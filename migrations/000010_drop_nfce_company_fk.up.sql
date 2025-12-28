-- Drop foreign key constraint for company_id (temporary until company management is implemented)
ALTER TABLE nfce_requests DROP CONSTRAINT IF EXISTS fk_nfce_requests_company_id;
