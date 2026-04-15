.PHONY: help run-gateway run-transaction run-inventory run-ai-auditor run-worker run-prometheus run-grafana stop-prometheus stop-grafana reset-transaction-data demo-check

help:
	@echo "Available targets:"
	@echo "  run-gateway             - Run gateway service"
	@echo "  run-transaction         - Run transaction service"
	@echo "  run-inventory           - Run inventory service"
	@echo "  run-ai-auditor          - Run ai-auditor service"
	@echo "  run-worker              - Run worker service"
	@echo "  run-prometheus          - Run Prometheus container"
	@echo "  stop-prometheus         - Stop Prometheus container"
	@echo "  run-grafana             - Run Grafana container"
	@echo "  stop-grafana            - Stop Grafana container"
	@echo "  reset-transaction-data  - Reset transaction tables"
	@echo "  demo-check              - Print demo readiness checklist"

run-gateway:
	powershell -ExecutionPolicy Bypass -File .\scripts\run-gateway.ps1

run-transaction:
	powershell -ExecutionPolicy Bypass -File .\scripts\run-transaction.ps1

run-inventory:
	powershell -ExecutionPolicy Bypass -File .\scripts\run-inventory.ps1

run-ai-auditor:
	powershell -ExecutionPolicy Bypass -File .\scripts\run-ai-auditor.ps1

run-worker:
	powershell -ExecutionPolicy Bypass -File .\scripts\run-worker.ps1

run-prometheus:
	powershell -ExecutionPolicy Bypass -File .\scripts\run-prometheus.ps1

stop-prometheus:
	powershell -ExecutionPolicy Bypass -File .\scripts\stop-prometheus.ps1

run-grafana:
	powershell -ExecutionPolicy Bypass -File .\scripts\run-grafana.ps1

stop-grafana:
	powershell -ExecutionPolicy Bypass -File .\scripts\stop-grafana.ps1

reset-transaction-data:
	powershell -ExecutionPolicy Bypass -File .\scripts\reset-transaction-data.ps1

demo-check:
	@echo "Demo readiness checklist:"
	@echo "  [ ] inventory service running"
	@echo "  [ ] transaction service running"
	@echo "  [ ] ai-auditor service running"
	@echo "  [ ] gateway service running"
	@echo "  [ ] worker service running"
	@echo "  [ ] prometheus up on http://localhost:9090"
	@echo "  [ ] grafana up on http://localhost:3000"
	@echo "  [ ] rabbitmq up on http://localhost:15672"
	@echo "  [ ] transaction data reset if needed"
	@echo "  [ ] demo curl commands prepared"