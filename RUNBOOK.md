# Finpharm-AI Runbook

## Tujuan
Dokumen operasional singkat untuk menjalankan, mereset, menguji, dan mendemokan project secara lokal tanpa kebingungan.

---

## Lokasi File Penting
- `Makefile` → root project
- `RUNBOOK.md` → root project
- PowerShell scripts → `scripts/*.ps1`
- Shell scripts → `scripts/*.sh`

---

## Catatan Developer Experience (DX)
Project ini saat ini tetap menjadikan **PowerShell (`.ps1`) sebagai jalur utama**, karena environment utama pengembangan adalah Windows.

Namun untuk memperbaiki developer experience dan membuat repo lebih reviewer-friendly, script operasional penting juga disediakan dalam versi `.sh`.

Jadi:
- `.ps1` **tetap dipertahankan**
- `.sh` **ditambahkan sebagai parity script**
- `Makefile` **ditambahkan sebagai wrapper command**

**Catatan:**
- Pada Windows, jalur yang paling direkomendasikan tetap memakai `scripts/*.ps1`
- Script `.sh` terutama berguna untuk Linux / macOS / WSL / reviewer yang ingin melihat parity command

---

## Port Utama
- Gateway: `http://localhost:8080`
- Transaction Service: `http://localhost:8081`
- Inventory Service: `http://localhost:8082`
- AI Auditor Service: `http://localhost:8083`
- Worker Metrics: `http://localhost:9094/metrics`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`
- RabbitMQ UI: `http://localhost:15672`

---

## Default Local Credentials

### PostgreSQL
- user: `finpharm`
- password: `finpharm`

### RabbitMQ
- user: `finpharm`
- password: `finpharm`

### Grafana
- login default local: `admin` / `admin`

**Catatan Penting:**
- nilai di atas hanya untuk local/dev/demo
- jangan dipakai untuk environment non-local
- untuk gateway non-local, ganti `JWT_SECRET` dari default dev value

---

## Quick Start

### 1. Jalankan PostgreSQL
```powershell
docker compose up -d postgres
```

### 2. Jalankan RabbitMQ
```powershell
.\scripts\rabbitmq-up.ps1
```

### 3. Jalankan migration
```powershell
.\scripts\migrate-inventory-up.ps1
.\scripts\migrate-transaction-up.ps1
```

### 4. Jalankan service
Buka terminal terpisah:
```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

### 5. Jalankan observability
```powershell
.\scripts\run-prometheus.ps1
.\scripts\run-grafana.ps1
```

### 6. Jalankan demo readiness
```powershell
.\scripts\demo-readiness.ps1
```

---

## Menjalankan Service

### PowerShell
```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

### Shell (`.sh`)
```bash
./scripts/run-inventory.sh
./scripts/run-transaction.sh
./scripts/run-ai-auditor.sh
./scripts/run-gateway.sh
./scripts/run-worker.sh
```

---

## Menjalankan Observability

### Prometheus
**PowerShell:**
```powershell
.\scripts\run-prometheus.ps1
```
**Shell:**
```bash
./scripts/run-prometheus.sh
```

### Grafana
**PowerShell:**
```powershell
.\scripts\run-grafana.ps1
```
**Shell:**
```bash
./scripts/run-grafana.sh
```

---

## Menghentikan Observability

### Prometheus
**PowerShell:**
```powershell
.\scripts\stop-prometheus.ps1
```
**Shell:**
```bash
./scripts/stop-prometheus.sh
```

### Grafana
**PowerShell:**
```powershell
.\scripts\stop-grafana.ps1
```
**Shell:**
```bash
./scripts/stop-grafana.sh
```

---

## RabbitMQ Helper

**PowerShell:**
```powershell
.\scripts\rabbitmq-up.ps1
.\scripts\rabbitmq-logs.ps1
.\scripts\rabbitmq-down.ps1
```

**Shell:**
```bash
./scripts/rabbitmq-up.sh
./scripts/rabbitmq-logs.sh
./scripts/rabbitmq-down.sh
```

---

## Reset Data Transaksi

Gunakan saat ingin membersihkan data transaksi tanpa mengubah stock medicines.

### PowerShell
```powershell
.\scripts\reset-transaction-data.ps1
```

### Shell
```bash
./scripts/reset-transaction-data.sh
```

**Catatan Penting:**
- script ini hanya membersihkan data transaksi
- yang dibersihkan biasanya: `transactions` dan `transaction_items`
- script ini **tidak mereset stock inventory**
- stock tetap dikelola oleh Inventory Service

---

## Generate Traffic untuk Observability

**PowerShell:**
```powershell
.\scripts\generate-traffic.ps1
```

**Shell:**
```bash
./scripts/generate-traffic.sh
```

Gunakan script ini untuk:
- menaikkan HTTP traffic dashboard
- memunculkan transaction outcomes
- memunculkan AI auditor request count
- memunculkan worker activity

---

## Demo Readiness Helper

**PowerShell:**
```powershell
.\scripts\demo-readiness.ps1
```

**Shell:**
```bash
./scripts/demo-readiness.sh
```

Checklist ini membantu memastikan:
- semua service sudah hidup
- observability stack sudah hidup
- token demo sudah siap
- dashboard siap dipresentasikan

---

## Migration Helper

**Catatan:**
- script `.sh` migration memakai CLI `migrate`
- DSN database diambil dari environment variable
- ini sengaja aman supaya tidak mengasumsikan DSN lokal yang salah

### Inventory
**PowerShell:**
```powershell
.\scripts\migrate-inventory-up.ps1
.\scripts\migrate-inventory-down.ps1
```
**Shell:**
```bash
export INVENTORY_DB_DSN="postgres://user:pass@localhost:5432/inventory_db?sslmode=disable"
./scripts/migrate-inventory-up.sh
./scripts/migrate-inventory-down.sh
```

### Transaction
**PowerShell:**
```powershell
.\scripts\migrate-transaction-up.ps1
.\scripts\migrate-transaction-down.ps1
```
**Shell:**
```bash
export TRANSACTION_DB_DSN="postgres://user:pass@localhost:5432/transaction_db?sslmode=disable"
./scripts/migrate-transaction-up.sh
./scripts/migrate-transaction-down.sh
```

---

## Makefile

Jika environment punya `make`, kamu bisa pakai wrapper command berikut:

```makefile
make help
make run-rabbitmq
make run-prometheus
make run-grafana
make demo-readiness
make demo-traffic
make test
```

**Catatan:**
- Makefile di repo ini tetap memanggil PowerShell
- tujuan Makefile adalah merapikan command, bukan menggantikan `.ps1`

---

## Catatan Makefile di Windows

Makefile di repo ini ditambahkan untuk merapikan command dan memberi entry point yang lebih enak untuk reviewer.
Namun perlu dicatat:
- environment utama project tetap Windows + PowerShell
- target Makefile saat ini memanggil command PowerShell
- output `make help` bisa sedikit berbeda tergantung implementasi `make` yang terpasang di Windows

**Contoh:**
- beberapa implementasi `make` bisa menampilkan output `echo` dengan tanda kutip
- itu **bukan bug pada repo**, hanya perbedaan perilaku tool `make`

Jadi:
- Makefile tetap berguna sebagai wrapper command
- tetapi jalur utama yang paling stabil di Windows tetap `scripts/*.ps1`

---

## Demo Flow Final yang Direkomendasikan

### 1. Persiapan
- jalankan PostgreSQL
- jalankan RabbitMQ
- jalankan migrations
- jalankan semua service
- jalankan Prometheus dan Grafana
- jalankan `.\scripts\demo-readiness.ps1`

### 2. Flow Inti
- issue token staff
- issue token supervisor
- cek medicines
- cek stock
- create transaction approved
- create transaction pending_review / flagged
- replay request dengan idempotency key yang sama
- list transactions

### 3. Flow Observability
- buka Prometheus targets
- buka dashboard Grafana
- jalankan `.\scripts\generate-traffic.ps1`
- tunjukkan:
  - HTTP traffic
  - latency
  - transaction outcomes
  - audit decisions
  - worker metrics
  - fallback total
  - 4xx/5xx panel bila perlu

### 4. Flow Error Demo (Opsional)
Gunakan untuk memunculkan panel error:
- unauthorized request → 401
- invalid limit → 400
- missing idempotency key → 400
- rate limit → 429

---

## Example Demo Requests

**Issue Token Staff**
```cmd
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"staff-001\",\"role\":\"staff\"}"
```

**Issue Token Supervisor**
```cmd
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"supervisor-001\",\"role\":\"supervisor\"}"
```

**Cek Medicines**
```cmd
curl -i "http://localhost:8080/v1/medicines?limit=2&offset=0" ^
  -H "Authorization: Bearer <STAFF_TOKEN>"
```

**Cek Stock**
```cmd
curl -i -X POST http://localhost:8080/v1/stock/check ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -d "{\"medicine_id\":\"PARA500\",\"qty\":1}"
```

**Create Transaction**
```cmd
curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-demo-001" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

**Replay Request Sama**
```cmd
curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-demo-001" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

**List Transactions**
```cmd
curl -i "http://localhost:8080/v1/transactions?limit=5&offset=0" ^
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>"
```

---

## Observability Check

### Prometheus
- buka `http://localhost:9090/targets`
- pastikan semua target `UP`

### Grafana
- buka `http://localhost:3000`
- login default local: `admin` / `admin`
- buka dashboard: **Finpharm Overview**

---

## Catatan Demo Penting

- gunakan idempotency key yang berbeda untuk transaksi baru
- replay dengan idempotency key yang sama tidak membuat event baru
- worker metric hanya naik jika benar-benar ada event baru
- DLQ yang berisi message lama tidak otomatis berarti bug aktif saat ini
- dashboard Grafana JSON **harus berada di folder**: `observability/grafana/dashboards/`
- bukan di: `observability/grafana/provisioning/dashboards/`

---

## Troubleshooting Singkat

**Metrics HTTP Tidak Muncul**
- generate traffic dulu
- lalu refresh Prometheus / Grafana

**Panel 4xx/5xx Kosong**
- generate error runtime:
  - unauthorized 401
  - invalid request 400
  - rate limit 429
- tunggu scrape Prometheus
- refresh Grafana

**Worker Metrics Kosong**
- pastikan worker hidup
- kirim transaksi baru dengan idempotency key baru
- cek worker log

**Grafana Folder Muncul Tapi Dashboard Kosong**
- pastikan JSON dashboard ada di: `observability/grafana/dashboards/`
- bukan di folder provisioning dashboards

**Script `.sh` Gagal Dieksekusi**
```bash
chmod +x ./scripts/*.sh
```

**Migration `.sh` Gagal**
- pastikan CLI `migrate` terpasang
- pastikan env var DSN sudah di-set