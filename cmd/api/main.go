package main

import (
	"context"
	"os"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/config"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/di"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/logger"
)

func main() {
	// Inicializar contexto
	ctx := context.Background()
	// Inicializar logger
	l := logger.NewZapLogger()
	l.Info("Starting NFC-e API...")

	// Load configuration
	cfg, err := config.InitConfig()
	if err != nil {
		l.Error("Failed to load configuration", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
	}

	// Init dependency injection
	server, err := di.InitializeAPI(ctx, cfg, l)
	if err != nil {
		l.Error("Failed to initialize application", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
	}

	// Start server
	if err := server.Start(ctx); err != nil {
		l.Error("Server failed", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
	}
}
