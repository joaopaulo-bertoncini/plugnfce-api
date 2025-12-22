package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/config"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/di"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/logger"
)

func main() {
	// Inicializar contexto
	ctx := context.Background()
	// Inicializar logger
	l := logger.NewZapLogger()
	l.Info("Starting NFC-e Worker...")

	// Load configuration
	cfg, err := config.InitConfig()
	if err != nil {
		l.Error("Failed to load configuration", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
	}

	// Init dependency injection
	worker, err := di.InitializeWorkerManual(ctx, cfg, l)
	if err != nil {
		l.Error("Failed to initialize worker", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
	}

	// Setup graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Start worker
	go func() {
		if err := worker.Start(ctx); err != nil {
			l.Error("Worker failed", logger.Field{Key: "error", Value: err.Error()})
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-shutdown
	l.Info("Shutting down worker...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := worker.Stop(shutdownCtx); err != nil {
		l.Error("Error during worker shutdown", logger.Field{Key: "error", Value: err.Error()})
		os.Exit(1)
	}

	l.Info("Worker shutdown complete")
}
