# Day 21 — Transaction Initial Migration (Transactions + Transaction Items)

## Tujuan
- Memulai persistence untuk `Transaction Service` dari lapisan paling bawah: **schema database**.
- Menambahkan migration pertama untuk tabel `transactions` dan `transaction_items`.
- Menjaga boundary microservices tetap sehat dengan migration yang terpisah dari inventory.
- Menyiapkan pondasi supaya modul berikutnya bisa fokus ke repository `sqlx` dan create transaction use case.

## Yang dibangun / diubah
- Menambahkan migration SQL pertama untuk `Transaction Service`:
  - create table `transactions`
  - create table `transaction_items`
  - index dasar untuk query yang paling umum
- Menambahkan script PowerShell untuk menjalankan migration transaction:
  - `scripts/migrate-transaction-up.ps1`
  - `scripts/migrate-transaction-down.ps1`
- Memperbarui `services/transaction/migrations/README.md` agar menjelaskan alasan pemisahan migration per service dan alasan tidak memakai foreign key lintas service.
- Memperbarui `README.md` supaya status repo sesuai dengan progress terbaru.

## Konsep yang dipelajari
- Persistence yang sehat biasanya dimulai dari **schema design**, bukan langsung dari handler atau endpoint.
- Pada microservices, walaupun masih memakai 1 Postgres instance untuk local dev, migration harus tetap dipisah per service.
- `Transaction Service` tidak boleh membuat foreign key langsung ke tabel milik `Inventory Service`, karena itu akan merusak service boundary.
- `transaction_items` dibuat terpisah dari `transactions` agar satu transaksi bisa punya banyak item dan lebih mudah dikembangkan saat create transaction use case nanti.

## Desain schema hari ini

### Tabel `transactions`
Menyimpan header transaksi:
- `id`
- `status`
- `created_at`
- `updated_at`

### Tabel `transaction_items`
Menyimpan item-item dalam transaksi:
- `id`
- `transaction_id`
- `medicine_id`
- `qty`
- `created_at`

## Kenapa belum ada foreign key ke Inventory?
Karena `medicine_id` adalah referensi antar service, bukan relasi tabel dalam satu service boundary.

Walaupun database local kita masih satu instance PostgreSQL, secara desain kita tetap harus berpikir seperti ini:
- `Inventory Service` mengelola obat dan stok
- `Transaction Service` mengelola catatan transaksi
- integrasi antar service dilakukan lewat **HTTP / event**, bukan lewat foreign key lintas schema ownership

## File yang berubah / ditambah

### Root
- [MOD] `README.md`

### Scripts
- [ADD] `scripts/migrate-transaction-up.ps1`
- [ADD] `scripts/migrate-transaction-down.ps1`

### Transaction
- [ADD] `services/transaction/migrations/000001_create_transactions_and_transaction_items.up.sql`
- [ADD] `services/transaction/migrations/000001_create_transactions_and_transaction_items.down.sql`
- [MOD] `services/transaction/migrations/README.md`

### Docs
- [ADD] `docs/day21.md`

## Cara verifikasi

### 1. Pastikan PostgreSQL hidup
```bash
docker compose up -d postgres
```

### 2. Jalankan migration transaction
```powershell
.\scripts\migrate-transaction-up.ps1
```

Kalau berhasil, akan muncul output migration `1/u ...`.

### 3. Cek daftar tabel di `transaction_db`
```bash
docker exec -it finpharm-postgres psql -U finpharm -d transaction_db -c "\dt"
```

Expected minimal ada:
- `transactions`
- `transaction_items`
- `schema_migrations`

### 4. Lihat struktur tabel `transactions`
```bash
docker exec -it finpharm-postgres psql -U finpharm -d transaction_db -c "\d transactions"
```

### 5. Lihat struktur tabel `transaction_items`
```bash
docker exec -it finpharm-postgres psql -U finpharm -d transaction_db -c "\d transaction_items"
```

### 6. Rollback jika perlu
```powershell
.\scripts\migrate-transaction-down.ps1
```

### 7. Naikkan lagi untuk memastikan migration repeatable
```powershell
.\scripts\migrate-transaction-up.ps1
```

## Self-Review
- Kenapa kita mulai persistence transaction dari migration schema dulu, bukan langsung membuat endpoint `create transaction`?
- Kenapa `transaction_items` dipisah dari `transactions`, bukan semua kolom disimpan dalam satu tabel besar?
- Kenapa `medicine_id` di transaction tidak dibuat foreign key langsung ke tabel inventory, walaupun database local masih 1 instance?
- Kenapa migration per service penting untuk menjelaskan ownership data saat project ini dipresentasikan sebagai portfolio?