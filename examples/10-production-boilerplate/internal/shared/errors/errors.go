// Package errors defines domain errors used across bounded contexts.
package errors

import "errors"

// Common domain errors.
var (
	// Generic errors
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
)
