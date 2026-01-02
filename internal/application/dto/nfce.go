package dto

import (
	"time"
)

// RequestStatus represents the lifecycle state of an NFC-e request
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

// EmitNFceRequest represents the request to emit a NFC-e
type EmitNFceRequest struct {
	UF         string      `json:"uf" binding:"required"`
	Ambiente   string      `json:"ambiente" binding:"required,oneof=producao homologacao"`
	Emitente   Emitente    `json:"emitente" binding:"required"`
	Itens      []Item      `json:"itens" binding:"required,min=1"`
	Pagamentos []Payment   `json:"pagamentos" binding:"required,min=1"`
	Options    EmitOptions `json:"options"`
}

// NFceResponse represents the response containing NFC-e data
type NFceResponse struct {
	ID             string        `json:"id"`
	IdempotencyKey string        `json:"idempotency_key"`
	Status         RequestStatus `json:"status"`
	ChaveAcesso    string        `json:"chave_acesso,omitempty"`
	Protocolo      string        `json:"protocolo,omitempty"`
	RejectionCode  string        `json:"rejection_code,omitempty"`
	RejectionMsg   string        `json:"rejection_msg,omitempty"`
	RetryCount     int           `json:"retry_count,omitempty"`
	NextRetryAt    *time.Time    `json:"next_retry_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Links          NFceLinks     `json:"links,omitempty"`
}

// NFceLinks contains URLs to NFC-e resources
type NFceLinks struct {
	XML    string `json:"xml,omitempty"`
	PDF    string `json:"pdf,omitempty"`
	QrCode string `json:"qr_code,omitempty"`
}

// NFceListResponse represents a list of NFC-e requests
type NFceListResponse struct {
	NFces []NFceResponse `json:"nfces"`
	Total int            `json:"total"`
}

// CancelNFceRequest represents the request to cancel a NFC-e
type CancelNFceRequest struct {
	Justificativa string `json:"justificativa" binding:"required,min=15,max=255"`
}

// NFceEventResponse represents an event in NFC-e lifecycle
type NFceEventResponse struct {
	ID         string        `json:"id"`
	RequestID  string        `json:"request_id"`
	StatusFrom RequestStatus `json:"status_from"`
	StatusTo   RequestStatus `json:"status_to"`
	CStat      string        `json:"cstat,omitempty"`
	Message    string        `json:"message,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

// NFceEventListResponse represents a list of NFC-e events
type NFceEventListResponse struct {
	Events []NFceEventResponse `json:"events"`
	Total  int                 `json:"total"`
}
