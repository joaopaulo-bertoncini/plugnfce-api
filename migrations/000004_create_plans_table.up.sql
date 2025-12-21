-- Create plans table
CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(20) DEFAULT 'monthly' CHECK (type IN ('monthly', 'yearly', 'package')),
    billing_cycle VARCHAR(20) DEFAULT 'monthly' CHECK (billing_cycle IN ('monthly', 'yearly', 'once')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'archived')),

    -- Pricing
    price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'BRL',

    -- Quotas
    quota_type VARCHAR(20) DEFAULT 'monthly' CHECK (quota_type IN ('monthly', 'package', 'unlimited')),
    max_nfce_per_month INTEGER,
    max_nfce_total INTEGER,

    -- Features
    features_max_nfce_per_month INTEGER,
    features_max_nfce_total INTEGER,
    features_allow_contingency BOOLEAN DEFAULT TRUE,
    features_allow_cancellation BOOLEAN DEFAULT TRUE,
    features_allow_inutilization BOOLEAN DEFAULT TRUE,
    features_webhook_support BOOLEAN DEFAULT TRUE,
    features_priority_support BOOLEAN DEFAULT FALSE,
    features_storage_days INTEGER DEFAULT 365,

    -- Metadata
    is_popular BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    trial_days INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for plans
CREATE INDEX IF NOT EXISTS idx_plans_status ON plans(status);
CREATE INDEX IF NOT EXISTS idx_plans_type ON plans(type);
CREATE INDEX IF NOT EXISTS idx_plans_is_popular ON plans(is_popular);
CREATE INDEX IF NOT EXISTS idx_plans_sort_order ON plans(sort_order);
CREATE INDEX IF NOT EXISTS idx_plans_created_at ON plans(created_at);

-- Add constraint to ensure at least one quota is set when not unlimited
ALTER TABLE plans ADD CONSTRAINT chk_plans_quota
    CHECK (
        quota_type = 'unlimited' OR
        (quota_type = 'monthly' AND max_nfce_per_month > 0) OR
        (quota_type = 'package' AND max_nfce_total > 0)
    );
