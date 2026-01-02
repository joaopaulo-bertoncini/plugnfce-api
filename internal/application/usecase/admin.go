package usecase

import (
	"context"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/mapper"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
)

// AdminUseCase defines the interface for admin operations
type AdminUseCase interface {
	CreateCompany(ctx context.Context, req dto.CreateCompanyRequest) (*dto.CompanyDTO, error)
	GetCompany(ctx context.Context, id string) (*dto.CompanyDTO, error)
	ListCompanies(ctx context.Context, limit, offset int) (*dto.CompanyListResponse, error)
	UpdateCompany(ctx context.Context, id string, req dto.UpdateCompanyRequest) error
	CreatePlan(ctx context.Context, req dto.CreatePlanRequest) (*dto.PlanDTO, error)
	GetPlan(ctx context.Context, id string) (*dto.PlanDTO, error)
	ListPlans(ctx context.Context, limit, offset int) (*dto.PlanListResponse, error)
	UpdatePlan(ctx context.Context, id string, req dto.UpdatePlanRequest) error
	CreateSubscription(ctx context.Context, req dto.CreateSubscriptionRequest) (*dto.SubscriptionDTO, error)
	GetSubscription(ctx context.Context, id string) (*dto.SubscriptionDTO, error)
	ListSubscriptions(ctx context.Context, limit, offset int) (*dto.SubscriptionListResponse, error)
	UpdateSubscription(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) error
}

// AdminUseCaseImpl handles admin operations
type AdminUseCaseImpl struct {
	companyRepo        ports.CompanyRepository
	planRepo           ports.PlanRepository
	subscriptionRepo   ports.SubscriptionRepository
	nfceRepo           ports.NFCeRepository
	companyMapper      *mapper.CompanyMapper
	planMapper         *mapper.PlanMapper
	subscriptionMapper *mapper.SubscriptionMapper
}

// NewAdminUseCase creates a new AdminUseCase
func NewAdminUseCase(
	companyRepo ports.CompanyRepository,
	planRepo ports.PlanRepository,
	subscriptionRepo ports.SubscriptionRepository,
) AdminUseCase {
	return &AdminUseCaseImpl{
		companyRepo:        companyRepo,
		planRepo:           planRepo,
		subscriptionRepo:   subscriptionRepo,
		companyMapper:      mapper.NewCompanyMapper(),
		planMapper:         mapper.NewPlanMapper(),
		subscriptionMapper: mapper.NewSubscriptionMapper(),
	}
}

// CreateCompany creates a new company
func (uc *AdminUseCaseImpl) CreateCompany(ctx context.Context, req dto.CreateCompanyRequest) (*dto.CompanyDTO, error) {
	company, err := entity.NewCompany(req.CNPJ, req.RazaoSocial)
	if err != nil {
		return nil, err
	}

	// Apply additional fields from request
	company.NomeFantasia = req.NomeFantasia
	company.InscricaoEstadual = req.InscricaoEstadual
	company.Email = req.Email
	company.Endereco = *mapper.NewCompanyMapper().ToAddressEntity(&req.Endereco)
	company.RegimeTributario = entity.TaxRegime(req.RegimeTributario)

	err = uc.companyRepo.Create(ctx, company)
	if err != nil {
		return nil, err
	}

	return uc.companyMapper.ToCompanyDTO(company), nil
}

// GetCompany gets a company by ID
func (uc *AdminUseCaseImpl) GetCompany(ctx context.Context, id string) (*dto.CompanyDTO, error) {
	company, err := uc.companyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.companyMapper.ToCompanyDTO(company), nil
}

// ListCompanies lists companies with pagination
func (uc *AdminUseCaseImpl) ListCompanies(ctx context.Context, limit, offset int) (*dto.CompanyListResponse, error) {
	companies, total, err := uc.companyRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.CompanyDTO, len(companies))
	for i, company := range companies {
		dtos[i] = *uc.companyMapper.ToCompanyDTO(company)
	}

	return &dto.CompanyListResponse{
		Companies: dtos,
		Total:     total,
	}, nil
}

// UpdateCompany updates a company
func (uc *AdminUseCaseImpl) UpdateCompany(ctx context.Context, id string, req dto.UpdateCompanyRequest) error {
	company, err := uc.companyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates from request
	if req.NomeFantasia != nil {
		company.NomeFantasia = *req.NomeFantasia
	}
	if req.InscricaoEstadual != nil {
		company.InscricaoEstadual = *req.InscricaoEstadual
	}
	if req.Email != nil {
		company.Email = *req.Email
	}
	if req.Endereco != nil {
		company.Endereco = *uc.companyMapper.ToAddressEntity(req.Endereco)
	}
	if req.RegimeTributario != nil {
		company.RegimeTributario = entity.TaxRegime(*req.RegimeTributario)
	}
	if req.Status != nil {
		company.Status = entity.CompanyStatus(*req.Status)
	}

	return uc.companyRepo.Update(ctx, company)
}

// CreatePlan creates a new plan
func (uc *AdminUseCaseImpl) CreatePlan(ctx context.Context, req dto.CreatePlanRequest) (*dto.PlanDTO, error) {
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

// GetPlan gets a plan by ID
func (uc *AdminUseCaseImpl) GetPlan(ctx context.Context, id string) (*dto.PlanDTO, error) {
	plan, err := uc.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.planMapper.ToPlanDTO(plan), nil
}

// ListPlans lists plans with pagination
func (uc *AdminUseCaseImpl) ListPlans(ctx context.Context, limit, offset int) (*dto.PlanListResponse, error) {
	plans, total, err := uc.planRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	response := uc.planMapper.ToPlanListDTO(plans)
	response.Total = total
	return &response, nil
}

// UpdatePlan updates a plan
func (uc *AdminUseCaseImpl) UpdatePlan(ctx context.Context, id string, req dto.UpdatePlanRequest) error {
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

// CreateSubscription creates a new subscription
func (uc *AdminUseCaseImpl) CreateSubscription(ctx context.Context, req dto.CreateSubscriptionRequest) (*dto.SubscriptionDTO, error) {
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

// GetSubscription gets a subscription by ID
func (uc *AdminUseCaseImpl) GetSubscription(ctx context.Context, id string) (*dto.SubscriptionDTO, error) {
	subscription, err := uc.subscriptionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.subscriptionMapper.ToSubscriptionDTO(subscription), nil
}

// ListSubscriptions lists subscriptions with pagination
func (uc *AdminUseCaseImpl) ListSubscriptions(ctx context.Context, limit, offset int) (*dto.SubscriptionListResponse, error) {
	subscriptions, total, err := uc.subscriptionRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	response := uc.subscriptionMapper.ToSubscriptionListDTO(subscriptions)
	response.Total = total
	return &response, nil
}

// UpdateSubscription updates a subscription
func (uc *AdminUseCaseImpl) UpdateSubscription(ctx context.Context, id string, req dto.UpdateSubscriptionRequest) error {
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
