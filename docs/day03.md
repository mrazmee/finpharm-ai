# Day 03 — Request-ID + Structured Logging + Standard Error

## Tujuan
- Propagate X-Request-ID gateway -> transaction
- Standard error format berisi request_id
- Structured logging (JSON)

## Yang dibangun / diubah
- Middleware request-id + logger
- Error response helper
- Header X-From-Gateway untuk tracing

## Konsep yang dipelajari
- Observability basic di microservices

## File yang berubah / ditambah
### Gateway
- [ADD] services/gateway/internal/httpapi/middleware/request_id.go
- [ADD] services/gateway/internal/httpapi/middleware/logger.go
- [ADD] services/gateway/internal/httpapi/handler/response.go
- [MOD] services/gateway/internal/httpapi/router.go
- [MOD] services/gateway/internal/httpapi/handler/stock_proxy.go
- [MOD] services/gateway/cmd/api/main.go (slog JSON)

### Transaction
- [ADD] services/transaction/internal/httpapi/middleware/request_id.go
- [ADD] services/transaction/internal/httpapi/middleware/logger.go
- [ADD] services/transaction/internal/httpapi/handler/response.go
- [MOD] services/transaction/internal/httpapi/router.go
- [MOD] services/transaction/internal/httpapi/handler/stock.go
- [MOD] services/transaction/cmd/api/main.go (slog JSON)

## Cara verifikasi
- Request tanpa X-Request-ID -> response punya X-Request-ID
- Log gateway & transaction request_id harus sama