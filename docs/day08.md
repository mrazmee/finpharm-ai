# Day 08 — Clean Architecture Step 1 (Transaction)

## Tujuan
- Refactor transaksi stock check ke domain/usecase/repository
- Endpoint tetap jalan

## Yang dibangun / diubah
- domain contract + repository in-memory + usecase
- handler hanya parsing & mapping response
- manual DI di router

## Konsep yang dipelajari
- Clean Architecture layering
- Manual dependency injection

## File yang berubah / ditambah
### Transaction
- [ADD] services/transaction/internal/domain/stock.go
- [ADD] services/transaction/internal/repository/stock_memory_repo.go
- [ADD] services/transaction/internal/usecase/stock_usecase.go
- [MOD] services/transaction/internal/httpapi/handler/stock.go
- [MOD] services/transaction/internal/httpapi/router.go

## Cara verifikasi
- go test ./services/transaction/... -v
- POST via gateway masih OK