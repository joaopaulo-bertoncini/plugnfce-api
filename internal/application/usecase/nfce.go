package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/mapper"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/storage"
)

// NFCeUseCase defines the interface for NFC-e business logic
type NFCeUseCase interface {
	EmitNFce(ctx context.Context, idempotencyKey string, req dto.EmitNFceRequest) (*dto.NFceResponse, error)
	GetNFceByID(ctx context.Context, id string) (*dto.NFceResponse, error)
	ListNFces(ctx context.Context, limit, offset int) (*dto.NFceListResponse, error)
	CancelNFce(ctx context.Context, id string, req dto.CancelNFceRequest) error
	GetNFceEvents(ctx context.Context, requestID string, limit, offset int) (*dto.NFceEventListResponse, error)
	DownloadXML(ctx context.Context, id string) ([]byte, error)
	DownloadPDF(ctx context.Context, id string) ([]byte, error)
	DownloadQRCode(ctx context.Context, id string) ([]byte, error)
}

// nfceUseCase implements NFCeUseCase
type nfceUseCase struct {
	repo      ports.NFCeRepository
	publisher dto.Publisher
	mapper    *mapper.NFceMapper
	storage   storage.StorageService
}

// NewNFCeUseCase creates a new NFCeUseCase
func NewNFCeUseCase(repo ports.NFCeRepository, publisher dto.Publisher, storage storage.StorageService) NFCeUseCase {
	return &nfceUseCase{
		repo:      repo,
		publisher: publisher,
		mapper:    mapper.NewNFceMapper(),
		storage:   storage,
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

	// Create request entity (this needs to be refactored to use entity constructors)
	// TODO: This is still a violation - should use entity.NewRequest() or similar
	nfceRequest := &entity.Request{
		IdempotencyKey: idempotencyKey,
		Status:         entity.RequestStatusPending,
		Payload:        uc.mapper.ToEmitPayload(req),
	}
	fmt.Printf("DEBUG: Created nfceRequest with initial ID: %s\n", nfceRequest.ID)

	// Persist request
	if err := uc.repo.Create(ctx, nfceRequest); err != nil {
		return nil, fmt.Errorf("failed to create NFC-e request: %w", err)
	}
	fmt.Printf("DEBUG: Persisted nfceRequest with final ID: %s\n", nfceRequest.ID)

	// Use the ID assigned by database
	requestID := nfceRequest.ID

	// Publish to queue for async processing
	emitMsg := dto.EmitMessage{
		RequestID:      requestID,
		IdempotencyKey: idempotencyKey,
		EnqueuedAt:     nfceRequest.CreatedAt,
	}
	fmt.Printf("DEBUG: Created emitMsg with RequestID: %s\n", emitMsg.RequestID)

	if err := uc.publisher.PublishEmit(ctx, emitMsg); err != nil {
		// Log error but don't fail the request - it will be retried
		// TODO: Add proper logging
		fmt.Printf("Failed to publish NFC-e message: %v\n", err)
		return nil, fmt.Errorf("failed to publish NFC-e message: %w", err)
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

	// Update status to canceled (temporarily mark as processing for queue)
	err = uc.repo.UpdateStatus(ctx, id, entity.RequestStatusAuthorized, entity.RequestStatusProcessing, func(r *entity.Request) {
		// Add cancellation metadata if needed
		r.XMotivo = req.Justificativa
	})
	if err != nil {
		return fmt.Errorf("failed to update NFC-e status: %w", err)
	}

	// Publish cancellation event to queue
	cancelMsg := dto.CancelMessage{
		RequestID:      id,
		IdempotencyKey: nfceReq.IdempotencyKey,
		Justificativa:  req.Justificativa,
		EnqueuedAt:     time.Now(),
	}

	if err := uc.publisher.PublishCancel(ctx, cancelMsg); err != nil {
		// Revert status if publish fails
		uc.repo.UpdateStatus(ctx, id, entity.RequestStatusProcessing, entity.RequestStatusAuthorized, func(r *entity.Request) {})
		return fmt.Errorf("failed to publish cancellation event: %w", err)
	}

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

// DownloadXML downloads the XML file for an NFC-e
func (uc *nfceUseCase) DownloadXML(ctx context.Context, id string) ([]byte, error) {
	// Get NFC-e request
	nfce, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFC-e: %w", err)
	}

	// Check if authorized
	if nfce.Status != entity.RequestStatusAuthorized && nfce.Status != entity.RequestStatusContingency {
		return nil, errors.New("NFC-e is not authorized")
	}

	// Check if XML URL exists
	if nfce.XMLURL == "" {
		return nil, errors.New("XML file not found")
	}

	// Extract bucket and key from URL (assuming format: http://minio:9000/bucket/key)
	// For now, construct the key from the known pattern
	key := fmt.Sprintf("nfce/%s/xml/%s.xml", nfce.CompanyID, nfce.ChaveAcesso)

	// Download file
	data, err := uc.storage.DownloadFile(ctx, "", key)
	if err != nil {
		return nil, fmt.Errorf("failed to download XML file: %w", err)
	}

	return data, nil
}

// DownloadPDF downloads the PDF file for an NFC-e
func (uc *nfceUseCase) DownloadPDF(ctx context.Context, id string) ([]byte, error) {
	// Get NFC-e request
	nfce, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFC-e: %w", err)
	}

	// Check if authorized
	if nfce.Status != entity.RequestStatusAuthorized && nfce.Status != entity.RequestStatusContingency {
		return nil, errors.New("NFC-e is not authorized")
	}

	// Check if PDF URL exists
	if nfce.PDFURL == "" {
		return nil, errors.New("PDF file not found")
	}

	// Construct the key from the known pattern
	key := fmt.Sprintf("nfce/%s/pdf/%s.pdf", nfce.CompanyID, nfce.ChaveAcesso)

	// Download file
	data, err := uc.storage.DownloadFile(ctx, "", key)
	if err != nil {
		return nil, fmt.Errorf("failed to download PDF file: %w", err)
	}

	return data, nil
}

// DownloadQRCode downloads the QR Code image for an NFC-e
func (uc *nfceUseCase) DownloadQRCode(ctx context.Context, id string) ([]byte, error) {
	// Get NFC-e request
	nfce, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFC-e: %w", err)
	}

	// Check if authorized
	if nfce.Status != entity.RequestStatusAuthorized && nfce.Status != entity.RequestStatusContingency {
		return nil, errors.New("NFC-e is not authorized")
	}

	// Check if QR Code URL exists
	if nfce.QRCodeURL == "" {
		return nil, errors.New("QR Code file not found")
	}

	// Construct the key from the known pattern
	key := fmt.Sprintf("nfce/%s/qr/%s.png", nfce.CompanyID, nfce.ChaveAcesso)

	// Download file
	data, err := uc.storage.DownloadFile(ctx, "", key)
	if err != nil {
		return nil, fmt.Errorf("failed to download QR Code file: %w", err)
	}

	return data, nil
}
