package handlers

import (
	"net/http"
	"strconv"

	"github.com/SalehAlobaylan/CRM-Service/src/middleware"
	"github.com/SalehAlobaylan/CRM-Service/src/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TagHandler handles tag-related endpoints
type TagHandler struct {
	db *gorm.DB
}

// NewTagHandler creates a new TagHandler
func NewTagHandler(db *gorm.DB) *TagHandler {
	return &TagHandler{db: db}
}

// TagCreateRequest represents the request body for creating a tag
type TagCreateRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=100"`
	Color string `json:"color,omitempty"`
}

// TagUpdateRequest represents the request body for updating a tag
type TagUpdateRequest struct {
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

// ListTags returns all tags
// GET /admin/tags
func (h *TagHandler) ListTags(c *gin.Context) {
	db, _, ok := tenantScopedDB(c, h.db)
	if !ok {
		return
	}

	var tags []models.Tag
	if err := db.Order("name ASC").Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch tags",
		})
		return
	}

	c.JSON(http.StatusOK, models.TagListResponse{
		Data:  tags,
		Total: int64(len(tags)),
	})
}

// CreateTag creates a new tag
// POST /admin/tags
func (h *TagHandler) CreateTag(c *gin.Context) {
	db, tenantID, ok := tenantScopedDB(c, h.db)
	if !ok {
		return
	}

	var req TagCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Check uniqueness
	var existing models.Tag
	if err := db.Where("name = ?", req.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "conflict",
			"code":    "TAG_EXISTS",
			"message": "A tag with this name already exists",
		})
		return
	}

	tag := models.Tag{
		BaseModel: models.BaseModel{
			TenantID: tenantID,
		},
		Name:  req.Name,
		Color: req.Color,
	}

	if err := db.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to create tag",
		})
		return
	}

	// Log audit
	h.logAudit(c, "tag", tag.ID, models.AuditActionCreate, nil, &tag)

	c.JSON(http.StatusCreated, tag)
}

// UpdateTag updates a tag
// PUT /admin/tags/:id
func (h *TagHandler) UpdateTag(c *gin.Context) {
	db, _, ok := tenantScopedDB(c, h.db)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid tag ID",
		})
		return
	}

	var tag models.Tag
	if err := db.First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "TAG_NOT_FOUND",
				"message": "Tag not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch tag",
		})
		return
	}

	oldTag := tag

	var req TagUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Check uniqueness if name is being changed
	if req.Name != "" && req.Name != tag.Name {
		var existing models.Tag
		if err := db.Where("name = ? AND id != ?", req.Name, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "conflict",
				"code":    "TAG_EXISTS",
				"message": "A tag with this name already exists",
			})
			return
		}
		tag.Name = req.Name
	}

	if req.Color != "" {
		tag.Color = req.Color
	}

	if err := db.Save(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to update tag",
		})
		return
	}

	// Log audit
	h.logAudit(c, "tag", tag.ID, models.AuditActionUpdate, &oldTag, &tag)

	c.JSON(http.StatusOK, tag)
}

// DeleteTag deletes a tag
// DELETE /admin/tags/:id
func (h *TagHandler) DeleteTag(c *gin.Context) {
	db, tenantID, ok := tenantScopedDB(c, h.db)
	if !ok {
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid tag ID",
		})
		return
	}

	var tag models.Tag
	if err := db.First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "TAG_NOT_FOUND",
				"message": "Tag not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch tag",
		})
		return
	}

	// Remove associations for the same tenant.
	db.Where("tag_id = ? AND tenant_id = ?", tag.ID, tenantID).Delete(&models.CustomerTag{})

	// Delete tag
	if err := db.Delete(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to delete tag",
		})
		return
	}

	// Log audit
	h.logAudit(c, "tag", tag.ID, models.AuditActionDelete, &tag, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag deleted successfully",
	})
}

// AssignTagToCustomer assigns a tag to a customer
// POST /admin/customers/:id/tags/:tagId
func (h *TagHandler) AssignTagToCustomer(c *gin.Context) {
	db, tenantID, ok := tenantScopedDB(c, h.db)
	if !ok {
		return
	}

	customerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid customer ID",
		})
		return
	}

	tagID, err := strconv.ParseUint(c.Param("tagId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid tag ID",
		})
		return
	}

	// Verify customer exists
	var customer models.Customer
	if err := db.First(&customer, customerID).Error; err != nil {
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

	// Verify tag exists
	var tag models.Tag
	if err := db.First(&tag, tagID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "TAG_NOT_FOUND",
				"message": "Tag not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch tag",
		})
		return
	}

	// Add association explicitly with tenant scope.
	association := models.CustomerTag{
		CustomerID: uint(customerID),
		TagID:      uint(tagID),
		TenantID:   tenantID,
	}
	if err := db.Where("customer_id = ? AND tag_id = ? AND tenant_id = ?", association.CustomerID, association.TagID, association.TenantID).
		FirstOrCreate(&association).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to assign tag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag assigned successfully",
	})
}

// RemoveTagFromCustomer removes a tag from a customer
// DELETE /admin/customers/:id/tags/:tagId
func (h *TagHandler) RemoveTagFromCustomer(c *gin.Context) {
	db, tenantID, ok := tenantScopedDB(c, h.db)
	if !ok {
		return
	}

	customerID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid customer ID",
		})
		return
	}

	tagID, err := strconv.ParseUint(c.Param("tagId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid tag ID",
		})
		return
	}

	// Verify customer exists
	var customer models.Customer
	if err := db.First(&customer, customerID).Error; err != nil {
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

	// Verify tag exists
	var tag models.Tag
	if err := db.First(&tag, tagID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "TAG_NOT_FOUND",
				"message": "Tag not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch tag",
		})
		return
	}

	// Remove association with tenant scope.
	if err := db.Where("customer_id = ? AND tag_id = ? AND tenant_id = ?", customerID, tagID, tenantID).
		Delete(&models.CustomerTag{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to remove tag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag removed successfully",
	})
}

// logAudit creates an audit log entry
func (h *TagHandler) logAudit(c *gin.Context, resourceType string, resourceID uint, action models.AuditAction, oldValue, newValue interface{}) {
	user, _ := middleware.GetUserFromContext(c)
	_, tenantID, ok := tenantScopedDB(c, h.db)
	if !ok {
		return
	}

	audit := models.AuditLog{
		TenantID:     tenantID,
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
