# Day 46 — Final Alerting Foundation & Verification

## Objective
Menyelesaikan alerting end-to-end pada stack observability FinPharm-AI agar sistem tidak hanya bisa dipantau lewat metrics dan dashboard, tetapi juga mampu:
- mengevaluasi alert rules di Prometheus
- merutekan alert ke Alertmanager
- mengirim notifikasi ke local webhook receiver
- membuktikan alert benar-benar firing melalui skenario verifikasi terkontrol

## Outcome
Day 46 berhasil menutup gap observability yang sebelumnya masih ada pada area alerting. Hasil akhirnya:
- Prometheus terhubung ke Alertmanager
- Alertmanager terhubung ke local webhook receiver
- alert rules tersimpan sebagai code di repo
- semua alert utama berhasil diverifikasi end-to-end
- blind spot observability pada HTTP 429 gateway berhasil diperbaiki
- verification path untuk transaction failure dibuat deterministik dan local-only

## Scope yang Diselesaikan

### 1. Alerting foundation
- menambahkan blok `alerting` pada Prometheus
- menambahkan Alertmanager config
- menambahkan docker compose Alertmanager
- menambahkan local webhook receiver
- menambahkan script run/stop Alertmanager dan webhook
- menambahkan target Makefile untuk alerting

### 2. Verification end-to-end
Alert berikut berhasil dibuktikan benar-benar firing:
1. `FinpharmServiceDown`
2. `FinpharmGatewayHigh5xxRatio`
3. `FinpharmGatewayHigh429Burst`
4. `FinpharmAIAuditorFallbackDetected`
5. `FinpharmWorkerDLQDetected`
6. `FinpharmTransactionFailedDetected`

### 3. Quality improvements during verification
- route template metrics sudah stabil dari tahap pre-Day46 cleanup
- alert `GatewayHigh5xxRatio` diperjelas ownership-nya dengan label `service=gateway`
- blind spot observability untuk response `429` diperbaiki dengan memindahkan rate limiting ke Gin middleware
- verifikasi transaction failure dibuat deterministik dengan local-only fault injection

## Files Changed During Day 46

### Alerting foundation
- `observability/prometheus/prometheus.yml`
- `observability/alertmanager/alertmanager.yml`
- `docker-compose.alertmanager.yml`
- `cmd/alert-webhook/main.go`
- `scripts/run-alertmanager.ps1`
- `scripts/stop-alertmanager.ps1`
- `scripts/run-alertmanager.sh`
- `scripts/stop-alertmanager.sh`
- `scripts/run-alert-webhook.ps1`
- `scripts/run-alert-webhook.sh`
- `Makefile`

### Alert rules and readability improvements
- `observability/prometheus/alerts/finpharm.rules.yml`

### Gateway 429 instrumentation fix
- `services/gateway/cmd/api/main.go`
- `services/gateway/internal/httpapi/router.go`
- `services/gateway/internal/httpapi/middleware/rate_limit_http.go`
- `services/gateway/internal/httpapi/middleware/rate_limit_http_test.go`

### Transaction failed deterministic verification
- `services/transaction/internal/config/config.go`
- `services/transaction/internal/config/config_fault_injection_test.go`
- `services/transaction/internal/repository/fault_injection.go`
- `services/transaction/internal/repository/fault_injection_test.go`
- `services/transaction/cmd/api/main.go`

## Final Alert Rules Covered

### 1. `FinpharmServiceDown`
**Intent:** mendeteksi service yang tidak bisa di-scrape oleh Prometheus.  
**Verification:** mematikan `inventory`, lalu memastikan:
- `up{job="inventory"} = 0`
- alert masuk `pending` lalu `firing`
- Alertmanager menerima alert
- webhook menerima payload firing
- setelah service hidup kembali, alert resolve

### 2. `FinpharmGatewayHigh5xxRatio`
**Intent:** mendeteksi gateway yang banyak mengembalikan 5xx ke client.  
**Verification:** mematikan `transaction`, lalu memicu request melalui gateway ke endpoint yang bergantung pada upstream tersebut hingga rasio 5xx melebihi threshold.  
**Catatan:** label `service=gateway` ditambahkan agar alert lebih mudah dibaca dan ownership komponen lebih jelas.

### 3. `FinpharmGatewayHigh429Burst`
**Intent:** mendeteksi lonjakan rate limiting pada gateway.  
**Verification:** spam endpoint auth token hingga response 429 muncul dan benar-benar tercatat di metric gateway.  
**Important fix:** sebelumnya 429 tidak masuk metric karena rate limiter berada di luar pipeline Gin observability. Ini diperbaiki dengan memindahkan rate limiting ke Gin middleware.

### 4. `FinpharmAIAuditorFallbackDetected`
**Intent:** mendeteksi fallback dari AI auditor saat provider utama gagal.  
**Verification:** menjalankan ai-auditor dengan konfigurasi yang memicu fallback, lalu mengirim beberapa audit request sampai counter fallback naik dan alert firing.

### 5. `FinpharmWorkerDLQDetected`
**Intent:** mendeteksi pesan yang masuk ke DLQ / invalid DLQ pada worker.  
**Verification:** mengirim payload invalid ke exchange RabbitMQ sehingga worker memproses jalur `invalid_dlq`, counter naik, dan alert firing.

### 6. `FinpharmTransactionFailedDetected`
**Intent:** mendeteksi transaksi yang mencapai status `FAILED`.  
**Verification:** menggunakan local-only deterministic verification hook agar audit selalu approved dan deduct stock selalu gagal, sehingga status transaction secara konsisten masuk ke `FAILED` dan alert bisa diverifikasi tanpa bergantung pada race condition acak.

## Why Gateway 429 Needed an Additional Fix
Selama verifikasi awal, runtime request memang menghasilkan HTTP 429, tetapi metric gateway tidak mencatat status 429.  
Akar masalahnya:
- rate limiting saat itu diimplementasikan sebagai outer `net/http` wrapper
- observability metric dicatat di Gin middleware
- request yang ditolak lebih awal tidak pernah melewati middleware metric

Perbaikan:
- refactor rate limiting menjadi Gin middleware
- hasil akhirnya:
  - response 429 tetap sama ke client
  - tetapi sekarang juga masuk ke metric `finpharm_http_requests_total`
  - alert `FinpharmGatewayHigh429Burst` menjadi bisa diverifikasi end-to-end

## Why Transaction Failed Needed Deterministic Verification
Jalur `FAILED` pada transaction service tidak mudah dipicu secara black-box biasa karena status ini baru terjadi setelah:
1. validasi lolos
2. pre-check stok lolos
3. transaction dibuat
4. audit tidak berhenti di review
5. deduct stock gagal

Kalau mengandalkan timing atau race condition, verifikasi akan tidak stabil.  
Untuk itu ditambahkan **local-only fault injection**:

- `TX_FORCE_AUDIT_APPROVED=true`
- `TX_FORCE_DEDUCT_FAILURE=true`

### Safeguards
Fault injection ini:
- **off by default**
- hanya boleh aktif di environment local/dev
- ditolak oleh `Validate()` bila diaktifkan di environment non-local
- dipakai hanya untuk verifikasi alert / resilience path secara deterministik

### Why this is acceptable
Ini bukan hack sementara, tetapi local-only verification hook yang terkontrol.  
Untuk portfolio production-minded, pendekatan ini justru lebih baik daripada:
- mengandalkan skenario gagal acak
- memalsukan verifikasi
- menghapus alert yang sulit dibuktikan

## Verification Matrix

| Alert | Status | Evidence |
|---|---|---|
| `FinpharmServiceDown` | Verified | `up=0`, alert firing, Alertmanager group, webhook firing/resolved |
| `FinpharmGatewayHigh5xxRatio` | Verified | raw 502 metric naik, ratio > threshold, alert firing, webhook firing |
| `FinpharmGatewayHigh429Burst` | Verified | 429 runtime produced, 429 metric recorded, `increase(...) > 10`, alert firing |
| `FinpharmAIAuditorFallbackDetected` | Verified | fallback counter naik, `increase(...) > 0`, alert firing |
| `FinpharmWorkerDLQDetected` | Verified | worker `invalid_dlq` metric naik, alert firing |
| `FinpharmTransactionFailedDetected` | Verified | transaction outcome `FAILED` naik, `increase(...) > 0`, alert firing |

## Operational Notes After Verification

### Return services to normal mode
Setelah verifikasi selesai:
- transaction service harus dijalankan kembali **tanpa**
  - `TX_FORCE_AUDIT_APPROVED`
  - `TX_FORCE_DEDUCT_FAILURE`

Karena dua flag itu hanya untuk verifikasi lokal.

### Alert resolve behavior
Beberapa alert berbasis:
- `increase(...[10m])`
- `for: 1m`

Karena itu, setelah skenario verifikasi dihentikan, alert tidak selalu langsung hilang.  
Selama event masih berada dalam window query, alert masih dapat tetap `firing`.  
Ini adalah perilaku yang benar, bukan bug.

## Why This Day Matters for Portfolio
Day 46 membuat observability project ini naik kelas dari:
- “punya metrics dan dashboard”

menjadi:
- “punya alerting pipeline yang benar-benar bisa mendeteksi gangguan operasional”

Nilai yang terlihat ke recruiter:
- pemisahan tanggung jawab Prometheus vs Alertmanager vs webhook receiver
- alerting as code
- end-to-end verification, bukan sekadar config load
- kemampuan menemukan blind spot observability lalu memperbaikinya
- penggunaan deterministic local-only verification hook untuk failure path yang sulit diuji

Ini memberi kesan bahwa project tidak dibuat sekadar untuk demo happy path, tetapi sudah dipikirkan dari sudut reliability dan operability.

## Recommended Commit Message
```txt
feat(observability): finalize day 46 alerting foundation and verification
```

atau jika ingin lebih eksplisit:

```txt
feat(alerting): add alertmanager, webhook receiver, gateway 429 instrumentation, and transaction failed verification hooks
```

## Suggested README / RUNBOOK Notes
Tambahkan ringkasan singkat bahwa:
- Alertmanager digunakan untuk routing alert dari Prometheus
- local webhook receiver dipakai sebagai proof of delivery saat demo
- verification hook transaction bersifat local-only dan tidak aktif default
- rate limiter gateway sekarang berada di Gin middleware agar response 429 ikut tercatat ke metrics

## Final Status
**Day 46 complete.**

Setelah ini repo siap lanjut ke tahap berikutnya dengan pondasi alerting yang:
- version-controlled
- demoable
- verifiable
- production-minded
