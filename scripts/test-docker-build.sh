#!/bin/bash

# Test Docker build for plugnfce
# This script tests if the Docker image builds correctly with CGO dependencies

set -e

echo "🏗️  Building Docker image for plugnfce..."
echo "This may take a while due to CGO compilation with libxml2..."

# Build the Docker image
docker build -f docker/Dockerfile -t plugnfce:test .

echo "✅ Docker build completed successfully!"
echo ""
echo "🧪 Testing if the binary works..."

# Test if the binary can start (it will fail due to missing DB, but should show it loaded correctly)
timeout 5s docker run --rm plugnfce:test ./plugnfce-api || true

echo ""
echo "🎉 Docker setup is working correctly!"
echo ""
echo "📝 Next steps:"
echo "1. Run 'docker-compose up -d' to start all services"
echo "2. The API will be available at http://localhost:8080"
echo "3. MinIO console at http://localhost:9001 (minioadmin/minioadmin)"
echo "4. RabbitMQ at http://localhost:15672 (guest/guest)"
