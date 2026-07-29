---
name: Workspace Domain Documentation
relation: RULES.md → modules/workspace/
description: Documentation for the workspace domain — organizational units owned by users
type: editable
---

# Workspace Domain

## Overview

The `workspace` domain represents organizational units. Each workspace is owned by a user and serves as the parent container for pipelines.

## Entity

```go
type Workspace struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    OwnerID   string    `json:"owner_id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

## Endpoints

| Method | Path | Handler | Description |
|---|---|---|---|
| GET | /api/v1/workspaces?owner_id={id} | List | List workspaces by owner |
| POST | /api/v1/workspaces | Create | Create a new workspace |
| GET | /api/v1/workspaces/{id} | Get | Get workspace by ID |
| PUT | /api/v1/workspaces/{id} | Update | Update workspace |
| DELETE | /api/v1/workspaces/{id} | Delete | Delete workspace |

> **Note:** The `List` endpoint requires the `owner_id` query parameter. This is intentional — workspaces are always scoped to a specific owner.

## Data flow

```
POST /api/v1/workspaces
  → handler.Create: decode JSON → validate name+owner_id → call service.Create
    → service.Create: generate UUID → set timestamps → call repo.Create
      → repository_pg.Create: INSERT INTO workspaces (id, name, owner_id, ...)
```

## Rules

- `owner_id` references an existing user (foreign key constraint).
- Listing requires `?owner_id=` — unauthenticated listing of all workspaces is not supported.
- Workspace deletion cascades to owned pipelines (DB foreign key `ON DELETE CASCADE`).

## cURL examples

```bash
curl -X POST http://localhost:8080/api/v1/workspaces \
  -H 'Content-Type: application/json' \
  -d '{"name":"My Project","owner_id":"<user-id>"}'

curl "http://localhost:8080/api/v1/workspaces?owner_id=<user-id>"
curl http://localhost:8080/api/v1/workspaces/{id}
```
