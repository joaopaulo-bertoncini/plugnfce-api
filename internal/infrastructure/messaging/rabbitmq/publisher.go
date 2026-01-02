package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	amqp "github.com/rabbitmq/amqp091-go"
)

// publisher implements Publisher interface
type publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher(url string) (dto.Publisher, error) {
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

	// Declare emit queue
	_, err = channel.QueueDeclare(
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
		return nil, fmt.Errorf("failed to declare emit queue: %w", err)
	}

	// Bind emit queue to exchange
	err = channel.QueueBind(
		"nfce.emit",     // queue name
		"nfce.emit",     // routing key
		"nfce.exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind emit queue: %w", err)
	}

	// Declare cancel queue
	_, err = channel.QueueDeclare(
		"nfce.cancel", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare cancel queue: %w", err)
	}

	// Bind cancel queue to exchange
	err = channel.QueueBind(
		"nfce.cancel",   // queue name
		"nfce.cancel",   // routing key
		"nfce.exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind cancel queue: %w", err)
	}

	return &publisher{
		conn:    conn,
		channel: channel,
	}, nil
}

// PublishEmit publishes an NFC-e emission message
func (p *publisher) PublishEmit(ctx context.Context, msg dto.EmitMessage) error {
	fmt.Printf("DEBUG: Publishing message to nfce.exchange with routing key nfce.emit: %+v\n", msg)

	body, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("DEBUG: Failed to marshal message: %v\n", err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	fmt.Printf("DEBUG: Publishing to exchange nfce.exchange, routing key nfce.emit\n")

	err = p.channel.PublishWithContext(ctx,
		"nfce.exchange", // exchange
		"nfce.emit",     // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		fmt.Printf("DEBUG: Failed to publish message: %v\n", err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	fmt.Printf("DEBUG: Message published successfully\n")
	return nil
}

// PublishCancel publishes an NFC-e cancellation message
func (p *publisher) PublishCancel(ctx context.Context, msg dto.CancelMessage) error {
	fmt.Printf("DEBUG: Publishing cancel message to nfce.exchange with routing key nfce.cancel: %+v\n", msg)

	body, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("DEBUG: Failed to marshal cancel message: %v\n", err)
		return fmt.Errorf("failed to marshal cancel message: %w", err)
	}

	fmt.Printf("DEBUG: Publishing cancel to exchange nfce.exchange, routing key nfce.cancel\n")

	err = p.channel.PublishWithContext(ctx,
		"nfce.exchange", // exchange
		"nfce.cancel",   // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		fmt.Printf("DEBUG: Failed to publish cancel message: %v\n", err)
		return fmt.Errorf("failed to publish cancel message: %w", err)
	}

	fmt.Printf("DEBUG: Cancel message published successfully\n")
	return nil
}

// Close closes the publisher connections
func (p *publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
