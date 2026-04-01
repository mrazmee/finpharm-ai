# Day 32 — Persist Audit Metadata in Transaction DB

## Tujuan

- Menyimpan hasil audit AI ke database `Transaction Service`
- Menampilkan metadata audit di response API transaction
- Membuat transaksi lebih traceable untuk kebutuhan review, debugging, dan portfolio

---

# Kenapa Day 32 Penting?

Sampai Day 31, AI audit sudah mempengaruhi status transaksi:
- `APPROVED`
- `PENDING_REVIEW`
- `FLAGGED`

Tetapi hasil audit itu masih sifatnya “transient” secara sistem:
- status final terlihat
- namun alasan audit, provider, model, dan risk score belum tersimpan secara permanen

Setelah Day 32, transaction punya jejak audit yang lebih lengkap.

---

# Metadata Audit Yang Disimpan

Kolom baru di tabel `transactions`:

- `audit_decision`
- `audit_risk_score`
- `audit_reason`
- `audit_provider`
- `audit_model`
- `audited_at`

---

# Response API Setelah Day 32

Sekarang response transaction bisa punya blok baru:

```json
{
  "id": "TXN-20260401100000-AAAA1111",
  "status": "FLAGGED",
  "items": [
    {
      "medicine_id": "OBATKERAS-X",
      "qty": 2
    }
  ],
  "audit": {
    "decision": "REVIEW",
    "risk_score": 0.91,
    "reason": "high-risk medicine detected: OBATKERAS-X",
    "provider": "mock",
    "model": "rule-based-v1",
    "audited_at": "2026-04-01T10:00:05Z"
  },
  "created_at": "2026-04-01T10:00:00Z"
}
```

---

# Scope Day 32

Hari ini fokusnya:

- menambah kolom audit di DB transaction
- menyimpan hasil audit ke DB
- mengembalikan audit metadata di create/list transaction response
- menyimpan audit fallback saat AI unavailable

---

# Mapping Audit Persistence

## 1. AI Approve

Disimpan:
- decision = `APPROVED`
- `risk_score` dari AI
- `reason` dari AI
- provider/model dari AI
- `audited_at` = waktu audit

Lalu:
- stock dideduct
- status transaction = `APPROVED`

## 2. AI Review

Disimpan:
- decision = `REVIEW`
- `risk_score` dari AI
- `reason` dari AI
- provider/model dari AI
- `audited_at` = waktu audit

Lalu:
- status transaction = `PENDING_REVIEW` atau `FLAGGED`
- stock **tidak** dideduct

## 3. AI Unavailable / Error

Disimpan audit fallback seperti:
- decision = `REVIEW`
- `risk_score` = `0.50`
- reason = `ai auditor unavailable; manual review required`
- provider = `system`
- model = `fallback-pending-review`

Lalu:
- status transaction = `PENDING_REVIEW`

---

# Kenapa Fallback Audit Tetap Disimpan?

Karena untuk audit trail, penting juga tahu:
- transaksi ini tidak lolos ke audit AI normal
- kenapa status-nya `PENDING_REVIEW`
- apakah karena AI menilai review atau karena AI unavailable

Kalau fallback tidak disimpan, reviewer hanya melihat status tanpa konteks.

---

# File Yang Berubah / Ditambah

## Transaction Service

```text
[REPLACE] services/transaction/internal/domain/transaction.go
[ADD]     services/transaction/migrations/000003_add_audit_columns_to_transactions.up.sql
[ADD]     services/transaction/migrations/000003_add_audit_columns_to_transactions.down.sql
[REPLACE] services/transaction/internal/repository/transaction_sqlx_repo.go
[REPLACE] services/transaction/internal/usecase/transaction_usecase.go
[REPLACE] services/transaction/internal/httpapi/handler/transaction.go
[REPLACE] services/transaction/internal/usecase/transaction_usecase_test.go
```

## Docs

```text
[REPLACE] docs/day32.md
```

---

# Langkah Sebelum Run

## 1. Jalankan Migration Transaction Baru

Gunakan script migration transaction milikmu seperti biasa. Migration Day 32 yang benar untuk repo ini adalah:
- `000003_add_audit_columns_to_transactions.up.sql`
- `000003_add_audit_columns_to_transactions.down.sql`

## 2. Restart Service

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-transaction.ps1
.\scripts\run-gateway.ps1
```

---

# Cara Verifikasi

## 1. Transaksi Low-Risk

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-601" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:
- status `APPROVED`
- ada blok audit
- provider/model terisi

---

## 2. Transaksi High-Risk

Untuk deterministic test, jalankan ai-auditor dengan mock:

```powershell
$env:AUDIT_PROVIDER = "mock"
.\scripts\run-ai-auditor.ps1
```

Lalu jalankan transaksi:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-602" \
  -d "{\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":2}]}"
```

Expected:
- status `FLAGGED`
- ada blok audit
- reason / provider / model terisi

---

## 3. AI Unavailable

Matikan service `ai-auditor`, lalu:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-603" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:
- status `PENDING_REVIEW`
- ada blok audit
- provider = `system`
- model = `fallback-pending-review`

---

## 4. List Transactions

```bash
curl -i "http://localhost:8080/v1/transactions?status=flagged"
curl -i "http://localhost:8080/v1/transactions?status=pending_review"
curl -i "http://localhost:8080/v1/transactions?status=approved"
```

Expected:
- item transaction sekarang membawa blok audit jika sudah diaudit

---

## 5. Jalankan Semua Test

```bash
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./... -count=1 -v
```

---

# Self-Review

- Kenapa status saja belum cukup tanpa metadata audit?
- Kenapa `audit_reason` perlu disimpan, bukan hanya ditampilkan sekali?
- Kenapa fallback audit juga perlu dipersist?
- Kenapa `audited_at` penting untuk traceability?
- Apa manfaat portofolio ketika API response sudah membawa `audit.provider` dan `audit.model`?