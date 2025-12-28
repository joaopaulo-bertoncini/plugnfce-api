//go:build wireinject
// +build wireinject

package di

import (
	"context"
	"fmt"
	"time"

	"github.com/google/wire"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/usecase"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/config"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/domain/service"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/database/postgres"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/http/handler"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/http/server"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/messaging/rabbitmq"
	nfceInfra "github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/nfce"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/qr"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/signer"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/soap/soapclient"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/sefaz/validator"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/storage"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/worker"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/database"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/logger"
	"gorm.io/gorm"
)

// InitializeAPI initializes the entire API application with dependency injection
func InitializeAPI(ctx context.Context, cfg *config.AppConfig, l logger.Logger) (*server.Server, error) {
	wire.Build(
		// Infrastructure
		provideDatabase,
		postgres.NewNFCeRepository,
		postgres.NewCompanyRepository,
		postgres.NewPlanRepository,
		postgres.NewSubscriptionRepository,
		postgres.NewWebhookRepository,
		providePublisher,
		providePort,
		server.NewServer,

		// Application
		provideStorage,
		usecase.NewNFCeUseCase,
		usecase.NewAdminUseCase,
		usecase.NewCompanyUseCase,
		usecase.NewPlanUseCase,
		usecase.NewSubscriptionUseCase,
		usecase.NewWebhookUseCase,

		// HTTP
		handler.NewNFCeHandler,
		handler.NewAdminHandler,
		handler.NewCompanyHandler,
		handler.NewPlanHandler,
		handler.NewSubscriptionHandler,
		handler.NewWebhookHandler,
	)
	return &server.Server{}, nil
}

// InitializeWorker initializes the worker with dependency injection
func InitializeWorker(ctx context.Context, cfg *config.AppConfig, l logger.Logger) (*worker.Worker, error) {
	wire.Build(
		// Infrastructure
		provideDatabase,
		postgres.NewNFCeRepository,
		providePublisher,
		provideConsumer,
		provideXMLBuilder,
		provideXMLSigner,
		provideXMLValidator,
		provideSOAPClient,
		provideQRGenerator,
		provideStorage,
		service.NewNFCeWorkerService,
		worker.NewWorker,
		provideMaxRetries,
	)
	return &worker.Worker{}, nil
}

// provideDatabase provides database instance
func provideDatabase(cfg *config.AppConfig) (*gorm.DB, error) {
	// Initialize database if not already initialized
	if database.GetDB() == nil {
		ctx := context.Background()
		if err := database.InitDatabase(ctx, cfg.GetDatabaseDSN(), cfg.Env); err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
	}
	return database.GetDB(), nil
}

// providePublisher provides RabbitMQ publisher
func providePublisher(cfg *config.AppConfig) (dto.Publisher, error) {
	publisher, err := rabbitmq.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		return nil, err
	}
	return dto.Publisher(publisher), nil
}

// providePort provides the server port
func providePort(cfg *config.AppConfig) string {
	return cfg.Port
}

// provideConsumer provides RabbitMQ consumer
func provideConsumer(cfg *config.AppConfig) (dto.Consumer, error) {
	consumer, err := rabbitmq.NewConsumer(cfg.RabbitMQURL)
	if err != nil {
		return nil, err
	}
	return dto.Consumer(consumer), nil
}

// provideXMLBuilder provides XML builder
func provideXMLBuilder(db *gorm.DB) nfceInfra.Builder {
	companyRepo := postgres.NewCompanyRepository(db)
	return nfceInfra.NewBuilder(companyRepo)
}

// provideXMLSigner provides XML signer
func provideXMLSigner() signer.Signer {
	return signer.NewSigner()
}

// provideXMLValidator provides XML validator
func provideXMLValidator() (validator.XMLValidator, error) {
	return validator.NewXMLValidator("./internal/infrastructure/sefaz/schemas")
}

// provideSOAPClient provides SOAP client
func provideSOAPClient() soapclient.Client {
	return soapclient.NewSOAPClient(30 * time.Second)
}

// provideQRGenerator provides QR code generator
func provideQRGenerator() qr.Generator {
	return qr.NewGenerator()
}

// provideMaxRetries provides max retry count
func provideMaxRetries() int {
	return 5
}

// provideWorkerCount provides worker count
func provideWorkerCount() int {
	return 3
}

// provideStorage provides storage service
func provideStorage(cfg *config.AppConfig) (storage.StorageService, error) {
	switch cfg.StorageType {
	case "minio":
		return storage.NewMinIOStorage(
			cfg.StorageEndpoint,
			cfg.StorageAccessKey,
			cfg.StorageSecretKey,
			cfg.StorageBucket,
			cfg.StorageUseSSL,
		)
	case "local":
		return storage.NewLocalStorage(
			cfg.StorageBasePath,
			cfg.StoragePublicURL,
			cfg.StorageBucket,
		)
	default:
		return storage.NewLocalStorage(
			cfg.StorageBasePath,
			cfg.StoragePublicURL,
			cfg.StorageBucket,
		)
	}
}
