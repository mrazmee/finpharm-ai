# Day 04 — Config + http.Server timeouts + Graceful Shutdown

## Tujuan
- Pindah dari router.Run ke http.Server
- Timeout server aktif (read/write/idle)
- Graceful shutdown (SIGINT/SIGTERM)

## Yang dibangun / diubah
- Config loader dari env
- Shutdown timeout

## Konsep yang dipelajari
- Inbound server timeouts vs outbound client timeouts
- Shutdown yang aman untuk microservices

## File yang berubah / ditambah
### Gateway
- [ADD] services/gateway/internal/config/config.go
- [MOD] services/gateway/cmd/api/main.go
- [MOD] services/gateway/internal/httpapi/router.go
- [MOD] services/gateway/internal/httpapi/handler/stock_proxy.go (baseURL inject)

### Transaction
- [ADD] services/transaction/internal/config/config.go
- [MOD] services/transaction/cmd/api/main.go
- [MOD] services/transaction/internal/httpapi/router.go

## Cara verifikasi
- Ctrl+C -> log server_shutdown_signal lalu server_stopped