# Day 20 (Addendum) — Pre-Day21 Cleanup: Documentation Sync + Stock Check Contract Fix

## Tujuan
- Merapikan repo sebelum masuk ke Day 21 agar pondasi persistence transaction tidak dibangun di atas dokumentasi dan contract yang drift.
- Menyelaraskan `README.md` dengan state codebase yang sebenarnya.
- Melengkapi dokumentasi harian Day 18, Day 19, dan Day 20 yang sempat terpotong.
- Memperbaiki contract `Transaction -> Inventory` pada stock check agar `qty` asli dari client benar-benar diteruskan ke downstream.
- Menambahkan test agar bug contract serupa tidak lolos lagi di kemudian hari.

## Yang dirapikan / diubah
- Memperbarui `README.md`:
  - status project saat ini
  - host port PostgreSQL yang benar (`55432`)
  - langkah migration inventory sebelum menjalankan inventory service
- Menambahkan `.gitignore` dasar untuk project Go.
- Melengkapi dokumen `docs/day18.md`, `docs/day19.md`, dan `docs/day20.md` termasuk section `Self-Review`.
- Memperbaiki `services/inventory/migrations/README.md` agar benar-benar menjelaskan migration inventory.
- Menambahkan `services/transaction/migrations/README.md` sebagai placeholder migration transaction.
- Memperbaiki contract transaction stock repo agar menerima `requestedQty` dan meneruskannya ke `Inventory Service`.
- Menambahkan test `Transaction Service` untuk memastikan `qty` yang dikirim client benar-benar diteruskan ke inventory.

## Konsep yang dipelajari
- Dokumentasi yang drift bisa sama berbahayanya dengan kode yang drift, karena bisa membuat implementasi berikutnya dibangun di atas asumsi yang salah.
- Sebuah interface mungkin terlihat kecil, tetapi kalau contract-nya tidak jujur, bug konseptual akan terbawa ke phase berikutnya.
- Test yang baik tidak hanya mengecek status code, tetapi juga bisa mengunci perilaku penting seperti payload forwarding ke downstream service.
- Cleanup sebelum lanjut feature baru adalah kebiasaan yang sehat dalam kerja backend, terutama saat masuk phase persistence.

## File yang berubah / ditambah

### Root
- [MOD] `README.md`
- [ADD] `.gitignore`

### Docs
- [MOD] `docs/day18.md`
- [MOD] `docs/day19.md`
- [MOD] `docs/day20.md`
- [ADD] `docs/day20add.md`

### Transaction
- [MOD] `services/transaction/internal/domain/stock.go`
- [MOD] `services/transaction/internal/usecase/stock_usecase.go`
- [MOD] `services/transaction/internal/repository/stock_http_repo.go`
- [MOD] `services/transaction/internal/repository/stock_memory_repo.go`
- [MOD] `services/transaction/internal/httpapi/handler/stock_test.go`
- [ADD] `services/transaction/migrations/README.md`

### Inventory
- [MOD] `services/inventory/migrations/README.md`

## Cara verifikasi

### 1. Baca ulang README
Pastikan langkah menjalankan project sekarang sudah berurutan:
1. `docker compose up -d postgres`
2. `./scripts/migrate-inventory-up.ps1`
3. `./scripts/run-inventory.ps1`
4. `./scripts/run-transaction.ps1`
5. `./scripts/run-gateway.ps1`

### 2. Verifikasi forwarding qty
Kirim request berikut lewat gateway atau transaction:
```bash
curl -i -X POST http://localhost:8080/v1/stock/check \
  -H "Content-Type: application/json" \
  -d "{\"medicine_id\":\"PARA500\",\"qty\":7}"
```

Expected behavior:
- transaction meneruskan `qty=7` ke inventory
- hasil response tetap konsisten dengan contract lama

### 3. Jalankan test transaction
```bash
go test ./services/transaction/... -v
```

Fokus pada test baru yang memastikan `qty` diteruskan ke downstream inventory.

## Self-Review
- Kenapa `requestedQty` harus ikut masuk ke adapter HTTP transaction, walaupun transaction masih menghitung `is_available` sendiri?
- Kenapa `README.md` harus diperlakukan sebagai bagian dari product, bukan sekadar catatan tambahan?
- Sebelum masuk Day 21, kenapa lebih sehat merapikan drift dulu daripada langsung menambah tabel transaction?


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