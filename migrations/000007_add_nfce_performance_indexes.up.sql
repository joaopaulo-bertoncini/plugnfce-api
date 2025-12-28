-- Add performance indexes for NFC-e requests table
-- These indexes optimize the most critical queries used by the worker and API

-- Composite index for GetPendingRetries query: status + next_retry_at
-- This is the most critical index as it's used every 15 seconds by the retry scheduler
CREATE INDEX IF NOT EXISTS idx_nfce_requests_status_next_retry_at
ON nfce_requests(status, next_retry_at)
WHERE next_retry_at IS NOT NULL;

-- Partial index for retrying status (more selective than the composite above)
CREATE INDEX IF NOT EXISTS idx_nfce_requests_retrying_only
ON nfce_requests(next_retry_at)
WHERE status = 'retrying' AND next_retry_at IS NOT NULL;

-- Composite index for status + created_at (useful for cleanup queries and monitoring)
CREATE INDEX IF NOT EXISTS idx_nfce_requests_status_created_at
ON nfce_requests(status, created_at);

-- Index for company-based queries (useful for multi-tenant operations)
CREATE INDEX IF NOT EXISTS idx_nfce_requests_company_status_created
ON nfce_requests(company_id, status, created_at DESC);

-- Index for cleanup of old failed requests (simplified for development)
CREATE INDEX IF NOT EXISTS idx_nfce_requests_failed_old
ON nfce_requests(created_at)
WHERE status IN ('rejected', 'canceled');
