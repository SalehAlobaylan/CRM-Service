package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SalehAlobaylan/CRM-Service/internal/config"
	"github.com/SalehAlobaylan/CRM-Service/internal/database"
	"github.com/SalehAlobaylan/CRM-Service/internal/middleware"
	"github.com/SalehAlobaylan/CRM-Service/internal/routes"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := middleware.InitLogger(cfg.IsDevelopment()); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer middleware.Logger.Sync()

	middleware.Logger.Info("Starting CRM Service...")

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		middleware.Logger.Fatal("Failed to connect to database: " + err.Error())
	}
	defer database.Close(db)

	middleware.Logger.Info("Connected to database")

	// Run migrations (AutoMigrate for development)
	if cfg.IsDevelopment() {
		middleware.Logger.Info("Running database migrations...")
		if err := database.AutoMigrate(db); err != nil {
			middleware.Logger.Fatal("Failed to run migrations: " + err.Error())
		}

		// Seed default data
		if err := database.SeedPipelineStages(db); err != nil {
			middleware.Logger.Warn("Failed to seed pipeline stages: " + err.Error())
		}
	}

	// Setup router
	router := routes.SetupRouter(db, cfg)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		middleware.Logger.Info("Server starting on port " + cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			middleware.Logger.Fatal("Failed to start server: " + err.Error())
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	middleware.Logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		middleware.Logger.Fatal("Server forced to shutdown: " + err.Error())
	}

	middleware.Logger.Info("Server exited gracefully")
}
