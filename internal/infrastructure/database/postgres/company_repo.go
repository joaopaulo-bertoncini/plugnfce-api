package postgres

import (
	"context"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/ports"
	"gorm.io/gorm"
)

// Company repository implementation
type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) ports.CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) Create(ctx context.Context, company *entity.Company) error {
	return r.db.WithContext(ctx).Create(company).Error
}

func (r *companyRepository) GetByID(ctx context.Context, id string) (*entity.Company, error) {
	var company entity.Company
	err := r.db.WithContext(ctx).First(&company, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) GetByCNPJ(ctx context.Context, cnpj string) (*entity.Company, error) {
	var company entity.Company
	err := r.db.WithContext(ctx).Where("cnpj = ?", cnpj).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) Update(ctx context.Context, company *entity.Company) error {
	return r.db.WithContext(ctx).Save(company).Error
}

func (r *companyRepository) List(ctx context.Context, limit, offset int) ([]*entity.Company, int, error) {
	var companies []*entity.Company
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Company{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&companies).Error
	return companies, int(total), err
}

func (r *companyRepository) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Company{}).Count(&count).Error
	return int(count), err
}

func (r *companyRepository) CountByStatus(ctx context.Context, status entity.CompanyStatus) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Company{}).Where("status = ?", status).Count(&count).Error
	return int(count), err
}

// Default NFC-e series used by the system
const defaultNFCeSerie = "1"

// GetNextNFCeNumber atomically gets and increments the next NFC-e number for a company
func (r *companyRepository) GetNextNFCeNumber(ctx context.Context, companyID string) (int64, error) {
	var nextNumber int64

	// Use PostgreSQL function for atomic sequence generation with default series
	err := r.db.WithContext(ctx).Raw("SELECT get_next_nfce_number(?::uuid, ?)", companyID, defaultNFCeSerie).Scan(&nextNumber).Error
	if err != nil {
		return 0, err
	}

	return nextNumber, nil
}

// GetCertificateByCompanyID retrieves the certificate for a company
func (r *companyRepository) GetCertificateByCompanyID(ctx context.Context, companyID string) (*entity.Certificate, error) {
	var company entity.Company
	err := r.db.WithContext(ctx).Select("certificado_pfx_data, certificado_password").First(&company, "id = ? AND status = ?", companyID, entity.CompanyStatusActive).Error
	if err != nil {
		return nil, err
	}

	// Check if certificate exists
	if company.Certificado.PFXData == nil || company.Certificado.Password == "" {
		return nil, gorm.ErrRecordNotFound
	}

	// Convert bytea to base64 string (assuming it's stored as base64 in the DB)
	certData := &entity.Certificate{
		PFXBase64: string(company.Certificado.PFXData),
		Password:  company.Certificado.Password,
	}

	return certData, nil
}
