# Day 14 — Resilience Basics: Downstream Timeout + Upstream Error Mapping (502)

## Tujuan
- Menambahkan ketahanan dasar saat Transaction bergantung pada Inventory:
  - call ke Inventory punya timeout policy yang jelas
  - jika Inventory down/timeout → Transaction merespon **502 Bad Gateway** (bukan 500)
- Menambah test untuk skenario downstream lambat/timeout.

## Yang dibangun / diubah
- Tambah domain error baru: `UpstreamError` (untuk dependency failure)
- `StockHTTPRepo`:
  - memakai `context.WithTimeout` per-call (SLA call inventory)
  - network error / timeout dimapping jadi `UpstreamError`
- Handler Transaction:
  - `UpstreamError` dimapping ke HTTP **502** dengan `code: UPSTREAM_ERROR`
- Test:
  - menambah test mock inventory yang sleep lebih lama dari timeout untuk memastikan return 502 dan timeout terjadi lebih cepat.

## Konsep yang dipelajari
- Perbedaan error internal (500) vs error dependency/upstream (502)
- Timeout policy per downstream call
- Cara mengetes timeout dengan `httptest.NewServer` + `time.Sleep`

## File yang berubah / ditambah

### Transaction
- [MOD] services/transaction/internal/domain/errors.go  
  - tambah `UpstreamError` + helper `IsUpstream`
- [MOD] services/transaction/internal/repository/stock_http_repo.go  
  - tambah per-call timeout + mapping upstream errors
- [MOD] services/transaction/internal/httpapi/handler/stock.go  
  - mapping `UpstreamError` → 502 `UPSTREAM_ERROR`
- [MOD] services/transaction/internal/httpapi/handler/stock_test.go  
  - tambah test `TestCheckStock_InventoryTimeout_Returns502`

## Cara verifikasi
### Manual
- Matikan Inventory service
- Hit via gateway: `POST /v1/stock/check`
- Expected:
  - HTTP 502
  - `code: UPSTREAM_ERROR`
  - ada `request_id`

### Testing
- `go test ./... -v` harus PASS

## Catatan / gotcha
- Pada test timeout, mock server tetap tidur walau client sudah timeout—ini normal karena server-side handler tetap berjalan sampai selesai.