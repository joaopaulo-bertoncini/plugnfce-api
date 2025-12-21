package dto

import (
	"time"
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

// WebhookDTO represents a webhook configuration for notifications
type WebhookDTO struct {
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

// CreateWebhookRequest represents the request to create a new webhook
type CreateWebhookRequest struct {
	CompanyID   string              `json:"company_id" validate:"required"`
	Name        string              `json:"name" validate:"required"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url" validate:"required,url"`
	Method      HTTPMethod          `json:"method,omitempty"`
	Events      []WebhookEvent      `json:"events" validate:"required,min=1"`
	Headers     WebhookHeaders      `json:"headers,omitempty"`
	Secret      string              `json:"secret,omitempty"`
	RetryConfig *WebhookRetryConfig `json:"retry_config,omitempty"`
}

// UpdateWebhookRequest represents the request to update a webhook
type UpdateWebhookRequest struct {
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	URL         *string             `json:"url,omitempty"`
	Method      *HTTPMethod         `json:"method,omitempty"`
	Status      *WebhookStatus      `json:"status,omitempty"`
	Events      []WebhookEvent      `json:"events,omitempty"`
	Headers     WebhookHeaders      `json:"headers,omitempty"`
	Secret      *string             `json:"secret,omitempty"`
	RetryConfig *WebhookRetryConfig `json:"retry_config,omitempty"`
}

// WebhookListResponse represents a paginated list of webhooks
type WebhookListResponse struct {
	Webhooks []WebhookDTO `json:"webhooks"`
	Total    int          `json:"total"`
}

// WebhookDeliveryListResponse represents a list of webhook deliveries
type WebhookDeliveryListResponse struct {
	Deliveries []WebhookDelivery `json:"deliveries"`
	Total      int               `json:"total"`
}
