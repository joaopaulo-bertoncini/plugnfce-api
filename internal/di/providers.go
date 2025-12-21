package di

import (
	"context"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/usecase"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/config"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/database/postgres"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/http/handler"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/http/server"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/messaging/rabbitmq"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/logger"
)

// InitializeAPIManual initializes the entire API application manually (alternative to wire)
func InitializeAPIManual(ctx context.Context, cfg *config.AppConfig, l logger.Logger) (*server.Server, error) {
	// Initialize database
	db, err := postgres.NewDatabase(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	nfceRepo := postgres.NewNFCeRepository(db)
	companyRepo := postgres.NewCompanyRepository(db)
	planRepo := postgres.NewPlanRepository(db)
	subscriptionRepo := postgres.NewSubscriptionRepository(db)
	webhookRepo := postgres.NewWebhookRepository(db)

	// Initialize publisher
	publisher, err := rabbitmq.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		return nil, err
	}

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
