#!/usr/bin/env bash
set -euo pipefail

QUERY="${1:-}"
TOPK="${2:-5}"
MINSCORE="${3:-0.45}"

if [ -z "$QUERY" ]; then
  echo "usage: ./scripts/run-knowledge-answer.sh \"your question\" [topk] [min_score]"
  exit 1
fi

echo "[finpharm] running knowledge answer..."
go run ./services/knowledge/cmd/answer -q "$QUERY" -k "$TOPK" -min-score "$MINSCORE"
echo "[finpharm] knowledge answer complete"