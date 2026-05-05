# Day 17 — Gateway Routing to Inventory (Medicines List & Detail)

## Tujuan
- Membuat client cukup berinteraksi lewat **Gateway**, termasuk untuk fitur Inventory.
- Menambahkan proxy route di Gateway untuk:
  - `GET /v1/medicines`
  - `GET /v1/medicines/:id`
- Menjaga contract tetap konsisten (Gateway pass-through response envelope dari Inventory).
- Propagate request-id ke upstream.

## Yang dibangun / diubah
- Gateway config menambahkan `INVENTORY_BASE_URL` (default `http://localhost:8082`).
- Gateway menambahkan handler proxy baru untuk Inventory:
  - meneruskan query string (limit/offset)
  - meneruskan path param `:id`
  - propagate header `X-Request-ID`
  - menandai asal request via `X-From-Gateway: finpharm-gateway`
- Menambahkan unit test gateway untuk memastikan proxy berjalan dengan upstream mock (httptest).

## Konsep yang dipelajari
- API Gateway routing pattern (edge service sebagai single entrypoint).
- Pass-through design: gateway tidak perlu unwrap/wrap ulang response dari downstream.
- Pengujian proxy handler dengan upstream mock server.

## File yang berubah / ditambah

### Gateway
- [MOD] services/gateway/internal/config/config.go  
  - tambah `InventoryBaseURL`
- [ADD] services/gateway/internal/httpapi/handler/inventory_proxy.go  
  - proxy endpoint medicines list & detail
- [MOD] services/gateway/internal/httpapi/router.go  
  - register routes `/v1/medicines` dan `/v1/medicines/:id`
- [ADD] services/gateway/internal/httpapi/handler/inventory_proxy_test.go  
  - test proxy list & detail via upstream mock

## Cara verifikasi
### Manual
- Jalankan 3 service: gateway, transaction, inventory
- Hit via gateway:
  - `GET http://localhost:8080/v1/medicines?limit=2&offset=0`
  - `GET http://localhost:8080/v1/medicines/PARA500`
  - `GET http://localhost:8080/v1/medicines/PARA200` (expected 404)

### Testing
- `go test ./... -v` harus PASS

## Catatan
- Log `path` pada route param bisa tampil sebagai `/v1/medicines/:id` (template route) dan itu normal.