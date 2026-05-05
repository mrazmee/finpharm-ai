# Day 30 — Integration & Verification

## Cek Kondisi Awal Stock

Sebelum memulai verifikasi transaksi, pastikan stock awal sudah sesuai:

```bash
curl -i -X POST http://localhost:8080/v1/stock/check \
  -H "Content-Type: application/json" \
  -d "{\"medicine_id\":\"OBATKERAS-X\",\"qty\":1}"
```

**Expected Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Request-Id: e6274c966ffad3dc6ea5c9d041870bc8
Date: Tue, 31 Mar 2026 07:47:00 GMT
Content-Length: 142

{"data":{"medicine_id":"OBATKERAS-X","requested_qty":1,"available_qty":5,"is_available":true},"request_id":"e6274c966ffad3dc6ea5c9d041870bc8"}
```

---

# Cara Menjalankan

Buka beberapa terminal berbeda dan jalankan semua service:

```powershell
.\scripts\run-inventory.ps1
.\scripts\run-ai-auditor.ps1
.\scripts\run-transaction.ps1
.\scripts\run-gateway.ps1
```

---

# Verifikasi

## 1. Low-Risk Transaction

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-301" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:

- status `201 Created`
- status `APPROVED`

---

## 2. High-Risk Transaction

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-302" \
  -d "{\"items\":[{\"medicine_id\":\"OBATKERAS-X\",\"qty\":2}]}"
```

Expected:

- status `201 Created`
- status `PENDING`

---

## 3. AI Auditor Mati

Matikan service `ai-auditor`, lalu jalankan:

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-303" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:

- status `201 Created`
- status `PENDING`

---

## 4. Replay

```bash
curl -i -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-301" \
  -d "{\"items\":[{\"medicine_id\":\"PARA500\",\"qty\":1}]}"
```

Expected:

- status `200 OK`
- status sama dengan transaksi sebelumnya

---

## 5. Filter Pending

```bash
curl -i "http://localhost:8080/v1/transactions?status=pending"
```

Expected:

- transaksi high-risk atau saat AI error muncul di list `PENDING`

---

## 6. Cek Stock High-Risk Tidak Berkurang

```bash
curl -i -X POST http://localhost:8080/v1/stock/check \
  -H "Content-Type: application/json" \
  -d "{\"medicine_id\":\"OBATKERAS-X\",\"qty\":1}"
```

Expected:

- stock tetap tidak berkurang untuk transaksi high-risk yang masih `PENDING`

---

## 7. Jalankan Semua Test

```bash
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./services/ai-auditor/... -count=1 -v
go test ./... -count=1 -v
```

---

# Hasil Verifikasi Yang Sudah Terbukti

Dari pengujian yang sudah dilakukan:

- low-risk transaction `PARA500` qty=1 berhasil menjadi `APPROVED`
- high-risk transaction `OBATKERAS-X` qty=2 tetap `PENDING`
- replay idempotency tetap mengembalikan transaction yang sama
- stock `OBATKERAS-X` tetap `5`, sehingga high-risk flow tidak mengurangi stock
- seluruh test suite tetap pass

---

# Self-Review

- Kenapa Day 30 belum perlu `PENDING_REVIEW`?
- Kenapa hasil audit `REVIEW` cukup ditahan di `PENDING` dulu?
- Kenapa AI error juga cukup ditahan di `PENDING` dulu?
- Kenapa stock hanya dikurangi setelah audit `APPROVED`?
- Kenapa replay tetap harus mengembalikan transaksi yang sama?