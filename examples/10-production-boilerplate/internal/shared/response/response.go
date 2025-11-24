// Package response provides HTTP response helpers.
package response

import (
	"errors"

	"example.com/production-boilerplate/internal/user"
	"github.com/coregx/fursy"
)

// Error converts domain errors to HTTP responses (RFC 9457 Problem Details).
func Error(c *fursy.Context, err error) error {
	// Map domain errors to HTTP status codes
	switch {
	case errors.Is(err, user.ErrUserNotFound):
		return c.Problem(fursy.NotFound("User Not Found: " + err.Error()))

	case errors.Is(err, user.ErrUserAlreadyExists):
		return c.Problem(fursy.Conflict("User Already Exists: " + err.Error()))

	case errors.Is(err, user.ErrInvalidCredentials):
		return c.Problem(fursy.Unauthorized("Invalid Credentials: Email or password is incorrect"))

	case errors.Is(err, user.ErrUserInactive):
		return c.Problem(fursy.Forbidden("User Inactive: Your account is inactive"))

	default:
		// Generic 500 error
		return c.Problem(fursy.InternalServerError("Internal Server Error: An unexpected error occurred"))
	}
}

// ValidationError returns validation error response.
func ValidationError(c *fursy.Context, err error) error {
	return c.Problem(fursy.BadRequest("Validation Error: " + err.Error()))
}
