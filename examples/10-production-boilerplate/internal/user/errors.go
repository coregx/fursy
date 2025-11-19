package user

import (

	"github.com/coregx/fursy"
)

// MapError converts domain errors to HTTP responses (RFC 9457 Problem Details).
func MapError(c *fursy.Context, err error) error {
	// Map domain errors to HTTP status codes
	switch err {
	case ErrUserNotFound:
		return c.Problem(fursy.NotFound("User not found"))

	case ErrUserAlreadyExists:
		return c.Problem(fursy.Conflict("User already exists"))

	case ErrInvalidCredentials:
		return c.Problem(fursy.Unauthorized("Invalid credentials"))

	case ErrUserInactive:
		return c.Problem(fursy.Forbidden("Account is inactive"))

	default:
		// Generic 500 error
		return c.Problem(fursy.InternalServerError("An unexpected error occurred"))
	}
}
