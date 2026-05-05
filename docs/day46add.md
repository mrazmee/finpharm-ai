# Day 46 Addendum — Gateway Rate Limit Instrumentation Fix

## Why This Change Was Needed

Alert `FinpharmGatewayHigh429Burst` depends on gateway metric `finpharm_http_requests_total{service="gateway",status="429"}`.

During verification:
- runtime requests clearly returned HTTP 429
- but Prometheus metric did not record status 429
- therefore the alert could not move to pending/firing

**Root cause:**
- rate limiting was implemented as an outer `net/http` wrapper in `main.go`
- gateway HTTP metrics were recorded inside Gin middleware
- rejected 429 requests returned before they reached the Gin observability pipeline

---

## Design Decision

Refactor rate limiting into **Gin middleware**.

This keeps cross-cutting concerns in one request pipeline:
- request ID
- logging
- metrics
- recovery
- rate limiting
- auth

This is a better fit for the application gateway architecture in this project than keeping rate limiting outside the router.

---

## Files Changed

- `services/gateway/cmd/api/main.go`
- `services/gateway/internal/httpapi/router.go`
- `services/gateway/internal/httpapi/middleware/rate_limit_http.go`
- `services/gateway/internal/httpapi/middleware/rate_limit_http_test.go`

---

## Expected Result

After this refactor:
- `429` responses are still returned to clients exactly as before
- but they are now also recorded by `finpharm_http_requests_total`
- alert `FinpharmGatewayHigh429Burst` becomes verifiable end-to-end

---

## Verification Plan

1. Run gateway
2. Spam `POST /v1/auth/token`
3. Confirm many runtime responses return `429`
4. Query in Prometheus:
   - `sum by (status) (finpharm_http_requests_total{service="gateway"})`
   - `increase(finpharm_http_requests_total{service="gateway",status="429"}[5m])`
5. Wait 1 minute
6. Confirm:
   - `ALERTS{alertname="FinpharmGatewayHigh429Burst",alertstate=~"pending|firing"}`
   - Alertmanager UI shows the alert
   - webhook logs receive the alert payload