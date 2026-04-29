#!/usr/bin/env bash
set -euo pipefail

echo "[finpharm] starting Alertmanager..."
docker compose -f docker-compose.alertmanager.yml up -d
echo "[finpharm] Alertmanager started at http://localhost:9093"