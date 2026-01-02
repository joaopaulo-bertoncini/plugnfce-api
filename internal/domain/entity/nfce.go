package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// RequestStatus represents the lifecycle state of an NFC-e request.
type RequestStatus string

const (
	// RequestStatusPending is persisted right after intake.
	RequestStatusPending RequestStatus = "pending"
	// RequestStatusProcessing is set when the worker starts handling the job.
	RequestStatusProcessing RequestStatus = "processing"
	// RequestStatusAuthorized means SEFAZ authorized the NFC-e.
	RequestStatusAuthorized RequestStatus = "authorized"
	// RequestStatusRejected means SEFAZ rejected the NFC-e with a business rule.
	RequestStatusRejected RequestStatus = "rejected"
	// RequestStatusContingency is used when falling back to SVC-AN/SVC-RS.
	RequestStatusContingency RequestStatus = "contingency"
	// RequestStatusRetrying indicates the message is being re-enqueued.
	RequestStatusRetrying RequestStatus = "retrying"
	// RequestStatusCanceled is for cancellation events.
	RequestStatusCanceled RequestStatus = "canceled"
)

// EmitOptions controls sync/async behavior and contingency flags.
type EmitOptions struct {
	Contingencia bool `json:"contingencia"`
	Sync         bool `json:"sync"`
}

// Certificate holds the encrypted PFX and its password.
type Certificate struct {
	PFXBase64 string `json:"cert_pfx_b64"`
	Password  string `json:"cert_password"`
}

// Emitente aggregates issuer data required to build the XML and QR.
type Emitente struct {
	CNPJ     string `json:"cnpj"`
	IE       string `json:"ie,omitempty"`
	Regime   string `json:"regime"`
	CSCID    string `json:"csc_id"`
	CSCToken string `json:"csc_token"`
}

// Item is a minimal representation of a product line.
type Item struct {
	Descricao  string  `json:"descricao"`
	NCM        string  `json:"ncm"`
	CFOP       string  `json:"cfop"`
	GTIN       string  `json:"gtin,omitempty"`
	Valor      float64 `json:"valor"`
	Quantidade float64 `json:"quantidade"`
	Unidade    string  `json:"unidade"`
}

// Payment captures the payment mix used in the sale.
type Payment struct {
	Forma string  `json:"forma"`
	Valor float64 `json:"valor"`
	Troco float64 `json:"troco,omitempty"`
}

// EmitPayload is the normalized payload used to generate the NFC-e XML.
type EmitPayload struct {
	UF         string      `json:"uf"`
	Ambiente   string      `json:"ambiente"`
	Emitente   Emitente    `json:"emitente"`
	Itens      []Item      `json:"itens"`
	Pagamentos []Payment   `json:"pagamentos"`
	Options    EmitOptions `json:"options"`
}

// Value implements the driver.Valuer interface for GORM JSONB serialization
func (e EmitPayload) Value() (driver.Value, error) {
	return json.Marshal(e)
}

// Scan implements the sql.Scanner interface for GORM JSONB deserialization
func (e *EmitPayload) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("EmitPayload.Scan: value must be []byte")
	}

	return json.Unmarshal(bytes, e)
}

// NFCE represents an NFC-e document and its processing state
type NFCE struct {
	ID             string        `json:"id"`
	CompanyID      string        `json:"company_id"` // Reference to issuing company
	IdempotencyKey string        `json:"idempotency_key"`
	Status         RequestStatus `json:"status"`

	// NFC-e data
	Payload EmitPayload `json:"payload" gorm:"type:jsonb"`

	// SEFAZ response data
	ChaveAcesso string `json:"chave_acesso,omitempty"`
	Protocolo   string `json:"protocolo,omitempty"`
	Numero      string `json:"numero,omitempty"` // NFC-e number
	Serie       string `json:"serie,omitempty"`  // NFC-e series

	// Error handling
	RejectionCode string `json:"rejection_code,omitempty" gorm:"column:rejection_code"`
	RejectionMsg  string `json:"rejection_msg,omitempty" gorm:"column:rejection_msg"`
	CStat         string `json:"cstat,omitempty" gorm:"column:cstat"`     // SEFAZ status code
	XMotivo       string `json:"xmotivo,omitempty" gorm:"column:xmotivo"` // SEFAZ status message

	// Processing metadata
	RetryCount   int        `json:"retry_count,omitempty"`
	NextRetryAt  *time.Time `json:"next_retry_at,omitempty"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	AuthorizedAt *time.Time `json:"authorized_at,omitempty"`

	// Contingency
	InContingency   bool   `json:"in_contingency,omitempty"`
	ContingencyType string `json:"contingency_type,omitempty"` // SVC-AN, SVC-RS

	// Storage references
	XMLURL    string `json:"xml_url,omitempty" gorm:"column:xml_url"`       // S3 URL for XML
	PDFURL    string `json:"pdf_url,omitempty" gorm:"column:pdf_url"`       // S3 URL for DANFE
	QRCodeURL string `json:"qrcode_url,omitempty" gorm:"column:qrcode_url"` // QR Code image URL

	// Relationships (not serialized to JSON)
	Events []Event `json:"-" gorm:"foreignKey:RequestID;references:ID"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewNFCE creates a new NFC-e request
func NewNFCE(companyID, idempotencyKey string, payload EmitPayload) (*NFCE, error) {
	if companyID == "" {
		return nil, errors.New("company ID é obrigatório")
	}

	if idempotencyKey == "" {
		return nil, errors.New("chave de idempotência é obrigatória")
	}

	now := time.Now()
	return &NFCE{
		ID:             generateNFCEID(),
		CompanyID:      companyID,
		IdempotencyKey: idempotencyKey,
		Status:         RequestStatusPending,
		Payload:        payload,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// MarkAsProcessing marks the NFC-e as being processed
func (n *NFCE) MarkAsProcessing() {
	n.Status = RequestStatusProcessing
	n.UpdatedAt = time.Now()
}

// MarkAsAuthorized marks the NFC-e as authorized by SEFAZ
func (n *NFCE) MarkAsAuthorized(chaveAcesso, protocolo, numero, serie string) {
	now := time.Now()
	n.Status = RequestStatusAuthorized
	n.ChaveAcesso = chaveAcesso
	n.Protocolo = protocolo
	n.Numero = numero
	n.Serie = serie
	n.AuthorizedAt = &now
	n.ProcessedAt = &now
	n.UpdatedAt = now
}

// MarkAsRejected marks the NFC-e as rejected by SEFAZ
func (n *NFCE) MarkAsRejected(cstat, xmotivo string) {
	now := time.Now()
	n.Status = RequestStatusRejected
	n.CStat = cstat
	n.XMotivo = xmotivo
	n.RejectionCode = cstat
	n.RejectionMsg = xmotivo
	n.ProcessedAt = &now
	n.UpdatedAt = now
}

// MarkAsContingency marks the NFC-e as using contingency
func (n *NFCE) MarkAsContingency(contingencyType string) {
	n.Status = RequestStatusContingency
	n.InContingency = true
	n.ContingencyType = contingencyType
	n.UpdatedAt = time.Now()
}

// IncrementRetry increments the retry count
func (n *NFCE) IncrementRetry() {
	n.RetryCount++
	n.Status = RequestStatusRetrying
	n.UpdatedAt = time.Now()
}

// CanRetry checks if the NFC-e can be retried
func (n *NFCE) CanRetry(maxRetries int) bool {
	// Don't retry if already successful or canceled
	if n.Status == RequestStatusAuthorized || n.Status == RequestStatusCanceled {
		return false
	}

	// Don't retry if exceeded max attempts
	if n.RetryCount >= maxRetries {
		return false
	}

	// Don't retry if more than 48 hours have passed since creation
	maxAge := 48 * time.Hour
	if time.Since(n.CreatedAt) > maxAge {
		return false
	}

	return true
}

// SetStorageURLs sets the URLs for stored documents
func (n *NFCE) SetStorageURLs(xmlURL, pdfURL, qrCodeURL string) {
	n.XMLURL = xmlURL
	n.PDFURL = pdfURL
	n.QRCodeURL = qrCodeURL
	n.UpdatedAt = time.Now()
}

// Event captures status transitions for auditability and observability.
type Event struct {
	ID         string                 `json:"id" gorm:"type:varchar(36);primaryKey"`
	RequestID  string                 `json:"request_id" gorm:"type:varchar(36);index"`
	StatusFrom RequestStatus          `json:"status_from" gorm:"type:varchar(20)"`
	StatusTo   RequestStatus          `json:"status_to" gorm:"type:varchar(20)"`
	CStat      string                 `json:"cstat,omitempty" gorm:"type:varchar(10)"`
	Message    string                 `json:"message,omitempty" gorm:"type:text"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for GORM
func (NFCE) TableName() string {
	return "nfce_requests"
}

// Request represents an NFC-e emission request (alias for NFCE for backward compatibility)
type Request = NFCE

// generateNFCEID generates a unique UUID for NFC-e
func generateNFCEID() string {
	return uuid.New().String()
}
