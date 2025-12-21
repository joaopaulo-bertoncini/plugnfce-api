package postgres

import (
	"context"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
	"gorm.io/gorm"
)

// Subscription repository implementation
type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) ports.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, subscription *entity.Subscription) error {
	return r.db.WithContext(ctx).Create(subscription).Error
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id string) (*entity.Subscription, error) {
	var subscription entity.Subscription
	err := r.db.WithContext(ctx).First(&subscription, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepository) GetActiveByCompanyID(ctx context.Context, companyID string) (*entity.Subscription, error) {
	var subscription entity.Subscription
	err := r.db.WithContext(ctx).Where("company_id = ? AND status IN ('active', 'trial')", companyID).Order("created_at DESC").First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, subscription *entity.Subscription) error {
	return r.db.WithContext(ctx).Save(subscription).Error
}

func (r *subscriptionRepository) List(ctx context.Context, limit, offset int) ([]*entity.Subscription, int, error) {
	var subscriptions []*entity.Subscription
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Subscription{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&subscriptions).Error
	return subscriptions, int(total), err
}

func (r *subscriptionRepository) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Subscription{}).Count(&count).Error
	return int(count), err
}

func (r *subscriptionRepository) CountByStatus(ctx context.Context, status entity.SubscriptionStatus) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Subscription{}).Where("status = ?", status).Count(&count).Error
	return int(count), err
}
