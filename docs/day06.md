# Day 06 — Feature Flag Debug (APP_ENV) + Tests

## Tujuan
- Debug endpoint hanya aktif di local/dev
- Prod harus 404

## Yang dibangun / diubah
- APP_ENV + IsDebugEnabled()
- Router conditional register debug routes
- Test debug enabled/disabled

## Konsep yang dipelajari
- Feature flag untuk safety (debug tidak bocor ke prod)

## File yang berubah / ditambah
### Gateway
- [MOD] services/gateway/internal/config/config.go
- [MOD] services/gateway/internal/httpapi/router.go
- [ADD] services/gateway/internal/httpapi/handler/debug_proxy_test.go

### Transaction
- [MOD] services/transaction/internal/config/config.go
- [MOD] services/transaction/internal/httpapi/router.go
- [ADD] services/transaction/internal/httpapi/handler/debug_test.go
- [MOD] services/transaction/internal/httpapi/handler/stock_test.go (gunakan router + middleware)

## Cara verifikasi
- APP_ENV=local -> debug ok
- APP_ENV=prod -> debug 404
- go test ./... -v