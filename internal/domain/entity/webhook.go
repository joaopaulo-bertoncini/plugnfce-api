package entity

import (
	"errors"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// WebhookEvent represents the types of events that can trigger webhooks
type WebhookEvent string

const (
	WebhookEventNFCEAuthorized      WebhookEvent = "nfce.authorized"
	WebhookEventNFCERejected        WebhookEvent = "nfce.rejected"
	WebhookEventNFCECanceled        WebhookEvent = "nfce.canceled"
	WebhookEventNFCEContingency     WebhookEvent = "nfce.contingency"
	WebhookEventSubscriptionExpired WebhookEvent = "subscription.expired"
	WebhookEventQuotaExceeded       WebhookEvent = "quota.exceeded"
)

// WebhookStatus represents the status of a webhook configuration
type WebhookStatus string

const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusInactive WebhookStatus = "inactive"
	WebhookStatusFailed   WebhookStatus = "failed"
)

// HTTPMethod represents HTTP methods for webhook delivery
type HTTPMethod string

const (
	HTTPMethodPOST  HTTPMethod = "POST"
	HTTPMethodPUT   HTTPMethod = "PUT"
	HTTPMethodPATCH HTTPMethod = "PATCH"
)

// WebhookHeaders contains custom headers for webhook requests
type WebhookHeaders map[string]string

// WebhookRetryConfig contains retry configuration
type WebhookRetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval"` // Base interval between retries
	MaxInterval   time.Duration `json:"max_interval"`   // Maximum interval
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           string                 `json:"id"`
	WebhookID    string                 `json:"webhook_id"`
	Event        WebhookEvent           `json:"event"`
	Payload      map[string]interface{} `json:"payload"`
	Attempt      int                    `json:"attempt"`
	StatusCode   int                    `json:"status_code,omitempty"`
	ResponseBody string                 `json:"response_body,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Succeeded    bool                   `json:"succeeded"`
	DeliveredAt  *time.Time             `json:"delivered_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// Webhook represents a webhook configuration for notifications
type Webhook struct {
	ID          string        `json:"id"`
	CompanyID   string        `json:"company_id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	URL         string        `json:"url"`
	Method      HTTPMethod    `json:"method"`
	Status      WebhookStatus `json:"status"`

	// Events to listen for
	Events []WebhookEvent `json:"events"`

	// Authentication and headers
	Headers WebhookHeaders `json:"headers,omitempty"`
	Secret  string         `json:"secret,omitempty"` // For HMAC validation

	// Retry configuration
	RetryConfig WebhookRetryConfig `json:"retry_config"`

	// Statistics
	TotalDeliveries      int `json:"total_deliveries"`
	SuccessfulDeliveries int `json:"successful_deliveries"`
	FailedDeliveries     int `json:"failed_deliveries"`

	// Metadata
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
}

// NewWebhook creates a new webhook configuration
func NewWebhook(companyID, name, webhookURL string, events []WebhookEvent) (*Webhook, error) {
	if companyID == "" {
		return nil, errors.New("company ID é obrigatório")
	}

	if name == "" {
		return nil, errors.New("nome do webhook é obrigatório")
	}

	if err := validateWebhookURL(webhookURL); err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, errors.New("pelo menos um evento deve ser especificado")
	}

	now := time.Now()
	return &Webhook{
		ID:        generateWebhookID(),
		CompanyID: companyID,
		Name:      name,
		URL:       webhookURL,
		Method:    HTTPMethodPOST,
		Status:    WebhookStatusActive,
		Events:    events,
		Headers:   make(WebhookHeaders),
		RetryConfig: WebhookRetryConfig{
			MaxRetries:    3,
			RetryInterval: 5 * time.Second,
			MaxInterval:   5 * time.Minute,
		},
		TotalDeliveries:      0,
		SuccessfulDeliveries: 0,
		FailedDeliveries:     0,
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

// IsActive returns true if the webhook is active
func (w *Webhook) IsActive() bool {
	return w.Status == WebhookStatusActive
}

// ListensToEvent returns true if the webhook listens to the specified event
func (w *Webhook) ListensToEvent(event WebhookEvent) bool {
	for _, e := range w.Events {
		if e == event {
			return true
		}
	}
	return false
}

// AddEvent adds an event to the webhook's event list
func (w *Webhook) AddEvent(event WebhookEvent) {
	for _, e := range w.Events {
		if e == event {
			return // Already exists
		}
	}
	w.Events = append(w.Events, event)
	w.UpdatedAt = time.Now()
}

// RemoveEvent removes an event from the webhook's event list
func (w *Webhook) RemoveEvent(event WebhookEvent) {
	for i, e := range w.Events {
		if e == event {
			w.Events = append(w.Events[:i], w.Events[i+1:]...)
			w.UpdatedAt = time.Now()
			return
		}
	}
}

// SetHeaders sets custom headers for the webhook
func (w *Webhook) SetHeaders(headers WebhookHeaders) {
	w.Headers = headers
	w.UpdatedAt = time.Now()
}

// SetSecret sets the webhook secret for HMAC validation
func (w *Webhook) SetSecret(secret string) {
	w.Secret = secret
	w.UpdatedAt = time.Now()
}

// RecordDelivery records a webhook delivery attempt
func (w *Webhook) RecordDelivery(success bool) {
	w.TotalDeliveries++
	if success {
		w.SuccessfulDeliveries++
	} else {
		w.FailedDeliveries++
	}
	now := time.Now()
	w.LastDeliveryAt = &now
	w.UpdatedAt = now

	// Auto-disable webhook if too many failures
	if w.getFailureRate() > 0.8 && w.TotalDeliveries > 10 {
		w.Status = WebhookStatusFailed
	}
}

// GetSuccessRate returns the success rate as a percentage (0-100)
func (w *Webhook) GetSuccessRate() float64 {
	if w.TotalDeliveries == 0 {
		return 100.0
	}
	return float64(w.SuccessfulDeliveries) / float64(w.TotalDeliveries) * 100
}

// getFailureRate returns the failure rate (0-1)
func (w *Webhook) getFailureRate() float64 {
	if w.TotalDeliveries == 0 {
		return 0.0
	}
	return float64(w.FailedDeliveries) / float64(w.TotalDeliveries)
}

// Activate activates the webhook
func (w *Webhook) Activate() {
	w.Status = WebhookStatusActive
	w.UpdatedAt = time.Now()
}

// Deactivate deactivates the webhook
func (w *Webhook) Deactivate() {
	w.Status = WebhookStatusInactive
	w.UpdatedAt = time.Now()
}

// validateWebhookURL validates the webhook URL
func validateWebhookURL(webhookURL string) error {
	if webhookURL == "" {
		return errors.New("URL do webhook é obrigatória")
	}

	u, err := url.Parse(webhookURL)
	if err != nil {
		return errors.New("URL do webhook é inválida")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("URL do webhook deve usar HTTP ou HTTPS")
	}

	if u.Host == "" {
		return errors.New("URL do webhook deve ter um host válido")
	}

	return nil
}

// generateWebhookID generates a unique UUID for the webhook
func generateWebhookID() string {
	return uuid.New().String()
}
