# Ambil token staff
$staffResp = Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/v1/auth/token" `
  -ContentType "application/json" `
  -Body '{"user_id":"staff-001","role":"staff"}'

$staffToken = $staffResp.data.access_token

# Ambil token supervisor
$supervisorResp = Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/v1/auth/token" `
  -ContentType "application/json" `
  -Body '{"user_id":"supervisor-001","role":"supervisor"}'

$supervisorToken = $supervisorResp.data.access_token

Write-Host "Traffic generation started..."

for ($i = 1; $i -le 15; $i++) {
    $traceId = "traffic-demo-$i"
    $idemKey = "idem-traffic-$i-$(Get-Date -Format 'yyyyMMddHHmmssfff')"

    # GET medicines
    Invoke-WebRequest -Method Get `
      -Uri "http://localhost:8080/v1/medicines?limit=2&offset=0" `
      -Headers @{
        Authorization = "Bearer $staffToken"
        "X-Trace-Id"  = $traceId
      } | Out-Null

    # POST stock check
    Invoke-WebRequest -Method Post `
      -Uri "http://localhost:8080/v1/stock/check" `
      -ContentType "application/json" `
      -Headers @{
        Authorization = "Bearer $staffToken"
        "X-Trace-Id"  = $traceId
      } `
      -Body '{"medicine_id":"PARA500","qty":1}' | Out-Null

    # POST transaction (unique idempotency key)
    Invoke-WebRequest -Method Post `
      -Uri "http://localhost:8080/v1/transactions" `
      -ContentType "application/json" `
      -Headers @{
        Authorization      = "Bearer $staffToken"
        "X-Trace-Id"       = $traceId
        "Idempotency-Key"  = $idemKey
      } `
      -Body '{"items":[{"medicine_id":"PARA500","qty":1}]}' | Out-Null

    # GET transactions as supervisor
    Invoke-WebRequest -Method Get `
      -Uri "http://localhost:8080/v1/transactions?limit=5&offset=0" `
      -Headers @{
        Authorization = "Bearer $supervisorToken"
        "X-Trace-Id"  = $traceId
      } | Out-Null

    Write-Host "Batch $i done"
    Start-Sleep -Milliseconds 500
}

Write-Host "Traffic generation finished."