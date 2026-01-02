-- Rollback: Add back obsolete NFC-e fields to companies table
-- This rollback is provided for compatibility but these fields are no longer used

-- Add back ultimo_numero_nfce field
ALTER TABLE companies ADD COLUMN IF NOT EXISTS ultimo_numero_nfce BIGINT DEFAULT 0;

-- Add back serie_atual_nfce field
ALTER TABLE companies ADD COLUMN IF NOT EXISTS serie_atual_nfce VARCHAR(3) DEFAULT '1';

-- Add comments to explain these fields are obsolete
COMMENT ON COLUMN companies.ultimo_numero_nfce IS 'OBSOLETE: Use nfce_sequences.ultimo_numero instead';
COMMENT ON COLUMN companies.serie_atual_nfce IS 'OBSOLETE: This field is no longer used';
