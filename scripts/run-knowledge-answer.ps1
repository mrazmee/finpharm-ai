param(
    [Parameter(Mandatory = $true)]
    [string]$Query,

    [int]$TopK = 5,

    [double]$MinScore = 0.45
)

$ErrorActionPreference = "Stop"

Write-Host "[finpharm] running knowledge answer..." -ForegroundColor Cyan
go run ./services/knowledge/cmd/answer -q "$Query" -k $TopK -min-score $MinScore

if ($LASTEXITCODE -ne 0) {
    throw "knowledge answer failed"
}

Write-Host "[finpharm] knowledge answer complete" -ForegroundColor Green