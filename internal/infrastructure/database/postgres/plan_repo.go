package postgres

import (
	"context"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
	"gorm.io/gorm"
)

// Plan repository implementation
type planRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) ports.PlanRepository {
	return &planRepository{db: db}
}

func (r *planRepository) Create(ctx context.Context, plan *entity.Plan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *planRepository) GetByID(ctx context.Context, id string) (*entity.Plan, error) {
	var plan entity.Plan
	err := r.db.WithContext(ctx).First(&plan, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) Update(ctx context.Context, plan *entity.Plan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *planRepository) List(ctx context.Context, limit, offset int) ([]*entity.Plan, int, error) {
	var plans []*entity.Plan
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Plan{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&plans).Error
	return plans, int(total), err
}

func (r *planRepository) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Plan{}).Count(&count).Error
	return int(count), err
}
