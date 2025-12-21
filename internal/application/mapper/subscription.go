package mapper

import (
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
)

// SubscriptionMapper handles mapping between subscription entities and DTOs
type SubscriptionMapper struct{}

// NewSubscriptionMapper creates a new SubscriptionMapper
func NewSubscriptionMapper() *SubscriptionMapper {
	return &SubscriptionMapper{}
}

// ToSubscriptionDTO converts a Subscription entity to a SubscriptionDTO
func (m *SubscriptionMapper) ToSubscriptionDTO(subscription *entity.Subscription) *dto.SubscriptionDTO {
	dtoSubscription := &dto.SubscriptionDTO{
		ID:          subscription.ID,
		CompanyID:   subscription.CompanyID,
		PlanID:      subscription.PlanID,
		Status:      dto.SubscriptionStatus(subscription.Status),
		StartedAt:   subscription.StartedAt,
		EndsAt:      subscription.EndsAt,
		CanceledAt:  subscription.CanceledAt,
		SuspendedAt: subscription.SuspendedAt,
		IsTrial:     subscription.IsTrial,
		TrialEndsAt: subscription.TrialEndsAt,
		CurrentUsage: dto.UsageStats{
			PeriodStart:   subscription.CurrentUsage.PeriodStart,
			PeriodEnd:     subscription.CurrentUsage.PeriodEnd,
			NFCeIssued:    subscription.CurrentUsage.NFCeIssued,
			NFCeRemaining: subscription.CurrentUsage.NFCeRemaining,
			LastNFCeAt:    subscription.CurrentUsage.LastNFCeAt,
		},
		BillingInfo: dto.BillingInfo{
			NextBillingAt: subscription.BillingInfo.NextBillingAt,
			LastBilledAt:  subscription.BillingInfo.LastBilledAt,
			Amount:        subscription.BillingInfo.Amount,
			Currency:      subscription.BillingInfo.Currency,
			PaymentMethod: subscription.BillingInfo.PaymentMethod,
		},
		AutoRenew:    subscription.AutoRenew,
		CancelReason: subscription.CancelReason,
		CreatedAt:    subscription.CreatedAt,
		UpdatedAt:    subscription.UpdatedAt,
	}

	// Include references if populated
	if subscription.Company != nil {
		companyMapper := NewCompanyMapper()
		dtoSubscription.Company = companyMapper.ToCompanyDTO(subscription.Company)
	}

	if subscription.Plan != nil {
		planMapper := NewPlanMapper()
		dtoSubscription.Plan = planMapper.ToPlanDTO(subscription.Plan)
	}

	return dtoSubscription
}

// ToSubscriptionEntity converts a SubscriptionDTO to a Subscription entity
func (m *SubscriptionMapper) ToSubscriptionEntity(subscription *dto.SubscriptionDTO) *entity.Subscription {
	return &entity.Subscription{
		ID:          subscription.ID,
		CompanyID:   subscription.CompanyID,
		PlanID:      subscription.PlanID,
		Status:      entity.SubscriptionStatus(subscription.Status),
		StartedAt:   subscription.StartedAt,
		EndsAt:      subscription.EndsAt,
		CanceledAt:  subscription.CanceledAt,
		SuspendedAt: subscription.SuspendedAt,
		IsTrial:     subscription.IsTrial,
		TrialEndsAt: subscription.TrialEndsAt,
		CurrentUsage: entity.UsageStats{
			PeriodStart:   subscription.CurrentUsage.PeriodStart,
			PeriodEnd:     subscription.CurrentUsage.PeriodEnd,
			NFCeIssued:    subscription.CurrentUsage.NFCeIssued,
			NFCeRemaining: subscription.CurrentUsage.NFCeRemaining,
			LastNFCeAt:    subscription.CurrentUsage.LastNFCeAt,
		},
		BillingInfo: entity.BillingInfo{
			NextBillingAt: subscription.BillingInfo.NextBillingAt,
			LastBilledAt:  subscription.BillingInfo.LastBilledAt,
			Amount:        subscription.BillingInfo.Amount,
			Currency:      subscription.BillingInfo.Currency,
			PaymentMethod: subscription.BillingInfo.PaymentMethod,
		},
		AutoRenew:    subscription.AutoRenew,
		CancelReason: subscription.CancelReason,
		CreatedAt:    subscription.CreatedAt,
		UpdatedAt:    subscription.UpdatedAt,
	}
}

// ToSubscriptionListDTO converts a slice of Subscription entities to SubscriptionListResponse
func (m *SubscriptionMapper) ToSubscriptionListDTO(subscriptions []*entity.Subscription) dto.SubscriptionListResponse {
	dtos := make([]dto.SubscriptionDTO, len(subscriptions))
	for i, subscription := range subscriptions {
		dtos[i] = *m.ToSubscriptionDTO(subscription)
	}

	return dto.SubscriptionListResponse{
		Subscriptions: dtos,
		Total:         len(dtos),
	}
}
