## `services/transaction/migrations/README.md`

```markdown
# Transaction Migrations

Folder ini disiapkan untuk migration milik `Transaction Service`.

Pada tahap sebelum Day 21, migration konkret untuk transaction memang belum dibuat karena fokus Phase 3 baru menyelesaikan persistence di `Inventory Service` terlebih dahulu.

Rencana awal migration transaction:

- tabel `transactions`
- tabel `transaction_items`
- kebutuhan pendukung seperti kolom atau tabel untuk idempotency pada tahap berikutnya

Tujuannya agar boundary persistence per service tetap jelas sejak awal dan tidak bercampur dengan migration inventory.
```