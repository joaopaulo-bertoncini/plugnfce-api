package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
)

// NFCeDomainService contém regras de negócio puras da NFC-e
// Não depende de infraestrutura (banco, APIs externas, etc.)
type NFCeDomainService struct{}

// NewNFCeDomainService cria uma nova instância do serviço de domínio
func NewNFCeDomainService() *NFCeDomainService {
	return &NFCeDomainService{}
}

// ValidateCSC valida se o CSC (Código de Segurança do Contribuinte) é válido
// Regras: CSC deve ter exatamente 36 caracteres (32 dígitos + 4 letras)
func (s *NFCeDomainService) ValidateCSC(csc string) error {
	if len(csc) != 36 {
		return errors.New("CSC deve ter exatamente 36 caracteres")
	}

	// CSC deve ter formato: 32 dígitos + 4 letras maiúsculas
	digits := csc[:32]
	letters := csc[32:]

	// Validar dígitos
	for _, char := range digits {
		if char < '0' || char > '9' {
			return errors.New("primeiros 32 caracteres do CSC devem ser dígitos")
		}
	}

	// Validar letras maiúsculas
	for _, char := range letters {
		if char < 'A' || char > 'Z' {
			return errors.New("últimos 4 caracteres do CSC devem ser letras maiúsculas")
		}
	}

	return nil
}

// CalculateTotal calcula o valor total da NFC-e
// Inclui produtos + frete - descontos
func (s *NFCeDomainService) CalculateTotal(items []entity.Item, freight, discount float64) (float64, error) {
	if len(items) == 0 {
		return 0, errors.New("NFC-e deve ter pelo menos um item")
	}

	var totalProducts float64
	for _, item := range items {
		if item.Quantidade <= 0 {
			return 0, fmt.Errorf("quantidade do item %s deve ser maior que zero", item.Descricao)
		}
		if item.Valor <= 0 {
			return 0, fmt.Errorf("preço unitário do item %s deve ser maior que zero", item.Descricao)
		}
		totalProducts += item.Quantidade * item.Valor
	}

	total := totalProducts + freight - discount

	if total <= 0 {
		return 0, errors.New("valor total da NFC-e deve ser maior que zero")
	}

	return total, nil
}

// ValidateEmissionDate valida se a data de emissão é válida
// Regras: não pode ser futura, não pode ser muito antiga
func (s *NFCeDomainService) ValidateEmissionDate(emissionDate time.Time) error {
	now := time.Now()

	// Não pode ser futura (tolerância de 5 minutos para sincronismo de relógio)
	if emissionDate.After(now.Add(5 * time.Minute)) {
		return errors.New("data de emissão não pode ser futura")
	}

	// Não pode ser muito antiga (máximo 5 dias atrás)
	if emissionDate.Before(now.AddDate(0, 0, -5)) {
		return errors.New("data de emissão não pode ser anterior a 5 dias")
	}

	return nil
}

// ValidateContingencyRules valida regras específicas para emissão em contingência
func (s *NFCeDomainService) ValidateContingencyRules(req *entity.Request, contingencyType string) error {
	if req.Status != entity.RequestStatusAuthorized {
		return errors.New("NFC-e deve estar autorizada para usar contingência")
	}

	validTypes := []string{"SVC-AN", "SVC-RS", "OFFLINE"}
	isValidType := false
	for _, validType := range validTypes {
		if contingencyType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		return fmt.Errorf("tipo de contingência inválido: %s", contingencyType)
	}

	// Outras regras específicas de contingência podem ser adicionadas aqui

	return nil
}

// IsRetryAllowed determina se uma requisição pode ser retentada
// Baseado no código de rejeição e contador de tentativas
func (s *NFCeDomainService) IsRetryAllowed(rejectionCode string, retryCount int, maxRetries int) bool {
	if retryCount >= maxRetries {
		return false
	}

	// Códigos que NÃO devem ser retentados (rejeições definitivas)
	noRetryCodes := []string{
		"101", // Cancelamento de NF-e homologado fora de prazo
		"102", // Cancelamento de NF-e homologado fora de prazo
		"135", // Evento registrado e vinculado à NF-e
		"151", // Temporariamente indisponível para atendimento
		"301", // Uso Denegado
		"302", // Irregularidade fiscal do emitente
	}

	for _, code := range noRetryCodes {
		if rejectionCode == code {
			return false
		}
	}

	return true
}

// GenerateSequentialNumber gera número sequencial para NFC-e
// Em produção, isso viria de uma sequence no banco
func (s *NFCeDomainService) GenerateSequentialNumber(lastNumber int) int {
	return lastNumber + 1
}
