.PHONY: help \
	run-gateway run-transaction run-inventory run-ai-auditor run-worker \
	run-rabbitmq stop-rabbitmq logs-rabbitmq \
	run-prometheus stop-prometheus \
	run-grafana stop-grafana \
	reset-transaction-data demo-check demo-readiness demo-traffic \
	test test-gateway test-transaction test-inventory test-ai-auditor test-worker

help:
	@echo "Available targets:"
	@echo "  run-gateway             - Run gateway service"
	@echo "  run-transaction         - Run transaction service"
	@echo "  run-inventory           - Run inventory service"
	@echo "  run-ai-auditor          - Run ai-auditor service"
	@echo "  run-worker              - Run worker service"
	@echo "  run-rabbitmq            - Run RabbitMQ container"
	@echo "  stop-rabbitmq           - Stop RabbitMQ container"
	@echo "  logs-rabbitmq           - Tail RabbitMQ logs"
	@echo "  run-prometheus          - Run Prometheus container"
	@echo "  stop-prometheus         - Stop Prometheus container"
	@echo "  run-grafana             - Run Grafana container"
	@echo "  stop-grafana            - Stop Grafana container"
	@echo "  reset-transaction-data  - Reset transaction tables"
	@echo "  demo-readiness          - Show demo readiness checklist"
	@echo "  demo-traffic            - Generate demo traffic"
	@echo "  test                    - Run all tests"
	@echo "  test-gateway            - Run gateway tests"
	@echo "  test-transaction        - Run transaction tests"
	@echo "  test-inventory          - Run inventory tests"
	@echo "  test-ai-auditor         - Run ai-auditor tests"
	@echo "  test-worker             - Run worker tests"

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

run-rabbitmq:
	powershell -ExecutionPolicy Bypass -File .\scripts\rabbitmq-up.ps1

stop-rabbitmq:
	powershell -ExecutionPolicy Bypass -File .\scripts\rabbitmq-down.ps1

logs-rabbitmq:
	powershell -ExecutionPolicy Bypass -File .\scripts\rabbitmq-logs.ps1

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
	powershell -ExecutionPolicy Bypass -File .\scripts\demo-readiness.ps1

demo-readiness:
	powershell -ExecutionPolicy Bypass -File .\scripts\demo-readiness.ps1

demo-traffic:
	powershell -ExecutionPolicy Bypass -File .\scripts\generate-traffic.ps1

test:
	go test ./... -count=1 -v

test-gateway:
	go test ./services/gateway/... -count=1 -v

test-transaction:
	go test ./services/transaction/... -count=1 -v

test-inventory:
	go test ./services/inventory/... -count=1 -v

test-ai-auditor:
	go test ./services/ai-auditor/... -count=1 -v

test-worker:
	go test ./services/worker/... -count=1 -v