---
name: Berita Domain Documentation
relation: RULES.md → modules/berita/
description: Documentation for the berita (news) domain — article CRUD and image upload
type: editable
---

# Berita Domain

## Overview

The `berita` domain handles news article CRUD and image management. Articles are authored by users with the `jurnal` role. Images are uploaded to S3-compatible storage (Backblaze B2) and served via presigned URLs.

## Entity

```go
type Berita struct {
    ID        string    `json:"id"`
    AuthorID  string    `json:"author_id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    ImageURL  string    `json:"image_url,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

## Endpoints

All berita endpoints require authentication (`Bearer` token) and the `jurnal` role.

| Method | Path | Handler | Description |
|---|---|---|---|
| GET | /api/v1/berita | List | List all news articles |
| GET | /api/v1/berita/{id} | Get | Get article by ID |
| POST | /api/v1/berita | Create | Create a new article |
| PUT | /api/v1/berita/{id} | Update | Update article (author only) |
| DELETE | /api/v1/berita/{id} | Delete | Delete article + its image (author only) |
| POST | /api/v1/berita/{id}/image | UploadImage | Upload/replace article image |

## Data flow

```
GET /api/v1/berita
  → handler.List → service.List → pg.List (SELECT ... ORDER BY created_at DESC)
  → handler signs each image_url via PresignGet (24h TTL)

POST /api/v1/berita
  → handler.Create: decode JSON → validate title+content → call service.Create
    → service.Create: generate ID → create record
      → pg.Create: INSERT INTO berita (id, author_id, title, content, ...)
  → handler signs image_url before response

POST /api/v1/berita/{id}/image
  → handler.UploadImage: parse multipart form → validate MIME type (jpeg/png/gif/webp)
    → enforce 5 MB max via MaxBytesReader
    → generate UUID filename → upload to storage via store.Upload
    → call service.SetImage → pg.Update (set image_url)
  → handler signs image_url before response

DELETE /api/v1/berita/{id}
  → handler.Delete: fetch article → if has image, delete from storage → call service.Delete
    → service.Delete: verify caller is author → pg.Delete
```

## Rules

- All endpoints require `jurnal` role (auth + role middleware applied in `Register`).
- Only the article author can update or delete their own articles (`forbidden: not the author`).
- Image upload is limited to 5 MB, accepted types: `image/jpeg`, `image/png`, `image/gif`, `image/webp`.
- Images are stored at `berita/{beritaID}/{uuid}.{ext}` in object storage.
- Image URLs returned are presigned (24-hour TTL). The handler signs them on every response.
- Deleting an article also deletes its stored image from object storage.
- `image_url` is a plain object key (not a full URL) as returned by the storage layer.

## cURL examples

```bash
# Create article
curl -X POST http://localhost:8080/api/v1/berita \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"title":"Breaking News","content":"Something happened today."}'

# List articles
curl http://localhost:8080/api/v1/berita \
  -H 'Authorization: Bearer <token>'

# Get article by ID
curl http://localhost:8080/api/v1/berita/{id} \
  -H 'Authorization: Bearer <token>'

# Update article
curl -X PUT http://localhost:8080/api/v1/berita/{id} \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"title":"Updated Headline","content":"Updated content."}'

# Upload image
curl -X POST http://localhost:8080/api/v1/berita/{id}/image \
  -H 'Authorization: Bearer <token>' \
  -F 'image=@photo.jpg'

# Delete article
curl -X DELETE http://localhost:8080/api/v1/berita/{id} \
  -H 'Authorization: Bearer <token>'
```
