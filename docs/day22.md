# Day 22 — Transaction Service Save Transaction to PostgreSQL with sqlx

## Tujuan
- Menambahkan endpoint `POST /v1/transactions` di `Transaction Service`.
- Melakukan validasi request transaksi yang berisi banyak item obat.
- Mengecek ketersediaan stok ke `Inventory Service` sebelum transaksi disimpan.
- Menyimpan `transaction header` dan `transaction_items` ke PostgreSQL memakai `sqlx`.
- Mengenalkan **DB transaction** (`BEGIN / COMMIT / ROLLBACK`) sebagai pondasi transaksi finansial yang lebih benar.

## Yang dibangun / diubah
- Menambahkan domain entity untuk transaksi dan item transaksi.
- Menambahkan `TransactionRepository` contract pada layer domain.
- Menambahkan repository PostgreSQL berbasis `sqlx` untuk menyimpan:
  - tabel `transactions`
  - tabel `transaction_items`
- Menambahkan `TransactionUsecase` untuk flow:
  - validasi payload
  - cek stok ke Inventory Service
  - simpan transaksi ke DB bila stok cukup
- Menambahkan handler HTTP `POST /v1/transactions`.
- Mengubah wiring `Transaction Service` di `main.go` agar dependency dibuat di composition root, bukan di router.
- Mengubah `httpapi.NewRouter(...)` agar menerima handler yang sudah di-wire dari luar.
- Menambahkan test untuk:
  - usecase create transaction
  - handler create transaction
  - test stock handler tetap jalan setelah refactor router

## Konsep yang dipelajari
- Service boundary tetap dijaga: `Transaction Service` tidak membaca tabel inventory secara langsung, tetapi tetap minta data stok lewat HTTP.
- Usecase menjadi tempat orkestrasi business flow: validasi -> check stock -> persist.
- Repository `sqlx` menyimpan data ke DB dengan query yang tetap eksplisit.
- Menyimpan header transaksi dan items transaksi harus dilakukan dalam **satu DB transaction** agar tidak terjadi partial write.
- Refactor wiring ke `main.go` membuat router tetap fokus pada routing, bukan pembuatan dependency.

## Catatan desain penting hari ini
- Hari ini kita baru melakukan **check stock lalu save transaction**.
- Kita **belum** melakukan stock reservation atau stock deduction.
- Artinya flow ini masih cocok untuk tahap belajar persistence, tetapi belum final untuk distributed consistency tingkat lanjut.
- Nanti saat masuk modul lebih lanjut, kita akan mulai berpikir ke arah:
  - idempotency
  - event-driven update
  - reserve / deduct stock
  - saga / compensation mindset

## Penyesuaian roadmap setelah molor Phase 2
Supaya pengerjaan tetap terukur sampai akhir, kita pakai target baru berikut:

### Target core portfolio (tanpa RAG optional)
- **Finish di Day 41**

### Target full project (termasuk RAG optional)
- **Finish di Day 45**

### Snapshot sisa roadmap sesudah Day 22
- **Day 23**: Gateway proxy `POST /v1/transactions` + end-to-end create transaction
- **Day 24**: Pagination & filtering list transactions
- **Day 25**: Idempotency key untuk create transaction
- **Day 26**: Milestone 3 stabilization + README/API collection sync
- **Day 27 - 31**: AI Auditor Service + Gemini integration + status review flow
- **Day 32 - 35**: RAG + vector storage + citation *(optional advanced)*
- **Day 36 - 39**: RabbitMQ + worker + reliability pattern
- **Day 40 - 43**: JWT, metrics, audit logging, tracing mindset
- **Day 44 - 45**: final hardening, runbook, portfolio demo readiness

Jadi mulai sekarang kita pakai mindset:
- **core selesai Day 41**
- **full selesai Day 45**

## File yang berubah / ditambah

### Root
- [MOD] `README.md`

### Transaction
- [MOD] `services/transaction/cmd/api/main.go`
- [ADD] `services/transaction/internal/domain/transaction.go`
- [MOD] `services/transaction/internal/domain/errors.go`
- [MOD] `services/transaction/internal/httpapi/router.go`
- [ADD] `services/transaction/internal/httpapi/handler/transaction.go`
- [MOD] `services/transaction/internal/httpapi/handler/debug_test.go`
- [MOD] `services/transaction/internal/httpapi/handler/stock_test.go`
- [ADD] `services/transaction/internal/httpapi/handler/transaction_test.go`
- [ADD] `services/transaction/internal/repository/postgres.go`
- [ADD] `services/transaction/internal/repository/transaction_sqlx_repo.go`
- [ADD] `services/transaction/internal/usecase/transaction_usecase.go`
- [ADD] `services/transaction/internal/usecase/transaction_usecase_test.go`

### Docs
- [ADD] `docs/day22.md`

## Cara verifikasi

### 1. Pastikan PostgreSQL hidup
```bash
docker compose up -d postgres
```

### 2. Pastikan migration inventory dan transaction sudah naik
```powershell
.\scripts\migrate-inventory-up.ps1
.\scripts\migrate-transaction-up.ps1
```

### 3. Jalankan Inventory Service
```powershell
.\scripts\run-inventory.ps1
```

### 4. Jalankan Transaction Service
```powershell
.\scripts\run-transaction.ps1
```

Saat start, `Transaction Service` sekarang seharusnya log bahwa persistence `postgres` dipakai.

### 5. Kirim create transaction langsung ke Transaction Service
```bash
curl -i -X POST http://localhost:8081/v1/transactions \
  -H "Content-Type: application/json" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":10},{\"medicine_id\":\"AMOX500\",\"qty\":2}]}"
```

Expected:
- status `201 Created`
- ada `id` transaksi
- status transaksi `PENDING`
- items transaksi muncul di response

### 6. Cek data masuk ke database
```bash
docker exec -it finpharm-postgres psql -U finpharm -d transaction_db -c "SELECT id, status, created_at FROM transactions ORDER BY created_at DESC;"
```

```bash
docker exec -it finpharm-postgres psql -U finpharm -d transaction_db -c "SELECT transaction_id, medicine_id, qty FROM transaction_items ORDER BY id DESC;"
```

### 7. Coba skenario stok tidak cukup
```bash
curl -i -X POST http://localhost:8081/v1/transactions \
  -H "Content-Type: application/json" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":999}]}"
```

Expected:
- status `409 Conflict`
- code error `INSUFFICIENT_STOCK`

### 8. Jalankan test
```bash
go test ./... -count=1 -v
```

## Self-Review
- Kenapa `CreateTransaction` tidak langsung menulis ke DB dari handler, tetapi lewat usecase?
- Kenapa penyimpanan `transactions` dan `transaction_items` harus dibungkus dalam satu DB transaction?
- Kenapa `Transaction Service` tetap check stock lewat HTTP ke `Inventory Service`, padahal database local masih satu instance?
- Kenapa hari ini kita belum langsung mengurangi stok inventory?
- Kenapa setelah refactor, dependency di-wire di `main.go` dan router hanya menerima handler yang sudah siap?
