package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SalehAlobaylan/CRM-Service/internal/config"
	"github.com/SalehAlobaylan/CRM-Service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes connection to the PostgreSQL database
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// Configure GORM logger
	logLevel := logger.Warn
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  cfg.IsDevelopment(),
		},
	)

	// Open connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	return db, nil
}

// AutoMigrate runs GORM AutoMigrate for all models
// Note: Use golang-migrate for production, AutoMigrate for development only
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Customer{},
		&models.Contact{},
		&models.Deal{},
		&models.PipelineStage{},
		&models.Activity{},
		&models.Note{},
		&models.Tag{},
		&models.AuditLog{},
	)
}

// SeedPipelineStages seeds default pipeline stages if not present
func SeedPipelineStages(db *gorm.DB) error {
	stages := []models.PipelineStage{
		{Name: "prospecting", DisplayName: "Prospecting", Order: 1, Color: "#6366f1", IsActive: true},
		{Name: "qualification", DisplayName: "Qualification", Order: 2, Color: "#8b5cf6", IsActive: true},
		{Name: "proposal", DisplayName: "Proposal", Order: 3, Color: "#a855f7", IsActive: true},
		{Name: "negotiation", DisplayName: "Negotiation", Order: 4, Color: "#f59e0b", IsActive: true},
		{Name: "closed_won", DisplayName: "Closed Won", Order: 5, Color: "#22c55e", IsActive: true},
		{Name: "closed_lost", DisplayName: "Closed Lost", Order: 6, Color: "#ef4444", IsActive: true},
	}

	for _, stage := range stages {
		// Use FirstOrCreate to avoid duplicates
		var existing models.PipelineStage
		result := db.Where("name = ?", stage.Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&stage).Error; err != nil {
				return fmt.Errorf("failed to seed pipeline stage %s: %w", stage.Name, err)
			}
		}
	}

	return nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
