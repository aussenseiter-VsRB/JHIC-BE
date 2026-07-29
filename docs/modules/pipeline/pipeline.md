---
name: Pipeline Domain Documentation
relation: RULES.md → modules/pipeline/
description: Documentation for the pipeline domain — n8n workflow management
type: editable
---

# Pipeline Domain

## Overview

The `pipeline` domain manages automation workflows (n8n pipelines). Each pipeline belongs to a workspace and stores its configuration as JSONB for flexibility.

## Entity

```go
type Pipeline struct {
    ID          string          `json:"id"`
    WorkspaceID string          `json:"workspace_id"`
    Name        string          `json:"name"`
    Description string          `json:"description,omitempty"`
    Status      string          `json:"status"`
    Config      json.RawMessage `json:"config,omitempty"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
}
```

## Endpoints

| Method | Path | Handler | Description |
|---|---|---|---|
| GET | /api/v1/pipelines?workspace_id={id} | List | List pipelines by workspace |
| POST | /api/v1/pipelines | Create | Create a new pipeline |
| GET | /api/v1/pipelines/{id} | Get | Get pipeline by ID |
| PUT | /api/v1/pipelines/{id} | Update | Update pipeline (name, description, status, config) |
| DELETE | /api/v1/pipelines/{id} | Delete | Delete pipeline |

> **Note:** The `List` endpoint requires the `workspace_id` query parameter. Pipelines are always scoped to a workspace.

## Data flow

```
POST /api/v1/pipelines
  → handler.Create: decode JSON → validate workspace_id+name → call service.Create
    → service.Create: generate UUID → set defaults (status=inactive, config={}) → call repo.Create
      → repository_pg.Create: INSERT INTO pipelines (...)
```

## Rules

- `Status` is one of: `active`, `inactive`, `error` (database enum `pipeline_status`).
- `Config` is JSONB — flexible storage for arbitrary n8n workflow configuration. Defaults to `{}`.
- Listing requires `?workspace_id=` — always scoped to a workspace.
- Pipeline deletion does not cascade further (pipelines are leaf entities).

## cURL examples

```bash
curl -X POST http://localhost:8080/api/v1/pipelines \
  -H 'Content-Type: application/json' \
  -d '{"workspace_id":"<ws-id>","name":"Daily Backup","description":"Runs every night"}'

curl "http://localhost:8080/api/v1/pipelines?workspace_id=<ws-id>"

curl -X PUT http://localhost:8080/api/v1/pipelines/{id} \
  -H 'Content-Type: application/json' \
  -d '{"name":"Daily Backup","status":"active","config":{"schedule":"0 2 * * *"}}'

curl http://localhost:8080/api/v1/pipelines/{id}
curl -X DELETE http://localhost:8080/api/v1/pipelines/{id}
```
