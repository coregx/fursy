// Package main demonstrates production-ready setup with validation, JWT auth, and logging.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
	"github.com/coregx/fursy/plugins/validator"
)

func main() {
	// Load configuration.
	cfg := LoadConfig()

	// Setup structured logging.
	setupLogging(cfg.LogLevel)

	slog.Info("Starting application",
		"environment", cfg.Environment,
		"port", cfg.Port,
	)

	// Create router.
	router := fursy.New()

	// Setup validator with custom messages.
	router.SetValidator(setupValidator())

	// Global middleware.
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// Setup routes.
	setupRoutes(router, cfg)

	// Create HTTP server with timeouts.
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	go func() {
		slog.Info("Server started", "port", cfg.Port)
		slog.Info("Default credentials: admin@example.com / password123")
		slog.Info("API endpoints:")
		slog.Info("  POST   /api/login - Login and get JWT token")
		slog.Info("  GET    /api/profile - Get current user profile (requires auth)")
		slog.Info("  PUT    /api/profile - Update current user profile (requires auth)")
		slog.Info("  POST   /api/users - Create user (admin only)")
		slog.Info("  GET    /api/users - List all users (admin only)")

		if err := srv.ListenAndServe(); err != nil {
			slog.Error("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// Graceful shutdown with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited")
}

func setupLogging(level string) {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}

func setupValidator() *validator.Validator {
	return validator.New(&validator.Options{
		CustomMessages: map[string]string{
			"required": "{field} is required",
			"email":    "Please provide a valid email address",
			"min":      "{field} must be at least {param} characters",
			"max":      "{field} must not exceed {param} characters",
			"alphanum": "{field} must contain only letters and numbers",
			"oneof":    "{field} must be one of: {param}",
		},
	})
}

func setupRoutes(router *fursy.Router, cfg *Config) {
	// Public routes.
	fursy.POST[LoginRequest, LoginResponse](router, "/api/login", HandleLogin(cfg))

	// Protected routes (require authentication).
	router.Use(AuthMiddleware(cfg.JWTSecret))
	{
		fursy.GET[fursy.Empty, UserResponse](router, "/api/profile", HandleGetProfile)
		fursy.PUT[UpdateProfileRequest, UserResponse](router, "/api/profile", HandleUpdateProfile)

		// Admin-only routes.
		router.Use(RequireRole("admin"))
		{
			fursy.POST[CreateUserRequest, UserResponse](router, "/api/users", HandleCreateUser)
			fursy.GET[fursy.Empty, UserListResponse](router, "/api/users", HandleListUsers)
		}
	}
}
