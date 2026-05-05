# Day 15 — Inventory: List Medicines (Pagination)

## Tujuan
- Menambahkan capability Inventory sebagai catalog obat.
- Membuat endpoint list obat dengan pagination sederhana.
- Tetap menggunakan response envelope `{ data, request_id }`.

## Yang dibangun / diubah
- Endpoint baru di Inventory:
  - `GET /v1/medicines?limit=10&offset=0`
- Implementasi data masih in-memory (nanti akan dipindah ke Postgres).
- Pagination basic:
  - default limit=10, max limit=100
  - offset >= 0
  - response mengembalikan `items`, `limit`, `offset`, `total`.

## Konsep yang dipelajari
- Read model / listing endpoint di microservices.
- Pagination sederhana untuk API.
- Handler DTO mapping (domain -> JSON DTO).
- Testing endpoint GET dengan httptest.

## File yang berubah / ditambah

### Inventory
- [ADD] services/inventory/internal/domain/medicine.go
- [ADD] services/inventory/internal/repository/medicine_memory_repo.go
- [ADD] services/inventory/internal/usecase/medicine_usecase.go
- [ADD] services/inventory/internal/httpapi/handler/medicine.go
- [MOD] services/inventory/internal/httpapi/router.go (register `GET /v1/medicines`)
- [ADD] services/inventory/internal/httpapi/handler/medicine_test.go

## Cara verifikasi
### Manual
- Jalankan inventory service
- Hit:
  - `GET http://localhost:8082/v1/medicines?limit=2&offset=0`
- Expected:
  - 200
  - body:
    ```json
    { "data": { "items": [...], "limit": 2, "offset": 0, "total": 3 }, "request_id": "..." }
    ```

### Testing
- `go test ./... -v` harus PASS

## Catatan / gotcha
- Pada implementasi in-memory, `total` dihitung dari panjang slice data.
- Nanti saat pindah ke DB, pagination akan berubah menjadi query SQL (LIMIT/OFFSET) dan `total` biasanya perlu query count terpisah.