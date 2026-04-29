$ErrorActionPreference = "Stop"

Write-Host "[finpharm] starting Alertmanager..." -ForegroundColor Cyan
docker compose -f docker-compose.alertmanager.yml up -d

if ($LASTEXITCODE -ne 0) {
    throw "failed to start Alertmanager"
}

Write-Host "[finpharm] Alertmanager started at http://localhost:9093" -ForegroundColor Green