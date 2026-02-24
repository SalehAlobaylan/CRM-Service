package handlers

import (
	"net/http"
	"strings"

	"github.com/SalehAlobaylan/CRM-Service/src/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func tenantScopedDB(c *gin.Context, db *gorm.DB) (*gorm.DB, string, bool) {
	tenantID, ok := middleware.GetTenantIDFromContext(c)
	if !ok || strings.TrimSpace(tenantID) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"code":    "MISSING_TENANT",
			"message": "tenant_id claim is required",
		})
		return nil, "", false
	}

	return db.Where("tenant_id = ?", tenantID), tenantID, true
}
