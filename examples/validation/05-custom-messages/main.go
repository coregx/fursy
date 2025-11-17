// Package main demonstrates custom validation error messages with fursy validator plugin.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
	"github.com/coregx/fursy/plugins/validator"
)

// CreateUserRequest represents user creation with validation.
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Age      int    `json:"age" validate:"required,gte=18,lte=120"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// UserResponse represents the user creation response.
type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	// Create router.
	router := fursy.New()

	// Create validator with custom error messages.
	v := validator.New(&validator.Options{
		CustomMessages: map[string]string{
			// Generic messages.
			"required": "{field} is required and cannot be empty",
			"email":    "Please provide a valid email address for {field}",

			// String length messages.
			"min": "{field} must be at least {param} characters long",
			"max": "{field} must not exceed {param} characters",

			// Number comparison messages.
			"gte": "{field} must be {param} or greater",
			"lte": "{field} must be {param} or less",
		},
	})

	// Set validator with custom messages.
	router.SetValidator(v)

	// Use middleware.
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// Type-safe POST endpoint.
	fursy.POST[CreateUserRequest, UserResponse](router, "/users", createUser)

	// Start server.
	slog.Info("Server starting", "port", 8080)
	slog.Info("Custom error messages configured:")
	slog.Info("  required: {field} is required and cannot be empty")
	slog.Info("  email: Please provide a valid email address for {field}")
	slog.Info("  min/max: {field} must be at least/not exceed {param} characters")

	if err := http.ListenAndServe(":8080", router); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func createUser(c *fursy.Box[CreateUserRequest, UserResponse]) error {
	// Bind and validate - errors use custom messages!
	if err := c.Bind(); err != nil {
		return err
	}

	req := c.ReqBody

	// Simulate user creation.
	user := UserResponse{
		ID:       1,
		Username: req.Username,
		Email:    req.Email,
	}

	slog.Info("User created", "username", user.Username)

	return c.Created("/users/1", user)
}
