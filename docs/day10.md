# Day 10 — Typed Errors (ValidationError + NotFoundError)

## Tujuan
- Validasi domain punya detail field+reason
- NotFound punya resource+key
- Handler mapping jadi lebih rapi

## Yang dibangun / diubah
- domain error types (Error() untuk memenuhi interface error)
- repo return NotFoundError
- usecase return ValidationError
- handler mapping error -> response details
- test tambah case untuk medicine_id spasi

## File yang berubah / ditambah
### Transaction
- [MOD] services/transaction/internal/domain/errors.go
- [MOD] services/transaction/internal/repository/stock_memory_repo.go
- [MOD] services/transaction/internal/usecase/stock_usecase.go
- [MOD] services/transaction/internal/httpapi/handler/stock.go
- [MOD] services/transaction/internal/httpapi/handler/stock_test.go