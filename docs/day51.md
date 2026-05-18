# Day 51 — HTTP Chatbot Endpoint + Frontend-Ready JSON

## Objective
Menutup fase fitur inti project dengan membuat endpoint HTTP chatbot SOP yang:
- bisa dipanggil frontend atau client lain
- menerima pertanyaan user
- mengembalikan jawaban grounded
- menyertakan citations, sources, confidence, dan fallback flag
- tetap backend-only

---

## Why This Day Matters
Sampai Day 50, sistem SOP assistant masih dominan berbentuk CLI flow.
Day 51 mengubahnya menjadi bentuk yang benar-benar realistis untuk integrasi aplikasi internal:
- endpoint HTTP stabil
- response JSON jelas
- health endpoint tersedia
- logic answer dipakai ulang oleh CLI dan HTTP layer

Ini penting supaya portofolio punya ujung yang jelas sebagai backend product surface, bukan hanya eksperimen internal.

---

## Scope
Day 51 mencakup:
- service layer `chat.Service`
- refactor CLI answer agar memakai service yang sama
- HTTP router untuk knowledge API
- `GET /health`
- `POST /v1/chat/sop`
- request validation
- success/error JSON envelope
- script run knowledge API
- dokumentasi kontrak request/response

**Day 51 belum mencakup:**
- frontend UI
- auth khusus chatbot endpoint
- streaming response
- session / chat history
- gateway proxy integration

---

## Endpoint

### Health
`GET /health`

**Response:**
```json
{
  "data": {
    "status": "ok",
    "service": "knowledge"
  },
  "request_id": "..."
}
```

### SOP Chat
`POST /v1/chat/sop`

**Request body:**
```json
{
  "question": "apakah amoxicillin bisa dijual tanpa resep?",
  "top_k": 5,
  "min_score": 0.45
}
```

**Success response:**
```json
{
  "data": {
    "question": "apakah amoxicillin bisa dijual tanpa resep?",
    "answer": "Amoxicillin tidak boleh dijual bebas tanpa verifikasi resep yang valid [S1].",
    "fallback": false,
    "citations": ["[S1]"],
    "sources": [
      {
        "ref": "[S1]",
        "title": "SOP Penjualan Antibiotik Amoxicillin 500mg",
        "category": "antibiotic-dispensation",
        "source_key": "antibiotic-amoxicillin.md",
        "heading": "Aturan Dasar",
        "score": 0.7474
      }
    ],
    "confidence": {
      "top_score": 0.7474,
      "min_top_score": 0.62,
      "retrieved_count": 2,
      "used_source_count": 1
    }
  },
  "request_id": "..."
}
```

**Validation error response:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "details": {
      "field": "question",
      "reason": "is required"
    },
    "request_id": "..."
  }
}
```

**Runtime error response:**
```json
{
  "error": {
    "code": "KNOWLEDGE_CHAT_ERROR",
    "message": "failed to generate SOP answer",
    "request_id": "..."
  }
}
```

---

## Files Changed
```text
services/knowledge/internal/config/config.go
services/knowledge/internal/config/config_test.go
services/knowledge/internal/chat/service.go
services/knowledge/internal/observability/metrics.go
services/knowledge/internal/httpapi/middleware/request_id.go
services/knowledge/internal/httpapi/middleware/request_logger.go
services/knowledge/internal/httpapi/handler/response.go
services/knowledge/internal/httpapi/handler/health.go
services/knowledge/internal/httpapi/handler/chat.go
services/knowledge/internal/httpapi/router.go
services/knowledge/cmd/answer/main.go
services/knowledge/cmd/api/main.go
scripts/run-knowledge-api.ps1
scripts/run-knowledge-api.sh
docs/day51.md
```

---

## Run

### Run tests
```powershell
go test ./services/knowledge/... -count=1 -v
```

### Start API
```powershell
.\scripts\run-knowledge-api.ps1
```
*Default port:* `8084`

### Health check
```powershell
Invoke-RestMethod -Method Get -Uri "http://localhost:8084/health"
```

### Example request
```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8084/v1/chat/sop" `
  -ContentType "application/json" `
  -Body (@{
    question = "apa edukasi minimal untuk paracetamol otc?"
    top_k = 5
    min_score = 0.45
  } | ConvertTo-Json)
```

---

## Why Backend-Only Is Enough

Day 51 sengaja berhenti di endpoint HTTP dan JSON contract.

Ini cukup kuat untuk portfolio backend karena yang ditunjukkan adalah:
- API design
- service orchestration
- grounded answer backend flow
- stable response schema
- health endpoint
- integration readiness untuk frontend tanpa harus membangun frontend