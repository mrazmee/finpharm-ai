# Day 12 — Add Inventory Service + Transaction calls Inventory (HTTP)

## Tujuan
- Membuat service baru **Inventory Service** sebagai **source of truth** untuk stok obat.
- Mengubah Transaction Service agar **tidak menyimpan stok sendiri**, melainkan **mengambil stok dari Inventory** via HTTP.
- Menjaga kontrak client tetap sama:
  - Client tetap hit **Gateway** → **Transaction** → (baru) **Inventory**
- Memastikan flow microservices makin “beneran”: service ownership jelas dan komunikasi antar service terjadi.

---

## Yang dibangun / diubah

### 1) Inventory Service (baru)
- Service baru `inventory` (default port `8082`)
- Endpoint:
  - `POST /v1/stock/check`
- Response sudah mengikuti standar yang sama:
  - sukses: `{ "data": ..., "request_id": "..." }`
  - error: `{ "error": { ... , "request_id": "..." } }`
- Middleware:
  - `RequestID` (menerima `X-Request-ID` jika ada, generate jika tidak)
  - structured logging JSON (`slog`)
- Graceful shutdown + HTTP server timeouts (sama seperti service lain)

### 2) Transaction Service
- Tambah repo baru: `stock_http_repo.go`
  - Implementasi `domain.StockRepository`
  - Mengambil data stok dari Inventory melalui HTTP
- DI di router Transaction berubah:
  - dulu: `StockMemoryRepo`
  - sekarang: `StockHTTPRepo (InventoryBaseURL)`
- Tambah config:
  - `INVENTORY_BASE_URL` (default `http://localhost:8082`)

### 3) Scripts
- Tambah `scripts/run-inventory.ps1` agar mudah menjalankan Inventory Service.

---

## Konsep yang dipelajari
- **Service ownership (microservices):** stok & obat dimiliki Inventory, bukan Transaction.
- **Communication antar service:** Transaction melakukan sync HTTP call ke Inventory.
- **Repository pattern di microservices:** interface repo tetap sama, implementasinya bisa:
  - in-memory (untuk dev/test)
  - HTTP (untuk call service lain)
  - nanti DB (Postgres)
- **Test isolation:** karena Transaction sekarang depend ke Inventory, test harus memakai mock upstream (httptest) agar unit test tidak tergantung service beneran.

---

## File yang berubah / ditambah

### Inventory (Service baru)
- [ADD] services/inventory/cmd/api/main.go
- [ADD] services/inventory/internal/config/config.go
- [ADD] services/inventory/internal/httpapi/router.go
- [ADD] services/inventory/internal/httpapi/handler/health.go
- [ADD] services/inventory/internal/httpapi/handler/response.go
- [ADD] services/inventory/internal/httpapi/handler/stock.go
- [ADD] services/inventory/internal/httpapi/middleware/request_id.go
- [ADD] services/inventory/internal/httpapi/middleware/logger.go
- [ADD] services/inventory/internal/domain/stock.go
- [ADD] services/inventory/internal/domain/errors.go
- [ADD] services/inventory/internal/repository/stock_memory_repo.go
- [ADD] services/inventory/internal/usecase/stock_usecase.go

### Transaction
- [MOD] services/transaction/internal/config/config.go (tambah INVENTORY_BASE_URL)
- [ADD] services/transaction/internal/repository/stock_http_repo.go
- [MOD] services/transaction/internal/httpapi/router.go (DI pakai HTTP repo)
- [MOD] services/transaction/internal/httpapi/handler/stock_test.go (mock inventory via httptest)

### Scripts
- [ADD] scripts/run-inventory.ps1

---

## Cara menjalankan (3 terminal)

### Terminal 1 — Inventory
```powershell
.\scripts\run-inventory.ps1

## Catatan penting (Day 12)
- Source of truth stok dipindahkan ke **Inventory Service**.
- Transaction sekarang menggunakan repository **HTTP** (`stock_http_repo.go`) untuk mengambil stok dari Inventory.
- File `services/transaction/internal/repository/stock_memory_repo.go` **tidak dipakai saat runtime** karena DI router memakai HTTP repo.
  - File ini tetap disimpan untuk kebutuhan **unit test/dev mode/fallback** dan pembelajaran perbedaan implementasi repository.
- File `services/transaction/internal/domain/errors.go` **tetap digunakan**, karena:
  - Transaction tetap melakukan validasi domain (menghasilkan `ValidationError`)
  - Transaction melakukan mapping error dari Inventory (mis. 404) menjadi `NotFoundError` agar handler Transaction bisa memberi response konsisten (400/404/500) ke client.