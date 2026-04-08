# Day 38 — Metrics Prometheus (`/metrics`) per service

## Tujuan

Menambahkan metrics Prometheus agar setiap service punya observability dasar.

---

# Scope Day 38

## HTTP Services

Expose endpoint metrics pada port service yang sama:
- Gateway → `http://localhost:8080/metrics`
- Transaction Service → `http://localhost:8081/metrics`
- Inventory Service → `http://localhost:8082/metrics`
- AI Auditor Service → `http://localhost:8083/metrics`

## Worker

Karena worker tidak punya HTTP API utama, worker expose metrics server terpisah:
- Worker → `http://localhost:9094/metrics`

---

# Metrics Yang Ditambahkan

## HTTP Services

- `finpharm_http_requests_total`
- `finpharm_http_request_duration_seconds`
- `finpharm_http_inflight_requests`

Label utama:
- `service`
- `method`
- `path`
- `status`

## Worker

- `finpharm_worker_events_total`
- `finpharm_worker_processing_duration_seconds`
- `finpharm_worker_inflight_messages`

Result label di worker saat ini:
- `success`
- `duplicate`
- `invalid_dlq`
- `retry`
- `dlq`

---

# File Yang Ditambahkan / Diubah

## Gateway

```text
[ADD]     services/gateway/internal/observability/metrics.go
[REPLACE] services/gateway/cmd/api/main.go
```

## Transaction

```text
[ADD]     services/transaction/internal/observability/metrics.go
[REPLACE] services/transaction/cmd/api/main.go
```

## Inventory

```text
[ADD]     services/inventory/internal/observability/metrics.go
[REPLACE] services/inventory/cmd/api/main.go
```

## AI Auditor

```text
[ADD]     services/ai-auditor/internal/observability/metrics.go
[REPLACE] services/ai-auditor/cmd/api/main.go
```

## Worker

```text
[REPLACE] services/worker/internal/config/config.go
[ADD]     services/worker/internal/observability/metrics.go
[REPLACE] services/worker/internal/consumer/transaction_approved_consumer.go
[REPLACE] services/worker/cmd/worker/main.go
[REPLACE] scripts/run-worker.ps1
```

## Docs

```text
[REPLACE] docs/day38.md
```

---

# Persiapan Environment

## Dependency Baru

Jalankan perintah ini di root project:

```powershell
go get [github.com/prometheus/client_golang@v1.20.5](https://github.com/prometheus/client_golang@v1.20.5)
go mod tidy
```

---

# Cara Menjalankan

Jalankan semua service secara berurutan:

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

---

# Cara Verifikasi Manual

## 1. Cek Endpoint Metrics Sebelum Ada Traffic

```bash
curl http://localhost:8080/metrics
curl http://localhost:8081/metrics
curl http://localhost:8082/metrics
curl http://localhost:8083/metrics
curl http://localhost:9094/metrics
```

**Expected:**
- Pada HTTP services, awalnya bisa saja baru terlihat metrics bawaan: `go_*`, `process_*`, `promhttp_*`.
- Pada worker, jika worker sudah pernah memproses event, bisa langsung terlihat `finpharm_worker_*`.

---

## 2. Generate Traffic Aplikasi

Ambil token staff:
```bash
curl -i -X POST http://localhost:8080/v1/auth/token \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"staff-001\",\"role\":\"staff\"}"
```

Ambil token supervisor:
```bash
curl -i -X POST http://localhost:8080/v1/auth/token \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"supervisor-001\",\"role\":\"supervisor\"}"
```

Hit endpoint bisnis (gunakan token yang didapat):
```bash
curl -i "http://localhost:8080/v1/medicines?limit=2&offset=0" \
  -H "Authorization: Bearer <STAFF_TOKEN>"

curl -i -X POST http://localhost:8080/v1/stock/check \
  -H "Authorization: Bearer <STAFF_TOKEN>" \
  -H "Content-Type: application/json" \
  -d "{\"medicine_id\":\"PARA500\",\"qty\":1}"

curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Authorization: Bearer <STAFF_TOKEN>" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-metrics-001" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"

curl -i "http://localhost:8080/v1/transactions?limit=5&offset=0" \
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>"
```

---

## 3. Cek Metrics Lagi Setelah Ada Traffic

Gunakan mode silent supaya output lebih bersih:

```bash
curl -s http://localhost:8080/metrics | findstr finpharm_
curl -s http://localhost:8081/metrics | findstr finpharm_
curl -s http://localhost:8082/metrics | findstr finpharm_
curl -s http://localhost:8083/metrics | findstr finpharm_
curl -s http://localhost:9094/metrics | findstr finpharm_
```

**Expected:**

**Gateway**
Muncul metric seperti:
- `finpharm_http_requests_total{service="gateway", ...}`
- `finpharm_http_request_duration_seconds{service="gateway", ...}`
- `finpharm_http_inflight_requests{service="gateway"}`

**Transaction**
Muncul metric seperti:
- `finpharm_http_requests_total{service="transaction", ...}`
- `finpharm_http_request_duration_seconds{service="transaction", ...}`
- `finpharm_http_inflight_requests{service="transaction"}`

**Inventory**
Muncul metric seperti:
- `finpharm_http_requests_total{service="inventory", ...}`
- `finpharm_http_request_duration_seconds{service="inventory", ...}`
- `finpharm_http_inflight_requests{service="inventory"}`

**AI Auditor**
Muncul metric seperti:
- `finpharm_http_requests_total{service="ai-auditor", ...}`
- `finpharm_http_request_duration_seconds{service="ai-auditor", ...}`
- `finpharm_http_inflight_requests{service="ai-auditor"}`

**Worker**
Muncul metric seperti:
- `finpharm_worker_events_total`
- `finpharm_worker_processing_duration_seconds`
- `finpharm_worker_inflight_messages`

---

# Hasil Validasi Day 38 & Interpretasi

Day 38 sudah berhasil divalidasi dengan hasil:

- **Gateway metrics muncul untuk:** `POST /v1/auth/token`, `GET /v1/medicines`, `POST /v1/stock/check`, `POST /v1/transactions`, `GET /v1/transactions`
- **Transaction metrics muncul untuk:** `POST /v1/stock/check`, `POST /v1/transactions`, `GET /v1/transactions`
- **Inventory metrics muncul untuk:** `GET /v1/medicines`, `POST /v1/stock/check`, `POST /v1/stock/deduct`
- **AI Auditor metrics muncul untuk:** `POST /v1/audit/transaction`
- **Worker metrics muncul** pada port `9094` dan menunjukkan custom metric worker sudah berjalan.

**Interpretasi Hasil:**
- Jika pada awalnya HTTP service hanya menampilkan metrics bawaan Go/Prometheus, itu **normal**.
- Metrics `finpharm_http_*` baru benar-benar terlihat setelah ada traffic ke endpoint bisnis.
- Worker metrics bisa langsung terlihat jika worker sudah pernah memproses event sebelumnya.
- Durasi `POST /v1/transactions` bisa lebih tinggi karena endpoint ini menunggu proses audit AI.
- Jika `ai-auditor` punya durasi request tinggi, biasanya itu akan ikut terlihat di gateway/transaction pada jalur transaksi.

**Catatan Penting:**
Saat ini label path memakai URL Path mentah. Ini cukup untuk portfolio dan pembelajaran, tetapi untuk production nanti lebih baik memakai route pattern yang *low-cardinality* (Contoh: lebih baik `/v1/medicines/:id` daripada `/v1/medicines/PARA500`).

---

# Checklist Saran Kecil Untuk Final Hardening

Bagian ini sengaja ditulis agar tidak terlupakan di hari-hari akhir project.

## Observability
- [ ] Ubah label path dari raw URL menjadi route pattern low-cardinality
- [ ] Pertimbangkan tambah metric per dependency call (inventory, ai-auditor, rabbitmq publish)
- [ ] Pertimbangkan tambah business metrics (total approved, flagged, pending review transactions)
- [ ] Pertimbangkan tambah metric retry / dlq yang lebih detail di worker
- [ ] Pertimbangkan scrape config Prometheus lokal via Docker Compose
- [ ] Pertimbangkan dashboard Grafana sederhana untuk demo

## Logging & Tracing Mindset
- [ ] Pastikan semua service selalu log `request_id`
- [ ] Pastikan `request_id` dipropagasikan konsisten antar microservice
- [ ] Tambahkan audit log untuk perubahan status transaksi
- [ ] Pertimbangkan tracing header propagation sebagai persiapan distributed tracing

## Security
- [ ] Review endpoint mana yang seharusnya supervisor-only
- [ ] Pertimbangkan proteksi `/metrics` untuk environment non-local
- [ ] Pertimbangkan rate limiting di gateway
- [ ] Pertimbangkan masking data sensitif di log

## Reliability
- [ ] Validasi retry queue / DLQ lebih lanjut dengan skenario runtime failure injection
- [ ] Pertimbangkan idempotent consumer berbasis persistence (Redis/DB), bukan hanya in-memory
- [ ] Pertimbangkan transactional outbox untuk publish event yang lebih kuat
- [ ] Pertimbangkan health check dependency yang lebih detail

## Documentation & DX
- [ ] Tambahkan section observability di README utama
- [ ] Tambahkan daftar port service dan port metrics
- [ ] Tambahkan contoh command curl untuk `/metrics`
- [ ] Tambahkan penjelasan bahwa metrics custom baru terlihat setelah ada traffic
- [ ] Rapikan script demo supaya urutan validasi mudah direplikasi recruiter/reviewer

---

# Verifikasi Akhir (Test)

Jalankan semua test suite:

```powershell
go test ./services/gateway/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/inventory/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./services/worker/... -count=1 -v
go test ./... -count=1 -v
```

---

# Self-Review

- Kenapa worker perlu metrics server sendiri?
- Kenapa metrics custom HTTP service belum selalu muncul sebelum ada traffic?
- Kenapa request duration metric penting?
- Kenapa label path mentah bisa jadi masalah di production?
- Kenapa durasi transaksi bisa mengikuti durasi AI auditor?
- Metric apa yang paling berguna untuk worker?