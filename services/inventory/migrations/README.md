# Transaction Migrations

Folder ini disiapkan untuk migration milik `Transaction Service`.

Pada tahap Day 19, migration yang sudah dibuat baru untuk `Inventory Service`, karena inventory akan menjadi service pertama yang dipindahkan dari in-memory ke PostgreSQL.

Migration transaction akan ditambahkan ketika mulai membuat tabel seperti:

- `transactions`
- `transaction_items`

Tujuannya agar ownership database per service tetap jelas sejak awal.