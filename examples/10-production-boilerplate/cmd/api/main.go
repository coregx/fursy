// Production Boilerplate - DDD Architecture with fursy + stream.
//
// This is a complete, production-ready REST API with real-time features demonstrating:
// - Domain-Driven Design (DDD) with Rich Models
// - Clean Architecture (by meaning, not directories)
// - JWT Authentication
// - SSE Notifications (Server-Sent Events)
// - WebSocket Chat
// - Database persistence with SQLite
// - RFC 9457 Problem Details
// - Graceful shutdown
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/production-boilerplate/internal/chat"
	"example.com/production-boilerplate/internal/config"
	"example.com/production-boilerplate/internal/notification"
	"example.com/production-boilerplate/internal/shared/auth"
	"example.com/production-boilerplate/internal/shared/database"
	"example.com/production-boilerplate/internal/user"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
	"github.com/coregx/stream/sse"
	"github.com/coregx/stream/websocket"
)

func main() {
	// Load configuration from environment
	cfg := config.Load()

	// Setup structured logger (log/slog)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting application", "env", cfg.Env, "port", cfg.Port)

	// Connect to database
	db, err := database.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	slog.Info("Database connected", "dsn", cfg.DatabaseDSN)

	// Initialize JWT service
	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration)

	// Initialize repositories
	userRepo := user.NewRepository(db.DB)

	// Initialize services
	userService := user.NewService(userRepo, jwtService)

	// Initialize SSE Hub for notifications
	notificationHub := sse.NewHub[*notification.Notification]()
	go notificationHub.Run() // Start hub goroutine
	notificationService := notification.NewService(notificationHub)

	// Initialize WebSocket Hub for chat
	chatHub := websocket.NewHub()
	go chatHub.Run() // Start hub goroutine
	chatService := chat.NewService(chatHub)

	// Initialize API handlers
	userAPI := user.NewAPI(userService)
	notificationAPI := notification.NewAPI(notificationService, notificationHub)
	chatAPI := chat.NewAPI(chatService, chatHub)

	// Create fursy router
	router := fursy.New()

	// Global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *fursy.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"env":    cfg.Env,
		})
	})

	// API documentation endpoint
	router.GET("/", func(c *fursy.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"name":    "Production Boilerplate API",
			"version": "1.0.0",
			"docs":    "/api/docs",
			"health":  "/health",
			"endpoints": map[string]interface{}{
				"auth": map[string]string{
					"register": "POST /api/auth/register",
					"login":    "POST /api/auth/login",
				},
				"user": map[string]string{
					"profile":        "GET /api/users/me",
					"updateProfile":  "PUT /api/users/me",
					"changePassword": "POST /api/users/me/password",
					"listUsers":      "GET /api/users (admin only)",
					"banUser":        "POST /api/users/:id/ban (admin only)",
					"promoteUser":    "POST /api/users/:id/promote (admin only)",
				},
				"notifications": map[string]string{
					"stream":    "GET /api/notifications/stream (SSE)",
					"broadcast": "POST /api/notifications/broadcast (admin only)",
				},
				"chat": map[string]string{
					"websocket": "GET /api/chat/ws (WebSocket)",
				},
			},
		})
	})

	// Register routes with auth middleware
	authMiddleware := auth.Middleware(jwtService)
	userAPI.RegisterRoutes(router, authMiddleware)
	notificationAPI.RegisterRoutes(router, authMiddleware)
	chatAPI.RegisterRoutes(router, authMiddleware)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		slog.Info("Shutting down server...")

		// Shutdown SSE and WebSocket hubs
		notificationHub.Close()
		chatHub.Close()

		// Shutdown HTTP server
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("Server shutdown error", "error", err)
		}

		slog.Info("Server stopped")
	}()

	// Start server
	slog.Info("Server starting", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error:", err)
	}
}
