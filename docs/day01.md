# Day 01 — Microservices Bootstrap (Gin)

## Tujuan
- Menjalankan 2 service terpisah: gateway & transaction
- Punya endpoint dasar `/` dan `/health`

## Yang dibangun / diubah
- Membuat struktur repo microservices monorepo
- Menjalankan gateway di 8080 dan transaction di 8081

## Konsep yang dipelajari
- Microservices = proses/binary terpisah
- Struktur Go: `cmd/api` + `internal`

## File yang berubah / ditambah
### Gateway
- [ADD] services/gateway/cmd/api/main.go
- [ADD] services/gateway/internal/httpapi/router.go
- [ADD] services/gateway/internal/httpapi/handler/health.go

### Transaction
- [ADD] services/transaction/cmd/api/main.go
- [ADD] services/transaction/internal/httpapi/router.go
- [ADD] services/transaction/internal/httpapi/handler/health.go

## Cara verifikasi
- Run:
  - `go run ./services/transaction/cmd/api`
  - `go run ./services/gateway/cmd/api`
- Test:
  - GET http://localhost:8080/health
  - GET http://localhost:8081/health