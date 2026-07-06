# Out of Scope — Grux PoC

This document captures what is **intentionally not built** in the Grux
Proof-of-Concept, why each item is deferred, and the production-build
roadmap for getting there. Every feature below has a concrete code-shape
target so the PoC can evolve into a production system without a
ground-up rewrite.

The PoC demonstrates the **stack**: a Go HTTP API on Cloud Run, a
Flutter mobile client, and a Cloud Build pipeline that ships them
together. It is single-tenant, unauthenticated, in-memory, and ships a
generic "items" resource instead of the real Grux domain model.

## Deferred features

### 1. Authentication / OAuth

- **What's deferred**: User login, JWT issuance and verification,
  Workspace SSO, session management, password reset flows.
- **Why deferred**: A whole-user-system is its own scope: identity
  provider integration, token rotation, RBAC, audit logging. None of
  that is needed to prove the API shape, the deploy pipeline, or the
  mobile client work end-to-end.
- **Production path**: Google Workspace SSO via OAuth 2.0 (authorization
  code + PKCE on the Flutter client). API authenticates with signed JWT
  bearer tokens; verification happens in middleware before handlers
  run. Tokens are short-lived (15 min) with refresh tokens stored
  securely on-device. Service-to-service calls use Google ID tokens.
- **Code reference**: future `internal/auth/middleware.go` (JWT
  verification + tenant extraction), `internal/auth/oauth.go` (token
  exchange + refresh), `internal/auth/session.go` (server-side session
  store on Cloud Firestore or Redis).

### 2. Postgres database

- **What's deferred**: Persistent storage, schema migrations, connection
  pooling, backup/restore, indexes tuned for production query patterns.
- **Why deferred**: The PoC uses an in-memory store to demonstrate the
  API surface and the contract between the API and the Flutter client.
  Wiring Postgres now adds infra (Cloud SQL, VPC connector, secret
  manager) without changing what the PoC proves.
- **Production path**: Cloud SQL for Postgres with `golang-migrate`
  migrations checked into the repo under `db/migrations/`. The
  `internal/store/memory.go` package is swapped for
  `internal/store/postgres.go` against the same `ItemStore` interface,
  so handlers stay unchanged. Connection pool is tuned for Cloud Run's
  per-instance concurrency. Read replicas for analytics queries.
- **Code reference**: future `internal/store/postgres.go` (uses
  `database/sql` + `pgx` pool), `internal/store/memory.go` (kept only
  for tests), `db/migrations/0001_init.sql` and friends.

### 3. Multi-tenancy

- **What's deferred**: Clinic / tenant isolation, per-tenant data
  partitioning, tenant-scoped queries, tenant onboarding flow,
  per-tenant rate limits and quotas.
- **Why deferred**: Multi-tenancy is a cross-cutting concern that
  touches every layer (DB schema, middleware, error reporting,
  observability). Building it on top of a single-tenant PoC is cheaper
  than retrofitting later — but only after the entity model is real.
- **Production path**: `tenant_id` (UUID) on every domain entity, a
  Postgres `SET app.current_tenant = $1` on every connection acquired
  from the pool (via `RLS`-compatible views or app-layer middleware),
  and a tenant extractor middleware that reads the JWT claim and
  attaches the tenant to `context.Context`. For strongly regulated
  tenants, schema-per-tenant isolation with a router.
- **Code reference**: future `internal/middleware/tenant.go` (extracts
  tenant from JWT, sets context), `internal/store/tenant_scoped.go`
  (wraps any store with tenant filter), `db/migrations/0002_tenants.sql`.

### 4. WhatsApp Business API integration

- **What's deferred**: Outbound WhatsApp message sends, inbound webhook
  reception, template management, opt-in / opt-out tracking, delivery
  status reconciliation.
- **Why deferred**: WhatsApp Cloud API requires Meta Business Manager
  approval, a verified business, a phone number registration, and
  per-template approval — none of which fit a PoC timeline. The
  integration also introduces async webhooks (Pub/Sub-backed) that
  need their own reliability story.
- **Production path**: WhatsApp Cloud API via Meta with a webhook
  handler fronted by a Google Cloud Function that publishes events to
  Pub/Sub. The API subscribes to a `whatsapp_inbound` topic and fans
  out to handlers. Outbound uses Meta's templated messages with a
  template-approval workflow surfaced to ops. Opt-in is tracked per
  user with explicit consent capture.
- **Code reference**: future `internal/integrations/whatsapp/client.go`
  (Cloud API wrapper), `internal/integrations/whatsapp/webhook.go`
  (signature verification + Pub/Sub publish),
  `internal/integrations/whatsapp/templates.go` (template render + send).

### 5. Stripe payments

- **What's deferred**: Payment intents, subscriptions, invoices,
  webhook signature verification, refunds, dunning, revenue
  recognition, reconciliation reports.
- **Why deferred**: Payments are a separate domain with strict
  regulatory and idempotency requirements. The PoC has no concept of a
  paying customer yet.
- **Production path**: Stripe SDK with idempotent payment intents keyed
  by `(tenant_id, idempotency_key)`. Webhook endpoint verifies Stripe
  signatures with the `Stripe-Signature` header and a stored webhook
  secret in Secret Manager. Reconciliation is a daily Cloud Run job
  that diffs Stripe state against our internal ledger. Subscription
  state changes flow into Pub/Sub and trigger entitlement updates.
- **Code reference**: future `internal/integrations/stripe/client.go`,
  `internal/integrations/stripe/webhook.go`,
  `internal/integrations/stripe/reconcile.go` (daily job entrypoint).

### 6. Real Grux / Rook product logic

- **What's deferred**: The actual Grux domain model (knowledge
  management, document collections, semantic search), and the Rook
  companion features (chat history, voice notes, agent memory).
- **Why deferred**: The PoC's job is to validate the stack — Go API on
  Cloud Run, Flutter client, deploy pipeline. The generic `Item` model
  is intentionally minimal so reviewers can see the architecture
  without needing product context.
- **Production path**: Replace the `Item` model with the real domain
  entities (Documents, Workspaces, Users, Permissions, KnowledgeNodes,
  Agents, Conversations, etc.). The patterns proven by the PoC —
  handler → service → store, provider state on the client, JSON
  contracts — apply unchanged. The `ItemStore` interface becomes a set
  of entity-specific interfaces behind the same handler pattern.
- **Code reference**: future `internal/domain/<entity>/` packages,
  `mobile/lib/models/` swap from `item.dart` to per-entity files, same
  Provider pattern.

### 7. CI/CD beyond Cloud Build

- **What's deferred**: GitHub Actions for PR checks, multi-environment
  promotion (dev → staging → prod), automated rollback on health-check
  failure, preview environments per PR.
- **Why deferred**: A single Cloud Build trigger that ships on merge
  to `main` is sufficient for one deployment target in a PoC.
- **Production path**: GitHub Actions for PR-time checks (Go `go vet`,
  `go test`, `golangci-lint`; Flutter `flutter analyze`, `flutter test`;
  Docker hadolint). Cloud Build handles the deploy itself. Manual
  promotion via `gcloud builds submit` from staging image tag to prod.
  Optional: Cloud Deploy for progressive delivery with canary
  percentages and automatic rollback on SLO breach.
- **Code reference**: future `.github/workflows/pr-checks.yaml`,
  `.github/workflows/release.yaml`, `deploy/cloudbuild.staging.yaml`,
  `deploy/cloudbuild.prod.yaml`.

### 8. Helm chart for GKE

- **What's deferred**: Kubernetes manifests, Helm templates, Kustomize
  overlays, custom resource definitions for Postgres operators.
- **Why deferred**: Cloud Run is enough for stateless APIs in the PoC.
  K8s adds operational surface area (node pools, upgrades, RBAC,
  ingress, cert-manager) that pays off only when stateful workloads or
  multi-region failover justify it.
- **Production path**: Helm chart for stateful services that genuinely
  need GKE — Cloud SQL Postgres operators, custom GPU workloads, jobs
  that need privileged sidecars. Stateless APIs stay on Cloud Run. The
  Helm chart lives in `infra/helm/` and is rendered by Argo CD or
  Flux against the prod cluster.
- **Code reference**: future `infra/helm/grux-poc/Chart.yaml`,
  `infra/helm/grux-poc/templates/`, `infra/argocd/application.yaml`.

### 9. Test coverage > 70%

- **What's deferred**: Comprehensive unit tests, integration tests
  against real databases, end-to-end Flutter integration tests, load
  tests, chaos tests, contract tests between the API and the mobile
  client.
- **Why deferred**: The PoC targets **at least one test per layer** to
  prove the stack is testable end-to-end — one Go handler test, one
  Flutter widget test. That demonstrates the *capability* of testing
  without spending a PoC's timeline on coverage.
- **Production path**: Unit tests cover business logic (handlers,
  services, validators). Integration tests run against
  `testcontainers-go` Postgres + ephemeral Cloud Run instances. E2E
  Flutter tests with `package:integration_test` exercise the full
  create-list flow against a deployed test environment. Load tests with
  k6 target the documented SLOs (e.g., 95p < 200ms, 100 RPS sustained).
  Contract tests between the Flutter `ApiClient` and the Go handlers
  run on every PR via GitHub Actions.
- **Code reference**: future `internal/handlers/*_test.go`,
  `internal/services/*_test.go`, `mobile/test/integration/` (E2E),
  `tests/load/k6.js`, `tests/contract/openapi.yaml`.

### 10. Observability beyond Cloud Logging

- **What's deferred**: Structured tracing (OpenTelemetry), custom
  metrics dashboards, alerting on SLO burn rate, error tracking (Sentry
  or equivalent), user-facing audit logs.
- **Why deferred**: Cloud Run ships with Cloud Logging out of the
  box, which is enough to debug a PoC. Tracing and metrics only earn
  their cost when there's enough traffic to be statistically
  meaningful.
- **Production path**: OpenTelemetry SDK in the Go API auto-instruments
  `net/http` and exports to Cloud Trace + Cloud Monitoring. Flutter
  client emits client-side errors and latency to the same backend via
  a fire-and-forget `telemetry` topic on Pub/Sub. SLO dashboards in
  Cloud Monitoring with multi-window burn-rate alerts. PagerDuty
  rotation on tier-1 alerts.
- **Code reference**: future `internal/observability/tracing.go`,
  `internal/observability/metrics.go`,
  `mobile/lib/services/telemetry_client.dart`.

### 11. Internationalization (i18n) and accessibility (a11y)

- **What's deferred**: Multi-locale support on the Flutter client,
  RTL layouts, screen-reader semantics audits, WCAG 2.1 AA
  conformance, dynamic type support across breakpoints.
- **Why deferred**: PoC content is English-only and the team is
  English-only at PoC stage. i18n retrofit is expensive once hard-coded
  strings are everywhere.
- **Production path**: `flutter_localizations` + ARB files from day
  one of the real product build. Locale picker in settings.
  Accessibility tests via `package:accessibility` and manual screen
  reader passes per release. Design tokens support RTL by construction
  (logical properties, not left/right).
- **Code reference**: future `mobile/lib/l10n/app_en.arb`,
  `mobile/lib/l10n/app_es.arb`, `mobile/test/accessibility/`.

### 12. Web client

- **What's deferred**: A Flutter Web build, browser-specific
  optimizations, PWA installability, search-engine landing pages.
- **Why deferred**: The PoC proves mobile with Flutter; shipping a
  web build is a `flutter build web` away, but routing, auth flows,
  and SEO are different problems. Worth doing once the mobile client
  is stable.
- **Production path**: Enable Flutter Web with `flutter build web`,
  deploy the static bundle to Firebase Hosting or Cloud Storage +
  Cloud CDN. Reuse the existing `ApiClient` and provider state. Web-
  specific routing with `go_router` deep links. PWA manifest for
  install.
- **Code reference**: future `web/index.html`, `web/manifest.json`,
  `mobile/lib/router/web_router.dart`.

## How to use this document

- **New engineer onboarding**: read this top-to-bottom before opening
  the codebase. Every "why deferred" tells you the constraint that
  shaped the PoC.
- **Production planning**: each "production path" is sized to a single
  quarter's work; the "code reference" is the directory or file the
  change will land in.
- **Reviewing this PoC**: if a feature you expected is missing, check
  this list first. If it's not here, the omission is unintentional —
  flag it.

## Production-build roadmap (suggested)

| Quarter | Theme | Items |
|---|---|---|
| Q1 | Make it real | #1 Auth, #2 Postgres, #6 Real domain model |
| Q2 | Make it multi-tenant | #3 Multi-tenancy, #9 Tests to 70%+ |
| Q3 | Make it integrate | #4 WhatsApp, #5 Stripe, #10 Observability |
| Q4 | Make it global | #7 Multi-env CI/CD, #8 Helm if needed, #11 i18n + #12 Web if needed |

Order is a suggestion, not a contract. Auth (#1) and Postgres (#2) are
the only items that should be done before the real Grux / Rook domain
model (#6) is feasible.