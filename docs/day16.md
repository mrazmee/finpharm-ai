# Day 16 — Inventory: Get Medicine By ID (Detail Endpoint)

## Tujuan
- Menambahkan endpoint detail obat di Inventory:
  - `GET /v1/medicines/:id`
- Mengembalikan response dengan format standar:
  - sukses: `{ "data": ..., "request_id": "..." }`
  - error not found: HTTP 404 `MEDICINE_NOT_FOUND`

## Yang dibangun / diubah
- Menambahkan fungsi `GetByID` pada repository obat (in-memory).
- Menambahkan usecase `GetMedicine` (dengan validasi basic untuk parameter `id`).
- Menambahkan handler `GetMedicine` untuk endpoint `GET /v1/medicines/:id`.
- Menambahkan routing endpoint baru di router Inventory.
- Menambahkan unit test:
  - case sukses untuk id valid
  - case 404 untuk id tidak ditemukan

## Konsep yang dipelajari
- Endpoint “detail by id” sebagai pola REST standar.
- Mapping domain error (`NotFoundError`, `ValidationError`) ke HTTP status code (404/400).
- Menjaga konsistensi API contract lewat response envelope.
- Testing endpoint route params (`:id`) dengan `httptest`.

## File yang berubah / ditambah

### Inventory
- [MOD] services/inventory/internal/domain/medicine.go  
  - tambah `GetByID` di `MedicineRepository`
  - tambah `GetMedicine` di `MedicineUsecase`
- [MOD] services/inventory/internal/repository/medicine_memory_repo.go  
  - implement `GetByID` + return `NotFoundError` jika tidak ada
- [MOD] services/inventory/internal/usecase/medicine_usecase.go  
  - tambah `GetMedicine` + validasi `id`
- [MOD] services/inventory/internal/httpapi/handler/medicine.go  
  - tambah handler `GetMedicine`
- [MOD] services/inventory/internal/httpapi/router.go  
  - register route `GET /v1/medicines/:id`
- [MOD] services/inventory/internal/httpapi/handler/medicine_test.go  
  - tambah test `TestGetMedicine_OK` dan `TestGetMedicine_NotFound`

## Cara verifikasi
### Manual
- Jalankan inventory:
  - `.\scripts\run-inventory.ps1`
- Test sukses:
  - `GET http://localhost:8082/v1/medicines/PARA500`
  - Expected: 200 + `data` + `request_id`
- Test not found:
  - `GET http://localhost:8082/v1/medicines/PARA200`
  - Expected: 404 + `MEDICINE_NOT_FOUND`

### Testing
- `go test ./... -v` harus PASS

## Catatan
- Pada log, `path` untuk endpoint param akan terlihat sebagai template route: `/v1/medicines/:id` (ini normal dari `c.FullPath()`).