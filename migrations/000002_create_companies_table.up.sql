-- Create companies table
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cnpj VARCHAR(14) NOT NULL UNIQUE,
    razao_social VARCHAR(255) NOT NULL,
    nome_fantasia VARCHAR(255),
    inscricao_estadual VARCHAR(20),

    -- Contact info
    email VARCHAR(255) NOT NULL,

    -- Address
    endereco_logradouro VARCHAR(255) NOT NULL,
    endereco_numero VARCHAR(10) NOT NULL,
    endereco_complemento VARCHAR(100),
    endereco_bairro VARCHAR(100) NOT NULL,
    endereco_codigo_municipio VARCHAR(7) NOT NULL,
    endereco_municipio VARCHAR(100) NOT NULL,
    endereco_uf VARCHAR(2) NOT NULL,
    endereco_cep VARCHAR(8) NOT NULL,

    -- Certificate
    certificado_type VARCHAR(10) DEFAULT 'a1' CHECK (certificado_type IN ('a1')),
    certificado_pfx_data BYTEA,
    certificado_password VARCHAR(255),
    certificado_expires_at TIMESTAMPTZ,
    certificado_subject VARCHAR(500),

    -- CSC Configuration
    csc_id VARCHAR(10),
    csc_token VARCHAR(32),
    csc_valid_from TIMESTAMPTZ,
    csc_valid_until TIMESTAMPTZ,

    -- Business configuration
    regime_tributario VARCHAR(50) DEFAULT 'simples_nacional' CHECK (regime_tributario IN ('simples_nacional', 'lucro_presumido', 'lucro_real')),
    serie_nfce VARCHAR(3) DEFAULT '1',

    -- Status
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'blocked')),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for companies
CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_cnpj ON companies(cnpj);
CREATE INDEX IF NOT EXISTS idx_companies_email ON companies(email);
CREATE INDEX IF NOT EXISTS idx_companies_status ON companies(status);
CREATE INDEX IF NOT EXISTS idx_companies_created_at ON companies(created_at);

-- Add foreign key constraint to nfce_requests
ALTER TABLE nfce_requests ADD CONSTRAINT fk_nfce_requests_company_id
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE;
