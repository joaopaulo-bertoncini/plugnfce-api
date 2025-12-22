# 🐳 NFC-e API - Docker Development Environment

Este documento explica como usar o ambiente de desenvolvimento Docker para a API NFC-e.

## 📋 Pré-requisitos

- Docker 20.10+
- Docker Compose 2.0+
- Pelo menos 4GB de RAM disponível
- Pelo menos 10GB de espaço em disco

## 🚀 Início Rápido

### 1. Clonar e configurar

```bash
git clone <repository-url>
cd plugnfce

# Copiar arquivo de configuração de exemplo
cp env.example .env
```

### 2. Iniciar todos os serviços

```bash
# Usando o script helper
./scripts/docker-dev.sh up

# Ou diretamente com docker-compose
docker-compose up -d
```

### 3. Aguardar inicialização

Os serviços levarão alguns segundos para ficarem prontos. Você pode verificar o status com:

```bash
./scripts/docker-dev.sh status
```

## 🌐 Serviços Disponíveis

Após iniciar, os seguintes serviços estarão disponíveis:

| Serviço | URL | Descrição |
|---------|-----|-----------|
| **API NFC-e** | http://localhost:8080 | API REST principal |
| **MinIO Console** | http://localhost:9001 | Interface web do MinIO (user: `minioadmin`, pass: `minioadmin`) |
| **RabbitMQ Management** | http://localhost:15672 | Interface de gerenciamento do RabbitMQ (user: `guest`, pass: `guest`) |
| **PostgreSQL** | localhost:5432 | Banco de dados (user: `plugnfce`, pass: `plugnfce`, db: `plugnfce`) |

## 🛠️ Comandos Úteis

### Gerenciamento de Serviços

```bash
# Ver logs de todos os serviços
./scripts/docker-dev.sh logs

# Ver logs específicos
./scripts/docker-dev.sh logs api
./scripts/docker-dev.sh logs worker
./scripts/docker-dev.sh logs db

# Acessar shell de um container
./scripts/docker-dev.sh shell api
./scripts/docker-dev.sh shell worker
./scripts/docker-dev.sh shell db

# Parar todos os serviços
./scripts/docker-dev.sh down

# Limpar tudo (containers, volumes, imagens)
./scripts/docker-dev.sh clean
```

### Desenvolvimento

```bash
# Reconstruir e reiniciar serviços
./scripts/docker-dev.sh rebuild

# Executar testes
./scripts/docker-dev.sh test

# Verificar status dos serviços
./scripts/docker-dev.sh status
```

## 🏗️ Arquitetura dos Containers

### API Container
- **Base**: Alpine Linux com Go 1.24
- **Porta**: 8080
- **Responsabilidades**:
  - Receber requisições HTTP REST
  - Validar entrada e idempotência
  - Publicar mensagens na fila RabbitMQ
  - Persistir estado inicial no PostgreSQL

### Worker Container
- **Base**: Mesma imagem da API
- **Responsabilidades**:
  - Consumir mensagens da fila RabbitMQ
  - Processar emissão de NFC-e
  - Validar XML contra schemas XSD
  - Assinar digitalmente
  - Comunicar com SEFAZ via SOAP
  - Gerar QR Code
  - Atualizar status no banco

### Infraestrutura

- **PostgreSQL**: Persistência de dados e eventos
- **RabbitMQ**: Fila de mensagens assíncronas
- **MinIO**: Armazenamento de arquivos (XML, PDF)
- **Redis**: Cache (opcional, não usado atualmente)

## 🔧 Configuração

### Variáveis de Ambiente

As principais variáveis estão definidas no `docker-compose.yml`. Para personalizar:

1. Copie `env.example` para `.env`
2. Edite as variáveis necessárias
3. Reinicie os serviços: `./scripts/docker-dev.sh rebuild`

### Volumes Persistentes

- `db_data`: Dados do PostgreSQL
- `minio_data`: Arquivos armazenados no MinIO
- `rabbitmq_data`: Configurações e filas do RabbitMQ
- `redis_data`: Dados do Redis

## 🧪 Testes

### Testes Unitários

```bash
# Dentro do container da API
./scripts/docker-dev.sh shell api
go test ./...
```

### Testes de Integração

```bash
# Testar comunicação entre serviços
./scripts/docker-dev.sh test
```

### Teste Manual da API

```bash
# Health check
curl http://localhost:8080/health

# Listar NFC-e (se implementado)
curl http://localhost:8080/nfce
```

## 🐛 Troubleshooting

### Serviços não iniciam
```bash
# Verificar logs detalhados
./scripts/docker-dev.sh logs

# Verificar recursos do sistema
docker system df
```

### API retorna erro de conexão
```bash
# Verificar se todos os serviços estão saudáveis
./scripts/docker-dev.sh status

# Aguardar mais tempo para inicialização completa
sleep 30 && ./scripts/docker-dev.sh logs api
```

### Problemas de memória
```bash
# Limpar recursos não utilizados
./scripts/docker-dev.sh clean

# Reiniciar Docker daemon se necessário
sudo systemctl restart docker
```

## 📊 Monitoramento

### Logs em Tempo Real
```bash
# Todos os logs
./scripts/docker-dev.sh logs

# Apenas erros
./scripts/docker-dev.sh logs | grep -i error
```

### Métricas dos Containers
```bash
# Uso de recursos
docker stats

# Logs do RabbitMQ (mensagens processadas)
./scripts/docker-dev.sh logs rabbitmq
```

## 🚀 Deploy em Produção

Este setup é otimizado para desenvolvimento. Para produção:

1. Use imagens específicas de versão
2. Configure secrets adequadamente
3. Adicione healthchecks mais robustos
4. Configure limites de recursos
5. Use Docker Swarm ou Kubernetes
6. Configure backups automáticos
7. Adicione monitoring (Prometheus/Grafana)

## 📚 Recursos Adicionais

- [Documentação da API](./docs/api.md)
- [Arquitetura do Sistema](./ARQUITETURA-NFCE-GO.md)
- [Guia de Desenvolvimento](./docs/development.md)
