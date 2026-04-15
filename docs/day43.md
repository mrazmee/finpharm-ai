# Day 43 — Final Hardening II: Runtime Safety + Gateway Polish

## Tujuan

Menambahkan lapisan hardening runtime dasar agar gateway lebih aman, lebih production-like, dan lebih siap untuk diuji/didemokan.

---

## Fokus Day 43

- basic in-memory rate limiting di gateway
- config validation / fail-fast
- startup safety
- startup log yang lebih jelas
- dokumentasi hardening runtime

## Kenapa Day 43 Penting?

Project yang terasa production-like bukan hanya punya fitur, tetapi juga punya guardrail:
- tidak menerima config penting yang rusak secara diam-diam
- tidak membiarkan abuse request tanpa batas
- punya startup behaviour yang lebih aman
- memberi sinyal operasional yang lebih jelas di log

---

## Scope Day 43

Day 43 mencakup:
- basic in-memory rate limiting di gateway
- auth endpoint memiliki limit lebih ketat
- general endpoint memiliki limit umum
- config validation saat startup
- fail fast jika config penting tidak valid
- update docs hardening

**Scope Yang Belum Dikerjakan di Day 43:**
Supaya tetap lurus roadmap hardening, Day 43 belum mencakup: Redis/distributed rate limit, API gateway eksternal, WAF, global throttling lintas node, advanced abuse detection, atau full healthcheck orchestration lintas stack.

---

## File Yang Ditambahkan / Diubah

### Gateway Middleware
```text
[NEW] services/gateway/internal/httpapi/middleware/rate_limit_http.go
[NEW] services/gateway/internal/httpapi/middleware/rate_limit_http_test.go
```

### Gateway Config
```text
[REPLACE] services/gateway/internal/config/config.go
[NEW]     services/gateway/internal/config/config_test.go
```

### Gateway Startup
```text
[REPLACE] services/gateway/cmd/api/main.go
```

### Docs
```text
[NEW] docs/day43.md
```

---

## Pendekatan Rate Limiting Yang Dipakai

Agar aman terhadap struktur gateway yang sudah berjalan, rate limiting diterapkan sebagai **standard `net/http` wrapper** di level gateway app handler.

Keuntungan pendekatan ini:
- tidak terlalu mengganggu struktur router yang sudah ada
- mudah dipasang di gateway saat ini
- tetap bisa mengontrol request berdasarkan path

### Aturan Rate Limit

Saat ini ada dua kelompok utama:

**1. Auth endpoint**
- Path: `/v1/auth/token`
- Limit: lebih ketat (default `20 request / 60 detik / IP`)

**2. General endpoint**
- Path: Semua endpoint lain (medicines, stock check, transactions, debug sleep, dll)
- Limit: lebih longgar (default `60 request / 60 detik / IP`)

### Header Response Rate Limit

Saat request melewati gateway, response sekarang akan membawa header:
- `X-RateLimit-Limit`
- `X-RateLimit-Remaining`
- `X-RateLimit-Window-Seconds`

Kalau terkena rate limit:
- status `429 Too Many Requests`
- header `Retry-After`
- body error code: `RATE_LIMITED`

---

## Config Baru & Fail-Fast Validation

Environment variable baru di gateway:
```env
RATE_LIMIT_ENABLED=true
RATE_LIMIT_GENERAL_LIMIT=60
RATE_LIMIT_AUTH_LIMIT=20
RATE_LIMIT_WINDOW_SECONDS=60
```

### Config Validation

Gateway sekarang fail fast kalau config penting tidak valid. Contoh validasi:
- `PORT` tidak boleh kosong
- `INVENTORY_BASE_URL` dan `TRANSACTION_BASE_URL` tidak boleh kosong
- `timeout` harus > 0
- Jika auth aktif: `JWT_SECRET` dan `JWT_ISSUER` tidak boleh kosong, `JWT_EXPIRE_MINUTES` harus > 0
- Jika rate limit aktif: `limit` harus > 0, `window` harus > 0

**Tambahan hardening:** Di luar *local environment*, `JWT_SECRET` default (`dev-secret-change-me`) dianggap tidak valid.

### Cara Kerja Startup Baru

Saat gateway start:
1. Load config
2. Validate config
3. Kalau invalid → log `config_invalid` lalu exit
4. Kalau valid → lanjut start server

Ini lebih aman daripada membiarkan gateway jalan dengan config yang salah.

---

## Cara Test & Validasi Manual

### 1. Unit Test
```powershell
go test ./services/gateway/internal/httpapi/middleware/... -count=1 -v
go test ./services/gateway/internal/config/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./... -count=1 -v
```

### 2. Jalankan Gateway
```powershell
.\scripts\run-gateway.ps1
```

### 3. Pastikan Startup Log Memuat Info Baru
Expected log:
- `rate_limit_enabled`
- `rate_limit_general_limit`
- `rate_limit_auth_limit`
- `rate_limit_window_seconds`

### 4. Uji Auth Token Endpoint (Rate Limit Ketat)
Kirim request beberapa kali dengan cepat menggunakan PowerShell:

```powershell
1..25 | ForEach-Object {
  try {
    $resp = Invoke-WebRequest -UseBasicParsing -Method Post `
      -Uri "http://localhost:8080/v1/auth/token" `
      -ContentType "application/json" `
      -Body '{"user_id":"staff-001","role":"staff"}'
    $resp.StatusCode
  } catch {
    if ($_.Exception.Response) {
      [int]$_.Exception.Response.StatusCode
    } else {
      "ERROR"
    }
  }
}
```
**Expected:** request awal `200`, setelah limit tercapai mulai muncul `429`.

*Alternatif menggunakan cURL:*
```powershell
1..25 | ForEach-Object {
  curl -s -o NUL -w "%{http_code}`n" -X POST http://localhost:8080/v1/auth/token -H "Content-Type: application/json" -d "{\"user_id\":\"staff-001\",\"role\":\"staff\"}"
}
```

### 5. Uji General Endpoint (Rate Limit Longgar)

**Endpoint GET /v1/transactions:**
```powershell
$supervisorResp = Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/v1/auth/token" `
  -ContentType "application/json" `
  -Body '{"user_id":"supervisor-001","role":"supervisor"}'

$supervisorToken = $supervisorResp.data.access_token

1..70 | ForEach-Object {
  try {
    $resp = Invoke-WebRequest -UseBasicParsing -Method Get `
      -Uri "http://localhost:8080/v1/transactions?limit=5&offset=0" `
      -Headers @{ Authorization = "Bearer $supervisorToken" }
    Write-Output $resp.StatusCode
  } catch {
    if ($_.Exception.Response) {
      Write-Output ([int]$_.Exception.Response.StatusCode)
    } else {
      Write-Output "ERROR"
    }
  }
}
```

**Endpoint POST stock/check:**
```powershell
$staffResp = Invoke-RestMethod -Method Post `
  -Uri "http://localhost:8080/v1/auth/token" `
  -ContentType "application/json" `
  -Body '{"user_id":"staff-001","role":"staff"}'

$staffToken = $staffResp.data.access_token

1..70 | ForEach-Object {
  try {
    $resp = Invoke-WebRequest -UseBasicParsing -Method Post `
      -Uri "http://localhost:8080/v1/stock/check" `
      -Headers @{ Authorization = "Bearer $staffToken" } `
      -ContentType "application/json" `
      -Body '{"medicine_id":"PARA500","qty":1}'
    Write-Output $resp.StatusCode
  } catch {
    if ($_.Exception.Response) {
      Write-Output ([int]$_.Exception.Response.StatusCode)
    } else {
      Write-Output "ERROR"
    }
  }
}
```

### 6. Cek Response Header
- Saat request biasa akan ada: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Window-Seconds`.
- Saat blocked akan ada tambahan: `Retry-After`.

---

## Interpretasi Hasil

Kalau manual check berjalan:
- gateway sekarang punya guardrail dasar terhadap burst request
- config salah akan gagal sejak startup
- reviewer bisa melihat adanya lapisan hardening runtime, bukan hanya fitur

### Keterbatasan Implementasi Day 43
Penting dicatat:
- rate limiter saat ini in-memory
- key saat ini berbasis client IP
- tidak cocok untuk multi-instance distributed deployment
- restart process akan mereset counter

Itu masih acceptable untuk local hardening, portfolio, dan baseline production-minded design.

---

## Troubleshooting

**1. Semua request langsung kena 429**
- Cek: limit terlalu kecil, testing dari IP yang sama, atau worker/test tool sedang membanjiri endpoint.

**2. Auth endpoint cepat sekali kena limit**
- Itu bisa normal karena auth endpoint memang sengaja diberi limit lebih ketat.

**3. Gateway gagal start**
- Cek log: `config_invalid`. Biasanya ada env yang kosong / timeout invalid / secret tidak valid.

---

## Checklist Hardening Lanjutan

Bagian ini tetap ditahan untuk hardening berikutnya:
- [ ] distributed rate limit
- [ ] healthcheck endpoint dan readiness model yang lebih kuat
- [ ] config sanity check lintas service
- [ ] observability polish
- [ ] release readiness final

---

## Self-Review

- Kenapa auth endpoint perlu limit lebih ketat?
- Kenapa in-memory rate limit masih acceptable untuk tahap ini?
- Kenapa fail-fast config validation penting?
- Kenapa default dev secret tidak boleh lolos di environment non-local?
- Kenapa rate limit di Day 43 dipasang di gateway, bukan di semua service?