if ([string]::IsNullOrEmpty($env:APP_ENV)) { $env:APP_ENV = "local" }
if ([string]::IsNullOrEmpty($env:PORT)) { $env:PORT = "8081" }

if ([string]::IsNullOrEmpty($env:READ_TIMEOUT_MS)) { $env:READ_TIMEOUT_MS = "5000" }
if ([string]::IsNullOrEmpty($env:WRITE_TIMEOUT_MS)) { $env:WRITE_TIMEOUT_MS = "5000" }
if ([string]::IsNullOrEmpty($env:IDLE_TIMEOUT_MS)) { $env:IDLE_TIMEOUT_MS = "30000" }
if ([string]::IsNullOrEmpty($env:SHUTDOWN_TIMEOUT_MS)) { $env:SHUTDOWN_TIMEOUT_MS = "7000" }

if ([string]::IsNullOrEmpty($env:INVENTORY_BASE_URL)) { $env:INVENTORY_BASE_URL = "http://localhost:8082" }
if ([string]::IsNullOrEmpty($env:AI_AUDITOR_BASE_URL)) { $env:AI_AUDITOR_BASE_URL = "http://localhost:8083" }
if ([string]::IsNullOrEmpty($env:AI_AUDITOR_TIMEOUT_MS)) { $env:AI_AUDITOR_TIMEOUT_MS = "5000" }

if ([string]::IsNullOrEmpty($env:DB_HOST)) { $env:DB_HOST = "127.0.0.1" }
if ([string]::IsNullOrEmpty($env:DB_PORT)) { $env:DB_PORT = "55432" }
if ([string]::IsNullOrEmpty($env:DB_USER)) { $env:DB_USER = "finpharm" }
if ([string]::IsNullOrEmpty($env:DB_PASSWORD)) { $env:DB_PASSWORD = "finpharm" }
if ([string]::IsNullOrEmpty($env:DB_NAME)) { $env:DB_NAME = "transaction_db" }
if ([string]::IsNullOrEmpty($env:DB_SSLMODE)) { $env:DB_SSLMODE = "disable" }

go run .\services\transaction\cmd\api