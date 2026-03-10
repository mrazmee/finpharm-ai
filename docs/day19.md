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

### Transaction
- [ADD] `services/transaction/migrations/README.md`

### Docs
- [ADD] `docs/day19.md`

## Cara verifikasi

### 1. Pastikan PostgreSQL hidup
```bash
docker compose up -d postgres