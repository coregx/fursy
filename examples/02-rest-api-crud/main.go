// Package main demonstrates a complete REST API with CRUD operations.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
)

func main() {
	// Create router.
	router := fursy.New()

	// Use middleware.
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// Create database and handlers.
	db := NewDatabase()
	handlers := NewHandlers(db)

	// Setup routes.
	setupRoutes(router, handlers)

	// Start server.
	slog.Info("Server starting", "port", 8080)
	slog.Info("API endpoints available:")
	slog.Info("  POST   /users       - Create user")
	slog.Info("  GET    /users       - List all users")
	slog.Info("  GET    /users/:id   - Get user by ID")
	slog.Info("  PUT    /users/:id   - Update user")
	slog.Info("  DELETE /users/:id   - Delete user")

	if err := http.ListenAndServe(":8080", router); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func setupRoutes(router *fursy.Router, h *Handlers) {
	// User CRUD routes.
	router.POST("/users", h.CreateUser)
	router.GET("/users", h.ListUsers)
	router.GET("/users/:id", h.GetUser)
	router.PUT("/users/:id", h.UpdateUser)
	router.DELETE("/users/:id", h.DeleteUser)
}
