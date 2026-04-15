# Finpharm-AI Runbook

## Tujuan

Dokumen singkat untuk menjalankan, mereset, dan mendemokan project secara lokal tanpa kebingungan.

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

## Menjalankan Service

Buka terminal terpisah untuk masing-masing service.

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
- Script ini hanya membersihkan data transaksi.
- Yang dibersihkan biasanya: `transactions` dan `transaction_items`.
- Script ini **tidak mereset stock inventory**.
- Data stock / medicines tetap dikelola oleh boundary `Inventory Service`.

---

## Generate Traffic Untuk Observability

**PowerShell:**
```powershell
.\scripts\generate-traffic.ps1
```

**Shell:**
```bash
./scripts/generate-traffic.sh
```

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

---

## Migration Helper

**Catatan:**
- Script `.sh` migration memakai CLI `migrate`.
- DSN database diambil dari environment variable.
- Ini sengaja aman supaya tidak mengasumsikan DSN lokal yang salah.

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
make run-prometheus
make run-grafana
make reset-transaction-data
make demo-check
```

**Catatan:**
- Di repo ini `Makefile` tetap memanggil PowerShell, karena environment utama project adalah Windows.
- Tujuan `Makefile` adalah merapikan command, bukan menggantikan `.ps1`.

---

## Catatan Makefile di Windows

`Makefile` di repo ini ditambahkan untuk merapikan command dan memberi entry point yang lebih enak untuk reviewer.

Namun perlu dicatat:
- Environment utama project tetap Windows + PowerShell.
- Target `Makefile` saat ini memanggil command PowerShell.
- Output `make help` bisa sedikit berbeda tergantung implementasi `make` yang terpasang di Windows.

**Contoh:**
- Beberapa implementasi `make` bisa menampilkan output `echo` dengan tanda kutip.
- Itu **bukan bug pada repo**, hanya perbedaan perilaku tool `make`.

**Jadi:**
- `Makefile` tetap berguna sebagai wrapper command.
- Tetapi jalur utama yang paling stabil di Windows tetap `scripts/*.ps1`.

---

## Demo Flow Yang Direkomendasikan

### 1. Issue Token Staff
```cmd
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"staff-001\",\"role\":\"staff\"}"
```

### 2. Issue Token Supervisor
```cmd
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"supervisor-001\",\"role\":\"supervisor\"}"
```

### 3. Cek Medicines
```cmd
curl -i "http://localhost:8080/v1/medicines?limit=2&offset=0" ^
  -H "Authorization: Bearer <STAFF_TOKEN>"
```

### 4. Cek Stock
```cmd
curl -i -X POST http://localhost:8080/v1/stock/check ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -d "{\"medicine_id\":\"PARA500\",\"qty\":1}"
```

### 5. Create Transaction
```cmd
curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-demo-001" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

### 6. List Transactions
```cmd
curl -i "http://localhost:8080/v1/transactions?limit=5&offset=0" ^
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>"
```

---

## Observability Check

### Prometheus
- Buka `http://localhost:9090/targets`
- Pastikan semua target `UP`

### Grafana
- Buka `http://localhost:3000`
- Login default: `admin` / `admin`
- Buka dashboard: **Finpharm Overview**

---

## Catatan Demo Penting

- Gunakan idempotency key yang berbeda untuk transaksi baru.
- Replay dengan idempotency key yang sama tidak membuat event baru.
- Worker metric hanya naik jika benar-benar ada event baru.
- DLQ yang berisi message lama tidak otomatis berarti bug aktif saat ini.
- Dashboard Grafana JSON **harus berada di folder**: `observability/grafana/dashboards/`
- Bukan di: `observability/grafana/provisioning/dashboards/`

---

## Troubleshooting Singkat

**Metrics HTTP Tidak Muncul**
- Generate traffic dulu
- Lalu refresh Prometheus / Grafana

**Worker Metrics Kosong**
- Pastikan worker hidup
- Kirim transaksi baru dengan idempotency key baru
- Cek worker log

**Grafana Folder Muncul Tapi Dashboard Kosong**
- Pastikan JSON dashboard ada di folder: `observability/grafana/dashboards/`
- Bukan di folder provisioning dashboards

**Script `.sh` Gagal Dieksekusi**
Jalankan perintah ini terlebih dahulu:
```bash
chmod +x ./scripts/*.sh
```

**Migration `.sh` Gagal**
- Pastikan CLI `migrate` terpasang
- Pastikan env var DSN sudah di-set