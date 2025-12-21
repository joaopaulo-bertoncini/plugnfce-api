-- Create webhooks table
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(500) NOT NULL,
    method VARCHAR(10) DEFAULT 'POST' CHECK (method IN ('POST', 'PUT', 'PATCH')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'failed')),

    -- Events (stored as JSON array)
    events JSONB NOT NULL DEFAULT '[]',

    -- Headers (stored as JSON object)
    headers JSONB DEFAULT '{}',

    -- Authentication
    secret VARCHAR(255),

    -- Retry configuration
    retry_max_retries INTEGER DEFAULT 3,
    retry_interval_seconds INTEGER DEFAULT 5, -- in seconds
    retry_max_interval_seconds INTEGER DEFAULT 300, -- in seconds (5 minutes)

    -- Statistics
    total_deliveries INTEGER DEFAULT 0,
    successful_deliveries INTEGER DEFAULT 0,
    failed_deliveries INTEGER DEFAULT 0,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_delivery_at TIMESTAMPTZ
);

-- Create webhook_deliveries table
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event VARCHAR(50) NOT NULL,
    payload JSONB,
    attempt INTEGER DEFAULT 1,
    status_code INTEGER,
    response_body TEXT,
    error_message TEXT,
    succeeded BOOLEAN DEFAULT FALSE,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for webhooks
CREATE INDEX IF NOT EXISTS idx_webhooks_company_id ON webhooks(company_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_status ON webhooks(status);
CREATE INDEX IF NOT EXISTS idx_webhooks_events ON webhooks USING GIN(events);
CREATE INDEX IF NOT EXISTS idx_webhooks_created_at ON webhooks(created_at);
CREATE INDEX IF NOT EXISTS idx_webhooks_last_delivery_at ON webhooks(last_delivery_at);

-- Create indexes for webhook_deliveries
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook_id ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_event ON webhook_deliveries(event);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_succeeded ON webhook_deliveries(succeeded);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_created_at ON webhook_deliveries(created_at);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_delivered_at ON webhook_deliveries(delivered_at);

-- Add constraints
ALTER TABLE webhooks ADD CONSTRAINT chk_webhooks_events_not_empty
    CHECK (jsonb_array_length(events) > 0);

ALTER TABLE webhooks ADD CONSTRAINT chk_webhooks_retry_config
    CHECK (retry_max_retries >= 0 AND retry_interval_seconds > 0 AND retry_max_interval_seconds > 0);

ALTER TABLE webhooks ADD CONSTRAINT chk_webhooks_statistics
    CHECK (total_deliveries >= 0 AND successful_deliveries >= 0 AND failed_deliveries >= 0
           AND successful_deliveries + failed_deliveries <= total_deliveries);

ALTER TABLE webhook_deliveries ADD CONSTRAINT chk_webhook_deliveries_attempt
    CHECK (attempt > 0);
