package models

// Tag represents a tag/label for categorization
type Tag struct {
	BaseModel
	Name  string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Color string `gorm:"size:7" json:"color,omitempty"` // Hex color like #FF5733

	// Relations (many-to-many with customers)
	Customers []Customer `gorm:"many2many:customer_tags;" json:"customers,omitempty"`
}

// TableName specifies the table name for Tag
func (Tag) TableName() string {
	return "tags"
}

// CustomerTag represents the join table for customer-tag relationship
type CustomerTag struct {
	CustomerID uint `gorm:"primaryKey" json:"customer_id"`
	TagID      uint `gorm:"primaryKey" json:"tag_id"`
}

// TableName specifies the table name for CustomerTag
func (CustomerTag) TableName() string {
	return "customer_tags"
}

// TagListResponse is used for tag lists
type TagListResponse struct {
	Data  []Tag `json:"data"`
	Total int64 `json:"total"`
}
