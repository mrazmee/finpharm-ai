# Day 26 — Transaction Status Lifecycle + Stock Deduction

## Tujuan

- Menambahkan lifecycle status transaction:
  - `PENDING`
  - `APPROVED`
  - `FAILED`
- Menambahkan **deduct stock nyata** ke `Inventory Service`
- Mengubah create transaction agar tidak berhenti di simpan ke DB saja
- Membuat **Transaction Service** menjadi orchestrator lintas service

---

# Flow Baru Hari Ini

Alur `POST /v1/transactions` sekarang menjadi:

1. Validasi request + `Idempotency-Key`
2. Cek replay by `idempotency_key`
3. Pre-check stock ke `Inventory Service`
4. Simpan transaction dengan status awal `PENDING`
5. Deduct stock ke `Inventory Service`
6. Jika semua sukses → update status jadi `APPROVED`
7. Jika deduct gagal setelah transaction dibuat → update status jadi `FAILED`

---

# Lifecycle Status

## Kenapa create sukses sekarang statusnya `APPROVED`?

Setelah Day 26, create transaction **tidak berhenti saat data berhasil masuk database**.  

Transaction dianggap sukses penuh jika:

- Transaction tersimpan
- Stock benar-benar berhasil dikurangi

Sehingga response sukses normal sekarang adalah **status `APPROVED`**.

---

## Kenapa masih simpan dulu sebagai `PENDING`?

Status `PENDING` dipakai sebagai **status transisi** agar ada jejak lifecycle:

- “sudah tercatat”
- “belum selesai orchestration ke inventory”

Kalau langsung insert sebagai `APPROVED`, state transisi hilang.

---

## Apa yang terjadi kalau deduct stock gagal?

- Transaction akan di-update ke `FAILED`
- API create mengembalikan error
- Jika client replay dengan `Idempotency-Key` sama, transaction yang sama akan ditemukan

Contoh kasus:

- Pre-check lolos
- Transaction berhasil dibuat
- Saat deduct stock, inventory gagal / stock berubah / race condition

---

# Perubahan Penting pada Inventory Service

Inventory Service punya endpoint baru:

```http
POST /v1/stock/deduct
```

**Body:**

```json
{
  "medicine_id": "PARA500",
  "qty": 2
}
```

**Success response:**

```json
{
  "data": {
    "medicine_id": "PARA500",
    "deducted_qty": 2,
    "remaining_qty": 78
  },
  "request_id": "..."
}
```

---

# Perubahan Penting pada Transaction Service

Transaction repository sekarang punya method baru:

```go
UpdateStatus(...)
```

Stock repository di Transaction Service juga punya method baru:

```go
DeductStock(...)
```

---

# Known Limitation Hari Ini

- Deduct stock untuk multi-item **masih dilakukan satu per satu**
- Artinya:

  1. Item pertama berhasil deduct
  2. Item kedua gagal
  3. Transaction menjadi `FAILED`
  4. Item pertama belum otomatis di-roll back

- Ini diterima dulu untuk tahap belajar
- Bisa ditingkatkan nanti dengan:

  - Batch atomic deduct di Inventory
  - Compensation
  - Saga / event-driven orchestration

---

# File Yang Berubah / Ditambah

## Inventory Service

```
[MOD] services/inventory/internal/domain/errors.go
[MOD] services/inventory/internal/domain/stock.go
[MOD] services/inventory/internal/usecase/stock_usecase.go
[MOD] services/inventory/internal/repository/stock_memory_repo.go
[MOD] services/inventory/internal/repository/stock_sqlx_repo.go
[MOD] services/inventory/internal/httpapi/handler/stock.go
[MOD] services/inventory/internal/httpapi/router.go
[ADD] services/inventory/internal/httpapi/handler/stock_test.go
```

## Transaction Service

```
[MOD] services/transaction/internal/domain/stock.go
[MOD] services/transaction/internal/domain/transaction.go
[MOD] services/transaction/internal/repository/transaction_sqlx_repo.go
[MOD] services/transaction/internal/repository/stock_memory_repo.go
[MOD] services/transaction/internal/repository/stock_http_repo.go
[MOD] services/transaction/internal/usecase/transaction_usecase.go
[MOD] services/transaction/internal/usecase/transaction_usecase_test.go
[MOD] services/transaction/internal/httpapi/handler/transaction.go
[MOD] services/transaction/internal/httpapi/handler/transaction_test.go
```

## Documentation

```
[ADD] docs/day26.md
```

---

# Cara Verifikasi

## 1. Jalankan Service

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-gateway.ps1
```

---

## 2. Create Transaction Baru

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-101" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"
```

Expected:

- status `201 Created`
- transaction status = `APPROVED`

---

## 3. Replay Request Yang Sama

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-101" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"
```

Expected:

- status `200 OK`
- transaction id sama
- status tetap `APPROVED`

---

## 4. Coba Deduct Berlebih

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-102" \
  -d "{\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":999}]}"
```

Expected:

- status `409 Conflict`
- code `INSUFFICIENT_STOCK`

---

## 5. Lihat List Transaction

```bash
curl -i "http://localhost:8080/v1/transactions?status=approved"
```

Expected:

- transaksi baru sukses tampil sebagai `APPROVED`

---

## 6. Jalankan Test

```bash
go test ./services/inventory/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./... -count=1 -v
```

---

# Self-Review

- Kenapa create transaction sekarang berakhir sebagai `APPROVED`, bukan `PENDING`?  
- Kenapa transaction tetap dibuat dulu sebagai `PENDING` sebelum deduct stock?  
- Kenapa status `FAILED` tetap berguna walaupun API create akhirnya mengembalikan error?  
- Kenapa retry pada check stock masih bisa diterima, tapi retry blind pada deduct stock berbahaya?  
- Kenapa Day 26 ini belum benar-benar menyelesaikan consistency untuk multi-item transaction?

---

# Verifikasi yang Perlu Kamu Jalankan

Setelah paste semua file di atas:

```bash
go test ./services/inventory/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./... -count=1 -v
```

Lalu coba manual:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-101" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"

curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-101" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"

curl -i "http://localhost:8080/v1/transactions?status=approved"
```