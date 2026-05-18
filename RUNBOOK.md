# Finpharm-AI Runbook

## Tujuan
Dokumen operasional singkat untuk menjalankan, mereset, menguji, dan mendemokan project secara lokal tanpa kebingungan.

---

## Lokasi File Penting
- `Makefile` → root project
- `RUNBOOK.md` → root project
- `README.md` → ringkasan project final
- PowerShell scripts → `scripts/*.ps1`
- Shell scripts → `scripts/*.sh`
- Day-by-day docs → `docs/dayXX.md`

---

## Catatan Developer Experience (DX)
Project ini tetap menjadikan **PowerShell (`.ps1`) sebagai jalur utama**, karena environment utama pengembangan adalah Windows.

Untuk reviewer / parity command:
- script `.sh` juga tersedia
- `Makefile` juga tersedia sebagai wrapper

Jadi:
- `.ps1` = jalur utama Windows
- `.sh` = parity untuk Linux / macOS / WSL
- `Makefile` = wrapper command

---

## Port Utama
- Gateway: `http://localhost:8080`
- Transaction Service: `http://localhost:8081`
- Inventory Service: `http://localhost:8082`
- AI Auditor Service: `http://localhost:8083`
- Knowledge Service: `http://localhost:8084`
- Worker Metrics: `http://localhost:9094/metrics`
- Prometheus: `http://localhost:9090`
- Alertmanager: `http://localhost:9093`
- Grafana: `http://localhost:3000`
- RabbitMQ UI: `http://localhost:15672`
- Alert Webhook: `http://localhost:18080`

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

### 3. Jalankan migrations domain utama
```powershell
.\scripts\migrate-inventory-up.ps1
.\scripts\migrate-transaction-up.ps1
```

### 4. Jalankan knowledge migration
```powershell
.\scripts\run-knowledge-migrate.ps1
```

### 5. Jalankan knowledge ingestion
```powershell
.\scripts\run-knowledge-ingest.ps1
```

### 6. Jalankan services
Buka terminal terpisah:
```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-knowledge-api.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

### 7. Jalankan observability
```powershell
.\scripts\run-prometheus.ps1
.\scripts\run-grafana.ps1
```

### 8. Jalankan alerting local
```powershell
.\scripts\run-alertmanager.ps1
.\scripts\run-alert-webhook.ps1
```

### 9. Jalankan demo readiness
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
.\scripts\run-knowledge-api.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

### Shell (`.sh`)
```bash
./scripts/run-inventory.sh
./scripts/run-transaction.sh
./scripts/run-ai-auditor.sh
./scripts/run-knowledge-api.sh
./scripts/run-gateway.sh
./scripts/run-worker.sh
```

---

## Menjalankan Knowledge Flow

**Migration**
```powershell
.\scripts\run-knowledge-migrate.ps1
```

**Ingestion**
```powershell
.\scripts\run-knowledge-ingest.ps1
```

**Manual retrieval**
```powershell
.\scripts\run-knowledge-query.ps1 -Query "apakah amoxicillin bisa dijual tanpa resep?"
```

**Manual answer**
```powershell
.\scripts\run-knowledge-answer.ps1 -Query "apa edukasi minimal untuk paracetamol otc?"
```

**HTTP knowledge API**
```powershell
.\scripts\run-knowledge-api.ps1
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

### Alertmanager
**PowerShell:**
```powershell
.\scripts\run-alertmanager.ps1
```
**Shell:**
```bash
./scripts/run-alertmanager.sh
```

### Alert Webhook
**PowerShell:**
```powershell
.\scripts\run-alert-webhook.ps1
```
**Shell:**
```bash
./scripts/run-alert-webhook.sh
```

---

## Menghentikan Observability

**Prometheus**
```powershell
.\scripts\stop-prometheus.ps1
```

**Grafana**
```powershell
.\scripts\stop-grafana.ps1
```

**Alertmanager**
```powershell
.\scripts\stop-alertmanager.ps1
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

**PowerShell:**
```powershell
.\scripts\reset-transaction-data.ps1
```

**Shell:**
```bash
./scripts/reset-transaction-data.sh
```

**Catatan Penting:**
- script ini hanya membersihkan data transaksi
- yang dibersihkan: `transactions` dan `transaction_items`
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
- knowledge flow sudah siap
- token demo sudah siap
- dashboard siap dipresentasikan

---

## Migration Helper

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

### Knowledge
**PowerShell:**
```powershell
.\scripts\run-knowledge-migrate.ps1
```
**Shell:**
```bash
./scripts/run-knowledge-migrate.sh
```

---

## Makefile

Jika environment punya `make`, kamu bisa pakai wrapper command berikut:

```makefile
make help
make run-rabbitmq
make run-prometheus
make run-grafana
make run-alertmanager
make run-alert-webhook
make demo-readiness
make demo-traffic
make test
```

**Catatan:**
- Makefile di repo ini tetap memanggil PowerShell
- tujuan Makefile adalah merapikan command, bukan menggantikan `.ps1`

---

## Demo Flow Final yang Direkomendasikan

### 1. Persiapan
- jalankan PostgreSQL
- jalankan RabbitMQ
- jalankan migrations:
  - inventory
  - transaction
  - knowledge
- jalankan knowledge ingestion
- jalankan semua service
- jalankan Prometheus, Grafana, Alertmanager, dan alert webhook
- jalankan `.\scripts\demo-readiness.ps1`

### 2. Flow Inti Backend
- issue token staff
- issue token supervisor
- cek medicines
- cek stock
- create transaction approved
- create transaction pending_review / flagged
- replay request dengan idempotency key yang sama
- list transactions

### 3. Flow Chatbot SOP
- jalankan query chatbot SOP via gateway
- tunjukkan positive grounded answer
- tunjukkan fallback out-of-domain answer
- tunjukkan bahwa response berisi:
  - `answer`
  - `fallback`
  - `citations`
  - `sources`
  - `confidence`

### 4. Flow Observability
- buka Prometheus targets
- buka Grafana dashboard
- jalankan `.\scripts\generate-traffic.ps1`
- tunjukkan:
  - HTTP traffic
  - latency
  - transaction outcomes
  - audit decisions
  - worker metrics
  - AI auditor fallback total
  - panel 4xx/5xx bila perlu

### 5. Flow Alerting (Opsional)
- buka Alertmanager
- jalankan local alert webhook
- tunjukkan alert yang sudah diverifikasi di Day 46 bila ingin demo reliability lebih dalam

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

**SOP Chatbot via Gateway**
```cmd
curl -i -X POST http://localhost:8080/v1/chat/sop ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -d "{\"question\":\"apakah amoxicillin bisa dijual tanpa resep?\",\"top_k\":5,\"min_score\":0.45}"
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

### Alertmanager
- buka `http://localhost:9093`
- pastikan route alerting hidup

---

## Troubleshooting Singkat

**Metrics HTTP tidak muncul**
- generate traffic dulu
- refresh Prometheus / Grafana

**Panel 4xx/5xx kosong**
- generate error runtime:
  - unauthorized `401`
  - invalid request `400`
  - rate limit `429`
- tunggu scrape Prometheus
- refresh Grafana

**Worker metrics kosong**
- pastikan worker hidup
- kirim transaksi baru dengan idempotency key baru
- cek worker log

**Knowledge query / answer gagal**
- pastikan `GEMINI_API_KEY` sudah tersedia
- pastikan knowledge migration dan ingestion sudah dijalankan
- pastikan knowledge API hidup di `:8084`

**Chatbot gateway gagal**
- pastikan knowledge API hidup
- pastikan `KNOWLEDGE_BASE_URL` mengarah ke `http://localhost:8084`
- pastikan token gateway valid

**Script `.sh` gagal dieksekusi**
```bash
chmod +x ./scripts/*.sh
```

**Migration `.sh` gagal**
- pastikan CLI `migrate` terpasang
- pastikan env var DSN sudah di-set

---

## Final Note

Runbook ini cukup untuk:
- menjalankan project lokal
- menyiapkan demo
- menunjukkan flow backend utama
- menunjukkan chatbot SOP via gateway
- menunjukkan observability dan alerting