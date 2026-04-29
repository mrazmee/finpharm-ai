$ErrorActionPreference = "Stop"

Write-Host "[finpharm] stopping Alertmanager..." -ForegroundColor Yellow
docker compose -f docker-compose.alertmanager.yml down

if ($LASTEXITCODE -ne 0) {
    throw "failed to stop Alertmanager"
}

Write-Host "[finpharm] Alertmanager stopped" -ForegroundColor Green