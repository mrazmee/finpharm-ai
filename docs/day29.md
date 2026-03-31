# Day 29 — Gemini Integration with Safe Fallback

## Tujuan

- Menghubungkan `ai-auditor` ke Gemini sebagai primary provider
- Menjaga service tetap hidup walaupun API key belum ada atau provider gagal
- Menjaga contract response tetap stabil:
  - `decision`
  - `risk_score`
  - `reason`
  - `provider`
  - `model`

---

# Prinsip Desain Hari Ini

Day 29 memakai pola dua provider:

1. **Primary provider**: Gemini
2. **Fallback provider**: safe fallback

Kalau Gemini berhasil:
- response datang dari provider `gemini`

Kalau Gemini gagal:
- service tidak crash
- endpoint tetap balas `200`
- response datang dari provider `fallback`

## Kenapa Fallback Default-nya REVIEW, Bukan APPROVED?

Karena domain project ini menyentuh:
- transaksi
- stok
- farmasi
- audit

Kalau provider AI unavailable, pendekatan yang lebih aman adalah **safe-review**, bukan auto-approve.

Itu sebabnya default:
- `decision = REVIEW`
- `provider = fallback`

---

# Config Baru

- `AUDIT_PROVIDER` default `gemini`
- `AUDIT_FAIL_OPEN` default `false`
- `GEMINI_API_KEY`
- `GOOGLE_API_KEY` (opsional alternatif)
- `GEMINI_MODEL` default `gemini-2.5-flash`
- `GEMINI_TIMEOUT_MS` default `3000`

---

# Provider Yang Ada Sekarang

- `GeminiProvider`
- `SafeFallbackProvider`
- `RuleBasedProvider` (masih dipertahankan kalau mau pakai mode mock)

---

# Contract Tetap Sama

Baik Gemini maupun fallback mengembalikan schema yang sama:

```json
{
  "data": {
    "decision": "APPROVED",
    "risk_score": 0.12,
    "reason": "....",
    "provider": "gemini",
    "model": "gemini-2.5-flash"
  },
  "request_id": "..."
}
```

Fallback juga memakai format yang sama, hanya nilainya berbeda.

---

# File Yang Berubah / Ditambah

## AI Auditor Service

```text
[MOD] services/ai-auditor/internal/config/config.go
[MOD] services/ai-auditor/internal/domain/audit.go
[MOD] services/ai-auditor/internal/usecase/audit_usecase.go
[MOD] services/ai-auditor/internal/usecase/audit_usecase_test.go
[MOD] services/ai-auditor/cmd/api/main.go
```

## Provider Baru

```text
[ADD] services/ai-auditor/internal/provider/gemini_provider.go
[ADD] services/ai-auditor/internal/provider/rule_based_provider.go
[ADD] services/ai-auditor/internal/provider/safe_fallback_provider.go
```

## Script & Docs

```text
[MOD] scripts/run-ai-auditor.ps1
[ADD] docs/day29.md
```

---

# Cara Verifikasi

## 1. Set API Key Lokal

Contoh PowerShell:

```powershell
$env:GEMINI_API_KEY = "YOUR_REAL_KEY"
```

---

## 2. Jalankan Service

```powershell
.\scripts\run-ai-auditor.ps1
```

---

## 3. Cek Audit Normal

```bash
curl -i -X POST http://localhost:8083/v1/audit/transaction \
  -H "Content-Type: application/json" \
  -d "{\"transaction_id\":\"TXN-20260322130000-AAAA1111\",\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2}]}"
```

Expected:

- status `200 OK`
- provider = `gemini` (bila Gemini berhasil) atau `fallback` (bila Gemini gagal)

---

## 4. Cek Fallback Saat API Key Kosong

Jalankan shell baru tanpa env key, lalu:

```powershell
Remove-Item Env:GEMINI_API_KEY -ErrorAction SilentlyContinue
.\scripts\run-ai-auditor.ps1
```

Kemudian panggil endpoint audit.

Expected:

- service tetap hidup
- response tetap `200 OK`
- provider = `fallback`
- default decision aman = `REVIEW`

---

## 5. Jalankan Test

```bash
go test ./services/ai-auditor/... -count=1 -v
```

---

# Self-Review

- Kenapa Gemini dijadikan primary provider, bukan satu-satunya provider?
- Kenapa fallback tetap mengembalikan contract yang sama?
- Kenapa fallback default lebih aman memakai `REVIEW`?
- Kenapa `AUDIT_FAIL_OPEN` dibuat config, walaupun default-nya false?
- Kenapa `RuleBasedProvider` masih berguna walaupun Gemini sudah aktif?

---

# Verifikasi Akhir

Sebelum run, set key-mu di shell lokal dulu, lalu jalankan service:

```powershell
$env:GEMINI_API_KEY = "YOUR_REAL_KEY"
.\scripts\run-ai-auditor.ps1
```

Jalankan test:

```bash
go test ./services/ai-auditor/... -count=1 -v
```

Lalu coba lakukan smoke test:

```bash
curl -i http://localhost:8083/health

curl -i -X POST http://localhost:8083/v1/audit/transaction ^
  -H "Content-Type: application/json" ^
  -d "{\"transaction_id\":\"TXN-20260331100000-AAAA1111\",\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2}]}"

curl -i -X POST http://localhost:8083/v1/audit/transaction ^
  -H "Content-Type: application/json" ^
  -d "{\"transaction_id\":\"TXN-20260331100000-BBBB2222\",\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":2}]}"
```