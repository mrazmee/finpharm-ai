#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="${RABBITMQ_CONTAINER_NAME:-finpharm-rabbitmq}"

docker logs "$CONTAINER_NAME" --tail 100 -f