# Day 40 — Prometheus Local Aggregation

## Tujuan

Mengumpulkan metrics dari semua service ke satu tempat menggunakan Prometheus lokal.

---

# Kenapa Day 40 Ada Setelah Day 38 dan Day 39?

Karena:
- Day 38 sudah menambahkan endpoint `/metrics`
- Day 39 sudah memperkuat observability dengan audit log dan trace mindset
- Day 40 membuat metrics itu bisa di-scrape dan dipantau secara terpusat

Jadi Day 40 tetap satu jalur roadmap observability.

---

# Scope Day 40

Day 40 hanya mencakup:
- menambahkan config Prometheus
- menjalankan Prometheus lokal dengan Docker Compose
- scrape semua service metrics
- memverifikasi semua target `UP`
- memverifikasi metric `finpharm_*` bisa di-query dari Prometheus UI

**Scope Yang Belum Dikerjakan di Day 40:**
Supaya tidak keluar roadmap, Day 40 belum mencakup Grafana, alerting rules, Loki, Jaeger, OpenTelemetry, atau dashboard production-grade.

---

# File Yang Ditambahkan

```text
[NEW]     observability/prometheus/prometheus.yml
[NEW]     docker-compose.prometheus.yml
[NEW]     scripts/run-prometheus.ps1
[NEW]     scripts/stop-prometheus.ps1
[REPLACE] docs/day40.md
```

---

# Arsitektur Sederhana

Service berjalan di host lokal:
- gateway → `localhost:8080`
- transaction → `localhost:8081`
- inventory → `localhost:8082`
- ai-auditor → `localhost:8083`
- worker metrics → `localhost:9094`

Prometheus berjalan di Docker:
- Prometheus UI → `localhost:9090`

Karena Prometheus berjalan di container dan service lain berjalan di host, target scrape menggunakan: `host.docker.internal`

## Catatan Penting Tentang `host.docker.internal`

`host.docker.internal` dipakai oleh **container Prometheus** untuk mengakses service yang berjalan di host.

Jadi:
- di file `prometheus.yml` → `host.docker.internal:8080` dan seterusnya adalah **benar**
- di browser host Windows → gunakan `localhost:8080`, `localhost:8081`, dst

Kalau `host.docker.internal` tidak bisa dibuka dari browser Windows, itu **normal** dan bukan bug Prometheus setup.

---

# Struktur Folder

```text
finpharm-ai/
├─ docs/
├─ observability/
│  └─ prometheus/
│     └─ prometheus.yml
├─ scripts/
│  ├─ run-prometheus.ps1
│  └─ stop-prometheus.ps1
├─ services/
└─ docker-compose.prometheus.yml
```

---

# Cara Menjalankan

## 1. Pastikan Semua Service Aktif

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

## 2. Jalankan Prometheus

```powershell
.\scripts\run-prometheus.ps1
```

Atau manual:
```bash
docker compose -f .\docker-compose.prometheus.yml up -d
docker compose -f .\docker-compose.prometheus.yml ps
```

## 3. Buka Prometheus UI

Buka di browser: `http://localhost:9090`

---

# Validasi Manual

## A. Cek Menu Targets

Masuk ke: `http://localhost:9090/targets`

**Expected:**
- job gateway → `UP`
- job transaction → `UP`
- job inventory → `UP`
- job ai-auditor → `UP`
- job worker → `UP`

---

# Query Prometheus Untuk Validasi Day 40

Jalankan query berikut di Prometheus UI:

## 1. HTTP Request Counter
`finpharm_http_requests_total`
**Expected:** Muncul series dari service gateway, transaction, inventory, dan ai-auditor (jika ada traffic audit pada sesi itu).

## 2. HTTP Duration Histogram Count
`finpharm_http_request_duration_seconds_count`
**Expected:** Muncul count histogram request dari service HTTP.

## 3. Aggregated HTTP Requests Per Service
`sum by (service) (finpharm_http_requests_total)`
**Expected:** Muncul service yang memang sudah menerima traffic pada sesi pengujian.

## 4. Worker Event Counter
`finpharm_worker_events_total`
**Expected:** Muncul minimal series seperti `{result="success"}`

## 5. Aggregated Worker Events By Result
`sum by (result) (finpharm_worker_events_total)`
**Expected:** Muncul minimal `{result="success"}`

## 6. Worker Inflight
`finpharm_worker_inflight_messages`
**Expected:** Biasanya `0` saat tidak ada pesan yang sedang aktif diproses.

## 7. Worker Processing Duration Count
`finpharm_worker_processing_duration_seconds_count`
**Expected:** Bertambah setelah worker berhasil memproses event.

---

# Traffic Contoh Untuk Memunculkan Metrics

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
  -H "Idempotency-Key: idem-day40-001" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"

:: List Transactions (Supervisor)
curl -i "http://localhost:8080/v1/transactions?limit=5&offset=0" ^
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>"
```

---

# Hasil Validasi Runtime & Penjelasan Q&A

Day 40 sudah tervalidasi dengan hasil:
- Semua target Prometheus `UP`
- Semua query metrics HTTP dan Worker berhasil dan menampilkan data.

## Interpretasi Hasil Worker Metrics

Hasil runtime menunjukkan:
- `finpharm_worker_events_total{result="success"} = 1`
- `sum by (result)` juga menunjukkan `success = 1`
- `finpharm_worker_inflight_messages = 0`
- `finpharm_worker_processing_duration_seconds_count = 1`

Artinya: worker endpoint metrics aktif, benar-benar memproses 1 event sukses, dan saat metrics di-scrape tidak ada message yang sedang aktif diproses.

## Kenapa Metric Worker Sempat Kosong?

Karena metric itu baru muncul kalau worker benar-benar memproses event dan memanggil increment counter. Kalau worker baru start, belum ada event baru, atau transaksi yang dikirim hanya replay response (idempotency key sama), maka metric itu bisa kosong.

## Penjelasan DLQ Yang Tetap Berisi 1 Message

Jika RabbitMQ menunjukkan:
- `transaction.approved.queue = 0`
- `transaction.approved.retry.queue = 0`
- `transaction.approved.dlq = 1`

Itu artinya ada 1 pesan gagal lama yang disimpan di DLQ (hasil pengujian invalid message sebelumnya). **Ini normal dan bukan bug.**

Pesan di DLQ tidak akan diproses ulang oleh worker utama sampai ada aksi manual. Selama queue utama tetap kosong atau diproses normal dan metric worker bertambah saat event baru diproses, DLQ = 1 hanya menandakan ada satu pesan lama untuk investigasi.

---

# Troubleshooting

**1. Targets DOWN**
- Cek service target benar-benar sedang jalan
- Cek port target benar
- Cek worker metrics port 9094 benar-benar aktif
- Cek Docker Desktop aktif

**2. Query Kosong**
- Belum ada traffic aplikasi
- Worker belum memproses event baru
- Transaksi yang dipakai hanya replay response karena idempotency key sama

**3. `host.docker.internal` tidak bisa dibuka di browser**
- Normal untuk host/browser. Gunakan `localhost` untuk akses manual dari Windows host.

---

# Checklist Hardening Akhir

Bagian ini dikumpulkan agar tidak terlupakan saat final hardening nanti.

## Observability Stack
- [ ] Tambahkan Grafana untuk dashboard lokal
- [ ] Tambahkan dashboard demo sederhana per service
- [ ] Pertimbangkan alert rules basic (target down, high latency, worker DLQ meningkat)
- [ ] Pertimbangkan recording rules untuk query penting

## Metrics Quality
- [ ] Evaluasi label path agar low-cardinality
- [ ] Tambah business metrics (total approved, flagged, pending review transactions)
- [ ] Tambah metric publish success/failure RabbitMQ
- [ ] Tambah metric ai-auditor fallback usage
- [ ] Tambah metric retry / DLQ worker yang lebih detail

## Deployment / DX
- [ ] Pertimbangkan satukan observability compose dengan stack lokal
- [ ] Tambahkan README observability
- [ ] Dokumentasikan semua port (8080, 8081, 8082, 8083, 9094, 9090)

## Future Tracing / Logging
- [ ] Sambungkan Prometheus journey ini dengan trace/log journey di day berikutnya
- [ ] Pertimbangkan OpenTelemetry di hardening akhir
- [ ] Pertimbangkan korelasi dashboard metrics dengan `trace_id` dan audit log

---

# Self-Review

- Kenapa Prometheus di Day 40 masih sesuai roadmap?
- Kenapa worker metric sempat kosong sebelum ada event baru?
- Kenapa replay idempotency tidak menambah worker metric?
- Kenapa DLQ yang berisi 1 message belum tentu bug?
- Kapan DLQ sebaiknya di-purge dan kapan dibiarkan?
- Metric apa yang paling bagus dipakai saat demo recruiter/reviewer?