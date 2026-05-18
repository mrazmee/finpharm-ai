# Day 52 — Frontend-Ready JSON Contract + Gateway Integration

## Objective
Menstandarkan chatbot SOP agar siap dipakai consumer/frontend melalui **gateway**, tanpa memindahkan business logic dari service `knowledge`.

---

## Why this day matters

Sebelum Day 52:
- chatbot SOP hanya tersedia langsung dari knowledge service
- consumer/client harus mengetahui alamat service `knowledge`
- standardisasi ingress microservices belum lengkap

Setelah Day 52:
- chatbot SOP tersedia lewat **gateway**
- request/response JSON tetap stabil
- auth dan role check tetap berada di gateway
- knowledge service tetap menjadi domain owner untuk logic retrieval + answer synthesis

Ini membuat arsitektur lebih konsisten dengan service lain di project:
- gateway = ingress, auth, validation edge, observability
- service domain = business logic

---

## Scope

Day 52 mencakup:
- penambahan `KNOWLEDGE_BASE_URL` di config gateway
- penambahan proxy route gateway untuk chatbot SOP
- schema request/response tetap stabil
- error response konsisten di edge/gateway
- contoh payload untuk frontend / Postman
- verifikasi bahwa identity, request ID, dan trace ID terpropagasi sampai knowledge service

**Day 52 tidak mencakup:**
- frontend UI
- perubahan besar pada logic RAG
- auth baru di knowledge service
- session/chat history
- streaming response

---

## Design Decision

- `knowledge service` tetap domain owner untuk chatbot SOP
- `gateway` hanya bertugas untuk:
  - JWT auth
  - role check
  - request validation di edge
  - proxy request ke knowledge service
  - pass-through response upstream

Dengan begitu, Day 52 tetap sesuai prinsip project:
- gateway tidak mengambil business logic chatbot
- knowledge service tetap fokus pada retrieval dan answer synthesis

---

## Gateway Route

### Public consumer-facing endpoint
`POST /v1/chat/sop`

Route ini tersedia di gateway dan diproteksi oleh:
- `JWTAuth`
- `RequireRoles("staff", "supervisor")`

Jadi chatbot SOP sekarang mengikuti pola route client-facing yang sama seperti:
- `/v1/medicines`
- `/v1/stock/check`
- `/v1/transactions`

---

## Request Contract

### Request body
```json
{
  "question": "apakah amoxicillin bisa dijual tanpa resep?",
  "top_k": 5,
  "min_score": 0.45
}
```

### Field Explanation
- **question:** pertanyaan user (wajib diisi)
- **top_k:** jumlah maksimum chunk yang diminta ke knowledge flow (harus > 0)
- **min_score:** minimum similarity score (harus antara 0 dan 1)

### Success Response Contract
```json
{
  "data": {
    "question": "apakah amoxicillin bisa dijual tanpa resep?",
    "answer": "Amoxicillin tidak boleh dijual bebas tanpa verifikasi resep yang valid [S1]. Staff wajib menahan transaksi dan meminta review apoteker atau supervisor bila tidak ada resep [S2].",
    "fallback": false,
    "citations": [
      "[S1]",
      "[S2]"
    ],
    "sources": [
      {
        "ref": "[S1]",
        "title": "SOP Penjualan Antibiotik Amoxicillin 500mg",
        "category": "antibiotic-dispensation",
        "source_key": "antibiotic-amoxicillin.md",
        "heading": "Aturan Dasar",
        "score": 0.7473900248103168
      },
      {
        "ref": "[S2]",
        "title": "SOP Penjualan Antibiotik Amoxicillin 500mg",
        "category": "antibiotic-dispensation",
        "source_key": "antibiotic-amoxicillin.md",
        "heading": "Kondisi yang Mengharuskan Penolakan Sementara",
        "score": 0.7237687523953019
      }
    ],
    "confidence": {
      "top_score": 0.7473900248103168,
      "min_top_score": 0.62,
      "retrieved_count": 2,
      "used_source_count": 2
    }
  },
  "request_id": "2097b7d9d775cd2b7f2c9a539b0408aa"
}
```

### Fallback Response Contract
```json
{
  "data": {
    "question": "berapa gaji apoteker?",
    "answer": "Saya belum menemukan dasar SOP yang cukup untuk menjawab pertanyaan ini.",
    "fallback": true,
    "citations": [],
    "sources": [],
    "confidence": {
      "top_score": 0.5767629259755432,
      "min_top_score": 0.62,
      "retrieved_count": 0,
      "used_source_count": 0
    }
  },
  "request_id": "a3518afc9f6948a8b6c54fbb67b6ff8b"
}
```

### Error Response Contract

**Validation error**
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

**Upstream error**
```json
{
  "error": {
    "code": "UPSTREAM_ERROR",
    "message": "knowledge service unreachable",
    "details": "...",
    "request_id": "..."
  }
}
```

**Gateway internal error**
```json
{
  "error": {
    "code": "GATEWAY_ERROR",
    "message": "failed to encode request",
    "request_id": "..."
  }
}
```

---

## Config Added

Gateway sekarang memiliki config tambahan:
- `KNOWLEDGE_BASE_URL` (default local: `http://localhost:8084`)

Config ini divalidasi seperti base URL service lain:
- `INVENTORY_BASE_URL`
- `TRANSACTION_BASE_URL`

Jadi contract config gateway tetap konsisten.

---

## Implementation Summary

Day 52 menambahkan:

### 1. Gateway config support
Gateway sekarang mengenal `KnowledgeBaseURL`.

### 2. Proxy handler baru
Handler baru: `KnowledgeProxyHandler`
Tugasnya:
- validasi request body di edge
- membuat request ke upstream knowledge service
- meneruskan header yang relevan melalui `setProxyForwardHeaders(...)`
- pass-through response upstream ke client

### 3. Gateway route baru
Route baru: `POST /v1/chat/sop`
Dengan auth aktif: hanya role staff dan supervisor yang boleh mengakses.

### 4. Startup log gateway diperbarui
Gateway sekarang juga mencatat: `knowledge_base_url`.
Ini membantu observability saat startup.

---

## Verification Summary

### Unit / Integration Test
Seluruh test gateway lulus, termasuk:
- config validation untuk `KnowledgeBaseURL`
- auth/router existing behavior
- `KnowledgeProxyHandler` success
- `KnowledgeProxyHandler` validation error
- `KnowledgeProxyHandler` upstream error

### Runtime Verification
Verifikasi runtime menunjukkan:
- token debug bisa dibuat lewat gateway
- request `POST /v1/chat/sop` lewat gateway berhasil
- query SOP mengembalikan `fallback=false`
- query non-SOP mengembalikan `fallback=true`
- `request_id` sama di gateway dan knowledge service
- `trace_id` sama di gateway dan knowledge service
- `user_id` dan role terpropagasi dari gateway ke knowledge service

Ini menunjukkan gateway integration benar-benar bekerja, bukan hanya lolos unit test.

---

## Example Commands

### 1. Jalankan knowledge API
```powershell
.\scripts\run-knowledge-api.ps1
```

### 2. Set env gateway
```powershell
$env:KNOWLEDGE_BASE_URL="http://localhost:8084"
```

### 3. Jalankan gateway
```powershell
go run ./services/gateway/cmd/api
```

### 4. Ambil token debug
```powershell
$token = (Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/v1/auth/token" `
  -ContentType "application/json" `
  -Body (@{
    user_id = "staff-001"
    role = "staff"
  } | ConvertTo-Json)
).data.access_token
```

### 5. Query SOP via gateway
```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/v1/chat/sop" `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
  } `
  -Body (@{
    question = "apakah amoxicillin bisa dijual tanpa resep?"
    top_k = 5
    min_score = 0.45
  } | ConvertTo-Json) | ConvertTo-Json -Depth 10
```

### 6. Query fallback via gateway
```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri "http://localhost:8080/v1/chat/sop" `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
  } `
  -Body (@{
    question = "berapa gaji apoteker?"
    top_k = 5
    min_score = 0.45
  } | ConvertTo-Json) | ConvertTo-Json -Depth 10
```

---

## What Day 52 Demonstrates

Day 52 menunjukkan bahwa project ini tidak berhenti di CLI internal, tetapi sudah sampai ke bentuk backend product surface yang siap dikonsumsi client:
- consumer-facing chatbot endpoint tersedia lewat gateway
- auth dan role enforcement tetap di ingress
- service knowledge tetap domain owner
- JSON contract stabil
- observability dan identity propagation tetap terjaga
- fallback behavior tetap jujur untuk query di luar SOP

---

## Final Status
**Day 52 complete.**

Setelah Day 52, project ini sudah punya bentuk yang kuat sebagai:
- backend microservices portfolio
- observability-aware system
- internal SOP assistant API
- gateway-integrated AI feature