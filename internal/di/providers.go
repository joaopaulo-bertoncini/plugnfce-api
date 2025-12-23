package di

import (
	"context"
	"time"

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
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/worker"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/database"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/logger"
)

// InitializeAPIManual initializes the entire API application manually (alternative to wire)
func InitializeAPIManual(ctx context.Context, cfg *config.AppConfig, l logger.Logger) (*server.Server, error) {
	// Initialize database
	err := database.InitDatabase(ctx, cfg.GetDatabaseDSN(), cfg.Env)
	if err != nil {
		return nil, err
	}
	db := database.GetDB()

	// Initialize repositories
	nfceRepo := postgres.NewNFCeRepository(db)
	companyRepo := postgres.NewCompanyRepository(db)
	planRepo := postgres.NewPlanRepository(db)
	subscriptionRepo := postgres.NewSubscriptionRepository(db)
	webhookRepo := postgres.NewWebhookRepository(db)

	// Initialize publisher
	rabbitmqPublisher, err := rabbitmq.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		return nil, err
	}
	publisher := dto.Publisher(rabbitmqPublisher)

	// Initialize use cases
	nfceUseCase := usecase.NewNFCeUseCase(nfceRepo, publisher)
	adminUseCase := usecase.NewAdminUseCase(companyRepo, planRepo, subscriptionRepo)
	companyUseCase := usecase.NewCompanyUseCase(companyRepo, subscriptionRepo)
	planUseCase := usecase.NewPlanUseCase(planRepo)
	subscriptionUseCase := usecase.NewSubscriptionUseCase(subscriptionRepo, planRepo, companyRepo)
	webhookUseCase := usecase.NewWebhookUseCase(webhookRepo)

	// Initialize handlers
	nfceHandler := handler.NewNFCeHandler(nfceUseCase)
	adminHandler := handler.NewAdminHandler(adminUseCase)
	companyHandler := handler.NewCompanyHandler(companyUseCase)
	planHandler := handler.NewPlanHandler(planUseCase)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionUseCase)
	webhookHandler := handler.NewWebhookHandler(webhookUseCase)

	// Initialize server
	srv := server.NewServer(
		nfceHandler,
		adminHandler,
		companyHandler,
		planHandler,
		subscriptionHandler,
		webhookHandler,
		l,
		cfg.Port,
	)

	return srv, nil
}

// InitializeWorkerManual initializes the worker manually
func InitializeWorkerManual(ctx context.Context, cfg *config.AppConfig, l logger.Logger) (*worker.Worker, error) {
	// Initialize database
	err := database.InitDatabase(ctx, cfg.GetDatabaseDSN(), cfg.Env)
	if err != nil {
		return nil, err
	}
	db := database.GetDB()

	// Initialize repositories
	nfceRepo := postgres.NewNFCeRepository(db)

	// Initialize messaging
	rabbitmqPublisher, err := rabbitmq.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		return nil, err
	}
	publisher := dto.Publisher(rabbitmqPublisher)

	rabbitmqConsumer, err := rabbitmq.NewConsumer(cfg.RabbitMQURL)
	if err != nil {
		return nil, err
	}
	consumer := dto.Consumer(rabbitmqConsumer)

	// Initialize SEFAZ components
	xmlBuilder := nfceInfra.NewBuilder()
	xmlSigner := signer.NewSigner()
	xmlValidator, err := validator.NewXMLValidator("./internal/infrastructure/sefaz/schemas")
	if err != nil {
		return nil, err
	}
	soapClient := soapclient.NewSOAPClient(30 * time.Second) // 30 second timeout
	qrGenerator := qr.NewGenerator()

	// Initialize domain service
	workerService := service.NewNFCeWorkerService(
		xmlBuilder,
		xmlSigner,
		xmlValidator,
		soapClient,
		qrGenerator,
	)

	// Initialize worker
	w := worker.NewWorker(
		nfceRepo,
		publisher,
		consumer,
		workerService,
		l,
		5, // max retries
	)

	return w, nil
}
