-- Recreate foreign key constraint for company_id
ALTER TABLE nfce_requests ADD CONSTRAINT fk_nfce_requests_company_id FOREIGN KEY (company_id) REFERENCES companies(id);
