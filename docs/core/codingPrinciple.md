---
name: Coding Principles
relation: index.md → core/
description: Code standards, naming conventions, and best practices
type: Enforce
---

# Coding Principles

## Core Tenets

### 1. Explicit over Implicit

- Errors are values, not exceptions. Always handle them.
- `context.Context` is always the first parameter of any function that performs I/O.
- Imports for the domain package are explicit — no `.` imports, no `init()` side effects.
- If something is required, the type system enforces it, not a runtime check.

### 2. Composition over Embedding

Prefer small interfaces composed together over large monolithic interfaces.

```go
// Good
type Reader interface { Read(ctx, id) (*T, error) }
type Writer interface { Create(ctx, *T) error }

// Avoid
type Repository interface {
    Reader
    Writer
    List(ctx) ([]T, error)
    Delete(ctx, id) error
}
```

Split repository interfaces by concern. The concrete `pg/` adapter implements all of them, but consumers only depend on what they need.

### 3. Colocation over Centralization

Code lives next to what it serves:
- Domain entity lives in `internal/domain/{name}/entity.go`
- Repository interface and implementation live in the same domain, with the pgx adapter in a `pg/` subpackage
- HTTP handlers live in the domain package, not in a central `handlers/` directory
- Only truly shared code goes in `internal/infrastructure/`

### 4. No Dead Code

Don't generate files, exports, or types that nothing consumes. Empty stubs create confusion and maintenance burden. If a file has no consumer, it shouldn't exist.

### 5. Progressive Complexity

Start simple. Add complexity only when the simple solution fails:
- Inline errors → sentinel errors → custom error types
- Direct repository calls in service → decorated repositories → event-sourced repositories
- Simple `http.ServeMux` → middleware chains → sub-routers

Don't build infrastructure for problems you don't have yet.

### 6. Interface Segregation

Define repository interfaces in `repository.go` with only the methods the domain needs. Don't create a "god interface" with every possible CRUD operation if the service only needs `Create` and `ByID`.

## Error Handling

- Validate inputs at the handler boundary (`handler.go`). Return 400 immediately.
- Business logic errors (duplicate email, not found) surface as `fmt.Errorf("descriptive message: %w", err)` from the service layer.
- Repository methods wrap errors with context: `fmt.Errorf("create user: %w", err)`.
- Handler maps errors to HTTP status codes. Never expose raw database errors to the client.
- Use `pgx.ErrNoRows` to detect missing records at the repository level — return `nil, nil` to the service layer, not the pgx error.

## Naming Conventions

| Layer | File | Naming |
|---|---|---|
| Entity | `entity.go` | `type User struct` |
| Repository interface | `repository.go` | `type Repository interface` |
| Postgres implementation | `pg/{adapter}.go` | `type Repository struct` (in `pg/`) |
| Service | `service.go` | `type Service struct` |
| Handler | `handler.go` | `type Handler struct` |

- Packages: lowercase, one word (`auth`, `user`), plus the `pg` adapter package.
- Files: snake_case (`pg/repository.go`).
- Types: PascalCase (`Repository`, `CreateUserInput`).
- Handlers: `func (h *Handler) List`, `Create`, `Get`, `Update`, `Delete`.

## Code Review Self-Check

After writing any code, ask:

1. Is it simple and necessary right now?
2. Does every exported symbol have a consumer?
3. Are errors wrapped with context?
4. Are contexts propagated from handler → service → repository?
5. Could a unit test verify this without a real database?
