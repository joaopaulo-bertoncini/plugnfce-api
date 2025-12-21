-- Create subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'trial', 'suspended', 'canceled', 'expired')),

    -- Period
    started_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    suspended_at TIMESTAMPTZ,

    -- Trial
    is_trial BOOLEAN DEFAULT FALSE,
    trial_ends_at TIMESTAMPTZ,

    -- Usage stats
    usage_period_start TIMESTAMPTZ,
    usage_period_end TIMESTAMPTZ,
    usage_nfce_issued INTEGER DEFAULT 0,
    usage_nfce_remaining INTEGER DEFAULT -1, -- -1 = unlimited
    usage_last_nfce_at TIMESTAMPTZ,

    -- Billing info
    billing_next_billing_at TIMESTAMPTZ,
    billing_last_billed_at TIMESTAMPTZ,
    billing_amount DECIMAL(10,2),
    billing_currency VARCHAR(3) DEFAULT 'BRL',
    billing_payment_method VARCHAR(50),

    -- Metadata
    auto_renew BOOLEAN DEFAULT TRUE,
    cancel_reason TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for subscriptions
CREATE INDEX IF NOT EXISTS idx_subscriptions_company_id ON subscriptions(company_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_started_at ON subscriptions(started_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_ends_at ON subscriptions(ends_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_trial_ends_at ON subscriptions(trial_ends_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_usage_period_end ON subscriptions(usage_period_end);
CREATE INDEX IF NOT EXISTS idx_subscriptions_billing_next_billing_at ON subscriptions(billing_next_billing_at);
CREATE INDEX IF NOT EXISTS idx_subscriptions_created_at ON subscriptions(created_at);

-- Add constraint to ensure only one active subscription per company
CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_company_active
    ON subscriptions(company_id)
    WHERE status IN ('active', 'trial');

-- Add constraint to ensure valid usage remaining
ALTER TABLE subscriptions ADD CONSTRAINT chk_subscriptions_usage_remaining
    CHECK (usage_nfce_remaining >= -1);
