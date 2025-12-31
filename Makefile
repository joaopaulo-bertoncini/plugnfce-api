# Makefile para ImobCheck API

# Variáveis
BINARY_API=plugnfce-api
BUILD_DIR_API=/bin
MAIN_FILE_API=cmd/api/main.go

BINARY_WORKER=plugnfce-worker
BUILD_DIR_WORKER=/bin
MAIN_FILE_WORKER=cmd/worker/main.go

# Database connection variables (use defaults or override from env file)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= plugnfce
DB_PASSWORD ?= plugnfce
DB_NAME ?= plugnfce
DB_SSL_MODE ?= disable

# Migrate command path
MIGRATE_CMD = $(HOME)/go/bin/migrate

# Comandos principais
.PHONY: build run test clean deps migrate

# Construir a aplicação
build-api:
	@echo "Building plugnfce-api..."
	@go build -o $(BUILD_DIR_API)/$(BINARY_API) $(MAIN_FILE_API)
	@echo "Build completed: $(BUILD_DIR_API)/$(BINARY_API)"

build-worker:
	@echo "Building plugnfce-worker..."
	@go build -o $(BUILD_DIR_WORKER)/$(BINARY_WORKER) $(MAIN_FILE_WORKER)
	@echo "Build completed: $(BUILD_DIR_WORKER)/$(BINARY_WORKER)"

# Executar API
run-api:
	@echo "Running API..."
	@go run $(MAIN_FILE_API)

# Executar Worker
run-worker:
	@echo "Running Worker..."
	@go run $(MAIN_FILE_WORKER)

# Executar em modo de desenvolvimento
dev:
	@echo "Running in development mode..."
	@ENV=development go run $(MAIN_FILE_API)

# Executar sem banco de dados (para testes)
dev-no-db:
	@echo "Running in development mode without database..."
	@ENV=development SKIP_DB=true go run $(MAIN_FILE_API)

# Executar testes
test:
	@echo "Running tests..."
	@go test -v ./...

# Executar testes com coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Limpar arquivos de build
clean:
	@echo "Cleaning build files..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean

# Instalar dependências
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Executar migrações do banco
migrate:
	@echo "Running database migrations..."
	@$(MIGRATE_CMD) -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" up

# Reverter migrações
migrate-down:
	@echo "Rolling back database migrations..."
	@$(MIGRATE_CMD) -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)" down

# Criar nova migração
migrate-create:
	@echo "Creating new migration..."
	@read -p "Enter migration name: " name; \
	$(MIGRATE_CMD) create -ext sql -dir migrations -seq $$name

# Executar linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Formatar código
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Verificar imports
imports:
	@echo "Checking imports..."
	@goimports -w .

# Executar com Docker
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_API) -t $(BINARY_WORKER) .

docker-run:
	@echo "Running with Docker..."
	@docker run -p 8080:8080 --env-file env $(BINARY_API) $(BINARY_WORKER)

# Docker Compose
docker-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

docker-up-all:
	@echo "Starting all services with Docker Compose..."
	@docker-compose --profile api up -d

docker-down:
	@echo "Stopping Docker Compose services..."
	@docker-compose down

docker-logs:
	@echo "Showing Docker Compose logs..."
	@docker-compose logs -f

# Desenvolvimento com Docker
dev-docker:
	@echo "Starting development environment with Docker..."
	@docker-compose up -d postgres
	@sleep 5
	@make dev

# Testar API
test-api:
	@echo "Testing API endpoints..."
	@./scripts/test_api.sh

# Ajuda
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  run-api       - Run API server"
	@echo "  run-worker    - Run worker process"
	@echo "  dev           - Run in development mode"
	@echo "  dev-no-db     - Run without database"
	@echo "  dev-docker    - Run with Docker PostgreSQL"
	@echo ""
	@echo "Testing:"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  test-api      - Test API endpoints"
	@echo ""
	@echo "Database:"
	@echo "  migrate       - Run database migrations"
	@echo "  migrate-down  - Rollback migrations"
	@echo "  migrate-create- Create new migration"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  imports       - Check imports"
	@echo "  deps          - Install dependencies"
	@echo "  clean         - Clean build files"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker"
	@echo "  docker-up     - Start PostgreSQL with Docker"
	@echo "  docker-up-all - Start all services with Docker"
	@echo "  docker-down   - Stop Docker services"
	@echo "  docker-logs   - Show Docker logs"
	@echo ""
	@echo "  help          - Show this help"
