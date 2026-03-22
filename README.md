# FinPharm-AI

FinPharm-AI adalah project **portfolio backend microservices berbasis Go** yang dibangun dengan pendekatan *learning by doing*. Domain yang dipilih adalah farmasi dan transaksi agar tetap relevan dengan target role backend engineer yang menuntut kombinasi:

- Go / Golang
- Microservices architecture
- REST API
- PostgreSQL / SQL
- Reliability dasar
- Idempotency
- Orchestration antar service
- Kesiapan untuk AI dan event-driven phase berikutnya

Project ini dibangun **bertahap per hari** agar proses belajar, keputusan desain, dan progres teknis bisa dibaca dengan jelas.

---

## Current Status

Posisi repo saat ini:

- **Phase 1 selesai**
  - Gateway + Transaction hidup
  - Contract awal microservices stabil
  - Health check, request-id, logging, timeout dasar

- **Phase 2 selesai**
  - Clean architecture dasar
  - Inventory Service
  - Gateway routing ke Inventory dan Transaction
  - Retry ringan + circuit breaker minimal
  - Testing handler dasar

- **Phase 3 selesai sampai Day 26**
  - Inventory Service sudah memakai **PostgreSQL + sqlx**
  - Transaction Service sudah menyimpan transaksi ke **PostgreSQL + sqlx**
  - `GET /v1/transactions` tersedia lewat Gateway
  - `POST /v1/transactions` sudah memakai **Idempotency-Key**
  - Transaction lifecycle berjalan:
    - `PENDING`
    - `APPROVED`
    - `FAILED`
  - Create transaction sekarang juga melakukan **deduct stock**
  - Replay dengan `Idempotency-Key` yang sama:
    - Tidak membuat transaksi baru
    - Tidak deduct stock lagi

> Fondasi microservices sudah stabil, persistence dua service domain sudah hidup, create transaction sudah idempotent, dan orchestration transaksi-ke-inventory sudah berjalan sinkron.

---

## Architecture

FinPharm-AI menggunakan arsitektur microservices sederhana dengan tiga service utama:

### 1. Gateway Service

Single entry point untuk client.

**Tugas:**

- Routing / proxy
- Request-id propagation
- Edge validation dasar
- Logging

Gateway **tidak** menyimpan business logic transaksi.

### 2. Transaction Service

Orchestration service untuk:

- Stock check
- Create transaction
- Idempotency
- Lifecycle status transaction
- Pemanggilan Inventory Service untuk deduct stock

### 3. Inventory Service

Source of truth untuk:

- Medicines catalog
- Stock availability
- Stock deduction

---

## Main Flow

### Stock Check

`Client -> Gateway -> Transaction -> Inventory`

### Medicines List / Detail

`Client -> Gateway -> Inventory`

### Create Transaction

`Client -> Gateway -> Transaction -> Inventory`

Flow create transaction saat ini:

1. Client kirim `POST /v1/transactions` ke Gateway
2. Gateway validasi request dasar + `Idempotency-Key`
3. Transaction Service cek replay by `idempotency_key`
4. Transaction Service pre-check stock ke Inventory
5. Transaction disimpan sebagai `PENDING`
6. Transaction Service minta Inventory deduct stock
7. Jika sukses → transaction diupdate jadi `APPROVED`
8. Jika gagal setelah transaction tercatat → transaction diupdate jadi `FAILED`

---

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

### Services

| Service     | Port                            | Description                                                                                         |
|------------ |-------------------------------- | --------------------------------------------------------------------------------------------------- |
| gateway     | 8080                            | Edge service untuk routing, proxy, request-id propagation, dan logging                              |
| transaction | 8081                            | Orchestration service untuk stock check, create transaction, idempotency, dan transaction lifecycle |
| inventory   | 8082                            | Source of truth untuk medicines catalog dan stock                                                   |
| postgres    | 55432 (host) / 5432 (container)| Database local dev, 1 instance dengan database logical terpisah per service                         |

---

## Requirements

- Go sesuai versi pada `go.mod`
- Docker Desktop
- PowerShell untuk menjalankan script `.ps1`

**Cek versi Go:**

```bash
go version
```

**Cek Docker:**

```bash
docker --version
docker compose version
```

---

## Running the Project

### 1. Jalankan PostgreSQL

```bash
docker compose up -d postgres
```

### 2. Jalankan migration Inventory

```powershell
.\scripts\migrate-inventory-up.ps1
```

### 3. Jalankan migration Transaction

```powershell
.\scripts\migrate-transaction-up.ps1
```

> Migration transaction saat ini sudah mencakup:
>
> - tabel `transactions`
> - tabel `transaction_items`
> - kolom `idempotency_key`

### 4. Jalankan Services

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-gateway.ps1
```

**Stop PostgreSQL:**

```bash
docker compose down
```

---

## Environment Variables

### Common

- `APP_ENV` — default: `local`
- `PORT`
- `READ_TIMEOUT_MS` — default: `5000`
- `WRITE_TIMEOUT_MS` — default: `5000`
- `IDLE_TIMEOUT_MS` — default: `30000`
- `SHUTDOWN_TIMEOUT_MS` — default: `7000`

### Gateway

- `TRANSACTION_BASE_URL` — default: `http://localhost:8081`
- `INVENTORY_BASE_URL` — default: `http://localhost:8082`

### Transaction

- `INVENTORY_BASE_URL` — default: `http://localhost:8082`
- `DB_HOST` — default: `127.0.0.1`
- `DB_PORT` — default: `55432`
- `DB_USER` — default: `finpharm`
- `DB_PASSWORD` — default: `finpharm`
- `DB_NAME` — default: `transaction_db`
- `DB_SSLMODE` — default: `disable`

### Inventory

- `STORAGE_DRIVER` — default: `postgres` (dari `scripts/run-inventory.ps1`)
- `DB_HOST` — default: `127.0.0.1`
- `DB_PORT` — default: `55432`
- `DB_USER` — default: `finpharm`
- `DB_PASSWORD` — default: `finpharm`
- `DB_NAME` — default: `inventory_db`
- `DB_SSLMODE` — default: `disable`

---

## API Endpoints

### Gateway

- `GET  /`
- `GET  /health`
- `POST /v1/stock/check`
- `POST /v1/transactions`
- `GET  /v1/transactions`
- `GET  /v1/medicines`
- `GET  /v1/medicines/:id`
- `GET  /v1/debug/sleep?ms=1000` — local/dev only

### Transaction

- `GET  /`
- `GET  /health`
- `POST /v1/stock/check`
- `POST /v1/transactions`
- `GET  /v1/transactions`
- `GET  /v1/debug/sleep?ms=1000` — local/dev only

### Inventory

- `GET  /`
- `GET  /health`
- `POST /v1/stock/check`
- `POST /v1/stock/deduct`
- `GET  /v1/medicines`
- `GET  /v1/medicines/:id`

---

## Example Requests

### Create Transaction via Gateway

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-001" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"
```

### Replay Request Yang Sama

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-001" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"
```

### List Approved Transactions

```bash
curl -i "http://localhost:8080/v1/transactions?status=approved"
```

### Check Stock

```bash
curl -i -X POST http://localhost:8080/v1/stock/check \
  -H "Content-Type: application/json" \
  -d "{\"medicine_id\":\"PARA500\",\"qty\":1}"
```

---

## Testing

### Jalankan semua test

```bash
go test ./... -count=1 -v
```

### Jalankan per service

```bash
go test ./services/inventory/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
```

---

## Known Limitations

- Deduct stock multi-item masih dilakukan satu per satu
- Jika item pertama berhasil deduct lalu item kedua gagal, transaction akan ditandai `FAILED`
- Belum ada mekanisme kompensasi untuk item pertama

**Masa depan / improvement:**

- Batch atomic deduct di Inventory
- Compensation
- Saga / event-driven orchestration

---

## Why This Project Matters for Portfolio

Project ini selaras dengan kebutuhan backend role target:

- Go / Golang
- Microservices
- SQL / PostgreSQL
- Distributed system mindset
- Idempotency
- API Gateway
- Orchestration service
- Reliability basics
- Next step ke AI dan event-driven architecture

---

## Next Focus

Setelah Milestone 3 stabil, fokus berikutnya:

- AI Auditor Service
- Gemini integration
- Transaction review status
- RabbitMQ / event-driven worker
- Security & observability