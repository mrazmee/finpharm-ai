if ([string]::IsNullOrEmpty($env:APP_ENV)) { $env:APP_ENV = "local" }
if ([string]::IsNullOrEmpty($env:PORT)) { $env:PORT = "8080" }

if ([string]::IsNullOrEmpty($env:READ_TIMEOUT_MS)) { $env:READ_TIMEOUT_MS = "5000" }
if ([string]::IsNullOrEmpty($env:WRITE_TIMEOUT_MS)) { $env:WRITE_TIMEOUT_MS = "5000" }
if ([string]::IsNullOrEmpty($env:IDLE_TIMEOUT_MS)) { $env:IDLE_TIMEOUT_MS = "30000" }
if ([string]::IsNullOrEmpty($env:SHUTDOWN_TIMEOUT_MS)) { $env:SHUTDOWN_TIMEOUT_MS = "7000" }

if ([string]::IsNullOrEmpty($env:INVENTORY_BASE_URL)) { $env:INVENTORY_BASE_URL = "http://localhost:8082" }
if ([string]::IsNullOrEmpty($env:TRANSACTION_BASE_URL)) { $env:TRANSACTION_BASE_URL = "http://localhost:8081" }

if ([string]::IsNullOrEmpty($env:AUTH_ENABLED)) { $env:AUTH_ENABLED = "true" }
if ([string]::IsNullOrEmpty($env:JWT_SECRET)) { $env:JWT_SECRET = "finpharm-local-secret" }
if ([string]::IsNullOrEmpty($env:JWT_ISSUER)) { $env:JWT_ISSUER = "finpharm-gateway" }
if ([string]::IsNullOrEmpty($env:JWT_EXPIRE_MINUTES)) { $env:JWT_EXPIRE_MINUTES = "60" }

go run .\services\gateway\cmd\api