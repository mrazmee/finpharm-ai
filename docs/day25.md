# Day 25 — Idempotency Key for Create Transaction

## Tujuan

- Menambahkan proteksi **anti double submit** pada `POST /v1/transactions`.
- Mewajibkan header `Idempotency-Key` dari client.
- Menyimpan `idempotency_key` ke PostgreSQL.
- Mengembalikan transaksi yang sama bila request identik dikirim ulang dengan key yang sama.
- Membedakan status response:
  - `201 Created` untuk create pertama
  - `200 OK` untuk replay request yang sama.

---

# Yang Dibangun / Diubah

### Database

- Menambahkan migration baru untuk kolom `idempotency_key` pada tabel `transactions`.
- Menambahkan **unique index** pada `transactions.idempotency_key`.

### Domain

- Menambahkan `IdempotencyKey` ke domain `Transaction`.

### Repository

- Menambahkan method:

```
GetByIdempotencyKey(...)
```

di repository contract.

### Usecase

- Menambahkan `CreateTransactionResult` agar usecase dan handler tahu apakah request ini:
  - create baru
  - replay dari request sebelumnya.

### HTTP Layer

- Menambahkan validasi header `Idempotency-Key` di:
  - **Gateway Service**
  - **Transaction Service**

- Mengubah perilaku endpoint:

```
POST /v1/transactions
```

menjadi:

| Kondisi | Response |
|------|------|
| request pertama | `201 Created` |
| request ulang dengan key sama | `200 OK` |

### Testing

Menambahkan test untuk:

- missing `Idempotency-Key`
- replay response
- propagation header dari gateway ke transaction service.

---

# Konsep yang Dipelajari

Pada domain transaksi, retry dari client atau network timeout bisa menyebabkan **submit ganda**.

Idempotency membantu memastikan **satu niat transaksi dari client tidak menghasilkan banyak transaksi berbeda**.

Beberapa prinsip yang diterapkan:

### Validasi Gateway

Gateway tetap melakukan validasi untuk:

```
fail fast
```

tetapi transaction service tetap harus memiliki validasi sendiri agar tetap aman jika dipanggil langsung.

### Database sebagai Lapisan Pertahanan Terakhir

Unique index di database penting untuk melindungi dari:

```
race condition
```

yang tidak selalu bisa ditangani oleh application code.

---

# Contract Baru Hari Ini

## Request Header Wajib

```http
Idempotency-Key: idem-001
```

---

# Perilaku Endpoint

## Request dengan Key Baru

```
POST /v1/transactions
```

Perilaku:

- menyimpan transaksi baru
- response:

```
201 Created
```

---

## Request dengan Key yang Sama

Jika client mengirim ulang request dengan **Idempotency-Key yang sama**:

- tidak membuat transaksi baru
- mengembalikan transaksi yang sudah ada

Response:

```
200 OK
```

---

# Kenapa Tetap Menggunakan Database Unique Index?

Karena pengecekan di application code saja **tidak cukup aman untuk concurrent request**.

Contoh skenario:

1. Request **A** dan **B** datang hampir bersamaan.
2. Keduanya mengecek bahwa key belum ada.
3. Tanpa unique index, dua transaksi bisa tetap tersimpan.

Karena itu sistem memakai **dua lapisan proteksi**:

1️⃣ Pre-check di **usecase**  
2️⃣ **Unique index di database**

---

# Kenapa Old Rows Di-Backfill Saat Migration?

Database sudah memiliki transaksi lama dari **Day 22–24** yang dibuat sebelum fitur idempotency ada.

Jika kolom baru langsung dibuat:

```
NOT NULL UNIQUE
```

migration bisa gagal.

Karena itu row lama diberi nilai:

```
legacy-<transaction_id>
```

Tujuannya:

- migration tetap sukses
- data lama tetap valid
- transaksi baru mulai memakai `Idempotency-Key` dari client

---

# File Yang Berubah / Ditambah

## Transaction Service

```
[ADD] services/transaction/migrations/000002_add_idempotency_key_to_transactions.up.sql
[ADD] services/transaction/migrations/000002_add_idempotency_key_to_transactions.down.sql

[MOD] services/transaction/internal/domain/transaction.go

[MOD] services/transaction/internal/usecase/transaction_usecase.go
[MOD] services/transaction/internal/usecase/transaction_usecase_test.go

[MOD] services/transaction/internal/repository/transaction_sqlx_repo.go

[MOD] services/transaction/internal/httpapi/handler/transaction.go
[MOD] services/transaction/internal/httpapi/handler/transaction_test.go
```

---

## Gateway Service

```
[MOD] services/gateway/internal/httpapi/handler/transaction_proxy.go
[MOD] services/gateway/internal/httpapi/handler/transaction_proxy_test.go
```

---

## Documentation

```
[ADD] docs/day25.md
```

---

# Cara Verifikasi

## 1. Jalankan Migration Transaction Terbaru

```powershell
.\scripts\migrate-transaction-up.ps1
```

---

## 2. Cek Schema Table `transactions`

```bash
docker exec -it finpharm-postgres \
psql -U finpharm -d transaction_db -c "\d transactions"
```

Expected:

- ada kolom `idempotency_key`
- ada unique index:

```
idx_transactions_idempotency_key
```

---

## 3. Create Transaction Pertama

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-001" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"
```

Expected:

```
201 Created
```

---

## 4. Replay Request Yang Sama

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-001" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"
```

Expected:

```
200 OK
```

dan **id transaksi sama dengan request pertama**.

---

## 5. Coba Tanpa Header

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2}]}"
```

Expected:

```
400 Bad Request
```

Error field:

```
header.Idempotency-Key
```

---

## 6. Jalankan Test

```bash
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./... -count=1 -v
```

---

# Self-Review

Beberapa pertanyaan untuk memastikan pemahaman:

- Kenapa **create transaction** perlu `Idempotency-Key`, tetapi **list transaction tidak**?
- Kenapa usecase tetap melakukan **pre-check**, walaupun database sudah punya unique index?
- Kenapa replay request dikembalikan sebagai **200**, bukan **201**?
- Kenapa old data perlu di-backfill saat migration idempotency?
- Kenapa gateway juga ikut validasi `Idempotency-Key`, padahal transaction service sudah validasi sendiri?

---

# Langkah Verifikasi Tambahan

Setelah semua file ditempel, jalankan:

```bash
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./... -count=1 -v
```

Kemudian tes manual:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-001" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"

curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-001" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"

curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2}]}"
```