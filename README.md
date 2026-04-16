# FinPharm-AI

FinPharm-AI adalah **portfolio backend microservices berbasis Go** yang dibangun dengan pendekatan *learning by doing* untuk menampilkan fondasi yang relevan dengan target role backend engineer:

- Go / Golang
- Microservices architecture
- REST API
- PostgreSQL / SQL
- Reliability dasar
- Idempotency
- Event-driven processing
- JWT auth & authorization
- Observability
- Kesiapan untuk AI dan optional advanced RAG phase berikutnya

Project ini dibangun **bertahap per hari** agar proses belajar, keputusan desain, dan progres teknis bisa dibaca dengan jelas.

---

## Current Status

Posisi repo saat ini:

### Core backend sudah selesai
- Gateway Service
- Transaction Service
- Inventory Service
- AI Auditor Service
- Worker Service

### Persistence & orchestration sudah hidup
- Inventory Service memakai **PostgreSQL + sqlx**
- Transaction Service memakai **PostgreSQL + sqlx**
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

### Observability sudah hidup
- Request metrics per service
- Worker metrics dasar
- Prometheus local scrape
- Grafana local dashboard
- Audit logging + trace header propagation

> **Catatan:** Fondasi utama portfolio sudah ada dan terasa production-like untuk local/demo: domain service terpisah, persistence dua domain utama hidup, AI auditor dan event-driven worker berjalan, auth ada, observability ada, dan hardening dasar sudah masuk.

---

## Architecture

FinPharm-AI menggunakan arsitektur microservices sederhana dengan lima komponen utama:

### 1. Gateway Service
Single entry point untuk client.

**Tugas:**
- Routing / proxy
- Request-id propagation
- Trace header propagation
- Edge validation dasar
- JWT issue token
- Authorization dasar
- Rate limit dasar
- Logging

*Gateway **tidak** menyimpan business logic transaksi.*

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

---

## Main Flow

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

---

## Repository Structure

```text
finpharm-ai/
├─ .github/
│  └─ workflows/
├─ docs/
├─ observability/
│  ├─ prometheus/
│  └─ grafana/
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
│  └─ worker/
├─ docker-compose.yml
├─ docker-compose.rabbitmq.yml
├─ docker-compose.prometheus.yml
├─ docker-compose.grafana.yml
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
| **Gateway** | `8080` | Edge service untuk routing, auth, rate limit, request-id, trace, dan logging |
| **Transaction** | `8081` | Orchestration service untuk stock check, create tx, idempotency, audit result, event publish |
| **Inventory** | `8082` | Source of truth untuk medicines dan stock |
| **AI Auditor** | `8083` | Audit service untuk decision transaksi |
| **Worker Metrics** | `9094` | Metrics endpoint worker |
| **Prometheus** | `9090` | Local metrics aggregation |
| **Grafana** | `3000` | Dashboard observability |
| **RabbitMQ UI** | `15672`| RabbitMQ management UI |
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

**3. Jalankan Migration**
```powershell
.\scripts\migrate-inventory-up.ps1
.\scripts\migrate-transaction-up.ps1
```

**4. Jalankan Services (Buka terminal terpisah)**
```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

**5. Jalankan Observability**
```powershell
.\scripts\run-prometheus.ps1
.\scripts\run-grafana.ps1
```

**6. Cek Readiness Demo**
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
- `SHUTDOWN_TIMEOUT_MS` — default: `5000` atau sesuai service

### Gateway
- `TRANSACTION_BASE_URL` — default: `http://localhost:8081`
- `INVENTORY_BASE_URL` — default: `http://localhost:8082`
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
- `AI_AUDITOR_BASE_URL` — sesuai script / config lokal bila dipakai
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
- `GEMINI_API_KEY` — opsional jika memakai provider utama (fallback tetap tersedia untuk local/demo)

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

---

## Default Local Credentials / Notes

**PostgreSQL**
- user: `finpharm`
- password: `finpharm`

**RabbitMQ**
- user: `finpharm`
- password: `finpharm`

**Grafana**
- default local login: `admin` / `admin`

*Catatan:*
- Credential di atas hanya untuk local/dev/demo.
- Untuk non-local environment, ganti secret/default password.
- Gateway non-local akan menolak `JWT_SECRET=dev-secret-change-me`.

---

## API Endpoints

### Gateway
```text
GET  /
GET  /health
GET  /metrics
POST /v1/auth/token
POST /v1/stock/check
POST /v1/transactions
GET  /v1/transactions
GET  /v1/medicines
GET  /v1/medicines/:id
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

### Worker
```text
GET  /metrics — (worker metrics endpoint)
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

**Replay Request Yang Sama**
```cmd
curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-001" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

**List Transactions**
```cmd
curl -i "http://localhost:8080/v1/transactions?limit=5&offset=0" ^
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>"
```

**Check Stock**
```cmd
curl -i -X POST http://localhost:8080/v1/stock/check ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -d "{\"medicine_id\":\"PARA500\",\"qty\":1}"
```

---

## Observability

### Prometheus
- Buka `http://localhost:9090/targets`
- Pastikan semua target `UP`

### Grafana
- Buka `http://localhost:3000`
- Login local default: `admin` / `admin`
- Buka dashboard: **Finpharm Overview**

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
```

---

## Demo Flow yang Disarankan

1. Jalankan database, RabbitMQ, semua service, Prometheus, dan Grafana
2. Jalankan `.\scripts\demo-readiness.ps1`
3. Issue token staff dan supervisor
4. Tunjukkan:
   - list medicines
   - stock check
   - create transaction approved
   - create transaction flagged / pending review
   - replay by idempotency key
5. Tunjukkan RabbitMQ / worker log
6. Tunjukkan Prometheus targets
7. Tunjukkan Grafana dashboard

*Optional:*
- generate traffic `.\scripts\generate-traffic.ps1`
- generate 401/400/429 untuk panel error

---

## Known Limitations

- Deduct stock multi-item masih satu per satu
- Belum ada kompensasi jika partial deduct terjadi
- Rate limit saat ini masih in-memory dan single-instance oriented
- Worker retry/DLQ business metrics spesifik belum dipisah kaya observability utama
- Grafana alerting dasar belum diaktifkan
- Credential local masih memakai default value untuk kemudahan demo/dev

---

## Why This Project Matters for Portfolio

Project ini menampilkan kombinasi yang kuat untuk backend role:
- Go / Golang
- microservices
- REST contract
- PostgreSQL / sqlx
- idempotency
- orchestration
- AI service integration
- event-driven worker
- JWT auth
- observability
- runtime hardening dasar

---

## Next Focus

Setelah fondasi production-like local ini stabil, optional advanced berikutnya adalah:
- RAG + Vector DB
- SOP ingestion
- retrieval
- citation
- chatbot