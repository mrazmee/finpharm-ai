# Day 39 — Audit Logging + Tracing Mindset

## Tujuan

Menambahkan:
- audit logging lintas service
- correlation tracing lintas HTTP service
- fondasi observability yang lebih matang tanpa menambah schema database

---

# Prinsip Day 39

Day 39 sengaja dibuat **aman** untuk codebase saat ini:
- tidak menambah migration / tabel audit log
- tidak menambah dependency tracing besar seperti OpenTelemetry
- fokus pada:
  - audit log di level aplikasi
  - `X-Trace-Id` untuk correlation
  - propagasi trace dan actor identity lintas HTTP service

---

# Scope Day 39

## 1. Audit Logging

Audit log ditambahkan untuk request penting:
- **Semua mutation HTTP:**
  - `POST`
  - `PUT`
  - `PATCH`
  - `DELETE`
- **Protected read / important read:**
  - `GET /v1/transactions`
  - `GET /v1/debug/sleep`

Field utama audit log:
- `audited_service`
- `method`
- `path`
- `status`
- `request_id`
- `trace_id`
- `user_id`
- `role`
- `remote_addr`

## 2. Tracing Mindset

Ditambahkan `X-Trace-Id`:
- jika request sudah punya `X-Trace-Id`, service akan pakai itu
- jika belum ada, service akan fallback ke `X-Request-Id`
- jika `X-Request-Id` juga belum ada, service akan generate trace id baru

Response juga mengembalikan:
- `X-Trace-Id`

## 3. Propagation Yang Dikerjakan

**Gateway → Downstream**
Gateway sekarang meneruskan header berikut ke downstream:
- `X-Request-Id`
- `X-Trace-Id`
- `X-User-Id`
- `X-User-Role`
- `X-Caller-Service`

**Transaction → Dependency HTTP**
Transaction service sekarang meneruskan:
- `X-Request-Id`
- `X-Trace-Id`
- `X-User-Id`
- `X-User-Role`
- `X-Caller-Service`

Ke service berikut:
- inventory service
- ai-auditor service

---

# Kenapa Audit Log Belum Disimpan Ke Database?

Karena untuk tahap ini kita prioritaskan:
- konsistensi lintas service
- kemudahan validasi manual
- tidak menambah migration baru
- tidak mengganggu fitur yang sudah stabil

Jadi Day 39 fokus pada app-level audit logging, correlation tracing, dan runtime verification. Persistent audit store bisa menjadi bagian hardening di akhir.

---

# File Yang Ditambahkan / Diubah

## Shared Telemetry

```text
[ADD] internal/telemetry/tracehttp/tracehttp.go
[ADD] internal/telemetry/tracehttp/tracehttp_test.go
[ADD] internal/telemetry/audithttp/audithttp.go
```

## Gateway

```text
[REPLACE] services/gateway/cmd/api/main.go
[ADD]     services/gateway/internal/httpapi/handler/proxy_headers.go
[REPLACE] services/gateway/internal/httpapi/handler/inventory_proxy.go
[REPLACE] services/gateway/internal/httpapi/handler/stock_proxy.go
[REPLACE] services/gateway/internal/httpapi/handler/transaction_proxy.go
```

## Transaction

```text
[REPLACE] services/transaction/cmd/api/main.go
[ADD]     services/transaction/internal/repository/http_headers.go
[REPLACE] services/transaction/internal/repository/stock_http_repo.go
[REPLACE] services/transaction/internal/repository/ai_auditor_http_repo.go
```

## Inventory, AI Auditor, Worker, Docs

```text
[REPLACE] services/inventory/cmd/api/main.go
[REPLACE] services/ai-auditor/cmd/api/main.go
[REPLACE] services/worker/internal/processor/transaction_approved_processor.go
[REPLACE] docs/day39.md
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

## 1. Ambil Token Staff

```cmd
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"staff-001\",\"role\":\"staff\"}"
```

## 2. Ambil Token Supervisor

```cmd
curl -i -X POST http://localhost:8080/v1/auth/token ^
  -H "Content-Type: application/json" ^
  -d "{\"user_id\":\"supervisor-001\",\"role\":\"supervisor\"}"
```

## 3. Uji Create Transaction Dengan Trace ID Manual

```cmd
curl -i -X POST http://localhost:8080/v1/transactions ^
  -H "Authorization: Bearer <STAFF_TOKEN>" ^
  -H "X-Trace-Id: demo-trace-day39-004" ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: idem-day39-manual-004" ^
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

**Expected:**
Response headers membawa:
```text
X-Request-Id: <generated_id>
X-Trace-Id: demo-trace-day39-004
```

## 4. Uji Supervisor List Transactions

```cmd
curl -i "http://localhost:8080/v1/transactions?limit=5&offset=0" ^
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>" ^
  -H "X-Trace-Id: demo-trace-day39-002"
```

**Expected:**
```text
HTTP/1.1 200 OK
X-Trace-Id: demo-trace-day39-002
```

## 5. Uji Supervisor Debug Route

```cmd
curl -i "http://localhost:8080/v1/debug/sleep?ms=1" ^
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>" ^
  -H "X-Trace-Id: demo-trace-day39-003"
```

**Expected:**
```text
HTTP/1.1 200 OK
X-Trace-Id: demo-trace-day39-003
```

---

# Hasil Validasi Day 39

## Validasi Supervisor Path

Sudah tervalidasi bahwa gateway audit log untuk `GET /v1/transactions` membawa:
- `trace_id = demo-trace-day39-002`
- `user_id = supervisor-001`
- `role = supervisor`

Transaction audit log untuk request yang sama juga membawa:
- `trace_id = demo-trace-day39-002`
- `user_id = supervisor-001`
- `role = supervisor`

Selain itu, `GET /v1/debug/sleep` oleh supervisor juga berhasil dan tercatat dengan `trace_id = demo-trace-day39-003`.

## Validasi Create Transaction Path

Semua service HTTP di jalur `POST /v1/transactions` (gateway, transaction, inventory, ai-auditor) memakai:
- `trace_id = demo-trace-day39-004`
- `user_id = staff-001`
- `role = staff`

*Jadi propagation HTTP lintas service sudah berhasil.*

## Validasi Worker

Worker sudah memiliki:
- `audit_domain_event`
- log domain saat approved event diproses
- notification/report dummy log

Namun worker **belum** ikut membawa `trace_id` yang sama dari jalur HTTP sebelumnya. 
Artinya:
- jalur HTTP sudah *one-trace correlation*
- jalur async RabbitMQ → worker belum berada dalam satu trace yang sama.

---

# Interpretasi Hasil & Tanya Jawab

## Day 39 Saat Ini Berada Pada Level:

**Sudah Ada:**
- audit logging lintas service
- correlation tracing lintas HTTP service
- actor propagation (user_id, role) lintas HTTP service
- domain event logging di worker

**Belum Ada:**
- async trace propagation penuh sampai worker
- span model / distributed tracing penuh
- persistent audit log store

## Apakah Ini Sudah Sesuai Praktik Industri?

**Ya**, untuk level dasar–menengah, ini sudah realistis dan sesuai praktik industri. Yang sekarang kita punya adalah *correlation tracing*, bukan full distributed tracing seperti OpenTelemetry lengkap. 

Itu tetap sangat berguna karena:
- memudahkan debugging lintas service
- cukup kuat untuk portfolio
- cocok untuk sistem yang belum butuh tracing stack besar

---

# Jalankan Test

```powershell
go test ./internal/telemetry/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/inventory/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./services/worker/... -count=1 -v
go test ./... -count=1 -v
```

---

# Checklist Hardening Akhir

Bagian ini sengaja dikumpulkan agar tidak terlupakan saat final hardening.

## A. Audit Logging Hardening
- [ ] Tentukan event audit domain yang wajib disimpan permanen
- [ ] Tambahkan persistent audit store (DB / log sink / storage)
- [ ] Tambahkan struktur audit log yang lebih business-oriented:
  - `resource`
  - `action`
  - `result`
- [ ] Tambahkan actor metadata yang lebih kaya:
  - `user_id`
  - `role`
  - `source_service`
  - `ip`
- [ ] Review field audit log agar tidak menyimpan data sensitif berlebihan

## B. Trace Propagation Hardening
- [ ] Tambahkan `trace_id` ke event RabbitMQ saat publish `transaction.approved`
- [ ] Baca `trace_id` itu di worker saat consume message
- [ ] Masukkan `trace_id` ke log worker (`audit_domain_event`, `worker_notification_sent`, dll)
- [ ] Pastikan jalur async worker ikut satu trace correlation dengan jalur HTTP

## C. Distributed Tracing Evolution
- [ ] Pertimbangkan migrasi dari custom `X-Trace-Id` ke header standar seperti `traceparent`
- [ ] Pertimbangkan OpenTelemetry untuk full distributed tracing
- [ ] Tambahkan konsep `span_id` bila nanti tracing dinaikkan levelnya
- [ ] Tambahkan timing/span untuk dependency call (Gateway → Transaction, Transaction → Inventory, dll)

## D. Security Hardening
- [ ] Masking token/header sensitif di log
- [ ] Review apakah `/debug/sleep` harus tetap ada di environment tertentu
- [ ] Review `/metrics` apakah tetap public di non-local environment
- [ ] Review audit log agar tidak membocorkan informasi sensitif

## E. Reliability Hardening
- [ ] Pertimbangkan persistent idempotent consumer untuk worker
- [ ] Pertimbangkan transactional outbox untuk publish event
- [ ] Pertimbangkan structured retry reason yang lebih jelas
- [ ] Pertimbangkan DLQ observability yang lebih kaya

## F. Documentation Hardening
- [ ] Tambahkan contoh log audit di README
- [ ] Tambahkan penjelasan `request_id` vs `trace_id`
- [ ] Tambahkan penjelasan bahwa Day 39 saat ini sudah full di jalur HTTP, tetapi async worker belum satu trace
- [ ] Tambahkan bagian “future evolution to OpenTelemetry”

---

# Self-Review

- Kenapa audit log penting pada domain farmasi?
- Apa perbedaan `request_id` dan `trace_id`?
- Kenapa HTTP propagation sudah dianggap berhasil?
- Kenapa worker belum dianggap satu trace penuh?
- Apa bedanya *correlation tracing* dan *distributed tracing*?
- Kapan sistem perlu naik dari custom trace header ke OpenTelemetry?