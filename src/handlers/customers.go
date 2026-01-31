package handlers

import (
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SalehAlobaylan/CRM-Service/src/middleware"
	"github.com/SalehAlobaylan/CRM-Service/src/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CustomerHandler handles customer-related endpoints
type CustomerHandler struct {
	db *gorm.DB
}

// NewCustomerHandler creates a new CustomerHandler
func NewCustomerHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{db: db}
}

// CustomerCreateRequest represents the request body for creating a customer
type CustomerCreateRequest struct {
	Name           string              `json:"name" binding:"required,min=1,max=255"`
	Email          string              `json:"email" binding:"required,email"`
	Phone          string              `json:"phone,omitempty"`
	Company        string              `json:"company,omitempty"`
	Role           string              `json:"role,omitempty"`
	Status         models.CustomerStatus `json:"status,omitempty"`
	AssignedTo     *uint               `json:"assigned_to,omitempty"`
	Notes          string              `json:"notes,omitempty"`
	NextFollowUpAt *time.Time          `json:"next_follow_up_at,omitempty"`
}

// CustomerUpdateRequest represents the request body for updating a customer
type CustomerUpdateRequest struct {
	Name           string              `json:"name" binding:"omitempty,min=1,max=255"`
	Email          string              `json:"email" binding:"omitempty,email"`
	Phone          string              `json:"phone,omitempty"`
	Company        string              `json:"company,omitempty"`
	Role           string              `json:"role,omitempty"`
	Status         models.CustomerStatus `json:"status,omitempty"`
	AssignedTo     *uint               `json:"assigned_to,omitempty"`
	Contacted      *bool               `json:"contacted,omitempty"`
	Notes          string              `json:"notes,omitempty"`
	NextFollowUpAt *time.Time          `json:"next_follow_up_at,omitempty"`
}

// CustomerPatchRequest represents the request body for patching a customer
type CustomerPatchRequest struct {
	Status         *models.CustomerStatus `json:"status,omitempty"`
	AssignedTo     *uint                  `json:"assigned_to,omitempty"`
	Contacted      *bool                  `json:"contacted,omitempty"`
	NextFollowUpAt *time.Time             `json:"next_follow_up_at,omitempty"`
}

// ListCustomers returns a paginated list of customers with filtering
// GET /admin/customers
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build query
	query := h.db.Model(&models.Customer{})

	// Apply filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		query = query.Where("assigned_to = ?", assignedTo)
	}
	if search := c.Query("search"); search != "" {
		searchTerm := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(company) LIKE ?",
			searchTerm, searchTerm, searchTerm)
	}
	if createdFrom := c.Query("created_from"); createdFrom != "" {
		if t, err := time.Parse(time.RFC3339, createdFrom); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if createdTo := c.Query("created_to"); createdTo != "" {
		if t, err := time.Parse(time.RFC3339, createdTo); err == nil {
			query = query.Where("created_at <= ?", t)
		}
	}
	if tagIDs := c.Query("tags"); tagIDs != "" {
		ids := strings.Split(tagIDs, ",")
		query = query.Joins("JOIN customer_tags ON customer_tags.customer_id = customers.id").
			Where("customer_tags.tag_id IN ?", ids)
	}

	// Sorting
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	allowedSortFields := map[string]bool{
		"created_at": true, "updated_at": true, "name": true, "email": true, "status": true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "created_at"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Get total count
	var total int64
	query.Count(&total)

	// Apply pagination
	offset := (page - 1) * pageSize
	var customers []models.Customer
	if err := query.Preload("Tags").Offset(offset).Limit(pageSize).Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to fetch customers",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	c.JSON(http.StatusOK, models.CustomerListResponse{
		Data:       customers,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// CreateCustomer creates a new customer
// POST /admin/customers
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req CustomerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_EMAIL",
			"message": "Invalid email format",
		})
		return
	}

	// Check email uniqueness
	var existing models.Customer
	if err := h.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "conflict",
			"code":    "EMAIL_EXISTS",
			"message": "A customer with this email already exists",
		})
		return
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = models.CustomerStatusLead
	}

	customer := models.Customer{
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		Company:        req.Company,
		Role:           req.Role,
		Status:         status,
		AssignedTo:     req.AssignedTo,
		Notes:          req.Notes,
		NextFollowUpAt: req.NextFollowUpAt,
	}

	if err := h.db.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to create customer",
		})
		return
	}

	// Log audit
	h.logAudit(c, "customer", customer.ID, models.AuditActionCreate, nil, &customer)

	c.JSON(http.StatusCreated, customer)
}

// GetCustomer returns a single customer by ID with related entities
// GET /admin/customers/:id
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid customer ID",
		})
		return
	}

	var customer models.Customer
	if err := h.db.Preload("Tags").First(&customer, id).Error; err != nil {
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

	// Get related counts
	var contactsCount int64
	h.db.Model(&models.Contact{}).Where("customer_id = ?", id).Count(&contactsCount)

	var openDealsCount int64
	h.db.Model(&models.Deal{}).Where("customer_id = ? AND stage NOT IN ?", id,
		[]string{string(models.DealStageClosedWon), string(models.DealStageClosedLost)}).Count(&openDealsCount)

	var upcomingActivitiesCount int64
	h.db.Model(&models.Activity{}).Where("customer_id = ? AND status = ? AND due_date > ?",
		id, models.ActivityStatusScheduled, time.Now()).Count(&upcomingActivitiesCount)

	// Get recent activities
	var recentActivities []models.Activity
	h.db.Where("customer_id = ?", id).Order("created_at DESC").Limit(5).Find(&recentActivities)

	response := models.CustomerDetailResponse{
		Customer:                customer,
		ContactsCount:           int(contactsCount),
		OpenDealsCount:          int(openDealsCount),
		UpcomingActivitiesCount: int(upcomingActivitiesCount),
		RecentActivities:        recentActivities,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateCustomer fully updates a customer
// PUT /admin/customers/:id
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid customer ID",
		})
		return
	}

	var customer models.Customer
	if err := h.db.First(&customer, id).Error; err != nil {
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

	oldCustomer := customer

	var req CustomerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// If email is being changed, check uniqueness
	if req.Email != "" && req.Email != customer.Email {
		if !isValidEmail(req.Email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation_error",
				"code":    "INVALID_EMAIL",
				"message": "Invalid email format",
			})
			return
		}

		var existing models.Customer
		if err := h.db.Where("email = ? AND id != ?", req.Email, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "conflict",
				"code":    "EMAIL_EXISTS",
				"message": "A customer with this email already exists",
			})
			return
		}
		customer.Email = req.Email
	}

	// Update fields
	if req.Name != "" {
		customer.Name = req.Name
	}
	if req.Phone != "" {
		customer.Phone = req.Phone
	}
	if req.Company != "" {
		customer.Company = req.Company
	}
	if req.Role != "" {
		customer.Role = req.Role
	}
	if req.Status != "" {
		customer.Status = req.Status
	}
	if req.AssignedTo != nil {
		customer.AssignedTo = req.AssignedTo
	}
	if req.Contacted != nil {
		customer.Contacted = *req.Contacted
	}
	if req.Notes != "" {
		customer.Notes = req.Notes
	}
	if req.NextFollowUpAt != nil {
		customer.NextFollowUpAt = req.NextFollowUpAt
	}

	if err := h.db.Save(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to update customer",
		})
		return
	}

	// Log audit
	h.logAudit(c, "customer", customer.ID, models.AuditActionUpdate, &oldCustomer, &customer)

	c.JSON(http.StatusOK, customer)
}

// PatchCustomer partially updates a customer
// PATCH /admin/customers/:id
func (h *CustomerHandler) PatchCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid customer ID",
		})
		return
	}

	var customer models.Customer
	if err := h.db.First(&customer, id).Error; err != nil {
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

	oldCustomer := customer

	var req CustomerPatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		})
		return
	}

	// Apply patch updates
	updates := make(map[string]interface{})
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.AssignedTo != nil {
		updates["assigned_to"] = *req.AssignedTo
	}
	if req.Contacted != nil {
		updates["contacted"] = *req.Contacted
	}
	if req.NextFollowUpAt != nil {
		updates["next_follow_up_at"] = *req.NextFollowUpAt
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "NO_UPDATES",
			"message": "No fields to update",
		})
		return
	}

	if err := h.db.Model(&customer).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to update customer",
		})
		return
	}

	// Reload customer
	h.db.First(&customer, id)

	// Log audit
	h.logAudit(c, "customer", customer.ID, models.AuditActionUpdate, &oldCustomer, &customer)

	c.JSON(http.StatusOK, customer)
}

// DeleteCustomer soft-deletes a customer
// DELETE /admin/customers/:id
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"code":    "INVALID_ID",
			"message": "Invalid customer ID",
		})
		return
	}

	var customer models.Customer
	if err := h.db.First(&customer, id).Error; err != nil {
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

	// Soft delete
	if err := h.db.Delete(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"code":    "DATABASE_ERROR",
			"message": "Failed to delete customer",
		})
		return
	}

	// Log audit
	h.logAudit(c, "customer", customer.ID, models.AuditActionDelete, &customer, nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Customer deleted successfully",
	})
}

// logAudit creates an audit log entry
func (h *CustomerHandler) logAudit(c *gin.Context, resourceType string, resourceID uint, action models.AuditAction, oldValue, newValue interface{}) {
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

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
