package models

// Contact represents a contact person for a customer
type Contact struct {
	BaseModel
	CustomerID uint   `gorm:"not null;index" json:"customer_id"`
	FirstName  string `gorm:"size:100;not null" json:"first_name"`
	LastName   string `gorm:"size:100" json:"last_name,omitempty"`
	Email      string `gorm:"size:255" json:"email,omitempty"`
	Phone      string `gorm:"size:50" json:"phone,omitempty"`
	Position   string `gorm:"size:100" json:"position,omitempty"`
	IsPrimary  bool   `gorm:"default:false" json:"is_primary"`
	Notes      string `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Customer Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

// TableName specifies the table name for Contact
func (Contact) TableName() string {
	return "contacts"
}

// ContactListResponse is used for paginated contact lists
type ContactListResponse struct {
	Data       []Contact `json:"data"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}
