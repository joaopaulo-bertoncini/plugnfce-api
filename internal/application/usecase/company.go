package usecase

import (
	"context"
	"time"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/mapper"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
)

// CompanyUseCase defines the interface for company operations
type CompanyUseCase interface {
	GetProfile(ctx context.Context, companyID string) (*dto.CompanyDTO, error)
	UpdateProfile(ctx context.Context, company *dto.CompanyDTO) error
	UpdateCertificate(ctx context.Context, companyID string, pfxData []byte, password string, expiresAt time.Time) error
	UpdateCSC(ctx context.Context, companyID, cscID, cscToken string, validUntil time.Time) error
}

// CompanyUseCaseImpl handles company operations
type CompanyUseCaseImpl struct {
	companyRepo      ports.CompanyRepository
	subscriptionRepo ports.SubscriptionRepository
}

// NewCompanyUseCase creates a new CompanyUseCase
func NewCompanyUseCase(
	companyRepo ports.CompanyRepository,
	subscriptionRepo ports.SubscriptionRepository,
) CompanyUseCase {
	return &CompanyUseCaseImpl{
		companyRepo:      companyRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// GetProfile gets the company profile
func (uc *CompanyUseCaseImpl) GetProfile(ctx context.Context, companyID string) (*dto.CompanyDTO, error) {
	company, err := uc.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, err
	}

	return mapper.NewCompanyMapper().ToCompanyDTO(company), nil
}

// UpdateProfile updates the company profile
func (uc *CompanyUseCaseImpl) UpdateProfile(ctx context.Context, company *dto.CompanyDTO) error {
	return uc.companyRepo.Update(ctx, mapper.NewCompanyMapper().ToCompanyEntity(company))
}

// UpdateCertificate updates the company certificate
func (uc *CompanyUseCaseImpl) UpdateCertificate(ctx context.Context, companyID string, pfxData []byte, password string, expiresAt time.Time) error {
	company, err := uc.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}

	err = company.UpdateCertificate(entity.CertificateTypeA1, pfxData, password, expiresAt)
	if err != nil {
		return err
	}

	return uc.companyRepo.Update(ctx, company)
}

// UpdateCSC updates the company CSC configuration
func (uc *CompanyUseCaseImpl) UpdateCSC(ctx context.Context, companyID, cscID, cscToken string, validUntil time.Time) error {
	company, err := uc.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return err
	}

	return company.UpdateCSC(cscID, cscToken, validUntil)
}
