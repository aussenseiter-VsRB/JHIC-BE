---
name: PKL Domain Documentation
relation: RULES.md → modules/pkl/
description: Documentation for the pkl domain — student PKL request submission and sequential guru approval workflow
type: editable
---

# PKL Domain

## Overview

The `pkl` domain handles student internship (PKL) request submissions and the sequential approval workflow. A request must be approved by four guru roles in order: `wali_kelas` → `bk` → `kesiswaan` → `kaprog`. Approvers are resolved and snapshotted onto the request at creation time. Admins can view all requests but cannot approve; each guru only sees requests where they are an approver.

## Entity

```go
type PklRequest struct {
    ID           string    `json:"id"`
    RequesterID  string    `json:"requester_id"`
    Company      string    `json:"company"`
    Location     string    `json:"location"`
    StartDate    time.Time `json:"start_date"`
    EndDate      time.Time `json:"end_date"`
    Description  string    `json:"description"`
    Status       string    `json:"status"`
    CancelReason string    `json:"cancel_reason,omitempty"`
    CurrentStep  int       `json:"current_step"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    Steps        []Step    `json:"steps,omitempty"`
}

type Step struct {
    ID         string     `json:"id"`
    RequestID  string     `json:"request_id"`
    Position   string     `json:"position"`
    ApproverID string     `json:"approver_id"`
    Status     string     `json:"status"`
    Note       string     `json:"note,omitempty"`
    Sequence   int        `json:"sequence"`
    DecidedAt  *time.Time `json:"decided_at,omitempty"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
}
```

Request statuses: `pending`, `accepted`, `rejected`, `needs_further_action`, `cancelled`.
Step statuses: `pending`, `approved`, `rejected`, `needs_further_action`.

## Endpoints

All endpoints require a `Bearer` token.

| Method | Path | Role | Description |
|---|---|---|---|
| POST | /api/v1/approval/pkl | user | Create a PKL request |
| GET | /api/v1/approval/pkl | user/guru/admin | List requests (scoped by role) |
| GET | /api/v1/approval/pkl/{id} | user/guru/admin | Get a single request (scoped) |
| POST | /api/v1/approval/pkl/{id}/decide | guru | Approve / reject / request further action |
| DELETE | /api/v1/approval/pkl/{id} | user | Cancel own request (reason required) |

## Data flow

```
POST /api/v1/approval/pkl
  → handler.Create: decode JSON → validate dates (YYYY-MM-DD) and required fields
    → service.Create: load requester → resolve 4 approvers via FindByPosition
      (wali_kelas by class, kaprog by jurusan, bk/kesiswaan school-wide)
      → error "no {position} assigned for this request" if unresolvable
      → build request + 4 steps (approver snapshot) → pg.CreateRequest
        (tx: INSERT pkl_requests + INSERT 4 pkl_approval_steps)

GET /api/v1/approval/pkl
  → handler.List → service.List: admin → ListAll; guru → ListForApprover;
    user → ListByRequester → attach steps to each request

GET /api/v1/approval/pkl/{id}
  → handler.Get → service.Get: load request + steps → role-based access check
    (user: requester only; guru: approver on request only; admin: any)

POST /api/v1/approval/pkl/{id}/decide
  → handler.Decide → service.Decide: load request+steps →
    pending: current guru at CurrentStep decides; needs_further_action: only the
    guru whose step is flagged may resolve → approve advances CurrentStep,
    reject sets status=rejected, needs_further_action freezes the request
    → pg.Decide (tx with optimistic guards on request status and step status)

DELETE /api/v1/approval/pkl/{id}
  → handler.Cancel → service.Cancel: verify requester + reason + status
    (pending or needs_further_action only) → pg.Cancel
```

## Rules

- Approval order is fixed: `wali_kelas` → `bk` → `kesiswaan` → `kaprog`.
- Approvers are resolved and snapshotted when the request is created; later user edits do not affect existing requests.
- `wali_kelas` is matched by the requester's `class` (e.g. `PPLG 1`), `kaprog` by the requester's `jurusan` (`PPLG`/`AK`/`HTL`), `bk` and `kesiswaan` are school-wide (single user each, enforced by partial unique indexes).
- A guru may only decide their own step, in sequence. Out-of-turn or foreign requests return 403.
- `needs_further_action` freezes the request; only the guru who raised it may later resolve it to `approve` (resumes the next step) or `reject`.
- `rejected` closes the request; the student may submit a new one.
- Cancellation is requester-only, requires a reason, and is allowed only from `pending` or `needs_further_action`.
- Admins can view every request but never approve (`/decide` is guru-only).
- Concurrent decisions and cancels are guarded by optimistic checks; conflicts return 409.

## cURL examples

```bash
# Create a request (student)
curl -X POST http://localhost:8080/api/v1/approval/pkl \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"company":"PT Maju","location":"Jl. Merdeka 1","start_date":"2026-08-03","end_date":"2026-08-28","description":"PKL di bagian IT"}'

# List my requests (student)
curl http://localhost:8080/api/v1/approval/pkl \
  -H 'Authorization: Bearer <token>'

# Get request by ID
curl http://localhost:8080/api/v1/approval/pkl/{id} \
  -H 'Authorization: Bearer <token>'

# Decide (guru): approve / reject / needs_further_action
curl -X POST http://localhost:8080/api/v1/approval/pkl/{id}/decide \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"decision":"approve","note":"Setuju"}'

# Cancel own request (student, reason required)
curl -X DELETE http://localhost:8080/api/v1/approval/pkl/{id} \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"reason":"Salah perusahaan"}'
```
