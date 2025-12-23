# 🛠️ Guia de Desenvolvimento - NFC-e API

## Visão Geral

Este guia ajuda desenvolvedores a entenderem a estrutura do código, padrões de desenvolvimento e como contribuir para o projeto NFC-e.

## 📁 Estrutura do Projeto

```
plugnfce/
├── cmd/                    # Aplicações executáveis
│   ├── api/               # Servidor HTTP REST API
│   └── worker/            # Worker de processamento assíncrono
├── internal/              # Código privado da aplicação
│   ├── application/       # Camada de aplicação (Use Cases)
│   │   ├── dto/          # Data Transfer Objects
│   │   ├── mapper/       # Mapeadores de dados
│   │   └── usecase/      # Casos de uso da aplicação
│   ├── domain/           # Camada de domínio (Regras de negócio)
│   │   ├── entity/       # Entidades de negócio
│   │   ├── ports/        # Interfaces (Ports & Adapters)
│   │   └── service/      # Serviços de domínio
│   ├── infrastructure/   # Camada de infraestrutura
│   │   ├── database/     # PostgreSQL repositories
│   │   ├── http/         # HTTP handlers e server
│   │   ├── messaging/    # RabbitMQ clients
│   │   ├── sefaz/        # SEFAZ integration
│   │   └── worker/       # Worker orchestration
│   └── di/               # Dependency injection (Wire)
├── pkg/                  # Código compartilhado público
│   ├── database/         # Database utilities
│   └── logger/           # Logging utilities
├── migrations/           # Database migrations
├── scripts/              # Scripts de desenvolvimento
├── docker/               # Docker configuration
├── docs/                 # Documentação
└── internal/infrastructure/sefaz/schemas/  # XSD schemas
```

## 🏗️ Arquitetura - Clean Architecture

### Princípios Seguidos

1. **Separação de Responsabilidades**: Cada camada tem uma responsabilidade clara
2. **Injeção de Dependência**: Interfaces definem contratos, implementações são injetadas
3. **Ports & Adapters**: Domínio não depende de infraestrutura
4. **Testabilidade**: Código isolado facilita testes unitários

### Fluxo de Dependências

```
Infrastructure ──► Application ──► Domain
     ▲                │              │
     └────────────────┴──────────────┘
          (Dependency Injection)
```

## 🚀 Início Rápido

### Pré-requisitos
- Go 1.24+
- Docker & Docker Compose
- Git
- Make (opcional)

### Configuração do Ambiente

```bash
# 1. Clonar repositório
git clone <repository-url>
cd plugnfce

# 2. Configurar ambiente
cp env.example .env

# 3. Iniciar serviços
./scripts/docker-dev.sh up

# 4. Verificar saúde
curl http://localhost:8080/health
```

## 🧪 Desenvolvimento e Testes

### Executando Testes

```bash
# Todos os testes
go test ./...

# Testes de uma package específica
go test ./internal/domain/service/...

# Testes com coverage
go test -cover ./...

# Testes de integração (com Docker)
./scripts/docker-dev.sh test
```

### Debug e Desenvolvimento

```bash
# Executar API localmente
go run cmd/api/main.go

# Executar Worker localmente
go run cmd/worker/main.go

# Ver logs em tempo real
./scripts/docker-dev.sh logs api

# Acessar container para debug
./scripts/docker-dev.sh shell api
```

## 📝 Padrões de Código

### Go Standards
- Seguir [Effective Go](https://golang.org/doc/effective_go.html)
- Usar `gofmt` para formatação
- Imports organizados: standard → third-party → internal

### Estrutura de Arquivos

#### Handlers HTTP
```go
// internal/infrastructure/http/handler/nfce.go
type NFCeHandler struct {
    usecase usecase.NFCeUseCase
}

func (h *NFCeHandler) EmitNFce(c *gin.Context) {
    // 1. Parse request
    // 2. Validate input
    // 3. Call use case
    // 4. Return response
}
```

#### Use Cases
```go
// internal/application/usecase/nfce.go
type NFCeUseCase struct {
    repo    ports.NFCeRepository
    publisher rabbitmq.Publisher
}

func (uc *NFCeUseCase) EmitNFce(ctx context.Context, key string, req dto.EmitNFceRequest) (*dto.NFceResponse, error) {
    // 1. Validate idempotency
    // 2. Create entity
    // 3. Persist to database
    // 4. Publish to queue
    // 5. Return response
}
```

#### Domain Services
```go
// internal/domain/service/worker.go
type NFCeWorkerService struct {
    xmlBuilder  nfce.Builder
    xmlSigner   signer.Signer
    validator   validator.XMLValidator
    soapClient  soapclient.Client
    qrGenerator qr.Generator
}

func (s *NFCeWorkerService) ProcessNFceEmission(ctx context.Context, nfce *entity.NFCE) error {
    // 1. Validate idempotency
    // 2. Build XML
    // 3. Validate XSD
    // 4. Sign XML
    // 5. Send to SEFAZ
    // 6. Update status
}
```

### Repositories
```go
// internal/infrastructure/database/postgres/nfce.go
type nfceRepository struct {
    db *gorm.DB
}

func (r *nfceRepository) Create(ctx context.Context, req *entity.NFCE) error {
    req.ID = uuid.New().String()
    req.CreatedAt = time.Now()
    return r.db.WithContext(ctx).Create(req).Error
}
```

## 🔄 Dependency Injection

### Usando Wire

O projeto usa [Wire](https://github.com/google/wire) para DI.

**Providers são definidos em:**
```go
// internal/di/providers.go
func InitializeAPIManual(ctx context.Context, cfg *config.AppConfig, l logger.Logger) (*server.Server, error) {
    // Manual dependency injection for development
}
```

**Wire generation:**
```go
// internal/di/wire.go
//go:build wireinject

func InitializeAPI(ctx context.Context, cfg *config.AppConfig, l logger.Logger) (*server.Server, error) {
    wire.Build(/* ... */)
    return &server.Server{}, nil
}
```

**Regenerar após mudanças:**
```bash
go generate ./internal/di
```

## 🧪 Testes

### Estrutura de Testes

```
internal/domain/service/
├── worker.go
└── worker_test.go

internal/application/usecase/
├── nfce.go
└── nfce_test.go
```

### Exemplo de Teste Unitário

```go
func TestNFCeWorkerService_ProcessNFceEmission(t *testing.T) {
    // Arrange
    mockBuilder := &mocks.Builder{}
    mockSigner := &mocks.Signer{}
    // ... setup mocks

    service := service.NewNFCeWorkerService(
        mockBuilder,
        mockSigner,
        // ...
    )

    // Act
    err := service.ProcessNFceEmission(context.Background(), nfce)

    // Assert
    assert.NoError(t, err)
    mockBuilder.AssertExpectations(t)
}
```

### Testes de Integração

```go
func TestNFCeAPI_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test database
    // Make HTTP requests
    // Assert responses
}
```

## 🐛 Debugging

### Logs Estruturados

```go
// Info com campos estruturados
w.logger.Info("Processing NFC-e request",
    logger.Field{Key: "request_id", Value: requestID},
    logger.Field{Key: "status", Value: status})

// Error com contexto
w.logger.Error("Failed to process NFC-e",
    logger.Field{Key: "request_id", Value: requestID},
    logger.Field{Key: "error", Value: err.Error()})
```

### Debug Mode

```bash
# Executar com debug
DEBUG=1 go run cmd/api/main.go

# Ver logs detalhados
./scripts/docker-dev.sh logs api --tail 100 -f
```

## 🔒 Segurança

### Certificado Digital
- Certificado A1 (PFX) descriptografado apenas em memória
- Senha nunca logada ou armazenada
- Certificado válido para NFC-e

### Validações
- Idempotency-Key para evitar duplicatas
- Validação de entrada em todos os endpoints
- Sanitização de dados

### Headers de Segurança
```go
// CORS, Rate Limiting, etc.
router.Use(gin.Recovery())
router.Use(gin.Logger())
router.Use(middleware.CORS())
router.Use(middleware.RateLimit())
```

## 🚀 Deployment

### Build Otimizado

```bash
# Build com otimizações
CGO_ENABLED=1 GOOS=linux go build \
    -a -installsuffix cgo \
    -o bin/plugnfce-api \
    ./cmd/api

# Imagem Docker multi-stage
docker build -f docker/Dockerfile -t plugnfce:latest .
```

### Variáveis de Produção

```bash
# Database
DB_HOST=postgres-prod
DB_USER=prod_user
DB_PASSWORD=${DB_PASSWORD}

# RabbitMQ
RABBITMQ_HOST=rabbitmq-prod
RABBITMQ_USER=prod_user
RABBITMQ_PASSWORD=${RABBITMQ_PASSWORD}

# Environment
ENV=production
LOG_LEVEL=info
```

## 📊 Monitoramento

### Métricas Implementadas
- Tempo de resposta da API
- Taxa de sucesso de emissão
- Latência da SEFAZ
- Uso de recursos do Worker

### Health Checks
```bash
# API health
GET /health

# Database connectivity
GET /health/db

# Queue status
GET /health/queue
```

## 🤝 Contribuição

### Pull Request Process
1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Add nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

### Code Review Checklist
- [ ] Testes passando
- [ ] Lint passando (`golangci-lint`)
- [ ] Documentação atualizada
- [ ] Migration scripts (se aplicável)
- [ ] Breaking changes documentados

### Commits Convention
```
feat: add NFC-e cancellation endpoint
fix: resolve XML validation issue
docs: update API documentation
test: add integration tests for worker
refactor: improve error handling in repository
```

## 📚 Recursos Adicionais

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Wire Documentation](https://github.com/google/wire)
- [GORM Documentation](https://gorm.io/)
- [Gin Web Framework](https://gin-gonic.com/)

---

**Mantido por**: Equipe NFC-e
**Última atualização**: Dezembro 2024
