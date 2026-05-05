# Day 46 Prep — Pre-Alerting Cleanup

## Tujuan
Merapikan fondasi observability dan startup safety sebelum masuk ke implementasi alerting Day 46.

## Kenapa tahap ini perlu?
Alerting yang bagus bergantung pada dua hal:
1. metric yang stabil dan tidak noisy
2. startup config yang fail-fast agar service tidak jalan dengan konfigurasi rusak

Kalau dua hal ini belum rapi, alert bisa misleading dan repo akan terasa masih seperti project belajar biasa, bukan portfolio yang production-minded.

## Scope tahap ini
- normalisasi label path metric HTTP agar tidak memakai raw URL path
- penyesuaian histogram bucket supaya latency lebih masuk akal untuk local API service
- penambahan validation config di inventory, transaction, ai-auditor, dan worker
- penyelarasan startup main supaya semua service fail-fast saat config invalid
- penambahan config tests dasar

## File yang diubah
### Observability HTTP
- [REPLACE] `services/gateway/internal/observability/metrics.go`
- [REPLACE] `services/transaction/internal/observability/metrics.go`
- [REPLACE] `services/inventory/internal/observability/metrics.go`
- [REPLACE] `services/ai-auditor/internal/observability/metrics.go`

### Router HTTP
- [REPLACE] `services/gateway/internal/httpapi/router.go`
- [REPLACE] `services/transaction/internal/httpapi/router.go`
- [REPLACE] `services/inventory/internal/httpapi/router.go`
- [REPLACE] `services/ai-auditor/internal/httpapi/router.go`

### Main entrypoints
- [REPLACE] `services/gateway/cmd/api/main.go`
- [REPLACE] `services/transaction/cmd/api/main.go`
- [REPLACE] `services/inventory/cmd/api/main.go`
- [REPLACE] `services/ai-auditor/cmd/api/main.go`
- [REPLACE] `services/worker/cmd/worker/main.go`

### Config hardening
- [REPLACE] `services/gateway/internal/config/config.go`
- [REPLACE] `services/inventory/internal/config/config.go`
- [REPLACE] `services/transaction/internal/config/config.go`
- [REPLACE] `services/ai-auditor/internal/config/config.go`
- [REPLACE] `services/worker/internal/config/config.go`

### Tests
- [REPLACE] `services/gateway/internal/config/config_test.go`
- [NEW] `services/inventory/internal/config/config_test.go`
- [NEW] `services/transaction/internal/config/config_test.go`
- [NEW] `services/ai-auditor/internal/config/config_test.go`
- [NEW] `services/worker/internal/config/config_test.go`

### Worker observability
- [REPLACE] `services/worker/internal/observability/metrics.go`

## Perubahan utama

### 1. Path metric tidak lagi raw URL
Sebelumnya metric HTTP membaca `r.URL.Path` dari wrapper HTTP luar.

Akibatnya, endpoint dinamis seperti:
- `/v1/medicines/1`
- `/v1/medicines/2`
- `/v1/medicines/999`

akan menghasilkan label path yang berbeda-beda.

Sekarang metric dipindahkan ke middleware Gin dan menggunakan `c.FullPath()`.
Artinya path akan menjadi template route seperti:
- `/v1/medicines/:id`
- `/v1/transactions`
- `/v1/audit/transaction`

Ini lebih aman untuk cardinality, tapi tetap backward-compatible karena nama label tetap `path`.

### 2. Histogram bucket dibuat lebih realistis
Bucket default Prometheus tidak salah, tapi terlalu generic.

Sekarang latency HTTP memakai bucket:
- 10ms
- 25ms
- 50ms
- 100ms
- 250ms
- 500ms
- 1s
- 2s
- 5s

Untuk worker processing duration dipakai bucket sampai 10 detik agar lebih cocok untuk async processing.

### 3. Semua service sekarang fail-fast saat config invalid
Sebelumnya validasi config paling matang baru ada di gateway.
Sekarang inventory, transaction, ai-auditor, dan worker juga punya `Validate()`.

Contoh validasi yang sekarang ditambahkan:
- port wajib integer positif
- timeout wajib > 0
- base URL antar service wajib absolute URL
- AMQP URL wajib valid `amqp://` atau `amqps://`
- storage driver inventory wajib `memory` atau `postgres`
- queue/exchange worker wajib terisi
- Gemini API key diwajibkan di env non-local jika provider = `gemini`

## Dampak ke dashboard/query lama
Tidak perlu ubah query dashboard yang sudah ada, karena:
- nama metric tetap sama
- nama label tetap sama
- hanya isi label `path` yang sekarang lebih stabil

## Cara verifikasi

### 1. Jalankan test config
```powershell
go test ./services/gateway/internal/config -count=1 -v
go test ./services/inventory/internal/config -count=1 -v
go test ./services/transaction/internal/config -count=1 -v
go test ./services/ai-auditor/internal/config -count=1 -v
go test ./services/worker/internal/config -count=1 -v
```

### 2. Jalankan seluruh test project
```powershell
go test ./... -count=1 -v
```

### 3. Verifikasi metric path di Prometheus
Setelah service hidup, cek query berikut:
```promql
sum by (service, path) (finpharm_http_requests_total)
```

Expected:
- route dinamis tampil sebagai template seperti `/v1/medicines/:id`
- bukan pecah per ID aktual

### 4. Verifikasi startup fail-fast
Coba salahkan salah satu env penting, misalnya:
- `INVENTORY_BASE_URL=localhost:8082`
- `RABBITMQ_URL=http://localhost:5672`
- `STORAGE_DRIVER=sqlite`

Expected:
- service gagal start
- log menampilkan `config_invalid`

## Hasil yang diharapkan
Setelah tahap ini:
- observability lebih stabil
- latency panel lebih bermakna
- startup service lebih aman
- repo terlihat lebih production-minded
- Day 46 alerting bisa dibangun di atas metric yang lebih sehat

## Yang sengaja belum dikerjakan di tahap ini
- Alertmanager
- routing notifikasi alert
- alert receiver lokal
- rules alert baru
- `.env.example` global
- CI lint tambahan

Semua itu akan lebih tepat dibahas di Day 46 atau setelahnya.

## Self-review
- Kenapa raw URL path berbahaya untuk metric label?
- Kenapa route template lebih aman daripada path mentah?
- Kenapa fail-fast config penting untuk service production-like?
- Kenapa bucket histogram sebaiknya tidak asal default?