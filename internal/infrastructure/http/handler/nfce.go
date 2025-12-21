package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/usecase"
)

// NFCeHandler manages HTTP requests related to NFC-e
type NFCeHandler struct {
	nfceUseCase usecase.NFCeUseCase
}

type NFCeHandlerInterface interface {
	EmitNFce(c *gin.Context)
	GetNFceByID(c *gin.Context)
	ListNFces(c *gin.Context)
	CancelNFce(c *gin.Context)
	GetNFceEvents(c *gin.Context)
}

// NewNFCeHandler creates a new NFCeHandler
func NewNFCeHandler(nfceUseCase usecase.NFCeUseCase) *NFCeHandler {
	return &NFCeHandler{
		nfceUseCase: nfceUseCase,
	}
}

// EmitNFce emits a new NFC-e
func (h *NFCeHandler) EmitNFce(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.EmitNFceRequest

	// Get idempotency key from header
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is required"})
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.nfceUseCase.EmitNFce(ctx, idempotencyKey, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, response)
}

// GetNFceByID gets a NFC-e by ID
func (h *NFCeHandler) GetNFceByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	response, err := h.nfceUseCase.GetNFceByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "NFC-e not found"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListNFces lists NFC-e requests with pagination
func (h *NFCeHandler) ListNFces(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be >= 0"})
		return
	}

	response, err := h.nfceUseCase.ListNFces(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list NFC-es"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelNFce cancels a NFC-e
func (h *NFCeHandler) CancelNFce(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var req dto.CancelNFceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.nfceUseCase.CancelNFce(ctx, id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "NFC-e cancellation requested"})
}

// GetNFceEvents gets events for a NFC-e
func (h *NFCeHandler) GetNFceEvents(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.Param("id")

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 200"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be >= 0"})
		return
	}

	response, err := h.nfceUseCase.GetNFceEvents(ctx, requestID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get NFC-e events"})
		return
	}

	c.JSON(http.StatusOK, response)
}
