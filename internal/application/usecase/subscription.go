package usecase

import (
	"context"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/mapper"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
)

// SubscriptionUseCase defines the interface for subscription operations
type SubscriptionUseCase interface {
	Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*dto.SubscriptionDTO, error)
	GetByID(ctx context.Context, id string) (*dto.SubscriptionDTO, error)
	GetCurrent(ctx context.Context, companyID string) (*dto.SubscriptionDTO, error)
	List(ctx context.Context, limit, offset int) (*dto.SubscriptionListResponse, error)
	Update(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) error
	Cancel(ctx context.Context, id string, req dto.CancelSubscriptionRequest) error
	GetUsage(ctx context.Context, companyID string) (*dto.UsageStats, error)
}

// SubscriptionUseCaseImpl handles subscription operations
type SubscriptionUseCaseImpl struct {
	subscriptionRepo   ports.SubscriptionRepository
	planRepo           ports.PlanRepository
	companyRepo        ports.CompanyRepository
	subscriptionMapper *mapper.SubscriptionMapper
}

// NewSubscriptionUseCase creates a new SubscriptionUseCase
func NewSubscriptionUseCase(
	subscriptionRepo ports.SubscriptionRepository,
	planRepo ports.PlanRepository,
	companyRepo ports.CompanyRepository,
) SubscriptionUseCase {
	return &SubscriptionUseCaseImpl{
		subscriptionRepo:   subscriptionRepo,
		planRepo:           planRepo,
		companyRepo:        companyRepo,
		subscriptionMapper: mapper.NewSubscriptionMapper(),
	}
}

// Create creates a new subscription
func (uc *SubscriptionUseCaseImpl) Create(ctx context.Context, req dto.CreateSubscriptionRequest) (*dto.SubscriptionDTO, error) {
	plan, err := uc.planRepo.GetByID(ctx, req.PlanID)
	if err != nil {
		return nil, err
	}

	subscription, err := entity.NewSubscription(req.CompanyID, req.PlanID, plan)
	if err != nil {
		return nil, err
	}

	err = uc.subscriptionRepo.Create(ctx, subscription)
	if err != nil {
		return nil, err
	}

	return uc.subscriptionMapper.ToSubscriptionDTO(subscription), nil
}

// GetByID gets a subscription by ID
func (uc *SubscriptionUseCaseImpl) GetByID(ctx context.Context, id string) (*dto.SubscriptionDTO, error) {
	subscription, err := uc.subscriptionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.subscriptionMapper.ToSubscriptionDTO(subscription), nil
}

// GetCurrent gets the current active subscription for a company
func (uc *SubscriptionUseCaseImpl) GetCurrent(ctx context.Context, companyID string) (*dto.SubscriptionDTO, error) {
	subscription, err := uc.subscriptionRepo.GetActiveByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return uc.subscriptionMapper.ToSubscriptionDTO(subscription), nil
}

// List lists subscriptions with pagination
func (uc *SubscriptionUseCaseImpl) List(ctx context.Context, limit, offset int) (*dto.SubscriptionListResponse, error) {
	subscriptions, total, err := uc.subscriptionRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	response := uc.subscriptionMapper.ToSubscriptionListDTO(subscriptions)
	response.Total = total
	return &response, nil
}

// Update updates a subscription
func (uc *SubscriptionUseCaseImpl) Update(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) error {
	subscription, err := uc.subscriptionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates from request
	if req.Status != nil {
		subscription.Status = entity.SubscriptionStatus(*req.Status)
	}
	if req.AutoRenew != nil {
		subscription.AutoRenew = *req.AutoRenew
	}
	if req.CancelReason != nil {
		subscription.CancelReason = *req.CancelReason
	}

	return uc.subscriptionRepo.Update(ctx, subscription)
}

// Cancel cancels a subscription
func (uc *SubscriptionUseCaseImpl) Cancel(ctx context.Context, id string, req dto.CancelSubscriptionRequest) error {
	subscription, err := uc.subscriptionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	subscription.Cancel(req.Reason)
	return uc.subscriptionRepo.Update(ctx, subscription)
}

// GetUsage gets the usage statistics for a company's current subscription
func (uc *SubscriptionUseCaseImpl) GetUsage(ctx context.Context, companyID string) (*dto.UsageStats, error) {
	subscription, err := uc.subscriptionRepo.GetActiveByCompanyID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	usageStats := &dto.UsageStats{
		PeriodStart:   subscription.CurrentUsage.PeriodStart,
		PeriodEnd:     subscription.CurrentUsage.PeriodEnd,
		NFCeIssued:    subscription.CurrentUsage.NFCeIssued,
		NFCeRemaining: subscription.CurrentUsage.NFCeRemaining,
		LastNFCeAt:    subscription.CurrentUsage.LastNFCeAt,
	}

	return usageStats, nil
}
