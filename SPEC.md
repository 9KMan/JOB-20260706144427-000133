# SPEC — Grux Senior Full-Stack PoC (Go + Flutter + GCP)

## PoC Purpose

A working reference scaffold that demonstrates **the architectural shape of the
Grux + Rook products** (cloud-native, cross-platform, GCP-deployable) using the
exact stack from the JD: Go backend + Flutter frontend + GCP infrastructure.

**This is a PoC for the bid, not a production system.** It's intentionally
minimal — a single end-to-end slice that proves the stack works together and
deploys to GCP Cloud Run.

## PoC Scope (decision-driven — 4 phases)

### Phase 1 — Go Backend (Go 1.22, Gin framework, single binary)

**Goal**: Working Go HTTP server with health check + one CRUD endpoint, deployable
to Cloud Run as a stateless container.

**Deliverables**:
- `cmd/api/main.go` — entrypoint with structured logging + graceful shutdown
- `internal/api/handlers.go` — health + items handlers
- `internal/api/middleware.go` — request ID + CORS + recover
- `internal/store/memory.go` — in-memory store (production would use Postgres)
- `go.mod` + `go.sum` — pinned deps
- `Dockerfile` — multi-stage build, distroless final image
- `.dockerignore`

**Acceptance**:
- `go run cmd/api/main.go` starts on :8080
- `curl http://localhost:8080/health` returns `{"status":"ok"}`
- `curl http://localhost:8080/api/items` returns `[]` initially
- `curl -X POST http://localhost:8080/api/items -d '{"name":"foo"}'` adds an item
- `go build ./...` returns 0 errors
- `go test ./...` returns 0 failures

### Phase 2 — Flutter App (Flutter 3.x, single screen)

**Goal**: Flutter mobile/web app that calls the Go backend and displays the
items list, demonstrating cross-platform capability.

**Deliverables**:
- `mobile/lib/main.dart` — app entry, Material theme, routing
- `mobile/lib/screens/items_screen.dart` — list + add UI
- `mobile/lib/services/api_client.dart` — HTTP client wrapping the Go API
- `mobile/lib/models/item.dart` — Item model
- `mobile/pubspec.yaml` — pinned deps (http, provider or riverpod)
- `mobile/analysis_options.yaml` — lints config

**Acceptance**:
- `cd mobile && flutter pub get` succeeds
- `cd mobile && flutter analyze` returns 0 errors
- `cd mobile && flutter test` returns 0 failures (at least 1 widget test)
- App compiles for at least Android (Flutter's default platform)

### Phase 3 — GCP Infrastructure (Cloud Build + Cloud Run deploy)

**Goal**: One-command GCP deployment that proves the Cloud Run path works.

**Deliverables**:
- `deploy/cloudbuild.yaml` — Cloud Build config: build → push to Artifact Registry → deploy to Cloud Run
- `deploy/deploy.sh` — bootstrap script: enable APIs, create Artifact Registry, run Cloud Build
- `deploy/env.example` — env var template (GCP_PROJECT, GCP_REGION, SERVICE_NAME)
- `deploy/README.md` — deploy instructions

**Acceptance**:
- `bash -n deploy/deploy.sh` parses without errors
- `deploy/cloudbuild.yaml` validates against Cloud Build schema
- README explains the GCP setup steps clearly

### Phase 4 — Top-level README + RUNBOOK + canonical files

**Goal**: Repo surface that's reviewable in 5 minutes.

**Deliverables**:
- `README.md` — top-level: what this is, quick start, deploy to GCP, link to SPEC.md
- `RUNBOOK.md` — common operations: run locally, run tests, deploy to GCP, troubleshoot
- `OUT_OF_SCOPE.md` — what's intentionally not built (auth, DB, etc.)
- `.gitignore` — Go + Flutter + Python ignores
- `.shipped` — sentinel preventing rebuild (Pitfall 11)
- `.planning/JOB-133-BUILD-STATUS.md` — post-build status note

**Acceptance**:
- `grep -c "Business Problem Solved" README.md` ≥ 1 (per Bug 200 audit pattern)
- All canonical files present
- A reviewer can clone + read README + understand the stack in 5 minutes

## What's NOT in the PoC (and why)

| Out of scope | Why |
|---|---|
| Authentication / OAuth | Out of scope per PoC decision (production concern) |
| Postgres database | In-memory store is sufficient to demonstrate API shape |
| Multi-tenancy | Single-tenant PoC; production would add tenant isolation |
| WhatsApp Business API integration | Different stack layer, separate PoC |
| Stripe payments | Different stack layer, separate PoC |
| Real Grux/Rook product logic | PoC uses generic "items" entity to demonstrate patterns |
| CI/CD beyond Cloud Build | One deployment target is enough for PoC |
| Helm chart for GKE | Cloud Run is sufficient for the PoC |
| Test coverage > 70% | PoC targets ≥1 test per layer to prove testability |

## Why this shape demonstrates Grux/Rook fit

1. **Cloud-native architecture** — same pattern Grux/Rook use (Cloud Run + Pub/Sub + Cloud SQL)
2. **Cross-platform capability** — Go backend serves web + mobile + future clients (Flutter for mobile/web)
3. **Production discipline** — multi-stage Docker, structured logging, graceful shutdown, health checks
4. **GCP-native deployment** — `cloudbuild.yaml` is the exact pattern a Grux engineer would write
5. **Architectural fit** — function-oriented service decomposition (Phase 1's `internal/api` + `internal/store` separation matches Grux's service boundaries)

## How this maps to the JD

| JD requirement | PoC evidence |
|---|---|
| Go backend | `cmd/api/main.go` + `internal/api/handlers.go` |
| Flutter frontend | `mobile/lib/main.dart` + `mobile/lib/screens/items_screen.dart` |
| GCP deploy | `deploy/cloudbuild.yaml` + `deploy/deploy.sh` |
| Cloud-native architecture | Multi-stage Docker + Cloud Run config |
| Production-grade quality | Structured logging + graceful shutdown + health check |
| Test discipline | Go tests + Flutter widget tests |
| Function-oriented architecture | `internal/api` vs `internal/store` separation |

## Estimated time

- Phase 1 (Go): ~20-30 min OpenCode
- Phase 2 (Flutter): ~20-30 min OpenCode
- Phase 3 (GCP): ~10-15 min OpenCode
- Phase 4 (README): ~10-15 min OpenCode
- Total: ~60-90 min OpenCode execution

## Deliverable check

After all phases complete:
```bash
cd /home/deploy/squad/build-worker/JOB-20260706144427-000133
find . -name "*.go" -not -path "*/.git/*" | wc -l   # expect ≥6
find ./mobile -name "*.dart" -not -path "*/.dart_tool/*" | wc -l  # expect ≥4
go build ./... 2>&1 | head -5                       # expect 0 errors
cd mobile && flutter analyze 2>&1 | tail -5          # expect 0 errors
```