# Day 44 — Final Hardening III: Observability Polish

## Tujuan

Memoles observability agar lebih siap demo, lebih business-aware, dan lebih production-like tanpa menambah fitur bisnis baru.

---

## Fokus Day 44

- menambah business metrics transaksi
- menambah audit decision metrics
- menambah AI auditor fallback metrics
- merapikan dashboard Grafana
- menambah panel observability yang lebih berguna untuk reviewer/demo

## Kenapa Day 44 Penting?

Sebelum Day 44, observability project ini sudah punya:
- HTTP metrics dasar
- worker metrics dasar
- Prometheus scrape
- Grafana dashboard awal

Namun dashboard masih lebih banyak menunjukkan “trafik teknis”. Day 44 membuat observability lebih berguna untuk menjelaskan:
- outcome transaksi
- keputusan audit
- fallback AI auditor
- error responses
- endpoint yang paling penting

---

## Scope Day 44

Day 44 mencakup:
- business metric transaksi
- audit metric transaksi
- replay metric transaksi
- AI auditor fallback metric
- panel error rate
- panel P99 latency
- panel transaction outcomes
- panel transaction audit decisions
- panel AI auditor fallback total
- rename panel worker processing count

**Scope Yang Belum Dikerjakan di Day 44:**
Karena file consumer worker untuk retry/DLQ increment belum dijadikan basis perubahan hari ini, Day 44 belum mencakup:
- worker retry metric spesifik
- worker DLQ metric spesifik
- worker `invalid_dlq` metric spesifik

Kalau nanti mau menambah bagian itu, titik perubahan paling aman ada di layer consumer worker.

---

## File Yang Ditambahkan / Diubah

### Transaction Service
```text
[REPLACE] services/transaction/internal/observability/metrics.go
[REPLACE] services/transaction/internal/usecase/transaction_usecase.go
```

### AI Auditor Service
```text
[REPLACE] services/ai-auditor/internal/observability/metrics.go
[REPLACE] services/ai-auditor/internal/usecase/audit_usecase.go
```

### Grafana
```text
[REPLACE] observability/grafana/dashboards/finpharm-overview.json
```

### Docs
```text
[NEW]     docs/day44.md
```

---

## Metric Baru Yang Ditambahkan

### Transaction Service
- `finpharm_transaction_outcomes_total{status}`
- `finpharm_transaction_audit_decisions_total{decision,provider,model}`
- `finpharm_transaction_replays_total`

Metric ini ditambahkan karena status final transaksi dan hasil audit diputuskan di transaction usecase.

### AI Auditor Service
- `finpharm_ai_auditor_decisions_total{decision,provider,model}`
- `finpharm_ai_auditor_fallback_total{reason}`

Metric ini ditambahkan karena primary/fallback behaviour terjadi di AI auditor usecase.

---

## Dashboard Update

### Panel Baru / Diperbarui
- Total HTTP Requests by Service
- HTTP Request Rate by Service
- HTTP 4xx/5xx by Service
- HTTP P95 Latency by Service
- HTTP P99 Latency by Service
- Transaction Outcomes by Status
- Transaction Audit Decisions
- Transaction Endpoint Traffic
- AI Auditor Request Count
- AI Auditor Fallback Total
- Worker Events by Result
- Worker Inflight Messages
- Worker Processed Events Count
- HTTP Requests Breakdown

### Rename Panel
- Dari `Worker Processing Count`
- Menjadi `Worker Processed Events Count`

Tujuannya supaya reviewer lebih cepat paham bahwa panel itu menunjukkan jumlah proses yang terekam histogram, bukan ukuran queue.

---

## Cara Test

### 1. Unit / Package Test
```powershell
go test ./services/transaction/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./... -count=1 -v
```

### 2. Generate Traffic
Gunakan script pembantu:
```powershell
.\scripts\generate-traffic.ps1
```
Atau kirim beberapa transaksi manual dengan status:
- `approved`
- `pending_review`
- `flagged`
- `replay`

---

## Validasi Manual di Prometheus

Coba jalankan query berikut di `http://localhost:9090`:

- **Transaction outcomes:** `finpharm_transaction_outcomes_total`
- **Transaction audit decisions:** `finpharm_transaction_audit_decisions_total`
- **Transaction replays:** `finpharm_transaction_replays_total`
- **AI auditor fallbacks:** `finpharm_ai_auditor_fallback_total`
- **AI auditor decisions:** `finpharm_ai_auditor_decisions_total`

---

## Validasi Manual di Grafana

Buka: `http://localhost:3000`
Lalu cek dashboard: **Finpharm Overview**

**Expected:**
- Panel *Transaction Outcomes* menampilkan status seperti: `APPROVED`, `PENDING_REVIEW`, `FLAGGED`, `FAILED` (bila pernah terjadi).
- Panel *Audit Decisions* menampilkan kombinasi `decision` dan `provider`.
- Panel *Fallback Total* naik saat primary AI provider gagal.
- Panel *HTTP Error Rate* tidak kosong bila ada 4xx/5xx traffic.
- Panel *Worker Processed Events Count* tetap bekerja.

---

## Interpretasi Hasil

Setelah Day 44:
- Dashboard tidak lagi hanya menunjukkan traffic teknis.
- Reviewer bisa melihat outcome bisnis transaksi.
- Fallback AI auditor bisa dijelaskan dengan angka.
- Error responses bisa dilihat lebih cepat.
- Endpoint penting transaction / ai-auditor lebih mudah dianalisis.

### Keterbatasan Saat Ini
- Worker retry / DLQ metrics spesifik belum ditambahkan pada Day 44.
- Metric worker yang sekarang masih bergantung pada result yang memang di-increment dari jalur worker saat ini.
- Dashboard masih satu file besar, belum dipisah modular per service.

---

## Checklist Hardening Lanjutan

Untuk Day 45 nanti, bagian yang masih layak dibawa:
- [ ] Screenshot dashboard di README
- [ ] Urutan demo observability final
- [ ] Login / credential notes
- [ ] Runbook observability yang lebih ringkas
- [ ] Release readiness final polish

---

## Self-Review

- Kenapa business metrics transaksi lebih cocok ditaruh di transaction service?
- Kenapa fallback metric lebih cocok ditaruh di AI auditor usecase?
- Kenapa P99 latency berguna selain P95?
- Kenapa panel 4xx/5xx penting untuk demo?
- Kenapa worker retry/DLQ metric belum dimasukkan pada pass ini?

---

## Catatan Tambahan: Kenapa Worker Retry/DLQ Tidak Dipaksa Sekarang?

Karena dari file saat ini, yang terlihat jelas baru:
- helper metric worker umum
- processor log domain event

Tetapi belum ada titik consumer yang pasti menaikkan result seperti `retry`, `dlq`, atau `invalid_dlq`. Jadi kalau pembaruan metrik tersebut dipaksa sekarang, risikonya justru membuat implementasi menjadi tidak presisi atau rusak pada repo Anda.

---

## Langkah Verifikasi Akhir

Setelah menempelkan file-file baru, jalankan perintah berikut:

```powershell
go test ./services/transaction/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./... -count=1 -v
```

Lalu:
1. Jalankan semua service.
2. Generate traffic lagi.
3. Buka Prometheus dan Grafana.
4. Cek panel/metric baru yang telah ditambahkan.