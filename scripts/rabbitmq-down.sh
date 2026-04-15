#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${RABBITMQ_COMPOSE_FILE:-./docker-compose.rabbitmq.yml}"

docker compose -f "$COMPOSE_FILE" down