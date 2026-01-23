package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SalehAlobaylan/CRM-Service/internal/middleware"
	"github.com/SalehAlobaylan/CRM-Service/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DealHandler handles deal-related endpoints
type DealHandler struct {
	db *gorm.DB
}

// NewDealHandler creates a new DealHandler
func NewDealHandler(db *gorm.DB) *DealHandler {
	return &DealHandler{db: db}
}

// DealCreateRequest represents the request body for creating a deal
type DealCreateRequest struct {
	Title             string           `json:"title" binding:"required,min=1,max=255"`
	Description       string           `json:"description,omitempty"`
	CustomerID        uint             `json:"customer_id" binding:"required"`
	ContactID         *uint            `json:"contact_id,omitempty"`
	Stage             models.DealStage `json:"stage,omitempty"`
	Amount            float64          `json:"amount,omitempty"`
	Currency          string           `json:"currency,omitempty"`
	Probability       int              `json:"probability,omitempty"`
	ExpectedCloseDate *time.Time       `json:"expected_close_date,omitempty"`
	OwnerID           *uint            `json:"owner_id,omitempty"`
}

// DealUpdateRequest represents the request body for updating a deal
type DealUpdateRequest struct {
	Title             string           `json:"title,omitempty"`
	Description       string           `json:"description,omitempty"`
	CustomerID        *uint            `json:"customer_id,omitempty"`
	ContactID         *uint            `json:"contact_id,omitempty"`
	Stage             models.DealStage `json:"stage,omitempty"`
	Amount            *float64         `json:"amount,omitempty"`
	Currency          string           `json:"currency,omitempty"`
	Probability       *int             `json:"probability,omitempty"`
	ExpectedCloseDate *time.Time       `json:"expected_close_date,omitempty"`
	ActualCloseDate   *time.Time       `json:"actual_close_date,omitempty"`
	OwnerID           *uint            `json:"owner_id,omitempty"`
	LostReason        string           `json:"lost_reason,omitempty"`
}

// DealStageTransitionRequest represents a stage transition request
type DealStageTransitionRequest struct {
	Stage      models.DealStage `json:"stage" binding:"required"`
	LostReason string           `json:"lost_reason,omitempty"`
}

// ListDeals returns a paginated list of deals with filtering
// GET /admin/deals
func (h *DealHandler) ListDeals(c *gin.Context) {
	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&models.Deal{})

	// Filters
	if stage := c.Query("stage"); stage != "" {
		query = query.Where("stage = ?", stage)
	}
	if ownerID := c.Query("owner_id"); ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}
	if customerID := c.Query("customer_id"); customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}
	if search := c.Query("search"); search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ?", searchTerm)
	}
	if amountMin := c.Query("amount_min"); amountMin != "" {
		if val, err := strconv.ParseFloat(amountMin, 64); err == nil {
			query = query.Where("amount >= ?", val)
		}
	}
	if amountMax := c.Query("amount_max"); amountMax != "" {
		if val, err := strconv.ParseFloat(amountMax, 64); err == nil {
			query = query.Where("amount <= ?", val)
		}
	}
	if closeDateFrom := c.Query("expected_close_from"); closeDateFrom != "" {
		if t, err := time.Parse(time.RFC3339, closeDateFrom); err == nil {
			query = query.Where("expected_close_date >= ?", t)
		}
	}
	if closeDateTo := c.Query("expected_close_to"); closeDateTo != "" {
		if t, err := time.Parse(time.RFC3339, closeDateTo); err == nil {
			query = query.Where("expected_close_date <= ?", t)
		}
	}

	// Sorting
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	allowedSortFields := map[string]bool{
		"created_at": true, "updated_at": true, "title": true, "amount": true,
		"expected_close_date": true, "stage": true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "created_at"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Count total
	var total int64
	query.Count(&total)

	// Get deals
	var deals []models.Deal
	offset := (page - 1) * pageSize
	if err := query.Preload("Customer").Offset(offset).Limit(pageSize).Find(&deals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch deals",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	c.JSON(http.StatusOK, models.DealListResponse{
		Data:       deals,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// CreateDeal creates a new deal
// POST /admin/deals
func (h *DealHandler) CreateDeal(c *gin.Context) {
	var req DealCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Verify customer exists
	var customer models.Customer
	if err := h.db.First(&customer, req.CustomerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"code":    "CUSTOMER_NOT_FOUND",
				"message": "Customer not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to verify customer",
		})
		return
	}

	// Set defaults
	stage := req.Stage
	if stage == "" {
		stage = models.DealStageProspecting
	}
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Validate probability
	probability := req.Probability
	if probability < 0 {
		probability = 0
	}
	if probability > 100 {
		probability = 100
	}

	deal := models.Deal{
		Title:             req.Title,
		Description:       req.Description,
		CustomerID:        req.CustomerID,
		ContactID:         req.ContactID,
		Stage:             stage,
		Amount:            req.Amount,
		Currency:          currency,
		Probability:       probability,
		ExpectedCloseDate: req.ExpectedCloseDate,
		OwnerID:           req.OwnerID,
	}

	if err := h.db.Create(&deal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to create deal",
		})
		return
	}

	// Reload with customer
	h.db.Preload("Customer").First(&deal, deal.ID)

	// Log audit
	h.logAudit(c, "deal", deal.ID, models.AuditActionCreate, nil, &deal)

	c.JSON(http.StatusCreated, deal)
}

// GetDeal returns a single deal by ID
// GET /admin/deals/:id
func (h *DealHandler) GetDeal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid deal ID",
		})
		return
	}

	var deal models.Deal
	if err := h.db.Preload("Customer").Preload("Contact").Preload("Activities").Preload("Notes").First(&deal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "DEAL_NOT_FOUND",
				"message": "Deal not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch deal",
		})
		return
	}

	c.JSON(http.StatusOK, deal)
}

// UpdateDeal updates a deal
// PUT /admin/deals/:id
func (h *DealHandler) UpdateDeal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid deal ID",
		})
		return
	}

	var deal models.Deal
	if err := h.db.First(&deal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "DEAL_NOT_FOUND",
				"message": "Deal not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch deal",
		})
		return
	}

	oldDeal := deal

	var req DealUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Update fields
	if req.Title != "" {
		deal.Title = req.Title
	}
	if req.Description != "" {
		deal.Description = req.Description
	}
	if req.CustomerID != nil {
		deal.CustomerID = *req.CustomerID
	}
	if req.ContactID != nil {
		deal.ContactID = req.ContactID
	}
	if req.Stage != "" {
		if !models.IsValidDealStage(req.Stage) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"code":    "INVALID_STAGE",
				"message": "Invalid deal stage",
			})
			return
		}
		deal.Stage = req.Stage
	}
	if req.Amount != nil {
		deal.Amount = *req.Amount
	}
	if req.Currency != "" {
		deal.Currency = req.Currency
	}
	if req.Probability != nil {
		prob := *req.Probability
		if prob < 0 {
			prob = 0
		}
		if prob > 100 {
			prob = 100
		}
		deal.Probability = prob
	}
	if req.ExpectedCloseDate != nil {
		deal.ExpectedCloseDate = req.ExpectedCloseDate
	}
	if req.ActualCloseDate != nil {
		deal.ActualCloseDate = req.ActualCloseDate
	}
	if req.OwnerID != nil {
		deal.OwnerID = req.OwnerID
	}
	if req.LostReason != "" {
		deal.LostReason = req.LostReason
	}

	if err := h.db.Save(&deal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to update deal",
		})
		return
	}

	// Reload with customer
	h.db.Preload("Customer").First(&deal, deal.ID)

	// Log audit
	h.logAudit(c, "deal", deal.ID, models.AuditActionUpdate, &oldDeal, &deal)

	c.JSON(http.StatusOK, deal)
}

// PatchDeal handles stage transitions
// PATCH /admin/deals/:id
func (h *DealHandler) PatchDeal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid deal ID",
		})
		return
	}

	var deal models.Deal
	if err := h.db.First(&deal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "DEAL_NOT_FOUND",
				"message": "Deal not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch deal",
		})
		return
	}

	oldDeal := deal

	var req DealStageTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Validate stage
	if !models.IsValidDealStage(req.Stage) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_STAGE",
			"message": "Invalid deal stage",
		})
		return
	}

	// Update stage
	deal.Stage = req.Stage

	// If closed, set actual close date
	if req.Stage == models.DealStageClosedWon || req.Stage == models.DealStageClosedLost {
		now := time.Now()
		deal.ActualCloseDate = &now
		if req.Stage == models.DealStageClosedLost && req.LostReason != "" {
			deal.LostReason = req.LostReason
		}
	}

	if err := h.db.Save(&deal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to update deal",
		})
		return
	}

	// Reload with customer
	h.db.Preload("Customer").First(&deal, deal.ID)

	// Log audit
	h.logAudit(c, "deal", deal.ID, models.AuditActionUpdate, &oldDeal, &deal)

	c.JSON(http.StatusOK, deal)
}

// DeleteDeal soft-deletes a deal
// DELETE /admin/deals/:id
func (h *DealHandler) DeleteDeal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid deal ID",
		})
		return
	}

	var deal models.Deal
	if err := h.db.First(&deal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "DEAL_NOT_FOUND",
				"message": "Deal not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch deal",
		})
		return
	}

	if err := h.db.Delete(&deal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to delete deal",
		})
		return
	}

	// Log audit
	h.logAudit(c, "deal", deal.ID, models.AuditActionDelete, &deal, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Deal deleted successfully",
	})
}

// logAudit creates an audit log entry
func (h *DealHandler) logAudit(c *gin.Context, resourceType string, resourceID uint, action models.AuditAction, oldValue, newValue interface{}) {
	user, _ := middleware.GetUserFromContext(c)

	audit := models.AuditLog{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		UserID:       user.ID,
		UserName:     user.Name,
		UserRole:     user.Role,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	}

	h.db.Create(&audit)
}
