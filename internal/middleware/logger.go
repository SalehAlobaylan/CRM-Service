package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global structured logger
var Logger *zap.Logger

// InitLogger initializes the global logger
func InitLogger(isDevelopment bool) error {
	var config zap.Config

	if isDevelopment {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Logger = logger
	return nil
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// StructuredLogger creates a structured logging middleware
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID from context
		requestID, _ := c.Get("request_id")

		// Get user ID if authenticated
		userID, _ := c.Get(ContextKeyUserID)

		// Build log fields
		fields := []zap.Field{
			zap.String("request_id", requestID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", c.Writer.Size()),
		}

		// Add user ID if present
		if userID != nil {
			fields = append(fields, zap.Any("user_id", userID))
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// Log based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			Logger.Error("Server error", fields...)
		case status >= 400:
			Logger.Warn("Client error", fields...)
		case status >= 300:
			Logger.Info("Redirect", fields...)
		default:
			Logger.Info("Request", fields...)
		}
	}
}

// Recovery creates a panic recovery middleware with logging
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")

				Logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("request_id", requestID.(string)),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				c.AbortWithStatusJSON(500, ErrorResponse{
					Error:   "internal_server_error",
					Code:    "INTERNAL_ERROR",
					Message: "An unexpected error occurred",
				})
			}
		}()

		c.Next()
	}
}
