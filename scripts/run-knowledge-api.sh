#!/usr/bin/env bash
set -euo pipefail

echo "[finpharm] running knowledge api..."
go run ./services/knowledge/cmd/api
echo "[finpharm] knowledge api stopped"