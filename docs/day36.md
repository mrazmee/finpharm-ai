# Day 36 — Reliability: Ack, Retry, Dead-Letter, Idempotent Consumer

## Tujuan

- Membuat `Worker Service` lebih tahan banting saat memproses event
- Menambahkan retry path
- Menambahkan dead-letter queue
- Menambahkan idempotent consumer sederhana

---

# Scope Day 36

Hari ini fokusnya:

- message sukses → `ack`
- message gagal → dipindah ke retry queue atau DLQ
- invalid message → langsung ke DLQ
- message yang sama tidak diproses dua kali kalau sudah sukses diproses sebelumnya
- topology RabbitMQ ditambah:
  - main queue
  - retry queue
  - dead-letter queue

---

# Topology RabbitMQ

Exchange tetap:
- `finpharm.events`

Queue yang dipakai sekarang:
- main queue: `transaction.approved.queue`
- retry queue: `transaction.approved.retry.queue`
- dead-letter queue: `transaction.approved.dlq`

Routing key:
- `transaction.approved`
- `transaction.approved.retry`
- `transaction.approved.dlq`

---

# Flow Day 36

## 1. Success Path
- worker process event
- worker log notification/report
- worker mark `transaction_id` sebagai processed
- worker `ack`

## 2. Invalid Message
Kalau payload invalid:
- worker tidak memproses
- worker publish ke DLQ
- worker `ack` message awal

## 3. Processing Failure
Kalau handler gagal:
- kalau retry count masih di bawah limit → publish ke retry queue
- kalau retry count sudah mencapai limit → publish ke DLQ
- lalu worker `ack` message awal

## 4. Duplicate Message
Kalau `transaction_id` yang sama sudah pernah sukses diproses:
- worker skip
- worker `ack`

---

# Q&A Desain Reliability

## Kenapa Pakai Idempotent Consumer?

Dalam sistem event-driven:
- publish ulang bisa terjadi
- worker restart bisa terjadi
- message duplicate bisa saja muncul

Kalau consumer tidak idempotent:
- notifikasi bisa dikirim dua kali
- report bisa tergenerate dua kali

Hari ini kita pakai versi sederhana:
- in-memory processed store
- key = `transaction_id`

Ini belum persisten lintas restart, tapi cukup untuk belajar konsep dengan jelas.

## Kenapa Gagal Dipublish Ulang ke Retry/DLQ Lalu Message Awal Di-ack?

Karena kita mau:
- menghindari loop `nack requeue` tak berujung
- memindahkan kontrol retry ke jalur yang lebih eksplisit
- membuat flow failure lebih mudah diobservasi

---

# File Yang Berubah / Ditambah

## Worker Service

```text
[REPLACE] services/worker/internal/config/config.go
[ADD]     services/worker/internal/consumer/idempotency_store.go
[REPLACE] services/worker/internal/consumer/transaction_approved_consumer.go
[REPLACE] services/worker/internal/consumer/transaction_approved_consumer_test.go
```

## Scripts & Docs

```text
[REPLACE] scripts/run-worker.ps1
[REPLACE] docs/day36.md
```

---

# Environment Baru

```env
RABBITMQ_TRANSACTION_APPROVED_RETRY_QUEUE=transaction.approved.retry.queue
RABBITMQ_TRANSACTION_APPROVED_RETRY_ROUTING_KEY=transaction.approved.retry
RABBITMQ_TRANSACTION_APPROVED_DLQ=transaction.approved.dlq
RABBITMQ_TRANSACTION_APPROVED_DLQ_ROUTING_KEY=transaction.approved.dlq
RABBITMQ_MAX_RETRY_COUNT=3
RABBITMQ_RETRY_DELAY_MS=5000
```

## Catatan Tentang Retry Delay

Hari ini `RABBITMQ_RETRY_DELAY_MS` baru disiapkan sebagai config untuk langkah berikutnya. Belum dipakai sebagai TTL-based delayed retry. 

Hari ini fokus dulu pada:
- jalur retry queue ada
- jalur DLQ ada
- ack/reroute flow jelas

---

# Cara Menjalankan

Jalankan semua service secara berurutan:

```powershell
.\scripts\rabbitmq-up.ps1
.\scripts\run-inventory.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-transaction.ps1
.\scripts\run-gateway.ps1
.\scripts\run-worker.ps1
```

---

# Cara Verifikasi Dasar

## 1. Queue Baru Muncul

Di RabbitMQ UI (`http://localhost:15672`), pastikan sekarang ada:
- `transaction.approved.queue`
- `transaction.approved.retry.queue`
- `transaction.approved.dlq`

## 2. Approved Message Tetap Diproses Normal

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-901" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:
- status `APPROVED`
- worker process lalu `ack`

## 3. Duplicate Handling

Kalau message approved yang sama masuk lagi dengan `transaction_id` yang sama:
- worker harus skip
- worker log `worker_duplicate_skipped`

## 4. Invalid Message ke DLQ

Kalau payload invalid dikirim ke `transaction.approved.queue`:
- worker harus log invalid message
- message harus masuk ke `transaction.approved.dlq`

---

# Validasi Manual Yang Sudah Berhasil Dilakukan

## A. Duplicate Message Berhasil Divalidasi

Manual publish ulang event approved dengan `transaction_id` yang sama menghasilkan log:
```text
worker_duplicate_skipped
```
Artinya:
- worker tidak memproses ulang event yang sudah pernah sukses diproses
- idempotent consumer berjalan dengan benar di runtime

## B. Invalid Message Berhasil Masuk DLQ

Manual publish payload invalid menghasilkan log:
```text
worker_invalid_message_to_dlq
```
dan queue:
```text
transaction.approved.dlq
```
bertambah menjadi `Ready = 1`, `Total = 1`. 

Artinya:
- invalid payload benar-benar tidak diproses normal
- invalid payload benar-benar diarahkan ke DLQ

---

# Hasil Verifikasi Day 36

Dengan hasil manual dan test yang sudah dijalankan, Day 36 sekarang sudah membuktikan:
- topology queue reliability berhasil dibuat
- approved path tetap aman
- duplicate message bisa di-skip
- invalid message masuk ke DLQ
- seluruh test tetap hijau

---

# Jalankan Semua Test

```bash
go test ./services/worker/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./... -count=1 -v
```

---

# Self-Review

- Kenapa `ack` lebih aman setelah reroute berhasil?
- Kenapa duplicate message perlu di-skip?
- Kenapa retry queue dan DLQ dipisah?
- Kenapa invalid message langsung masuk DLQ?
- Kenapa idempotent consumer in-memory cukup untuk tahap belajar, tapi belum cukup untuk production?
- Kenapa validasi manual duplicate dan DLQ penting meskipun unit test sudah hijau?