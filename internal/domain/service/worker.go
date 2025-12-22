package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/entity"
	nfceInfra "github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/nfce"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/qr"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/signer"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/soap/soapclient"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/validator"
)

// NFCeWorkerService handles the complete NFC-e emission process
type NFCeWorkerService struct {
	xmlBuilder   nfceInfra.Builder
	xmlSigner    signer.Signer
	xmlValidator validator.XMLValidator
	soapClient   soapclient.Client
	qrGenerator  qr.Generator
}

// NewNFCeWorkerService creates a new NFC-e worker service
func NewNFCeWorkerService(
	xmlBuilder nfceInfra.Builder,
	xmlSigner signer.Signer,
	xmlValidator validator.XMLValidator,
	soapClient soapclient.Client,
	qrGenerator qr.Generator,
) *NFCeWorkerService {
	return &NFCeWorkerService{
		xmlBuilder:   xmlBuilder,
		xmlSigner:    xmlSigner,
		xmlValidator: xmlValidator,
		soapClient:   soapClient,
		qrGenerator:  qrGenerator,
	}
}

// ProcessNFceEmission handles the complete NFC-e emission workflow
func (s *NFCeWorkerService) ProcessNFceEmission(ctx context.Context, nfceRequest *entity.NFCE) error {
	// Update status to processing
	nfceRequest.MarkAsProcessing()

	// Step 1: Check idempotency - if already authorized, skip processing
	if nfceRequest.Status == entity.RequestStatusAuthorized {
		return nil
	}

	// Step 2: Generate chave de acesso
	nfceInput := s.convertToNFCeInput(nfceRequest.Payload)
	nfceData, err := s.xmlBuilder.BuildNFCe(nfceInput)
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
		UF:       nfceRequest.Payload.UF,
		Ambiente: nfceRequest.Payload.Ambiente,
		XML:      signedXML,
	}

	response, err := s.soapClient.Authorize(ctx, authReq)
	if err != nil {
		return fmt.Errorf("SEFAZ authorization failed: %w", err)
	}

	// Step 8: Process SEFAZ response
	switch response.Status {
	case "authorized":
		return s.handleAuthorized(ctx, nfceRequest, chaveAcesso, response)
	case "denied":
		return s.handleRejected(ctx, nfceRequest, response)
	default:
		return fmt.Errorf("unexpected SEFAZ status: %s", response.Status)
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
func (s *NFCeWorkerService) convertToNFCeInput(payload entity.EmitPayload) nfceInfra.NFCeInput {
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
		UF:       payload.UF,
		Ambiente: payload.Ambiente,
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
func (s *NFCeWorkerService) handleAuthorized(ctx context.Context, nfceRequest *entity.NFCE, chaveAcesso string, response soapclient.AuthorizationResponse) error {
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
	}

	qrURL, err := s.qrGenerator.BuildURL(ctx, qrParams)
	if err != nil {
		// Log error but don't fail the process
		fmt.Printf("Failed to generate QR code: %v\n", err)
	}

	// Set storage URLs (placeholder - actual storage implementation needed)
	xmlURL := fmt.Sprintf("s3://bucket/xml/%s.xml", chaveAcesso)
	pdfURL := fmt.Sprintf("s3://bucket/pdf/%s.pdf", chaveAcesso)
	qrCodeURL := qrURL

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

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}
