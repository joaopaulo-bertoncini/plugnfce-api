// Package validator provides XML validation against XSD schemas, specifically designed for SEFAZ NFC-e validation.
//
// Example usage:
//
//	validator, err := NewXMLValidator("./schemas")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Download official SEFAZ schemas (one time setup)
//	ctx := context.Background()
//	if err := validator.DownloadSEFAZSchemas(ctx, "4.00"); err != nil {
//		log.Fatal(err)
//	}
//
//	// Validate NFC-e XML
//	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>...`)
//	if err := validator.ValidateNFCe(ctx, xmlData, "4.00"); err != nil {
//		log.Printf("Validation failed: %v", err)
//	}
//
package validator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	xsdvalidate "github.com/terminalstatic/go-xsd-validate"
)

// XMLValidator enforces compliance of generated XML against XSDs.
type XMLValidator interface {
	Validate(ctx context.Context, xml []byte, schemaName string) error
	ValidateNFCe(ctx context.Context, xml []byte, version string) error
	ValidateWithCustomSchema(ctx context.Context, xml []byte, schemaContent []byte) error
	ListAvailableSchemas() ([]string, error)
	DownloadSEFAZSchemas(ctx context.Context, version string) error
}

// xmlValidator implements XMLValidator interface
type xmlValidator struct {
	schemasDir string
	schemas    map[string]*xsdvalidate.XsdHandler
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewXMLValidator creates a new XML validator
func NewXMLValidator(schemasDir string) (XMLValidator, error) {
	validator := &xmlValidator{
		schemasDir: schemasDir,
		schemas:    make(map[string]*xsdvalidate.XsdHandler),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Initialize schemas directory if it doesn't exist
	if err := os.MkdirAll(schemasDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create schemas directory: %w", err)
	}

	return validator, nil
}

// Validate validates XML against XSD schema
func (v *xmlValidator) Validate(ctx context.Context, xmlData []byte, schemaName string) error {
	// Get or load schema
	handler, err := v.getSchema(schemaName)
	if err != nil {
		return fmt.Errorf("failed to load schema %s: %w", schemaName, err)
	}

	// Validate XML against schema
	if err := handler.ValidateMem(xmlData, xsdvalidate.ValidErrDefault); err != nil {
		return fmt.Errorf("XML validation failed for schema %s: %w", schemaName, err)
	}

	return nil
}

// getSchema loads or returns cached XSD schema
func (v *xmlValidator) getSchema(schemaName string) (*xsdvalidate.XsdHandler, error) {
	v.mu.RLock()
	if handler, exists := v.schemas[schemaName]; exists {
		v.mu.RUnlock()
		return handler, nil
	}
	v.mu.RUnlock()

	// Load schema from file
	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check after acquiring write lock
	if handler, exists := v.schemas[schemaName]; exists {
		return handler, nil
	}

	schemaPath := filepath.Join(v.schemasDir, schemaName+".xsd")
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("schema file not found: %s", schemaPath)
	}

	handler, err := xsdvalidate.NewXsdHandlerUrl(schemaPath, xsdvalidate.ParsErrDefault)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema from file: %w", err)
	}

	v.schemas[schemaName] = handler
	return handler, nil
}

// ValidateNFCe validates NFC-e XML against the appropriate schema
func (v *xmlValidator) ValidateNFCe(ctx context.Context, xmlData []byte, version string) error {
	// NFC-e schema naming convention (e.g., "nfe_v4.00.xsd" for version 4.00)
	schemaName := fmt.Sprintf("nfe_v%s", version)
	return v.Validate(ctx, xmlData, schemaName)
}

// ValidateWithCustomSchema validates XML against a custom schema content
func (v *xmlValidator) ValidateWithCustomSchema(ctx context.Context, xmlData []byte, schemaContent []byte) error {
	// Create temporary XSD handler from content
	handler, err := xsdvalidate.NewXsdHandlerMem(schemaContent, xsdvalidate.ParsErrDefault)
	if err != nil {
		return fmt.Errorf("failed to load custom schema: %w", err)
	}
	defer handler.Free()

	// Validate XML against schema
	if err := handler.ValidateMem(xmlData, xsdvalidate.ValidErrDefault); err != nil {
		return fmt.Errorf("XML validation failed against custom schema: %w", err)
	}

	return nil
}

// DownloadSEFAZSchemas downloads official SEFAZ schemas for NFC-e
func (v *xmlValidator) DownloadSEFAZSchemas(ctx context.Context, version string) error {
	// NFC-e schemas required for version 4.00
	schemas := map[string]string{
		"nfe_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/nfe_v4.00.xsd",
		"infNFe_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/infNFe_v4.00.xsd",
		"infIntermed_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/infIntermed_v4.00.xsd",
		"infRespTec_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/infRespTec_v4.00.xsd",
		"infSolicNFF_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/infSolicNFF_v4.00.xsd",
		"procNFe_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/procNFe_v4.00.xsd",
		"retConsSitNFe_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/retConsSitNFe_v4.00.xsd",
		"retConsStatServ_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/retConsStatServ_v4.00.xsd",
		"retEnviNFe_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/retEnviNFe_v4.00.xsd",
		"retInutNFe_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/retInutNFe_v4.00.xsd",
		"tiposBasico_v4.00.xsd": "http://www.portalfiscal.inf.br/nfe/xsd/tiposBasico_v4.00.xsd",
	}

	// Download each schema
	for schemaName, url := range schemas {
		if err := v.downloadSchema(ctx, schemaName, url); err != nil {
			return fmt.Errorf("failed to download schema %s: %w", schemaName, err)
		}
	}

	// Clear cache to force reload of updated schemas
	v.mu.Lock()
	v.schemas = make(map[string]*xsdvalidate.XsdHandler)
	v.mu.Unlock()

	return nil
}

// downloadSchema downloads a single schema file
func (v *xmlValidator) downloadSchema(ctx context.Context, schemaName, url string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error %d downloading schema", resp.StatusCode)
	}

	// Create schema file
	schemaPath := filepath.Join(v.schemasDir, schemaName)
	file, err := os.Create(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to create schema file: %w", err)
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	return nil
}

// ListAvailableSchemas returns list of available schema files
func (v *xmlValidator) ListAvailableSchemas() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(v.schemasDir, "*.xsd"))
	if err != nil {
		return nil, fmt.Errorf("failed to list schema files: %w", err)
	}

	schemas := make([]string, 0, len(files))
	for _, file := range files {
		schemas = append(schemas, filepath.Base(file[:len(file)-4])) // Remove .xsd extension
	}

	return schemas, nil
}
