# Day 20 — Inventory Service Switch to PostgreSQL with sqlx

## Tujuan
- Memindahkan `Inventory Service` dari in-memory repository ke PostgreSQL saat dijalankan normal.
- Memakai `sqlx` agar query SQL tetap eksplisit dan cocok untuk learning backend fundamentals.
- Menjaga API contract tetap sama meskipun source data sudah pindah ke database.
- Menjaga test handler tetap cepat dan stabil tanpa ketergantungan ke database.

## Yang dibangun / diubah
- Menambahkan `STORAGE_DRIVER` pada config inventory.
- Menyetel `scripts/run-inventory.ps1` agar default memakai `postgres`.
- Menambahkan helper koneksi PostgreSQL menggunakan `sqlx`.
- Menambahkan repository PostgreSQL untuk:
  - medicines list/detail
  - stock quantity lookup
- Mengubah wiring `Inventory Service` di `main.go` agar memilih repository berdasarkan `STORAGE_DRIVER`.
- Mengubah `httpapi.NewRouter(...)` agar menerima handler yang sudah di-wire dari luar.
- Memperbarui test inventory agar tetap menggunakan in-memory repository.

## Konsep yang dipelajari
- Repository implementation bisa diganti tanpa mengubah contract handler/usecase.
- `sqlx` memberi keseimbangan yang bagus: tetap nyaman dipakai, tapi query SQL tetap terlihat jelas.
- Wiring dependency di `main.go` membantu pemisahan tanggung jawab yang lebih bersih.
- Test tidak harus selalu memakai DB; unit test tetap bisa memakai in-memory dependency selama yang diuji adalah layer HTTP handler.

## File yang berubah / ditambah

### Inventory
- [MOD] `services/inventory/internal/config/config.go`
- [MOD] `services/inventory/cmd/api/main.go`
- [MOD] `services/inventory/internal/httpapi/router.go`
- [ADD] `services/inventory/internal/repository/postgres.go`
- [ADD] `services/inventory/internal/repository/medicine_sqlx_repo.go`
- [ADD] `services/inventory/internal/repository/stock_sqlx_repo.go`
- [ADD] `services/inventory/internal/repository/sql_errors.go`
- [MOD] `services/inventory/internal/httpapi/handler/medicine_test.go`

### Scripts
- [MOD] `scripts/run-inventory.ps1`

## Cara verifikasi

### 1. Pastikan dependency terpasang
```bash
go get github.com/jmoiron/sqlx
go get github.com/lib/pq
go mod tidy
```

### 2. Jalankan PostgreSQL
```bash
docker compose up -d postgres
```

### 3. Jalankan migration inventory
```powershell
.\scripts\migrate-inventory-up.ps1
```

### 4. Jalankan Inventory Service
```powershell
.\scripts\run-inventory.ps1
```

Karena `STORAGE_DRIVER` default sekarang `postgres`, service akan mencoba connect ke database inventory.

### 5. Uji endpoint inventory langsung
```bash
curl -i "http://localhost:8082/v1/medicines?limit=2&offset=0"
```

### 6. Uji stock check
```bash
curl -i -X POST http://localhost:8082/v1/stock/check \
  -H "Content-Type: application/json" \
  -d "{\"medicine_id\":\"PARA500\",\"qty\":10}"
```

### 7. Jalankan test inventory
```bash
go test ./services/inventory/... -v
```

## Self-Review
- Kenapa `sqlx` cocok untuk tahap belajar backend yang ingin tetap dekat dengan SQL?
- Kenapa wiring dependency dipindahkan ke `main.go`, bukan dibuat di dalam handler?
- Kenapa unit test handler tetap lebih baik memakai dependency in-memory meskipun service normal sudah pindah ke PostgreSQL?