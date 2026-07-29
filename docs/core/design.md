---
name: Design Decisions
relation: index.md → core/
description: Design decisions and rationale to prevent agent drift
type: Enforce
---

# Design Decisions

## No framework (2026-07-20)

**Decision:** Use Go stdlib `net/http` with Go 1.22+ `http.ServeMux`. No Gin, Echo, Chi, or any third-party HTTP framework.

**Rationale:** Go's stdlib is production-grade. Go 1.22 added method-based routing (`GET /users/{id}`, `POST /users`) and path parameters (`r.PathValue("id")`), eliminating the two main reasons teams reached for third-party routers. Adding a framework introduces dependency risk, abstraction leakage, and framework-specific magic that fights Go's explicit nature.

## Raw SQL over ORM (2026-07-20)

**Decision:** Use `pgx` directly with hand-written SQL. No GORM, ent, or sqlboiler.

**Rationale:** ORMs in Go introduce reflection overhead, surprising behavior (silent zero-value issues in GORM), and obscure the actual queries hitting the database. Complex queries inevitably require raw SQL anyway, creating a hybrid that's worse than either pure approach. Direct pgx gives full control over query plans, native Postgres type support (JSONB, arrays, UUID), and predictable performance.

## Domain isolation (2026-07-20)

**Decision:** Each domain is a standalone Go package under `internal/domain/`. Domains do not import each other's entities.

**Rationale:** Cross-domain imports create tight coupling and circular dependency issues. If a domain needs data from another domain, it goes through the service layer (e.g., `workspace.Service` called by a higher-level orchestrator), not by importing workspace entities into the user package.

## Manual dependency injection (2026-07-20)

**Decision:** Wire all dependencies in `cmd/server/main.go`. No DI framework, no reflection, no `wire`.

**Rationale:** Explicit wiring is the Go way — it's traceable, debuggable, and requires no code generation. The tradeoff is more lines in `main.go` as the project grows, but this is acceptable for a monolith with a handful of domains. If the project grows to dozens of domains, consider a `wire`-based approach, but start simple.

## Repository interfaces for testability (2026-07-20)

**Decision:** Every domain has a `repository.go` file defining an interface, and a `repository_pg.go` implementing it with pgx.

**Rationale:** The interface lets services be unit-tested with mock repositories without spinning up a database. It also makes it trivial to swap implementations (e.g., adding Redis caching via a decorated repository).

## Custom migration runner (2026-07-20)

**Decision:** Use a simple migration runner in `internal/infrastructure/database/migrate.go` that reads `.sql` files and tracks applied versions in a `schema_migrations` table.

**Rationale:** `golang-migrate/migrate` requires driver registration and database/sql compatibility layers. A custom runner with `embed` or filesystem reads is ~60 lines, has zero extra dependencies, and is fully transparent. Down migrations are handled manually via SQL scripts when needed.

## UUID generation (2026-07-20)

**Decision:** Generate UUIDs client-side in Go using `crypto/rand` (`internal/pkg/id/id.go`), not database-side `gen_random_uuid()`.

**Rationale:** Client-side UUIDs let services create entities and return them immediately without a round-trip to get the generated ID. The stdlib approach avoids adding the `google/uuid` dependency.
