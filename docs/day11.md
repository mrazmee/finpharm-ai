# Day 11 — Success Response Envelope (data + request_id)

## Tujuan
- Menstandarkan response sukses agar konsisten dan traceable:
  - semua response sukses dibungkus dengan envelope `{ "data": ..., "request_id": "..." }`
- Error response tetap seperti sebelumnya (punya request_id juga)

## Yang dibangun / diubah
- Menambahkan helper `RespondOK()` untuk response sukses dengan request_id
- Mengubah handler `stock` agar memakai `RespondOK()` (bukan `c.JSON(...)` langsung)
- Menyesuaikan test agar memvalidasi format envelope

## Konsep yang dipelajari
- API response envelope untuk konsistensi contract (success & error punya metadata)
- Traceability: request_id tidak hanya untuk error, tapi juga untuk success
- Gateway idealnya pass-through response dari downstream (tidak unwrap/wrap ulang tanpa alasan)

## File yang berubah / ditambah

### Transaction
- [MOD] services/transaction/internal/httpapi/handler/response.go  
  - tambah `SuccessResponse` + `RespondOK()`
- [MOD] services/transaction/internal/httpapi/handler/stock.go  
  - response sukses sekarang memakai `RespondOK()`
- [MOD] services/transaction/internal/httpapi/handler/stock_test.go  
  - cek response mengandung `"data"` dan `"request_id"`

### Gateway
- [MOD] services/gateway/internal/httpapi/handler/stock_proxy_test.go  
  - upstream fake response mengikuti envelope format

## Cara verifikasi
- Manual (via Gateway):
  - POST `http://localhost:8080/v1/stock/check`
  - body: `{"medicine_id":"PARA500","qty":10}`
  - expected:
    ```json
    {
      "data": { ... },
      "request_id": "..."
    }
    ```
- Testing:
  - `go test ./... -v` harus PASS

## Catatan / gotcha
- Ini perubahan contract response sukses (breaking change untuk client), jadi di project nyata biasanya dilakukan dengan versioning atau migrasi bertahap.