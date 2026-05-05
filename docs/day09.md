# Day 09 — Error Mapping (Not Found -> 404)

## Tujuan
- medicine tidak ditemukan -> 404 MEDICINE_NOT_FOUND
- validasi binding -> 400
- internal error -> 500

## Yang dibangun / diubah
- Usecase tidak lagi mengubah not-found jadi stok 0
- Handler mapping error -> status code

## File yang berubah / ditambah
### Transaction
- [ADD/MOD] services/transaction/internal/domain/errors.go (versi day09)
- [MOD] services/transaction/internal/repository/stock_memory_repo.go
- [MOD] services/transaction/internal/usecase/stock_usecase.go
- [MOD] services/transaction/internal/httpapi/handler/stock.go
- [MOD] services/transaction/internal/httpapi/handler/stock_test.go (tambah case 404)

## Cara verifikasi
- POST medicine_id tidak ada -> 404
- go test ./... -v