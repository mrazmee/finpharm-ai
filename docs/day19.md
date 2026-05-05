# Day 19 — Inventory Initial Migration (Medicines + Stocks)

## Tujuan
- Menambahkan struktur migration untuk persistence.
- Membuat schema awal `Inventory Service` di PostgreSQL.
- Menyimpan data awal medicines dan stock ke database.
- Menyiapkan fondasi agar pada langkah berikutnya inventory repository bisa dipindahkan dari in-memory ke DB.

## Yang dibangun / diubah
- Menambahkan script PowerShell untuk menjalankan migration inventory:
  - `scripts/migrate-inventory-up.ps1`
  - `scripts/migrate-inventory-down.ps1`
- Menambahkan migration SQL pertama untuk inventory:
  - create table `medicines`
  - create table `stocks`
  - seed initial data
- Menambahkan placeholder folder migration untuk transaction agar boundary per service tetap jelas.

## Konsep yang dipelajari
- Migration digunakan untuk melacak perubahan schema database secara terstruktur.
- Pada microservices, migration sebaiknya dipisahkan per service.
- `Inventory Service` dipindahkan lebih dulu ke DB karena menjadi source of truth untuk obat dan stok.
- Seed data awal membantu transisi dari in-memory ke database tanpa mengubah contract API.

## File yang berubah / ditambah

### Scripts
- [ADD] `scripts/migrate-inventory-up.ps1`
- [ADD] `scripts/migrate-inventory-down.ps1`

### Inventory
- [ADD] `services/inventory/migrations/000001_create_medicines_and_stocks.up.sql`
- [ADD] `services/inventory/migrations/000001_create_medicines_and_stocks.down.sql`
- [ADD] `services/inventory/migrations/README.md`

### Transaction
- [ADD] `services/transaction/migrations/README.md`

### Docs
- [ADD] `docs/day19.md`

## Cara verifikasi

### 1. Pastikan PostgreSQL hidup
```bash
docker compose up -d postgres
```

### 2. Jalankan migration inventory
```powershell
.\scripts\migrate-inventory-up.ps1
```

### 3. Pastikan migration sukses
Kalau sukses, command migration akan selesai tanpa error.

### 4. Cek koneksi DB dengan helper script
```bash
go run ./scripts/check_inventory_db.go
```

Expected output minimal:
- connection string tercetak
- `db ping ok`

### 5. Rollback jika perlu
```powershell
.\scripts\migrate-inventory-down.ps1
```

Setelah rollback, kamu bisa menjalankan `migrate-inventory-up.ps1` lagi untuk memastikan migration repeatable.

## Self-Review
- Kenapa tabel tidak dibuat manual lewat pgAdmin, tetapi lewat migration?
- Kenapa migration inventory dipisah dari migration transaction walaupun masih 1 Postgres instance?
- Kenapa seed data awal berguna saat transisi dari in-memory ke PostgreSQL?