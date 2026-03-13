# Day 23 — Gateway Proxy Create Transaction

## Tujuan
- Menambahkan endpoint `POST /v1/transactions` di `Gateway Service`.
- Menjadikan flow create transaction berjalan end-to-end lewat entry point yang benar: **Client -> Gateway -> Transaction Service**.
- Menambahkan validasi dasar di edge agar request yang jelas-jelas salah bisa ditolak lebih cepat.
- Menjaga peran Gateway tetap sehat: **proxy/routing only**, bukan tempat business logic transaksi.

## Yang dibangun / diubah
- Menambahkan `TransactionProxyHandler` baru di `Gateway Service`.
- Menambahkan route `POST /v1/transactions` pada router gateway.
- Menambahkan validasi dasar request create transaction di edge:
  - `items` wajib ada dan minimal 1 item
  - `medicine_id` tidak boleh kosong
  - `qty` harus `> 0`
- Menambahkan forward request ke `Transaction Service` dengan:
  - `context timeout`
  - propagasi `X-Request-ID`
  - header `X-Caller-Service: gateway`
- Menambahkan test untuk:
  - proxy create transaction sukses
  - validasi gateway gagal lebih awal tanpa memanggil upstream

## Konsep yang dipelajari
- Client idealnya tidak memanggil `Transaction Service` langsung, tetapi masuk lewat `Gateway` sebagai single entry point.
- Validasi dasar di gateway berguna untuk **fail fast**, tetapi validasi business rule yang lebih dalam tetap milik downstream service.
- Gateway boleh mengenal contract request/response, tetapi tidak boleh mengambil alih orchestration transaksi.
- Propagasi `request-id` tetap penting agar trace log dari Gateway ke Transaction Service tetap nyambung.

## Catatan desain penting hari ini
- Hari ini fokus pada **proxy dan edge validation**.
- Gateway tidak membuat transaction ID atau menyimpan transaksi.
- Flow create transaction masih sepenuhnya dilakukan oleh `Transaction Service`.
- Stock deduction belum dilakukan hari ini supaya debugging lebih mudah dan self-review tetap jelas.

## Penyesuaian roadmap setelah Day 22
- **Day 23**: Gateway proxy `POST /v1/transactions` + end-to-end create transaction
- **Day 24**: Pagination & filtering list transactions
- **Day 25**: Idempotency key untuk create transaction
- **Day 26**: Milestone 3 stabilization + README/API collection sync
- **Day 27 - 31**: AI Auditor Service + Gemini integration + status review flow
- **Day 32 - 35**: RAG + vector storage + citation *(optional advanced)*
- **Day 36 - 39**: RabbitMQ + worker + reliability pattern
- **Day 40 - 43**: JWT, metrics, audit logging, tracing mindset
- **Day 44 - 45**: final hardening, runbook, portfolio demo readiness

## File yang berubah / ditambah

### Gateway
- [ADD] `services/gateway/internal/httpapi/handler/transaction_proxy.go`
- [ADD] `services/gateway/internal/httpapi/handler/transaction_proxy_test.go`
- [MOD] `services/gateway/internal/httpapi/router.go`

### Docs
- [ADD] `docs/day23.md`

## Cara verifikasi

### 1. Pastikan PostgreSQL hidup
```bash
docker compose up -d postgres
```

### 2. Pastikan migration inventory dan transaction sudah naik
```powershell
.\scripts\migrate-inventory-up.ps1
.\scripts\migrate-transaction-up.ps1
```

### 3. Jalankan Inventory Service
```powershell
.\scripts\run-inventory.ps1
```

### 4. Jalankan Transaction Service
```powershell
.\scripts\run-transaction.ps1
```

### 5. Jalankan Gateway Service
```powershell
.\scripts\run-gateway.ps1
```

### 6. Kirim create transaction via Gateway
```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2},{\"medicine_id\":\"AMOX500\",\"qty\":1}]}"
```

Expected:
- status `201 Created`
- response tetap memakai envelope dari Transaction Service
- ada `id` transaksi
- status transaksi `PENDING`

### 7. Coba request invalid agar gateway fail fast
```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -d "{\"items\":[{\"medicine_id\":\"\",\"qty\":0}]}"
```

Expected:
- status `400 Bad Request`
- code error `VALIDATION_ERROR`
- request tidak diteruskan ke Transaction Service

### 8. Jalankan test gateway
```bash
go test ./services/gateway/... -count=1 -v
```

### 9. Jalankan semua test
```bash
go test ./... -count=1 -v
```

## Self-Review
- Kenapa `POST /v1/transactions` sebaiknya diekspos lewat Gateway, bukan client memanggil Transaction Service langsung?
- Kenapa validasi dasar request tetap berguna di Gateway walaupun Transaction Service juga melakukan validasi?
- Kenapa Gateway hanya forward request dan tidak ikut membuat transaction ID atau menyimpan transaksi?
- Kenapa propagasi request-id tetap harus dijaga pada flow create transaction?

## Verifikasi cepat
```bash
go test ./services/gateway/... -count=1 -v
go test ./... -count=1 -v
```

Lalu coba:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":2}]}"
```