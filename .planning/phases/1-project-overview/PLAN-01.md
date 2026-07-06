---
phase: 1
plan: 01
type: IMPLEMENTATION
wave: 1
depends_on: []
files_modified:
  - SPEC.md
  - README.md
autonomous: true
requirements:
  - Go 1.22 + Gin HTTP server
  - Health check endpoint
  - One CRUD endpoint (items)
  - In-memory store
  - Multi-stage Dockerfile
  - Structured logging
  - Graceful shutdown
---

# Phase 1 — Go Backend (HTTP server + Cloud Run-ready)

## What this phase delivers

A working Go HTTP server with structured logging, health check, and one CRUD
endpoint, packaged as a Cloud Run-ready Docker container. This is the spine
of the PoC — the Flutter app and GCP deploy scripts depend on it.

## Files to Create

| File | Purpose |
|---|---|
| `cmd/api/main.go` | Entrypoint: load config, set up Gin router, start server with graceful shutdown |
| `internal/api/handlers.go` | HTTP handlers: `health`, `listItems`, `createItem`, `getItem` |
| `internal/api/middleware.go` | Middleware: request ID, CORS, recover, structured logging |
| `internal/api/handlers_test.go` | Tests: health returns 200, items CRUD round-trips |
| `internal/store/memory.go` | In-memory `ItemStore` interface + implementation |
| `internal/store/memory_test.go` | Tests: store CRUD operations |
| `internal/model/item.go` | Item struct: `id`, `name`, `created_at` |
| `go.mod` | Module declaration: `github.com/9KMan/JOB-20260706144427-000133` |
| `go.sum` | Pinned dependencies (auto-generated) |
| `Dockerfile` | Multi-stage: `golang:1.22-alpine` builder → `gcr.io/distroless/static:nonroot` runtime |
| `.dockerignore` | Exclude `.git`, `.planning`, `*.md` (except SPEC/README) |
| `Makefile` | Targets: `build`, `test`, `run`, `docker-build`, `docker-run` |

## Implementation requirements

### `cmd/api/main.go`
- Use `github.com/gin-gonic/gin` for routing
- Use `log/slog` for structured JSON logging
- Config via env vars: `PORT` (default 8080), `LOG_LEVEL` (default info)
- Graceful shutdown on SIGINT/SIGTERM with 10s timeout
- Bind to `0.0.0.0:$PORT` (Cloud Run requirement)

### `internal/api/handlers.go`
- `GET /health` → 200 `{"status":"ok","service":"grux-poc","version":"0.1.0"}`
- `GET /api/items` → 200 `[]` (empty list initially)
- `POST /api/items` → 201 with created item (validates name non-empty)
- `GET /api/items/:id` → 200 or 404
- All responses use JSON

### `internal/api/middleware.go`
- RequestID middleware: generates UUID, sets `X-Request-Id` header, adds to context
- CORS middleware: allow `*` for PoC (production would restrict origins)
- Recovery middleware: catches panics, returns 500 with generic message
- Logger middleware: logs method/path/status/latency/duration/request_id

### `internal/store/memory.go`
- Define `ItemStore` interface: `List()`, `Create(item)`, `Get(id)`
- Implement with `sync.RWMutex` for thread safety
- Items have UUID v4 IDs
- Pre-seed with 2 items on startup so list isn't empty

### `internal/model/item.go`
```go
type Item struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}
```

### `Dockerfile`
```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api ./cmd/api

# Runtime stage
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /api /api
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
```

### `Makefile`
```makefile
.PHONY: build test run docker-build docker-run

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -v

run:
	go run ./cmd/api

docker-build:
	docker build -t grux-poc-api:latest .

docker-run:
	docker run -p 8080:8080 grux-poc-api:latest
```

## Verification

After implementation, run from the repo root:
```bash
go build ./...                    # expect 0 errors
go test ./... -v                  # expect all PASS
PORT=8080 go run ./cmd/api &       # start server
curl http://localhost:8080/health # expect {"status":"ok",...}
curl http://localhost:8080/api/items # expect [{...},{...}] (seeded)
curl -X POST http://localhost:8080/api/items -H 'Content-Type: application/json' -d '{"name":"test"}'
# expect 201 + new item
docker build -t grux-poc-api .    # expect success
```

## Acceptance criteria

- [ ] `go build ./...` returns 0 errors
- [ ] `go test ./...` returns all PASS (≥4 tests)
- [ ] `curl /health` returns 200 with valid JSON
- [ ] `curl /api/items` returns seeded items
- [ ] `docker build` succeeds, image is < 50 MB
- [ ] Server starts and shuts down gracefully (Ctrl-C exits cleanly)
- [ ] Logs are structured JSON with request_id field