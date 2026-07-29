---
name: Auth Domain Documentation
relation: RULES.md → modules/auth/
description: Documentation for the auth domain — user registration, login, and session management
type: editable
---

# Auth Domain

## Overview

The `auth` domain handles user identity and session management. It provides registration and login with bcrypt-hashed passwords and opaque bearer tokens stored in a sessions table.

## Entity

```go
type User struct {
    ID           string    `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"` // never serialized
    Name         string    `json:"name"`
    Role         string    `json:"role"`
    AvatarURL    string    `json:"avatar_url,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

## Endpoints

| Method | Path | Handler | Description |
|---|---|---|---|
| POST | /api/v1/auth/register | RegisterUser | Create account + return token |
| POST | /api/v1/auth/login | Login | Authenticate + return token |
| POST | /api/v1/auth/logout | Logout | Invalidate all sessions for user |

## Data flow

```
POST /api/v1/auth/register
  → handler.RegisterUser: decode JSON → validate email+password → call service.Register
    → service.Register: check duplicate email → bcrypt hash → create user → generate token → create session
      → usersRepo.Create: INSERT INTO users (id, email, password_hash, ...)
      → sessionsRepo.Create: INSERT INTO sessions (token, user_id, expires_at)

POST /api/v1/auth/login
  → handler.Login: decode JSON → validate → call service.Login
    → service.Login: lookup by email → bcrypt compare → generate token → create session
      → sessionsRepo.Create: INSERT INTO sessions (...)

POST /api/v1/auth/logout
  → handler.Logout: extract Bearer token → validate → delete all sessions for user
```

## Rules

- Email must be unique. Service returns an error on duplicate.
- Passwords are hashed with bcrypt (never stored in plaintext).
- New users default to role `user`. Roles are managed by the user domain.
- `Session` entity (in `entity.go`) stores opaque bearer tokens linked to users.
- Tokens are 64-char hex strings, valid for 72 hours.
- Tokens are opaque — no claims, no signing. Validated via DB lookup.
- Logout invalidates all sessions for the user (not just the current token).

## cURL examples

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"supersecret","name":"Alice"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"supersecret"}'

# Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H 'Authorization: Bearer <token>'

# Health
curl http://localhost:8080/api/v1/health
```
