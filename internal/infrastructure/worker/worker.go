package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/service"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/logger"
)

// Worker processes NFC-e emission requests from the message queue
type Worker struct {
	repo          ports.NFCeRepository
	publisher     dto.Publisher
	consumer      dto.Consumer
	workerService *service.NFCeWorkerService
	logger        logger.Logger
	maxRetries    int
	shutdown      chan struct{}
	wg            sync.WaitGroup
}

// NewWorker creates a new NFC-e worker
func NewWorker(
	repo ports.NFCeRepository,
	publisher dto.Publisher,
	consumer dto.Consumer,
	workerService *service.NFCeWorkerService,
	logger logger.Logger,
	maxRetries int,
) *Worker {
	return &Worker{
		repo:          repo,
		publisher:     publisher,
		consumer:      consumer,
		workerService: workerService,
		logger:        logger,
		maxRetries:    maxRetries,
		shutdown:      make(chan struct{}),
	}
}

// Start begins processing NFC-e emission requests
func (w *Worker) Start(ctx context.Context) error {
	w.logger.Info("Starting NFC-e worker")

	// Start message consumer with handler
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		err := w.consumer.ConsumeEmit(ctx, w.handleMessage)
		if err != nil && err.Error() != "context canceled" {
			w.logger.Error("Consumer error", logger.Field{Key: "error", Value: err.Error()})
		}
	}()

	// Start retry scheduler
	w.wg.Add(1)
	go w.scheduleRetries(ctx)

	w.logger.Info("NFC-e worker started successfully")
	return nil
}

// Stop gracefully shuts down the worker
func (w *Worker) Stop(ctx context.Context) error {
	w.logger.Info("Stopping NFC-e worker")

	// Signal shutdown
	close(w.shutdown)

	// Wait for all goroutines to finish or context timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("NFC-e worker stopped gracefully")
		return nil
	case <-ctx.Done():
		w.logger.Warn("NFC-e worker shutdown timed out")
		return ctx.Err()
	}
}

// handleMessage processes a single message from the queue
func (w *Worker) handleMessage(ctx context.Context, msg dto.EmitMessage) error {
	w.logger.Info("Processing NFC-e emission request",
		logger.Field{Key: "request_id", Value: msg.RequestID},
		logger.Field{Key: "idempotency_key", Value: msg.IdempotencyKey})

	// Get the NFC-e request from database
	nfceRequest, err := w.repo.GetByID(ctx, msg.RequestID)
	if err != nil {
		return fmt.Errorf("failed to get NFC-e request: %w", err)
	}

	// Check idempotency - if already processed successfully, skip
	if nfceRequest.Status == entity.RequestStatusAuthorized {
		w.logger.Info("NFC-e already authorized, skipping")
		return nil
	}

	// Process the NFC-e emission
	if err := w.workerService.ProcessNFceEmission(ctx, nfceRequest); err != nil {
		w.logger.Error("NFC-e emission failed", logger.Field{Key: "error", Value: err.Error()})

		// Check if we can retry
		if w.workerService.CanRetry(nfceRequest, w.maxRetries) {
			w.scheduleRetry(ctx, nfceRequest)
		} else {
			// Mark as rejected if max retries exceeded
			nfceRequest.MarkAsRejected("999", "Número máximo de tentativas excedido")
		}
	}

	// Update the request in database
	if err := w.repo.Update(ctx, nfceRequest); err != nil {
		return fmt.Errorf("failed to update NFC-e request: %w", err)
	}

	// Create event for tracking
	event := &entity.Event{
		ID:         fmt.Sprintf("%s-%d", nfceRequest.ID, time.Now().Unix()),
		RequestID:  nfceRequest.ID,
		StatusFrom: entity.RequestStatusProcessing,
		StatusTo:   nfceRequest.Status,
		CStat:      nfceRequest.CStat,
		Message:    nfceRequest.XMotivo,
		CreatedAt:  time.Now(),
	}

	if err := w.repo.CreateEvent(ctx, event); err != nil {
		w.logger.Error("Failed to create event", logger.Field{Key: "error", Value: err.Error()})
	}

	w.logger.Info("NFC-e emission completed",
		logger.Field{Key: "status", Value: string(nfceRequest.Status)})

	return nil
}

// processMessage processes a single NFC-e emission message
func (w *Worker) processMessage(ctx context.Context, msg dto.EmitMessage, log logger.Logger) error {
	w.logger.Info("Processing NFC-e emission request",
		logger.Field{Key: "request_id", Value: msg.RequestID},
		logger.Field{Key: "idempotency_key", Value: msg.IdempotencyKey})

	// Get the NFC-e request from database
	nfceRequest, err := w.repo.GetByID(ctx, msg.RequestID)
	if err != nil {
		return fmt.Errorf("failed to get NFC-e request: %w", err)
	}

	// Check idempotency - if already processed successfully, skip
	if nfceRequest.Status == entity.RequestStatusAuthorized {
		w.logger.Info("NFC-e already authorized, skipping")
		return nil
	}

	// Process the NFC-e emission
	if err := w.workerService.ProcessNFceEmission(ctx, nfceRequest); err != nil {
		w.logger.Error("NFC-e emission failed", logger.Field{Key: "error", Value: err.Error()})

		// Check if we can retry
		if w.workerService.CanRetry(nfceRequest, w.maxRetries) {
			w.scheduleRetry(ctx, nfceRequest)
			return nil
		}

		// Mark as rejected if max retries exceeded
		nfceRequest.MarkAsRejected("999", "Número máximo de tentativas excedido")
	}

	// Update the request in database
	if err := w.repo.Update(ctx, nfceRequest); err != nil {
		return fmt.Errorf("failed to update NFC-e request: %w", err)
	}

	// Create event for tracking
	event := &entity.Event{
		ID:         fmt.Sprintf("%s-%d", nfceRequest.ID, time.Now().Unix()),
		RequestID:  nfceRequest.ID,
		StatusFrom: entity.RequestStatusProcessing,
		StatusTo:   nfceRequest.Status,
		CStat:      nfceRequest.CStat,
		Message:    nfceRequest.XMotivo,
		CreatedAt:  time.Now(),
	}

	if err := w.repo.CreateEvent(ctx, event); err != nil {
		w.logger.Error("Failed to create event", logger.Field{Key: "error", Value: err.Error()})
	}

	w.logger.Info("NFC-e emission completed",
		logger.Field{Key: "status", Value: string(nfceRequest.Status)})

	return nil
}

// scheduleRetry schedules a retry for the NFC-e request
func (w *Worker) scheduleRetry(ctx context.Context, nfceRequest *entity.NFCE) {
	w.workerService.IncrementRetry(nfceRequest)

	// Calculate backoff delay (exponential backoff)
	delay := w.calculateBackoffDelay(nfceRequest.RetryCount)
	nextRetryAt := time.Now().Add(delay)

	nfceRequest.NextRetryAt = &nextRetryAt
	nfceRequest.Status = entity.RequestStatusRetrying

	w.logger.Info("Scheduled retry",
		logger.Field{Key: "request_id", Value: nfceRequest.ID},
		logger.Field{Key: "retry_count", Value: nfceRequest.RetryCount},
		logger.Field{Key: "next_retry_at", Value: nextRetryAt})
}

// calculateBackoffDelay calculates exponential backoff delay
func (w *Worker) calculateBackoffDelay(retryCount int) time.Duration {
	// Base delays: 1m, 5m, 15m, 1h, 6h, 24h
	baseDelays := []time.Duration{
		time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		time.Hour,
		6 * time.Hour,
		24 * time.Hour,
	}

	if retryCount <= len(baseDelays) {
		return baseDelays[retryCount-1]
	}

	// Max delay of 24 hours for retries beyond the base schedule
	return 24 * time.Hour
}

// scheduleRetries periodically checks for and processes retry requests
func (w *Worker) scheduleRetries(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-w.shutdown:
			return
		case <-ticker.C:
			if err := w.processPendingRetries(ctx); err != nil {
				w.logger.Error("Failed to process pending retries", logger.Field{Key: "error", Value: err.Error()})
			}
		}
	}
}

// processPendingRetries finds and processes NFC-e requests that are due for retry
func (w *Worker) processPendingRetries(ctx context.Context) error {
	// Get requests that are due for retry
	requests, err := w.repo.GetPendingRetries(ctx, time.Now(), 10) // Process up to 10 at a time
	if err != nil {
		return fmt.Errorf("failed to get pending retries: %w", err)
	}

	for _, req := range requests {
		// Reset status to processing and clear next retry time
		req.Status = entity.RequestStatusProcessing
		req.NextRetryAt = nil

		// Update in database
		if err := w.repo.Update(ctx, req); err != nil {
			w.logger.Error("Failed to update retry request",
				logger.Field{Key: "request_id", Value: req.ID},
				logger.Field{Key: "error", Value: err.Error()})
			continue
		}

		// Publish to queue for immediate processing
		emitMsg := dto.EmitMessage{
			RequestID:      req.ID,
			IdempotencyKey: req.IdempotencyKey,
			Payload:        req.Payload,
			EnqueuedAt:     time.Now(),
		}

		if err := w.publisher.PublishEmit(ctx, emitMsg); err != nil {
			w.logger.Error("Failed to publish retry message",
				logger.Field{Key: "request_id", Value: req.ID},
				logger.Field{Key: "error", Value: err.Error()})
		}
	}

	return nil
}
