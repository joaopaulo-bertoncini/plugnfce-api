package router

import (
	"github.com/gin-gonic/gin"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/infrastructure/http/handler"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	nfceHandler *handler.NFCeHandler,
	adminHandler *handler.AdminHandler,
	companyHandler *handler.CompanyHandler,
	planHandler *handler.PlanHandler,
	subscriptionHandler *handler.SubscriptionHandler,
	webhookHandler *handler.WebhookHandler,
) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public NFC-e endpoints
		nfce := v1.Group("/nfce")
		{
			nfce.POST("", nfceHandler.EmitNFce)
			nfce.GET("/:id", nfceHandler.GetNFceByID)
			nfce.POST("/:id/cancel", nfceHandler.CancelNFce)
			nfce.GET("/:id/events", nfceHandler.GetNFceEvents)
		}

		// Company endpoints (for authenticated companies)
		companies := v1.Group("/companies")
		if companyHandler != nil {
			companies.GET("/profile", companyHandler.GetProfile)
			companies.PUT("/profile", companyHandler.UpdateProfile)
			companies.PUT("/:id/certificate", companyHandler.UpdateCertificateByID)
			companies.PUT("/certificate", companyHandler.UpdateCertificate)
			companies.PUT("/csc", companyHandler.UpdateCSC)
		}

		// Subscription endpoints (for authenticated companies)
		subscriptions := v1.Group("/subscriptions")
		if subscriptionHandler != nil {
			subscriptions.GET("/current", subscriptionHandler.GetCurrent)
			subscriptions.GET("/usage", subscriptionHandler.GetUsage)
		}

		// Webhook endpoints (for authenticated companies)
		webhooks := v1.Group("/webhooks")
		if webhookHandler != nil {
			webhooks.POST("", webhookHandler.Create)
			webhooks.GET("", webhookHandler.List)
			webhooks.GET("/:id", webhookHandler.GetByID)
			webhooks.PUT("/:id", webhookHandler.Update)
			webhooks.DELETE("/:id", webhookHandler.Delete)
		}
	}

	// Admin API routes
	admin := r.Group("/api/admin")
	{
		// Admin authentication (if admin handler exists)
		if adminHandler != nil {
			admin.POST("/login", adminHandler.Login)
		}

		// Company management
		companies := admin.Group("/companies")
		if adminHandler != nil {
			companies.POST("", adminHandler.CreateCompany)
			companies.GET("", adminHandler.ListCompanies)
			companies.GET("/:id", adminHandler.GetCompany)
			companies.PUT("/:id", adminHandler.UpdateCompany)
			companies.PUT("/:id/certificate", adminHandler.UpdateCompanyCertificate)
			companies.PUT("/:id/csc", adminHandler.UpdateCompanyCSC)
		}

		// Plan management
		plans := admin.Group("/plans")
		if planHandler != nil {
			plans.POST("", planHandler.Create)
			plans.GET("", planHandler.List)
			plans.GET("/:id", planHandler.GetByID)
			plans.PUT("/:id", planHandler.Update)
			plans.DELETE("/:id", planHandler.Archive)
		}

		// Subscription management
		subscriptions := admin.Group("/subscriptions")
		if subscriptionHandler != nil {
			subscriptions.POST("", subscriptionHandler.Create)
			subscriptions.GET("", subscriptionHandler.List)
			subscriptions.GET("/:id", subscriptionHandler.GetByID)
			subscriptions.PUT("/:id", subscriptionHandler.Update)
			subscriptions.DELETE("/:id", subscriptionHandler.Cancel)
		}

		// Webhook management
		webhooks := admin.Group("/webhooks")
		if webhookHandler != nil {
			webhooks.POST("", webhookHandler.Create)
			webhooks.GET("", webhookHandler.List)
			webhooks.GET("/:id", webhookHandler.GetByID)
			webhooks.PUT("/:id", webhookHandler.Update)
			webhooks.DELETE("/:id", webhookHandler.Delete)
		}

		// NFC-e management
		nfceAdmin := admin.Group("/nfce")
		if adminHandler != nil {
			nfceAdmin.GET("", adminHandler.ListNFCE)
		}

		// Statistics
		if adminHandler != nil {
			admin.GET("/stats", adminHandler.GetStats)
		}
	}

	return r
}
