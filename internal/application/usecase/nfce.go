package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/mapper"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/messaging/rabbitmq"
)

// NFCeUseCase defines the interface for NFC-e business logic
type NFCeUseCase interface {
	EmitNFce(ctx context.Context, idempotencyKey string, req dto.EmitNFceRequest) (*dto.NFceResponse, error)
	GetNFceByID(ctx context.Context, id string) (*dto.NFceResponse, error)
	ListNFces(ctx context.Context, limit, offset int) (*dto.NFceListResponse, error)
	CancelNFce(ctx context.Context, id string, req dto.CancelNFceRequest) error
	GetNFceEvents(ctx context.Context, requestID string, limit, offset int) (*dto.NFceEventListResponse, error)
}

// nfceUseCase implements NFCeUseCase
type nfceUseCase struct {
	repo      ports.NFCeRepository
	publisher rabbitmq.Publisher
	mapper    *mapper.NFceMapper
}

// NewNFCeUseCase creates a new NFCeUseCase
func NewNFCeUseCase(repo ports.NFCeRepository, publisher rabbitmq.Publisher) NFCeUseCase {
	return &nfceUseCase{
		repo:      repo,
		publisher: publisher,
		mapper:    mapper.NewNFceMapper(),
	}
}

// EmitNFce handles the NFC-e emission request
func (uc *nfceUseCase) EmitNFce(ctx context.Context, idempotencyKey string, req dto.EmitNFceRequest) (*dto.NFceResponse, error) {
	// Check for existing request with same idempotency key
	existing, err := uc.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err == nil && existing != nil {
		// Return existing request if already authorized or processing
		if existing.Status == entity.RequestStatusAuthorized ||
			existing.Status == entity.RequestStatusProcessing {
			response := uc.mapper.ToResponse(existing)
			return &response, nil
		}
		// Return error if rejected
		if existing.Status == entity.RequestStatusRejected {
			return nil, fmt.Errorf("NFC-e already rejected: %s", existing.RejectionMsg)
		}
	}

	// Generate new request ID
	requestID := uuid.New().String()

	// Create request entity (this needs to be refactored to use entity constructors)
	// TODO: This is still a violation - should use entity.NewRequest() or similar
	nfceRequest := &entity.Request{
		ID:             requestID,
		IdempotencyKey: idempotencyKey,
		Status:         entity.RequestStatusPending,
		Payload:        uc.mapper.ToEmitPayload(req),
	}

	// Persist request
	if err := uc.repo.Create(ctx, nfceRequest); err != nil {
		return nil, fmt.Errorf("failed to create NFC-e request: %w", err)
	}

	// Publish to queue for async processing
	emitMsg := rabbitmq.EmitMessage{
		RequestID:      requestID,
		IdempotencyKey: idempotencyKey,
		Payload:        uc.mapper.ToEmitPayload(req),
		EnqueuedAt:     nfceRequest.CreatedAt,
	}

	if err := uc.publisher.PublishEmit(ctx, emitMsg); err != nil {
		// Log error but don't fail the request - it will be retried
		// TODO: Add proper logging
		_ = err
	}

	response := uc.mapper.ToResponse(nfceRequest)
	return &response, nil
}

// GetNFceByID retrieves a NFC-e by ID
func (uc *nfceUseCase) GetNFceByID(ctx context.Context, id string) (*dto.NFceResponse, error) {
	req, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFC-e: %w", err)
	}

	response := uc.mapper.ToResponse(req)

	// Add links if authorized
	if dto.RequestStatus(req.Status) == dto.RequestStatusAuthorized && req.ChaveAcesso != "" {
		response.Links = dto.NFceLinks{
			XML:    fmt.Sprintf("/nfce/%s/xml", id),
			PDF:    fmt.Sprintf("/nfce/%s/pdf", id),
			QrCode: fmt.Sprintf("/nfce/%s/qrcode", id),
		}
	}

	return &response, nil
}

// ListNFces lists NFC-e requests with pagination
func (uc *nfceUseCase) ListNFces(ctx context.Context, limit, offset int) (*dto.NFceListResponse, error) {
	requests, err := uc.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list NFC-es: %w", err)
	}

	response := uc.mapper.ToResponseList(requests)
	return &response, nil
}

// CancelNFce cancels a NFC-e
func (uc *nfceUseCase) CancelNFce(ctx context.Context, id string, req dto.CancelNFceRequest) error {
	// Get current request
	nfceReq, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get NFC-e: %w", err)
	}

	// Check if can be canceled
	if dto.RequestStatus(nfceReq.Status) != dto.RequestStatusAuthorized {
		return errors.New("only authorized NFC-e can be canceled")
	}

	// Update status to canceled
	err = uc.repo.UpdateStatus(ctx, id, entity.RequestStatusAuthorized, entity.RequestStatusCanceled, func(r *entity.Request) {
		// Add cancellation metadata if needed
	})
	if err != nil {
		return fmt.Errorf("failed to cancel NFC-e: %w", err)
	}

	// TODO: Publish cancellation event to queue

	return nil
}

// GetNFceEvents retrieves events for a NFC-e request
func (uc *nfceUseCase) GetNFceEvents(ctx context.Context, requestID string, limit, offset int) (*dto.NFceEventListResponse, error) {
	events, err := uc.repo.GetEventsByRequestID(ctx, requestID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFC-e events: %w", err)
	}

	response := uc.mapper.ToEventResponseList(events)
	return &response, nil
}
