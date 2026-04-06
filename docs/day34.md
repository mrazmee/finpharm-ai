# Day 34 — Transaction Publish Event: `transaction.approved`

## Tujuan

- Menjadikan `Transaction Service` sebagai publisher event ke RabbitMQ
- Publish event hanya untuk transaksi yang sudah final `APPROVED`
- Menyiapkan topology dasar RabbitMQ untuk event transaksi approved

---

# Scope Day 34

Hari ini fokusnya:

- `Transaction Service` publish event `transaction.approved`
- event dipublish ke RabbitMQ exchange `finpharm.events`
- queue dasar dibuat otomatis: `transaction.approved.queue`
- routing key: `transaction.approved`
- hanya status `APPROVED` yang publish event
- `FLAGGED`, `PENDING_REVIEW`, dan `FAILED` **tidak** publish event

## Kenapa Hanya APPROVED Dulu?

Karena kita ingin mulai dari event yang paling jelas secara domain:
- transaksi sudah selesai
- stock sudah dideduct
- status final sudah stabil

Kalau langsung publish semua status sekaligus, scope hari ini jadi terlalu lebar.

## Kenapa Publish Failure Belum Membuat Request Gagal?

Untuk Day 34 kita sengaja pakai pendekatan **best effort publish**:
- transaksi tetap sukses kalau status final sudah `APPROVED`
- kalau publish ke RabbitMQ gagal, service hanya log warning

Kenapa? Karena:
- outbox pattern belum dibuat
- retry/reliability yang lebih kuat belum dibuat
- kita sedang belajar publisher dulu

Jadi fokus hari ini adalah memastikan format event benar, topology RabbitMQ benar, dan approved transaction benar-benar menghasilkan message.

---

# Topology RabbitMQ

Hari ini publisher akan memastikan topology ini ada:

- exchange: `finpharm.events`
- exchange type: `direct`
- queue: `transaction.approved.queue`
- routing key: `transaction.approved`

---

# Bentuk Event

Contoh payload event yang dipublish:

```json
{
  "event_name": "transaction.approved",
  "transaction_id": "TXN-20260401120000-AAAA1111",
  "idempotency_key": "idem-701",
  "status": "APPROVED",
  "items": [
    {
      "medicine_id": "PARA500",
      "qty": 1
    }
  ],
  "audit": {
    "decision": "APPROVED",
    "risk_score": 0.1,
    "reason": "Single common over-the-counter medication, low quantity.",
    "provider": "gemini",
    "model": "gemini-2.5-flash",
    "audited_at": "2026-04-01T12:00:05Z"
  },
  "created_at": "2026-04-01T12:00:00Z",
  "published_at": "2026-04-01T12:00:06Z"
}
```

---

# File Yang Berubah / Ditambah

## Transaction Service

```text
[ADD]     services/transaction/internal/domain/transaction_event.go
[ADD]     services/transaction/internal/repository/rabbitmq_event_publisher.go
[REPLACE] services/transaction/internal/config/config.go
[REPLACE] services/transaction/internal/usecase/transaction_usecase.go
[REPLACE] services/transaction/internal/usecase/transaction_usecase_test.go
[REPLACE] services/transaction/cmd/api/main.go
```

## Scripts & Docs

```text
[REPLACE] scripts/run-transaction.ps1
[ADD]     docs/day34.md
```

---

# Persiapan Environment

## Dependency Baru

Jalankan perintah ini di root project:

```bash
go get [github.com/rabbitmq/amqp091-go@v1.10.0](https://github.com/rabbitmq/amqp091-go@v1.10.0)
go mod tidy
```

## Env Baru

`Transaction Service` sekarang memakai:

```env
RABBITMQ_URL=amqp://finpharm:finpharm@localhost:5672/
RABBITMQ_EXCHANGE=finpharm.events
RABBITMQ_TRANSACTION_APPROVED_QUEUE=transaction.approved.queue
RABBITMQ_TRANSACTION_APPROVED_ROUTING_KEY=transaction.approved
```

---

# Cara Verifikasi

## 1. Pastikan RabbitMQ Hidup

```powershell
.\scripts\rabbitmq-up.ps1
```

## 2. Jalankan Service

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-transaction.ps1
.\scripts\run-gateway.ps1
```

## 3. Buat Transaksi Low-Risk

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-701" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:
- status `APPROVED`

## 4. Cek Queue di RabbitMQ UI

Buka browser: `http://localhost:15672` (login: `finpharm` / `finpharm`) lalu buka tab **Queues and Streams**.

Expected:
- queue `transaction.approved.queue` ada
- jumlah message bertambah setelah transaction approved

## 5. Cek Transaksi Non-Approved (Tidak Publish)

**Skenario High-Risk:**
```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-702" \
  -d "{\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":2}]}"
```
Expected:
- status `FLAGGED`
- queue `transaction.approved.queue` **tidak bertambah** karena transaksi ini bukan approved.

**Skenario AI Unavailable:**
Matikan `ai-auditor`, lalu jalankan:
```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-703" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```
Expected:
- status `PENDING_REVIEW`
- queue **tidak bertambah**.

## 6. Jalankan Semua Test

```bash
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./... -count=1 -v
```

---

# Self-Review

- Kenapa hanya `APPROVED` yang publish event?
- Kenapa event publish failure belum membuat request gagal?
- Kenapa queue `transaction.approved.queue` dibuat sekarang meskipun consumer belum dibuat?
- Apa beda status transaction di database dengan event di RabbitMQ?
- Kenapa Day 34 belum menyelesaikan reliability secara penuh?

---

# Urutan Kerjanya Sekarang

**1. Install dependency baru**
```powershell
go get [github.com/rabbitmq/amqp091-go@v1.10.0](https://github.com/rabbitmq/amqp091-go@v1.10.0)
go mod tidy
```

**2. Replace/add file-file yang disebutkan di atas**

**3. Jalankan test**
```powershell
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./... -count=1 -v
```

**4. Nyalakan RabbitMQ dan semua service**
```powershell
.\scripts\rabbitmq-up.ps1
.\scripts\run-inventory.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-transaction.ps1
.\scripts\run-gateway.ps1
```

**5. Verifikasi approved publish event**
Buat 1 transaksi approved, lalu pastikan pesan masuk ke queue melalui RabbitMQ UI.