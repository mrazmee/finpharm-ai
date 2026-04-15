# Day 42 — Final Hardening I: DX + Demo Foundation

## Tujuan

Meratakan developer experience (DX) dan demo readiness agar project lebih terasa production-like saat dijalankan lokal dan saat ditunjukkan ke reviewer.

---

## Fokus Day 42

- Makefile
- Shell script `.sh`
- Runbook dasar
- Demo readiness helper
- Command yang lebih konsisten lintas environment

## Kenapa Day 42 Penting?

Project yang bagus tidak cukup hanya “fiturnya jalan”. Reviewer juga akan menilai:
- seberapa mudah project dijalankan
- seberapa rapi command-nya
- apakah ada runbook
- apakah ada demo flow yang jelas

---

## Scope Day 42

Day 42 mencakup:
- menambahkan `Makefile`
- menambahkan script `.sh` untuk jalur operasional penting
- menambahkan runbook
- menambahkan helper demo readiness
- merapikan cara run observability stack
- menambahkan parity script untuk operasional utama agar repo tidak terasa Windows-only

---

## File Yang Ditambahkan / Diubah

### Root
```text
[NEW] Makefile
[NEW] RUNBOOK.md
```

### Scripts `.sh`
```text
[NEW] scripts/run-gateway.sh
[NEW] scripts/run-transaction.sh
[NEW] scripts/run-inventory.sh
[NEW] scripts/run-ai-auditor.sh
[NEW] scripts/run-worker.sh
[NEW] scripts/run-prometheus.sh
[NEW] scripts/stop-prometheus.sh
[NEW] scripts/run-grafana.sh
[NEW] scripts/stop-grafana.sh
[NEW] scripts/demo-readiness.sh
[NEW] scripts/generate-traffic.sh
[NEW] scripts/migrate-inventory-up.sh
[NEW] scripts/migrate-inventory-down.sh
[NEW] scripts/migrate-transaction-up.sh
[NEW] scripts/migrate-transaction-down.sh
[NEW] scripts/rabbitmq-up.sh
[NEW] scripts/rabbitmq-down.sh
[NEW] scripts/rabbitmq-logs.sh
[NEW] scripts/reset-transaction-data.sh
```

### Scripts PowerShell
```text
[NEW] scripts/demo-readiness.ps1
```

### Docs
```text
[REPLACE] docs/day42.md
```

---

## Keputusan DX Yang Dipakai

Project ini saat ini tetap menjadikan **PowerShell (`.ps1`) sebagai jalur utama**, karena environment utama pengembangan adalah Windows.

Namun untuk memperbaiki developer experience dan membuat repo lebih reviewer-friendly, script operasional penting juga disediakan dalam versi `.sh`.

Jadi:
- `.ps1` **tetap dipertahankan**
- `.sh` **ditambahkan sebagai parity script**
- `Makefile` **ditambahkan sebagai wrapper command**

---

## Letak File Penting

- `Makefile` → root project
- `RUNBOOK.md` → root project
- PowerShell scripts → `scripts/*.ps1`
- Shell scripts → `scripts/*.sh`

---

## Struktur File

```text
finpharm-ai/
├─ Makefile
├─ RUNBOOK.md
├─ docs/
│  └─ day42.md
├─ observability/
├─ scripts/
│  ├─ run-gateway.ps1
│  ├─ run-gateway.sh
│  ├─ run-transaction.ps1
│  ├─ run-transaction.sh
│  ├─ run-inventory.ps1
│  ├─ run-inventory.sh
│  ├─ run-ai-auditor.ps1
│  ├─ run-ai-auditor.sh
│  ├─ run-worker.ps1
│  ├─ run-worker.sh
│  ├─ run-prometheus.ps1
│  ├─ run-prometheus.sh
│  ├─ stop-prometheus.ps1
│  ├─ stop-prometheus.sh
│  ├─ run-grafana.ps1
│  ├─ run-grafana.sh
│  ├─ stop-grafana.ps1
│  ├─ stop-grafana.sh
│  ├─ demo-readiness.ps1
│  ├─ demo-readiness.sh
│  ├─ generate-traffic.ps1
│  ├─ generate-traffic.sh
│  ├─ migrate-inventory-up.ps1
│  ├─ migrate-inventory-up.sh
│  ├─ migrate-inventory-down.ps1
│  ├─ migrate-inventory-down.sh
│  ├─ migrate-transaction-up.ps1
│  ├─ migrate-transaction-up.sh
│  ├─ migrate-transaction-down.ps1
│  ├─ migrate-transaction-down.sh
│  ├─ rabbitmq-up.ps1
│  ├─ rabbitmq-up.sh
│  ├─ rabbitmq-down.ps1
│  ├─ rabbitmq-down.sh
│  ├─ rabbitmq-logs.ps1
│  ├─ rabbitmq-logs.sh
│  ├─ reset-transaction-data.ps1
│  ├─ reset-transaction-data.sh
│  └─ reset_transaction_data.go
└─ services/
```

---

## Makefile

Makefile diletakkan di root project dan bernama persis: `Makefile` (Tanpa extension tambahan).

Tujuan Makefile di Day 42:
- merapikan command
- memberi satu entry point yang enak untuk reviewer
- BUKAN menggantikan `.ps1`

Karena environment utama project adalah Windows, target Makefile saat ini tetap memanggil PowerShell.

Contoh:
```makefile
make help
make run-prometheus
make run-grafana
make reset-transaction-data
make demo-check
```

---

## Script `.sh`

Script `.sh` ditambahkan untuk jalur operasional penting, terutama agar repo:
- lebih reviewer-friendly
- lebih enak dijalankan di Linux / macOS / WSL
- terasa lebih matang secara DX

**Catatan:**
- `.sh` tidak menggantikan `.ps1`
- `.ps1` tetap menjadi jalur utama di environment Windows

Jika memakai Linux / macOS / WSL, jalankan:
```bash
chmod +x ./scripts/*.sh
```

---

## Runbook

Runbook utama diletakkan di: `RUNBOOK.md` di root project, BUKAN di folder `docs/`.

Kenapa di root:
- lebih mudah ditemukan reviewer
- sejajar dengan `README.md` dan `Makefile`
- cocok sebagai dokumen operasional utama project

---

## Hasil Yang Diharapkan

Setelah Day 42:
- command project lebih rapi
- reviewer lebih mudah mencoba project
- ada jalur demo yang jelas
- ada runbook yang membantu saat lupa urutan menjalankan service
- repo tidak terasa terlalu bergantung pada PowerShell saja

---

## Cara Menggunakan

### Makefile
```makefile
make help
make run-prometheus
make run-grafana
make reset-transaction-data
make demo-check
```

### PowerShell Scripts
```powershell
.\scripts\run-prometheus.ps1
.\scripts\run-grafana.ps1
.\scripts\demo-readiness.ps1
```

### Shell Scripts
```bash
./scripts/run-inventory.sh
./scripts/run-transaction.sh
./scripts/run-ai-auditor.sh
./scripts/run-gateway.sh
./scripts/run-worker.sh
./scripts/run-prometheus.sh
./scripts/run-grafana.sh
./scripts/demo-readiness.sh
```

---

## Operasional Yang Sekarang Punya Parity Script

### Menjalankan Service
**PowerShell:**
```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```
**Shell:**
```bash
./scripts/run-inventory.sh
./scripts/run-transaction.sh
./scripts/run-ai-auditor.sh
./scripts/run-gateway.sh
./scripts/run-worker.sh
```

### Menjalankan Observability
**PowerShell:**
```powershell
.\scripts\run-prometheus.ps1
.\scripts\run-grafana.ps1
```
**Shell:**
```bash
./scripts/run-prometheus.sh
./scripts/run-grafana.sh
```

### Menghentikan Observability
**PowerShell:**
```powershell
.\scripts\stop-prometheus.ps1
.\scripts\stop-grafana.ps1
```
**Shell:**
```bash
./scripts/stop-prometheus.sh
./scripts/stop-grafana.sh
```

### RabbitMQ Helper
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

### Reset Data Transaksi
**PowerShell:**
```powershell
.\scripts\reset-transaction-data.ps1
```
**Shell:**
```bash
./scripts/reset-transaction-data.sh
```

### Generate Traffic
**PowerShell:**
```powershell
.\scripts\generate-traffic.ps1
```
**Shell:**
```bash
./scripts/generate-traffic.sh
```

### Demo Readiness
**PowerShell:**
```powershell
.\scripts\demo-readiness.ps1
```
**Shell:**
```bash
./scripts/demo-readiness.sh
```

### Migration Helper

**Inventory:**
```powershell
:: PowerShell
.\scripts\migrate-inventory-up.ps1
.\scripts\migrate-inventory-down.ps1
```
```bash
# Shell
export INVENTORY_DB_DSN="postgres://user:pass@localhost:5432/inventory_db?sslmode=disable"
./scripts/migrate-inventory-up.sh
./scripts/migrate-inventory-down.sh
```

**Transaction:**
```powershell
:: PowerShell
.\scripts\migrate-transaction-up.ps1
.\scripts\migrate-transaction-down.ps1
```
```bash
# Shell
export TRANSACTION_DB_DSN="postgres://user:pass@localhost:5432/transaction_db?sslmode=disable"
./scripts/migrate-transaction-up.sh
./scripts/migrate-transaction-down.sh
```

---

## Validasi Manual Day 42

- [x] Makefile tersedia di root
- [x] `RUNBOOK.md` tersedia di root
- [x] script `.sh` operasional utama tersedia di folder `scripts/`
- [x] script `.ps1` lama tetap ada
- [x] `demo-readiness.ps1` berjalan
- [x] `demo-readiness.sh` bisa dijalankan di shell-compatible environment
- [x] `RUNBOOK.md` bisa diikuti tanpa kebingungan besar

## Hasil Validasi Runtime Day 42

Day 42 tervalidasi dengan hasil:
- `demo-readiness.ps1` berjalan normal
- Makefile terbaca oleh tool `make` di environment Windows
- `RUNBOOK.md` tersedia di root
- file `.sh` operasional utama tersedia di folder `scripts/`

**Catatan hasil validasi:**
- output `make help` dapat terlihat sedikit berbeda tergantung implementation `make` di Windows. Hal ini tidak dianggap bug pada repo.
- reset transaction tetap hanya membersihkan data transaksi, bukan stock inventory.

---

## Catatan Implementasi Penting

### 1. File `.ps1` Tidak Dihapus
PowerShell tetap dipertahankan karena environment utama development adalah Windows, dan semua flow yang sudah tervalidasi selama pengerjaan memang memakai `.ps1`.

### 2. Script `.sh` Bukan Pengganti `.ps1`
Script `.sh` ditambahkan sebagai parity script, peningkatan DX, dan nilai plus untuk reviewer.
Jadi:
- `.ps1` tetap jalur utama di Windows
- `.sh` adalah jalur tambahan untuk Linux / macOS / WSL

### 3. Migration `.sh` Memakai Env Var
Script migration `.sh` tidak meng-hardcode DSN. Itu sengaja agar lebih aman, lebih fleksibel, dan tidak mengasumsikan konfigurasi lokal yang salah.

### 4. Reset Transaction Tidak Mereset Stock
Script reset transaction hanya membersihkan domain transaksi, biasanya:
- `transactions`
- `transaction_items`

Script ini tidak mereset stock inventory. Boundary inventory tetap dipisahkan agar sesuai desain microservices.

### 5. Perilaku `make` Di Windows Bisa Berbeda
Makefile di Day 42 tetap valid, tetapi output `make` pada Windows bisa sedikit berbeda tergantung implementation yang terpasang (contoh: hasil echo dengan tanda kutip). Itu bukan bug repo. Makefile tetap berfungsi sebagai wrapper command.

---

## Troubleshooting Singkat

**`make` tidak tersedia di Windows**
Tidak masalah besar untuk Day 42. Yang penting file `Makefile` ada, script `.ps1` tetap bisa dipakai, dan repo punya entry point command yang lebih rapi.

**Script `.sh` gagal dijalankan**
Jalankan:
```bash
chmod +x ./scripts/*.sh
```

**Migration `.sh` gagal**
Cek:
- CLI `migrate` terpasang
- env var `DSN` sudah di-set
- path migrations sesuai

**Shell script tidak dipakai reviewer**
Tidak masalah. Tetap ada nilainya karena menunjukkan repo lebih matang dan tidak terlalu platform-locked.

---

## Checklist Hardening Lanjutan

Bagian ini tetap ditahan untuk day hardening berikutnya:
- [ ] rate limit dasar
- [ ] runtime safety
- [ ] observability polish lanjutan
- [ ] dashboard/alerting polish
- [ ] release readiness final
- [ ] README final polish

---

## Self-Review

- Kenapa DX penting untuk portfolio?
- Kenapa `.ps1` tetap dipertahankan?
- Kenapa `.sh` tetap layak ditambahkan?
- Kenapa Makefile diletakkan di root?
- Kenapa `RUNBOOK.md` diletakkan di root?
- Kenapa migration `.sh` memakai env var dan bukan hardcoded DSN?