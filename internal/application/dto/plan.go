package dto

import (
	"time"
)

// PlanType represents the type of plan
type PlanType string

const (
	PlanTypeMonthly PlanType = "monthly" // Recurring monthly
	PlanTypeYearly  PlanType = "yearly"  // Recurring yearly
	PlanTypePackage PlanType = "package" // One-time package
)

// BillingCycle represents the billing cycle
type BillingCycle string

const (
	BillingCycleMonthly BillingCycle = "monthly"
	BillingCycleYearly  BillingCycle = "yearly"
	BillingCycleOnce    BillingCycle = "once"
)

// PlanStatus represents the status of a plan
type PlanStatus string

const (
	PlanStatusActive   PlanStatus = "active"
	PlanStatusInactive PlanStatus = "inactive"
	PlanStatusArchived PlanStatus = "archived"
)

// QuotaType represents how quotas are calculated
type QuotaType string

const (
	QuotaTypeMonthly   QuotaType = "monthly"   // X NFC-e per month
	QuotaTypePackage   QuotaType = "package"   // X NFC-e total in package
	QuotaTypeUnlimited QuotaType = "unlimited" // Unlimited NFC-e
)

// PlanFeatures represents the features included in a plan
type PlanFeatures struct {
	MaxNFCePerMonth    int  `json:"max_nfce_per_month,omitempty"` // 0 = unlimited
	MaxNFCeTotal       int  `json:"max_nfce_total,omitempty"`     // 0 = unlimited
	AllowContingency   bool `json:"allow_contingency"`
	AllowCancellation  bool `json:"allow_cancellation"`
	AllowInutilization bool `json:"allow_inutilization"`
	WebhookSupport     bool `json:"webhook_support"`
	PrioritySupport    bool `json:"priority_support"`
	StorageDays        int  `json:"storage_days"` // Days to keep XML/PDF
}

// PlanDTO represents a subscription plan
type PlanDTO struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Type         PlanType     `json:"type"`
	BillingCycle BillingCycle `json:"billing_cycle"`
	Status       PlanStatus   `json:"status"`

	// Pricing
	Price    float64 `json:"price"`    // Price per billing cycle
	Currency string  `json:"currency"` // Default: BRL

	// Quotas
	QuotaType       QuotaType `json:"quota_type"`
	MaxNFCePerMonth int       `json:"max_nfce_per_month,omitempty"` // For monthly quotas
	MaxNFCeTotal    int       `json:"max_nfce_total,omitempty"`     // For package quotas

	// Features
	Features PlanFeatures `json:"features"`

	// Metadata
	IsPopular bool      `json:"is_popular,omitempty"` // Highlight in UI
	SortOrder int       `json:"sort_order,omitempty"` // Display order
	TrialDays int       `json:"trial_days,omitempty"` // Trial period in days
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreatePlanRequest represents the request to create a new plan
type CreatePlanRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description,omitempty"`
	Type        PlanType `json:"type" validate:"required"`
	Price       float64  `json:"price" validate:"required,min=0"`
}

// UpdatePlanRequest represents the request to update a plan
type UpdatePlanRequest struct {
	Name            *string       `json:"name,omitempty"`
	Description     *string       `json:"description,omitempty"`
	Type            *PlanType     `json:"type,omitempty"`
	Status          *PlanStatus   `json:"status,omitempty"`
	Price           *float64      `json:"price,omitempty"`
	Currency        *string       `json:"currency,omitempty"`
	QuotaType       *QuotaType    `json:"quota_type,omitempty"`
	MaxNFCePerMonth *int          `json:"max_nfce_per_month,omitempty"`
	MaxNFCeTotal    *int          `json:"max_nfce_total,omitempty"`
	Features        *PlanFeatures `json:"features,omitempty"`
	IsPopular       *bool         `json:"is_popular,omitempty"`
	SortOrder       *int          `json:"sort_order,omitempty"`
	TrialDays       *int          `json:"trial_days,omitempty"`
}

// PlanListResponse represents a paginated list of plans
type PlanListResponse struct {
	Plans []PlanDTO `json:"plans"`
	Total int       `json:"total"`
}
