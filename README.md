# FinPharm-AI

FinPharm-AI adalah project portfolio backend microservices berbasis Go yang dibuat dengan pendekatan *learn by doing*. Domain yang dipilih adalah farmasi dan transaksi, agar tetap relevan dengan proses belajar backend service, sistem finansial ringan, serta kebutuhan integrasi AI di tahap berikutnya.

Project ini dirancang agar tidak hanya sekadar “jalan”, tetapi juga menunjukkan pemahaman terhadap praktik kerja industri seperti API Gateway, service boundary yang jelas, request tracing, standard response, resilience dasar, dan persiapan persistence per service. Saat ini Phase 2 sudah selesai, dan project sedang memasuki awal Phase 3 untuk fondasi PostgreSQL.

## Architecture

FinPharm-AI menggunakan arsitektur microservices sederhana dengan tiga service utama:

* **Gateway Service** berperan sebagai *single entry point* untuk client. Tugasnya fokus pada routing, proxy, request-id propagation, dan logging.
* **Transaction Service** berperan sebagai orchestration service untuk alur `stock check`. Service ini menerima request dari Gateway lalu memanggil Inventory Service.
* **Inventory Service** berperan sebagai source of truth untuk data obat dan stok, termasuk endpoint medicines list, medicine detail, dan stock check.

Alur utama saat ini:

* `Client -> Gateway -> Transaction -> Inventory` untuk `POST /v1/stock/check`
* `Client -> Gateway -> Inventory` untuk `GET /v1/medicines`
* `Client -> Gateway -> Inventory` untuk `GET /v1/medicines/:id`

## Repository Structure

```text
finpharm-ai/
  .github/
    workflows/
  docs/
  scripts/
    postgres/
  services/
    gateway/
    inventory/
    transaction/
  docker-compose.yml
  go.mod
  go.sum
  README.md
```

## Services

| Service     | Port | Description                                                                                |
|-------------|------|--------------------------------------------------------------------------------------------|
| gateway     | 8080 | Edge service untuk routing, proxy, request-id propagation, dan logging                   |
| transaction | 8081 | Orchestration service untuk stock check, termasuk retry ringan dan circuit breaker minimal |
| inventory   | 8082 | Source of truth untuk medicines catalog dan stock                                         |
| postgres    | 5432 | Database foundation untuk Phase 3, memakai 1 instance dengan multi-database per service  |

## Requirements

* Go sesuai versi pada `go.mod`
* Docker Desktop
* PowerShell untuk menjalankan script `.ps1`

Cek versi Go:

```bash
go version
```

Cek Docker:

```bash
docker --version
docker compose version
```

## Running the Project

### Terminal 1 — PostgreSQL

```bash
docker compose up -d postgres
```

### Terminal 2 — Inventory Service

```powershell
.\scripts\run-inventory.ps1
```

### Terminal 3 — Transaction Service

```powershell
.\scripts\run-transaction.ps1
```

### Terminal 4 — Gateway Service

```powershell
.\scripts\run-gateway.ps1
```

Untuk menghentikan PostgreSQL:

```bash
docker compose down
```

## Environment Variables

### Common

* `APP_ENV` — default `local`
* `PORT`
* `READ_TIMEOUT_MS` — default `5000`
* `WRITE_TIMEOUT_MS` — default `5000`
* `IDLE_TIMEOUT_MS` — default `30000`
* `SHUTDOWN_TIMEOUT_MS` — default `7000`

### Gateway

* `TRANSACTION_BASE_URL` — default `http://localhost:8081`
* `INVENTORY_BASE_URL` — default `http://localhost:8082`

### Transaction

* `INVENTORY_BASE_URL` — default `http://localhost:8082`
* `DB_HOST` — default `localhost`
* `DB_PORT` — default `5432`
* `DB_USER` — default `finpharm`
* `DB_PASSWORD` — default `finpharm`
* `DB_NAME` — default `transaction_db`
* `DB_SSLMODE` — default `disable`

### Inventory

* `DB_HOST` — default `localhost`
* `DB_PORT` — default `5432`
* `DB_USER` — default `finpharm`
* `DB_PASSWORD` — default `finpharm`
* `DB_NAME` — default `inventory_db`
* `DB_SSLMODE` — default `disable`

## API Endpoints

### Gateway

```text
GET  /
GET  /health
POST /v1/stock/check
GET  /v1/medicines
GET  /v1/medicines/:id
GET  /v1/debug/sleep?ms=1000   # local/dev only
```

### Transaction

```text
GET  /
GET  /health
POST /v1/stock/check
GET  /v1/debug/sleep?ms=1000   # local/dev only
```

### Inventory

```text
GET  /
GET  /health
POST /v1/stock/check
GET  /v1/medicines
GET  /v1/medicines/:id
```

## Example Request

Check stock via Gateway:

```bash
curl -i -X POST http://localhost:8080/v1/stock/check \
  -H "Content-Type: application/json" \
  -d "{\"medicine_id\":\"PARA500\",\"qty\":10}"
```

Contoh response:

```json
{
  "data": {
    "medicine_id": "PARA500",
    "requested_qty": 10,
    "available_qty": 80,
    "is_available": true
  },
  "request_id": "example-request-id"
}
```

Medicines list via Gateway:

```bash
curl -i "http://localhost:8080/v1/medicines?limit=2&offset=0"
```

Medicine detail via Gateway:

```bash
curl -i "http://localhost:8080/v1/medicines/PARA500"
```

## Testing

Jalankan semua test:

```bash
go test ./... -v
```

Jalankan per service:

```bash
go test ./services/gateway/... -v
go test ./services/inventory/... -v
go test ./services/transaction/... -v
```

Jalankan package handler tertentu:

```bash
go test ./services/gateway/internal/httpapi/handler -v
go test ./services/inventory/internal/httpapi/handler -v
go test ./services/transaction/internal/httpapi/handler -v
```

## CI

Project ini menggunakan GitHub Actions melalui file `.github/workflows/ci.yml`.

CI akan berjalan otomatis pada:

* push ke branch `main`
* setiap pull request

Command yang dijalankan di CI:

```bash
go test ./... -v
```

## Implemented Best Practices

* API Gateway sebagai edge service yang tipis
* Request ID propagation menggunakan `X-Request-ID`
* Caller service header standar menggunakan `X-Caller-Service`
* Structured logging dengan `log/slog`
* Graceful shutdown dan HTTP server timeouts
* Manual dependency injection
* Standard response envelope untuk success dan error
* Retry ringan pada Transaction Service saat call ke Inventory
* Circuit breaker minimal pada Transaction Service
* Unit testing handler dan proxy dengan `httptest`
* Dokumentasi progress harian di folder `docs/`

## Roadmap

* Phase 3 — PostgreSQL persistence untuk Inventory dan Transaction
* Migration per service dengan boundary yang jelas
* Inventory repository pindah dari in-memory ke database
* Transaction persistence untuk transaksi real
* Idempotency key untuk transaksi
* Pagination dan filtering yang lebih solid
* Phase 4 — AI Auditor Service dengan Gemini/OpenAI
* Phase 5 — RAG untuk SOP farmasi
* Phase 6 — Event-driven workflow dengan RabbitMQ
* Phase 7 — Security, metrics, audit logging, dan hardening