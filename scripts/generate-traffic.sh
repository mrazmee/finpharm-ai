#!/usr/bin/env bash
set -euo pipefail

command -v curl >/dev/null 2>&1 || { echo "curl is required"; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "python3 is required"; exit 1; }

staff_json="$(curl -s -X POST "http://localhost:8080/v1/auth/token" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"staff-001","role":"staff"}')"

staff_token="$(printf '%s' "$staff_json" | python3 -c 'import sys, json; print(json.load(sys.stdin)["data"]["access_token"])')"

supervisor_json="$(curl -s -X POST "http://localhost:8080/v1/auth/token" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"supervisor-001","role":"supervisor"}')"

supervisor_token="$(printf '%s' "$supervisor_json" | python3 -c 'import sys, json; print(json.load(sys.stdin)["data"]["access_token"])')"

echo "Traffic generation started..."

for i in $(seq 1 15); do
  trace_id="traffic-demo-$i"
  idem_key="idem-traffic-$i-$(python3 - <<'PY'
import uuid
print(uuid.uuid4().hex)
PY
)"

  curl -s -o /dev/null "http://localhost:8080/v1/medicines?limit=2&offset=0" \
    -H "Authorization: Bearer $staff_token" \
    -H "X-Trace-Id: $trace_id"

  curl -s -o /dev/null -X POST "http://localhost:8080/v1/stock/check" \
    -H "Authorization: Bearer $staff_token" \
    -H "X-Trace-Id: $trace_id" \
    -H "Content-Type: application/json" \
    -d '{"medicine_id":"PARA500","qty":1}'

  curl -s -o /dev/null -X POST "http://localhost:8080/v1/transactions" \
    -H "Authorization: Bearer $staff_token" \
    -H "X-Trace-Id: $trace_id" \
    -H "Idempotency-Key: $idem_key" \
    -H "Content-Type: application/json" \
    -d '{"items":[{"medicine_id":"PARA500","qty":1}]}'

  curl -s -o /dev/null "http://localhost:8080/v1/transactions?limit=5&offset=0" \
    -H "Authorization: Bearer $supervisor_token" \
    -H "X-Trace-Id: $trace_id"

  echo "Batch $i done"
  sleep 0.5
done

echo "Traffic generation finished."