# 🧪 Testing Guide - NFC-e API & Worker

This guide shows you how to test the NFC-e emission system, including the worker functionality and generating test NFC-e notes.

## 🚀 Quick Start Testing

### Prerequisites
- Docker and Docker Compose installed
- Go 1.24+ installed (for local development)
- All services running: `make docker-up`

### Step 1: Start Infrastructure Services
```bash
# Start all infrastructure services
make docker-up

# Or manually with docker-compose
docker-compose up -d
```

Verify services are running:
```bash
docker-compose ps
```

### Step 2: Run Database Migrations
```bash
make migrate
```

### Step 3: Start API and Worker

#### Option A: Run Locally (Recommended for Testing)
```bash
# Terminal 1: Start API server
make run-api

# Terminal 2: Start worker process
make run-worker
```

#### Option B: Run with Docker (if you add services to docker-compose.yml)
```yaml
# Add to docker-compose.yml:
api:
  build: .
  ports:
    - "8080:8080"
  depends_on:
    - db
    - rabbitmq
    - minio
  environment:
    - ENV=development
  command: ["./bin/plugnfce-api"]

worker:
  build: .
  depends_on:
    - db
    - rabbitmq
    - minio
  environment:
    - ENV=development
  command: ["./bin/plugnfce-worker"]
```

### Step 4: Run the Test Script
```bash
# Run the automated test script
make test-api

# Or run directly
./scripts/test_api.sh
```

## 📋 Manual Testing with cURL

### 1. Check API Health
```bash
curl http://localhost:8080/health
```

### 2. Create a Test NFC-e
```bash
# Generate a unique idempotency key
IDEMPOTENCY_KEY=$(uuidgen)

# Step 0: Upload Certificate (once per company)
COMPANY_ID="your-company-id"

# Upload Certificate via File (multipart/form-data)
curl -X PUT http://localhost:8080/api/v1/companies/$COMPANY_ID/certificate \
  -F "pfx_file=@certificado.pfx" \
  -F "password=your_certificate_password" \
  -F "expires_at=2025-12-31T23:59:59Z"

### Frontend JavaScript Example:

```javascript
async function uploadCertificate(companyId, pfxFile, password, expiresAt) {
  const formData = new FormData();
  formData.append('pfx_file', pfxFile);      // File object from <input type="file">
  formData.append('password', password);
  formData.append('expires_at', expiresAt);

  const response = await fetch(`/api/v1/companies/${companyId}/certificate`, {
    method: 'PUT',
    body: formData  // Browser sets Content-Type automatically
  });

  return response.json();
}

// Note: Company configuration fields removed:
// - ultimo_numero_nfce (now managed by nfce_sequences table)
// - serie_atual_nfce (redundant field removed)
```

#### HTML Form Example:
```html
<form id="certificateForm" enctype="multipart/form-data">
  <input type="file" id="pfxFile" accept=".pfx,.p12" required>
  <input type="password" id="password" placeholder="Certificate password" required>
  <input type="datetime-local" id="expiresAt" required>
  <button type="submit">Upload Certificate</button>
</form>

<script>
document.getElementById('certificateForm').addEventListener('submit', async (e) => {
  e.preventDefault();

  const companyId = 'your-company-id';
  const pfxFile = document.getElementById('pfxFile').files[0];
  const password = document.getElementById('password').value;
  const expiresAt = new Date(document.getElementById('expiresAt').value).toISOString();

  try {
    const result = await uploadCertificate(companyId, pfxFile, password, expiresAt);
    console.log('Certificate uploaded successfully:', result);
  } catch (error) {
    console.error('Upload failed:', error);
  }
});
</script>
```

# Create NFC-e with test data
# Note: Certificate is now managed per company, not sent in payload
curl -X POST http://localhost:8080/api/v1/nfce \
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
    "itens": [
      {
        "descricao": "Produto de Teste NFC-e",
        "ncm": "84713019",
        "cfop": "5102",
        "gtin": "7891234567890",
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
    "options": {
      "contingencia": false,
      "sync": false
    }
  }'
```

### 3. Monitor Processing
```bash
# Extract the NFC-e ID from the response and check status
NFCE_ID="550e8400-e29b-41d4-a716-446655440000"

# Check status (run repeatedly until processed)
curl http://localhost:8080/nfce/$NFCE_ID
```

### 4. Download Generated Files (when authorized)
```bash
# Download XML
curl -o nfce.xml http://localhost:8080/nfce/$NFCE_ID/xml

# Download PDF
curl -o nfce.pdf http://localhost:8080/nfce/$NFCE_ID/pdf

# Download QR Code image
curl -o qrcode.png http://localhost:8080/nfce/$NFCE_ID/qrcode
```

## 🔍 Monitoring & Debugging

### Check Worker Logs
```bash
# View worker logs in real-time
docker-compose logs worker -f

# Or if running locally, check terminal output
```

### Check Queue Status
```bash
# View RabbitMQ queue status
docker-compose exec rabbitmq rabbitmqctl list_queues name messages_ready messages_unacknowledged

# View queue contents (if any)
docker-compose exec rabbitmq rabbitmqctl list_queue_bindings
```

### Check Database
```bash
# Connect to PostgreSQL
docker-compose exec db psql -U plugnfce -d plugnfce

# Check NFC-e requests
SELECT id, status, chave_acesso, created_at FROM nfce_requests ORDER BY created_at DESC LIMIT 5;

# Check events
SELECT request_id, status_from, status_to, created_at FROM nfce_events ORDER BY created_at DESC LIMIT 10;
```

### Check MinIO Storage
```bash
# Access MinIO console: http://localhost:9001
# Login: minioadmin / minioadmin

# Or check via CLI
docker-compose exec minio mc ls plugnfce/
```

## 🎯 Testing Contingency Mode

### Force Contingency Testing
```bash
# Set options.contingencia to true in the request
curl -X POST http://localhost:8080/nfce \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{
    "uf": "SP",
    "ambiente": "homologacao",
    "emitente": { "cnpj": "12345678000123", "ie": "123456789", "regime": "simples", "csc_id": "000001", "csc_token": "ABCDEF123456" },
    "itens": [{ "descricao": "Produto Teste", "ncm": "84713019", "cfop": "5102", "gtin": "7891234567890", "valor": 29.90, "quantidade": 1, "unidade": "UN" }],
    "pagamentos": [{ "forma": "01", "valor": 29.90 }],
    "certificado": { "cert_pfx_b64": "TEST_CERTIFICATE_BASE64_PLACEHOLDER", "cert_password": "test_password" },
    "options": { "contingencia": true, "sync": false }
  }'
```

### Test Automatic Contingency Fallback
```bash
# The system will automatically switch to contingency if SEFAZ returns errors like:
# - cStat 108: Serviço Paralisado Temporariamente
# - cStat 109: Serviço Paralisado sem Previsão
# - cStat 691-693: Contingency-related errors
```

## 📊 Expected Test Results

### Successful NFC-e Emission
1. **Status progression**: `pending` → `processing` → `authorized`
2. **Files generated**: XML, PDF, QR Code in MinIO
3. **Database records**: Complete NFC-e record with chave_acesso and protocolo
4. **Worker logs**: Processing steps logged with correlation IDs

### Contingency Mode
1. **Status**: `contingency`
2. **XML TpEmis**: "6" (SVC-AN) or "7" (SVC-RS)
3. **QR Code**: Generated with contingency-aware parameters
4. **Storage**: Files saved with contingency markers

## 🐛 Troubleshooting

### Worker Not Processing
```bash
# Check if worker is running
ps aux | grep plugnfce-worker

# Check RabbitMQ connection
docker-compose logs rabbitmq | grep -i error

# Check worker logs for errors
docker-compose logs worker
```

### API Not Responding
```bash
# Check API health
curl http://localhost:8080/health

# Check if port 8080 is in use
netstat -tulpn | grep :8080

# Check API logs
docker-compose logs api 2>&1 | tail -50
```

### Database Connection Issues
```bash
# Test database connection
docker-compose exec db psql -U plugnfce -d plugnfce -c "SELECT version();"

# Check database logs
docker-compose logs db
```

## 📈 Performance Testing

### Load Testing
```bash
# Install hey for load testing
go install github.com/rakyll/hey@latest

# Test API with concurrent requests
hey -n 100 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: {{.RequestNumber}}" \
  -d '{"uf":"SP","ambiente":"homologacao",...}' \
  http://localhost:8080/nfce
```

### Monitor System Resources
```bash
# Monitor worker CPU/memory
docker stats

# Check queue backlog
docker-compose exec rabbitmq rabbitmqctl list_queues
```

## 🎉 Success Indicators

✅ **API responds to health checks**  
✅ **NFC-e creation returns 201 status**  
✅ **Worker processes messages from queue**  
✅ **Status changes from pending → processing → authorized**  
✅ **Files are stored in MinIO**  
✅ **Database contains complete NFC-e records**  
✅ **QR codes are scannable and valid**  

Your NFC-e system is working correctly when you can create NFC-e requests and see them progress through the worker pipeline to successful authorization! 🚀
