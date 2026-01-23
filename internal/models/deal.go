package models

import (
	"time"
)

// DealStage represents the stage of a deal in the pipeline
type DealStage string

const (
	DealStageProspecting  DealStage = "prospecting"
	DealStageQualification DealStage = "qualification"
	DealStageProposal     DealStage = "proposal"
	DealStageNegotiation  DealStage = "negotiation"
	DealStageClosedWon    DealStage = "closed_won"
	DealStageClosedLost   DealStage = "closed_lost"
)

// ValidDealStages contains all valid deal stages for validation
var ValidDealStages = []DealStage{
	DealStageProspecting,
	DealStageQualification,
	DealStageProposal,
	DealStageNegotiation,
	DealStageClosedWon,
	DealStageClosedLost,
}

// IsValidDealStage checks if a stage is valid
func IsValidDealStage(stage DealStage) bool {
	for _, s := range ValidDealStages {
		if s == stage {
			return true
		}
	}
	return false
}

// Deal represents a sales opportunity
type Deal struct {
	BaseModel
	Title             string     `gorm:"size:255;not null" json:"title"`
	Description       string     `gorm:"type:text" json:"description,omitempty"`
	CustomerID        uint       `gorm:"not null;index" json:"customer_id"`
	ContactID         *uint      `json:"contact_id,omitempty"`
	Stage             DealStage  `gorm:"size:50;default:'prospecting'" json:"stage"`
	Amount            float64    `gorm:"type:decimal(15,2);default:0" json:"amount"`
	Currency          string     `gorm:"size:3;default:'USD'" json:"currency"`
	Probability       int        `gorm:"default:0" json:"probability"` // 0-100
	ExpectedCloseDate *time.Time `json:"expected_close_date,omitempty"`
	ActualCloseDate   *time.Time `json:"actual_close_date,omitempty"`
	OwnerID           *uint      `json:"owner_id,omitempty"`
	LostReason        string     `gorm:"size:255" json:"lost_reason,omitempty"`

	// Relations
	Customer   Customer   `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Contact    *Contact   `gorm:"foreignKey:ContactID" json:"contact,omitempty"`
	Activities []Activity `gorm:"foreignKey:DealID" json:"activities,omitempty"`
	Notes      []Note     `gorm:"foreignKey:DealID" json:"notes,omitempty"`
}

// TableName specifies the table name for Deal
func (Deal) TableName() string {
	return "deals"
}

// DealListResponse is used for paginated deal lists
type DealListResponse struct {
	Data       []Deal `json:"data"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}

// PipelineStage represents a configurable pipeline stage
type PipelineStage struct {
	BaseModel
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	DisplayName string `gorm:"size:100;not null" json:"display_name"`
	Order       int    `gorm:"not null" json:"order"`
	Color       string `gorm:"size:7" json:"color,omitempty"` // Hex color
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

// TableName specifies the table name for PipelineStage
func (PipelineStage) TableName() string {
	return "pipeline_stages"
}
