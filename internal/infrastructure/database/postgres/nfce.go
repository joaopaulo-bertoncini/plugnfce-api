package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
	"gorm.io/gorm"
)

// nfceRepository implements ports.NFCeRepository
type nfceRepository struct {
	db *gorm.DB
}

// NewNFCeRepository creates a new NFC-e repository
func NewNFCeRepository(db *gorm.DB) ports.NFCeRepository {
	return &nfceRepository{db: db}
}

// Create creates a new NFC-e request
func (r *nfceRepository) Create(ctx context.Context, req *entity.NFCE) error {
	req.ID = uuid.New().String()
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(req).Error
}

// UpdateStatus updates the status of an NFC-e request
func (r *nfceRepository) UpdateStatus(ctx context.Context, id string, from entity.RequestStatus, to entity.RequestStatus, mutate func(*entity.NFCE)) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var req entity.NFCE
		if err := tx.First(&req, "id = ?", id).Error; err != nil {
			return err
		}

		if req.Status != from {
			return nil // Status already changed
		}

		if mutate != nil {
			mutate(&req)
		}

		req.Status = to
		req.UpdatedAt = time.Now()

		return tx.Save(&req).Error
	})
}

// GetByID gets an NFC-e request by ID
func (r *nfceRepository) GetByID(ctx context.Context, id string) (*entity.NFCE, error) {
	var req entity.NFCE
	err := r.db.WithContext(ctx).First(&req, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// GetByIdempotencyKey gets an NFC-e request by idempotency key
func (r *nfceRepository) GetByIdempotencyKey(ctx context.Context, key string) (*entity.NFCE, error) {
	var req entity.NFCE
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", key).First(&req).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// List lists NFC-e requests with pagination
func (r *nfceRepository) List(ctx context.Context, limit, offset int) ([]*entity.NFCE, error) {
	var requests []*entity.NFCE
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("created_at DESC").Find(&requests).Error
	return requests, err
}

// AppendEvent appends an event to the NFC-e request
func (r *nfceRepository) AppendEvent(ctx context.Context, evt *entity.Event) error {
	evt.ID = uuid.New().String()
	evt.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(evt).Error
}

// ListWithFilters lists NFC-e requests with filters and pagination
func (r *nfceRepository) ListWithFilters(ctx context.Context, limit, offset int, companyID, status string) ([]*entity.NFCE, int, error) {
	var requests []*entity.NFCE
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.NFCE{})

	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&requests).Error
	return requests, int(total), err
}

// Count counts total NFC-e requests
func (r *nfceRepository) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.NFCE{}).Count(&count).Error
	return int(count), err
}

// CountByStatus counts NFC-e requests by status
func (r *nfceRepository) CountByStatus(ctx context.Context, status entity.RequestStatus) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.NFCE{}).Where("status = ?", status).Count(&count).Error
	return int(count), err
}

// Update updates an NFC-e request
func (r *nfceRepository) Update(ctx context.Context, nfce *entity.NFCE) error {
	nfce.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(nfce).Error
}

// CreateEvent creates an event for NFC-e tracking (alias for AppendEvent)
func (r *nfceRepository) CreateEvent(ctx context.Context, event *entity.Event) error {
	return r.AppendEvent(ctx, event)
}

// GetPendingRetries gets NFC-e requests that are due for retry
func (r *nfceRepository) GetPendingRetries(ctx context.Context, beforeTime time.Time, limit int) ([]*entity.NFCE, error) {
	var requests []*entity.NFCE
	err := r.db.WithContext(ctx).
		Where("status = ? AND next_retry_at IS NOT NULL AND next_retry_at <= ?",
			entity.RequestStatusRetrying, beforeTime).
		Limit(limit).
		Order("next_retry_at ASC").
		Find(&requests).Error
	return requests, err
}

// GetEventsByRequestID gets events for a specific NFC-e request
func (r *nfceRepository) GetEventsByRequestID(ctx context.Context, requestID string, limit, offset int) ([]*entity.Event, error) {
	var events []*entity.Event
	err := r.db.WithContext(ctx).Where("request_id = ?", requestID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&events).Error
	return events, err
}
