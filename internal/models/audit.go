package models

import (
	"time"
)

// AuditAction represents the type of audit action
type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
)

// AuditLog represents an immutable audit trail entry
type AuditLog struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	ResourceType string      `gorm:"size:100;not null;index" json:"resource_type"` // customer, deal, activity, etc.
	ResourceID   uint        `gorm:"not null;index" json:"resource_id"`
	Action       AuditAction `gorm:"size:50;not null" json:"action"`
	UserID       uint        `gorm:"not null;index" json:"user_id"`
	UserName     string      `gorm:"size:255" json:"user_name,omitempty"`
	UserRole     string      `gorm:"size:50" json:"user_role,omitempty"`
	OldValues    string      `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues    string      `gorm:"type:jsonb" json:"new_values,omitempty"`
	IPAddress    string      `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent    string      `gorm:"size:500" json:"user_agent,omitempty"`
	CreatedAt    time.Time   `gorm:"not null" json:"created_at"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditLogListResponse is used for paginated audit log lists
type AuditLogListResponse struct {
	Data       []AuditLog `json:"data"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}
