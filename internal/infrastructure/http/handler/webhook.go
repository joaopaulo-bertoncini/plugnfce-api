package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/usecase"
)

// WebhookHandler manages HTTP requests related to webhook operations
type WebhookHandler struct {
	webhookUseCase usecase.WebhookUseCase
}

// NewWebhookHandler creates a new WebhookHandler
func NewWebhookHandler(webhookUseCase usecase.WebhookUseCase) *WebhookHandler {
	return &WebhookHandler{
		webhookUseCase: webhookUseCase,
	}
}

// Create creates a new webhook
func (h *WebhookHandler) Create(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set the company ID from the context
	req.CompanyID = companyID

	webhook, err := h.webhookUseCase.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

// GetByID gets a webhook by ID
func (h *WebhookHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook ID is required"})
		return
	}

	webhook, err := h.webhookUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// List lists webhooks for the authenticated company
func (h *WebhookHandler) List(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	response, err := h.webhookUseCase.List(c.Request.Context(), companyID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   response.Webhooks,
		"total":  response.Total,
		"limit":  limit,
		"offset": offset,
	})
}

// Update updates a webhook
func (h *WebhookHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook ID is required"})
		return
	}

	var req dto.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.webhookUseCase.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook updated successfully"})
}

// Delete deletes a webhook
func (h *WebhookHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook ID is required"})
		return
	}

	err := h.webhookUseCase.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook deleted successfully"})
}
