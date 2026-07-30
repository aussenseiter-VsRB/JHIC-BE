---
name: Domain Creation Rules
relation: index.md → modules/
description: Enforceable rules for creating and structuring Go backend domains
type: Enforce
---

# Domain Creation Rules

## 1. Domain structure

Every business domain is a **top-level package** under `internal/domain/`. Each domain must include this exact file set:

```
internal/domain/{domainName}/
├── entity.go          — Domain struct (no JSON tags on internal fields)
├── repository.go      — Repository interface (methods the service needs)
├── repository_pg.go   — pgx implementation of the repository
├── service.go         — Business logic (calls repository interface)
└── handler.go         — HTTP handlers + route registration
```

### Optional files

- `types.go` — Request/response types if they grow beyond the handler file.
- `errors.go` — Sentinel errors for the domain.

## 2. File contracts

### entity.go

Define the domain struct. Use `json:"name,omitempty"` tags for optional fields.

```go
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### repository.go

Define the interface the service layer depends on. Keep it minimal — only methods the service actually calls.

```go
type Repository interface {
    Create(ctx context.Context, u *User) error
    ByID(ctx context.Context, id string) (*User, error)
    List(ctx context.Context) ([]User, error)
    Update(ctx context.Context, u *User) error
    Delete(ctx context.Context, id string) error
}
```

### repository_pg.go

Implement the repository using `pgx`. Constructor takes `*pgxpool.Pool`.

```go
type RepositoryPG struct {
    pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *RepositoryPG {
    return &RepositoryPG{pool: pool}
}
```

- Use `pgx.ErrNoRows` for missing records → return `nil, nil`.
- Wrap all errors with context: `fmt.Errorf("method name: %w", err)`.

### service.go

Business logic layer. Constructor takes the `Repository` interface.

```go
type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}
```

- Use `internal/pkg/id.New()` for UUID generation.
- Timestamps use `time.Now().UTC()`.
- Validate business rules here (e.g., duplicate email check).

### handler.go

HTTP transport layer. Constructor takes `*Service`.

```go
type Handler struct {
    svc *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{svc: svc}
}
```

Must implement a `Register` method. For public or auth-only domains, accept only the mux:

```go
func (h *Handler) Register(mux *http.ServeMux) {
    mux.HandleFunc("GET /api/v1/users", h.List)
    mux.HandleFunc("GET /api/v1/users/{id}", h.Get)
    mux.HandleFunc("PUT /api/v1/users/{id}", h.Update)
    mux.HandleFunc("DELETE /api/v1/users/{id}", h.Delete)
}
```

For domains that require auth + role middleware on all routes, accept the middleware functions as parameters:

```go
func (h *Handler) Register(mux *http.ServeMux, authMw func(http.Handler) http.Handler, roleMw func(http.Handler) http.Handler) {
    mux.Handle("POST /api/v1/berita", authMw(roleMw(http.HandlerFunc(h.Create))))
    mux.Handle("GET /api/v1/berita", authMw(roleMw(http.HandlerFunc(h.List))))
    mux.Handle("GET /api/v1/berita/{id}", authMw(roleMw(http.HandlerFunc(h.Get))))
    mux.Handle("PUT /api/v1/berita/{id}", authMw(roleMw(http.HandlerFunc(h.Update))))
    mux.Handle("DELETE /api/v1/berita/{id}", authMw(roleMw(http.HandlerFunc(h.Delete))))
    mux.Handle("POST /api/v1/berita/{id}/image", authMw(roleMw(http.HandlerFunc(h.UploadImage))))
}
```

- Parse request bodies with `json.NewDecoder`.
- Validate input at the handler level (required fields, format).
- Use `response.JSON` and `response.Error` helpers for all responses.
- Path parameters use `r.PathValue("id")`.

## 3. Route registration

Each handler registers its own routes via `Register`. The central `internal/router.go` calls each handler's `Register`:

```go
func NewRouter(ah *auth.Handler, uh *user.Handler, bh *berita.Handler, authMw func(http.Handler) http.Handler, roleMw func(http.Handler) http.Handler) http.Handler {
    mux := http.NewServeMux()
    ah.Register(mux)
    uh.Register(mux)
    bh.Register(mux, authMw, roleMw)
    // ... global middleware chain
}
```

## 4. Wiring (DI)

All dependencies are wired in `cmd/server/main.go`:

```go
repo := domain.NewRepository(pool)
svc  := domain.NewService(repo)
hnd  := domain.NewHandler(svc)
```

Order: repository depends on pool → service depends on repository → handler depends on service.

## 5. Naming conventions

| Artifact | Pattern | Example |
|---|---|---|---|
| Directory | `internal/domain/{name}` | `internal/domain/auth` |
| Entity struct | PascalCase singular | `type User struct` |
| Repository interface | `Repository` | `type Repository interface` |
| PG implementation | `RepositoryPG` | `type RepositoryPG struct` |
| Service | `Service` | `type Service struct` |
| Handler | `Handler` | `type Handler struct` |
| Constructor | `New{Type}` | `func NewRepository(pool *pgxpool.Pool) *RepositoryPG` |
| Handler methods | `List`, `Create`, `Get`, `Update`, `Delete` | `func (h *Handler) List` |
| Input types | `{Action}{Domain}Input` | `CreateUserInput` |

## 6. No dead code

Don't generate files, exports, or types that nothing consumes. Empty stubs create confusion and maintenance burden. Every exported symbol must have at least one consumer.
