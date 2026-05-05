# Day 13 — Propagate X-Request-ID to Inventory (Full Trace Chain)

## Tujuan
- Memastikan `X-Request-ID` terpropagate penuh:
  - client → gateway → transaction → inventory
- Menambah traceability antar service (log correlation end-to-end)

## Yang dibangun / diubah
- Transaction sekarang meneruskan `X-Request-ID` saat memanggil Inventory melalui `StockHTTPRepo`
- Inventory logger membaca `from_service` untuk memastikan asal request
- Transaction test diperbarui untuk memastikan `X-Request-ID` benar-benar dikirim ke upstream (mock inventory)

## Konsep yang dipelajari
- Request correlation antar microservices (request_id)
- Observability: satu request harus bisa ditelusuri lintas service lewat log
- Teknik testing dependency service dengan `httptest.NewServer` (mock upstream)

## File yang berubah / ditambah

### Transaction
- [MOD] services/transaction/internal/repository/stock_http_repo.go  
  - menambahkan header `X-Request-ID` ketika call Inventory
- [MOD] services/transaction/internal/httpapi/handler/stock.go  
  - mengambil request_id dari middleware context dan menginjeksikan ke repo HTTP
- [MOD] services/transaction/internal/httpapi/router.go  
  - DI disederhanakan: handler menerima `InventoryBaseURL`
- [MOD] services/transaction/internal/httpapi/handler/stock_test.go  
  - menambah test `TestCheckStock_OK_PropagatesRequestIDToInventory`

### Inventory
- (tidak wajib berubah kode) tetapi log menunjukkan field `from_service:"transaction"` sudah terbaca dari header `X-From-Service`

## Cara verifikasi
### Manual (3 service running)
- Hit endpoint via gateway: `POST /v1/stock/check`
- Pastikan log di 3 service punya `request_id` yang sama:
  - gateway: request_id = X
  - transaction: request_id = X
  - inventory: request_id = X
- Pastikan inventory log juga menampilkan `from_service:"transaction"`

### Testing
- `go test ./... -v` harus PASS

## Catatan
- Request-ID tidak disimpan di `context.Context` (belum), jadi propagation dilakukan dengan mengambilnya dari Gin context middleware lalu disisipkan ke header outbound.