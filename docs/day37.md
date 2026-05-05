# Day 37 — JWT Auth di Gateway + Role Staff/Supervisor

## Tujuan

- Menambahkan auth di **Gateway**
- Menggunakan JWT bearer token
- Menambahkan kontrol role:
  - `staff`
  - `supervisor`

---

# Prinsip Day 37

Karena Gateway adalah *edge service*, auth memang paling tepat dipasang di sini:
- validasi token di edge
- role check di edge
- downstream service tetap fokus pada business logic

Gateway juga akan meneruskan identity ke downstream lewat header:
- `X-User-ID`
- `X-User-Role`

---

# Scope Route Protection

## Public
- `GET /health`
- `POST /v1/auth/token` → local/dev only

## Staff atau Supervisor
- `GET /v1/medicines`
- `GET /v1/medicines/:id`
- `POST /v1/stock/check`
- `POST /v1/transactions`

## Supervisor Only
- `GET /v1/transactions`
- `GET /v1/debug/sleep` (local/dev)

---

# Q&A Konfigurasi

## Kenapa `AUTH_ENABLED` Dibuat Via Env?

Supaya:
- runtime lokal benar-benar memakai auth
- test proxy lama tidak perlu dirombak total sekaligus

Jadi:
- config default `AUTH_ENABLED=false`
- `scripts/run-gateway.ps1` mengaktifkannya menjadi `true`

---

# Persiapan Environment

## Dependency Baru

Jalankan perintah ini di root project:

```powershell
go get [github.com/golang-jwt/jwt/v5@v5.2.1](https://github.com/golang-jwt/jwt/v5@v5.2.1)
go mod tidy
```

## Environment Baru

```env
AUTH_ENABLED=true
JWT_SECRET=finpharm-local-secret
JWT_ISSUER=finpharm-gateway
JWT_EXPIRE_MINUTES=60
```

---

# File Yang Berubah / Ditambah

## Gateway

```text
[REPLACE] services/gateway/internal/config/config.go
[NEW]     services/gateway/internal/httpapi/middleware/auth.go
[NEW]     services/gateway/internal/httpapi/handler/auth.go
[REPLACE] services/gateway/internal/httpapi/router.go
```

## Tests

```text
[NEW]     services/gateway/internal/httpapi/middleware/auth_test.go
[NEW]     services/gateway/internal/httpapi/handler/auth_test.go
[NEW]     services/gateway/internal/httpapi/router_auth_test.go
```

## Scripts & Docs

```text
[REPLACE] scripts/run-gateway.ps1
[REPLACE] docs/day37.md
```

---

# Cara Menjalankan

Jalankan semua service secara berurutan:

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-transaction.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-gateway.ps1
```

---

# Cara Ambil Token Local

## 1. Token Staff

```bash
curl -i -X POST http://localhost:8080/v1/auth/token \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"staff-001\",\"role\":\"staff\"}"
```

## 2. Token Supervisor

```bash
curl -i -X POST http://localhost:8080/v1/auth/token \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"supervisor-001\",\"role\":\"supervisor\"}"
```

---

# Contoh Pemakaian Token

*(Ganti `<STAFF_TOKEN>` atau `<SUPERVISOR_TOKEN>` dengan token asli yang didapat dari perintah sebelumnya)*

## 1. Staff Boleh Create Transaction

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Authorization: Bearer <STAFF_TOKEN>" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-day37-001" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

## 2. Staff Tidak Boleh List Transactions

```bash
curl -i "http://localhost:8080/v1/transactions" \
  -H "Authorization: Bearer <STAFF_TOKEN>"
```

Expected:
- `403 Forbidden`

## 3. Supervisor Boleh List Transactions

```bash
curl -i "http://localhost:8080/v1/transactions" \
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>"
```

Expected:
- `200 OK`

---

# Manual Verification Tambahan Yang Disarankan

## 1. Tanpa Token (Harus 401)

```bash
curl -i "http://localhost:8080/v1/transactions"
```
Expected: `401 Unauthorized`

## 2. Token Invalid (Harus 401)

```bash
curl -i "http://localhost:8080/v1/transactions" \
  -H "Authorization: Bearer invalid.token.value"
```
Expected: `401 Unauthorized`

## 3. Staff Tidak Boleh Akses Debug Route

```bash
curl -i "http://localhost:8080/v1/debug/sleep?ms=1" \
  -H "Authorization: Bearer <STAFF_TOKEN>"
```
Expected: `403 Forbidden`

## 4. Supervisor Boleh Akses Debug Route

```bash
curl -i "http://localhost:8080/v1/debug/sleep?ms=1" \
  -H "Authorization: Bearer <SUPERVISOR_TOKEN>"
```
Expected: `200 OK`

---

# Validasi Yang Sudah Tercapai

Dengan test dan pengecekan runtime, Day 37 sudah membuktikan:

- token `staff` dan `supervisor` berhasil dibuat
- `staff` bisa create transaction
- `staff` tidak bisa list transactions
- `supervisor` bisa list transactions
- middleware auth dan role check lolos test
- router-level auth protection juga dites langsung

---

# Jalankan Semua Test

```powershell
go test ./services/gateway/... -count=1 -v
go test ./services/transaction/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./services/worker/... -count=1 -v
go test ./... -count=1 -v
```

---

# Self-Review

- Kenapa auth paling tepat dipasang di Gateway?
- Kenapa `GET /v1/transactions` lebih cocok supervisor-only?
- Kenapa identity diteruskan ke downstream via header?
- Kenapa endpoint issue token dibatasi local/dev?
- Apa bedanya *authentication* dan *authorization* di implementasi ini?
- Kenapa validasi 401 dan debug-route role check membantu memperkuat Day 37?