package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Server
	ServerPort string

	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string

	// JWT
	JWTSecret string
	JWTIssuer string

	// CORS
	CORSAllowedOrigins []string

	// Environment
	Environment string
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		// Server
		ServerPort: getEnv("SERVER_PORT", "3000"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "crm_db"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
		JWTIssuer: getEnv("JWT_ISSUER", "cms"),

		// CORS
		CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:3001"}),

		// Environment
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsSlice reads an environment variable as a comma-separated slice
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as an integer
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool reads an environment variable as a boolean
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// GetDSN returns the PostgreSQL connection string
func (c *Config) GetDSN() string {
	return "host=" + c.DBHost +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" port=" + c.DBPort +
		" sslmode=" + c.DBSSLMode +
		" TimeZone=UTC"
}
