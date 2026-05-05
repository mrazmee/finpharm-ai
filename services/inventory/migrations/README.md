## `services/inventory/migrations/README.md`

# Inventory Migrations

Folder ini berisi migration milik `Inventory Service`.

Migration pertama yang sudah tersedia saat ini adalah:

- `000001_create_medicines_and_stocks.up.sql`
- `000001_create_medicines_and_stocks.down.sql`

Tujuan migration inventory saat ini:

- membuat tabel `medicines`
- membuat tabel `stocks`
- menambahkan seed data awal untuk medicines dan stock

Kenapa migration dipisah per service:

- menjaga ownership database tetap jelas
- memudahkan evolusi schema per service
- lebih dekat dengan praktik microservices di industri

Migration untuk `Transaction Service` disimpan di folder terpisah: `services/transaction/migrations`.