package usecase

import (
	"context"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/mapper"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
)

// PlanUseCase defines the interface for plan operations
type PlanUseCase interface {
	Create(ctx context.Context, req dto.CreatePlanRequest) (*dto.PlanDTO, error)
	GetByID(ctx context.Context, id string) (*dto.PlanDTO, error)
	List(ctx context.Context, limit, offset int) (*dto.PlanListResponse, error)
	Update(ctx context.Context, id string, req dto.UpdatePlanRequest) error
	Archive(ctx context.Context, id string) error
}

// PlanUseCaseImpl handles plan operations
type PlanUseCaseImpl struct {
	planRepo   ports.PlanRepository
	planMapper *mapper.PlanMapper
}

// NewPlanUseCase creates a new PlanUseCase
func NewPlanUseCase(planRepo ports.PlanRepository) PlanUseCase {
	return &PlanUseCaseImpl{
		planRepo:   planRepo,
		planMapper: mapper.NewPlanMapper(),
	}
}

// Create creates a new plan
func (uc *PlanUseCaseImpl) Create(ctx context.Context, req dto.CreatePlanRequest) (*dto.PlanDTO, error) {
	plan, err := entity.NewPlan(req.Name, req.Description, entity.PlanType(req.Type), req.Price)
	if err != nil {
		return nil, err
	}

	err = uc.planRepo.Create(ctx, plan)
	if err != nil {
		return nil, err
	}

	return uc.planMapper.ToPlanDTO(plan), nil
}

// GetByID gets a plan by ID
func (uc *PlanUseCaseImpl) GetByID(ctx context.Context, id string) (*dto.PlanDTO, error) {
	plan, err := uc.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.planMapper.ToPlanDTO(plan), nil
}

// List lists plans with pagination
func (uc *PlanUseCaseImpl) List(ctx context.Context, limit, offset int) (*dto.PlanListResponse, error) {
	plans, total, err := uc.planRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	response := uc.planMapper.ToPlanListDTO(plans)
	response.Total = total
	return &response, nil
}

// Update updates a plan
func (uc *PlanUseCaseImpl) Update(ctx context.Context, id string, req dto.UpdatePlanRequest) error {
	plan, err := uc.planRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates from request
	if req.Name != nil {
		plan.Name = *req.Name
	}
	if req.Description != nil {
		plan.Description = *req.Description
	}
	if req.Type != nil {
		plan.Type = entity.PlanType(*req.Type)
	}
	if req.Status != nil {
		plan.Status = entity.PlanStatus(*req.Status)
	}
	if req.Price != nil {
		plan.Price = *req.Price
	}
	if req.Currency != nil {
		plan.Currency = *req.Currency
	}
	if req.QuotaType != nil {
		plan.QuotaType = entity.QuotaType(*req.QuotaType)
	}
	if req.MaxNFCePerMonth != nil {
		plan.MaxNFCePerMonth = *req.MaxNFCePerMonth
	}
	if req.MaxNFCeTotal != nil {
		plan.MaxNFCeTotal = *req.MaxNFCeTotal
	}
	if req.Features != nil {
		plan.Features = entity.PlanFeatures{
			MaxNFCePerMonth:    req.Features.MaxNFCePerMonth,
			MaxNFCeTotal:       req.Features.MaxNFCeTotal,
			AllowContingency:   req.Features.AllowContingency,
			AllowCancellation:  req.Features.AllowCancellation,
			AllowInutilization: req.Features.AllowInutilization,
			WebhookSupport:     req.Features.WebhookSupport,
			PrioritySupport:    req.Features.PrioritySupport,
			StorageDays:        req.Features.StorageDays,
		}
	}
	if req.IsPopular != nil {
		plan.IsPopular = *req.IsPopular
	}
	if req.SortOrder != nil {
		plan.SortOrder = *req.SortOrder
	}
	if req.TrialDays != nil {
		plan.TrialDays = *req.TrialDays
	}

	return uc.planRepo.Update(ctx, plan)
}

// Archive archives a plan
func (uc *PlanUseCaseImpl) Archive(ctx context.Context, id string) error {
	plan, err := uc.planRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	plan.Archive()
	return uc.planRepo.Update(ctx, plan)
}
