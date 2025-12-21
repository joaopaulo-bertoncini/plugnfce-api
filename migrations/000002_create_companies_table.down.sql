-- Remove foreign key constraint from nfce_requests
ALTER TABLE nfce_requests DROP CONSTRAINT IF EXISTS fk_nfce_requests_company_id;

-- Drop companies table
DROP TABLE IF EXISTS companies;
