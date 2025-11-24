// Package config provides application configuration loading from environment variables.
package config

import (
	"log"
	"os"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Server settings
	Port string
	Env  string

	// Database settings
	DatabaseDSN string

	// JWT settings
	JWTSecret     []byte
	JWTExpiration time.Duration

	// CORS settings
	CORSOrigins []string
}

// Load loads configuration from environment variables.
func Load() *Config {
	// Load JWT expiration
	jwtExpiration, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		log.Fatal("Invalid JWT_EXPIRATION:", err)
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		Env:           getEnv("ENV", "development"),
		DatabaseDSN:   getEnv("DATABASE_DSN", "./data/app.db"),
		JWTSecret:     []byte(getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production")),
		JWTExpiration: jwtExpiration,
		CORSOrigins:   []string{getEnv("CORS_ORIGINS", "http://localhost:3000,http://localhost:8080")},
	}
}

// getEnv retrieves environment variable with default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
