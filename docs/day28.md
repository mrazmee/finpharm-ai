# Day 28 — AI Auditor Service Scaffold

## Tujuan

- Menambahkan service baru bernama **AI Auditor Service**
- Membuat fondasi service AI yang bisa dijalankan terpisah seperti service lain
- Menetapkan contract awal endpoint audit transaction
- Menjaga scope Day 28 tetap fokus pada **scaffold + contract**, belum integrasi Gemini sungguhan

---

# Kenapa Belum Langsung Pakai Gemini?

Karena sebelum integrasi provider AI, kita perlu lebih dulu menyiapkan:

- service boundary yang jelas
- endpoint contract yang stabil
- test dasar
- config dan script run

Kalau ini belum ada, Day 29 akan terlalu bercampur antara:

- desain service AI
- implementasi provider eksternal
- debugging HTTP / config / env

---

# File Yang Berubah / Ditambah

Service baru: `services/ai-auditor`

## AI Auditor Service

```text
[ADD] cmd/api/main.go
[ADD] internal/config/config.go
[ADD] internal/domain/errors.go
[ADD] internal/domain/audit.go
[ADD] internal/usecase/audit_usecase.go
[ADD] internal/usecase/audit_usecase_test.go
[ADD] internal/httpapi/router.go
[ADD] internal/httpapi/middleware/request_id.go
[ADD] internal/httpapi/middleware/logger.go
[ADD] internal/httpapi/handler/response.go
[ADD] internal/httpapi/handler/health.go
[ADD] internal/httpapi/handler/audit.go
[ADD] internal/httpapi/handler/audit_test.go
```

## Scripts

```text
[ADD] scripts/run-ai-auditor.ps1
```

---

# Endpoint Baru Hari Ini

## 1. Health

```text
GET /
GET /health
```

## 2. Audit Dummy

```text
POST /v1/audit/transaction
```

**Contract request:**
```json
{
  "transaction_id": "TXN-20260322130000-AAAA1111",
  "items": [
    {
      "medicine_id": "PARA500",
      "qty": 2
    }
  ]
}
```

**Contract response:**
```json
{
  "data": {
    "decision": "APPROVED",
    "risk_score": 0.12,
    "reason": "mock audit result: no suspicious pattern detected",
    "provider": "mock",
    "model": "rule-based-v1"
  },
  "request_id": "..."
}
```

---

# Mock Logic Hari Ini

Day 28 belum memanggil LLM sungguhan. Sebagai gantinya, kita pakai rule-based mock sederhana:

- jika medicine mengandung `OBATKERAS` → `REVIEW`
- jika qty besar (`>= 20`) → `REVIEW`
- selain itu → `APPROVED`

Tujuannya bukan untuk fraud detection final, tapi untuk memberi:

- contract yang nyata
- response yang bisa dites
- dasar yang mudah dihubungkan ke Gemini di Day 29

## Kenapa Decision Pakai APPROVED / REVIEW?

Karena service ini adalah auditor, bukan transaction state manager.

Jadi keputusan AI di Day 28 masih dipisahkan dulu dari status transaksi. Nanti saat integration phase:

- Transaction Service bisa memetakan hasil audit ke lifecycle domain-nya sendiri
- misalnya ke `APPROVED`, `PENDING_REVIEW`, atau `FLAGGED`

---

# Cara Verifikasi

Script run baru menggunakan default port **8083**.

## 1. Jalankan Service

```bash
.\scripts\run-ai-auditor.ps1
```

---

## 2. Cek Health

```bash
curl -i http://localhost:8083/health
```

Expected:

- status `200 OK`
- body berisi `service = ai-auditor`

---

## 3. Audit Transaksi Normal

```bash
curl -i -X POST http://localhost:8083/v1/audit/transaction \
  -H "Content-Type: application/json" \
  -d "{\"transaction_id\":\"TXN-20260322130000-AAAA1111\",\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2}]}"
```

Expected:

- status `200 OK`
- decision = `APPROVED`

---

## 4. Audit Transaksi High-Risk Mock

```bash
curl -i -X POST http://localhost:8083/v1/audit/transaction \
  -H "Content-Type: application/json" \
  -d "{\"transaction_id\":\"TXN-20260322130000-BBBB2222\",\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":2}]}"
```

Expected:

- status `200 OK`
- decision = `REVIEW`
- `risk_score` tinggi

---

## 5. Coba Request Invalid

```bash
curl -i -X POST http://localhost:8083/v1/audit/transaction \
  -H "Content-Type: application/json" \
  -d "{\"transaction_id\":\"\",\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2}]}"
```

Expected:

- status `400 Bad Request`
- field = `transaction_id`

---

## 6. Jalankan Semua Test

```bash
go test ./services/ai-auditor/... -count=1 -v
```

Expected:

- semua test pass

---

# Self-Review

- Kenapa Day 28 cukup membuat service AI berdiri sendiri dulu?
- Kenapa contract audit perlu dibuat sebelum integrasi provider AI?
- Kenapa mock rule-based tetap berguna walaupun belum pakai Gemini?
- Kenapa decision audit dipisahkan dulu dari status transaksi?
- Kenapa `GEMINI_API_KEY` boleh disiapkan dari sekarang walaupun belum dipakai?

---

# Verifikasi Akhir

Jalankan test:

```bash
go test ./services/ai-auditor/... -count=1 -v
```

Lalu lakukan smoke test:

```bash
curl -i http://localhost:8083/health

curl -i -X POST http://localhost:8083/v1/audit/transaction ^
  -H "Content-Type: application/json" ^
  -d "{\"transaction_id\":\"TXN-20260322130000-AAAA1111\",\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2}]}"

curl -i -X POST http://localhost:8083/v1/audit/transaction ^
  -H "Content-Type: application/json" ^
  -d "{\"transaction_id\":\"TXN-20260322130000-BBBB2222\",\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":2}]}"
```

Pastikan:

- service health merespons dengan benar
- mock logic membedakan transaksi normal dan high-risk