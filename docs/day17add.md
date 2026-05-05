# Day 17 (Addendum) — Transaction Wiring Cleanup + Request-ID Context + Retry/Circuit Breaker

## Tujuan
- Mendokumentasikan perubahan **tambahan Day 17** yang sempat dikerjakan di local tetapi belum ter-push ke repo.
- Merapikan wiring `Transaction Service` agar lebih dekat ke pola Clean Architecture.
- Menjaga `request_id` tetap ikut sampai ke downstream `Inventory Service` tanpa membuat handler membuat repo sendiri.
- Menambahkan resilience dasar saat `Transaction Service` memanggil `Inventory Service`.

## Yang dibangun / diubah
- `Transaction Service` handler `POST /v1/stock/check` tidak lagi membuat repository/usecase di dalam handler.
- Router transaction sekarang melakukan **manual dependency injection**:
  - buat shared `http.Client`
  - buat `CircuitBreaker`
  - buat `StockHTTPRepo`
  - inject ke `StockUsecase`
  - inject ke `StockHandler`
- Menambahkan helper context untuk menyimpan dan membaca `request_id`:
  - `WithRequestID(ctx, requestID)`
  - `RequestIDFromContext(ctx)`
- `StockHTTPRepo` di transaction diperbarui agar:
  - membaca `request_id` dari context
  - propagate header `X-Request-ID` ke inventory
  - mengirim header `X-Caller-Service: transaction`
  - retry ringan **1x** untuk error upstream yang retryable
  - memakai **circuit breaker minimal** agar fail fast saat inventory terus gagal
- Menambahkan test transaction untuk:
  - retry sekali saat upstream error lalu sukses
  - circuit breaker open lalu request berikutnya langsung `502` tanpa call tambahan ke inventory

## Konsep yang dipelajari
- Manual dependency injection di level router agar wiring dependency terlihat jelas.
- Pemakaian `context.Context` untuk membawa metadata request (`request_id`) ke layer repository.
- Retry ringan hanya untuk error upstream yang memang layak diulang.
- Circuit breaker pattern dasar (`CLOSED`, `OPEN`, `HALF_OPEN`) untuk melindungi service dari upstream yang tidak sehat.
- Transaction service sebagai orchestrator, sementara inventory tetap source of truth untuk stok.

## File yang berubah / ditambah

### Transaction
- [MOD] `services/transaction/internal/httpapi/handler/stock.go`
  - handler menerima usecase via constructor dan menyisipkan `request_id` ke context
- [MOD] `services/transaction/internal/httpapi/router.go`
  - tambah manual DI untuk `http.Client`, `CircuitBreaker`, repository, usecase, dan handler
- [MOD] `services/transaction/internal/repository/stock_http_repo.go`
  - baca `request_id` dari context, tambah retry ringan, tambah integrasi circuit breaker
- [MOD] `services/transaction/internal/httpapi/handler/stock_test.go`
  - test retry 1x dan fail-fast saat breaker open
- [ADD] `services/transaction/internal/repository/circuit_breaker.go`
  - implementasi circuit breaker minimal
- [ADD] `services/transaction/internal/repository/request_id_ctx.go`
  - helper simpan/ambil `request_id` dari context

## Cara verifikasi
### Manual
- Jalankan 3 service: gateway, transaction, inventory
- Cek stock via gateway:
  - `POST http://localhost:8080/v1/stock/check`
- Cek medicines list via gateway (pagination sederhana):
  - `GET http://localhost:8080/v1/medicines?limit=2&offset=0`

### Testing
- Jalankan:
  - `go test ./... -v`
- Fokus transaction:
  - `go test ./services/transaction/... -v`

## Catatan
- Addendum ini bukan menggantikan `docs/day17.md`, tetapi melengkapi perubahan Day 17 yang tertinggal di local working tree.
- Circuit breaker di tahap ini masih minimal dan belum persisten; targetnya untuk belajar konsep resilience dulu.
- Header antar-service yang dipakai project saat ini adalah `X-Request-ID` dan `X-Caller-Service`.