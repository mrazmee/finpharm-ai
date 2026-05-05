# Day 07 — Repo Ergonomics + CI

## Tujuan
- Repo lebih enak dipakai: scripts, README, log path rapi
- Setup CI GitHub Actions

## Yang dibangun / diubah
- Logger fallback: kalau FullPath kosong (404), pakai URL.Path
- Script PowerShell run gateway/transaction (PS 5.1 friendly)
- README lengkap
- GitHub Actions CI (go test ./...)

## File yang berubah / ditambah
### Gateway
- [MOD] services/gateway/internal/httpapi/middleware/logger.go
### Transaction
- [MOD] services/transaction/internal/httpapi/middleware/logger.go

### Repo
- [ADD] scripts/run-gateway.ps1
- [ADD] scripts/run-transaction.ps1
- [MOD] README.md
- [ADD] .github/workflows/ci.yml

## Cara verifikasi
- scripts jalan
- go test ./... -v
- PR/push main -> CI jalan