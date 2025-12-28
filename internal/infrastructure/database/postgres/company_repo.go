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

// GetNextNFCeNumber atomically gets and increments the next NFC-e number for a company
func (r *companyRepository) GetNextNFCeNumber(ctx context.Context, companyID string) (int64, error) {
	var nextNumber int64

	// Use a transaction to ensure atomicity
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return 0, tx.Error
	}

	// Get current sequence number
	var company entity.Company
	if err := tx.First(&company, "id = ?", companyID).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	nextNumber = company.UltimoNumeroNFCe + 1

	// Update the sequence
	if err := tx.Model(&entity.Company{}).Where("id = ?", companyID).Update("ultimo_numero_nfce", nextNumber).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}

	return nextNumber, nil
}

// UpdateNFCeSequence updates the NFC-e sequence number for a company (used for rollbacks)
func (r *companyRepository) UpdateNFCeSequence(ctx context.Context, companyID string, lastNumber int64) error {
	return r.db.WithContext(ctx).Model(&entity.Company{}).Where("id = ?", companyID).Update("ultimo_numero_nfce", lastNumber).Error
}
