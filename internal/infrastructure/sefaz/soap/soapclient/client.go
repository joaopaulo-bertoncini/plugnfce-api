package soapclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// AuthorizationRequest is the input for SEFAZ authorization.
type AuthorizationRequest struct {
	UF              string
	Ambiente        string
	XML             []byte
	Contingency     bool   // Whether to use contingency mode
	ContingencyType string // "SVC-AN" or "SVC-RS"
}

// AuthorizationResponse captures the SEFAZ reply.
type AuthorizationResponse struct {
	Status      string
	CStat       string
	Motivo      string
	Protocolo   string
	RawResponse []byte
}

// Client abstracts SOAP communication with SEFAZ.
type Client interface {
	Authorize(ctx context.Context, req AuthorizationRequest) (AuthorizationResponse, error)
	QueryStatus(ctx context.Context, uf, ambiente string) (AuthorizationResponse, error)
}

// soapClient implements Client interface
type soapClient struct {
	httpClient *http.Client
	endpoints  map[string]map[string]string // UF -> Ambiente -> URL
	timeout    time.Duration
}

// NewSOAPClient creates a new SOAP client for SEFAZ communication
func NewSOAPClient(timeout time.Duration) Client {
	return &soapClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		endpoints: getSEFAZEndpoints(),
		timeout:   timeout,
	}
}

// Authorize sends NFC-e authorization request to SEFAZ
func (c *soapClient) Authorize(ctx context.Context, req AuthorizationRequest) (AuthorizationResponse, error) {
	var endpoint string
	var err error

	if req.Contingency {
		endpoint, err = c.getContingencyEndpoint(req.ContingencyType, req.Ambiente)
		if err != nil {
			return AuthorizationResponse{}, fmt.Errorf("failed to get contingency endpoint: %w", err)
		}
	} else {
		endpoint, err = c.getEndpoint(req.UF, req.Ambiente)
		if err != nil {
			return AuthorizationResponse{}, fmt.Errorf("failed to get endpoint: %w", err)
		}
	}

	// Build SOAP envelope
	soapEnvelope := c.buildAuthorizationEnvelope(req.XML)

	// Send SOAP request
	resp, err := c.sendSOAPRequest(ctx, endpoint, soapEnvelope)
	if err != nil {
		return AuthorizationResponse{}, fmt.Errorf("SOAP request failed: %w", err)
	}

	// Parse response
	return c.parseAuthorizationResponse(resp)
}

// QueryStatus queries SEFAZ service status
func (c *soapClient) QueryStatus(ctx context.Context, uf, ambiente string) (AuthorizationResponse, error) {
	endpoint, err := c.getEndpoint(uf, ambiente)
	if err != nil {
		return AuthorizationResponse{}, fmt.Errorf("failed to get endpoint: %w", err)
	}

	// Build status query SOAP envelope
	soapEnvelope := c.buildStatusQueryEnvelope()

	// Send SOAP request
	resp, err := c.sendSOAPRequest(ctx, endpoint, soapEnvelope)
	if err != nil {
		return AuthorizationResponse{}, fmt.Errorf("SOAP request failed: %w", err)
	}

	// Parse response
	return c.parseStatusResponse(resp)
}

// sendSOAPRequest sends a SOAP request to the specified endpoint
func (c *soapClient) sendSOAPRequest(ctx context.Context, endpoint, soapEnvelope string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(soapEnvelope))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// buildAuthorizationEnvelope builds SOAP envelope for NFC-e authorization
func (c *soapClient) buildAuthorizationEnvelope(xmlContent []byte) string {
	envelope := `<?xml version="1.0" encoding="UTF-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	<soap12:Header>
		<nfeCabecMsg xmlns="http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4">
			<cUF>35</cUF>
			<versaoDados>4.00</versaoDados>
		</nfeCabecMsg>
	</soap12:Header>
	<soap12:Body>
		<nfeDadosMsg xmlns="http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4">
			<NFeAutorizacaoLote xmlns="http://www.portalfiscal.inf.br/nfe">
				<idLote>1</idLote>
				<indSinc>1</indSinc>
				<NFes>
					<NFe>
						<infNFe versao="4.00">
							<!-- NFC-e content will be inserted here -->
						</infNFe>
					</NFe>
				</NFes>
			</NFeAutorizacaoLote>
		</nfeDadosMsg>
	</soap12:Body>
</soap12:Envelope>`

	// Insert the XML content into the envelope
	// This is a simplified approach - in production, proper XML manipulation should be used
	xmlStr := string(xmlContent)
	envelope = strings.Replace(envelope, "<!-- NFC-e content will be inserted here -->", xmlStr, 1)

	return envelope
}

// buildStatusQueryEnvelope builds SOAP envelope for status query
func (c *soapClient) buildStatusQueryEnvelope() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
	<soap12:Header>
		<nfeCabecMsg xmlns="http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4">
			<cUF>35</cUF>
			<versaoDados>4.00</versaoDados>
		</nfeCabecMsg>
	</soap12:Header>
	<soap12:Body>
		<nfeDadosMsg xmlns="http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4">
			<consStatServ versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">
				<tpAmb>2</tpAmb>
				<cServ>SP</cServ>
			</consStatServ>
		</nfeDadosMsg>
	</soap12:Body>
</soap12:Envelope>`
}

// parseAuthorizationResponse parses the SOAP response for authorization
func (c *soapClient) parseAuthorizationResponse(soapResponse []byte) (AuthorizationResponse, error) {
	// This is a simplified parser - in production, use proper XML parsing
	response := AuthorizationResponse{
		RawResponse: soapResponse,
	}

	// Extract cStat
	if idx := bytes.Index(soapResponse, []byte("<cStat>")); idx != -1 {
		start := idx + 7
		if end := bytes.Index(soapResponse[start:], []byte("</cStat>")); end != -1 {
			response.CStat = string(soapResponse[start : start+end])
		}
	}

	// Extract motivo
	if idx := bytes.Index(soapResponse, []byte("<xMotivo>")); idx != -1 {
		start := idx + 9
		if end := bytes.Index(soapResponse[start:], []byte("</xMotivo>")); end != -1 {
			response.Motivo = string(soapResponse[start : start+end])
		}
	}

	// Extract protocolo
	if idx := bytes.Index(soapResponse, []byte("<nProt>")); idx != -1 {
		start := idx + 7
		if end := bytes.Index(soapResponse[start:], []byte("</nProt>")); end != -1 {
			response.Protocolo = string(soapResponse[start : start+end])
		}
	}

	// Determine status based on cStat
	response.Status = c.determineStatus(response.CStat)

	return response, nil
}

// parseStatusResponse parses the SOAP response for status query
func (c *soapClient) parseStatusResponse(soapResponse []byte) (AuthorizationResponse, error) {
	return c.parseAuthorizationResponse(soapResponse)
}

// determineStatus determines the status based on cStat
func (c *soapClient) determineStatus(cstat string) string {
	switch cstat {
	case "100", "101", "102", "103", "104", "105", "106", "107", "108", "109", "150":
		return "authorized"
	case "110", "111", "112", "113", "114", "115", "116", "117", "118", "119":
		return "denied"
	default:
		return "error"
	}
}

// IsRetryableError determines if an error is retryable based on cStat
func IsRetryableError(cstat string) bool {
	// SEFAZ error codes that indicate transient errors (should be retried)
	retryableCodes := map[string]bool{
		// Service temporarily unavailable
		"108": true, // Serviço Paralisado Temporariamente (SVC)
		"109": true, // Serviço Paralisado sem Previsão

		// Network/timeout related
		"500": true, // Erro interno do servidor
		"503": true, // Serviço temporariamente indisponível

		// Processing errors that might succeed on retry
		"301": true, // Uso Denegado: Irregularidade fiscal do emitente (temporary block)
		"302": true, // Irregularidade fiscal do destinatário (temporary)

		// Contingency situations
		"691": true, // Contingência EPEC: Sistema não autorizado
		"692": true, // Contingência SVC: Sistema não autorizado
		"693": true, // Contingência SVC: Autorização não concedida

		// Processing queue full or busy
		"204": true, // Duplicidade de NF-e (might be timing issue)
		"539": true, // Duplicidade de NF-e com diferença na Chave de Acesso
	}

	// Range-based checks for certain error categories
	if cstat >= "500" && cstat <= "599" {
		return true // 5xx errors are generally retryable
	}

	return retryableCodes[cstat]
}

// GetErrorCategory returns the category of the error for better handling
func GetErrorCategory(cstat string) string {
	switch {
	case cstat == "100" || (cstat >= "101" && cstat <= "109"):
		return "authorized"
	case cstat >= "110" && cstat <= "119":
		return "denied_permanent"
	case cstat >= "200" && cstat <= "299":
		return "denied_business_rule"
	case cstat >= "300" && cstat <= "399":
		return "denied_security"
	case cstat >= "400" && cstat <= "499":
		return "denied_schema"
	case cstat >= "500" && cstat <= "599":
		return "error_server"
	case cstat >= "600" && cstat <= "699":
		return "error_contingency"
	case cstat >= "700" && cstat <= "799":
		return "error_processing"
	default:
		return "error_unknown"
	}
}

// getEndpoint returns the SEFAZ endpoint for the given UF and environment
func (c *soapClient) getEndpoint(uf, ambiente string) (string, error) {
	ufMap, exists := c.endpoints[uf]
	if !exists {
		return "", fmt.Errorf("UF %s not supported", uf)
	}

	env := "prod"
	if ambiente == "2" || ambiente == "homologacao" {
		env = "hom"
	}

	endpoint, exists := ufMap[env]
	if !exists {
		return "", fmt.Errorf("environment %s not supported for UF %s", ambiente, uf)
	}

	return endpoint, nil
}

// getContingencyEndpoint returns the contingency endpoint for SVC-AN or SVC-RS
func (c *soapClient) getContingencyEndpoint(contingencyType, ambiente string) (string, error) {
	env := "prod"
	if ambiente == "2" || ambiente == "homologacao" {
		env = "hom"
	}

	switch contingencyType {
	case "SVC-AN":
		// SVC-AN (Sistema Virtual de Contingência - Ambiente Nacional)
		if env == "prod" {
			return "https://www.svc.fazenda.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx", nil
		}
		return "https://hom.svc.fazenda.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx", nil

	case "SVC-RS":
		// SVC-RS (Sistema Virtual de Contingência - Rio Grande do Sul)
		if env == "prod" {
			return "https://www.svrs.rs.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx", nil
		}
		return "https://hom.svrs.rs.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx", nil

	default:
		return "", fmt.Errorf("unsupported contingency type: %s", contingencyType)
	}
}

// getSEFAZEndpoints returns the SEFAZ endpoints for each UF and environment
func getSEFAZEndpoints() map[string]map[string]string {
	return map[string]map[string]string{
		"AC": {
			"prod": "https://www.sefaznet.ac.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://www.sefaznet.ac.gov.br/nfce/NFeAutorizacao4",
		},
		"AL": {
			"prod": "https://nfce.sefaz.al.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.al.gov.br/nfce/NFeAutorizacao4",
		},
		"AP": {
			"prod": "https://nfce.sefaz.ap.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.ap.gov.br/nfce/NFeAutorizacao4",
		},
		"AM": {
			"prod": "https://nfce.sefaz.am.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.am.gov.br/nfce/NFeAutorizacao4",
		},
		"BA": {
			"prod": "https://nfce.sefaz.ba.gov.br/webservices/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.ba.gov.br/webservices/NFeAutorizacao4",
		},
		"CE": {
			"prod": "https://nfce.sefaz.ce.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.ce.gov.br/nfce/NFeAutorizacao4",
		},
		"DF": {
			"prod": "https://www.nfce.fazenda.df.gov.br/NFeAutorizacao4",
			"hom":  "https://www.nfce.fazenda.df.gov.br/NFeAutorizacao4",
		},
		"ES": {
			"prod": "https://nfce.sefaz.es.gov.br/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.es.gov.br/NFeAutorizacao4",
		},
		"GO": {
			"prod": "https://nfce.sefaz.go.gov.br/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.go.gov.br/NFeAutorizacao4",
		},
		"MA": {
			"prod": "https://nfce.sefaz.ma.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.ma.gov.br/nfce/NFeAutorizacao4",
		},
		"MT": {
			"prod": "https://nfce.sefaz.mt.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.mt.gov.br/nfce/NFeAutorizacao4",
		},
		"MS": {
			"prod": "https://nfce.sefaz.ms.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.ms.gov.br/nfce/NFeAutorizacao4",
		},
		"MG": {
			"prod": "https://nfce.fazenda.mg.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.fazenda.mg.gov.br/nfce/NFeAutorizacao4",
		},
		"PA": {
			"prod": "https://nfce.sefa.pa.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefa.pa.gov.br/nfce/NFeAutorizacao4",
		},
		"PB": {
			"prod": "https://nfce.sefaz.pb.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.pb.gov.br/nfce/NFeAutorizacao4",
		},
		"PR": {
			"prod": "https://nfce.sefaz.pr.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.pr.gov.br/nfce/NFeAutorizacao4",
		},
		"PE": {
			"prod": "https://nfce.sefaz.pe.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.pe.gov.br/nfce/NFeAutorizacao4",
		},
		"PI": {
			"prod": "https://nfce.sefaz.pi.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.pi.gov.br/nfce/NFeAutorizacao4",
		},
		"RJ": {
			"prod": "https://nfce.sefaz.rj.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.rj.gov.br/nfce/NFeAutorizacao4",
		},
		"RN": {
			"prod": "https://nfce.sefaz.rn.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.rn.gov.br/nfce/NFeAutorizacao4",
		},
		"RS": {
			"prod": "https://nfce.sefaz.rs.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.rs.gov.br/nfce/NFeAutorizacao4",
		},
		"RO": {
			"prod": "https://nfce.sefaz.ro.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.ro.gov.br/nfce/NFeAutorizacao4",
		},
		"RR": {
			"prod": "https://nfce.sefaz.rr.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.rr.gov.br/nfce/NFeAutorizacao4",
		},
		"SC": {
			"prod": "https://nfce.sefaz.sc.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.sc.gov.br/nfce/NFeAutorizacao4",
		},
		"SP": {
			"prod": "https://nfce.fazenda.sp.gov.br/NFeAutorizacao4",
			"hom":  "https://nfce.fazenda.sp.gov.br/NFeAutorizacao4",
		},
		"SE": {
			"prod": "https://nfce.sefaz.se.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.se.gov.br/nfce/NFeAutorizacao4",
		},
		"TO": {
			"prod": "https://nfce.sefaz.to.gov.br/nfce/NFeAutorizacao4",
			"hom":  "https://nfce.sefaz.to.gov.br/nfce/NFeAutorizacao4",
		},
	}
}
