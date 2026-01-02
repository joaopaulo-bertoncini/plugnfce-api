package handler

import (
	"fmt"
	"io"
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

type CompanyHandlerInterface interface {
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	UpdateCertificate(c *gin.Context)
	UpdateCertificateByID(c *gin.Context)
	UpdateCSC(c *gin.Context)
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

// UpdateCertificate updates the authenticated company's certificate via file upload
func (h *CompanyHandler) UpdateCertificate(c *gin.Context) {
	companyID := c.GetString("company_id")
	if companyID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Process multipart/form-data file upload
	pfxData, password, expiresAt, err := h.processMultipartUpload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.companyUseCase.UpdateCertificate(c.Request.Context(), companyID, pfxData, password, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "certificate updated successfully"})
}

// UpdateCertificateByID updates a company's certificate by ID (admin endpoint) via file upload
func (h *CompanyHandler) UpdateCertificateByID(c *gin.Context) {
	companyID := c.Param("id")
	if companyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "company ID is required"})
		return
	}

	// Process multipart/form-data file upload
	pfxData, password, expiresAt, err := h.processMultipartUpload(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.companyUseCase.UpdateCertificate(c.Request.Context(), companyID, pfxData, password, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "certificate updated successfully"})
}

// processMultipartUpload handles file upload via multipart/form-data
func (h *CompanyHandler) processMultipartUpload(c *gin.Context) ([]byte, string, time.Time, error) {
	// Get PFX file
	file, _, err := c.Request.FormFile("pfx_file")
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("pfx_file is required")
	}
	defer file.Close()

	// Read file data
	pfxData, err := io.ReadAll(file)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("failed to read PFX file: %w", err)
	}

	// Get password
	password := c.PostForm("password")
	if password == "" {
		return nil, "", time.Time{}, fmt.Errorf("password is required")
	}

	// Get expiration date
	expiresAtStr := c.PostForm("expires_at")
	if expiresAtStr == "" {
		return nil, "", time.Time{}, fmt.Errorf("expires_at is required")
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, "", time.Time{}, fmt.Errorf("invalid expires_at format (use RFC3339): %w", err)
	}

	return pfxData, password, expiresAt, nil
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
