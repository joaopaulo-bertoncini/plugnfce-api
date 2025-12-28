package service

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	nfceInfra "github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/nfce"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/qr"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/signer"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/soap/soapclient"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/validator"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/storage"
	"github.com/jung-kurt/gofpdf"
)

// NFCeWorkerService handles the complete NFC-e emission process
type NFCeWorkerService struct {
	xmlBuilder   nfceInfra.Builder
	xmlSigner    signer.Signer
	xmlValidator validator.XMLValidator
	soapClient   soapclient.Client
	qrGenerator  qr.Generator
	storage      storage.StorageService
}

// NewNFCeWorkerService creates a new NFC-e worker service
func NewNFCeWorkerService(
	xmlBuilder nfceInfra.Builder,
	xmlSigner signer.Signer,
	xmlValidator validator.XMLValidator,
	soapClient soapclient.Client,
	qrGenerator qr.Generator,
	storage storage.StorageService,
) *NFCeWorkerService {
	return &NFCeWorkerService{
		xmlBuilder:   xmlBuilder,
		xmlSigner:    xmlSigner,
		xmlValidator: xmlValidator,
		soapClient:   soapClient,
		qrGenerator:  qrGenerator,
		storage:      storage,
	}
}

// ProcessNFceEmission handles the complete NFC-e emission workflow
func (s *NFCeWorkerService) ProcessNFceEmission(ctx context.Context, nfceRequest *entity.NFCE) error {
	return s.processNFceEmissionWithContingency(ctx, nfceRequest, false, "")
}

// processNFceEmissionWithContingency handles NFC-e emission with optional contingency
func (s *NFCeWorkerService) processNFceEmissionWithContingency(ctx context.Context, nfceRequest *entity.NFCE, contingency bool, contingencyType string) error {
	// Update status to processing
	nfceRequest.MarkAsProcessing()

	// Step 1: Check idempotency - if already authorized, skip processing
	if nfceRequest.Status == entity.RequestStatusAuthorized {
		return nil
	}

	// Step 2: Generate chave de acesso
	nfceInput := s.convertToNFCeInput(nfceRequest.Payload, contingency, contingencyType)
	nfceData, err := s.xmlBuilder.BuildNFCe(nfceInput, nfceRequest.CompanyID)
	if err != nil {
		return fmt.Errorf("failed to build NFC-e XML: %w", err)
	}

	// The chave de acesso is generated inside BuildNFCe and set in the XML
	// Extract it from the built XML
	chaveAcesso, err := s.extractChaveAcesso(nfceData)
	if err != nil {
		return fmt.Errorf("failed to extract chave acesso: %w", err)
	}

	// Step 3: Convert to XML bytes for signing
	xmlBytes, err := s.convertNFCeToXML(nfceData)
	if err != nil {
		return fmt.Errorf("failed to convert NFC-e to XML: %w", err)
	}

	// Step 4: Validate XML against XSD schema before signing
	if err := s.xmlValidator.ValidateNFCe(ctx, xmlBytes, "4.00"); err != nil {
		return fmt.Errorf("XSD validation failed: %w", err)
	}

	// Step 5: Sign the XML
	keyMaterial := signer.KeyMaterial{
		PFXBase64: nfceRequest.Payload.Certificado.PFXBase64,
		Password:  nfceRequest.Payload.Certificado.Password,
	}

	// Find the ID of the infNFe element for signing
	infNFeID, err := s.findInfNFeID(xmlBytes)
	if err != nil {
		return fmt.Errorf("failed to find infNFe ID: %w", err)
	}

	signedXML, err := s.xmlSigner.SignEnveloped(ctx, xmlBytes, keyMaterial, infNFeID)
	if err != nil {
		return fmt.Errorf("failed to sign XML: %w", err)
	}

	// Step 6: Validate signed XML against XSD schema
	if err := s.xmlValidator.ValidateNFCe(ctx, signedXML, "4.00"); err != nil {
		return fmt.Errorf("signed XML validation failed: %w", err)
	}

	// Step 7: Send to SEFAZ
	authReq := soapclient.AuthorizationRequest{
		UF:              nfceRequest.Payload.UF,
		Ambiente:        nfceRequest.Payload.Ambiente,
		XML:             signedXML,
		Contingency:     contingency,
		ContingencyType: contingencyType,
	}

	response, err := s.soapClient.Authorize(ctx, authReq)
	if err != nil {
		return fmt.Errorf("SEFAZ authorization failed: %w", err)
	}

	// Step 8: Process SEFAZ response
	switch response.Status {
	case "authorized":
		return s.handleAuthorized(ctx, nfceRequest, chaveAcesso, signedXML, response)
	case "denied":
		return s.handleRejected(ctx, nfceRequest, response)
	default:
		// Check if we should try contingency for service unavailable errors
		if s.shouldUseContingency(response.CStat) && !contingency {
			return s.tryContingency(ctx, nfceRequest)
		}

		// Check if it's a retryable error
		if soapclient.IsRetryableError(response.CStat) {
			return fmt.Errorf("SEFAZ error (retryable): cStat=%s, motivo=%s", response.CStat, response.Motivo)
		}
		// Non-retryable error
		nfceRequest.MarkAsRejected(response.CStat, response.Motivo)
		return fmt.Errorf("SEFAZ error (non-retryable): cStat=%s, motivo=%s", response.CStat, response.Motivo)
	}
}

// extractChaveAcesso extracts the access key from the NFC-e XML
func (s *NFCeWorkerService) extractChaveAcesso(nfceData *nfceInfra.NFCe) (string, error) {
	// The chave acesso is in the Id field of infNFe, format: "NFe{CHAVE}"
	if nfceData.InfNFe.Id == "" {
		return "", fmt.Errorf("infNFe ID is empty")
	}

	// Remove "NFe" prefix to get the chave
	if len(nfceData.InfNFe.Id) < 3 || nfceData.InfNFe.Id[:3] != "NFe" {
		return "", fmt.Errorf("invalid infNFe ID format: %s", nfceData.InfNFe.Id)
	}

	return nfceData.InfNFe.Id[3:], nil
}

// convertNFCeToXML converts NFC-e struct to XML bytes
func (s *NFCeWorkerService) convertNFCeToXML(nfceData *nfceInfra.NFCe) ([]byte, error) {
	// Marshal to XML
	xmlBytes, err := xml.MarshalIndent(nfceData, "", "  ")
	if err != nil {
		return nil, err
	}

	// Add XML declaration
	xmlWithDeclaration := []byte(`<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(xmlBytes))

	return xmlWithDeclaration, nil
}

// findInfNFeID finds the ID attribute of the infNFe element
func (s *NFCeWorkerService) findInfNFeID(xmlBytes []byte) (string, error) {
	// Parse XML to find infNFe ID
	// This is a simplified implementation - in production, use proper XML parsing
	xmlStr := string(xmlBytes)

	// Look for Id="NFe..." in the XML
	const idPrefix = `Id="NFe`
	start := len(idPrefix)
	if idx := findInString(xmlStr, idPrefix); idx != -1 {
		// Find the closing quote
		idStart := idx + start
		if endIdx := findInString(xmlStr[idStart:], `"`); endIdx != -1 {
			return xmlStr[idStart : idStart+endIdx], nil
		}
	}

	return "", fmt.Errorf("infNFe ID not found in XML")
}

// findInString finds substring in string and returns index
func findInString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// convertToNFCeInput converts entity payload to NFC-e builder input
func (s *NFCeWorkerService) convertToNFCeInput(payload entity.EmitPayload, contingency bool, contingencyType string) nfceInfra.NFCeInput {
	// Convert entity types to infrastructure types
	itens := make([]nfceInfra.ItemInput, len(payload.Itens))
	for i, item := range payload.Itens {
		itens[i] = nfceInfra.ItemInput{
			CProd:    item.GTIN, // Using GTIN as product code
			CEAN:     &item.GTIN,
			XProd:    item.Descricao,
			NCM:      item.NCM,
			CFOP:     item.CFOP,
			UCom:     item.Unidade,
			QCom:     fmt.Sprintf("%.4f", item.Quantidade),
			VUnCom:   fmt.Sprintf("%.10f", item.Valor),
			VProd:    fmt.Sprintf("%.2f", item.Quantidade*item.Valor),
			CEANTrib: &item.GTIN,
			UTrib:    item.Unidade,
			QTrib:    fmt.Sprintf("%.4f", item.Quantidade),
			VUnTrib:  fmt.Sprintf("%.10f", item.Valor),
			IndTot:   "1", // Always totalize
		}
	}

	pagamentos := make([]nfceInfra.PagamentoInput, len(payload.Pagamentos))
	for i, pag := range payload.Pagamentos {
		pagamentos[i] = nfceInfra.PagamentoInput{
			TPag: pag.Forma,
			VPag: fmt.Sprintf("%.2f", pag.Valor),
		}
	}

	return nfceInfra.NFCeInput{
		UF:              payload.UF,
		Ambiente:        payload.Ambiente,
		Contingency:     contingency,
		ContingencyType: contingencyType,
		Emitente: nfceInfra.EmitenteInput{
			CNPJ:  payload.Emitente.CNPJ,
			XNome: "EMPRESA EXEMPLO", // Should come from payload
			XFant: stringPtr("EXEMPLO"),
			EnderEmit: nfceInfra.EnderEmitInput{
				XLgr:    "RUA EXEMPLO",
				Nro:     "123",
				XBairro: "CENTRO",
				CMun:    "3550308", // São Paulo
				XMun:    "SAO PAULO",
				UF:      payload.UF,
				CEP:     "01234567",
				CPais:   stringPtr("1058"),
				XPais:   stringPtr("BRASIL"),
				Fone:    stringPtr("11999999999"),
			},
			IE:  payload.Emitente.IE,
			CRT: payload.Emitente.Regime, // Simples Nacional
		},
		Itens:      itens,
		Pagamentos: pagamentos,
		Transp: nfceInfra.TranspInput{
			ModFrete: "9", // Sem frete
		},
	}
}

// handleAuthorized processes successful SEFAZ authorization
func (s *NFCeWorkerService) handleAuthorized(ctx context.Context, nfceRequest *entity.NFCE, chaveAcesso string, signedXML []byte, response soapclient.AuthorizationResponse) error {
	// Extract protocol and other data from response
	protocolo := response.Protocolo
	numero := "1" // Should be extracted from response or generated
	serie := "1"  // Should be extracted from response or configured

	// Mark as authorized
	nfceRequest.MarkAsAuthorized(chaveAcesso, protocolo, numero, serie)

	// Generate QR Code
	qrParams := qr.Params{
		ChaveAcesso: chaveAcesso,
		TpAmb:       nfceRequest.Payload.Ambiente,
		DhEmi:       time.Now().Format("2006-01-02T15:04:05-07:00"),
		VNF:         "100.00",       // Should calculate from items
		VICMS:       "0.00",         // Should calculate from taxes
		DigVal:      "dummy_digest", // Should extract from signed XML
		CSCID:       nfceRequest.Payload.Emitente.CSCID,
		CSCToken:    nfceRequest.Payload.Emitente.CSCToken,
		UF:          nfceRequest.Payload.UF,
		Contingency: nfceRequest.InContingency,
	}

	qrURL, err := s.qrGenerator.BuildURL(ctx, qrParams)
	if err != nil {
		// Log error but don't fail the process
		fmt.Printf("Failed to generate QR code: %v\n", err)
	}

	// Store XML file
	xmlURL, err := s.storeXMLFile(ctx, signedXML, chaveAcesso, nfceRequest.CompanyID)
	if err != nil {
		// Log error but don't fail the process - use fallback URL
		fmt.Printf("Failed to store XML file: %v\n", err)
		xmlURL = fmt.Sprintf("http://localhost:9000/plugnfce/nfce/%s/xml/%s.xml", nfceRequest.CompanyID, chaveAcesso)
	}

	// Generate and store PDF (placeholder for now - would need DANFE generator)
	pdfURL, err := s.generateAndStorePDFFile(ctx, nfceRequest, chaveAcesso)
	if err != nil {
		// Log error but don't fail the process - use fallback URL
		fmt.Printf("Failed to generate/store PDF file: %v\n", err)
		pdfURL = fmt.Sprintf("http://localhost:9000/plugnfce/nfce/%s/pdf/%s.pdf", nfceRequest.CompanyID, chaveAcesso)
	}

	// Store QR Code as image
	qrCodeURL, err := s.storeQRCodeImage(ctx, qrURL, chaveAcesso, nfceRequest.CompanyID, nfceRequest.InContingency)
	if err != nil {
		// Log error but don't fail the process - use fallback URL
		fmt.Printf("Failed to store QR code image: %v\n", err)
		qrCodeURL = fmt.Sprintf("http://localhost:9000/plugnfce/nfce/%s/qr/%s.png", nfceRequest.CompanyID, chaveAcesso)
	}

	nfceRequest.SetStorageURLs(xmlURL, pdfURL, qrCodeURL)

	return nil
}

// handleRejected processes SEFAZ rejection
func (s *NFCeWorkerService) handleRejected(ctx context.Context, nfceRequest *entity.NFCE, response soapclient.AuthorizationResponse) error {
	nfceRequest.MarkAsRejected(response.CStat, response.Motivo)
	return nil
}

// CanRetry determines if the request can be retried
func (s *NFCeWorkerService) CanRetry(nfceRequest *entity.NFCE, maxRetries int) bool {
	return nfceRequest.CanRetry(maxRetries)
}

// IncrementRetry increments the retry counter
func (s *NFCeWorkerService) IncrementRetry(nfceRequest *entity.NFCE) {
	nfceRequest.IncrementRetry()
}

// storeXMLFile uploads the signed XML to storage
func (s *NFCeWorkerService) storeXMLFile(ctx context.Context, xmlContent []byte, chaveAcesso string, companyID string) (string, error) {
	key := fmt.Sprintf("nfce/%s/xml/%s.xml", companyID, chaveAcesso)
	reader := bytes.NewReader(xmlContent)

	url, err := s.storage.UploadFile(ctx, "", key, reader, "application/xml")
	if err != nil {
		return "", fmt.Errorf("failed to upload XML: %w", err)
	}

	return url, nil
}

// generateAndStorePDFFile generates DANFE PDF and uploads it
func (s *NFCeWorkerService) generateAndStorePDFFile(ctx context.Context, nfceRequest *entity.NFCE, chaveAcesso string) (string, error) {
	// Generate real DANFE PDF
	pdfContent := s.generateDANFE(nfceRequest, chaveAcesso)
	key := fmt.Sprintf("nfce/%s/pdf/%s.pdf", nfceRequest.CompanyID, chaveAcesso)
	reader := bytes.NewReader(pdfContent)

	url, err := s.storage.UploadFile(ctx, "", key, reader, "application/pdf")
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}

	return url, nil
}

// generateDANFE generates a real DANFE NFC-e PDF
func (s *NFCeWorkerService) generateDANFE(nfceRequest *entity.NFCE, chaveAcesso string) []byte {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set margins
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 10)

	// Title
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 10, "DOCUMENTO AUXILIAR DA NOTA FISCAL DE CONSUMIDOR ELETRÔNICA")
	pdf.Ln(15)

	// NFC-e Info
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(190, 6, "NFC-e")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 8)

	// Chave de Acesso
	pdf.Cell(30, 5, "Chave de Acesso:")
	pdf.SetFont("Courier", "", 7)
	pdf.MultiCell(160, 3, chaveAcesso, "", "L", false)
	pdf.Ln(2)

	// Emitente
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(190, 5, "EMITENTE")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 8)
	pdf.Cell(20, 4, "CNPJ:")
	pdf.Cell(50, 4, nfceRequest.Payload.Emitente.CNPJ)
	pdf.Cell(20, 4, "IE:")
	pdf.Cell(40, 4, nfceRequest.Payload.Emitente.IE)
	pdf.Cell(15, 4, "UF:")
	pdf.Cell(15, 4, nfceRequest.Payload.UF)
	pdf.Ln(5)

	// Ambiente
	env := "PRODUÇÃO"
	if nfceRequest.Payload.Ambiente == "2" || nfceRequest.Payload.Ambiente == "homologacao" {
		env = "HOMOLOGAÇÃO"
	}
	pdf.Cell(25, 4, "Ambiente:")
	pdf.Cell(40, 4, env)
	pdf.Cell(20, 4, "Número:")
	pdf.Cell(30, 4, nfceRequest.Numero)
	pdf.Cell(20, 4, "Série:")
	pdf.Cell(30, 4, nfceRequest.Serie)
	pdf.Ln(8)

	// Items Table Header
	pdf.SetFont("Arial", "B", 7)
	pdf.SetFillColor(240, 240, 240)

	// Simple table without borders for now
	pdf.Cell(15, 6, "Cód.")
	pdf.Cell(60, 6, "Descrição")
	pdf.Cell(15, 6, "Qtde")
	pdf.Cell(15, 6, "UN")
	pdf.Cell(20, 6, "V. Unit.")
	pdf.Cell(20, 6, "V. Total")
	pdf.Ln(6)

	// Items
	pdf.SetFont("Arial", "", 7)
	totalValue := 0.0

	for i, item := range nfceRequest.Payload.Itens {
		pdf.Cell(15, 5, item.GTIN)
		pdf.Cell(60, 5, truncateString(item.Descricao, 35))
		pdf.Cell(15, 5, fmt.Sprintf("%.2f", item.Quantidade))
		pdf.Cell(15, 5, item.Unidade)
		pdf.Cell(20, 5, fmt.Sprintf("R$ %.2f", item.Valor))

		itemTotal := item.Valor * item.Quantidade
		totalValue += itemTotal
		pdf.Cell(20, 5, fmt.Sprintf("R$ %.2f", itemTotal))
		pdf.Ln(5)

		// Add page break if needed
		if i > 0 && i%20 == 0 && i < len(nfceRequest.Payload.Itens)-1 {
			pdf.AddPage()
		}
	}

	// Totals
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(130, 6, "")
	pdf.Cell(30, 6, "TOTAL R$:")
	pdf.Cell(30, 6, fmt.Sprintf("%.2f", totalValue))
	pdf.Ln(10)

	// Payment Info
	if len(nfceRequest.Payload.Pagamentos) > 0 {
		pdf.SetFont("Arial", "B", 8)
		pdf.Cell(190, 5, "FORMA DE PAGAMENTO")
		pdf.Ln(6)

		pdf.SetFont("Arial", "", 8)
		for _, payment := range nfceRequest.Payload.Pagamentos {
			pdf.Cell(40, 4, payment.Forma)
			pdf.Cell(30, 4, fmt.Sprintf("R$ %.2f", payment.Valor))
			if payment.Troco > 0 {
				pdf.Cell(30, 4, fmt.Sprintf("Troco: R$ %.2f", payment.Troco))
			}
			pdf.Ln(5)
		}
		pdf.Ln(5)
	}

	// Protocol Info
	pdf.SetFont("Arial", "B", 8)
	pdf.Cell(190, 5, "PROTOCOLO DE AUTORIZAÇÃO")
	pdf.Ln(6)

	pdf.SetFont("Courier", "", 8)
	pdf.Cell(190, 4, fmt.Sprintf("Protocolo: %s", nfceRequest.Protocolo))
	pdf.Ln(5)
	if nfceRequest.AuthorizedAt != nil {
		pdf.Cell(190, 4, fmt.Sprintf("Data: %s", nfceRequest.AuthorizedAt.Format("02/01/2006 15:04:05")))
	}
	pdf.Ln(10)

	// Footer
	pdf.SetFont("Arial", "I", 6)
	pdf.MultiCell(190, 3, "Esta NFC-e foi emitida por ME ou EPP optante pelo Simples Nacional. Não gera direito a crédito fiscal de IPI ou ICMS.", "", "L", false)
	pdf.Ln(2)
	pdf.Cell(190, 3, "Emitida em contingência: Não")

	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		// Fallback to simple PDF if gofpdf fails
		return s.generateSimpleFallbackPDF(nfceRequest, chaveAcesso)
	}

	return buf.Bytes()
}

// generateSimpleFallbackPDF creates a minimal PDF if gofpdf fails
func (s *NFCeWorkerService) generateSimpleFallbackPDF(nfceRequest *entity.NFCE, chaveAcesso string) []byte {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 20, "DANFE NFC-e")
	pdf.Ln(20)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(190, 8, fmt.Sprintf("Chave de Acesso: %s", chaveAcesso))
	pdf.Ln(10)
	pdf.Cell(190, 8, fmt.Sprintf("Emitente: %s", nfceRequest.Payload.Emitente.CNPJ))
	pdf.Ln(10)

	totalValue := 0.0
	for _, item := range nfceRequest.Payload.Itens {
		totalValue += item.Valor * item.Quantidade
	}
	pdf.Cell(190, 8, fmt.Sprintf("Valor Total: R$ %.2f", totalValue))

	var buf bytes.Buffer
	pdf.Output(&buf)
	return buf.Bytes()
}

// Helper function to truncate strings
func truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}

// storeQRCodeImage generates QR code image and uploads to storage
func (s *NFCeWorkerService) storeQRCodeImage(ctx context.Context, qrURL, chaveAcesso, companyID string, contingency bool) (string, error) {
	// Extract parameters from the NFC-e request to regenerate QR code
	// For now, we'll use placeholder values - in production, these should come from the request
	qrParams := qr.Params{
		ChaveAcesso: chaveAcesso,
		TpAmb:       "2", // Assume homologation for now
		DhEmi:       time.Now().Format("2006-01-02T15:04:05-07:00"),
		VNF:         "100.00",       // Should be calculated from items
		VICMS:       "0.00",         // Should be calculated from taxes
		DigVal:      "dummy_digest", // Should be extracted from signed XML
		CSCID:       "001",          // Should come from company config
		CSCToken:    "dummy_token",  // Should come from company config
		UF:          "SP",           // Should come from request
		Contingency: contingency,
	}

	// Generate QR code image
	qrImage, err := s.qrGenerator.BuildImage(ctx, qrParams, 256)
	if err != nil {
		// Fallback to storing URL as text if image generation fails
		content := fmt.Sprintf("QR Code URL: %s\nGenerated at: %s", qrURL, time.Now().Format(time.RFC3339))
		key := fmt.Sprintf("nfce/%s/qr/%s.txt", companyID, chaveAcesso)
		reader := strings.NewReader(content)

		url, uploadErr := s.storage.UploadFile(ctx, "", key, reader, "text/plain")
		if uploadErr != nil {
			return "", fmt.Errorf("failed to generate QR image and fallback upload: %w", err)
		}
		return url, nil
	}

	// Upload QR code image
	key := fmt.Sprintf("nfce/%s/qr/%s.png", companyID, chaveAcesso)
	reader := bytes.NewReader(qrImage)

	url, err := s.storage.UploadFile(ctx, "", key, reader, "image/png")
	if err != nil {
		return "", fmt.Errorf("failed to upload QR code image: %w", err)
	}

	return url, nil
}

// shouldUseContingency determines if we should switch to contingency mode based on SEFAZ response
func (s *NFCeWorkerService) shouldUseContingency(cstat string) bool {
	// Contingency should be used for service unavailable errors
	contingencyCodes := map[string]bool{
		"108": true, // Serviço Paralisado Temporariamente (SVC)
		"109": true, // Serviço Paralisado sem Previsão
		"691": true, // Contingência EPEC: Sistema não autorizado
		"692": true, // Contingência SVC: Sistema não autorizado
		"693": true, // Contingência SVC: Autorização não concedida
	}

	// Also use contingency for general server errors (5xx)
	if cstat >= "500" && cstat <= "599" {
		return true
	}

	return contingencyCodes[cstat]
}

// tryContingency attempts to process the NFC-e using contingency mode
func (s *NFCeWorkerService) tryContingency(ctx context.Context, nfceRequest *entity.NFCE) error {
	// Determine which contingency to use based on UF
	contingencyType := "SVC-AN" // Default to SVC-AN
	if nfceRequest.Payload.UF == "RS" {
		contingencyType = "SVC-RS" // Use SVC-RS for Rio Grande do Sul
	}

	// Mark as contingency
	nfceRequest.MarkAsContingency(contingencyType)

	// Retry with contingency
	return s.processNFceEmissionWithContingency(ctx, nfceRequest, true, contingencyType)
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}
