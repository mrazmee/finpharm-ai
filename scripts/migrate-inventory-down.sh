#!/usr/bin/env bash
set -euo pipefail

: "${INVENTORY_DB_DSN:?set INVENTORY_DB_DSN first}"
MIGRATIONS_DIR="${INVENTORY_MIGRATIONS_DIR:-./services/inventory/migrations}"

migrate -path "$MIGRATIONS_DIR" -database "$INVENTORY_DB_DSN" down 1