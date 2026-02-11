# FinPharm-AI (Microservices Bootcamp)

FinPharm-AI adalah project latihan backend **microservices** menggunakan **Go + Gin** dengan pendekatan *learn by doing* (progress harian).  
Saat ini (Phase 1) repo berisi 2 service:

- **Gateway Service** → pintu masuk (proxy/routing, observability dasar)
- **Transaction Service** → service domain awal (logika stok in-memory + endpoint debug untuk simulasi)

---

## Struktur Repo

finpharm-ai/
services/
gateway/
transaction/
scripts/
.github/workflows/
go.mod
README.md


---

## Services

| Service      | Default Port | URL Base                  |
|-------------|--------------|---------------------------|
| gateway     | 8080         | http://localhost:8080     |
| transaction | 8081         | http://localhost:8081     |

---

## Requirements

- Go (mengikuti versi yang tertera di `go.mod`)

Cek versi Go kamu:
```bash
go version

Menjalankan Project (Local)

Jalankan dua terminal (PowerShell) agar 2 service hidup bersamaan.

Terminal 1 — Transaction

.\scripts\run-transaction.ps1

Terminal 2 — Gateway

.\scripts\run-gateway.ps1


Common (Gateway & Transaction)

APP_ENV

default: local

set prod untuk mematikan debug routes

READ_TIMEOUT_MS (default: 5000)

WRITE_TIMEOUT_MS (default: 5000)

IDLE_TIMEOUT_MS (default: 30000)

SHUTDOWN_TIMEOUT_MS (default: 7000)

Gateway

PORT (default: 8080)

TRANSACTION_BASE_URL (default: http://localhost:8081)

Transaction

PORT (default: 8081)

Endpoints
Gateway

GET / → hello gateway

GET /health → JSON status

POST /v1/stock/check → forward/proxy ke Transaction

GET /v1/debug/sleep?ms=1000 → proxy debug (hanya aktif jika APP_ENV=local/dev)

Transaction

GET / → hello transaction service

GET /health → JSON status

POST /v1/stock/check → check stock (in-memory)

GET /v1/debug/sleep?ms=1000 → sleep endpoint (hanya aktif jika APP_ENV=local/dev)

Contoh Request
Check Stock (via Gateway)

curl -i -X POST http://localhost:8080/v1/stock/check \
  -H "Content-Type: application/json" \
  -d "{\"medicine_id\":\"PARA500\",\"qty\":10}"

Contoh response:

{
  "medicine_id": "PARA500",
  "requested_qty": 10,
  "available_qty": 80,
  "is_available": true
}


Debug Sleep (Local Only)

curl -i "http://localhost:8080/v1/debug/sleep?ms=1000"

ika APP_ENV=prod, endpoint debug harus:

404 page not found

Testing

Jalankan semua test:

go test ./... -v

Jalankan per service:

go test ./services/gateway/... -v
go test ./services/transaction/... -v

Jalankan per package handler:

go test ./services/gateway/internal/httpapi/handler -v
go test ./services/transaction/internal/httpapi/handler -v

CI (GitHub Actions)

Repo memakai GitHub Actions untuk menjalankan:

go test ./... -v

CI berjalan otomatis pada:

push ke branch main

setiap Pull Request

Catatan Best Practice yang Sudah Diterapkan

Request ID propagation (X-Request-ID) dari Gateway ke Transaction

Structured logging (JSON) + correlation by request_id

Graceful shutdown dengan timeout

HTTP server timeouts (read/write/idle)

Debug routes dikontrol dengan APP_ENV (mati di prod)

Basic unit tests untuk handler & proxy

Next Milestone

Phase berikutnya akan mulai menerapkan Clean Architecture di Transaction Service:

domain → usecase → repository → delivery

dependency injection (manual)

persiapan masuk DB (Postgres + migrations)

::contentReference[oaicite:0]{index=0}
