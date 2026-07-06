---
phase: 3
plan: 01
type: IMPLEMENTATION
wave: 2
depends_on:
  - phase: 1
  - phase: 2
plans: ["01"]
files_modified:
  - internal/api/handlers.go
  - internal/api/handlers_test.go
autonomous: true
requirements:
  - REST API: /health, /api/items (GET, POST), /api/items/:id (GET)
  - CORS middleware
  - Request ID propagation
  - Structured JSON responses
---

# Phase 3 — Architecture (API surface + integration patterns)

## What this phase delivers

The architectural shape: HTTP API contract that the Flutter app will call,
with CORS for cross-origin requests, request ID propagation for tracing, and
structured JSON responses.

## Files to Create

Already created in Phase 1 — this phase verifies the architectural decisions:

| File | Architectural decision |
|---|---|
| `internal/api/handlers.go` | REST API: standard CRUD with resource-oriented URLs |
| `internal/api/middleware.go` | Request ID + CORS + Recover + Logger (cross-cutting concerns) |
| `internal/store/memory.go` | Interface-based store: swappable to Postgres later |

## Architecture decisions documented

### Why REST over GraphQL/gRPC
- Cloud Run + Flutter works fine with REST; lower client complexity
- gRPC would require Flutter gRPC client setup (more boilerplate)
- GraphQL adds server complexity not needed for PoC

### Why CORS = `*` for PoC
- Flutter web preview may call the API from `localhost:3000` etc.
- Production CORS would be `https://grux.example.com`
- Documented in middleware comment

### Why request ID propagation
- Production observability: trace a request from Flutter → API → logs
- PoC: just generate + log; production: add OpenTelemetry

## Verification

```bash
# Start server in one terminal
go run ./cmd/api

# Test CORS preflight
curl -X OPTIONS http://localhost:8080/api/items \
  -H 'Origin: http://localhost:3000' \
  -H 'Access-Control-Request-Method: POST' \
  -i
# Expect: Access-Control-Allow-Origin: *

# Verify request ID is returned
curl -i http://localhost:8080/health
# Expect: X-Request-Id header in response
```

## Acceptance criteria

- [ ] OPTIONS preflight returns CORS headers
- [ ] Every response includes `X-Request-Id` header
- [ ] Server logs include the same `request_id` field
- [ ] All endpoints return valid JSON (no raw HTML error pages)