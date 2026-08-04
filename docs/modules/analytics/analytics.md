---
name: Analytics Domain Documentation
relation: RULES.md → modules/analytics/
description: Privacy-safe aggregate product and Berita engagement analytics
type: editable
---

# Analytics Domain

## Overview

Analytics records product events in PostgreSQL without raw user content. Session identifiers are SHA-256 hashed before storage. Analytics writes are best-effort: an insertion failure never fails the user-facing request.

## Endpoints

| Method | Path | Access | Description |
|---|---|---|---|
| GET | /api/v1/analytics/summary?days=30 | admin | Aggregate chat and Nexxa-Match event counts |
| GET | /api/v1/analytics/berita/summary?days=30 | jurnal, admin | Aggregate Berita engagement event counts |

`days` defaults to 30 and accepts values from 1 through 90.

## Events

| Event | Stored properties |
|---|---|
| `chat.request` | message length, optional topic, success |
| `match.completed` | success, recommended major, three percentages |
| `berita.view` | article ID |
| `berita.read_50` | article ID |
| `berita.read_90` | article ID |
| `berita.share` | article ID |
| `berita.link_click` | article ID |

Raw chat messages, Nexxa-Match answers/reasons, Berita text/titles, IP addresses, and external URLs are not stored.
