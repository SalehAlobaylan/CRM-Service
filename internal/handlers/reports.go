package handlers

import (
	"net/http"

	"github.com/SalehAlobaylan/CRM-Service/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReportHandler handles reporting endpoints
type ReportHandler struct {
	db *gorm.DB
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(db *gorm.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

// OverviewReport represents the overview report response
type OverviewReport struct {
	Customers     CustomerStats     `json:"customers"`
	Deals         DealStats         `json:"deals"`
	Activities    ActivityStats     `json:"activities"`
	RecentDeals   []models.Deal     `json:"recent_deals"`
	TopCustomers  []CustomerSummary `json:"top_customers"`
}

// CustomerStats represents customer statistics
type CustomerStats struct {
	Total    int64            `json:"total"`
	ByStatus map[string]int64 `json:"by_status"`
}

// DealStats represents deal statistics
type DealStats struct {
	Total           int64            `json:"total"`
	TotalValue      float64          `json:"total_value"`
	WonValue        float64          `json:"won_value"`
	WonCount        int64            `json:"won_count"`
	LostCount       int64            `json:"lost_count"`
	OpenCount       int64            `json:"open_count"`
	AverageDealSize float64          `json:"average_deal_size"`
	ByStage         map[string]int64 `json:"by_stage"`
}

// ActivityStats represents activity statistics
type ActivityStats struct {
	Total       int64            `json:"total"`
	Scheduled   int64            `json:"scheduled"`
	Completed   int64            `json:"completed"`
	Overdue     int64            `json:"overdue"`
	ByType      map[string]int64 `json:"by_type"`
}

// CustomerSummary represents a customer summary for reports
type CustomerSummary struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Company    string  `json:"company"`
	DealsCount int64   `json:"deals_count"`
	DealsValue float64 `json:"deals_value"`
}

// GetOverview returns an overview report
// GET /admin/reports/overview
func (h *ReportHandler) GetOverview(c *gin.Context) {
	report := OverviewReport{
		Customers:  h.getCustomerStats(),
		Deals:      h.getDealStats(),
		Activities: h.getActivityStats(),
	}

	// Get recent deals
	var recentDeals []models.Deal
	h.db.Preload("Customer").Order("created_at DESC").Limit(5).Find(&recentDeals)
	report.RecentDeals = recentDeals

	// Get top customers by deal value
	report.TopCustomers = h.getTopCustomers(5)

	c.JSON(http.StatusOK, report)
}

// getCustomerStats returns customer statistics
func (h *ReportHandler) getCustomerStats() CustomerStats {
	stats := CustomerStats{
		ByStatus: make(map[string]int64),
	}

	// Total customers
	h.db.Model(&models.Customer{}).Count(&stats.Total)

	// By status
	statuses := []models.CustomerStatus{
		models.CustomerStatusLead,
		models.CustomerStatusProspect,
		models.CustomerStatusActive,
		models.CustomerStatusInactive,
		models.CustomerStatusChurned,
	}

	for _, status := range statuses {
		var count int64
		h.db.Model(&models.Customer{}).Where("status = ?", status).Count(&count)
		stats.ByStatus[string(status)] = count
	}

	return stats
}

// getDealStats returns deal statistics
func (h *ReportHandler) getDealStats() DealStats {
	stats := DealStats{
		ByStage: make(map[string]int64),
	}

	// Total deals
	h.db.Model(&models.Deal{}).Count(&stats.Total)

	// Total value
	h.db.Model(&models.Deal{}).Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalValue)

	// Won deals
	h.db.Model(&models.Deal{}).Where("stage = ?", models.DealStageClosedWon).Count(&stats.WonCount)
	h.db.Model(&models.Deal{}).Where("stage = ?", models.DealStageClosedWon).Select("COALESCE(SUM(amount), 0)").Scan(&stats.WonValue)

	// Lost deals
	h.db.Model(&models.Deal{}).Where("stage = ?", models.DealStageClosedLost).Count(&stats.LostCount)

	// Open deals
	h.db.Model(&models.Deal{}).Where("stage NOT IN ?", []string{
		string(models.DealStageClosedWon),
		string(models.DealStageClosedLost),
	}).Count(&stats.OpenCount)

	// Average deal size
	if stats.Total > 0 {
		stats.AverageDealSize = stats.TotalValue / float64(stats.Total)
	}

	// By stage
	for _, stage := range models.ValidDealStages {
		var count int64
		h.db.Model(&models.Deal{}).Where("stage = ?", stage).Count(&count)
		stats.ByStage[string(stage)] = count
	}

	return stats
}

// getActivityStats returns activity statistics
func (h *ReportHandler) getActivityStats() ActivityStats {
	stats := ActivityStats{
		ByType: make(map[string]int64),
	}

	// Total activities
	h.db.Model(&models.Activity{}).Count(&stats.Total)

	// By status
	h.db.Model(&models.Activity{}).Where("status = ?", models.ActivityStatusScheduled).Count(&stats.Scheduled)
	h.db.Model(&models.Activity{}).Where("status = ?", models.ActivityStatusCompleted).Count(&stats.Completed)
	h.db.Model(&models.Activity{}).Where("status = ?", models.ActivityStatusOverdue).Count(&stats.Overdue)

	// By type
	types := []models.ActivityType{
		models.ActivityTypeCall,
		models.ActivityTypeEmail,
		models.ActivityTypeMeeting,
		models.ActivityTypeTask,
		models.ActivityTypeNote,
	}

	for _, t := range types {
		var count int64
		h.db.Model(&models.Activity{}).Where("type = ?", t).Count(&count)
		stats.ByType[string(t)] = count
	}

	return stats
}

// getTopCustomers returns top customers by deal value
func (h *ReportHandler) getTopCustomers(limit int) []CustomerSummary {
	var results []CustomerSummary

	h.db.Model(&models.Customer{}).
		Select("customers.id, customers.name, customers.email, customers.company, COUNT(deals.id) as deals_count, COALESCE(SUM(deals.amount), 0) as deals_value").
		Joins("LEFT JOIN deals ON deals.customer_id = customers.id AND deals.deleted_at IS NULL").
		Group("customers.id, customers.name, customers.email, customers.company").
		Order("deals_value DESC").
		Limit(limit).
		Scan(&results)

	return results
}
