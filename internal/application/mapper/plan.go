package mapper

import (
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
)

// PlanMapper handles mapping between plan entities and DTOs
type PlanMapper struct{}

// NewPlanMapper creates a new PlanMapper
func NewPlanMapper() *PlanMapper {
	return &PlanMapper{}
}

// ToPlanDTO converts a Plan entity to a PlanDTO
func (m *PlanMapper) ToPlanDTO(plan *entity.Plan) *dto.PlanDTO {
	return &dto.PlanDTO{
		ID:              plan.ID,
		Name:            plan.Name,
		Description:     plan.Description,
		Type:            dto.PlanType(plan.Type),
		BillingCycle:    dto.BillingCycle(plan.BillingCycle),
		Status:          dto.PlanStatus(plan.Status),
		Price:           plan.Price,
		Currency:        plan.Currency,
		QuotaType:       dto.QuotaType(plan.QuotaType),
		MaxNFCePerMonth: plan.MaxNFCePerMonth,
		MaxNFCeTotal:    plan.MaxNFCeTotal,
		Features: dto.PlanFeatures{
			MaxNFCePerMonth:    plan.Features.MaxNFCePerMonth,
			MaxNFCeTotal:       plan.Features.MaxNFCeTotal,
			AllowContingency:   plan.Features.AllowContingency,
			AllowCancellation:  plan.Features.AllowCancellation,
			AllowInutilization: plan.Features.AllowInutilization,
			WebhookSupport:     plan.Features.WebhookSupport,
			PrioritySupport:    plan.Features.PrioritySupport,
			StorageDays:        plan.Features.StorageDays,
		},
		IsPopular: plan.IsPopular,
		SortOrder: plan.SortOrder,
		TrialDays: plan.TrialDays,
		CreatedAt: plan.CreatedAt,
		UpdatedAt: plan.UpdatedAt,
	}
}

// ToPlanEntity converts a PlanDTO to a Plan entity
func (m *PlanMapper) ToPlanEntity(plan *dto.PlanDTO) *entity.Plan {
	return &entity.Plan{
		ID:              plan.ID,
		Name:            plan.Name,
		Description:     plan.Description,
		Type:            entity.PlanType(plan.Type),
		BillingCycle:    entity.BillingCycle(plan.BillingCycle),
		Status:          entity.PlanStatus(plan.Status),
		Price:           plan.Price,
		Currency:        plan.Currency,
		QuotaType:       entity.QuotaType(plan.QuotaType),
		MaxNFCePerMonth: plan.MaxNFCePerMonth,
		MaxNFCeTotal:    plan.MaxNFCeTotal,
		Features: entity.PlanFeatures{
			MaxNFCePerMonth:    plan.Features.MaxNFCePerMonth,
			MaxNFCeTotal:       plan.Features.MaxNFCeTotal,
			AllowContingency:   plan.Features.AllowContingency,
			AllowCancellation:  plan.Features.AllowCancellation,
			AllowInutilization: plan.Features.AllowInutilization,
			WebhookSupport:     plan.Features.WebhookSupport,
			PrioritySupport:    plan.Features.PrioritySupport,
			StorageDays:        plan.Features.StorageDays,
		},
		IsPopular: plan.IsPopular,
		SortOrder: plan.SortOrder,
		TrialDays: plan.TrialDays,
		CreatedAt: plan.CreatedAt,
		UpdatedAt: plan.UpdatedAt,
	}
}

// ToPlanListDTO converts a slice of Plan entities to PlanListResponse
func (m *PlanMapper) ToPlanListDTO(plans []*entity.Plan) dto.PlanListResponse {
	dtos := make([]dto.PlanDTO, len(plans))
	for i, plan := range plans {
		dtos[i] = *m.ToPlanDTO(plan)
	}

	return dto.PlanListResponse{
		Plans: dtos,
		Total: len(dtos),
	}
}
