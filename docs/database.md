# 🗄️ Banco de Dados - NFC-e API

## Visão Geral

O sistema utiliza **PostgreSQL** como banco de dados principal, com foco em:
- **Integridade**: Constraints e foreign keys
- **Performance**: Índices otimizados
- **Auditabilidade**: Histórico completo de eventos
- **Escalabilidade**: Estrutura preparada para crescimento

## 📊 Esquema das Tabelas

### nfce_requests
Tabela principal que armazena todas as solicitações de NFC-e.

```sql
CREATE TABLE nfce_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN (
        'pending', 'processing', 'authorized', 'rejected',
        'contingency', 'retrying', 'canceled'
    )),
    payload JSONB NOT NULL,

    -- SEFAZ Response Data
    chave_acesso VARCHAR(44) UNIQUE,
    protocolo VARCHAR(15),
    numero VARCHAR(9),
    serie VARCHAR(3),

    -- Error Information
    rejection_code VARCHAR(3),
    rejection_msg TEXT,
    cstat VARCHAR(3),
    xmotivo TEXT,

    -- Processing Metadata
    retry_count INTEGER DEFAULT 0,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    authorized_at TIMESTAMP WITH TIME ZONE,

    -- Contingency
    in_contingency BOOLEAN DEFAULT FALSE,
    contingency_type VARCHAR(20),

    -- Storage References
    xml_url TEXT,
    pdf_url TEXT,
    qrcode_url TEXT,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_nfce_requests_idempotency_key ON nfce_requests(idempotency_key);
CREATE INDEX idx_nfce_requests_status ON nfce_requests(status);
CREATE INDEX idx_nfce_requests_company_id ON nfce_requests(company_id);
CREATE INDEX idx_nfce_requests_next_retry_at ON nfce_requests(next_retry_at);
CREATE INDEX idx_nfce_requests_chave_acesso ON nfce_requests(chave_acesso);
CREATE INDEX idx_nfce_requests_created_at ON nfce_requests(created_at DESC);
```

### nfce_events
Histórico de auditoria de todas as transições de estado.

```sql
CREATE TABLE nfce_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES nfce_requests(id) ON DELETE CASCADE,

    -- Status Transition
    status_from VARCHAR(20),
    status_to VARCHAR(20) NOT NULL,

    -- SEFAZ Information
    cstat VARCHAR(3),
    message TEXT,

    -- Additional Metadata
    metadata JSONB,

    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_nfce_events_request_id ON nfce_events(request_id);
CREATE INDEX idx_nfce_events_created_at ON nfce_events(created_at DESC);
CREATE INDEX idx_nfce_events_status_to ON nfce_events(status_to);
```

### companies
Empresas emitentes de NFC-e.

```sql
CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cnpj VARCHAR(14) UNIQUE NOT NULL,
    razao_social VARCHAR(255) NOT NULL,
    nome_fantasia VARCHAR(255),
    inscricao_estadual VARCHAR(20),
    inscricao_municipal VARCHAR(20),
    regime_tributario VARCHAR(20) CHECK (regime_tributario IN ('simples', 'normal', 'mei')),
    cnae VARCHAR(10),
    endereco JSONB,
    contato JSONB,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE UNIQUE INDEX idx_companies_cnpj ON companies(cnpj);
CREATE INDEX idx_companies_status ON companies(status);
```

### plans e subscriptions
Sistema de planos e assinaturas (futuro).

```sql
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    limits JSONB, -- {"nfce_per_month": 1000, "storage_gb": 10}
    price_monthly DECIMAL(10,2),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id),
    plan_id UUID NOT NULL REFERENCES plans(id),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'canceled', 'suspended')),
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_subscriptions_company_id ON subscriptions(company_id);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
```

### webhooks
Configuração de webhooks para notificações assíncronas.

```sql
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id),
    url TEXT NOT NULL,
    events TEXT[] NOT NULL, -- ['authorized', 'rejected', 'canceled']
    secret VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_webhooks_company_id ON webhooks(company_id);
CREATE INDEX idx_webhooks_status ON webhooks(status);
```

## 🔄 Migrações

As migrações são gerenciadas com [golang-migrate](https://github.com/golang-migrate/migrate).

### Estrutura dos Arquivos
```
migrations/
├── 000001_create_companies_table.up.sql
├── 000001_create_companies_table.down.sql
├── 000002_create_plans_table.up.sql
├── 000002_create_plans_table.down.sql
├── 000003_create_nfce_requests_table.up.sql
├── 000003_create_nfce_requests_table.down.sql
└── ...
```

### Executando Migrações

```bash
# Via Docker
docker-compose exec api migrate -path /app/migrations -database $DATABASE_URL up

# Via Go
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up
```

## 📈 Consultas Comuns

### Status das NFC-e por Empresa
```sql
SELECT
    company_id,
    status,
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE) as today
FROM nfce_requests
WHERE company_id = $1
GROUP BY company_id, status
ORDER BY status;
```

### NFC-e Autorizadas Hoje
```sql
SELECT
    id,
    chave_acesso,
    protocolo,
    authorized_at
FROM nfce_requests
WHERE status = 'authorized'
  AND DATE(authorized_at) = CURRENT_DATE
ORDER BY authorized_at DESC;
```

### Estatísticas de Performance
```sql
SELECT
    DATE(created_at) as date,
    COUNT(*) as total_requests,
    AVG(EXTRACT(EPOCH FROM (processed_at - created_at))) as avg_processing_time,
    COUNT(*) FILTER (WHERE status = 'authorized') as authorized,
    COUNT(*) FILTER (WHERE status = 'rejected') as rejected
FROM nfce_requests
WHERE created_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

### Retry Queue
```sql
SELECT
    id,
    retry_count,
    next_retry_at,
    created_at
FROM nfce_requests
WHERE status = 'retrying'
  AND next_retry_at <= NOW()
ORDER BY next_retry_at ASC;
```

## 🔍 Índices e Performance

### Índices Principais
- **idempotency_key**: UNIQUE para evitar duplicatas
- **status**: Filtrar por estado atual
- **company_id**: Agrupar por empresa
- **next_retry_at**: Buscar itens para retry
- **chave_acesso**: Consultar NFC-e autorizada
- **created_at**: Ordenação cronológica

### Índices Compostos Recomendados
```sql
CREATE INDEX idx_nfce_requests_company_status_date
ON nfce_requests(company_id, status, DATE(created_at));

CREATE INDEX idx_nfce_requests_status_retry
ON nfce_requests(status, next_retry_at)
WHERE status = 'retrying';
```

## 🛡️ Constraints e Validações

### Check Constraints
- **Status válido**: Apenas valores permitidos
- **CNPJ formatado**: 14 dígitos numéricos
- **Chave de acesso**: 44 caracteres alfanuméricos

### Foreign Keys
- **company_id**: Referência para tabela companies
- **request_id**: Em nfce_events referencia nfce_requests

### Unique Constraints
- **idempotency_key**: Uma por NFC-e
- **chave_acesso**: Uma por NFC-e autorizada
- **cnpj**: Uma por empresa

## 📊 Particionamento (Futuro)

Para volumes altos, considere particionar por:
- **created_at**: Partições mensais
- **company_id**: Para isolamento multi-tenant

```sql
-- Exemplo de particionamento por mês
CREATE TABLE nfce_requests_y2024m12 PARTITION OF nfce_requests
    FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');
```

## 🔄 Backup e Recovery

### Estratégia de Backup
- **Diário**: Backup completo das tabelas principais
- **Horário**: Backup incremental dos eventos
- **Point-in-Time Recovery**: Para recuperação precisa

### Comando de Backup
```bash
pg_dump -h localhost -U plugnfce -d plugnfce -F c -b -v -f backup_$(date +%Y%m%d_%H%M%S).dump
```

### Restauração
```bash
pg_restore -h localhost -U plugnfce -d plugnfce -v backup_file.dump
```

## 📈 Monitoramento

### Métricas Importantes
- **Tamanho das tabelas**: Crescimento ao longo do tempo
- **Queries lentas**: Identificar gargalos
- **Uso de índices**: Efetividade dos índices
- **Deadlocks**: Conflitos de transação

### Queries de Monitoramento
```sql
-- Tamanho das tabelas
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Índices não utilizados
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
ORDER BY tablename, indexname;
```

## 🚀 Otimizações

### Para Alto Volume
1. **Connection Pooling**: PgBouncer
2. **Read Replicas**: Para consultas de status
3. **Caching**: Redis para status frequentes
4. **Archiving**: Mover dados antigos para cold storage

### Configurações PostgreSQL
```ini
# postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
```

---

**Versão do Schema**: 1.0.0
**Última atualização**: Dezembro 2024
