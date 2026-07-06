package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/9KMan/JOB-20260706144427-000133/internal/model"
	"github.com/9KMan/JOB-20260706144427-000133/internal/store"
)

// newTestRouter builds a fully-wired gin engine using the real middleware
// stack and a real MemoryStore. Tests then exercise it through httptest.
func newTestRouter(t *testing.T) (*gin.Engine, *store.MemoryStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
	s := store.NewMemoryStore()
	h := New(s, logger)
	r := gin.New()
	r.Use(
		RequestID(),
		Logger(logger),
		Recover(logger),
		CORS(),
	)
	RegisterRoutes(r, h)
	return r, s
}

func TestHealth(t *testing.T) {
	r, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body["status"])
	}
	if body["service"] != ServiceName {
		t.Errorf("expected service=%s, got %v", ServiceName, body["service"])
	}
	if _, ok := body["version"]; !ok {
		t.Error("expected version field in health response")
	}
	if id := w.Header().Get(HeaderRequestID); id == "" {
		t.Error("expected X-Request-Id response header to be set")
	}
}

func TestListItemsSeeded(t *testing.T) {
	r, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var items []model.Item
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 seeded items, got %d", len(items))
	}
}

func TestCreateItem(t *testing.T) {
	r, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/items",
		strings.NewReader(`{"name":"test-item"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", w.Code, w.Body.String())
	}
	var item model.Item
	if err := json.Unmarshal(w.Body.Bytes(), &item); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if item.Name != "test-item" {
		t.Errorf("expected name=test-item, got %q", item.Name)
	}
	if item.ID == "" {
		t.Error("expected ID to be set")
	}
	if item.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCreateItemInvalidName(t *testing.T) {
	r, _ := newTestRouter(t)
	cases := map[string]string{
		"empty":      `{"name":""}`,
		"whitespace": `{"name":"   "}`,
		"tabs_only":  `{"name":"\t\n"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/items", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d, body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestCreateItemMalformedJSON(t *testing.T) {
	r, _ := newTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/api/items",
		strings.NewReader(`{not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", w.Code)
	}
}

func TestGetItem(t *testing.T) {
	r, s := newTestRouter(t)

	created, err := s.Create("target")
	if err != nil {
		t.Fatalf("setup create: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/items/"+created.ID, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}
	var got model.Item
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got.ID != created.ID || got.Name != created.Name {
		t.Errorf("mismatch: got=%+v want=%+v", got, created)
	}
}

func TestGetItemNotFound(t *testing.T) {
	r, _ := newTestRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/items/does-not-exist", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// TestItemCRUDRoundTrip verifies the create→get flow returns identical data,
// exercising the items endpoints as documented in PLAN-01.md.
func TestItemCRUDRoundTrip(t *testing.T) {
	r, _ := newTestRouter(t)

	// POST
	createBody := strings.NewReader(`{"name":"round-trip"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/items", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("POST: expected 201, got %d, body=%s", createW.Code, createW.Body.String())
	}
	var created model.Item
	if err := json.Unmarshal(createW.Body.Bytes(), &created); err != nil {
		t.Fatalf("POST json: %v", err)
	}

	// GET
	getReq := httptest.NewRequest(http.MethodGet, "/api/items/"+created.ID, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("GET: expected 200, got %d", getW.Code)
	}
	var fetched model.Item
	if err := json.Unmarshal(getW.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("GET json: %v", err)
	}
	if fetched.ID != created.ID || fetched.Name != created.Name {
		t.Errorf("round trip mismatch: created=%+v fetched=%+v", created, fetched)
	}
}

func TestCORSPreflight(t *testing.T) {
	r, _ := newTestRouter(t)
	req := httptest.NewRequest(http.MethodOptions, "/api/items", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS preflight, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Access-Control-Allow-Origin=*, got %q", got)
	}
}
