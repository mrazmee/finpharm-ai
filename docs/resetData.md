# Reset Transaction Data

## Tujuan

Membersihkan data transaksi untuk kebutuhan demo/portfolio tanpa mengubah:
- schema database
- migration history
- konfigurasi service

---

# Scope Reset

## Yang Di-reset

Script ini hanya membersihkan tabel:
- `transactions`
- `transaction_items`

Dengan cara:
- `TRUNCATE`
- `RESTART IDENTITY`

## Yang TIDAK Di-reset

Script ini tidak menyentuh:
- tabel medicines / inventory
- stock obat
- migration metadata
- service lain

---

# Kapan Dipakai?

Pakai reset ini saat:
- list transaction sudah terlalu penuh data lama
- ingin demo ulang Day 30 / Day 31 / Day 32 dengan data bersih
- ingin screenshot portfolio yang lebih rapi

---

# File Yang Berubah / Ditambah

```text
[ADD] scripts/reset_transaction_data.go
[ADD] scripts/reset-transaction-data.ps1
```

---

# Cara Menjalankan

## Menggunakan PowerShell

```powershell
.\scripts\reset-transaction-data.ps1
```

## Menggunakan Go Langsung

```bash
go run .\scripts\reset_transaction_data.go --yes
```

## Output Yang Diharapkan

Script akan menampilkan:
- jumlah row `transactions` sebelum reset
- jumlah row `transaction_items` sebelum reset
- jumlah row setelah reset

Contoh Output:
```text
=== RESET TRANSACTION DATA SUCCESS ===
DB       : 127.0.0.1:55432 / transaction_db
Before   : transactions=15, transaction_items=22
After    : transactions=0, transaction_items=0
```

---

# Cara Pakai Sekarang (Urutan Yang Disarankan)

## 1. Stop Service Terkait

Stop dulu service `gateway` dan `transaction` biar tidak ada request yang masuk saat reset.

## 2. Jalankan Reset

```powershell
.\scripts\reset-transaction-data.ps1
```

## 3. Nyalakan Lagi Semua Service

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-transaction.ps1
.\scripts\run-gateway.ps1
```

## 4. Verifikasi List Transaction Kosong

```bash
curl -i "http://localhost:8080/v1/transactions"
```

Expected:
- status `200 OK`
- `items` kosong (`[]`)
- `total: 0`

---

# Catatan Penting

Reset ini **tidak mengembalikan stock obat ke seed awal**.

Jadi setelah reset:
- data transaksi bersih
- tapi stock tetap seperti kondisi terakhir

Kalau setelah reset kamu ingin demo stock dari angka awal/seed lagi, itu butuh reset inventory juga, karena script ini hanya membersihkan data transaksi.