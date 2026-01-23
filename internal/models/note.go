package models

// Note represents a note/comment attached to a customer or deal
type Note struct {
	BaseModel
	Content    string `gorm:"type:text;not null" json:"content"`
	CustomerID *uint  `gorm:"index" json:"customer_id,omitempty"`
	DealID     *uint  `gorm:"index" json:"deal_id,omitempty"`
	ActivityID *uint  `gorm:"index" json:"activity_id,omitempty"`
	AuthorID   uint   `gorm:"not null" json:"author_id"`
	AuthorName string `gorm:"size:255" json:"author_name,omitempty"`

	// Relations
	Customer *Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Deal     *Deal     `gorm:"foreignKey:DealID" json:"deal,omitempty"`
}

// TableName specifies the table name for Note
func (Note) TableName() string {
	return "notes"
}

// NoteListResponse is used for paginated note lists
type NoteListResponse struct {
	Data       []Note `json:"data"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}
