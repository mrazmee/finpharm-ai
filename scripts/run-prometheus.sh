#!/usr/bin/env bash
set -euo pipefail

docker compose -f ./docker-compose.prometheus.yml up -d
docker compose -f ./docker-compose.prometheus.yml ps