---
name: System Design
relation: index.md → core/
description: Architecture overview and system design documentation
type: Enforce
---

# System Design

## Project architecture

This is a **Go** backend service with a **domain-based** structure. The architecture uses a clean layered pattern where each business domain is encapsulated in its own package under `internal/domain/`.

## Domain architecture

```
internal/domain/{domainName}/
├── entity.go          — Domain struct and value objects (e.g., Session)
├── repository.go      — Repository interface(s) for testability
├── repository_pg.go   — Postgres implementation using pgx
├── service.go         — Business logic
└── handler.go         — HTTP handlers (transport layer)
```

### Data flow

```
HTTP request
  → handler.go (parse request, validate input, call service)
    → service.go (business logic, call repository interface)
      → repository_pg.go (raw SQL via pgx)
        → PostgreSQL
```

Each layer depends on an **interface** (defined in `repository.go`), so services are unit-testable without a real database.

## Project layout

```
cmd/server/
├── main.go                  — Entry point: config → DB → migrate → wire DI → serve
└── migrations/              — SQL migration files

internal/
├── domain/                  — Business domains
│   ├── auth/
│   ├── user/
│   └── berita/
├── infrastructure/          — Shared infrastructure
│   ├── database/            — pgxpool connection, migration runner
│   ├── middleware/          — CORS, logging, auth, role
│   ├── response/            — JSON response helpers
│   └── storage/             — S3-compatible storage (Backblaze B2 / MinIO)
├── pkg/id/                  — UUID v4 generator (stdlib only)
└── router.go                — Route registration + middleware chain

config/
└── config.go                — Env-based config loader
```

## Entry point flow (`cmd/server/main.go`)

1. Load config from environment variables
2. Connect to Postgres via pgxpool
3. Run pending SQL migrations
4. Instantiate repositories → services → handlers (manual DI)
5. Register routes on `http.ServeMux`
6. Instantiate middleware (auth token validator, role checker)
7. Apply middleware chain (logger → CORS — globally; auth + role applied per-route on protected endpoints)
8. Start server with graceful shutdown on SIGINT/SIGTERM

## Key principles

- **No framework.** Uses Go stdlib `net/http` with Go 1.22+ `http.ServeMux` (supports method-based routing like `GET /users/{id}`).
- **Raw SQL.** All database queries are hand-written SQL via pgx. No ORM.
- **Domain isolation.** Each domain is a self-contained package. No cross-domain entity imports.
- **Interface-based repos.** `repository.go` defines the contract; `repository_pg.go` implements it. Swap implementations for testing.
- **Manual DI.** Dependencies are wired explicitly in `main.go`. No reflection, no DI framework.
- **Opaque bearer tokens.** Sessions stored in a `sessions` table. No JWT dependency. Instant revocation.
