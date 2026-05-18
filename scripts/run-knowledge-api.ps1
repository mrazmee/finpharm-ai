$ErrorActionPreference = "Stop"

Write-Host "[finpharm] running knowledge api..." -ForegroundColor Cyan
go run ./services/knowledge/cmd/api

if ($LASTEXITCODE -ne 0) {
    throw "knowledge api failed"
}

Write-Host "[finpharm] knowledge api stopped" -ForegroundColor Green