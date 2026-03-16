# Day 24 — List Transactions with Pagination & Filtering

## Tujuan
- Menambahkan endpoint `GET /v1/transactions` di `Transaction Service`.
- Menambahkan proxy `GET /v1/transactions` di `Gateway Service`.
- Menampilkan daftar transaksi dari PostgreSQL lengkap dengan item-item transaksi.
- Menambahkan pagination dasar (`limit`, `offset`) dan filtering sederhana (`status`).
- Menjaga flow client tetap masuk lewat Gateway sebagai entry point utama.

## Yang dibangun / diubah
- Menambahkan contract domain untuk list transactions:
  - `ListTransactionsRequest`
  - `ListTransactionsResult`
- Menambahkan method `List(...)` di `TransactionRepository`.
- Menambahkan method `ListTransactions(...)` di `TransactionUsecase`.
- Menambahkan query SQL untuk:
  - count total transaksi
  - ambil header transaksi sesuai pagination/filter
  - ambil seluruh item dari transaksi yang terambil pada page tersebut
- Menambahkan handler HTTP `GET /v1/transactions` di `Transaction Service`.
- Menambahkan proxy `GET /v1/transactions` di `Gateway Service`.
- Menambahkan test untuk:
  - list transactions handler
  - query param validation
  - gateway proxy list transactions

## Konsep yang dipelajari
- Pagination penting agar endpoint list tidak mengambil semua data sekaligus.
- Query list sering perlu memisahkan:
  - query `COUNT(*)`
  - query data page saat ini
- Untuk response transaksi, tetap butuh `transaction_items`, jadi repository melakukan query kedua untuk menggabungkan item berdasarkan `transaction_id`.
- Filter `status` sengaja dibuat sederhana dulu agar contract stabil sebelum masuk idempotency dan transaction lifecycle yang lebih kompleks.
- Gateway tetap hanya meneruskan request dan tidak ikut membaca database.

## Desain response hari ini
Response `GET /v1/transactions` memakai envelope standar:

```json
{
  "data": {
    "items": [
      {
        "id": "TXN-20260313064640-CC2116B4",
        "status": "PENDING",
        "items": [
          { "medicine_id": "PARA500", "qty": 2 },
          { "medicine_id": "AMOX500", "qty": 1 }
        ],
        "created_at": "2026-03-13T06:46:40.955307Z"
      }
    ],
    "limit": 10,
    "offset": 0,
    "total": 1
  },
  "request_id": "..."
}
```

## Catatan desain penting hari ini
- Hari ini fokus pada **list, pagination, dan filter sederhana**.
- Gateway tetap bertindak sebagai **single entry point**, tidak melakukan logika business atau query database.
- Repository `sqlx` menangani query pagination + items join agar response konsisten.
- Filter `status` dibuat sederhana agar contract stabil untuk development awal.

## File yang berubah / ditambah

### Transaction
- [ADD] `services/transaction/internal/domain/list_transactions.go`
- [MOD] `services/transaction/internal/usecase/transaction_usecase.go`
- [ADD] `services/transaction/internal/usecase/transaction_usecase_list_test.go`
- [ADD] `services/transaction/internal/httpapi/handler/transaction_list.go`
- [MOD] `services/transaction/internal/httpapi/router.go`

### Gateway
- [ADD] `services/gateway/internal/httpapi/handler/transaction_list_proxy.go`
- [ADD] `services/gateway/internal/httpapi/handler/transaction_list_proxy_test.go`
- [MOD] `services/gateway/internal/httpapi/router.go`

### Docs
- [ADD] `docs/day24.md`

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

### 6. List transactions via Gateway
```bash
curl -i -X GET "http://localhost:8080/v1/transactions?limit=10&offset=0&status=PENDING" \
  -H "Content-Type: application/json"
```

Expected:
- status `200 OK`
- response memuat field `data.items` dengan transaksi beserta `transaction_items`
- `limit`, `offset`, dan `total` sesuai query param

### 7. Jalankan test
```bash
go test ./services/transaction/... -count=1 -v
go test ./services/gateway/... -count=1 -v
go test ./... -count=1 -v
```

## Self-Review
- Kenapa pagination itu penting untuk list transactions di microservices?
- Kenapa repository perlu memisahkan query `COUNT(*)` dan query page data?
- Kenapa filter status tetap dikerjakan di repository, bukan di Gateway?
- Kenapa Gateway hanya forward request tanpa ikut baca database?  
- Bagaimana response format standar membantu konsistensi client dan tracing request-id?