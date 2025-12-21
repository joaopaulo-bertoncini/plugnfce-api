package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/dto"
	"github.com/joaopaulo-bertoncini/plugnfce-api/internal/application/usecase"
)

// CompanyHandler manages HTTP requests related to company operations
type CompanyHandler struct {
	companyUseCase usecase.CompanyUseCase
}

// NewCompanyHandler creates a new CompanyHandler
func NewCompanyHandler(companyUseCase usecase.CompanyUseCase) *CompanyHandler {
	return &CompanyHandler{
		companyUseCase: companyUseCase,
	}
}

// GetProfile gets the company profile
func (h *CompanyHandler) GetProfile(c *gin.Context) {
	companyID := c.GetString("company_id") // From auth middleware
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	company, err := h.companyUseCase.GetProfile(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, company)
}

// UpdateProfile updates the company profile
func (h *CompanyHandler) UpdateProfile(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current profile first
	currentProfile, err := h.companyUseCase.GetProfile(c.Request.Context(), companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply updates to the current profile
	if req.NomeFantasia != nil {
		currentProfile.NomeFantasia = *req.NomeFantasia
	}
	if req.InscricaoEstadual != nil {
		currentProfile.InscricaoEstadual = *req.InscricaoEstadual
	}
	if req.Email != nil {
		currentProfile.Email = *req.Email
	}
	if req.Endereco != nil {
		currentProfile.Endereco = *req.Endereco
	}
	if req.RegimeTributario != nil {
		currentProfile.RegimeTributario = *req.RegimeTributario
	}
	if req.SerieNFCe != nil {
		currentProfile.SerieNFCe = *req.SerieNFCe
	}
	if req.Status != nil {
		currentProfile.Status = *req.Status
	}

	err = h.companyUseCase.UpdateProfile(c.Request.Context(), currentProfile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}

// UpdateCertificate updates the company certificate
func (h *CompanyHandler) UpdateCertificate(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.UpdateCompanyCertificateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Decode base64 PFX data
	var pfxData []byte

	err := h.companyUseCase.UpdateCertificate(c.Request.Context(), companyID, &dto.CertificateDTO{
		Type:      req.Type,
		PFXData:   pfxData,
		Password:  req.Password,
		ExpiresAt: req.ExpiresAt,
	}, pfxData, req.Password, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "certificate updated successfully"})
}

// UpdateCSC updates the company CSC configuration
func (h *CompanyHandler) UpdateCSC(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		CSCID      string    `json:"csc_id"`
		CSCToken   string    `json:"csc_token"`
		ValidUntil time.Time `json:"valid_until"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.companyUseCase.UpdateCSC(c.Request.Context(), companyID, req.CSCID, req.CSCToken, req.ValidUntil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CSC updated successfully"})
}
