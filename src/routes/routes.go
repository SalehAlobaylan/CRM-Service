package routes

import (
	"github.com/SalehAlobaylan/CRM-Service/src/config"
	"github.com/SalehAlobaylan/CRM-Service/src/handlers"
	"github.com/SalehAlobaylan/CRM-Service/src/middleware"
	"github.com/SalehAlobaylan/CRM-Service/src/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery())
	router.Use(middleware.StructuredLogger())
	router.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	customerHandler := handlers.NewCustomerHandler(db)
	contactHandler := handlers.NewContactHandler(db)
	dealHandler := handlers.NewDealHandler(db)
	activityHandler := handlers.NewActivityHandler(db)
	tagHandler := handlers.NewTagHandler(db)
	reportHandler := handlers.NewReportHandler(db)
	healthHandler := handlers.NewHealthHandler(db)

	// Public routes (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	router.GET("/metrics", healthHandler.Metrics())

	// Admin routes (JWT auth required)
	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Auth endpoints
		admin.GET("/me", authHandler.GetMe)
		admin.GET("/me/activities", activityHandler.GetMyActivities)

		// Customer endpoints
		customers := admin.Group("/customers")
		{
			customers.GET("", customerHandler.ListCustomers)
			customers.POST("", middleware.RequirePermission(models.PermissionWrite), customerHandler.CreateCustomer)
			customers.GET("/:id", customerHandler.GetCustomer)
			customers.PUT("/:id", middleware.RequirePermission(models.PermissionWrite), customerHandler.UpdateCustomer)
			customers.PATCH("/:id", middleware.RequirePermission(models.PermissionWrite), customerHandler.PatchCustomer)
			customers.DELETE("/:id", middleware.RequirePermission(models.PermissionDelete), customerHandler.DeleteCustomer)

			// Nested contacts under customers
			customers.GET("/:id/contacts", contactHandler.ListContacts)
			customers.POST("/:id/contacts", middleware.RequirePermission(models.PermissionWrite), contactHandler.CreateContact)

			// Customer tags
			customers.POST("/:id/tags/:tagId", middleware.RequirePermission(models.PermissionWrite), tagHandler.AssignTagToCustomer)
			customers.DELETE("/:id/tags/:tagId", middleware.RequirePermission(models.PermissionWrite), tagHandler.RemoveTagFromCustomer)
		}

		// Contact endpoints (for update/delete by contact ID)
		contacts := admin.Group("/contacts")
		{
			contacts.PUT("/:id", middleware.RequirePermission(models.PermissionWrite), contactHandler.UpdateContact)
			contacts.DELETE("/:id", middleware.RequirePermission(models.PermissionDelete), contactHandler.DeleteContact)
		}

		// Deal endpoints
		deals := admin.Group("/deals")
		{
			deals.GET("", dealHandler.ListDeals)
			deals.POST("", middleware.RequirePermission(models.PermissionWrite), dealHandler.CreateDeal)
			deals.GET("/:id", dealHandler.GetDeal)
			deals.PUT("/:id", middleware.RequirePermission(models.PermissionWrite), dealHandler.UpdateDeal)
			deals.PATCH("/:id", middleware.RequirePermission(models.PermissionWrite), dealHandler.PatchDeal)
			deals.DELETE("/:id", middleware.RequirePermission(models.PermissionDelete), dealHandler.DeleteDeal)
		}

		// Activity endpoints
		activities := admin.Group("/activities")
		{
			activities.GET("", activityHandler.ListActivities)
			activities.POST("", middleware.RequirePermission(models.PermissionWrite), activityHandler.CreateActivity)
			activities.GET("/:id", activityHandler.GetActivity)
			activities.PUT("/:id", middleware.RequirePermission(models.PermissionWrite), activityHandler.UpdateActivity)
			activities.PATCH("/:id", middleware.RequirePermission(models.PermissionWrite), activityHandler.PatchActivity)
			activities.DELETE("/:id", middleware.RequirePermission(models.PermissionDelete), activityHandler.DeleteActivity)
		}

		// Tag endpoints
		tags := admin.Group("/tags")
		{
			tags.GET("", tagHandler.ListTags)
			tags.POST("", middleware.RequireRole(models.RoleAdmin), tagHandler.CreateTag)
			tags.PUT("/:id", middleware.RequireRole(models.RoleAdmin), tagHandler.UpdateTag)
			tags.DELETE("/:id", middleware.RequireRole(models.RoleAdmin), tagHandler.DeleteTag)
		}

		// Report endpoints
		reports := admin.Group("/reports")
		{
			reports.GET("/overview", reportHandler.GetOverview)
		}
	}

	return router
}
