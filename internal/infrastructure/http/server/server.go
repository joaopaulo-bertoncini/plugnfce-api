package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/http/handler"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/http/router"
	"github.com/joaopaulo-bertoncini/plugnfce-api/pkg/logger"
)

// Server represents the HTTP server
type Server struct {
	engine *gin.Engine
	port   string
	logger logger.Logger
	server *http.Server
}

// NewServer creates a new HTTP server
func NewServer(
	nfceHandler *handler.NFCeHandler,
	adminHandler *handler.AdminHandler,
	companyHandler *handler.CompanyHandler,
	planHandler *handler.PlanHandler,
	subscriptionHandler *handler.SubscriptionHandler,
	webhookHandler *handler.WebhookHandler,
	logger logger.Logger,
	port string,
) *Server {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Setup routes
	engine := router.SetupRoutes(
		nfceHandler,
		adminHandler,
		companyHandler,
		planHandler,
		subscriptionHandler,
		webhookHandler,
	)

	return &Server{
		engine: engine,
		port:   port,
		logger: logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%s", s.port),
		Handler:      s.engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for interrupt signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		s.logger.Info("Starting HTTP server", logger.Field{Key: "port", Value: s.port})
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Failed to start server", logger.Field{Key: "error", Value: err.Error()})
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-done
	s.logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server forced to shutdown", logger.Field{Key: "error", Value: err.Error()})
		return err
	}

	s.logger.Info("Server exited")
	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
