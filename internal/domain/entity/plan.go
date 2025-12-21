package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
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

// Plan represents a subscription plan
type Plan struct {
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

// NewPlan creates a new plan with validation
func NewPlan(name, description string, planType PlanType, price float64) (*Plan, error) {
	if name == "" {
		return nil, errors.New("nome do plano é obrigatório")
	}

	if price < 0 {
		return nil, errors.New("preço não pode ser negativo")
	}

	now := time.Now()
	return &Plan{
		ID:              generatePlanID(),
		Name:            name,
		Description:     description,
		Type:            planType,
		BillingCycle:    BillingCycleMonthly, // Default
		Status:          PlanStatusActive,
		Price:           price,
		Currency:        "BRL",
		QuotaType:       QuotaTypeMonthly,
		MaxNFCePerMonth: 100, // Default
		Features: PlanFeatures{
			MaxNFCePerMonth:    100,
			AllowContingency:   true,
			AllowCancellation:  true,
			AllowInutilization: true,
			WebhookSupport:     true,
			PrioritySupport:    false,
			StorageDays:        365,
		},
		SortOrder: 0,
		TrialDays: 0,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// IsActive returns true if the plan is active
func (p *Plan) IsActive() bool {
	return p.Status == PlanStatusActive
}

// HasQuotaLimit returns true if the plan has NFC-e quota limits
func (p *Plan) HasQuotaLimit() bool {
	return p.QuotaType != QuotaTypeUnlimited
}

// GetMaxNFCe returns the maximum NFC-e allowed based on quota type
func (p *Plan) GetMaxNFCe() (int, bool) {
	switch p.QuotaType {
	case QuotaTypeMonthly:
		return p.MaxNFCePerMonth, p.MaxNFCePerMonth > 0
	case QuotaTypePackage:
		return p.MaxNFCeTotal, p.MaxNFCeTotal > 0
	case QuotaTypeUnlimited:
		return 0, false // Unlimited
	default:
		return 0, false
	}
}

// AllowsFeature checks if a specific feature is allowed in this plan
func (p *Plan) AllowsFeature(feature string) bool {
	switch feature {
	case "contingency":
		return p.Features.AllowContingency
	case "cancellation":
		return p.Features.AllowCancellation
	case "inutilization":
		return p.Features.AllowInutilization
	case "webhook":
		return p.Features.WebhookSupport
	case "priority_support":
		return p.Features.PrioritySupport
	default:
		return false
	}
}

// UpdatePricing updates the plan pricing
func (p *Plan) UpdatePricing(price float64, currency string) error {
	if price < 0 {
		return errors.New("preço não pode ser negativo")
	}

	if currency == "" {
		currency = "BRL"
	}

	p.Price = price
	p.Currency = currency
	p.UpdatedAt = time.Now()
	return nil
}

// UpdateQuotas updates the plan quotas
func (p *Plan) UpdateQuotas(quotaType QuotaType, maxMonthly, maxTotal int) error {
	p.QuotaType = quotaType
	p.MaxNFCePerMonth = maxMonthly
	p.MaxNFCeTotal = maxTotal
	p.UpdatedAt = time.Now()
	return nil
}

// Archive marks the plan as archived (soft delete)
func (p *Plan) Archive() {
	p.Status = PlanStatusArchived
	p.UpdatedAt = time.Now()
}

// generatePlanID generates a unique UUID for the plan
func generatePlanID() string {
	return uuid.New().String()
}
