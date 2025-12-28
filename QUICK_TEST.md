# Quick NFC-e Test Commands

# 1. Start infrastructure
make docker-up

# 2. Run migrations  
make migrate

# 3. Start API (Terminal 1)
make run-api

# 4. Start Worker (Terminal 2) 
make run-worker

# 5. Run automated test
make test-api

# Manual test example:
IDEMPOTENCY_KEY=$(uuidgen)
curl -X POST http://localhost:8080/nfce \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d '{
    "uf": "SP",
    "ambiente": "homologacao",
    "emitente": {
      "cnpj": "12345678000123",
      "ie": "123456789",
      "regime": "simples",
      "csc_id": "000001",
      "csc_token": "ABCDEF123456"
    },
    "itens": [{
      "descricao": "Produto Teste",
      "ncm": "84713019",
      "cfop": "5102",
      "gtin": "7891234567890",
      "valor": 29.90,
      "quantidade": 1,
      "unidade": "UN"
    }],
    "pagamentos": [{
      "forma": "01",
      "valor": 29.90
    }],
    "certificado": {
      "cert_pfx_b64": "TEST_CERTIFICATE_PLACEHOLDER",
      "cert_password": "test_password"
    }
  }'
