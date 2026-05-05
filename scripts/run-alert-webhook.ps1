$ErrorActionPreference = "Stop"

Write-Host "[finpharm] starting local alert webhook on http://localhost:18080" -ForegroundColor Cyan
go run ./cmd/alert-webhook