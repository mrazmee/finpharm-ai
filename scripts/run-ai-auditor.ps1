if ([string]::IsNullOrEmpty($env:APP_ENV)) { $env:APP_ENV = "local" }
if ([string]::IsNullOrEmpty($env:PORT)) { $env:PORT = "8083" }

if ([string]::IsNullOrEmpty($env:READ_TIMEOUT_MS)) { $env:READ_TIMEOUT_MS = "5000" }
if ([string]::IsNullOrEmpty($env:WRITE_TIMEOUT_MS)) { $env:WRITE_TIMEOUT_MS = "5000" }
if ([string]::IsNullOrEmpty($env:IDLE_TIMEOUT_MS)) { $env:IDLE_TIMEOUT_MS = "30000" }
if ([string]::IsNullOrEmpty($env:SHUTDOWN_TIMEOUT_MS)) { $env:SHUTDOWN_TIMEOUT_MS = "7000" }

# Day 28 scaffold: belum pakai provider AI sungguhan.
# Env ini disiapkan dari sekarang agar Day 29 lebih mudah diintegrasikan.
if ([string]::IsNullOrEmpty($env:GEMINI_API_KEY)) { $env:GEMINI_API_KEY = "" }

go run .\services\ai-auditor\cmd\api