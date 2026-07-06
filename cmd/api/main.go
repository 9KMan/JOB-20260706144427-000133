// Package main is the entrypoint for the grux-poc-api HTTP service.
//
// The service exposes a small CRUD API over a single resource ("items"),
// a health endpoint, structured JSON logging, and graceful shutdown suitable
// for Cloud Run deployments on GCP.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/9KMan/JOB-20260706144427-000133/internal/api"
	"github.com/9KMan/JOB-20260706144427-000133/internal/store"
)

const (
	defaultPort     = "8080"
	defaultLogLevel = "info"
	shutdownTimeout = 10 * time.Second
	readTimeout     = 10 * time.Second
	readHdrTimeout  = 5 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
)

// config holds runtime configuration loaded from the environment.
type config struct {
	Port     string
	LogLevel slog.Level
}

// loadConfig reads PORT and LOG_LEVEL from the environment, applying
// defaults when unset.
func loadConfig() config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = defaultPort
	}

	levelStr := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	var level slog.Level
	switch levelStr {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	case "", "info":
		level = slog.LevelInfo
	default:
		level = slog.LevelInfo
	}
	return config{Port: port, LogLevel: level}
}

func main() {
	cfg := loadConfig()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	logger.Info("starting service",
		slog.String("service", api.ServiceName),
		slog.String("version", api.ServiceVersion),
		slog.String("port", cfg.Port),
		slog.String("log_level", cfg.LogLevel.String()),
	)

	// Gin in production emits JSON-ish logs we don't want; we use slog instead.
	gin.SetMode(gin.ReleaseMode)

	memStore := store.NewMemoryStore()
	h := api.New(memStore, logger)

	router := gin.New()
	router.Use(
		api.RequestID(),
		api.Logger(logger),
		api.Recover(logger),
		api.CORS(),
	)
	api.RegisterRoutes(router, h)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHdrTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	// Run the server in a goroutine so we can select between a startup
	// failure and an OS shutdown signal in the main goroutine.
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-serverErr:
		if err != nil {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info("shutting down http server", slog.Duration("timeout", shutdownTimeout))
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("server stopped cleanly")
}
