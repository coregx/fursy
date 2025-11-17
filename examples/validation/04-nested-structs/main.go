// Package main demonstrates nested struct validation with fursy validator plugin.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
	"github.com/coregx/fursy/plugins/validator"
)

// Address represents a physical address with validation.
type Address struct {
	Street  string `json:"street" validate:"required,min=5"`
	City    string `json:"city" validate:"required,min=2"`
	State   string `json:"state" validate:"required,len=2"`
	ZipCode string `json:"zip_code" validate:"required,numeric,len=5"`
	Country string `json:"country" validate:"required,len=2"`
}

// CreateUserRequest represents user creation with nested address validation.
type CreateUserRequest struct {
	Name    string   `json:"name" validate:"required,min=2,max=100"`
	Email   string   `json:"email" validate:"required,email"`
	Address *Address `json:"address" validate:"required"`
	Tags    []string `json:"tags" validate:"required,min=1,max=5,dive,min=2,max=20"`
}

// UserResponse represents the user creation response.
type UserResponse struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Address *Address `json:"address"`
	Tags    []string `json:"tags"`
}

func main() {
	// Create router.
	router := fursy.New()

	// Set validator plugin - validates nested structs automatically!
	router.SetValidator(validator.New())

	// Use middleware.
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// Type-safe POST endpoint with nested validation.
	fursy.POST[CreateUserRequest, UserResponse](router, "/users", createUser)

	// Start server.
	slog.Info("Server starting", "port", 8080)
	slog.Info("Demonstrates nested struct validation:")
	slog.Info("  - Address (required nested struct)")
	slog.Info("  - Tags (slice with dive validation)")

	if err := http.ListenAndServe(":8080", router); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func createUser(c *fursy.Box[CreateUserRequest, UserResponse]) error {
	// Bind request body and validate nested structures automatically.
	if err := c.Bind(); err != nil {
		return err
	}

	req := c.ReqBody

	// Simulate user creation.
	user := UserResponse{
		ID:      1,
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Tags:    req.Tags,
	}

	slog.Info("User created with address",
		"name", user.Name,
		"city", user.Address.City,
		"tags", len(user.Tags))

	return c.Created("/users/1", user)
}
