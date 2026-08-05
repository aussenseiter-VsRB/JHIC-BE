---
name: SPMB Domain Documentation
relation: RULES.md → modules/spmb/
description: Documentation for the spmb domain — public student registration with an admin review flow
type: editable
---

# SPMB Domain

## Overview

The `spmb` domain lets a prospective student (or parent) register for admission (SPMB/PPDB) without an account, and lets an admin review and decide each registration. It pairs with the AI-assisted `nexxa/spmb` sub-domain: uploading a Kartu Keluarga (KK) image/PDF via `parse-kk` auto-fills the registration form, and `ask` answers SPMB questions through a dedicated RAG workflow.

## Entity

```go
type SpmbRegistration struct {
    ID           id.ID     `json:"id"`
    Nama         string    `json:"nama"`
    Nik          string    `json:"nik"`
    Nisn         string    `json:"nisn,omitempty"`
    KkNo         string    `json:"kk_no,omitempty"`
    TempatLahir  string    `json:"tempat_lahir,omitempty"`
    TanggalLahir string    `json:"tanggal_lahir,omitempty"`
    JenisKelamin string    `json:"jenis_kelamin"`
    Agama        string    `json:"agama,omitempty"`
    Alamat       string    `json:"alamat"`
    AsalSekolah  string    `json:"asal_sekolah,omitempty"`
    NoHP         string    `json:"no_hp,omitempty"`
    NamaAyah     string    `json:"nama_ayah,omitempty"`
    NamaIbu      string    `json:"nama_ibu,omitempty"`
    Jurusan      string    `json:"jurusan"`
    Status       string    `json:"status"`
    CancelReason string    `json:"cancel_reason,omitempty"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

Status values: `proses` (default on create), `approve`, `cancel`. Allowed transitions: `proses → approve`, `proses|approve → cancel` (cancel requires a reason). `jurusan` must be one of `PPLG`, `Akuntansi`, `Perhotelan`, `Hotel`.

## Endpoints

| Method | Path | Description | Auth | Role |
|--------|------|-------------|------|------|
| POST | `/api/v1/spmb` | Create a registration (public) | No | — |
| GET | `/api/v1/spmb` | List all registrations | Yes | admin |
| GET | `/api/v1/spmb/{id}` | Get one registration | Yes | admin |
| POST | `/api/v1/spmb/{id}/status` | Approve or cancel a registration | Yes | admin |

The AI-assist endpoints live in the `nexxa` domain (see `docs/modules/nexxa/nexxa.md`):
`POST /api/v1/nexxa/spmb/parse-kk` and `POST /api/v1/nexxa/spmb/ask`.

## Data flow

```
POST /api/v1/spmb {nama, nik, jenis_kelamin, alamat, jurusan, ...}
  → spmb.Handler.Create
    → spmb.Service.Create: validate (nama, 16-digit nik, jenis_kelamin, alamat, jurusan)
      → id.New(), status=proses, timestamps
      → spmb.Repository.Create → INSERT spmb_registrations
    → 201 {registration} or 400 on validation error

POST /api/v1/spmb/{id}/status {status: approve|cancel, reason?}  (admin)
  → spmb.Service.SetStatus: ByID → transition rule check → UpdateStatus (optimistic, expected=current)
    → 200 {registration}, 409 on stale/invalid state
```

## Rules

- Create is public; a prospective student needs no account. All read/status endpoints require `admin`.
- `nik` must be exactly 16 digits.
- Cancel requires `reason`; approve must come from `proses`.
- The list is ordered by `created_at DESC` (no pagination yet).

## Examples

```bash
# Public registration
curl -X POST http://localhost:8080/api/v1/spmb \
  -H 'Content-Type: application/json' \
  -d '{"nama":"Budi Santoso","nik":"3204123456789012","jenis_kelamin":"Laki-laki","alamat":"Jl. Raya Soreang No. 1","jurusan":"PPLG"}'

# Admin approve (Bearer token from login)
curl -X POST http://localhost:8080/api/v1/spmb/42/status \
  -H 'Authorization: Bearer <admin-token>' -H 'Content-Type: application/json' \
  -d '{"status":"approve"}'

# Admin cancel
curl -X POST http://localhost:8080/api/v1/spmb/42/status \
  -H 'Authorization: Bearer <admin-token>' -H 'Content-Type: application/json' \
  -d '{"status":"cancel","reason":"berkas tidak lengkap"}'
```
