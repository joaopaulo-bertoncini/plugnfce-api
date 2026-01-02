-- Remove obsolete NFC-e fields from companies table
-- These fields are no longer used since we moved to nfce_sequences table

-- Remove ultimo_numero_nfce field (replaced by nfce_sequences.ultimo_numero)
ALTER TABLE companies DROP COLUMN IF EXISTS ultimo_numero_nfce;

-- Remove serie_atual_nfce field (redundant with serie_nfce)
ALTER TABLE companies DROP COLUMN IF EXISTS serie_atual_nfce;
