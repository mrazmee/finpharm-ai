# Day 31 — AI-Based Transaction Status Mapping

## Tujuan

- Meresmikan lifecycle transaction yang dipengaruhi hasil audit AI
- Memisahkan kasus:
  - review biasa
  - risiko tinggi
  - approval aman
- Mengurangi pemakaian status `PENDING` sebagai tempat semua kondisi review

---

# Scope Day 31

Hari ini fokusnya:

- `Transaction Service` tetap call `AI Auditor Service`
- hasil audit sekarang dipetakan ke status transaction yang lebih jelas:
  - `APPROVED`
  - `PENDING_REVIEW`
  - `FLAGGED`
- kasus AI error/unavailable dipetakan ke `PENDING_REVIEW`
- transaksi `FLAGGED` dan `PENDING_REVIEW` **belum deduct stock**
- transaction `APPROVED` tetap deduct stock

---

# Mapping Status Hari Ini

## 1. Audit APPROVED

Kalau AI audit mengembalikan `decision = APPROVED`:
- lanjut deduct stock
- update transaction jadi `APPROVED`

## 2. Audit REVIEW dengan Risk Score Sedang

Kalau AI audit mengembalikan `decision = REVIEW` dan `risk_score < 0.85`:
- transaction jadi `PENDING_REVIEW`
- stock belum dikurangi

## 3. Audit REVIEW dengan Risk Score Tinggi

Kalau AI audit mengembalikan `decision = REVIEW` dan `risk_score >= 0.85`:
- transaction jadi `FLAGGED`
- stock belum dikurangi

## 4. AI Auditor Error / Timeout / Unavailable

Kalau AI audit gagal:
- transaction jadi `PENDING_REVIEW`
- stock belum dikurangi

---

# Q&A Desain Status

## Kenapa Perlu FLAGGED?

Karena tidak semua hasil review itu setara. 

Ada kasus yang butuh review manual biasa, ada juga kasus yang sudah cukup kuat dianggap mencurigakan. Status `FLAGGED` memberi sinyal bahwa:
- transaksi lebih berisiko
- perlu perhatian lebih tinggi
- tidak seharusnya diperlakukan sama dengan review biasa

## Kenapa Threshold Pakai 0.85?

Karena kita butuh aturan sederhana yang eksplisit, mudah dites, dan mudah dijelaskan.
- `risk_score >= 0.85` → `FLAGGED`
- selain itu → `PENDING_REVIEW`

Ini bukan angka final production, tapi cukup bagus untuk phase belajar saat ini.

## Kenapa AI Error Dipetakan ke PENDING_REVIEW, Bukan FLAGGED?

Karena AI unavailable bukan berarti transaksi itu pasti berbahaya. Tapi transaksi juga belum aman untuk auto-approve. Jadi pilihan paling seimbang adalah `PENDING_REVIEW`.

---

# Status Yang Aktif Sekarang

Transaction sekarang punya lifecycle:

- `PENDING` (masih bisa muncul dari data lama)
- `PENDING_REVIEW`
- `APPROVED`
- `FLAGGED`
- `FAILED`

Target status final yang lebih bermakna untuk flow Day 31 ke depan adalah: `APPROVED`, `PENDING_REVIEW`, `FLAGGED`, `FAILED`.

---

# File Yang Berubah / Ditambah

## Transaction Service

```text
[REPLACE] services/transaction/internal/domain/transaction.go
[REPLACE] services/transaction/internal/usecase/transaction_usecase.go
[REPLACE] services/transaction/internal/usecase/transaction_usecase_test.go
[REPLACE] services/transaction/internal/httpapi/handler/transaction_test.go
```

## Docs

```text
[ADD] docs/day31.md
```

**Catatan:** Tidak ada migration baru hari ini karena status tetap disimpan sebagai string di kolom yang sama. Kita hanya menambah nilai status baru di level aplikasi.

---

# Cara Verifikasi

## 1. Low-Risk Transaction

Gunakan AI auditor normal.

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-501" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:
- status `201 Created`
- status `APPROVED`

---

## 2. High-Risk Transaction Deterministic

Untuk verifikasi deterministic, jalankan ai-auditor dengan mock provider:

```powershell
$env:AUDIT_PROVIDER = "mock"
.\scripts\run-ai-auditor.ps1
```

Lalu jalankan transaksi:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-502" \
  -d "{\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":2}]}"
```

Expected:
- status `201 Created`
- status `FLAGGED` (karena mock provider memberi `decision = REVIEW` dan `risk_score = 0.91`)

---

## 3. AI Unavailable

Matikan service `ai-auditor`, lalu:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-503" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:
- status `201 Created`
- status `PENDING_REVIEW`

---

## 4. Filter Status Baru

```bash
curl -i "http://localhost:8080/v1/transactions?status=flagged"
curl -i "http://localhost:8080/v1/transactions?status=pending_review"
```

Expected:
- filter `flagged` menampilkan transaksi high-risk
- filter `pending_review` menampilkan transaksi yang belum lolos audit / AI unavailable

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

- Kenapa `FLAGGED` perlu dipisahkan dari `PENDING_REVIEW`?
- Kenapa threshold 0.85 cukup masuk akal untuk tahap ini?
- Kenapa AI error dipetakan ke `PENDING_REVIEW`, bukan `APPROVED`?
- Kenapa transaksi `FLAGGED` tidak boleh mengurangi stock?
- Kenapa Day 31 tidak memerlukan migration database?

---

# Verifikasi Akhir Day 31

Setelah me-replace file-file di atas, pastikan untuk menjalankan test:

```bash
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./... -count=1 -v
```

Lalu cek tiga skenario ini (menggunakan CMD syntax dengan `^`):

```cmd
curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-501" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"

curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-502" ^
  -d "{\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":2}]}"

curl -i "http://localhost:8080/v1/transactions?status=flagged"
```