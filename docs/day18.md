# Day 18 — Phase 3 Kickoff: PostgreSQL Foundation with Docker Compose

## Tujuan
- Memulai **Phase 3 (Persistence)** dengan fondasi infra yang rapi untuk local development.
- Menyiapkan **PostgreSQL** via Docker Compose.
- Menerapkan pendekatan **database per service** untuk local dev, tetapi tetap memakai **1 Postgres instance** agar sederhana dan hemat resource.
- Menambahkan konfigurasi database pada Inventory dan Transaction untuk persiapan migrasi repository dari in-memory ke DB.
- Memperbarui `README.md` agar sesuai dengan state project setelah Phase 2 selesai.

## Keputusan arsitektur
- Local development memakai **1 container PostgreSQL**.
- Di dalam 1 instance tersebut dibuat **2 database logical**:
  - `inventory_db`
  - `transaction_db`
- Alasan memilih pendekatan ini:
  - lebih realistis untuk microservices dibanding 1 database campur semua tabel
  - tetap ringan untuk laptop dan proses belajar
  - menunjukkan pemahaman **data ownership per service**
  - mudah dijelaskan sebagai stepping stone menuju production-grade architecture

## Yang dibangun / diubah
- Menambahkan `docker-compose.yml` untuk menjalankan PostgreSQL.
- Menambahkan init script PostgreSQL untuk bootstrap multi-database:
  - `inventory_db`
  - `transaction_db`
- Menambahkan konfigurasi DB ke:
  - `services/inventory/internal/config/config.go`
  - `services/transaction/internal/config/config.go`
- Menambahkan default env DB di script run:
  - `scripts/run-inventory.ps1`
  - `scripts/run-transaction.ps1`
- Memperbarui `README.md` agar mencerminkan:
  - 3 services aktif
  - flow arsitektur saat ini
  - status akhir Phase 2
  - awal Phase 3 dengan PostgreSQL foundation

## Konsep yang dipelajari
- **Database per service** sebagai prinsip ownership data pada microservices.
- **Single Postgres instance, multi-database** sebagai kompromi yang baik untuk local development.
- Docker Compose sebagai fondasi environment yang repeatable.
- Menyiapkan config lebih awal sebelum repository benar-benar dipindahkan ke DB.
- Dokumentasi repo perlu di-update ketika 1 phase selesai agar portfolio terlihat rapi.

## File yang berubah / ditambah

### Root
- [ADD] `docker-compose.yml`
- [MOD] `README.md`

### Scripts
- [ADD] `scripts/postgres/init-multiple-dbs.sh`
- [MOD] `scripts/run-inventory.ps1`
- [MOD] `scripts/run-transaction.ps1`

### Inventory
- [MOD] `services/inventory/internal/config/config.go`
  - tambah field DB config dan helper connection string

### Transaction
- [MOD] `services/transaction/internal/config/config.go`
  - tambah field DB config dan helper connection string

## Cara verifikasi

### 1. Jalankan PostgreSQL
```bash
docker compose up -d postgres
```

### 2. Cek container hidup
```bash
docker ps
```

Pastikan ada container PostgreSQL untuk project ini.

### 3. Jalankan Inventory Service
```powershell
.\scripts\run-inventory.ps1
```

Pada tahap Day 18, inventory memang belum wajib memakai repository PostgreSQL. Fokusnya masih memastikan config DB dan environment local sudah siap.

### 4. Jalankan Transaction Service
```powershell
.\scripts\run-transaction.ps1
```

### 5. Pastikan config DB terbaca
Lihat log startup masing-masing service dan pastikan nilai default DB mengarah ke:
- host `127.0.0.1`
- port `55432`
- database `inventory_db` atau `transaction_db`

## Self-Review
- Kenapa untuk local dev kita boleh memakai **1 Postgres instance** tetapi tetap memisahkan **database logical per service**?
- Apa bedanya **menyiapkan DB config** dengan **benar-benar memindahkan repository ke DB**?
- Kenapa perubahan infra seperti Docker Compose juga harus ikut tercermin di `README.md`?