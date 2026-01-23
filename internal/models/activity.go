package models

import (
	"time"
)

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeCall    ActivityType = "call"
	ActivityTypeEmail   ActivityType = "email"
	ActivityTypeMeeting ActivityType = "meeting"
	ActivityTypeTask    ActivityType = "task"
	ActivityTypeNote    ActivityType = "note"
)

// ActivityStatus represents the status of an activity
type ActivityStatus string

const (
	ActivityStatusScheduled ActivityStatus = "scheduled"
	ActivityStatusCompleted ActivityStatus = "completed"
	ActivityStatusCancelled ActivityStatus = "cancelled"
	ActivityStatusOverdue   ActivityStatus = "overdue"
)

// Activity represents a CRM activity (call, email, meeting, task)
type Activity struct {
	BaseModel
	Title       string         `gorm:"size:255;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Type        ActivityType   `gorm:"size:50;not null" json:"type"`
	Status      ActivityStatus `gorm:"size:50;default:'scheduled'" json:"status"`
	CustomerID  *uint          `gorm:"index" json:"customer_id,omitempty"`
	DealID      *uint          `gorm:"index" json:"deal_id,omitempty"`
	ContactID   *uint          `json:"contact_id,omitempty"`
	AssignedTo  *uint          `gorm:"index" json:"assigned_to,omitempty"`
	DueDate     *time.Time     `json:"due_date,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Duration    int            `json:"duration,omitempty"` // Duration in minutes
	Outcome     string         `gorm:"type:text" json:"outcome,omitempty"`
	Priority    string         `gorm:"size:20;default:'normal'" json:"priority"` // low, normal, high

	// Relations
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Deal     *Deal     `gorm:"foreignKey:DealID" json:"deal,omitempty"`
	Contact  *Contact  `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
}

// TableName specifies the table name for Activity
func (Activity) TableName() string {
	return "activities"
}

// ActivityListResponse is used for paginated activity lists
type ActivityListResponse struct {
	Data       []Activity `json:"data"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}
