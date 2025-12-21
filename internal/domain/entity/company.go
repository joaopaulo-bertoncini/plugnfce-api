package entity

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// CompanyStatus represents the status of a company
type CompanyStatus string

const (
	CompanyStatusActive   CompanyStatus = "active"
	CompanyStatusInactive CompanyStatus = "inactive"
	CompanyStatusBlocked  CompanyStatus = "blocked"
)

// TaxRegime represents the tax regime of the company
type TaxRegime string

const (
	TaxRegimeSimplesNacional TaxRegime = "simples_nacional"
	TaxRegimeLucroPresumido  TaxRegime = "lucro_presumido"
	TaxRegimeLucroReal       TaxRegime = "lucro_real"
)

// CertificateType represents the type of digital certificate
type CertificateType string

const (
	CertificateTypeA1 CertificateType = "a1"
)

// Company represents an NFC-e issuing company
type Company struct {
	ID                string             `json:"id"`
	CNPJ              string             `json:"cnpj"`
	RazaoSocial       string             `json:"razao_social"`
	NomeFantasia      string             `json:"nome_fantasia,omitempty"`
	InscricaoEstadual string             `json:"inscricao_estadual,omitempty"`
	Email             string             `json:"email"`
	Endereco          Address            `json:"endereco"`
	Certificado       DigitalCertificate `json:"certificado"`
	CSC               CSCConfig          `json:"csc"`
	RegimeTributario  TaxRegime          `json:"regime_tributario"`
	SerieNFCe         string             `json:"serie_nfce"` // Série padrão para NFC-e
	Status            CompanyStatus      `json:"status"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// Address represents a company's address
type Address struct {
	Logradouro      string `json:"logradouro"`
	Numero          string `json:"numero"`
	Complemento     string `json:"complemento,omitempty"`
	Bairro          string `json:"bairro"`
	CodigoMunicipio string `json:"codigo_municipio"`
	Municipio       string `json:"municipio"`
	UF              string `json:"uf"`
	CEP             string `json:"cep"`
}

// DigitalCertificate holds the company's digital certificate information
type DigitalCertificate struct {
	Type      CertificateType `json:"type"`
	PFXData   []byte          `json:"pfx_data"` // Encrypted PFX data
	Password  string          `json:"password"` // Certificate password
	ExpiresAt time.Time       `json:"expires_at"`
	Subject   string          `json:"subject,omitempty"` // Certificate subject
}

// CSCConfig holds CSC (Código de Segurança do Contribuinte) configuration
type CSCConfig struct {
	CSCID      string    `json:"csc_id"`
	CSCToken   string    `json:"csc_token"`
	ValidFrom  time.Time `json:"valid_from"`
	ValidUntil time.Time `json:"valid_until"`
}

// NewCompany creates a new company with validation
func NewCompany(cnpj, razaoSocial string) (*Company, error) {
	if err := validateCNPJ(cnpj); err != nil {
		return nil, err
	}

	if razaoSocial == "" {
		return nil, errors.New("razão social é obrigatória")
	}

	now := time.Now()
	return &Company{
		ID:          generateID(), // You might want to use a proper ID generator
		CNPJ:        cnpj,
		RazaoSocial: razaoSocial,
		Status:      CompanyStatusActive,
		SerieNFCe:   "1", // Default series
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateCertificate updates the company's digital certificate
func (c *Company) UpdateCertificate(certType CertificateType, pfxData []byte, password string, expiresAt time.Time) error {
	if len(pfxData) == 0 {
		return errors.New("dados do certificado são obrigatórios")
	}

	if password == "" {
		return errors.New("senha do certificado é obrigatória")
	}

	if expiresAt.Before(time.Now()) {
		return errors.New("certificado já expirou")
	}

	c.Certificado = DigitalCertificate{
		Type:      certType,
		PFXData:   pfxData,
		Password:  password,
		ExpiresAt: expiresAt,
	}
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateCSC updates the company's CSC configuration
func (c *Company) UpdateCSC(cscID, cscToken string, validUntil time.Time) error {
	if cscID == "" {
		return errors.New("CSC ID é obrigatório")
	}

	if cscToken == "" {
		return errors.New("CSC Token é obrigatório")
	}

	if validUntil.Before(time.Now()) {
		return errors.New("CSC já expirou")
	}

	c.CSC = CSCConfig{
		CSCID:      cscID,
		CSCToken:   cscToken,
		ValidFrom:  time.Now(),
		ValidUntil: validUntil,
	}
	c.UpdatedAt = time.Now()
	return nil
}

// IsActive returns true if the company is active
func (c *Company) IsActive() bool {
	return c.Status == CompanyStatusActive
}

// IsCertificateValid returns true if the certificate is still valid
func (c *Company) IsCertificateValid() bool {
	return c.Certificado.ExpiresAt.After(time.Now())
}

// IsCSCValid returns true if the CSC is still valid
func (c *Company) IsCSCValid() bool {
	return c.CSC.ValidUntil.After(time.Now())
}

// validateCNPJ performs basic CNPJ validation
func validateCNPJ(cnpj string) error {
	// Remove non-numeric characters
	re := regexp.MustCompile(`[^\d]`)
	cleanCNPJ := re.ReplaceAllString(cnpj, "")

	if len(cleanCNPJ) != 14 {
		return errors.New("CNPJ deve ter 14 dígitos")
	}

	// Basic validation - you might want to implement full CNPJ validation
	for _, char := range cleanCNPJ {
		if char < '0' || char > '9' {
			return errors.New("CNPJ deve conter apenas números")
		}
	}

	return nil
}

// generateID generates a unique UUID for the company
func generateID() string {
	return uuid.New().String()
}
