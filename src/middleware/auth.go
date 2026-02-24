package middleware

import (
	"errors"
	"hash/fnv"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/SalehAlobaylan/CRM-Service/src/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents supported token claims from IAM (canonical) and CMS (legacy).
type JWTClaims struct {
	UserID      string   `json:"user_id,omitempty"`
	Sub         string   `json:"sub,omitempty"`
	Email       string   `json:"email,omitempty"`
	Name        string   `json:"name,omitempty"`
	TenantID    string   `json:"tenant_id,omitempty"`
	Role        string   `json:"role,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

func GetUserIDFromClaims(claims *JWTClaims) uint {
	if claims == nil {
		return 0
	}
	ref := strings.TrimSpace(claims.UserID)
	if ref == "" {
		ref = strings.TrimSpace(claims.Sub)
	}
	if ref == "" {
		return 0
	}
	if parsed, err := strconv.ParseUint(ref, 10, 32); err == nil {
		return uint(parsed)
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(ref))
	return uint(hasher.Sum32()) + 1
}

const (
	ContextKeyUser     = "user"
	ContextKeyUserID   = "user_id"
	ContextKeyUserRole = "user_role"
	ContextKeyTenantID = "tenant_id"
	ContextKeyClaims   = "claims"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JWTAuth(jwtSecret string, allowedIssuers []string) gin.HandlerFunc {
	normalizedIssuers := normalizeIssuerAllowlist(allowedIssuers)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "MISSING_TOKEN",
				Message: "Authorization header is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "INVALID_TOKEN_FORMAT",
				Message: "Authorization header must be in 'Bearer <token>' format",
			})
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "INVALID_TOKEN",
				Message: "Token is required",
			})
			return
		}

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			message := "Invalid token"
			if errors.Is(err, jwt.ErrTokenExpired) {
				message = "Token has expired"
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

		normalizeClaims(claims)
		if !isAllowedIssuer(claims.Issuer, normalizedIssuers) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "INVALID_ISSUER",
				Message: "Token issuer is not allowed",
			})
			return
		}
		if claims.Role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "MISSING_ROLE",
				Message: "Token must contain role claims",
			})
			return
		}
		if claims.TenantID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "MISSING_TENANT",
				Message: "Token must contain tenant_id claim",
			})
			return
		}

		userID := GetUserIDFromClaims(claims)
		externalID := strings.TrimSpace(claims.UserID)
		if externalID == "" {
			externalID = strings.TrimSpace(claims.Sub)
		}

		user := models.User{
			ID:          userID,
			ExternalID:  externalID,
			Email:       claims.Email,
			Name:        claims.Name,
			TenantID:    claims.TenantID,
			Role:        claims.Role,
			Roles:       claims.Roles,
			Permissions: claims.Permissions,
			IsActive:    true,
		}

		c.Set(ContextKeyUser, user)
		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyUserRole, claims.Role)
		c.Set(ContextKeyTenantID, claims.TenantID)
		c.Set(ContextKeyClaims, claims)
		c.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	normalizedAllowed := make([]string, 0, len(allowedRoles))
	for _, role := range allowedRoles {
		value := strings.ToLower(strings.TrimSpace(role))
		if value == "" {
			continue
		}
		normalizedAllowed = append(normalizedAllowed, value)
	}

	return func(c *gin.Context) {
		user, exists := GetUserFromContext(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "NO_USER_CONTEXT",
				Message: "User context not found",
			})
			return
		}

		for _, allowed := range normalizedAllowed {
			if strings.EqualFold(user.Role, allowed) {
				c.Next()
				return
			}
			for _, role := range user.Roles {
				if strings.EqualFold(role, allowed) {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Code:    "INSUFFICIENT_PERMISSIONS",
			Message: "You do not have permission to access this resource",
		})
	}
}

func RequirePermission(permission string) gin.HandlerFunc {
	required := strings.ToLower(strings.TrimSpace(permission))

	return func(c *gin.Context) {
		user, exists := GetUserFromContext(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "unauthorized",
				Code:    "NO_USER_CONTEXT",
				Message: "User context not found",
			})
			return
		}

		if strings.EqualFold(user.Role, models.RoleAdmin) {
			c.Next()
			return
		}

		if hasPermission(user.Permissions, required) || models.HasPermission(user.Role, required) {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, ErrorResponse{
			Error:   "forbidden",
			Code:    "INSUFFICIENT_PERMISSIONS",
			Message: "You do not have permission to perform this action",
		})
	}
}

func GetUserFromContext(c *gin.Context) (models.User, bool) {
	user, exists := c.Get(ContextKeyUser)
	if !exists {
		return models.User{}, false
	}
	return user.(models.User), true
}

func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

func GetTenantIDFromContext(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get(ContextKeyTenantID)
	if !exists {
		return "", false
	}
	value, ok := tenantID.(string)
	return value, ok
}

func hasPermission(granted []string, required string) bool {
	if required == "" {
		return false
	}

	required = strings.ToLower(strings.TrimSpace(required))
	requiredHasScope := strings.Contains(required, ":")
	requiredSuffix := ":" + required

	for _, permission := range granted {
		value := strings.ToLower(strings.TrimSpace(permission))
		if value == "" {
			continue
		}
		if value == required || value == "*:*" {
			return true
		}

		if requiredHasScope {
			parts := strings.Split(value, ":")
			requiredParts := strings.Split(required, ":")
			if len(parts) == 2 && len(requiredParts) == 2 {
				if parts[0] == requiredParts[0] && parts[1] == "*" {
					return true
				}
			}
		} else if strings.HasSuffix(value, requiredSuffix) || value == required {
			return true
		}
	}

	return false
}

func normalizeClaims(claims *JWTClaims) {
	claims.Role = strings.ToLower(strings.TrimSpace(claims.Role))
	claims.TenantID = strings.TrimSpace(claims.TenantID)
	claims.Email = strings.ToLower(strings.TrimSpace(claims.Email))
	claims.UserID = strings.TrimSpace(claims.UserID)
	claims.Sub = strings.TrimSpace(claims.Sub)

	roles := make([]string, 0, len(claims.Roles))
	for _, role := range claims.Roles {
		normalized := strings.ToLower(strings.TrimSpace(role))
		if normalized == "" || slices.Contains(roles, normalized) {
			continue
		}
		roles = append(roles, normalized)
	}
	if claims.Role != "" && !slices.Contains(roles, claims.Role) {
		roles = append(roles, claims.Role)
	}
	if claims.Role == "" && len(roles) > 0 {
		claims.Role = primaryRole(roles)
	}
	if claims.Role != "" && len(roles) == 0 {
		roles = append(roles, claims.Role)
	}
	claims.Roles = roles

	permissions := make([]string, 0, len(claims.Permissions))
	for _, permission := range claims.Permissions {
		normalized := strings.ToLower(strings.TrimSpace(permission))
		if normalized == "" || slices.Contains(permissions, normalized) {
			continue
		}
		permissions = append(permissions, normalized)
	}
	claims.Permissions = permissions
}

func primaryRole(roles []string) string {
	priority := []string{models.RoleAdmin, models.RoleManager, models.RoleAgent, "user"}
	for _, candidate := range priority {
		if slices.Contains(roles, candidate) {
			return candidate
		}
	}
	if len(roles) == 0 {
		return "user"
	}
	return roles[0]
}

func normalizeIssuerAllowlist(issuers []string) []string {
	normalized := make([]string, 0, len(issuers))
	for _, issuer := range issuers {
		value := strings.ToLower(strings.TrimSpace(issuer))
		if value == "" || slices.Contains(normalized, value) {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func isAllowedIssuer(issuer string, allowedIssuers []string) bool {
	normalized := strings.ToLower(strings.TrimSpace(issuer))
	if normalized == "" {
		return false
	}
	if len(allowedIssuers) == 0 {
		return true
	}
	return slices.Contains(allowedIssuers, normalized)
}
