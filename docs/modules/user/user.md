---
name: User Domain Documentation
relation: RULES.md → modules/user/
description: Documentation for the user domain — identity and account management
type: editable
---

# User Domain

## Overview

The `user` domain manages user identities. It provides CRUD operations for user accounts and is the owner reference for workspaces.

## Entity

```go
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    AvatarURL string    `json:"avatar_url,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
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
  → handler.Create: decode JSON → validate email+name required → call service.Create
    → service.Create: check duplicate email → generate UUID → set timestamps → call repo.Create
      → repository_pg.Create: INSERT INTO users (id, email, name, avatar_url, ...)
```

## Rules

- Email must be unique. Service returns an error if a user with the same email already exists.
- `avatar_url` is optional — defaults to empty string.
- User deletion cascades to owned workspaces (DB foreign key `ON DELETE CASCADE`).

## cURL examples

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","name":"Alice"}'

curl http://localhost:8080/api/v1/users
curl http://localhost:8080/api/v1/users/{id}
```
