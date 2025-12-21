package dto

import (
	"time"
)

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusTrial     SubscriptionStatus = "trial"
	SubscriptionStatusSuspended SubscriptionStatus = "suspended"
	SubscriptionStatusCanceled  SubscriptionStatus = "canceled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
)

// UsageStats tracks the usage of NFC-e within a billing period
type UsageStats struct {
	PeriodStart   time.Time  `json:"period_start"`
	PeriodEnd     time.Time  `json:"period_end"`
	NFCeIssued    int        `json:"nfce_issued"`
	NFCeRemaining int        `json:"nfce_remaining"` // -1 = unlimited
	LastNFCeAt    *time.Time `json:"last_nfce_at,omitempty"`
}

// BillingInfo contains billing-related information
type BillingInfo struct {
	NextBillingAt time.Time  `json:"next_billing_at"`
	LastBilledAt  *time.Time `json:"last_billed_at,omitempty"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	PaymentMethod string     `json:"payment_method,omitempty"`
}

// SubscriptionDTO represents a company's subscription to a plan
type SubscriptionDTO struct {
	ID        string             `json:"id"`
	CompanyID string             `json:"company_id"`
	PlanID    string             `json:"plan_id"`
	Status    SubscriptionStatus `json:"status"`

	// Period
	StartedAt   time.Time  `json:"started_at"`
	EndsAt      *time.Time `json:"ends_at,omitempty"` // For trial or fixed periods
	CanceledAt  *time.Time `json:"canceled_at,omitempty"`
	SuspendedAt *time.Time `json:"suspended_at,omitempty"`

	// Trial
	IsTrial     bool       `json:"is_trial,omitempty"`
	TrialEndsAt *time.Time `json:"trial_ends_at,omitempty"`

	// Usage and quotas
	CurrentUsage UsageStats  `json:"current_usage"`
	BillingInfo  BillingInfo `json:"billing_info"`

	// Metadata
	AutoRenew    bool      `json:"auto_renew"`
	CancelReason string    `json:"cancel_reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// References (populated when needed)
	Company *CompanyDTO `json:"company,omitempty"`
	Plan    *PlanDTO    `json:"plan,omitempty"`
}

// CreateSubscriptionRequest represents the request to create a new subscription
type CreateSubscriptionRequest struct {
	CompanyID string `json:"company_id" validate:"required"`
	PlanID    string `json:"plan_id" validate:"required"`
}

// UpdateSubscriptionRequest represents the request to update a subscription
type UpdateSubscriptionRequest struct {
	Status       *SubscriptionStatus `json:"status,omitempty"`
	AutoRenew    *bool               `json:"auto_renew,omitempty"`
	CancelReason *string             `json:"cancel_reason,omitempty"`
}

// CancelSubscriptionRequest represents the request to cancel a subscription
type CancelSubscriptionRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// SubscriptionListResponse represents a paginated list of subscriptions
type SubscriptionListResponse struct {
	Subscriptions []SubscriptionDTO `json:"subscriptions"`
	Total         int               `json:"total"`
}
