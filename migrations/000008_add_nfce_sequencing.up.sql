-- Add NFC-e sequencing fields to companies table
ALTER TABLE companies ADD COLUMN IF NOT EXISTS ultimo_numero_nfce BIGINT DEFAULT 0;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS serie_atual_nfce VARCHAR(3) DEFAULT '1';

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_companies_ultimo_numero_nfce ON companies(ultimo_numero_nfce);

-- Add comment
COMMENT ON COLUMN companies.ultimo_numero_nfce IS 'Último número sequencial usado para NFC-e desta empresa';
COMMENT ON COLUMN companies.serie_atual_nfce IS 'Série atual da NFC-e para esta empresa';
