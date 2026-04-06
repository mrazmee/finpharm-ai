# Day 35 — Worker Service Consume `transaction.approved`

## Tujuan

- Menambahkan service baru bernama `Worker Service`
- `Worker Service` consume event `transaction.approved` dari RabbitMQ
- Setelah consume, worker menjalankan aksi dummy:
  - log notifikasi
  - log report

---

# Scope Day 35

Hari ini fokusnya:

- worker sebagai aplikasi/binary terpisah
- worker connect ke RabbitMQ
- worker consume dari queue `transaction.approved.queue`
- worker parse payload event approved
- worker ack message setelah sukses diproses
- output worker masih dummy via log
- graceful shutdown worker sekarang memakai `ShutdownTimeout`

## Kenapa Worker Dibuat Terpisah?

Karena prinsip microservices yang kita pegang:
- setiap service = binary terpisah
- proses async dipisah dari request utama
- `Transaction Service` cukup publish event
- `Worker Service` yang ambil event dan proses pekerjaan lanjutan

Ini bikin sistem lebih realistis:
- request user tetap cepat
- pekerjaan sampingan jalan async
- coupling antar proses lebih longgar

---

# Flow Day 35

Flow-nya sekarang:

1. user create transaction
2. `Transaction Service` proses transaksi
3. kalau status final `APPROVED`, service publish `transaction.approved`
4. message masuk ke queue `transaction.approved.queue`
5. `Worker Service` consume message
6. worker log dummy notification
7. worker log dummy report
8. worker ack message

---

# Kenapa Queue Bisa Cepat Kembali Ke 0?

Karena kalau worker sedang hidup dan aktif consume:
- message bisa langsung diambil
- langsung diproses
- langsung di-ack

Akibatnya:
- queue count tidak selalu sempat terlihat naik lama
- indikator utama Day 35 adalah:
  - log worker
  - consumer count di RabbitMQ UI
  - acked message / queue kembali idle

## Catatan Penting Tentang Queue Count

Kalau worker aktif, kondisi seperti ini **normal**:
- transaksi `APPROVED` berhasil dibuat
- event berhasil dipublish
- queue di UI tetap terlihat `0`

Itu bukan berarti publish gagal. Itu biasanya berarti:
- worker langsung mengambil message
- worker langsung memproses
- worker langsung `ack`

Jadi untuk verifikasi Day 35, jangan hanya melihat angka queue. Lihat juga:
- log `worker_notification_sent`
- log `worker_report_generated`
- log `worker_message_acked`
- `Consumers = 1` di RabbitMQ UI

---

# Graceful Shutdown

Worker sekarang memakai `ShutdownTimeout`.

Artinya:
- saat menerima `CTRL+C` / `SIGTERM`
- worker tidak berhenti secara liar
- worker diberi waktu maksimal sesuai config untuk menyelesaikan proses shutdown

Kalau shutdown melebihi timeout:
- worker akan mencatat `worker_shutdown_timeout`

Ini lebih baik untuk service jangka panjang dibanding stop langsung tanpa batas waktu.

---

# File Yang Ditambahkan / Diubah

## Worker Service

```text
[ADD]     services/worker/internal/domain/event.go
[ADD]     services/worker/internal/config/config.go
[ADD]     services/worker/internal/processor/transaction_approved_processor.go
[ADD]     services/worker/internal/processor/transaction_approved_processor_test.go
[ADD]     services/worker/internal/consumer/transaction_approved_consumer.go
[ADD]     services/worker/internal/consumer/transaction_approved_consumer_test.go
[REPLACE] services/worker/cmd/worker/main.go
```

## Scripts & Docs

```text
[ADD]     scripts/run-worker.ps1
[REPLACE] docs/day35.md
```

---

# Dependency

Tidak ada dependency baru selain `github.com/rabbitmq/amqp091-go` yang sudah dipakai sejak Day 34.

---

# Cara Menjalankan

## 1. Pastikan RabbitMQ Hidup

```powershell
.\scripts\rabbitmq-up.ps1
```

## 2. Jalankan Service Utama

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-transaction.ps1
.\scripts\run-gateway.ps1
```

## 3. Jalankan Worker

```powershell
.\scripts\run-worker.ps1
```

---

# Cara Verifikasi

## Skenario A — Consume Existing Message

Kalau queue dari Day 34 masih punya 1 message approved:
- jalankan worker
- worker harus langsung consume message itu
- queue count turun jadi `0`

Expected log worker:
```text
worker_consumer_started
worker_notification_sent
worker_report_generated
worker_message_acked
```

## Skenario B — Publish Message Baru

Buat approved transaction baru:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-801" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:
- status `APPROVED`
- worker langsung memproses event
- queue bisa tetap 0 karena cepat dikonsumsi
- log worker bertambah

## Skenario C — Non-Approved Tidak Masuk Worker

Buat flagged transaction:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-802" \
  -d "{\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":2}]}"
```

Expected:
- status `FLAGGED`
- worker tidak memproses event baru (karena Day 34 memang hanya publish `transaction.approved`)

## RabbitMQ UI Yang Perlu Dicek

Di `http://localhost:15672`:
- tab **Queues and Streams**
- queue `transaction.approved.queue` ada
- saat worker hidup: `Consumers` pada queue bisa menjadi `1`
- setelah message diproses: `Ready` kembali `0`

## Jalankan Test

```bash
go test ./services/worker/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./... -count=1 -v
```

---

# Self-Review

- Kenapa worker dipisah jadi binary terpisah?
- Kenapa queue count tidak selalu terlihat naik walaupun publish berhasil?
- Kenapa `FLAGGED` transaction tidak diproses worker hari ini?
- Kenapa worker perlu ack message setelah sukses?
- Apa bedanya publish event di Day 34 dengan consume event di Day 35?
- Kenapa `ShutdownTimeout` penting untuk worker?

---

# Verifikasi Akhir

Setelah 2 koreksi ini, tidak ada perubahan behavior utama. Jadi yang disarankan cukup jalankan test:

```powershell
go test ./services/worker/... -count=1 -v
go test ./... -count=1 -v
```