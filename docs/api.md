# 🌐 API NFC-e - Documentação

## Visão Geral

A API NFC-e fornece endpoints REST para emissão de Nota Fiscal de Consumidor Eletrônica modelo 65. Utiliza processamento assíncrono para garantir alta performance e resiliência.

**Base URL**: `http://localhost:8080` (desenvolvimento)

## 🔐 Autenticação

Atualmente, a API não implementa autenticação. Em produção, considere:
- JWT Tokens
- API Keys
- OAuth 2.0
- mTLS

## 📋 Endpoints

### NFC-e

#### `POST /nfce`
Emite uma nova NFC-e de forma assíncrona.

**Headers:**
```
Content-Type: application/json
Idempotency-Key: <string> (obrigatório, único por requisição)
```

**Request Body:**
```json
{
  "uf": "SP",
  "ambiente": "producao",
  "emitente": {
    "cnpj": "12345678000123",
    "ie": "123456789",
    "regime": "simples",
    "csc_id": "000001",
    "csc_token": "ABCDEF123456"
  },
  "itens": [
    {
      "descricao": "Produto de exemplo",
      "ncm": "12345678",
      "cfop": "5102",
      "gtin": "789123456789",
      "valor": 29.90,
      "quantidade": 1,
      "unidade": "UN"
    }
  ],
  "pagamentos": [
    {
      "forma": "01",
      "valor": 29.90
    }
  ],
  "certificado": {
    "cert_pfx_b64": "base64-do-certificado-pfx",
    "cert_password": "senha-do-certificado"
  },
  "options": {
    "contingencia": false,
    "sync": false
  }
}
```

**Response (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "created_at": "2024-12-23T10:30:00Z",
  "links": {
    "self": "/nfce/550e8400-e29b-41d4-a716-446655440000",
    "status": "/nfce/550e8400-e29b-41d4-a716-446655440000/status"
  }
}
```

**Códigos de Erro:**
- `400 Bad Request` - Dados inválidos
- `409 Conflict` - Idempotency-Key já utilizado
- `422 Unprocessable Entity` - Erro de validação
- `500 Internal Server Error` - Erro interno

#### `GET /nfce/{id}`
Consulta o status de uma NFC-e pelo ID.

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "authorized",
  "chave_acesso": "35241234567890000126550010000000011234567890",
  "protocolo": "135412345678900",
  "numero": "000000001",
  "serie": "1",
  "created_at": "2024-12-23T10:30:00Z",
  "processed_at": "2024-12-23T10:30:15Z",
  "authorized_at": "2024-12-23T10:30:15Z",
  "links": {
    "xml": "/nfce/550e8400-e29b-41d4-a716-446655440000/xml",
    "pdf": "/nfce/550e8400-e29b-41d4-a716-446655440000/pdf",
    "qrcode": "/nfce/550e8400-e29b-41d4-a716-446655440000/qrcode"
  }
}
```

**Status Possíveis:**
- `pending` - Aguardando processamento
- `processing` - Sendo processado
- `authorized` - Autorizado pela SEFAZ
- `rejected` - Rejeitado pela SEFAZ
- `contingency` - Emitido em contingência
- `retrying` - Tentando novamente após erro
- `canceled` - Cancelado

#### `GET /nfce/{id}/xml`
Retorna o XML autorizado da NFC-e.

**Response (200 OK):**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <!-- XML completo da NFC-e -->
</NFe>
```

#### `GET /nfce/{id}/pdf`
Retorna o DANFE (PDF) da NFC-e.

**Response (200 OK):**
```
Content-Type: application/pdf
Content-Disposition: attachment; filename="nfce-35241234567890000126550010000000011234567890.pdf"
```

#### `GET /nfce/{id}/qrcode`
Retorna a imagem do QR Code da NFC-e.

**Response (200 OK):**
```
Content-Type: image/png
```

#### `POST /nfce/{id}/cancel`
Cancela uma NFC-e autorizada.

**Request Body:**
```json
{
  "justificativa": "Cancelamento solicitado pelo cliente"
}
```

**Response (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "canceled",
  "canceled_at": "2024-12-23T11:00:00Z"
}
```

### Sistema

#### `GET /health`
Verifica a saúde da API.

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2024-12-23T10:30:00Z",
  "services": {
    "database": "up",
    "queue": "up"
  }
}
```

#### `GET /status`
Retorna informações do sistema.

**Response (200 OK):**
```json
{
  "version": "1.0.0",
  "environment": "development",
  "uptime": "2h 30m",
  "stats": {
    "total_nfce": 1250,
    "authorized_today": 45,
    "rejected_today": 3
  }
}
```

## 📊 Campos Obrigatórios

### Emitente
- `cnpj`: CNPJ do emitente (14 dígitos)
- `ie`: Inscrição Estadual
- `regime`: Regime tributário ("simples", "normal")
- `csc_id`: ID do Código de Segurança do Contribuinte
- `csc_token`: Token do CSC

### Itens
- `descricao`: Descrição do produto (até 120 caracteres)
- `ncm`: Código NCM (8 dígitos)
- `cfop`: CFOP (4 dígitos)
- `valor`: Valor unitário (2 casas decimais)
- `quantidade`: Quantidade (4 casas decimais)
- `unidade`: Unidade de medida

### Pagamentos
- `forma`: Código da forma de pagamento (2 dígitos)
- `valor`: Valor do pagamento (2 casas decimais)

### Certificado Digital
- `cert_pfx_b64`: Certificado A1 em base64
- `cert_password`: Senha do certificado

## 🎯 Idempotência

Todas as requisições de emissão devem incluir o header `Idempotency-Key`. Este valor deve ser único e gerado pelo cliente. Se a mesma chave for enviada novamente:

- Se a NFC-e ainda não foi processada → retorna status atual
- Se a NFC-e foi autorizada → retorna dados completos
- Se a NFC-e foi rejeitada → retorna erro de rejeição

## ⚡ Limites e Rate Limiting

- **Máximo de itens por NFC-e**: 56
- **Tamanho máximo da descrição**: 120 caracteres
- **Valor máximo por item**: R$ 9.999.999,99
- **Rate limit**: 100 requisições/minuto por IP (configurável)

## 🚨 Tratamento de Erros

### Estrutura de Erro
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Dados de entrada inválidos",
    "details": {
      "field": "emitente.cnpj",
      "reason": "CNPJ deve ter 14 dígitos"
    }
  }
}
```

### Códigos de Erro Comuns
- `VALIDATION_ERROR` - Dados inválidos
- `IDEMPOTENCY_CONFLICT` - Chave de idempotência já utilizada
- `NFC_E_NOT_FOUND` - NFC-e não encontrada
- `NFC_E_ALREADY_CANCELED` - NFC-e já cancelada
- `SERVICE_UNAVAILABLE` - Serviço temporariamente indisponível

## 🔄 Webhooks (Futuro)

Para notificações assíncronas, configure webhooks:

```json
{
  "url": "https://minha-api.com/webhooks/nfce",
  "events": ["authorized", "rejected", "canceled"],
  "secret": "webhook-secret"
}
```

## 🧪 Exemplos de Uso

### cURL
```bash
# Emitir NFC-e
curl -X POST http://localhost:8080/nfce \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d @nfce-payload.json

# Consultar status
curl http://localhost:8080/nfce/550e8400-e29b-41d4-a716-446655440000
```

### JavaScript/Node.js
```javascript
const response = await fetch('/nfce', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Idempotency-Key': crypto.randomUUID()
  },
  body: JSON.stringify(nfceData)
});

const result = await response.json();
console.log('NFC-e criada:', result.id);
```

## 📈 Monitoramento

### Métricas Disponíveis
- Tempo de resposta da API
- Taxa de sucesso de emissão
- Tempo de processamento do Worker
- Latência da SEFAZ
- Uso de recursos (CPU/Memória)

### Logs
Todos os requests são logados com:
- Request ID (correlação)
- Timestamp
- Status HTTP
- Tempo de processamento
- Erros (se houver)

---

**Versão da API**: 1.0.0
**Última atualização**: Dezembro 2024
