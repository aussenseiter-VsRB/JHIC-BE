---
name: User Domain Documentation
relation: RULES.md → modules/user/
description: Documentation for the user domain — profile management and role assignment
type: editable
---

# User Domain

## Overview

The `user` domain manages user profiles and role assignments. It reads from the same `users` table as the auth domain but never exposes password hashes. Roles control access: `jurnal`, `guru`, `admin`, `user`.

## Entity

```go
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    Role      string    `json:"role"`
    AvatarURL string    `json:"avatar_url,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

## Endpoints

| Method | Path | Handler | Description |
|---|---|---|---|
| GET | /api/v1/users | List | List all users |
| GET | /api/v1/users/{id} | Get | Get user by ID |
| PUT | /api/v1/users/{id} | Update | Update profile (name, avatar_url) |
| PUT | /api/v1/users/{id}/role | UpdateRole | Update user role |
| DELETE | /api/v1/users/{id} | Delete | Delete user |

## Data flow

```
GET /api/v1/users
  → handler.List → service.List → repository_pg.List (SELECT without password_hash)

PUT /api/v1/users/{id}/role
  → handler.UpdateRole → service.UpdateRole (validates role is one of: jurnal, guru, admin, user)
    → repository_pg.UpdateRole (UPDATE users SET role = $2)
```

## Rules

- Valid roles: `jurnal`, `guru`, `admin`, `user`.
- The user domain shares the `users` table with the auth domain but never selects `password_hash`.
- Deleting a user cascades to their sessions (DB foreign key `ON DELETE CASCADE`).
- Profile updates (`PUT /api/v1/users/{id}`) only modify `name` and `avatar_url`.

## cURL examples

```bash
# List users
curl http://localhost:8080/api/v1/users

# Get user
curl http://localhost:8080/api/v1/users/{id}

# Update profile
curl -X PUT http://localhost:8080/api/v1/users/{id} \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice Updated","avatar_url":"https://example.com/avatar.jpg"}'

# Update role
curl -X PUT http://localhost:8080/api/v1/users/{id}/role \
  -H 'Content-Type: application/json' \
  -d '{"role":"admin"}'

# Delete user
curl -X DELETE http://localhost:8080/api/v1/users/{id}
```
