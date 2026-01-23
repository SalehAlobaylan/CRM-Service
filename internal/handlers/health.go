package handlers

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

// HealthHandler handles health check and metrics endpoints
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks"`
}

// Health returns the health status of the service
// GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	response := HealthResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Checks:  make(map[string]string),
	}

	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		response.Status = "unhealthy"
		response.Checks["database"] = "error: " + err.Error()
	} else {
		if err := sqlDB.Ping(); err != nil {
			response.Status = "unhealthy"
			response.Checks["database"] = "error: " + err.Error()
		} else {
			response.Checks["database"] = "ok"
		}
	}

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	response.Checks["memory"] = "ok"

	// Set HTTP status based on health
	statusCode := http.StatusOK
	if response.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Metrics returns Prometheus metrics
// GET /metrics
func (h *HealthHandler) Metrics() gin.HandlerFunc {
	// Register custom metrics
	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crm_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "crm_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Register metrics (ignore if already registered)
	prometheus.Register(httpRequestsTotal)
	prometheus.Register(httpRequestDuration)

	return gin.WrapH(promhttp.Handler())
}

// Ready returns the readiness status
// GET /ready
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check if service is ready to accept traffic
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"error":  err.Error(),
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not_ready",
			"error":  "database not available",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
