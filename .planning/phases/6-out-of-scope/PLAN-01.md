---
phase: 6
plan: 01
type: IMPLEMENTATION
wave: 3
depends_on:
  - phase: 5
plans: ["01"]
files_modified:
  - OUT_OF_SCOPE.md
autonomous: true
requirements:
  - Explicit list of features NOT built in this PoC
  - Reason for each deferral
  - Production-build roadmap
---

# Phase 6 — Out-of-Scope (deferred features)

## What this phase delivers

An explicit `OUT_OF_SCOPE.md` documenting what is intentionally NOT built
in the PoC and the production roadmap for each deferred feature.

## Files to Create

| File | Purpose |
|---|---|
| `OUT_OF_SCOPE.md` | Top-level doc: deferred features + roadmap |

## Deferred features

### Authentication / OAuth
- **What's deferred**: User login, JWT tokens, Workspace SSO
- **Why**: Different scope (whole-user-system). Out of scope for PoC.
- **Production path**: Google Workspace SSO via OAuth 2.0; JWT for API auth;
  middleware-based auth checks before handlers
- **Code ref**: future `internal/auth/middleware.go` + `internal/auth/oauth.go`

### Postgres database
- **What's deferred**: Persistent storage; SQL migrations
- **Why**: PoC uses in-memory store to demonstrate API shape
- **Production path**: Cloud SQL Postgres + `golang-migrate` migrations;
  swap `internal/store/memory.go` for `internal/store/postgres.go` (same
  `ItemStore` interface)
- **Code ref**: future `internal/store/postgres.go` (uses `database/sql`)

### Multi-tenancy
- **What's deferred**: Clinic/tenant isolation; per-tenant data partitioning
- **Why**: Different concern (cross-cutting). PoC is single-tenant.
- **Production path**: tenant_id column on all entities + middleware that
  sets `SET app.current_tenant = $1` on every Postgres connection; or
  schema-per-tenant for stronger isolation
- **Code ref**: future `internal/middleware/tenant.go`

### WhatsApp Business API integration
- **What's deferred**: WhatsApp message send/receive, template management
- **Why**: Different integration pattern; needs Meta Business approval
- **Production path**: WhatsApp Cloud API via Meta + webhook handler +
  template approval workflow + opt-in management
- **Code ref**: future `internal/integrations/whatsapp/`

### Stripe payments
- **What's deferred**: Payment intents, subscriptions, webhooks
- **Why**: Out of PoC scope (separate domain)
- **Production path**: Stripe SDK + webhook signature verification +
  idempotent payment intents + reconciliation report
- **Code ref**: future `internal/integrations/stripe/`

### Real Grux/Rook product logic
- **What's deferred**: Knowledge management, Rook companion features
- **Why**: PoC uses generic "items" to demonstrate the stack
- **Production path**: Replace `Item` model with domain entities
  (Documents, Workspaces, Users, Permissions, etc.) — same patterns apply

### CI/CD beyond Cloud Build
- **What's deferred**: GitHub Actions, multi-env promotion, staging
- **Why**: One deployment target sufficient for PoC
- **Production path**: GitHub Actions for PR checks → staging → manual
  promotion to prod via Cloud Build trigger

### Helm chart for GKE
- **What's deferred**: Kubernetes manifests, Helm templates
- **Why**: Cloud Run is sufficient for the PoC
- **Production path**: Helm chart for stateful services that need GKE
  (Postgres operators, custom workloads); Cloud Run for stateless APIs

### Test coverage > 70%
- **What's deferred**: Comprehensive unit + integration + E2E tests
- **Why**: PoC targets ≥1 test per layer to prove testability
- **Production path**: Add integration tests against testcontainers-go;
  E2E tests with Flutter integration_test; load tests with k6

## Verification

```bash
test -f OUT_OF_SCOPE.md && echo "EXISTS" || echo "MISSING"
grep -c "Deferred" OUT_OF_SCOPE.md  # expect ≥9
```

## Acceptance criteria

- [ ] `OUT_OF_SCOPE.md` exists at repo root
- [ ] Documents ≥9 deferred features
- [ ] Each feature has: what's deferred, why, production path, code ref
- [ ] Total document ≥ 1,500 chars