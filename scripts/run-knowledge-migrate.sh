#!/usr/bin/env bash
set -euo pipefail

echo "[finpharm] running knowledge migrations..."
go run ./services/knowledge/cmd/migrate
echo "[finpharm] knowledge migrations complete"