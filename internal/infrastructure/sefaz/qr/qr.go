package qr

import (
	"context"
	"crypto/sha1"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Params holds the data required to assemble the NFC-e QR Code v3 URL.
type Params struct {
	ChaveAcesso  string
	TpAmb        string
	Destinatario string // Optional
	DhEmi        string
	VNF          string
	VICMS        string
	DigVal       string
	CSCID        string
	CSCToken     string
	UF           string
}

// Generator builds the URL (and optionally image) for NFC-e QR Code v3.
type Generator interface {
	BuildURL(ctx context.Context, params Params) (string, error)
}

// generator implements Generator interface
type generator struct{}

// NewGenerator creates a new QR Code generator
func NewGenerator() Generator {
	return &generator{}
}

// BuildURL builds the NFC-e QR Code v3 URL
func (g *generator) BuildURL(ctx context.Context, params Params) (string, error) {
	// Validate required parameters
	if err := g.validateParams(params); err != nil {
		return "", fmt.Errorf("invalid parameters: %w", err)
	}

	// Build the payload string according to NT 2025.001
	payload := g.buildPayload(params)

	// Generate hash with CSC
	hash := g.generateHash(payload, params.CSCToken)

	// Build the final URL
	qrURL := g.buildQRURL(params, hash)

	return qrURL, nil
}

// validateParams validates the required parameters
func (g *generator) validateParams(params Params) error {
	if params.ChaveAcesso == "" {
		return fmt.Errorf("chave de acesso é obrigatória")
	}
	if params.TpAmb == "" {
		return fmt.Errorf("tipo de ambiente é obrigatório")
	}
	if params.DhEmi == "" {
		return fmt.Errorf("data de emissão é obrigatória")
	}
	if params.VNF == "" {
		return fmt.Errorf("valor da NF é obrigatório")
	}
	if params.VICMS == "" {
		return fmt.Errorf("valor do ICMS é obrigatório")
	}
	if params.DigVal == "" {
		return fmt.Errorf("digest value é obrigatório")
	}
	if params.CSCID == "" {
		return fmt.Errorf("CSC ID é obrigatório")
	}
	if params.CSCToken == "" {
		return fmt.Errorf("CSC token é obrigatório")
	}
	if params.UF == "" {
		return fmt.Errorf("UF é obrigatória")
	}

	return nil
}

// buildPayload builds the payload string for hash generation
func (g *generator) buildPayload(params Params) string {
	// Format according to NFC-e QR Code v3 specification
	// chNFe|tpAmb|dest|dhEmi|vNF|vICMS|digVal|cIdToken|cHashQRCode

	parts := []string{
		params.ChaveAcesso, // chNFe
		params.TpAmb,       // tpAmb (1=produção, 2=homologação)
	}

	// dest (destinatário) - optional
	if params.Destinatario != "" {
		parts = append(parts, params.Destinatario)
	} else {
		parts = append(parts, "")
	}

	parts = append(parts,
		params.DhEmi,  // dhEmi
		params.VNF,    // vNF
		params.VICMS,  // vICMS
		params.DigVal, // digVal
		params.CSCID,  // cIdToken
	)

	return strings.Join(parts, "|")
}

// generateHash generates SHA-1 hash of payload + CSC token
func (g *generator) generateHash(payload, cscToken string) string {
	// Concatenate payload with CSC token
	data := payload + cscToken

	// Generate SHA-1 hash
	hasher := sha1.New()
	hasher.Write([]byte(data))
	hash := hasher.Sum(nil)

	// Convert to uppercase hex string
	hashStr := fmt.Sprintf("%X", hash)

	return hashStr
}

// buildQRURL builds the final QR Code URL
func (g *generator) buildQRURL(params Params, hash string) string {
	// Base URL depends on UF and environment
	baseURL := g.getBaseURL(params.UF, params.TpAmb)
	if baseURL == "" {
		// Fallback to old URL format if UF not supported
		baseURL = "https://www.nfce.fazenda.sp.gov.br/qrcode"
	}

	// Build query parameters
	values := url.Values{}
	values.Set("chNFe", params.ChaveAcesso)
	values.Set("nVersao", "3") // Version 3
	values.Set("tpAmb", params.TpAmb)
	if params.Destinatario != "" {
		values.Set("dest", params.Destinatario)
	}
	values.Set("dhEmi", url.QueryEscape(params.DhEmi))
	values.Set("vNF", params.VNF)
	values.Set("vICMS", params.VICMS)
	values.Set("digVal", url.QueryEscape(params.DigVal))
	values.Set("cIdToken", params.CSCID)
	values.Set("cHashQRCode", hash)

	// Construct final URL
	fullURL := fmt.Sprintf("%s?%s", baseURL, values.Encode())

	return fullURL
}

// getBaseURL returns the base URL for QR Code according to UF and environment
func (g *generator) getBaseURL(uf, tpAmb string) string {
	// Environment: 1=produção, 2=homologação
	isProduction := tpAmb == "1"

	// URLs by UF for production and homologation
	ufURLs := map[string]map[string]string{
		"AC": {"prod": "https://www.sefaznet.ac.gov.br/nfce/qrcode", "hom": "https://www.sefaznet.ac.gov.br/nfce/qrcode"},
		"AL": {"prod": "https://nfce.sefaz.al.gov.br/QRCode/consultarNFCe.jsp", "hom": "https://nfce.sefaz.al.gov.br/QRCode/consultarNFCe.jsp"},
		"AP": {"prod": "https://www.sefaz.ap.gov.br/nfce/nfcep.php", "hom": "https://www.sefaz.ap.gov.br/nfce/nfcep.php"},
		"AM": {"prod": "https://www.sefaz.am.gov.br/nfce/qrcode", "hom": "https://www.sefaz.am.gov.br/nfce/qrcode"},
		"BA": {"prod": "https://nfce.sefaz.ba.gov.br/servicos/nfce/default.aspx", "hom": "https://nfce.sefaz.ba.gov.br/servicos/nfce/default.aspx"},
		"CE": {"prod": "https://nfce.sefaz.ce.gov.br/pages/ShowNFCe.html", "hom": "https://nfce.sefaz.ce.gov.br/pages/ShowNFCe.html"},
		"DF": {"prod": "https://www.fazenda.df.gov.br/nfce/qrcode", "hom": "https://www.fazenda.df.gov.br/nfce/qrcode"},
		"ES": {"prod": "https://www.sefaz.es.gov.br/nfce/qrcode", "hom": "https://www.sefaz.es.gov.br/nfce/qrcode"},
		"GO": {"prod": "https://nfce.sefaz.go.gov.br/nfce/qrcode", "hom": "https://nfce.sefaz.go.gov.br/nfce/qrcode"},
		"MA": {"prod": "https://www.sefaz.ma.gov.br/nfce/qrcode", "hom": "https://www.sefaz.ma.gov.br/nfce/qrcode"},
		"MT": {"prod": "https://www.sefaz.mt.gov.br/nfce/qrcode", "hom": "https://www.sefaz.mt.gov.br/nfce/qrcode"},
		"MS": {"prod": "https://www.dfe.ms.gov.br/nfce/qrcode", "hom": "https://www.dfe.ms.gov.br/nfce/qrcode"},
		"MG": {"prod": "https://nfce.fazenda.mg.gov.br/portalnfce/sistema/qrcode.xhtml", "hom": "https://nfce.fazenda.mg.gov.br/portalnfce/sistema/qrcode.xhtml"},
		"PA": {"prod": "https://www.sefa.pa.gov.br/nfce/qrcode", "hom": "https://www.sefa.pa.gov.br/nfce/qrcode"},
		"PB": {"prod": "https://www.sefaz.pb.gov.br/nfce/qrcode", "hom": "https://www.sefaz.pb.gov.br/nfce/qrcode"},
		"PR": {"prod": "https://www.fazenda.pr.gov.br/nfce/qrcode", "hom": "https://www.fazenda.pr.gov.br/nfce/qrcode"},
		"PE": {"prod": "https://nfce.sefaz.pe.gov.br/nfce/consulta", "hom": "https://nfce.sefaz.pe.gov.br/nfce/consulta"},
		"PI": {"prod": "https://www.sefaz.pi.gov.br/nfce/qrcode", "hom": "https://www.sefaz.pi.gov.br/nfce/qrcode"},
		"RJ": {"prod": "https://www.fazenda.rj.gov.br/nfce/qrcode", "hom": "https://www.fazenda.rj.gov.br/nfce/qrcode"},
		"RN": {"prod": "https://www.sefaz.rn.gov.br/nfce/qrcode", "hom": "https://www.sefaz.rn.gov.br/nfce/qrcode"},
		"RS": {"prod": "https://www.sefaz.rs.gov.br/nfce/qrcode", "hom": "https://www.sefaz.rs.gov.br/nfce/qrcode"},
		"RO": {"prod": "https://www.sefaz.ro.gov.br/nfce/qrcode", "hom": "https://www.sefaz.ro.gov.br/nfce/qrcode"},
		"RR": {"prod": "https://www.sefaz.rr.gov.br/nfce/qrcode", "hom": "https://www.sefaz.rr.gov.br/nfce/qrcode"},
		"SC": {"prod": "https://sat.sef.sc.gov.br/nfce/qrcode", "hom": "https://sat.sef.sc.gov.br/nfce/qrcode"},
		"SP": {"prod": "https://www.nfce.fazenda.sp.gov.br/qrcode", "hom": "https://www.nfce.fazenda.sp.gov.br/qrcode"},
		"SE": {"prod": "https://www.sefaz.se.gov.br/nfce/qrcode", "hom": "https://www.sefaz.se.gov.br/nfce/qrcode"},
		"TO": {"prod": "https://www.sefaz.to.gov.br/nfce/qrcode", "hom": "https://www.sefaz.to.gov.br/nfce/qrcode"},
	}

	env := "hom"
	if isProduction {
		env = "prod"
	}

	if ufData, exists := ufURLs[uf]; exists {
		if url, exists := ufData[env]; exists {
			return url
		}
	}

	return ""
}

// Helper function to convert string to proper format if needed
func formatDecimal(value string) string {
	// Ensure proper decimal formatting (2 decimal places)
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return fmt.Sprintf("%.2f", f)
	}
	return value
}
