package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
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

// Subscription represents a company's subscription to a plan
type Subscription struct {
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
	Company *Company `json:"company,omitempty"`
	Plan    *Plan    `json:"plan,omitempty"`
}

// NewSubscription creates a new subscription
func NewSubscription(companyID, planID string, plan *Plan) (*Subscription, error) {
	if companyID == "" {
		return nil, errors.New("company ID é obrigatório")
	}

	if planID == "" {
		return nil, errors.New("plan ID é obrigatório")
	}

	if plan == nil {
		return nil, errors.New("plan é obrigatório")
	}

	now := time.Now()
	subscription := &Subscription{
		ID:        generateSubscriptionID(),
		CompanyID: companyID,
		PlanID:    planID,
		Status:    SubscriptionStatusActive,
		StartedAt: now,
		AutoRenew: true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Initialize usage stats
	subscription.CurrentUsage = UsageStats{
		PeriodStart:   now,
		PeriodEnd:     subscription.calculatePeriodEnd(now, plan),
		NFCeRemaining: subscription.calculateInitialQuota(plan),
	}

	// Initialize billing info
	subscription.BillingInfo = BillingInfo{
		NextBillingAt: subscription.calculateNextBilling(now, plan),
		Amount:        plan.Price,
		Currency:      plan.Currency,
	}

	// Handle trial period if applicable
	if plan.TrialDays > 0 {
		subscription.Status = SubscriptionStatusTrial
		trialEnd := now.AddDate(0, 0, plan.TrialDays)
		subscription.TrialEndsAt = &trialEnd
		subscription.BillingInfo.NextBillingAt = trialEnd
	}

	return subscription, nil
}

// CanIssueNFCe checks if the subscription allows issuing another NFC-e
func (s *Subscription) CanIssueNFCe() (bool, string) {
	// Check subscription status
	if s.Status == SubscriptionStatusSuspended || s.Status == SubscriptionStatusCanceled || s.Status == SubscriptionStatusExpired {
		return false, "assinatura não está ativa"
	}

	// Check trial expiration
	if s.IsTrial && s.TrialEndsAt != nil && time.Now().After(*s.TrialEndsAt) {
		return false, "período de teste expirou"
	}

	// Check quota
	if s.CurrentUsage.NFCeRemaining == 0 {
		return false, "cota de NFC-e esgotada para o período"
	}

	return true, ""
}

// RecordNFCeUsage records the usage of one NFC-e
func (s *Subscription) RecordNFCeUsage() error {
	now := time.Now()

	// Check if period has changed (for monthly plans)
	if s.needsPeriodReset(now) {
		s.resetUsagePeriod(now)
	}

	// Check quota
	if s.CurrentUsage.NFCeRemaining == 0 {
		return errors.New("cota de NFC-e esgotada para o período")
	}

	// Record usage
	s.CurrentUsage.NFCeIssued++
	if s.CurrentUsage.NFCeRemaining > 0 {
		s.CurrentUsage.NFCeRemaining--
	}
	s.CurrentUsage.LastNFCeAt = &now
	s.UpdatedAt = now

	return nil
}

// GetUsagePercentage returns the usage percentage (0-100)
func (s *Subscription) GetUsagePercentage() float64 {
	if s.CurrentUsage.NFCeRemaining < 0 { // Unlimited
		return 0
	}

	totalQuota := s.CurrentUsage.NFCeIssued + s.CurrentUsage.NFCeRemaining
	if totalQuota == 0 {
		return 0
	}

	return float64(s.CurrentUsage.NFCeIssued) / float64(totalQuota) * 100
}

// Cancel cancels the subscription
func (s *Subscription) Cancel(reason string) {
	now := time.Now()
	s.Status = SubscriptionStatusCanceled
	s.CanceledAt = &now
	s.CancelReason = reason
	s.AutoRenew = false
	s.UpdatedAt = now
}

// Suspend suspends the subscription
func (s *Subscription) Suspend() {
	now := time.Now()
	s.Status = SubscriptionStatusSuspended
	s.SuspendedAt = &now
	s.UpdatedAt = now
}

// Reactivate reactivates a suspended subscription
func (s *Subscription) Reactivate() error {
	if s.Status != SubscriptionStatusSuspended {
		return errors.New("apenas assinaturas suspensas podem ser reativadas")
	}

	s.Status = SubscriptionStatusActive
	s.SuspendedAt = nil
	s.UpdatedAt = time.Now()
	return nil
}

// IsActive returns true if the subscription is active
func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive || s.Status == SubscriptionStatusTrial
}

// calculatePeriodEnd calculates the end of the current billing period
func (s *Subscription) calculatePeriodEnd(start time.Time, plan *Plan) time.Time {
	switch plan.BillingCycle {
	case BillingCycleMonthly:
		return start.AddDate(0, 1, 0)
	case BillingCycleYearly:
		return start.AddDate(1, 0, 0)
	default:
		return start.AddDate(0, 1, 0) // Default to monthly
	}
}

// calculateNextBilling calculates the next billing date
func (s *Subscription) calculateNextBilling(start time.Time, plan *Plan) time.Time {
	switch plan.BillingCycle {
	case BillingCycleMonthly:
		return start.AddDate(0, 1, 0)
	case BillingCycleYearly:
		return start.AddDate(1, 0, 0)
	default:
		return start.AddDate(0, 1, 0)
	}
}

// calculateInitialQuota calculates the initial quota based on the plan
func (s *Subscription) calculateInitialQuota(plan *Plan) int {
	switch plan.QuotaType {
	case QuotaTypeMonthly:
		return plan.MaxNFCePerMonth
	case QuotaTypePackage:
		return plan.MaxNFCeTotal
	case QuotaTypeUnlimited:
		return -1 // Unlimited
	default:
		return 100 // Default
	}
}

// needsPeriodReset checks if the usage period needs to be reset
func (s *Subscription) needsPeriodReset(now time.Time) bool {
	return now.After(s.CurrentUsage.PeriodEnd)
}

// resetUsagePeriod resets the usage stats for a new period
func (s *Subscription) resetUsagePeriod(now time.Time) {
	s.CurrentUsage.PeriodStart = now
	s.CurrentUsage.PeriodEnd = s.calculatePeriodEnd(now, s.Plan)
	s.CurrentUsage.NFCeIssued = 0
	s.CurrentUsage.NFCeRemaining = s.calculateInitialQuota(s.Plan)
}

// generateSubscriptionID generates a unique UUID for the subscription
func generateSubscriptionID() string {
	return uuid.New().String()
}
