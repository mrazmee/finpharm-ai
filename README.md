# FinPharm-AI

FinPharm-AI adalah project portfolio backend microservices berbasis Go yang dibuat dengan pendekatan *learn by doing*. Domain yang dipilih adalah farmasi dan transaksi agar tetap relevan dengan proses belajar backend service, sistem finansial ringan, dan integrasi AI pada tahap berikutnya.

Project ini dirancang supaya tidak hanya sekadar “jalan”, tetapi juga menunjukkan pemahaman terhadap praktik kerja industri seperti API Gateway, service boundary yang jelas, request tracing, standard response, resilience dasar, persistence per service, dan dokumentasi progres yang rapi.

## Current Status

Posisi project saat ini:

- **Phase 1 selesai**: gateway dan transaction sudah stabil untuk contract awal.
- **Phase 2 selesai**: clean architecture dasar, inventory service, routing gateway, request-id, logging, timeout, testing dasar, retry ringan, dan circuit breaker minimal sudah ada.
- **Phase 3 sedang berjalan**:
  - `Inventory Service` sudah memakai **PostgreSQL + sqlx** saat dijalankan normal.
  - migration inventory pertama sudah tersedia dan sudah dipakai untuk tabel `medicines` + `stocks`.
  - migration awal `Transaction Service` sudah tersedia untuk tabel `transactions` + `transaction_items`.
  - `Transaction Service` masih **belum** menyimpan transaksi ke database pada runtime; wiring repository `sqlx` dan create transaction use case akan menjadi fokus berikutnya.

Jadi kondisi repo saat ini paling tepat dibaca sebagai:

> fondasi microservices sudah stabil, inventory persistence sudah hidup, dan transaction persistence sedang mulai dibangun dari schema migration.

## Architecture

FinPharm-AI menggunakan arsitektur microservices sederhana dengan tiga service utama:

- **Gateway Service** sebagai *single entry point* untuk client. Tugasnya fokus pada routing, proxy, request-id propagation, dan logging.
- **Transaction Service** sebagai orchestration service untuk alur `stock check`. Service ini menerima request dari Gateway lalu memanggil Inventory Service.
- **Inventory Service** sebagai source of truth untuk data obat dan stok, termasuk endpoint medicines list, medicine detail, dan stock check.

Alur utama saat ini:

- `Client -> Gateway -> Transaction -> Inventory` untuk `POST /v1/stock/check`
- `Client -> Gateway -> Inventory` untuk `GET /v1/medicines`
- `Client -> Gateway -> Inventory` untuk `GET /v1/medicines/:id`

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

| Service     | Port | Description |
|-------------|------|-------------|
| gateway     | 8080 | Edge service untuk routing, proxy, request-id propagation, dan logging |
| transaction | 8081 | Orchestration service untuk stock check, dan mulai menyiapkan persistence transaction |
| inventory   | 8082 | Source of truth untuk medicines catalog dan stock |
| postgres    | 55432 (host) / 5432 (container) | Database foundation untuk Phase 3, memakai 1 instance dengan database logical terpisah per service |

## Requirements

- Go sesuai versi pada `go.mod`
- Docker Desktop
- PowerShell untuk menjalankan script `.ps1`

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

### 1. Jalankan PostgreSQL

```bash
docker compose up -d postgres
```

### 2. Jalankan migration inventory

```powershell
.\scripts\migrate-inventory-up.ps1
```

Catatan:
- langkah ini diperlukan karena `Inventory Service` default-nya sudah memakai `postgres`
- migration ini membuat tabel `medicines` dan `stocks`

### 3. Jalankan Inventory Service

```powershell
.\scripts\run-inventory.ps1
```

### 4. Jalankan Transaction Service

```powershell
.\scripts\run-transaction.ps1
```

### 5. Jalankan Gateway Service

```powershell
.\scripts\run-gateway.ps1
```

### 6. Jalankan migration transaction (untuk Phase 3 transaction persistence)

```powershell
.\scripts\migrate-transaction-up.ps1
```

Catatan:
- langkah ini mulai relevan saat masuk modul persistence `Transaction Service`
- migration ini membuat tabel `transactions` dan `transaction_items`
- saat ini service transaction belum menulis data ke tabel tersebut, tetapi schema awalnya sudah disiapkan

Untuk menghentikan PostgreSQL:

```bash
docker compose down
```

## Environment Variables

### Common

- `APP_ENV` — default `local`
- `PORT`
- `READ_TIMEOUT_MS` — default `5000`
- `WRITE_TIMEOUT_MS` — default `5000`
- `IDLE_TIMEOUT_MS` — default `30000`
- `SHUTDOWN_TIMEOUT_MS` — default `7000`

### Gateway

- `TRANSACTION_BASE_URL` — default `http://localhost:8081`
- `INVENTORY_BASE_URL` — default `http://localhost:8082`

### Transaction

- `INVENTORY_BASE_URL` — default `http://localhost:8082`
- `DB_HOST` — default `127.0.0.1`
- `DB_PORT` — default `55432`
- `DB_USER` — default `finpharm`
- `DB_PASSWORD` — default `finpharm`
- `DB_NAME` — default `transaction_db`
- `DB_SSLMODE` — default `disable`

### Inventory

- `STORAGE_DRIVER` — default `postgres` dari `scripts/run-inventory.ps1`
- `DB_HOST` — default `127.0.0.1`
- `DB_PORT` — default `55432`
- `DB_USER` — default `finpharm`
- `DB_PASSWORD` — default `finpharm`
- `DB_NAME` — default `inventory_db`
- `DB_SSLMODE` — default `disable`

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

- push ke branch `main`
- setiap pull request