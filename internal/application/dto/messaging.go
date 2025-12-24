package dto

import (
	"context"
	"time"
)

// EmitMessage is the payload published to the queue for NFC-e emission.
// Contains only the request ID for efficiency - worker fetches full data from database.
type EmitMessage struct {
	RequestID      string    `json:"request_id"`
	IdempotencyKey string    `json:"idempotency_key"`
	RetryCount     int       `json:"retry_count,omitempty"`
	EnqueuedAt     time.Time `json:"enqueued_at"`
}

// Publisher abstracts the message bus used by the API.
type Publisher interface {
	PublishEmit(ctx context.Context, msg EmitMessage) error
}

// Consumer abstracts the worker subscription to the emission queue.
type Consumer interface {
	ConsumeEmit(ctx context.Context, handler func(context.Context, EmitMessage) error) error
}
