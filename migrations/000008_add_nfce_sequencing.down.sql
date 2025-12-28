-- Remove NFC-e sequencing fields from companies table
DROP INDEX IF EXISTS idx_companies_ultimo_numero_nfce;
ALTER TABLE companies DROP COLUMN IF EXISTS serie_atual_nfce;
ALTER TABLE companies DROP COLUMN IF EXISTS ultimo_numero_nfce;
