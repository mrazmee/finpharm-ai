package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"finpharm-ai/services/worker/internal/config"
	"finpharm-ai/services/worker/internal/domain"
	"finpharm-ai/services/worker/internal/observability"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TransactionApprovedHandler interface {
	HandleTransactionApproved(ctx context.Context, event domain.TransactionApprovedEvent) error
}

type TransactionApprovedConsumer struct {
	cfg            config.Config
	conn           *amqp.Connection
	handler        TransactionApprovedHandler
	processedStore *InMemoryProcessedStore
}

func NewTransactionApprovedConsumer(cfg config.Config, handler TransactionApprovedHandler) (*TransactionApprovedConsumer, error) {
	if strings.TrimSpace(cfg.RabbitMQURL) == "" {
		return nil, fmt.Errorf("rabbitmq url is required")
	}
	if handler == nil {
		return nil, fmt.Errorf("transaction approved handler is required")
	}

	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	c := &TransactionApprovedConsumer{
		cfg:            cfg,
		conn:           conn,
		handler:        handler,
		processedStore: NewInMemoryProcessedStore(),
	}

	if err := c.ensureTopology(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return c, nil
}

func (c *TransactionApprovedConsumer) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *TransactionApprovedConsumer) Run(ctx context.Context) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()

	if err := ch.Qos(c.cfg.PrefetchCount, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	deliveries, err := ch.Consume(
		c.cfg.QueueName,
		c.cfg.ConsumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("start consume from queue %q: %w", c.cfg.QueueName, err)
	}

	slog.Info("worker_consumer_started",
		"queue", c.cfg.QueueName,
		"routing_key", c.cfg.RoutingKey,
		"consumer_tag", c.cfg.ConsumerTag,
		"prefetch_count", c.cfg.PrefetchCount,
		"max_retry_count", c.cfg.MaxRetryCount,
		"retry_queue", c.cfg.RetryQueueName,
		"dlq", c.cfg.DLQName,
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("worker_consumer_stopping", "queue", c.cfg.QueueName)
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("deliveries channel closed")
			}

			done := observability.BeginProcessing()

			event, err := decodeTransactionApprovedEvent(delivery.Body)
			if err != nil {
				slog.Warn("worker_invalid_message_to_dlq",
					"queue", c.cfg.QueueName,
					"error", err,
				)
				observability.IncResult("invalid_dlq")

				if dlqErr := c.publishToRoutingKey(ctx, c.cfg.DLQRoutingKey, delivery.Body, withRetryHeader(delivery.Headers, currentRetryCount(delivery.Headers))); dlqErr != nil {
					done()
					return fmt.Errorf("publish invalid message to dlq: %w", dlqErr)
				}
				if ackErr := delivery.Ack(false); ackErr != nil {
					done()
					return fmt.Errorf("ack invalid message: %w", ackErr)
				}

				done()
				continue
			}

			if c.processedStore.Exists(event.TransactionID) {
				slog.Info("worker_duplicate_skipped",
					"transaction_id", event.TransactionID,
					"queue", c.cfg.QueueName,
				)
				observability.IncResult("duplicate")

				if ackErr := delivery.Ack(false); ackErr != nil {
					done()
					return fmt.Errorf("ack duplicate message: %w", ackErr)
				}

				done()
				continue
			}

			if err := c.handler.HandleTransactionApproved(ctx, event); err != nil {
				retryCount := currentRetryCount(delivery.Headers)
				if retryCount >= c.cfg.MaxRetryCount {
					slog.Warn("worker_process_failed_to_dlq",
						"transaction_id", event.TransactionID,
						"retry_count", retryCount,
						"max_retry_count", c.cfg.MaxRetryCount,
						"error", err,
					)
					observability.IncResult("dlq")

					if dlqErr := c.publishToRoutingKey(ctx, c.cfg.DLQRoutingKey, delivery.Body, withRetryHeader(delivery.Headers, retryCount)); dlqErr != nil {
						done()
						return fmt.Errorf("publish failed message to dlq: %w", dlqErr)
					}
				} else {
					nextRetry := retryCount + 1
					slog.Warn("worker_process_failed_to_retry",
						"transaction_id", event.TransactionID,
						"retry_count", nextRetry,
						"max_retry_count", c.cfg.MaxRetryCount,
						"error", err,
					)
					observability.IncResult("retry")

					if retryErr := c.publishToRoutingKey(ctx, c.cfg.RetryRoutingKey, delivery.Body, withRetryHeader(delivery.Headers, nextRetry)); retryErr != nil {
						done()
						return fmt.Errorf("publish failed message to retry queue: %w", retryErr)
					}
				}

				if ackErr := delivery.Ack(false); ackErr != nil {
					done()
					return fmt.Errorf("ack failed message after reroute: %w", ackErr)
				}

				done()
				continue
			}

			c.processedStore.Mark(event.TransactionID)
			observability.IncResult("success")

			if err := delivery.Ack(false); err != nil {
				done()
				return fmt.Errorf("ack processed message: %w", err)
			}

			slog.Info("worker_message_acked",
				"transaction_id", event.TransactionID,
				"event_name", event.EventName,
				"queue", c.cfg.QueueName,
			)

			done()
		}
	}
}

func (c *TransactionApprovedConsumer) ensureTopology() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel for topology: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		c.cfg.RabbitMQExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange %q: %w", c.cfg.RabbitMQExchange, err)
	}

	_, err = ch.QueueDeclare(
		c.cfg.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue %q: %w", c.cfg.QueueName, err)
	}

	_, err = ch.QueueDeclare(
		c.cfg.RetryQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare retry queue %q: %w", c.cfg.RetryQueueName, err)
	}

	_, err = ch.QueueDeclare(
		c.cfg.DLQName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare dlq %q: %w", c.cfg.DLQName, err)
	}

	if err := ch.QueueBind(
		c.cfg.QueueName,
		c.cfg.RoutingKey,
		c.cfg.RabbitMQExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind queue %q: %w", c.cfg.QueueName, err)
	}

	if err := ch.QueueBind(
		c.cfg.RetryQueueName,
		c.cfg.RetryRoutingKey,
		c.cfg.RabbitMQExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind retry queue %q: %w", c.cfg.RetryQueueName, err)
	}

	if err := ch.QueueBind(
		c.cfg.DLQName,
		c.cfg.DLQRoutingKey,
		c.cfg.RabbitMQExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind dlq %q: %w", c.cfg.DLQName, err)
	}

	return nil
}

func (c *TransactionApprovedConsumer) publishToRoutingKey(ctx context.Context, routingKey string, body []byte, headers amqp.Table) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel for publish: %w", err)
	}
	defer ch.Close()

	return ch.PublishWithContext(
		ctx,
		c.cfg.RabbitMQExchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Headers:      headers,
		},
	)
}

func decodeTransactionApprovedEvent(body []byte) (domain.TransactionApprovedEvent, error) {
	var event domain.TransactionApprovedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return domain.TransactionApprovedEvent{}, fmt.Errorf("decode event json: %w", err)
	}
	if strings.TrimSpace(event.EventName) == "" {
		return domain.TransactionApprovedEvent{}, fmt.Errorf("event_name is required")
	}
	if strings.TrimSpace(event.TransactionID) == "" {
		return domain.TransactionApprovedEvent{}, fmt.Errorf("transaction_id is required")
	}
	return event, nil
}

func currentRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}

	raw, ok := headers["x-retry-count"]
	if !ok {
		return 0
	}

	switch v := raw.(type) {
	case int32:
		return int(v)
	case int64:
		return int(v)
	case int:
		return v
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

func withRetryHeader(headers amqp.Table, retryCount int) amqp.Table {
	out := amqp.Table{}
	for k, v := range headers {
		out[k] = v
	}
	out["x-retry-count"] = retryCount
	return out
}