package dto

import (
	"time"
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
	CertificateTypeA3 CertificateType = "a3"
)

// CompanyDTO represents company data
type CompanyDTO struct {
	ID                string         `json:"id"`
	CNPJ              string         `json:"cnpj"`
	RazaoSocial       string         `json:"razao_social"`
	NomeFantasia      string         `json:"nome_fantasia,omitempty"`
	InscricaoEstadual string         `json:"inscricao_estadual,omitempty"`
	Email             string         `json:"email"`
	Endereco          AddressDTO     `json:"endereco"`
	Certificado       CertificateDTO `json:"certificado"`
	CSC               CSCDTO         `json:"csc"`
	RegimeTributario  TaxRegime      `json:"regime_tributario"`
	Status            CompanyStatus  `json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// AddressDTO represents address data
type AddressDTO struct {
	Logradouro      string `json:"logradouro"`
	Numero          string `json:"numero"`
	Complemento     string `json:"complemento,omitempty"`
	Bairro          string `json:"bairro"`
	CodigoMunicipio string `json:"codigo_municipio"`
	Municipio       string `json:"municipio"`
	UF              string `json:"uf"`
	CEP             string `json:"cep"`
}

// CertificateDTO represents certificate data
type CertificateDTO struct {
	Type      CertificateType `json:"type"`
	PFXData   []byte          `json:"pfx_data"`
	Password  string          `json:"password"`
	ExpiresAt time.Time       `json:"expires_at"`
	Subject   string          `json:"subject,omitempty"`
	Valid     bool            `json:"valid"`
}

// CSCDTO represents CSC data
type CSCDTO struct {
	CSCID      string    `json:"csc_id"`
	ValidFrom  time.Time `json:"valid_from"`
	ValidUntil time.Time `json:"valid_until"`
	Valid      bool      `json:"valid"`
}

// CreateCompanyRequest represents the request to create a new company
type CreateCompanyRequest struct {
	CNPJ              string     `json:"cnpj" validate:"required"`
	RazaoSocial       string     `json:"razao_social" validate:"required"`
	NomeFantasia      string     `json:"nome_fantasia,omitempty"`
	InscricaoEstadual string     `json:"inscricao_estadual,omitempty"`
	Email             string     `json:"email" validate:"required,email"`
	Endereco          AddressDTO `json:"endereco" validate:"required"`
	RegimeTributario  TaxRegime  `json:"regime_tributario" validate:"required"`
}

// UpdateCompanyRequest represents the request to update a company
type UpdateCompanyRequest struct {
	NomeFantasia      *string        `json:"nome_fantasia,omitempty"`
	InscricaoEstadual *string        `json:"inscricao_estadual,omitempty"`
	Email             *string        `json:"email,omitempty"`
	Endereco          *AddressDTO    `json:"endereco,omitempty"`
	RegimeTributario  *TaxRegime     `json:"regime_tributario,omitempty"`
	Status            *CompanyStatus `json:"status,omitempty"`
}

// UpdateCompanyCSCRequest represents the request to update company CSC
type UpdateCompanyCSCRequest struct {
	CSCID      string    `json:"csc_id" validate:"required"`
	CSCToken   string    `json:"csc_token" validate:"required"`
	ValidUntil time.Time `json:"valid_until" validate:"required"`
}

// CompanyListResponse represents a paginated list of companies
type CompanyListResponse struct {
	Companies []CompanyDTO `json:"companies"`
	Total     int          `json:"total"`
}
