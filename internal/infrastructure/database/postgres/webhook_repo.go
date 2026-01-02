package postgres

import (
	"context"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
	"gorm.io/gorm"
)

// Webhook repository implementation
type webhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) ports.WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) Create(ctx context.Context, webhook *entity.Webhook) error {
	return r.db.WithContext(ctx).Create(webhook).Error
}

func (r *webhookRepository) GetByID(ctx context.Context, id string) (*entity.Webhook, error) {
	var webhook entity.Webhook
	err := r.db.WithContext(ctx).First(&webhook, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (r *webhookRepository) Update(ctx context.Context, webhook *entity.Webhook) error {
	return r.db.WithContext(ctx).Save(webhook).Error
}

func (r *webhookRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.Webhook{}, "id = ?", id).Error
}

func (r *webhookRepository) List(ctx context.Context, limit, offset int) ([]*entity.Webhook, int, error) {
	var webhooks []*entity.Webhook
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Webhook{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&webhooks).Error
	return webhooks, int(total), err
}

func (r *webhookRepository) ListByCompanyID(ctx context.Context, companyID string, limit, offset int) ([]*entity.Webhook, int, error) {
	var webhooks []*entity.Webhook
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Webhook{}).Where("company_id = ?", companyID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&webhooks).Error
	return webhooks, int(total), err
}

func (r *webhookRepository) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Webhook{}).Count(&count).Error
	return int(count), err
}
