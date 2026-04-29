# Day 46 Addendum 2 — Deterministic Verification for Transaction Failed Alert

## Why this addendum exists
Alert `FinpharmTransactionFailedDetected` depends on:

`increase(finpharm_transaction_outcomes_total{status="FAILED"}[10m]) > 0`

In the normal transaction flow, `FAILED` only happens after:
1. validation passes
2. stock pre-check passes
3. transaction is created
4. audit does not stop the flow in review mode
5. stock deduction fails

That means black-box verification can be nondeterministic if we depend only on live timing and external service behavior.

## Design decision
Introduce two **local-only** fault injection flags:
- `TX_FORCE_AUDIT_APPROVED=true`
- `TX_FORCE_DEDUCT_FAILURE=true`

These are guarded by config validation so they are allowed only in local/dev environments.

## What the flags do
### `TX_FORCE_AUDIT_APPROVED`
Replaces the AI auditor repository with a local fixture that always returns:
- decision: `APPROVED`
- provider: `local-fixture`

### `TX_FORCE_DEDUCT_FAILURE`
Wraps the stock repository so:
- stock pre-check still calls the real inventory service
- but stock deduction always fails with an upstream error

## Why this is acceptable
This is not a production behavior change.
It is a deterministic local-only verification hook, similar in spirit to fault injection / failpoint testing:
- off by default
- blocked outside local/dev
- documented explicitly
- used only to verify alerting and resilience behavior

## Files changed
- `services/transaction/internal/config/config.go`
- `services/transaction/cmd/api/main.go`
- `services/transaction/internal/repository/fault_injection.go`
- `services/transaction/internal/repository/fault_injection_test.go`
- `services/transaction/internal/config/config_fault_injection_test.go`

## Verification plan
1. start transaction service with:
   - `TX_FORCE_AUDIT_APPROVED=true`
   - `TX_FORCE_DEDUCT_FAILURE=true`
2. send a normal create transaction request with valid stock and valid idempotency key
3. confirm API response becomes upstream failure / failed path
4. query:
   - `sum by (status) (finpharm_transaction_outcomes_total)`
   - `increase(finpharm_transaction_outcomes_total{status="FAILED"}[10m])`
5. wait 1 minute
6. confirm:
   - `ALERTS{alertname="FinpharmTransactionFailedDetected",alertstate=~"pending|firing"}`
   - Alertmanager shows the alert
   - webhook receives the alert payload