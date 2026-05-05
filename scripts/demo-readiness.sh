#!/usr/bin/env bash
set -euo pipefail

echo "=== FINPHARM-AI DEMO READINESS ==="
echo "[ ] Inventory service running on :8082"
echo "[ ] Transaction service running on :8081"
echo "[ ] AI Auditor service running on :8083"
echo "[ ] Gateway service running on :8080"
echo "[ ] Worker running with metrics on :9094"
echo "[ ] Prometheus running on :9090"
echo "[ ] Grafana running on :3000"
echo "[ ] RabbitMQ running on :15672"
echo "[ ] Transaction data reset if needed"
echo "[ ] Fresh staff token ready"
echo "[ ] Fresh supervisor token ready"
echo "[ ] Demo transaction curl commands ready"
echo "[ ] Grafana dashboard opened"
echo "[ ] Prometheus targets all UP"