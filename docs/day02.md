# Day 02 — Contract /v1/stock/check + Gateway Forward

## Tujuan
- Transaction punya `POST /v1/stock/check`
- Gateway forward endpoint yang sama ke Transaction

## Yang dibangun / diubah
- In-memory stock di transaction
- Gateway proxy call via HTTP

## Konsep yang dipelajari
- JSON binding & validation (Gin)
- Microservices sync call (HTTP)

## File yang berubah / ditambah
### Transaction
- [MOD] services/transaction/internal/httpapi/router.go
- [ADD] services/transaction/internal/httpapi/handler/stock.go

### Gateway
- [MOD] services/gateway/internal/httpapi/router.go
- [ADD] services/gateway/internal/httpapi/handler/stock_proxy.go

## Cara verifikasi
- POST http://localhost:8080/v1/stock/check
  - body: {"medicine_id":"PARA500","qty":10}