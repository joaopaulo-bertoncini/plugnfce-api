-- Create NFC-e sequences table for per-company sequencing
CREATE TABLE IF NOT EXISTS nfce_sequences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL UNIQUE REFERENCES companies(id) ON DELETE CASCADE,
    serie VARCHAR(3) NOT NULL DEFAULT '1',
    ultimo_numero BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_nfce_sequences_company_id ON nfce_sequences(company_id);

-- Add comments
COMMENT ON TABLE nfce_sequences IS 'Sequences de numeração NFC-e por empresa';
COMMENT ON COLUMN nfce_sequences.company_id IS 'ID da empresa';
COMMENT ON COLUMN nfce_sequences.serie IS 'Série da NFC-e';
COMMENT ON COLUMN nfce_sequences.ultimo_numero IS 'Último número sequencial usado';

-- Function to get next NFC-e number atomically
CREATE OR REPLACE FUNCTION get_next_nfce_number(company_uuid UUID, nfce_serie VARCHAR(3) DEFAULT '1')
RETURNS BIGINT AS $$
DECLARE
    next_number BIGINT;
BEGIN
    -- Try to update existing sequence
    UPDATE nfce_sequences
    SET ultimo_numero = ultimo_numero + 1, updated_at = NOW()
    WHERE company_id = company_uuid AND serie = nfce_serie;

    -- If no row was updated, insert new sequence
    IF NOT FOUND THEN
        INSERT INTO nfce_sequences (company_id, serie, ultimo_numero)
        VALUES (company_uuid, nfce_serie, 1)
        RETURNING ultimo_numero INTO next_number;
    ELSE
        -- Get the updated value
        SELECT ultimo_numero INTO next_number
        FROM nfce_sequences
        WHERE company_id = company_uuid AND serie = nfce_serie;
    END IF;

    RETURN next_number;
END;
$$ LANGUAGE plpgsql;
