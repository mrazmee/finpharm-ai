package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"finpharm-ai/services/transaction/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQTransactionEventPublisher struct {
	conn                 *amqp.Connection
	exchange             string
	transactionQueue     string
	transactionRoutingKey string
}

func NewRabbitMQTransactionEventPublisher(
	amqpURL string,
	exchange string,
	transactionQueue string,
	transactionRoutingKey string,
) (*RabbitMQTransactionEventPublisher, error) {
	amqpURL = strings.TrimSpace(amqpURL)
	if amqpURL == "" {
		return nil, fmt.Errorf("amqp url is required")
	}

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	publisher := &RabbitMQTransactionEventPublisher{
		conn:                  conn,
		exchange:              strings.TrimSpace(exchange),
		transactionQueue:      strings.TrimSpace(transactionQueue),
		transactionRoutingKey: strings.TrimSpace(transactionRoutingKey),
	}

	if publisher.exchange == "" {
		publisher.exchange = "finpharm.events"
	}
	if publisher.transactionQueue == "" {
		publisher.transactionQueue = "transaction.approved.queue"
	}
	if publisher.transactionRoutingKey == "" {
		publisher.transactionRoutingKey = "transaction.approved"
	}

	if err := publisher.ensureTopology(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return publisher, nil
}

func (p *RabbitMQTransactionEventPublisher) Close() error {
	if p == nil || p.conn == nil {
		return nil
	}
	return p.conn.Close()
}

func (p *RabbitMQTransactionEventPublisher) PublishTransactionApproved(ctx context.Context, event domain.TransactionApprovedEvent) error {
	if p == nil || p.conn == nil {
		return fmt.Errorf("rabbitmq publisher is not initialized")
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal transaction approved event: %w", err)
	}

	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()

	return ch.PublishWithContext(
		ctx,
		p.exchange,
		p.transactionRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Type:         event.EventName,
			Body:         body,
			Timestamp:    event.PublishedAt,
		},
	)
}

func (p *RabbitMQTransactionEventPublisher) ensureTopology() error {
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel for topology: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		p.exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange %q: %w", p.exchange, err)
	}

	_, err = ch.QueueDeclare(
		p.transactionQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue %q: %w", p.transactionQueue, err)
	}

	if err := ch.QueueBind(
		p.transactionQueue,
		p.transactionRoutingKey,
		p.exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind queue %q to exchange %q with routing key %q: %w",
			p.transactionQueue,
			p.exchange,
			p.transactionRoutingKey,
			err,
		)
	}

	return nil
}