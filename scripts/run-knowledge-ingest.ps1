$ErrorActionPreference = "Stop"

Write-Host "[finpharm] running knowledge ingestion..." -ForegroundColor Cyan
go run ./services/knowledge/cmd/ingest

if ($LASTEXITCODE -ne 0) {
    throw "knowledge ingestion failed"
}

Write-Host "[finpharm] knowledge ingestion complete" -ForegroundColor Green