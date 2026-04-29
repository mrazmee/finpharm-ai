#!/usr/bin/env bash
set -euo pipefail

echo "[finpharm] stopping Alertmanager..."
docker compose -f docker-compose.alertmanager.yml down
echo "[finpharm] Alertmanager stopped"