package mapper

import (
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
)

// CompanyMapper handles mapping between company entities and DTOs
type CompanyMapper struct{}

// NewCompanyMapper creates a new CompanyMapper
func NewCompanyMapper() *CompanyMapper {
	return &CompanyMapper{}
}

// ToCompanyDTO converts a Company entity to a CompanyDTO
func (m *CompanyMapper) ToCompanyDTO(company *entity.Company) *dto.CompanyDTO {
	return &dto.CompanyDTO{
		ID:                company.ID,
		CNPJ:              company.CNPJ,
		RazaoSocial:       company.RazaoSocial,
		NomeFantasia:      company.NomeFantasia,
		InscricaoEstadual: company.InscricaoEstadual,
		Email:             company.Email,
		Endereco:          *m.ToAddressDTO(&company.Endereco),
		Certificado:       *m.ToCertificateDTO(&company.Certificado),
		CSC:               *m.ToCSCConfigDTO(&company.CSC),
		RegimeTributario:  dto.TaxRegime(company.RegimeTributario),
		SerieNFCe:         company.SerieNFCe,
		Status:            dto.CompanyStatus(company.Status),
		CreatedAt:         company.CreatedAt,
		UpdatedAt:         company.UpdatedAt,
	}
}

// ToAddressDTO converts an Address entity to a AddressDTO
func (m *CompanyMapper) ToAddressDTO(address *entity.Address) *dto.AddressDTO {
	return &dto.AddressDTO{
		Logradouro:      address.Logradouro,
		Numero:          address.Numero,
		Complemento:     address.Complemento,
		Bairro:          address.Bairro,
		CodigoMunicipio: address.CodigoMunicipio,
		Municipio:       address.Municipio,
		UF:              address.UF,
		CEP:             address.CEP,
	}
}

// ToCertificateDTO converts a DigitalCertificate entity to a CertificateDTO
func (m *CompanyMapper) ToCertificateDTO(certificate *entity.DigitalCertificate) *dto.CertificateDTO {
	return &dto.CertificateDTO{
		Type:      dto.CertificateType(certificate.Type),
		ExpiresAt: certificate.ExpiresAt,
		Subject:   certificate.Subject,
	}
}

// ToCSCConfigDTO converts a CSCConfig entity to a CSCDTO
func (m *CompanyMapper) ToCSCConfigDTO(csc *entity.CSCConfig) *dto.CSCDTO {
	return &dto.CSCDTO{
		CSCID:      csc.CSCID,
		ValidFrom:  csc.ValidFrom,
		ValidUntil: csc.ValidUntil,
	}
}

// ToCompanyEntity converts a CompanyDTO to a Company entity
func (m *CompanyMapper) ToCompanyEntity(company *dto.CompanyDTO) *entity.Company {
	return &entity.Company{
		ID:                company.ID,
		CNPJ:              company.CNPJ,
		RazaoSocial:       company.RazaoSocial,
		NomeFantasia:      company.NomeFantasia,
		InscricaoEstadual: company.InscricaoEstadual,
		Email:             company.Email,
		Endereco:          *m.ToAddressEntity(&company.Endereco),
		Certificado:       *m.ToCertificateEntity(&company.Certificado),
		CSC:               *m.ToCSCConfigEntity(&company.CSC),
		RegimeTributario:  entity.TaxRegime(company.RegimeTributario),
		SerieNFCe:         company.SerieNFCe,
		Status:            entity.CompanyStatus(company.Status),
		CreatedAt:         company.CreatedAt,
		UpdatedAt:         company.UpdatedAt,
	}
}

// ToAddressEntity converts an AddressDTO to a Address entity
func (m *CompanyMapper) ToAddressEntity(address *dto.AddressDTO) *entity.Address {
	return &entity.Address{
		Logradouro:      address.Logradouro,
		Numero:          address.Numero,
		Complemento:     address.Complemento,
		Bairro:          address.Bairro,
		CodigoMunicipio: address.CodigoMunicipio,
		Municipio:       address.Municipio,
		UF:              address.UF,
		CEP:             address.CEP,
	}
}

// ToCertificateEntity converts a CertificateDTO to a Certificate entity
func (m *CompanyMapper) ToCertificateEntity(certificate *dto.CertificateDTO) *entity.DigitalCertificate {
	return &entity.DigitalCertificate{
		Type:      entity.CertificateType(certificate.Type),
		ExpiresAt: certificate.ExpiresAt,
		Subject:   certificate.Subject,
		PFXData:   certificate.PFXData,
		Password:  certificate.Password,
	}
}

// ToCSCConfigEntity converts a CSCDTO to a CSCConfig entity
func (m *CompanyMapper) ToCSCConfigEntity(csc *dto.CSCDTO) *entity.CSCConfig {
	return &entity.CSCConfig{
		CSCID:      csc.CSCID,
		ValidFrom:  csc.ValidFrom,
		ValidUntil: csc.ValidUntil,
	}
}
