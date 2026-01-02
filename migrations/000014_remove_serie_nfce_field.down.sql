-- Rollback: Add back serie_nfce field to companies table

ALTER TABLE companies ADD COLUMN IF NOT EXISTS serie_nfce VARCHAR(3) DEFAULT '1';

-- Add comment explaining the field
COMMENT ON COLUMN companies.serie_nfce IS 'NFC-e series (currently always 1)';
