package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	amqp "github.com/rabbitmq/amqp091-go"
)

// consumer implements Consumer interface
type consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewConsumer creates a new RabbitMQ consumer
func NewConsumer(url string) (dto.Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		"nfce.exchange", // name
		"direct",        // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	queue, err := channel.QueueDeclare(
		"nfce.emit", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		queue.Name,      // queue name
		"nfce.emit",     // routing key
		"nfce.exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &consumer{
		conn:    conn,
		channel: channel,
	}, nil
}

// ConsumeEmit consumes NFC-e emission messages
func (c *consumer) ConsumeEmit(ctx context.Context, handler func(context.Context, dto.EmitMessage) error) error {
	msgs, err := c.channel.Consume(
		"nfce.emit", // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}

			// Parse message
			var msg dto.EmitMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				d.Nack(false, false) // Don't requeue invalid messages
				continue
			}

			// Handle message
			if err := handler(ctx, msg); err != nil {
				log.Printf("Handler error for message %s: %v", msg.RequestID, err)
				// Check if it's a retryable error
				if shouldRetry(err) {
					d.Nack(false, true) // Requeue
				} else {
					d.Nack(false, false) // Don't requeue
				}
				continue
			}

			// Acknowledge successful processing
			if err := d.Ack(false); err != nil {
				log.Printf("Failed to acknowledge message %s: %v", msg.RequestID, err)
			}
		}
	}
}

// shouldRetry determines if an error should trigger message requeue
func shouldRetry(err error) bool {
	// For now, retry all errors. In production, you might want to classify errors
	// as retryable (temporary failures) vs non-retryable (permanent failures)
	return true
}

// Close closes the consumer connections
func (c *consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
