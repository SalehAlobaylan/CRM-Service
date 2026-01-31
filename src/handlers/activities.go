package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SalehAlobaylan/CRM-Service/src/middleware"
	"github.com/SalehAlobaylan/CRM-Service/src/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ActivityHandler handles activity-related endpoints
type ActivityHandler struct {
	db *gorm.DB
}

// NewActivityHandler creates a new ActivityHandler
func NewActivityHandler(db *gorm.DB) *ActivityHandler {
	return &ActivityHandler{db: db}
}

// ActivityCreateRequest represents the request body for creating an activity
type ActivityCreateRequest struct {
	Title       string               `json:"title" binding:"required,min=1,max=255"`
	Description string               `json:"description,omitempty"`
	Type        models.ActivityType  `json:"type" binding:"required"`
	Status      models.ActivityStatus `json:"status,omitempty"`
	CustomerID  *uint                `json:"customer_id,omitempty"`
	DealID      *uint                `json:"deal_id,omitempty"`
	ContactID   *uint                `json:"contact_id,omitempty"`
	AssignedTo  *uint                `json:"assigned_to,omitempty"`
	DueDate     *time.Time           `json:"due_date,omitempty"`
	Duration    int                  `json:"duration,omitempty"`
	Priority    string               `json:"priority,omitempty"`
}

// ActivityUpdateRequest represents the request body for updating an activity
type ActivityUpdateRequest struct {
	Title       string                `json:"title,omitempty"`
	Description string                `json:"description,omitempty"`
	Type        models.ActivityType   `json:"type,omitempty"`
	Status      models.ActivityStatus `json:"status,omitempty"`
	CustomerID  *uint                 `json:"customer_id,omitempty"`
	DealID      *uint                 `json:"deal_id,omitempty"`
	ContactID   *uint                 `json:"contact_id,omitempty"`
	AssignedTo  *uint                 `json:"assigned_to,omitempty"`
	DueDate     *time.Time            `json:"due_date,omitempty"`
	CompletedAt *time.Time            `json:"completed_at,omitempty"`
	Duration    *int                  `json:"duration,omitempty"`
	Outcome     string                `json:"outcome,omitempty"`
	Priority    string                `json:"priority,omitempty"`
}

// ActivityStatusUpdateRequest represents a status update request
type ActivityStatusUpdateRequest struct {
	Status  models.ActivityStatus `json:"status" binding:"required"`
	Outcome string                `json:"outcome,omitempty"`
}

// ListActivities returns a paginated list of activities with filtering
// GET /admin/activities
func (h *ActivityHandler) ListActivities(c *gin.Context) {
	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&models.Activity{})

	// Filters
	if activityType := c.Query("type"); activityType != "" {
		query = query.Where("type = ?", activityType)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		query = query.Where("assigned_to = ?", assignedTo)
	}
	if customerID := c.Query("customer_id"); customerID != "" {
		query = query.Where("customer_id = ?", customerID)
	}
	if dealID := c.Query("deal_id"); dealID != "" {
		query = query.Where("deal_id = ?", dealID)
	}
	if search := c.Query("search"); search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ?", searchTerm)
	}
	if dueDateFrom := c.Query("due_date_from"); dueDateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dueDateFrom); err == nil {
			query = query.Where("due_date >= ?", t)
		}
	}
	if dueDateTo := c.Query("due_date_to"); dueDateTo != "" {
		if t, err := time.Parse(time.RFC3339, dueDateTo); err == nil {
			query = query.Where("due_date <= ?", t)
		}
	}
	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// Sorting
	sortBy := c.DefaultQuery("sort_by", "due_date")
	sortOrder := c.DefaultQuery("sort_order", "asc")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}
	allowedSortFields := map[string]bool{
		"created_at": true, "updated_at": true, "title": true, "due_date": true,
		"status": true, "type": true, "priority": true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "due_date"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Count total
	var total int64
	query.Count(&total)

	// Get activities
	var activities []models.Activity
	offset := (page - 1) * pageSize
	if err := query.Preload("Customer").Preload("Deal").Offset(offset).Limit(pageSize).Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch activities",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	c.JSON(http.StatusOK, models.ActivityListResponse{
		Data:       activities,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetMyActivities returns activities assigned to the current user
// GET /admin/me/activities
func (h *ActivityHandler) GetMyActivities(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"code":    "NO_USER_CONTEXT",
			"message": "User not found in context",
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

	query := h.db.Model(&models.Activity{}).Where("assigned_to = ?", user.ID)

	// Filter by status (default to scheduled/overdue for "my tasks")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	} else {
		// Default: show scheduled and overdue tasks
		query = query.Where("status IN ?", []string{
			string(models.ActivityStatusScheduled),
			string(models.ActivityStatusOverdue),
		})
	}

	// Order by due date ascending (upcoming first)
	query = query.Order("due_date ASC NULLS LAST")

	// Count total
	var total int64
	query.Count(&total)

	// Get activities
	var activities []models.Activity
	offset := (page - 1) * pageSize
	if err := query.Preload("Customer").Preload("Deal").Offset(offset).Limit(pageSize).Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch activities",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	c.JSON(http.StatusOK, models.ActivityListResponse{
		Data:       activities,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// CreateActivity creates a new activity
// POST /admin/activities
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req ActivityCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Validate at least one link (customer or deal)
	if req.CustomerID == nil && req.DealID == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "MISSING_LINK",
			"message": "Activity must be linked to a customer or deal",
		})
		return
	}

	// Set defaults
	status := req.Status
	if status == "" {
		status = models.ActivityStatusScheduled
	}
	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	activity := models.Activity{
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Status:      status,
		CustomerID:  req.CustomerID,
		DealID:      req.DealID,
		ContactID:   req.ContactID,
		AssignedTo:  req.AssignedTo,
		DueDate:     req.DueDate,
		Duration:    req.Duration,
		Priority:    priority,
	}

	if err := h.db.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to create activity",
		})
		return
	}

	// Reload with relations
	h.db.Preload("Customer").Preload("Deal").First(&activity, activity.ID)

	// Log audit
	h.logAudit(c, "activity", activity.ID, models.AuditActionCreate, nil, &activity)

	c.JSON(http.StatusCreated, activity)
}

// GetActivity returns a single activity by ID
// GET /admin/activities/:id
func (h *ActivityHandler) GetActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid activity ID",
		})
		return
	}

	var activity models.Activity
	if err := h.db.Preload("Customer").Preload("Deal").Preload("Contact").First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "ACTIVITY_NOT_FOUND",
				"message": "Activity not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch activity",
		})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// UpdateActivity updates an activity
// PUT /admin/activities/:id
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid activity ID",
		})
		return
	}

	var activity models.Activity
	if err := h.db.First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "ACTIVITY_NOT_FOUND",
				"message": "Activity not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch activity",
		})
		return
	}

	oldActivity := activity

	var req ActivityUpdateRequest
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
		activity.Title = req.Title
	}
	if req.Description != "" {
		activity.Description = req.Description
	}
	if req.Type != "" {
		activity.Type = req.Type
	}
	if req.Status != "" {
		activity.Status = req.Status
	}
	if req.CustomerID != nil {
		activity.CustomerID = req.CustomerID
	}
	if req.DealID != nil {
		activity.DealID = req.DealID
	}
	if req.ContactID != nil {
		activity.ContactID = req.ContactID
	}
	if req.AssignedTo != nil {
		activity.AssignedTo = req.AssignedTo
	}
	if req.DueDate != nil {
		activity.DueDate = req.DueDate
	}
	if req.CompletedAt != nil {
		activity.CompletedAt = req.CompletedAt
	}
	if req.Duration != nil {
		activity.Duration = *req.Duration
	}
	if req.Outcome != "" {
		activity.Outcome = req.Outcome
	}
	if req.Priority != "" {
		activity.Priority = req.Priority
	}

	if err := h.db.Save(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to update activity",
		})
		return
	}

	// Reload with relations
	h.db.Preload("Customer").Preload("Deal").First(&activity, activity.ID)

	// Log audit
	h.logAudit(c, "activity", activity.ID, models.AuditActionUpdate, &oldActivity, &activity)

	c.JSON(http.StatusOK, activity)
}

// PatchActivity handles status updates (complete/cancel)
// PATCH /admin/activities/:id
func (h *ActivityHandler) PatchActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid activity ID",
		})
		return
	}

	var activity models.Activity
	if err := h.db.First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "ACTIVITY_NOT_FOUND",
				"message": "Activity not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch activity",
		})
		return
	}

	oldActivity := activity

	var req ActivityStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Update status
	activity.Status = req.Status

	// If completed, set completed_at
	if req.Status == models.ActivityStatusCompleted {
		now := time.Now()
		activity.CompletedAt = &now
	}

	if req.Outcome != "" {
		activity.Outcome = req.Outcome
	}

	if err := h.db.Save(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to update activity",
		})
		return
	}

	// Reload with relations
	h.db.Preload("Customer").Preload("Deal").First(&activity, activity.ID)

	// Log audit
	h.logAudit(c, "activity", activity.ID, models.AuditActionUpdate, &oldActivity, &activity)

	c.JSON(http.StatusOK, activity)
}

// DeleteActivity soft-deletes an activity
// DELETE /admin/activities/:id
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid activity ID",
		})
		return
	}

	var activity models.Activity
	if err := h.db.First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"code":    "ACTIVITY_NOT_FOUND",
				"message": "Activity not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch activity",
		})
		return
	}

	if err := h.db.Delete(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to delete activity",
		})
		return
	}

	// Log audit
	h.logAudit(c, "activity", activity.ID, models.AuditActionDelete, &activity, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Activity deleted successfully",
	})
}

// logAudit creates an audit log entry
func (h *ActivityHandler) logAudit(c *gin.Context, resourceType string, resourceID uint, action models.AuditAction, oldValue, newValue interface{}) {
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
