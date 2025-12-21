-- Create nfce_requests table
CREATE TABLE IF NOT EXISTS nfce_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'authorized', 'rejected', 'contingency', 'retrying', 'canceled')),
    payload JSONB NOT NULL,

    -- SEFAZ response data
    chave_acesso VARCHAR(44),
    protocolo VARCHAR(15),
    numero VARCHAR(9),
    serie VARCHAR(3),

    -- Error handling
    rejection_code VARCHAR(10),
    rejection_msg TEXT,
    cstat VARCHAR(10),
    xmotivo TEXT,

    -- Processing metadata
    retry_count INTEGER DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    processed_at TIMESTAMPTZ,
    authorized_at TIMESTAMPTZ,

    -- Contingency
    in_contingency BOOLEAN DEFAULT FALSE,
    contingency_type VARCHAR(20),

    -- Storage references
    xml_url VARCHAR(500),
    pdf_url VARCHAR(500),
    qrcode_url VARCHAR(500),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for nfce_requests
CREATE INDEX IF NOT EXISTS idx_nfce_requests_idempotency_key ON nfce_requests(idempotency_key);
CREATE INDEX IF NOT EXISTS idx_nfce_requests_status ON nfce_requests(status);
CREATE INDEX IF NOT EXISTS idx_nfce_requests_next_retry_at ON nfce_requests(next_retry_at);
CREATE INDEX IF NOT EXISTS idx_nfce_requests_chave_acesso ON nfce_requests(chave_acesso);
CREATE INDEX IF NOT EXISTS idx_nfce_requests_company_id ON nfce_requests(company_id);
CREATE INDEX IF NOT EXISTS idx_nfce_requests_created_at ON nfce_requests(created_at);

-- Create nfce_events table
CREATE TABLE IF NOT EXISTS nfce_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES nfce_requests(id) ON DELETE CASCADE,
    status_from VARCHAR(50),
    status_to VARCHAR(50),
    cstat VARCHAR(10),
    message TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for nfce_events
CREATE INDEX IF NOT EXISTS idx_nfce_events_request_id ON nfce_events(request_id);
CREATE INDEX IF NOT EXISTS idx_nfce_events_created_at ON nfce_events(created_at);
CREATE INDEX IF NOT EXISTS idx_nfce_events_status_to ON nfce_events(status_to);
