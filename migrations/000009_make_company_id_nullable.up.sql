-- Make company_id nullable for NFC-e requests (temporary until company management is implemented)
ALTER TABLE nfce_requests DROP CONSTRAINT IF EXISTS fk_nfce_requests_company_id;
ALTER TABLE nfce_requests ALTER COLUMN company_id DROP NOT NULL;
