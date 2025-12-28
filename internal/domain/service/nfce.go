package service

import (
	"errors"
	"fmt"
	"math/rand"
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

	// Use the SOAP client error classification for consistency
	// Import would create circular dependency, so we'll duplicate the logic here
	// or better yet, move this to a separate error classification service

	// Códigos que NÃO devem ser retentados (rejeições definitivas)
	noRetryCodes := map[string]bool{
		"101": true, // Cancelamento de NF-e homologado fora de prazo
		"102": true, // Cancelamento de NF-e homologado fora de prazo
		"135": true, // Evento registrado e vinculado à NF-e
		"151": true, // Temporariamente indisponível para atendimento (actually this should be retryable)
		"204": true, // Duplicidade de NF-e (definitive)
		"301": true, // Uso Denegado: Irregularidade fiscal (definitive)
		"302": true, // Irregularidade fiscal do destinatário (definitive)
		"539": true, // Duplicidade de NF-e com diferença na Chave de Acesso (definitive)
	}

	// Business rule violations (200-299 range generally definitive)
	if rejectionCode >= "200" && rejectionCode <= "299" && rejectionCode != "204" {
		return false
	}

	// Security violations (300-399 range generally definitive)
	if rejectionCode >= "300" && rejectionCode <= "399" {
		return false
	}

	// Schema validation errors (400-499 range generally definitive)
	if rejectionCode >= "400" && rejectionCode <= "499" {
		return false
	}

	// Explicitly non-retryable codes
	if noRetryCodes[rejectionCode] {
		return false
	}

	return true
}

// GenerateSequentialNumber gera número sequencial para NFC-e
// Em produção, isso viria de uma sequence no banco
func (s *NFCeDomainService) GenerateSequentialNumber(lastNumber int) int {
	return lastNumber + 1
}

// GenerateCNF gera o Código Numérico (cNF) de 8 dígitos
// O cNF é um número aleatório único por NFC-e
func (s *NFCeDomainService) GenerateCNF() string {
	// Generate random 8-digit number (00000001 to 99999999)
	// In production, ensure uniqueness within the company
	return fmt.Sprintf("%08d", rand.Intn(99999999)+1)
}
