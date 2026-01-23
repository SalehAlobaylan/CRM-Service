package models

// User represents user information extracted from JWT (CRM doesn't store users)
type User struct {
	ID       uint   `json:"id"`
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

// Role constants
const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleAgent   = "agent"
)

// Permission constants
const (
	PermissionRead      = "read"
	PermissionWrite     = "write"
	PermissionDelete    = "delete"
	PermissionManageAll = "manage_all"
	PermissionManageOwn = "manage_own"
)

// RolePermissions defines what each role can do
var RolePermissions = map[string][]string{
	RoleAdmin: {
		PermissionRead,
		PermissionWrite,
		PermissionDelete,
		PermissionManageAll,
	},
	RoleManager: {
		PermissionRead,
		PermissionWrite,
		PermissionDelete,
		PermissionManageAll,
	},
	RoleAgent: {
		PermissionRead,
		PermissionWrite,
		PermissionManageOwn,
	},
}

// HasPermission checks if a role has a specific permission
func HasPermission(role, permission string) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// CanManageAll checks if a role can manage all records
func CanManageAll(role string) bool {
	return HasPermission(role, PermissionManageAll)
}

// MeResponse is the response for GET /admin/me
type MeResponse struct {
	User        User     `json:"user"`
	Permissions []string `json:"permissions"`
}
