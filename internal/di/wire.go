//go:build wireinject
// +build wireinject

package di

import (
	"context"

	"github.com/google/wire"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/usecase"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/config"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/database/postgres"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/http/handler"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/http/server"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/messaging/rabbitmq"
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

// provideDatabase provides database instance
func provideDatabase(ctx context.Context, cfg *config.AppConfig) (*gorm.DB, error) {
	return postgres.NewDatabase(ctx, cfg)
}

// providePublisher provides RabbitMQ publisher
func providePublisher(cfg *config.AppConfig) (rabbitmq.Publisher, error) {
	return rabbitmq.NewPublisher(cfg.RabbitMQURL)
}

// providePort provides the server port
func providePort(cfg *config.AppConfig) string {
	return cfg.Port
}
