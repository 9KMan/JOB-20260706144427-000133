package api

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// HeaderRequestID is the response and request header carrying the
	// per-request correlation ID.
	HeaderRequestID = "X-Request-Id"

	// ginContextKeyRequestID is the key under which the request ID is
	// stored on the gin.Context.
	ginContextKeyRequestID = "request_id"
)

// GetRequestID returns the request ID stored on the gin context, or the
// empty string if none is present.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(ginContextKeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if id := c.Writer.Header().Get(HeaderRequestID); id != "" {
		return id
	}
	return ""
}

// RequestID middleware mints a UUID v4 per request and exposes it via
// the X-Request-Id header and the gin context. Downstream middleware and
// handlers can retrieve it via GetRequestID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.NewString()
		c.Set(ginContextKeyRequestID, id)
		c.Writer.Header().Set(HeaderRequestID, id)
		c.Next()
	}
}

// Logger middleware emits one structured slog line per request after it
// completes. Fields include method, path, status, latency, client IP,
// and request_id.
func Logger(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "request",
			slog.String("request_id", GetRequestID(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

// Recover middleware catches panics from downstream handlers, logs them
// with a stack trace, and returns a generic 500 response so the server
// keeps serving subsequent requests.
func Recover(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					slog.String("request_id", GetRequestID(c)),
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(http.StatusInternalServerError,
						errorResponse{Error: "internal error"})
				}
			}
		}()
		c.Next()
	}
}

// CORS middleware sets permissive cross-origin headers appropriate for the
// PoC. Access-Control-Allow-Origin is "*" — production deployments should
// restrict this to an explicit allow-list of origins.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-Id")
		h.Set("Access-Control-Expose-Headers", "X-Request-Id")
		h.Set("Access-Control-Max-Age", "600")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
