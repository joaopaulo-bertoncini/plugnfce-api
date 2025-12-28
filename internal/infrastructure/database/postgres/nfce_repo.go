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
	// Set default company ID if not provided (temporary until company management is implemented)
	if req.CompanyID == "" {
		req.CompanyID = "550e8400-e29b-41d4-a716-446655440000" // Default company UUID
	}
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	// Omit associations to prevent GORM from trying to resolve Events relationship
	return r.db.WithContext(ctx).Omit("Events").Create(req).Error
}

// UpdateStatus updates the status of an NFC-e request
func (r *nfceRepository) UpdateStatus(ctx context.Context, id string, from entity.RequestStatus, to entity.RequestStatus, mutate func(*entity.NFCE)) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Optimized: Use direct UPDATE with WHERE clause to avoid SELECT + UPDATE
		result := tx.Model(&entity.NFCE{}).
			Where("id = ? AND status = ?", id, from).
			Update("status", to).
			Update("updated_at", time.Now())

		if result.Error != nil {
			return result.Error
		}

		// If no rows were affected, status was already changed
		if result.RowsAffected == 0 {
			return nil // Status already changed or record not found
		}

		// If we need to mutate other fields, we still need to fetch
		if mutate != nil {
			var req entity.NFCE
			if err := tx.First(&req, "id = ?", id).Error; err != nil {
				return err
			}
			mutate(&req)
			return tx.Save(&req).Error
		}

		return nil
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
	err := r.db.WithContext(ctx).
		Omit("Events"). // Prevent GORM from trying to load Events association
		Where("idempotency_key = ?", key).
		Order("created_at DESC"). // Get the most recent if duplicates (though UNIQUE constraint prevents this)
		First(&req).Error
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

	// Apply filters (order matters for index usage)
	if companyID != "" && status != "" {
		// Use composite index: idx_nfce_requests_company_status_created
		query = query.Where("company_id = ? AND status = ?", companyID, status)
	} else if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	} else if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count efficiently
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with optimized ordering
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

// UpdateFields updates specific fields of an NFC-e request efficiently
func (r *nfceRepository) UpdateFields(ctx context.Context, id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.NFCE{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetStats returns optimized statistics for dashboard
func (r *nfceRepository) GetStats(ctx context.Context, companyID string, since time.Time) (map[string]int, error) {
	var stats struct {
		Pending    int `json:"pending"`
		Processing int `json:"processing"`
		Authorized int `json:"authorized"`
		Rejected   int `json:"rejected"`
		Retrying   int `json:"retrying"`
		Canceled   int `json:"canceled"`
		Total      int `json:"total"`
	}

	query := r.db.WithContext(ctx).Model(&entity.NFCE{}).Where("created_at >= ?", since)

	if companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	// Use raw SQL for better performance on stats queries
	err := query.Select(`
		COUNT(*) FILTER (WHERE status = 'pending') as pending,
		COUNT(*) FILTER (WHERE status = 'processing') as processing,
		COUNT(*) FILTER (WHERE status = 'authorized') as authorized,
		COUNT(*) FILTER (WHERE status = 'rejected') as rejected,
		COUNT(*) FILTER (WHERE status = 'retrying') as retrying,
		COUNT(*) FILTER (WHERE status = 'canceled') as canceled,
		COUNT(*) as total
	`).Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]int{
		"pending":    stats.Pending,
		"processing": stats.Processing,
		"authorized": stats.Authorized,
		"rejected":   stats.Rejected,
		"retrying":   stats.Retrying,
		"canceled":   stats.Canceled,
		"total":      stats.Total,
	}, nil
}

// CreateEvent creates an event for NFC-e tracking (alias for AppendEvent)
func (r *nfceRepository) CreateEvent(ctx context.Context, event *entity.Event) error {
	return r.AppendEvent(ctx, event)
}

// GetPendingRetries gets NFC-e requests that are due for retry
func (r *nfceRepository) GetPendingRetries(ctx context.Context, beforeTime time.Time, limit int) ([]*entity.NFCE, error) {
	var requests []*entity.NFCE
	err := r.db.WithContext(ctx).
		Omit("Events"). // Prevent GORM from trying to load Events association
		Where("status = ? AND next_retry_at IS NOT NULL AND next_retry_at <= ?",
			entity.RequestStatusRetrying, beforeTime).
		Limit(limit).
		Order("next_retry_at ASC"). // Order by next_retry_at for priority (oldest first)
		Find(&requests).Error
	return requests, err
}

// GetEventsByRequestID gets events for a specific NFC-e request
func (r *nfceRepository) GetEventsByRequestID(ctx context.Context, requestID string, limit, offset int) ([]*entity.Event, error) {
	var events []*entity.Event
	err := r.db.WithContext(ctx).Where("request_id = ?", requestID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&events).Error
	return events, err
}
