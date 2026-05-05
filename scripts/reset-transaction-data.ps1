if ([string]::IsNullOrEmpty($env:DB_HOST)) { $env:DB_HOST = "127.0.0.1" }
if ([string]::IsNullOrEmpty($env:DB_PORT)) { $env:DB_PORT = "55432" }
if ([string]::IsNullOrEmpty($env:DB_USER)) { $env:DB_USER = "finpharm" }
if ([string]::IsNullOrEmpty($env:DB_PASSWORD)) { $env:DB_PASSWORD = "finpharm" }
if ([string]::IsNullOrEmpty($env:DB_NAME)) { $env:DB_NAME = "transaction_db" }
if ([string]::IsNullOrEmpty($env:DB_SSLMODE)) { $env:DB_SSLMODE = "disable" }

go run .\scripts\reset_transaction_data.go --yes