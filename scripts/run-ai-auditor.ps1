if ([string]::IsNullOrEmpty($env:APP_ENV)) { $env:APP_ENV = "local" }
if ([string]::IsNullOrEmpty($env:PORT)) { $env:PORT = "8083" }

if ([string]::IsNullOrEmpty($env:READ_TIMEOUT_MS)) { $env:READ_TIMEOUT_MS = "5000" }
if ([string]::IsNullOrEmpty($env:WRITE_TIMEOUT_MS)) { $env:WRITE_TIMEOUT_MS = "5000" }
if ([string]::IsNullOrEmpty($env:IDLE_TIMEOUT_MS)) { $env:IDLE_TIMEOUT_MS = "30000" }
if ([string]::IsNullOrEmpty($env:SHUTDOWN_TIMEOUT_MS)) { $env:SHUTDOWN_TIMEOUT_MS = "7000" }

if ([string]::IsNullOrEmpty($env:AUDIT_PROVIDER)) { $env:AUDIT_PROVIDER = "gemini" }
if ([string]::IsNullOrEmpty($env:AUDIT_FAIL_OPEN)) { $env:AUDIT_FAIL_OPEN = "false" }

if ([string]::IsNullOrEmpty($env:GEMINI_MODEL)) { $env:GEMINI_MODEL = "gemini-2.5-flash" }
if ([string]::IsNullOrEmpty($env:GEMINI_TIMEOUT_MS)) { $env:GEMINI_TIMEOUT_MS = "3000" }

# Set your key in the current shell before running, for example:
# $env:GEMINI_API_KEY = "YOUR_REAL_KEY"
# $env:AUDIT_PROVIDER="mock"
# Do not commit real keys to the repository.

go run .\services\ai-auditor\cmd\api