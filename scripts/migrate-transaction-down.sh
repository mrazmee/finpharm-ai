#!/usr/bin/env bash
set -euo pipefail

: "${TRANSACTION_DB_DSN:?set TRANSACTION_DB_DSN first}"
MIGRATIONS_DIR="${TRANSACTION_MIGRATIONS_DIR:-./services/transaction/migrations}"

migrate -path "$MIGRATIONS_DIR" -database "$TRANSACTION_DB_DSN" down 1