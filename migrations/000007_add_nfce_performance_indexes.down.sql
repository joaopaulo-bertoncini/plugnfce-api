-- Drop performance indexes for NFC-e requests table
DROP INDEX CONCURRENTLY IF EXISTS idx_nfce_requests_failed_old;
DROP INDEX CONCURRENTLY IF EXISTS idx_nfce_requests_company_status_created;
DROP INDEX CONCURRENTLY IF EXISTS idx_nfce_requests_status_created_at;
DROP INDEX CONCURRENTLY IF EXISTS idx_nfce_requests_retrying_only;
DROP INDEX CONCURRENTLY IF EXISTS idx_nfce_requests_status_next_retry_at;
