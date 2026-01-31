package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel contains common columns for all tables
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// CustomerStatus represents the status of a customer
type CustomerStatus string

const (
	CustomerStatusLead      CustomerStatus = "lead"
	CustomerStatusProspect  CustomerStatus = "prospect"
	CustomerStatusActive    CustomerStatus = "active"
	CustomerStatusInactive  CustomerStatus = "inactive"
	CustomerStatusChurned   CustomerStatus = "churned"
)

// Customer represents a customer in the CRM
type Customer struct {
	BaseModel
	Name           string         `gorm:"size:255;not null" json:"name"`
	Email          string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Phone          string         `gorm:"size:50" json:"phone,omitempty"`
	Company        string         `gorm:"size:255" json:"company,omitempty"`
	Role           string         `gorm:"size:100" json:"role,omitempty"`
	Status         CustomerStatus `gorm:"size:50;default:'lead'" json:"status"`
	AssignedTo     *uint          `json:"assigned_to,omitempty"`
	Contacted      bool           `gorm:"default:false" json:"contacted"`
	NextFollowUpAt *time.Time     `json:"next_follow_up_at,omitempty"`
	Notes          string         `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Contacts   []Contact   `gorm:"foreignKey:CustomerID" json:"contacts,omitempty"`
	Deals      []Deal      `gorm:"foreignKey:CustomerID" json:"deals,omitempty"`
	Activities []Activity  `gorm:"foreignKey:CustomerID" json:"activities,omitempty"`
	Tags       []Tag       `gorm:"many2many:customer_tags;" json:"tags,omitempty"`
}

// TableName specifies the table name for Customer
func (Customer) TableName() string {
	return "customers"
}

// CustomerListResponse is used for paginated customer lists
type CustomerListResponse struct {
	Data       []Customer `json:"data"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// CustomerDetailResponse includes customer with related entities summary
type CustomerDetailResponse struct {
	Customer
	ContactsCount          int        `json:"contacts_count"`
	OpenDealsCount         int        `json:"open_deals_count"`
	UpcomingActivitiesCount int       `json:"upcoming_activities_count"`
	RecentActivities       []Activity `json:"recent_activities,omitempty"`
}
