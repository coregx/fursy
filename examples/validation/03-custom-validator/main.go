// Package main demonstrates custom validation rules with fursy validator plugin.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
	"github.com/coregx/fursy/plugins/validator"
)

// SignupRequest represents user signup with custom validation.
type SignupRequest struct {
	Email    string `json:"email" validate:"required,email,company_domain"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,strong_password"`
	Phone    string `json:"phone" validate:"required,phone"`
}

// SignupResponse represents signup response.
type SignupResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

func main() {
	// Create router.
	router := fursy.New()

	// Create validator and register custom validators.
	v := validator.New()

	// Register custom password validator.
	if err := v.RegisterCustomValidator("strong_password", validateStrongPassword); err != nil {
		slog.Error("Failed to register strong_password", "error", err)
		os.Exit(1)
	}

	// Register custom phone validator.
	if err := v.RegisterCustomValidator("phone", validatePhone); err != nil {
		slog.Error("Failed to register phone", "error", err)
		os.Exit(1)
	}

	// Register custom domain validator.
	if err := v.RegisterCustomValidator("company_domain", validateCompanyDomain); err != nil {
		slog.Error("Failed to register company_domain", "error", err)
		os.Exit(1)
	}

	// Set validator with custom rules.
	router.SetValidator(v)

	// Use middleware.
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// Setup routes.
	fursy.POST[SignupRequest, SignupResponse](router, "/signup", handleSignup)

	// Start server.
	slog.Info("Server starting", "port", 8080)
	slog.Info("Custom validators registered:")
	slog.Info("  strong_password - Password must have uppercase, lowercase, digit, special char")
	slog.Info("  phone - Phone number in valid format")
	slog.Info("  company_domain - Email must end with @example.com, @company.org, or @enterprise.net")

	if err := http.ListenAndServe(":8080", router); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func handleSignup(c *fursy.Box[SignupRequest, SignupResponse]) error {
	// Bind and validate with custom validators.
	if err := c.Bind(); err != nil {
		return err
	}

	req := c.ReqBody

	// Simulate user creation.
	resp := SignupResponse{
		ID:       1,
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
	}

	slog.Info("User signed up", "username", resp.Username, "email", resp.Email)

	return c.Created("/users/1", resp)
}
