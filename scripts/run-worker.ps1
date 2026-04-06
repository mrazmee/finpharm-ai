if ([string]::IsNullOrEmpty($env:APP_ENV)) { $env:APP_ENV = "local" }
if ([string]::IsNullOrEmpty($env:WORKER_NAME)) { $env:WORKER_NAME = "notification-worker" }

if ([string]::IsNullOrEmpty($env:RABBITMQ_URL)) { $env:RABBITMQ_URL = "amqp://finpharm:finpharm@localhost:5672/" }
if ([string]::IsNullOrEmpty($env:RABBITMQ_EXCHANGE)) { $env:RABBITMQ_EXCHANGE = "finpharm.events" }
if ([string]::IsNullOrEmpty($env:RABBITMQ_TRANSACTION_APPROVED_QUEUE)) { $env:RABBITMQ_TRANSACTION_APPROVED_QUEUE = "transaction.approved.queue" }
if ([string]::IsNullOrEmpty($env:RABBITMQ_TRANSACTION_APPROVED_ROUTING_KEY)) { $env:RABBITMQ_TRANSACTION_APPROVED_ROUTING_KEY = "transaction.approved" }
if ([string]::IsNullOrEmpty($env:RABBITMQ_CONSUMER_TAG)) { $env:RABBITMQ_CONSUMER_TAG = "worker.transaction.approved" }
if ([string]::IsNullOrEmpty($env:RABBITMQ_PREFETCH_COUNT)) { $env:RABBITMQ_PREFETCH_COUNT = "10" }
if ([string]::IsNullOrEmpty($env:SHUTDOWN_TIMEOUT_MS)) { $env:SHUTDOWN_TIMEOUT_MS = "7000" }

go run .\services\worker\cmd\worker