// Package main demonstrates basic validation with fursy validator plugin.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
	"github.com/coregx/fursy/plugins/validator"
)

// CreateUserRequest represents user creation request with validation tags.
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Age      int    `json:"age" validate:"gte=18,lte=120"`
	Password string `json:"password" validate:"required,min=8"`
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

	// Set validator plugin - enables automatic validation!
	router.SetValidator(validator.New())

	// Use middleware.
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// Type-safe POST endpoint with automatic validation.
	fursy.POST[CreateUserRequest, UserResponse](router, "/users", createUser)

	// Start server.
	slog.Info("Server starting", "port", 8080)
	slog.Info("Try: curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' -d '{\"email\":\"test@example.com\",\"username\":\"john\",\"age\":25,\"password\":\"secret123\"}'")

	if err := http.ListenAndServe(":8080", router); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

// createUser handles user creation with automatic validation.
func createUser(c *fursy.Box[CreateUserRequest, UserResponse]) error {
	// Bind request body and validate automatically.
	// If validation fails, returns RFC 9457 Problem Details.
	if err := c.Bind(); err != nil {
		return err
	}

	// ReqBody is now validated and type-safe!
	req := c.ReqBody

	// Simulate user creation (in real app, save to database).
	user := UserResponse{
		ID:       1,
		Username: req.Username,
		Email:    req.Email,
	}

	slog.Info("User created", "username", user.Username, "email", user.Email)

	// Return 201 Created with user data.
	return c.Created("/users/1", user)
}
