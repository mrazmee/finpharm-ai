# Day 41 — Grafana Local Dashboard

## Tujuan
Menambahkan Grafana lokal untuk memvisualisasikan metrics Prometheus yang sudah dikumpulkan pada Day 40.

---

# Kenapa Day 41 Tetap Sesuai Roadmap?

Karena:
- Day 38 menambahkan `/metrics`
- Day 39 memperkuat observability dengan audit log dan trace mindset
- Day 40 mengumpulkan metrics dengan Prometheus
- Day 41 menambahkan visualisasi dashboard

Jadi Day 41 tetap satu jalur roadmap observability.

---

# Scope Day 41

Day 41 mencakup:
- menjalankan Grafana lokal
- provisioning datasource Prometheus
- provisioning dashboard awal
- memverifikasi dashboard otomatis ter-load
- memverifikasi panel membaca metric `finpharm_*`

**Scope Yang Belum Dikerjakan di Day 41:**
Supaya tetap lurus roadmap, Day 41 belum mencakup alerting Grafana, SSO, Grafana Cloud, dashboard production-grade yang kompleks, Loki / log panels, Jaeger / traces panel, atau OpenTelemetry.

---

# File Yang Ditambahkan

```text
[NEW]     observability/grafana/provisioning/datasources/prometheus.yml
[NEW]     observability/grafana/provisioning/dashboards/dashboard.yml
[NEW]     observability/grafana/dashboards/finpharm-overview.json
[NEW]     docker-compose.grafana.yml
[NEW]     scripts/run-grafana.ps1
[NEW]     scripts/stop-grafana.ps1
[REPLACE] docs/day41.md
```

---

# Dependensi & Arsitektur Day 41

**Dependensi:**
Day 41 membutuhkan Prometheus Day 40 sudah berjalan di `localhost:9090`, semua service tetap aktif, dan worker metrics tetap aktif di `localhost:9094`.

**Arsitektur Sederhana:**
- service metrics tetap berjalan di host
- Prometheus tetap scrape semua metrics
- Grafana berjalan di container sendiri
- Grafana terhubung ke Prometheus melalui: `http://host.docker.internal:9090`

---

# Catatan Penting Tentang Provisioning Dashboard

Ada satu hal penting yang tervalidasi saat implementasi:
File JSON dashboard **harus berada di folder dashboards yang memang di-mount sebagai dashboard JSON path**.

Jika file JSON diletakkan di folder provisioning, folder `Finpharm` mungkin muncul, tetapi isi dashboard tidak terbaca.

Struktur yang benar adalah:
- provider config → `observability/grafana/provisioning/dashboards/`
- dashboard JSON → `observability/grafana/dashboards/`

---

# Struktur Folder

```text
finpharm-ai/
├─ docs/
├─ observability/
│  ├─ prometheus/
│  │  └─ prometheus.yml
│  └─ grafana/
│     ├─ dashboards/
│     │  └─ finpharm-overview.json
│     └─ provisioning/
│        ├─ dashboards/
│        │  └─ dashboard.yml
│        └─ datasources/
│           └─ prometheus.yml
├─ scripts/
│  ├─ run-prometheus.ps1
│  ├─ stop-prometheus.ps1
│  ├─ run-grafana.ps1
│  └─ stop-grafana.ps1
├─ services/
├─ docker-compose.prometheus.yml
└─ docker-compose.grafana.yml
```

---

# Cara Menjalankan

## 1. Pastikan Service Lokal Aktif
```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

## 2. Pastikan Prometheus Aktif
```powershell
.\scripts\run-prometheus.ps1
```

## 3. Jalankan Grafana
```powershell
.\scripts\run-grafana.ps1
```

## 4. Buka Grafana
Buka di browser: `http://localhost:3000`
- **Username:** `admin`
- **Password:** `admin`

---

# Validasi Manual

## A. Cek Grafana Bisa Login
**Expected:** Halaman login muncul, login berhasil, dashboard otomatis tersedia.

## B. Cek Datasource
Masuk ke: **Connections / Data sources**
**Expected:** Datasource Prometheus sudah ada.

## C. Cek Dashboard
Masuk ke: **Dashboards** > folder **Finpharm** > buka dashboard **Finpharm Overview**
**Expected:** Dashboard terbuka tanpa perlu import manual.

---

# Panel Yang Ada di Dashboard

1. **Total HTTP Requests by Service**
   Query: `sum by (service) (finpharm_http_requests_total)`

2. **HTTP Request Rate by Service**
   Query: `sum by (service) (rate(finpharm_http_requests_total[5m]))`

3. **HTTP P95 Latency by Service**
   Query: `histogram_quantile(0.95, sum by (le, service) (rate(finpharm_http_request_duration_seconds_bucket[5m])))`

4. **Worker Events by Result**
   Query: `sum by (result) (finpharm_worker_events_total)`

5. **Worker Inflight Messages**
   Query: `finpharm_worker_inflight_messages`

6. **Worker Processing Count**
   Query: `finpharm_worker_processing_duration_seconds_count`

7. **HTTP Requests Breakdown**
   Query: `sum by (service, method, path, status) (finpharm_http_requests_total)`

---

# Traffic Contoh Untuk Mengisi Dashboard

*(Ganti `<STAFF_TOKEN>` dan `<SUPERVISOR_TOKEN>` dengan token yang didapat dari perintah Auth)*

```cmd
:: Ambil Token Staff
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"staff-001\",\"role\":\"staff\"}"

:: Ambil Token Supervisor
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"supervisor-001\",\"role\":\"supervisor\"}"

:: Hit Medicines
curl -i "http://localhost:8080/v1/medicines?limit=2&offset=0" ^
  -H "Authorization: Bearer <STAFF_TOKEN>"

:: Hit Stock Check
curl -i -X POST http://localhost:8080/v1/stock/check ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -d "{\"medicine_id\":\"PARA500\",\"qty\":1}"

:: Create Transaction
curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-day41-001" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"

:: List Transactions (Supervisor)
curl -i "http://localhost:8080/v1/transactions?limit=5&offset=0" ^
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>"
```

---

# Hasil Validasi Runtime & Interpretasi

Day 41 sudah tervalidasi dengan hasil:
- Folder dashboard Finpharm muncul dan dashboard tampil
- Semua panel (HTTP request total, request rate, worker events, worker inflight, worker processing count, dan tabel breakdown) berjalan normal.

## Interpretasi Hasil Dashboard

- Traffic HTTP antar service sudah terlihat jelas.
- Worker event sukses sudah tercatat.
- Inflight worker `0` adalah normal saat tidak ada message aktif.
- Worker processing count bertambah saat worker memproses event.
- Dashboard sudah cukup untuk menjelaskan arsitektur sistem saat demo.

## Hal Yang Terlihat Tapi Normal

**1. P95 Latency Bisa Terlihat Tinggi**
Ini bisa normal jika request transaction menunggu AI auditor, AI auditor timeout/fallback cukup lama, atau jumlah traffic masih kecil sehingga p95 terlihat tajam. Jadi panel latency tinggi tidak otomatis berarti bug.

**2. Tidak Semua Service Selalu Muncul Di Semua Panel**
Ini normal jika pada sesi itu service tertentu belum menerima traffic cukup atau belum menghasilkan series yang relevan untuk query panel tersebut.

---

# Troubleshooting

**1. Folder Muncul Tapi Dashboard Kosong**
- *Penyebab:* File JSON dashboard diletakkan di folder yang salah.
- *Solusi:* Pastikan file JSON ada di `observability/grafana/dashboards/`, BUKAN di `observability/grafana/provisioning/dashboards/`.

**2. Datasource Error**
- *Cek:* Prometheus masih aktif di `localhost:9090`, Grafana container bisa mengakses `host.docker.internal`, dan Docker Desktop aktif.

**3. Panel Worker Kosong**
- *Kemungkinan:* Worker belum memproses event baru, request transaction hanya replay karena idempotency key sama, atau worker baru restart dan belum ada event sukses.

**4. Panel HTTP Ada Tapi Service Tertentu Tidak Muncul**
- *Kemungkinan:* Service tersebut belum menerima traffic pada sesi itu.

---

# Checklist Hardening Akhir

Bagian ini dikumpulkan agar tidak terlupakan saat final hardening.

## Dashboard / Observability
- [ ] Rapikan dashboard per service
- [ ] Tambah panel error rate per service/status
- [ ] Tambah panel p99 latency
- [ ] Tambah panel worker DLQ / retry metrics jika metric itu nanti ditambahkan
- [ ] Tambah panel AI auditor fallback usage
- [ ] Tambah dashboard business metrics transaksi

## Improvement dari Hasil Validasi Runtime
- [ ] Ubah nama panel *Worker Processing Count* menjadi lebih jelas (misal: *Worker Processed Events Count* atau *Worker Processing Histogram Count*)
- [ ] Tambah panel HTTP 4xx/5xx by service
- [ ] Tambah panel khusus transaction service (`POST /v1/transactions`, `GET /v1/transactions`)
- [ ] Tambah panel khusus AI auditor (request count, latency, fallback usage)
- [ ] Tambah panel worker result yang lebih kaya (success, duplicate, retry, dlq, invalid_dlq)

## UX / Demo
- [ ] Ubah default password Grafana
- [ ] Dokumentasikan login Grafana di README
- [ ] Buat urutan demo observability: traffic → Prometheus query → Grafana dashboard
- [ ] Rapikan nama dashboard dan panel untuk reviewer

## Hardening Akhir
- [ ] Satukan compose Prometheus + Grafana bila ingin lebih rapi
- [ ] Pertimbangkan alerting dasar
- [ ] Pertimbangkan Grafana provisioning folder lebih modular
- [ ] Pertimbangkan screenshot dashboard di README
- [ ] Pertimbangkan korelasi metrics dengan trace dan audit log

---

# Self-Review

- Kenapa Grafana dikerjakan setelah Prometheus?
- Kenapa folder dashboard bisa muncul tapi isi kosong?
- Kenapa worker panel bisa kosong kalau tidak ada event baru?
- Kenapa replay idempotency tidak menambah worker metric?
- Kenapa p95 latency bisa terlihat tinggi saat traffic masih kecil?
- Dashboard mana yang paling kuat untuk demo reviewer?