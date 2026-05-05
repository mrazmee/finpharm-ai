#!/usr/bin/env bash
set -euo pipefail

echo "[finpharm] starting local alert webhook on http://localhost:18080"
go run ./cmd/alert-webhook