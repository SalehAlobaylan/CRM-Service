package handlers

import (
	"net/http"

	"github.com/SalehAlobaylan/CRM-Service/internal/middleware"
	"github.com/SalehAlobaylan/CRM-Service/internal/models"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related endpoints
type AuthHandler struct{}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// GetMe returns the current user's information from JWT claims
// GET /admin/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"code":    "NO_USER_CONTEXT",
			"message": "User not found in context",
		})
		return
	}

	// Get permissions for user's role
	permissions := models.RolePermissions[user.Role]
	if permissions == nil {
		permissions = []string{}
	}

	response := models.MeResponse{
		User:        user,
		Permissions: permissions,
	}

	c.JSON(http.StatusOK, response)
}
