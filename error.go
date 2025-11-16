// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package fursy provides common HTTP errors for the FURSY router.
package fursy

import "errors"

// Common HTTP errors.
var (
	// ErrUnauthorized is returned when authentication fails.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when user is authenticated but not authorized.
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = errors.New("not found")

	// ErrBadRequest is returned when the request is invalid.
	ErrBadRequest = errors.New("bad request")

	// ErrInternalServerError is returned for server errors.
	ErrInternalServerError = errors.New("internal server error")
)
