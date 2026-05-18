# FinPharm-AI

FinPharm-AI adalah **portfolio backend microservices berbasis Go** yang dibangun dengan pendekatan *learning by doing* untuk menunjukkan fondasi yang relevan dengan target role backend engineer:

- Go / Golang
- Microservices architecture
- REST API
- PostgreSQL / SQL / sqlx
- Idempotency
- Event-driven processing
- JWT auth & authorization
- Observability
- Alerting
- AI service integration
- RAG / knowledge retrieval
- Gateway-integrated SOP chatbot

Project ini dibangun **bertahap per hari** agar proses belajar, keputusan desain, dan progres teknis bisa dibaca dengan jelas dari dokumentasi `docs/dayXX.md`.

---

## Final Status

Posisi repo saat ini:

### Core backend services selesai
- Gateway Service
- Transaction Service
- Inventory Service
- AI Auditor Service
- Worker Service
- Knowledge Service

### Persistence & orchestration sudah hidup
- Inventory Service memakai **PostgreSQL + sqlx**
- Transaction Service memakai **PostgreSQL + sqlx**
- Knowledge Service memakai **PostgreSQL + pgvector**
- Create transaction sudah **idempotent** dengan `Idempotency-Key`
- Transaction lifecycle berjalan:
  - `PENDING`
  - `APPROVED`
  - `PENDING_REVIEW`
  - `FLAGGED`
  - `FAILED`
- Approved transaction mem-publish event ke RabbitMQ
- Worker memproses event approved secara async

### Security & runtime hardening dasar sudah hidup
- JWT issue token di gateway
- Role-based authorization dasar:
  - `staff`
  - `supervisor`
- Debug route hanya aktif di local/dev
- Basic in-memory rate limiting di gateway
- Config validation / fail-fast saat startup

### Observability & alerting sudah hidup
- Request metrics per service
- Worker metrics dasar
- Prometheus local scrape
- Grafana local dashboard
- Alert rules di Prometheus
- Alertmanager local routing
- Local webhook receiver untuk proof-of-delivery
- Audit logging + trace header propagation

### Knowledge / RAG flow sudah hidup
- SOP markdown ingestion ke knowledge database
- pgvector similarity retrieval
- grounded answer synthesis
- citation-ready response
- fallback jujur untuk pertanyaan di luar SOP
- HTTP knowledge API
- chatbot SOP tersedia lewat **gateway** di:
  - `POST /v1/chat/sop`

> **Ringkasnya:** repo ini sudah mencapai bentuk portfolio backend yang production-minded untuk local/demo: service terpisah, persistence aktif, async processing aktif, auth ada, observability & alerting ada, knowledge/RAG flow ada, dan chatbot SOP sudah tersedia sebagai endpoint consumer-facing lewat gateway.

---

## Architecture

FinPharm-AI menggunakan arsitektur microservices sederhana dengan enam komponen utama:

### 1. Gateway Service
Single entry point untuk client.

**Tugas:**
- routing / proxy
- request-id propagation
- trace header propagation
- edge validation dasar
- JWT issue token
- authorization dasar
- rate limit dasar
- logging
- consumer-facing ingress untuk chatbot SOP

**Gateway tidak menyimpan business logic transaksi atau logic RAG.**

### 2. Transaction Service
Orchestration service untuk:
- stock check
- create transaction
- idempotency
- transaction lifecycle
- pemanggilan AI Auditor
- pemanggilan Inventory Service untuk deduct stock
- publish approved transaction event

### 3. Inventory Service
Source of truth untuk:
- medicines catalog
- stock availability
- stock deduction

### 4. AI Auditor Service
Service audit transaksi untuk:
- menilai request transaksi
- memberi decision: `APPROVED` atau `REVIEW`
- fallback ke provider aman saat provider utama gagal

### 5. Worker Service
Consumer async untuk:
- consume approved transaction event dari RabbitMQ
- dummy notification
- dummy report output
- worker metrics dasar

### 6. Knowledge Service
Service knowledge / RAG untuk SOP farmasi:
- SOP ingestion
- chunking
- embedding
- vector similarity retrieval
- grounded answer synthesis
- citation-ready response
- fallback jujur untuk pertanyaan di luar SOP

---

## Main Flows

### Medicines List / Detail
`Client -> Gateway -> Inventory`

### Stock Check
`Client -> Gateway -> Transaction -> Inventory`

### Create Transaction
`Client -> Gateway -> Transaction -> Inventory + AI Auditor`

**Flow create transaction saat ini:**
1. Client kirim `POST /v1/transactions` ke Gateway
2. Gateway validasi request dasar + `Idempotency-Key`
3. Gateway validasi auth + rate limit
4. Transaction Service cek replay by `idempotency_key`
5. Transaction Service pre-check stock ke Inventory
6. Transaction disimpan sebagai `PENDING`
7. Transaction Service memanggil AI Auditor
8. Jika audit result `REVIEW`:
   - status menjadi `PENDING_REVIEW`, atau
   - `FLAGGED` jika risk tinggi
9. Jika audit result `APPROVED`:
   - Inventory deduct stock
   - transaction diupdate jadi `APPROVED`
10. Transaction Service publish event `transaction.approved`
11. Worker consume event approved dan memproses output async

### SOP Chatbot
`Client -> Gateway -> Knowledge`

**Flow chatbot SOP saat ini:**
1. Client kirim `POST /v1/chat/sop` ke Gateway
2. Gateway validasi JWT + role
3. Gateway validasi request dasar
4. Gateway proxy request ke Knowledge Service
5. Knowledge Service:
   - embed query
   - retrieval ke pgvector
   - filter & diversify results
   - grounded answer synthesis
   - citation normalization
   - fallback bila konteks tidak cukup
6. Response JSON kembali ke client lewat Gateway dengan:
   - `answer`
   - `fallback`
   - `citations`
   - `sources`
   - `confidence`

---

## Repository Structure

```text
finpharm-ai/
├─ .github/
│  └─ workflows/
├─ cmd/
│  └─ alert-webhook/
├─ docs/
├─ internal/
│  └─ telemetry/
├─ knowledge/
│  └─ sop/
├─ observability/
│  ├─ prometheus/
│  ├─ grafana/
│  └─ alertmanager/
├─ scripts/
│  ├─ postgres/
│  ├─ *.ps1
│  ├─ *.sh
│  └─ reset_transaction_data.go
├─ services/
│  ├─ gateway/
│  ├─ inventory/
│  ├─ transaction/
│  ├─ ai-auditor/
│  ├─ worker/
│  └─ knowledge/
├─ docker-compose.yml
├─ docker-compose.rabbitmq.yml
├─ docker-compose.prometheus.yml
├─ docker-compose.grafana.yml
├─ docker-compose.alertmanager.yml
├─ Makefile
├─ RUNBOOK.md
├─ README.md
├─ go.mod
└─ go.sum
```

---

## Services & Ports

| Component | Port | Description |
| :--- | :--- | :--- |
| **Gateway** | `8080` | Edge service untuk routing, auth, rate limit, request-id, trace, logging, dan chatbot ingress |
| **Transaction** | `8081` | Orchestration service untuk stock check, create tx, idempotency, audit result, event publish |
| **Inventory** | `8082` | Source of truth untuk medicines dan stock |
| **AI Auditor** | `8083` | Audit service untuk decision transaksi |
| **Knowledge** | `8084` | SOP retrieval, grounded answer synthesis, dan knowledge API |
| **Worker Metrics** | `9094` | Metrics endpoint worker |
| **Prometheus** | `9090` | Local metrics aggregation |
| **Alertmanager** | `9093` | Local alert routing |
| **Grafana** | `3000` | Dashboard observability |
| **RabbitMQ UI** | `15672` | RabbitMQ management UI |
| **Alert Webhook** | `18080` | Local webhook receiver untuk verifikasi alert |
| **PostgreSQL** | `55432` / `5432` | Local dev database (host/container) |

---

## Quick Start (PowerShell / Windows)

**1. Jalankan PostgreSQL**
```powershell
docker compose up -d postgres
```

**2. Jalankan RabbitMQ**
```powershell
.\scripts\rabbitmq-up.ps1
```

**3. Jalankan migrations domain utama**
```powershell
.\scripts\migrate-inventory-up.ps1
.\scripts\migrate-transaction-up.ps1
```

**4. Jalankan knowledge migration**
```powershell
.\scripts\run-knowledge-migrate.ps1
```

**5. Ingest SOP ke knowledge database**
```powershell
.\scripts\run-knowledge-ingest.ps1
```

**6. Jalankan services (buka terminal terpisah)**
```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-knowledge-api.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

**7. Jalankan observability**
```powershell
.\scripts\run-prometheus.ps1
.\scripts\run-grafana.ps1
```

**8. Jalankan alerting local**
```powershell
.\scripts\run-alertmanager.ps1
.\scripts\run-alert-webhook.ps1
```

**9. Cek readiness demo**
```powershell
.\scripts\demo-readiness.ps1
```

---

## Quick Start (Makefile)

Kalau environment kamu punya `make`:

```makefile
make help
make run-rabbitmq
make run-prometheus
make run-grafana
make run-alertmanager
make run-alert-webhook
make demo-readiness
```

*Catatan:*
- Di repo ini `Makefile` tetap memanggil PowerShell.
- Jalur utama di Windows tetap `scripts/*.ps1`.

---

## Environment Variables

### Common
- `APP_ENV` — default: `local`
- `PORT`
- `READ_TIMEOUT_MS` — default: `5000`
- `WRITE_TIMEOUT_MS` — default: `5000`
- `IDLE_TIMEOUT_MS` — default: `30000`
- `SHUTDOWN_TIMEOUT_MS`

### Gateway
- `TRANSACTION_BASE_URL` — default: `http://localhost:8081`
- `INVENTORY_BASE_URL` — default: `http://localhost:8082`
- `KNOWLEDGE_BASE_URL` — default: `http://localhost:8084`
- `AUTH_ENABLED` — default: `true`
- `JWT_SECRET`
- `JWT_ISSUER` — default: `finpharm-gateway`
- `JWT_EXPIRE_MINUTES` — default: `60`
- `RATE_LIMIT_ENABLED` — default: `true`
- `RATE_LIMIT_GENERAL_LIMIT` — default: `60`
- `RATE_LIMIT_AUTH_LIMIT` — default: `20`
- `RATE_LIMIT_WINDOW_SECONDS` — default: `60`

### Transaction
- `INVENTORY_BASE_URL` — default: `http://localhost:8082`
- `AI_AUDITOR_BASE_URL` — default: `http://localhost:8083`
- `AI_AUDITOR_TIMEOUT_MS`
- `DB_HOST` — default: `127.0.0.1`
- `DB_PORT` — default: `55432`
- `DB_USER` — default: `finpharm`
- `DB_PASSWORD` — default: `finpharm`
- `DB_NAME` — default: `transaction_db`
- `DB_SSLMODE` — default: `disable`

### Inventory
- `STORAGE_DRIVER` — default: `postgres`
- `DB_HOST` — default: `127.0.0.1`
- `DB_PORT` — default: `55432`
- `DB_USER` — default: `finpharm`
- `DB_PASSWORD` — default: `finpharm`
- `DB_NAME` — default: `inventory_db`
- `DB_SSLMODE` — default: `disable`

### AI Auditor
- `AUDIT_PROVIDER`
- `AUDIT_FAIL_OPEN`
- `GEMINI_API_KEY`
- `GEMINI_MODEL`
- `GEMINI_TIMEOUT_MS`

### Worker
- `RABBITMQ_URL`
- `RABBITMQ_EXCHANGE`
- `RABBITMQ_TRANSACTION_APPROVED_QUEUE`
- `RABBITMQ_TRANSACTION_APPROVED_ROUTING_KEY`
- `RABBITMQ_TRANSACTION_APPROVED_RETRY_QUEUE`
- `RABBITMQ_TRANSACTION_APPROVED_DLQ`
- `RABBITMQ_PREFETCH_COUNT`
- `RABBITMQ_MAX_RETRY_COUNT`
- `WORKER_METRICS_PORT`

### Knowledge
- `KNOWLEDGE_DB_HOST` — default: `127.0.0.1`
- `KNOWLEDGE_DB_PORT` — default: `55432`
- `KNOWLEDGE_DB_USER` — default: `finpharm`
- `KNOWLEDGE_DB_PASSWORD` — default: `finpharm`
- `KNOWLEDGE_DB_NAME` — default: `postgres`
- `KNOWLEDGE_DB_SSLMODE` — default: `disable`
- `KNOWLEDGE_SOURCE_DIR` — default: `./knowledge/sop`
- `KNOWLEDGE_EMBEDDING_MODEL` — default: `models/gemini-embedding-001`
- `KNOWLEDGE_EMBEDDING_DIMENSION` — default: `768`
- `KNOWLEDGE_ANSWER_MODEL`
- `KNOWLEDGE_ANSWER_TEMPERATURE`
- `KNOWLEDGE_ANSWER_MAX_OUTPUT_TOKENS`
- `KNOWLEDGE_ANSWER_MIN_TOP_SCORE`
- `KNOWLEDGE_ANSWER_MAX_CHUNKS_PER_DOCUMENT`
- `KNOWLEDGE_ANSWER_SCORE_WINDOW`

---

## Default Local Credentials / Notes

**PostgreSQL**
- user: `finpharm`
- password: `finpharm`

**RabbitMQ**
- user: `finpharm`
- password: `finpharm`

**Grafana**
- login default local: `admin` / `admin`

*Catatan:*
- Credential di atas hanya untuk local/dev/demo.
- Untuk non-local environment, ganti secret/default password.
- Gateway non-local akan menolak `JWT_SECRET=dev-secret-change-me`.

---

## API Endpoints

### Gateway
```text
GET  /health
GET  /metrics
POST /v1/auth/token
POST /v1/stock/check
POST /v1/transactions
GET  /v1/transactions
GET  /v1/medicines
GET  /v1/medicines/:id
POST /v1/chat/sop
GET  /v1/debug/sleep?ms=1000 — (local/dev only)
```

### Transaction
```text
GET  /
GET  /health
GET  /metrics
POST /v1/stock/check
POST /v1/transactions
GET  /v1/transactions
GET  /v1/debug/sleep?ms=1000 — (local/dev only)
```

### Inventory
```text
GET  /
GET  /health
GET  /metrics
POST /v1/stock/check
POST /v1/stock/deduct
GET  /v1/medicines
GET  /v1/medicines/:id
```

### AI Auditor
```text
GET  /
GET  /health
GET  /metrics
POST /v1/audit/transaction
```

### Knowledge
```text
GET  /
GET  /health
GET  /metrics
POST /v1/chat/sop
```

### Worker
```text
GET /metrics
```

---

## Example Requests

**Issue Token Staff**
```cmd
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"staff-001\",\"role\":\"staff\"}"
```

**Issue Token Supervisor**
```cmd
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"supervisor-001\",\"role\":\"supervisor\"}"
```

**Create Transaction via Gateway**
```cmd
curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-001" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

**Replay Request yang Sama**
```cmd
curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-001" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

**Check Stock**
```cmd
curl -i -X POST http://localhost:8080/v1/stock/check ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -d "{\"medicine_id\":\"PARA500\",\"qty\":1}"
```

**SOP Chatbot via Gateway**
```cmd
curl -i -X POST http://localhost:8080/v1/chat/sop ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -d "{\"question\":\"apakah amoxicillin bisa dijual tanpa resep?\",\"top_k\":5,\"min_score\":0.45}"
```

---

## Example Chatbot Response

### Positive / grounded SOP answer
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

### Fallback / out-of-domain answer
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

---

## Observability

### Prometheus
- buka `http://localhost:9090/targets`
- pastikan semua target `UP`

### Grafana
- buka `http://localhost:3000`
- login local default: `admin` / `admin`
- buka dashboard: **Finpharm Overview**

**Panel utama dashboard:**
- Total HTTP Requests by Service
- HTTP Request Rate by Service
- HTTP 4xx/5xx by Service
- HTTP P95 Latency by Service
- HTTP P99 Latency by Service
- Transaction Outcomes by Status
- Transaction Audit Decisions
- Transaction Endpoint Traffic
- AI Auditor Request Count
- AI Auditor Fallback Total
- Worker Events by Result
- Worker Inflight Messages
- Worker Processed Events Count
- HTTP Requests Breakdown

### Alerting
- buka `http://localhost:9093`
- pastikan Alertmanager hidup
- jalankan local webhook receiver bila ingin proof-of-delivery
- alert utama sudah diverifikasi end-to-end di Day 46

---

## Testing

**Jalankan semua test**
```powershell
go test ./... -count=1 -v
```

**Jalankan per area**
```powershell
go test ./services/gateway/... -count=1 -v
go test ./services/inventory/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./services/worker/... -count=1 -v
go test ./services/knowledge/... -count=1 -v
```

---

## Final Demo Flow yang Disarankan

1. Jalankan PostgreSQL dan RabbitMQ
2. Jalankan migrations:
   - `.\scripts\migrate-inventory-up.ps1`
   - `.\scripts\migrate-transaction-up.ps1`
   - `.\scripts\run-knowledge-migrate.ps1`
3. Jalankan knowledge ingestion:
   - `.\scripts\run-knowledge-ingest.ps1`
4. Jalankan semua service:
   - inventory
   - transaction
   - ai-auditor
   - knowledge API
   - gateway
   - worker
5. Jalankan Prometheus, Grafana, Alertmanager, dan alert webhook
6. Jalankan `.\scripts\demo-readiness.ps1`
7. Issue token staff dan supervisor
8. Tunjukkan:
   - list medicines
   - stock check
   - create transaction approved
   - create transaction flagged / pending review
   - replay by idempotency key
   - list transactions
9. Tunjukkan SOP chatbot via gateway:
   - positive SOP query
   - fallback out-of-domain query
10. Tunjukkan RabbitMQ / worker log
11. Tunjukkan Prometheus targets
12. Tunjukkan Grafana dashboard
13. Tunjukkan Alertmanager / webhook bila ingin demo observability lebih lengkap

---

## What This Project Demonstrates

FinPharm-AI menunjukkan kombinasi kemampuan backend yang kuat:
- Go / Golang service development
- microservices architecture
- REST API contract
- PostgreSQL / sqlx / migrations
- idempotency
- orchestration flow
- AI service integration
- event-driven worker
- JWT auth & authorization
- observability
- alerting pipeline
- pgvector + retrieval
- grounded answer synthesis
- gateway-integrated SOP chatbot

---

## Known Limitations

- deduct stock multi-item masih satu per satu
- belum ada kompensasi jika partial deduct terjadi
- rate limit saat ini masih in-memory dan single-instance oriented
- worker retry/DLQ business metrics spesifik belum dipisah lebih dalam
- Gin mode masih default debug saat local run
- credential local masih memakai default value untuk kemudahan demo/dev
- chatbot SOP masih single-turn, belum punya memory/session

---

## Documentation Map

- `RUNBOOK.md` → panduan operasional lokal
- `docs/day46.md` → final alerting foundation & verification
- `docs/day47.md` → RAG ingestion foundation
- `docs/day48.md` → retrieval foundation
- `docs/day49.md` → answer synthesis + citation-ready response
- `docs/day50.md` → retrieval quality improvement
- `docs/day51.md` → knowledge HTTP API
- `docs/day52.md` → gateway integration untuk chatbot SOP
- `docs/day53.md` → final portfolio closeout

---

## Final Note

Project ini sengaja berhenti di titik yang jelas:
- backend microservices portfolio
- observability-aware local system
- gateway-integrated knowledge assistant API

Setelah titik ini, penambahan fitur besar tidak lagi wajib untuk membuat repo ini kuat sebagai portfolio backend. Yang lebih penting sekarang adalah:
- menjaga repo tetap rapi
- menjaga demo flow tetap stabil
- menjelaskan keputusan desain dengan baik saat presentasi / interview