#!/bin/bash

# Test script for NFC-e API
# This script creates test data and tests the NFC-e emission workflow

set -e

API_URL="http://localhost:8080/api/v1"
TEST_DATA_FILE="/tmp/nfce_test_data.json"

echo "🧪 Testing NFC-e API and Worker"
echo "================================="

# Create test data
cat > "$TEST_DATA_FILE" << 'EOF'
{
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
  "certificado": {
    "cert_pfx_b64": "TEST_CERTIFICATE_BASE64_PLACEHOLDER",
    "cert_password": "test_password"
  },
  "options": {
    "contingencia": false,
    "sync": false
  }
}
EOF

echo "📝 Created test data file: $TEST_DATA_FILE"

# Generate idempotency key
IDEMPOTENCY_KEY=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "$(date +%s)-$RANDOM")

echo "🔑 Generated idempotency key: $IDEMPOTENCY_KEY"

# Test 2: Create NFC-e
echo ""
echo "📤 Creating NFC-e..."
CREATE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" \
  -X POST "$API_URL/api/v1/nfce" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $IDEMPOTENCY_KEY" \
  -d @"$TEST_DATA_FILE")

if echo "$CREATE_RESPONSE" | grep -q "HTTP_STATUS:201"; then
    echo "✅ NFC-e creation request accepted"

    # Extract NFC-e ID from response
    NFCE_ID=$(echo "$CREATE_RESPONSE" | head -n -1 | jq -r '.id' 2>/dev/null || echo "")

    if [ -n "$NFCE_ID" ] && [ "$NFCE_ID" != "null" ]; then
        echo "📋 NFC-e ID: $NFCE_ID"
    else
        echo "⚠️  Could not extract NFC-e ID from response"
        echo "Response: $CREATE_RESPONSE"
    fi
else
    echo "❌ NFC-e creation failed"
    echo "$CREATE_RESPONSE"
    exit 1
fi

# Test 3: Check status
if [ -n "$NFCE_ID" ] && [ "$NFCE_ID" != "null" ]; then
    echo ""
    echo "📊 Checking NFC-e status..."

    MAX_ATTEMPTS=30
    ATTEMPT=1

    while [ $ATTEMPT -le $MAX_ATTEMPTS ]; do
        STATUS_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$API_URL/api/v1/nfce/$NFCE_ID")
        HTTP_STATUS=$(echo "$STATUS_RESPONSE" | tail -n1)

        if echo "$HTTP_STATUS" | grep -q "200"; then
            STATUS=$(echo "$STATUS_RESPONSE" | head -n -1 | jq -r '.status' 2>/dev/null || echo "unknown")

            case $STATUS in
                "authorized")
                    echo "✅ NFC-e authorized!"
                    CHAVE=$(echo "$STATUS_RESPONSE" | head -n -1 | jq -r '.chave_acesso' 2>/dev/null || echo "")
                    if [ -n "$CHAVE" ]; then
                        echo "🔑 Chave de acesso: $CHAVE"
                    fi
                    break
                    ;;
                "rejected")
                    echo "❌ NFC-e rejected"
                    REASON=$(echo "$STATUS_RESPONSE" | head -n -1 | jq -r '.xmotivo' 2>/dev/null || echo "")
                    if [ -n "$REASON" ]; then
                        echo "📝 Reason: $REASON"
                    fi
                    exit 1
                    ;;
                "contingency")
                    echo "⚠️  NFC-e issued in contingency mode"
                    break
                    ;;
                "processing"|"pending"|"retrying")
                    echo "⏳ NFC-e status: $STATUS (attempt $ATTEMPT/$MAX_ATTEMPTS)"
                    ;;
                *)
                    echo "📋 NFC-e status: $STATUS"
                    ;;
            esac
        else
            echo "❌ Status check failed (HTTP $HTTP_STATUS)"
            exit 1
        fi

        sleep 2
        ATTEMPT=$((ATTEMPT + 1))
    done

    if [ $ATTEMPT -gt $MAX_ATTEMPTS ]; then
        echo "⏰ Timeout waiting for NFC-e processing"
        exit 1
    fi

    # Test 4: Download files (if authorized)
    if [ "$STATUS" = "authorized" ] || [ "$STATUS" = "contingency" ]; then
        echo ""
        echo "📁 Testing file downloads..."

        # Test XML download
        XML_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$API_URL/api/v1/nfce/$NFCE_ID/xml")
        if echo "$XML_RESPONSE" | grep -q "HTTP_STATUS:200"; then
            echo "✅ XML download successful"
        else
            echo "❌ XML download failed"
        fi

        # Test PDF download
        PDF_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$API_URL/api/v1/nfce/$NFCE_ID/pdf")
        if echo "$PDF_RESPONSE" | grep -q "HTTP_STATUS:200"; then
            echo "✅ PDF download successful"
        else
            echo "❌ PDF download failed"
        fi

        # Test QR Code download
        QR_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$API_URL/api/v1/nfce/$NFCE_ID/qrcode")
        if echo "$QR_RESPONSE" | grep -q "HTTP_STATUS:200"; then
            echo "✅ QR Code download successful"
        else
            echo "❌ QR Code download failed"
        fi
    fi
fi

echo ""
echo "🎉 Test completed successfully!"
echo ""
echo "💡 Useful commands:"
echo "  - View API logs: docker-compose logs api -f"
echo "  - View worker logs: docker-compose logs worker -f"
echo "  - View queue status: docker-compose exec rabbitmq rabbitmqctl list_queues"
echo "  - View MinIO files: open http://localhost:9001 (login: minioadmin/minioadmin)"
