---
name: User Domain Example
relation: RULES.md → modules/examples/user/
description: Example domain showing the full handler → service → repository pattern
type: editable
---

# User Domain

This is an example domain for the `/api/v1/users` resource. Follow this pattern when creating new domains.

## Structure

```
internal/domain/user/
├── entity.go          — User struct (id, email, name, avatar_url, timestamps)
├── repository.go      — Repository interface (Create, ByID, ByEmail, List, Update, Delete)
├── repository_pg.go   — pgx implementation with raw SQL
├── service.go         — Business logic (duplicate email check, UUID generation)
└── handler.go         — HTTP handlers + route registration
```

## Endpoints

| Method | Path | Handler | Description |
|---|---|---|---|
| GET | /api/v1/users | List | List all users |
| POST | /api/v1/users | Create | Create a new user |
| GET | /api/v1/users/{id} | Get | Get user by ID |
| PUT | /api/v1/users/{id} | Update | Update user |
| DELETE | /api/v1/users/{id} | Delete | Delete user |

## Data flow

```
POST /api/v1/users
  → handler.Create: decode JSON body → validate required fields → call service.Create
    → service.Create: check duplicate email → generate UUID → set timestamps → call repo.Create
      → repository_pg.Create: INSERT INTO users ...
```

## How to use this example

1. Copy the directory structure (not the code).
2. Replace `user` with your domain name.
3. Define your entity in `entity.go`.
4. Define the repository interface in `repository.go`.
5. Implement with `pgx` in `repository_pg.go`.
6. Add business logic in `service.go`.
7. Write HTTP handlers in `handler.go`.
8. Register routes via the `Register(*http.ServeMux)` method.
9. Wire in `cmd/server/main.go` and register in `internal/router.go`.

## cURL examples

```bash
# Create a user
curl -X POST http://localhost:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"john@example.com","name":"John"}'

# List users
curl http://localhost:8080/api/v1/users

# Get user by ID
curl http://localhost:8080/api/v1/users/{id}

# Update user
curl -X PUT http://localhost:8080/api/v1/users/{id} \
  -H 'Content-Type: application/json' \
  -d '{"email":"john@example.com","name":"John Doe"}'

# Delete user
curl -X DELETE http://localhost:8080/api/v1/users/{id}
```
