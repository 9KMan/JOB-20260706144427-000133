---
phase: 4
plan: 01
type: IMPLEMENTATION
wave: 2
depends_on:
  - phase: 1
  - phase: 2
  - phase: 3
plans: ["01"]
files_modified:
  - internal/model/item.go
  - internal/store/memory.go
autonomous: true
requirements:
  - Item entity: id (UUID), name (string), created_at (timestamp)
  - In-memory store with sync.RWMutex
  - Pre-seeded with 2 items
---

# Phase 4 — Data Model (Item entity + in-memory store)

## What this phase delivers

The data model for the PoC: a single `Item` entity and an in-memory store
with thread-safe CRUD operations. Production would swap to Postgres via
the same `ItemStore` interface.

## Files to Create

Already created in Phase 1 — this phase documents the model:

| File | Data model role |
|---|---|
| `internal/model/item.go` | Item struct definition |
| `internal/store/memory.go` | ItemStore interface + memory impl |

## Schema

### Item
```go
type Item struct {
    ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
    Name      string    `json:"name" example:"First item"`
    CreatedAt time.Time `json:"created_at" example:"2026-07-06T15:30:00Z"`
}
```

### ItemStore interface
```go
type ItemStore interface {
    List(ctx context.Context) ([]model.Item, error)
    Create(ctx context.Context, name string) (model.Item, error)
    Get(ctx context.Context, id string) (model.Item, error)
}
```

## Production migration path

The `ItemStore` interface allows swapping the in-memory implementation for a
Postgres-backed one without changing handlers:

```go
// Future: internal/store/postgres.go
type PostgresStore struct { db *sql.DB }
func (s *PostgresStore) List(ctx context.Context) ([]model.Item, error) { ... }
// ... implements ItemStore
```

The Cloud SQL Postgres connection would use the same Go `database/sql`
driver pattern. Schema migration would use a tool like `golang-migrate`.

## Verification

```bash
go test ./internal/store/... -v
# expect: TestMemoryStore passes all CRUD operations
```

## Acceptance criteria

- [ ] `Item` struct has 3 fields with JSON tags
- [ ] `ItemStore` interface has 3 methods
- [ ] Memory store uses `sync.RWMutex` for concurrency
- [ ] Store pre-seeds 2 items on startup
- [ ] UUIDs are v4 (RFC 4122 compliant)