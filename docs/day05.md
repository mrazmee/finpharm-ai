# Day 05 — Testing + Debug Sleep Endpoint

## Tujuan
- Buat endpoint debug untuk simulasi delay
- Tambah unit test untuk handler/proxy

## Yang dibangun / diubah
- /v1/debug/sleep di transaction
- gateway proxy untuk debug sleep
- tests menggunakan httptest

## Konsep yang dipelajari
- Unit test Go + httptest
- Avoid import cycle (pakai package xxx_test)

## File yang berubah / ditambah
### Transaction
- [MOD] services/transaction/internal/httpapi/router.go
- [ADD] services/transaction/internal/httpapi/handler/debug.go
- [ADD] services/transaction/internal/httpapi/handler/stock_test.go (atau versi awal test)

### Gateway
- [MOD] services/gateway/internal/httpapi/router.go
- [ADD] services/gateway/internal/httpapi/handler/debug_proxy.go
- [ADD] services/gateway/internal/httpapi/handler/stock_proxy_test.go

## Cara verifikasi
- GET http://localhost:8080/v1/debug/sleep?ms=1000 -> ok
- go test ./... -v