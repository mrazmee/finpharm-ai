#!/usr/bin/env bash
set -euo pipefail

echo "[finpharm] running knowledge ingestion..."
go run ./services/knowledge/cmd/ingest
echo "[finpharm] knowledge ingestion complete"