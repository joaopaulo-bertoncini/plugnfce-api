package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/usecase"
)

// SubscriptionHandler manages HTTP requests related to subscription operations
type SubscriptionHandler struct {
	subscriptionUseCase usecase.SubscriptionUseCase
}

// NewSubscriptionHandler creates a new SubscriptionHandler
func NewSubscriptionHandler(subscriptionUseCase usecase.SubscriptionUseCase) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionUseCase: subscriptionUseCase,
	}
}

// Create creates a new subscription
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req dto.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.subscriptionUseCase.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetByID gets a subscription by ID
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription ID is required"})
		return
	}

	subscription, err := h.subscriptionUseCase.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// GetCurrent gets the current active subscription for the authenticated company
func (h *SubscriptionHandler) GetCurrent(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	subscription, err := h.subscriptionUseCase.GetCurrent(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// List lists subscriptions with pagination
func (h *SubscriptionHandler) List(c *gin.Context) {
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

	response, err := h.subscriptionUseCase.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   response.Subscriptions,
		"total":  response.Total,
		"limit":  limit,
		"offset": offset,
	})
}

// Update updates a subscription
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription ID is required"})
		return
	}

	var req dto.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.subscriptionUseCase.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription updated successfully"})
}

// Cancel cancels a subscription
func (h *SubscriptionHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription ID is required"})
		return
	}

	var req dto.CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.subscriptionUseCase.Cancel(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription canceled successfully"})
}

// GetUsage gets the usage statistics for a company's current subscription
func (h *SubscriptionHandler) GetUsage(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	usage, err := h.subscriptionUseCase.GetUsage(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usage)
}
