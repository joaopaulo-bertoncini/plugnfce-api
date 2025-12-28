-- Recreate foreign key constraint and make company_id NOT NULL again
ALTER TABLE nfce_requests ADD CONSTRAINT fk_nfce_requests_company_id FOREIGN KEY (company_id) REFERENCES companies(id);
ALTER TABLE nfce_requests ALTER COLUMN company_id SET NOT NULL;
