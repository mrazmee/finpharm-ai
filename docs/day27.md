# Day 27 — Milestone 3 Stabilization

## Tujuan

- Menutup Milestone 3 dengan repo yang lebih rapi dan lebih aman untuk dipresentasikan
- Menyinkronkan dokumentasi utama (`README.md`) dengan kondisi kode terbaru sampai Day 26
- Mengurangi kebocoran raw internal error pada handler service
- Menjadikan kondisi project lebih siap dibaca recruiter, mentor, atau reviewer portfolio

---

# Flow Hari Ini

Tidak ada flow baru pada Day 27.

Seluruh flow utama tetap mengacu pada Day 26:

1. Validasi request + `Idempotency-Key`
2. Cek replay by `idempotency_key`
3. Pre-check stock ke `Inventory Service`
4. Simpan transaction dengan status awal `PENDING`
5. Deduct stock ke `Inventory Service`
6. Jika sukses → update status `APPROVED`
7. Jika gagal → update status `FAILED`

Fokus hari ini adalah **stabilisasi**, bukan perubahan behavior.

---

# Perubahan Utama Hari Ini

## 1. Sinkronisasi README

`README.md` diperbarui agar mencerminkan kondisi sistem sampai Day 26, termasuk:

- architecture terbaru
- endpoint yang sudah tersedia
- idempotency behavior
- transaction lifecycle
- stock deduction
- known limitation
- arah pengembangan berikutnya

---

## 2. Hardening Error Handling

Sebelumnya beberapa handler masih mengembalikan:

```go
err.Error()
```

Untuk kasus `500 INTERNAL_ERROR`.

Hari ini diperbaiki menjadi:

- response lebih stabil
- tidak membocorkan detail internal
- tetap konsisten dengan error contract

Handler yang diperbaiki:

- `Inventory Service` — medicine handler
- `Inventory Service` — stock handler
- `Transaction Service` — stock handler
- `Transaction Service` — transaction handler

---

# Kenapa Raw Internal Error Tidak Boleh?

`err.Error()` mentah sering mengandung:

- detail implementasi internal
- nama driver / query database
- error dari network / transport layer
- informasi yang tidak relevan untuk client

Ini menyebabkan:

- API sulit dibaca reviewer
- potensi kebocoran informasi
- response tidak konsisten

---

# Apa yang Tetap Boleh Ditampilkan?

Beberapa error tetap boleh detail karena memang bagian dari domain:

- validation error
- not found
- insufficient stock
- domain-specific upstream error

Yang dihentikan hanya **internal error level 500**.

---

# Status Milestone 3

Day 27 menegaskan bahwa **Milestone 3 sudah selesai** dengan cakupan:

- inventory persistence
- transaction persistence
- list transactions
- idempotency
- stock deduction
- transaction lifecycle (`PENDING`, `APPROVED`, `FAILED`)

---

# Known Limitation Saat Ini

- Deduct stock multi-item masih dilakukan satu per satu
- Belum ada rollback otomatis jika sebagian gagal
- Belum ada saga / compensation mechanism
- Consistency antar item belum atomic

Ini akan menjadi fokus perbaikan di tahap berikutnya.

---

# File Yang Berubah / Ditambah

## Root

```
[MOD] README.md
```

## Inventory Service

```
[MOD] services/inventory/internal/httpapi/handler/medicine.go
[MOD] services/inventory/internal/httpapi/handler/stock.go
```

## Transaction Service

```
[MOD] services/transaction/internal/httpapi/handler/stock.go
[MOD] services/transaction/internal/httpapi/handler/transaction.go
```

## Documentation

```
[ADD] docs/day27.md
```

---

# Cara Verifikasi

## 1. Jalankan Semua Test

```bash
go test ./... -count=1 -v
```

Expected:

- semua test tetap pass
- tidak ada regression dari Day 23–26

---

## 2. Create Transaction

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-201" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:

- status `201 Created`
- transaction status `APPROVED`

---

## 3. Replay Request

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-201" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:

- status `200 OK`
- transaction yang sama dikembalikan

---

## 4. Insufficient Stock

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-202" \
  -d "{\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":999}]}"
```

Expected:

- status `409 Conflict`
- code `INSUFFICIENT_STOCK`

---

## 5. Jalankan Semua Test (Per Service)

```bash
go test ./services/inventory/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./... -count=1 -v
```

---

# Self-Review

- Kenapa Day 27 tidak menambah feature besar baru, tapi tetap penting?
- Kenapa README.md perlu disinkronkan ke kondisi kode terbaru?
- Kenapa raw internal error tidak ideal untuk response API?
- Kenapa Milestone 3 sekarang sudah cukup kuat untuk dijadikan portfolio checkpoint?
- Apa limitation paling penting yang masih tersisa sebelum masuk phase berikutnya?

---

# Checklist Milestone 3

- [ ] Inventory pakai PostgreSQL + sqlx
- [ ] Transaction pakai PostgreSQL + sqlx
- [ ] POST /v1/transactions lewat Gateway
- [ ] GET /v1/transactions lewat Gateway
- [ ] pagination/filtering dasar
- [ ] idempotency key
- [ ] stock deduction
- [ ] transaction lifecycle
- [ ] semua test hijau
- [ ] README sinkron
- [ ] docs harian sinkron

---

# Verifikasi Akhir

Jalankan:

```bash
go test ./... -count=1 -v
```

Lalu lakukan smoke test:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-201" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Pastikan:

- flow tetap berjalan
- tidak ada perubahan behavior dari Day 26
- tidak ada raw internal error yang bocor