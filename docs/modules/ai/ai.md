---
name: AI Domain Documentation
relation: RULES.md → modules/ai/
description: Documentation for the ai domain — validation and server-side proxying of n8n webhooks (chatbot and Nexxa-Match)
type: editable
---

# AI Domain

## Overview

The `ai` domain exposes n8n's AI webhooks through the JHIC-BE backend so the frontend never talks to n8n directly. It validates incoming payloads, forwards them server-side to the n8n webhook, and relays the response as-is. Two endpoints are public (no auth) and rate-limited per client IP. This domain has **no database** — the `N8NClient` interface (defined in `client.go`) plays the role a repository would, and its implementation lives in `internal/infrastructure/n8n/`.

## Entity

```go
type ChatRequest struct {
    ChatInput string `json:"chatInput"`
    SessionID string `json:"sessionId"`
}

type ChatResponse struct {
    Output string `json:"output"`
}

type NexxaRequest struct {
    Jawaban1 string `json:"jawaban_1"`
    // ... Jawaban2 .. Jawaban8
}

type NexxaResponse struct {
    NamaJurusan string `json:"nama_jurusan"`
    Alasan      string `json:"alasan"`
}
```

## Endpoints

Both endpoints are public, rate-limited (default 10 req/min/IP), and capped at 32KB bodies.

| Method | Path | Description |
|---|---|---|
| POST | /api/v1/ai/chat | Validate + forward a chatbot message to n8n, relay `{output}` |
| POST | /api/v1/ai/nexxa-match | Validate 8 answers + forward to n8n, relay `{nama_jurusan, alasan}` |

## Data flow

```
POST /api/v1/ai/chat {chatInput, sessionId}
  → middleware.RateLimit → handler.Chat: MaxBytesReader + JSON decode
    → service.Chat: trim + require non-empty + ≤300 chars; generate sessionId if empty
      → n8n.Client.Chat: POST {chatInput, sessionId} to N8N_CHAT_PATH
        (Basic Auth header from N8N_CHAT_USERNAME/PASSWORD)
        → relay {output} as-is, or 502/504 on upstream failure

POST /api/v1/ai/nexxa-match {jawaban_1..8}
  → middleware.RateLimit → handler.NexxaMatch: MaxBytesReader + JSON decode
    → service.NexxaMatch: require exactly 8 answers, each non-empty and ≤500 chars
      → n8n.Client.NexxaMatch: POST jawaban_1..8 to N8N_NEXXA_PATH
        (X-JHIC-Secret header from N8N_NEXXA_SECRET)
        → relay {nama_jurusan, alasan} as-is, or 502/504 on upstream failure
```

The service propagates the request context into the upstream call, so a browser disconnect cancels the n8n execution mid-flight. The upstream client times out via `N8N_TIMEOUT` (default 115s); the server `WriteTimeout` is 120s so slow LLM responses are not cut off.

## Rules

- Input validation is business logic in the service, not the handler.
- `chatInput` is trimmed before length checks and forwarding; `sessionId` is auto-generated (32 hex chars) when absent.
- Nexxa requires exactly 8 answers; answers are trimmed and forwarded normalized.
- Upstream HTTP errors map to `502 Bad Gateway`; timeouts to `504 Gateway Timeout`. Neither leaks upstream response bodies.
- The n8n chat webhook is authenticated with Basic Auth; the Nexxa webhook with the `X-JHIC-Secret` header — both values come from env and are sent only when configured.
- Rate limiting is a token bucket keyed by client IP with no background goroutine (opportunistic cleanup only).

## cURL examples

```bash
# Chatbot
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H 'Content-Type: application/json' \
  -d '{"chatInput":"Bagaimana cara mendaftar PPDB?","sessionId":"123e4567-e89b-12d3-a456-426614174000"}'

# Nexxa-Match
curl -X POST http://localhost:8080/api/v1/ai/nexxa-match \
  -H 'Content-Type: application/json' \
  -d '{"jawaban_1":"a","jawaban_2":"b","jawaban_3":"c","jawaban_4":"d","jawaban_5":"e","jawaban_6":"f","jawaban_7":"g","jawaban_8":"h"}'
```
