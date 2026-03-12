# Transaction Service Migrations

Folder ini berisi migration SQL untuk database milik `Transaction Service`.

## Kenapa dipisah?
Walaupun saat ini local development masih memakai **1 PostgreSQL instance**, setiap service tetap harus punya ownership schema masing-masing.

Artinya:
- migration inventory ada di folder migration inventory
- migration transaction ada di folder migration transaction
- perubahan schema transaction tidak dicampur ke migration inventory

## Catatan desain
`transaction_items.medicine_id` sengaja **tidak** dibuat foreign key ke tabel `inventory.medicines`.

Alasannya:
- `Inventory Service` dan `Transaction Service` adalah boundary yang berbeda
- transaction hanya menyimpan referensi ID obat yang dipakai saat transaksi dibuat
- pada microservices, service tidak boleh bergantung pada foreign key lintas service meskipun saat local dev masih memakai 1 instance Postgres

## Cara menjalankan
Naikkan migration:

```powershell
.\scripts\migrate-transaction-up.ps1
```

Turunkan 1 migration terakhir:

```powershell
.\scripts\migrate-transaction-down.ps1
```