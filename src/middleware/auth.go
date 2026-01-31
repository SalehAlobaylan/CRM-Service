package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/SalehAlobaylan/CRM-Service/src/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID uint   `json:"user_id,omitempty"`
	Sub    string `json:"sub,omitempty"`
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Context keys for user information
const (
	ContextKeyUser     = "user"
	ContextKeyUserID   = "user_id"
	ContextKeyUserRole = "user_role"
	ContextKeyClaims   = "claims"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JWTAuth creates a JWT authentication middleware
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "MISSING_TOKEN",
				Message: "Authorization header is required",
			})
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "INVALID_TOKEN_FORMAT",
				Message: "Authorization header must be in 'Bearer <token>' format",
			})
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			var message string
			if errors.Is(err, jwt.ErrTokenExpired) {
				message = "Token has expired"
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				message = "Token is malformed"
			} else {
				message = "Invalid token"
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "INVALID_TOKEN",
				Message: message,
			})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "INVALID_TOKEN",
				Message: "Token is not valid",
			})
			return
		}

		// Extract user ID (could be in 'sub' or 'user_id')
		userID := claims.UserID
		if userID == 0 && claims.Sub != "" {
			// Sub claim might contain user ID as string - we'll just use 0 for now
			// In production, you might parse this from the sub field
		}

		// Validate role is present
		if claims.Role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "MISSING_ROLE",
				Message: "Token must contain a role claim",
			})
			return
		}

		// Create user object from claims
		user := models.User{
			ID:       userID,
			Email:    claims.Email,
			Name:     claims.Name,
			Role:     claims.Role,
			IsActive: true,
		}

		// Store user info in context
		c.Set(ContextKeyUser, user)
		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyUserRole, claims.Role)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// RequireRole creates middleware that requires specific roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyUserRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "NO_USER_CONTEXT",
				Message: "User context not found",
			})
			return
		}

		userRole := role.(string)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Code:    "INSUFFICIENT_PERMISSIONS",
			Message: "You do not have permission to access this resource",
		})
	}
}

// RequirePermission creates middleware that requires specific permissions
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyUserRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "NO_USER_CONTEXT",
				Message: "User context not found",
			})
			return
		}

		userRole := role.(string)
		if !models.HasPermission(userRole, permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
				Error:   "forbidden",
				Code:    "INSUFFICIENT_PERMISSIONS",
				Message: "You do not have permission to perform this action",
			})
			return
		}

		c.Next()
	}
}

// GetUserFromContext retrieves the user from the Gin context
func GetUserFromContext(c *gin.Context) (models.User, bool) {
	user, exists := c.Get(ContextKeyUser)
	if !exists {
		return models.User{}, false
	}
	return user.(models.User), true
}

// GetUserIDFromContext retrieves the user ID from the Gin context
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}
