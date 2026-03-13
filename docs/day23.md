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
- Pada microservices, client idealnya tidak memanggil `Transaction Service` langsung, tetapi masuk lewat `Gateway` sebagai single entry point.
- Validasi dasar di gateway berguna untuk **fail fast**, tetapi validasi business rule yang lebih dalam tetap milik downstream service.
- Gateway boleh mengenal contract request/response, tetapi tidak boleh mengambil alih orchestration transaksi.
- Propagasi `request-id` tetap penting agar trace log dari Gateway ke Transaction Service tetap nyambung.

## Kenapa Day 23 tidak sekalian deduct stock?
Karena fokus hari ini adalah **edge contract** dan **end-to-end proxy**.

Kalau hari ini langsung ditambah stock deduction, scope akan bercampur antara:
- routing/proxy di gateway
- orchestration lintas service
- consistency transaction vs inventory

Itu lebih aman dikerjakan terpisah di hari berikutnya supaya debugging dan self-review tetap jelas.

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