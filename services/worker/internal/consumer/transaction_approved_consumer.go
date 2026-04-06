package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"finpharm-ai/services/worker/internal/config"
	"finpharm-ai/services/worker/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TransactionApprovedHandler interface {
	HandleTransactionApproved(ctx context.Context, event domain.TransactionApprovedEvent) error
}

type TransactionApprovedConsumer struct {
	cfg     config.Config
	conn    *amqp.Connection
	handler TransactionApprovedHandler
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
		cfg:     cfg,
		conn:    conn,
		handler: handler,
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

			event, err := decodeTransactionApprovedEvent(delivery.Body)
			if err != nil {
				slog.Warn("worker_invalid_message_discarded",
					"queue", c.cfg.QueueName,
					"error", err,
				)
				if ackErr := delivery.Ack(false); ackErr != nil {
					return fmt.Errorf("ack invalid message: %w", ackErr)
				}
				continue
			}

			if err := c.handler.HandleTransactionApproved(ctx, event); err != nil {
				slog.Warn("worker_process_failed_requeue",
					"transaction_id", event.TransactionID,
					"error", err,
				)
				if nackErr := delivery.Nack(false, true); nackErr != nil {
					return fmt.Errorf("nack failed message: %w", nackErr)
				}
				continue
			}

			if err := delivery.Ack(false); err != nil {
				return fmt.Errorf("ack processed message: %w", err)
			}

			slog.Info("worker_message_acked",
				"transaction_id", event.TransactionID,
				"event_name", event.EventName,
				"queue", c.cfg.QueueName,
			)
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

	if err := ch.QueueBind(
		c.cfg.QueueName,
		c.cfg.RoutingKey,
		c.cfg.RabbitMQExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind queue %q to exchange %q with routing key %q: %w",
			c.cfg.QueueName,
			c.cfg.RabbitMQExchange,
			c.cfg.RoutingKey,
			err,
		)
	}

	return nil
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