// Package api wires HTTP handlers, middleware, and the route table for the
// grux-poc-api service. It depends on the storage interface defined in the
// internal/store package and the model package for payload types.
package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/9KMan/JOB-20260706144427-000133/internal/model"
	"github.com/9KMan/JOB-20260706144427-000133/internal/store"
)

const (
	// ServiceName and ServiceVersion are reported by the health endpoint.
	ServiceName    = "grux-poc"
	ServiceVersion = "0.1.0"
)

// Handler holds the dependencies required by HTTP handlers. Handlers are
// constructed once and reused across requests.
type Handler struct {
	store  store.ItemStore
	logger *slog.Logger
}

// New returns a Handler bound to the provided store. If logger is nil,
// slog.Default() is used.
func New(s store.ItemStore, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{store: s, logger: logger}
}

// RegisterRoutes attaches API routes to the provided gin engine.
func RegisterRoutes(r *gin.Engine, h *Handler) {
	r.GET("/health", h.Health)
	items := r.Group("/api/items")
	{
		items.GET("", h.ListItems)
		items.POST("", h.CreateItem)
		items.GET("/:id", h.GetItem)
	}
}

// createItemRequest is the JSON body accepted by POST /api/items.
type createItemRequest struct {
	Name string `json:"name"`
}

// errorResponse is the standard error envelope returned to clients.
type errorResponse struct {
	Error string `json:"error"`
}

// Health responds with service identity, status, and version. It is
// intended for platform health checks (e.g. Cloud Run startup probes).
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": ServiceName,
		"version": ServiceVersion,
	})
}

// ListItems returns all items in the store as a JSON array. An empty
// store yields an empty array, never null.
func (h *Handler) ListItems(c *gin.Context) {
	items := h.store.List()
	if items == nil {
		items = []model.Item{}
	}
	c.JSON(http.StatusOK, items)
}

// CreateItem validates the request body, persists the item, and returns
// the created resource with status 201. A missing or blank name returns 400.
func (h *Handler) CreateItem(c *gin.Context) {
	var req createItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	item, err := h.store.Create(req.Name)
	if err != nil {
		if errors.Is(err, store.ErrEmptyName) {
			c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		h.logger.Error("create item failed",
			slog.String("request_id", GetRequestID(c)),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// GetItem returns a single item by ID, or 404 when no such item exists.
func (h *Handler) GetItem(c *gin.Context) {
	id := c.Param("id")
	item, err := h.store.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			c.JSON(http.StatusNotFound, errorResponse{Error: "item not found"})
			return
		}
		h.logger.Error("get item failed",
			slog.String("request_id", GetRequestID(c)),
			slog.String("id", id),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}
	c.JSON(http.StatusOK, item)
}
