$ErrorActionPreference = "Stop"

Write-Host "[finpharm] running knowledge migrations..." -ForegroundColor Cyan
go run ./services/knowledge/cmd/migrate

if ($LASTEXITCODE -ne 0) {
    throw "knowledge migration failed"
}

Write-Host "[finpharm] knowledge migrations complete" -ForegroundColor Green