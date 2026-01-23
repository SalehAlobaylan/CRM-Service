package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/SalehAlobaylan/CRM-Service/internal/middleware"
	"github.com/SalehAlobaylan/CRM-Service/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ContactHandler handles contact-related endpoints
type ContactHandler struct {
	db *gorm.DB
}

// NewContactHandler creates a new ContactHandler
func NewContactHandler(db *gorm.DB) *ContactHandler {
	return &ContactHandler{db: db}
}

// ContactCreateRequest represents the request body for creating a contact
type ContactCreateRequest struct {
	FirstName string `json:"first_name" binding:"required,min=1,max=100"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Position  string `json:"position,omitempty"`
	IsPrimary bool   `json:"is_primary,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

// ContactUpdateRequest represents the request body for updating a contact
type ContactUpdateRequest struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Position  string `json:"position,omitempty"`
	IsPrimary *bool  `json:"is_primary,omitempty"`
	Notes     string `json:"notes,omitempty"`
}

// ListContacts returns all contacts for a customer
// GET /admin/customers/:id/contacts
func (h *ContactHandler) ListContacts(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid customer ID",
		})
		return
	}

	// Verify customer exists
	var customer models.Customer
	if err := h.db.First(&customer, customerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "CUSTOMER_NOT_FOUND",
				"message": "Customer not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch customer",
		})
		return
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Get contacts
	var total int64
	h.db.Model(&models.Contact{}).Where("customer_id = ?", customerID).Count(&total)

	var contacts []models.Contact
	offset := (page - 1) * pageSize
	if err := h.db.Where("customer_id = ?", customerID).
		Order("is_primary DESC, created_at ASC").
		Offset(offset).Limit(pageSize).
		Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch contacts",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	c.JSON(http.StatusOK, models.ContactListResponse{
		Data:       contacts,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// CreateContact creates a new contact for a customer
// POST /admin/customers/:id/contacts
func (h *ContactHandler) CreateContact(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid customer ID",
		})
		return
	}

	// Verify customer exists
	var customer models.Customer
	if err := h.db.First(&customer, customerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "CUSTOMER_NOT_FOUND",
				"message": "Customer not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch customer",
		})
		return
	}

	var req ContactCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// If this is set as primary, unset other primaries
	if req.IsPrimary {
		h.db.Model(&models.Contact{}).Where("customer_id = ?", customerID).Update("is_primary", false)
	}

	contact := models.Contact{
		CustomerID: uint(customerID),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Phone:      req.Phone,
		Position:   req.Position,
		IsPrimary:  req.IsPrimary,
		Notes:      req.Notes,
	}

	if err := h.db.Create(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to create contact",
		})
		return
	}

	// Log audit
	h.logAudit(c, "contact", contact.ID, models.AuditActionCreate, nil, &contact)

	c.JSON(http.StatusCreated, contact)
}

// UpdateContact updates a contact
// PUT /admin/contacts/:id
func (h *ContactHandler) UpdateContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid contact ID",
		})
		return
	}

	var contact models.Contact
	if err := h.db.First(&contact, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "CONTACT_NOT_FOUND",
				"message": "Contact not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch contact",
		})
		return
	}

	oldContact := contact

	var req ContactUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Update fields
	if req.FirstName != "" {
		contact.FirstName = req.FirstName
	}
	if req.LastName != "" {
		contact.LastName = req.LastName
	}
	if req.Email != "" {
		contact.Email = req.Email
	}
	if req.Phone != "" {
		contact.Phone = req.Phone
	}
	if req.Position != "" {
		contact.Position = req.Position
	}
	if req.Notes != "" {
		contact.Notes = req.Notes
	}
	if req.IsPrimary != nil {
		// If setting as primary, unset other primaries
		if *req.IsPrimary {
			h.db.Model(&models.Contact{}).Where("customer_id = ? AND id != ?", contact.CustomerID, id).Update("is_primary", false)
		}
		contact.IsPrimary = *req.IsPrimary
	}

	if err := h.db.Save(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to update contact",
		})
		return
	}

	// Log audit
	h.logAudit(c, "contact", contact.ID, models.AuditActionUpdate, &oldContact, &contact)

	c.JSON(http.StatusOK, contact)
}

// DeleteContact deletes a contact
// DELETE /admin/contacts/:id
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid contact ID",
		})
		return
	}

	var contact models.Contact
	if err := h.db.First(&contact, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "CONTACT_NOT_FOUND",
				"message": "Contact not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch contact",
		})
		return
	}

	if err := h.db.Delete(&contact).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to delete contact",
		})
		return
	}

	// Log audit
	h.logAudit(c, "contact", contact.ID, models.AuditActionDelete, &contact, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Contact deleted successfully",
	})
}

// logAudit creates an audit log entry
func (h *ContactHandler) logAudit(c *gin.Context, resourceType string, resourceID uint, action models.AuditAction, oldValue, newValue interface{}) {
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
