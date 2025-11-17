package main

import (
	"os"
	"strconv"
)

// Config represents application configuration.
type Config struct {
	Port         string
	JWTSecret    string
	LogLevel     string
	Environment  string
	ReadTimeout  int
	WriteTimeout int
}

// LoadConfig loads configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		ReadTimeout:  getEnvInt("READ_TIMEOUT", 15),
		WriteTimeout: getEnvInt("WRITE_TIMEOUT", 15),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
