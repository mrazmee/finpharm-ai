param(
    [Parameter(Mandatory = $true)]
    [string]$Query,

    [int]$TopK = 5,

    [double]$MinScore = 0.45
)

$ErrorActionPreference = "Stop"

Write-Host "[finpharm] running knowledge query..." -ForegroundColor Cyan
go run ./services/knowledge/cmd/query -q "$Query" -k $TopK -min-score $MinScore

if ($LASTEXITCODE -ne 0) {
    throw "knowledge query failed"
}

Write-Host "[finpharm] knowledge query complete" -ForegroundColor Green