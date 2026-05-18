#!/usr/bin/env bash
set -euo pipefail

QUERY="${1:-}"
TOPK="${2:-5}"
MINSCORE="${3:-0.45}"

if [ -z "$QUERY" ]; then
  echo "usage: ./scripts/run-knowledge-query.sh "your question" [topk] [min_score]"
  exit 1
fi

echo "[finpharm] running knowledge query..."
go run ./services/knowledge/cmd/query -q "$QUERY" -k "$TOPK" -min-score "$MINSCORE"
echo "[finpharm] knowledge query complete"