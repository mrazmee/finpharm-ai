#!/usr/bin/env bash
set -euo pipefail

export APP_ENV="${APP_ENV:-local}"
go run ./services/inventory/cmd/api