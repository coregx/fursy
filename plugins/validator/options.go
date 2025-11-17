// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package validator provides integration with go-playground/validator/v10 for fursy.
//
// This plugin allows automatic validation of request bodies using struct tags,
// with automatic conversion to fursy's RFC 9457 compliant error format.
//
// Example:
//
//	import (
//	    "github.com/coregx/fursy"
//	    "github.com/coregx/fursy/plugins/validator"
//	)
//
//	func main() {
//	    r := fursy.New()
//	    r.SetValidator(validator.New())
//
//	    r.POST[CreateUserRequest, UserResponse]("/users", createUserHandler)
//	    r.Run(":8080")
//	}
package validator

// Options configures the validator behavior.
type Options struct {
	// TagName is the struct tag name for validation rules.
	// Default: "validate"
	//
	// Example:
	//   type User struct {
	//       Email string `validate:"required,email"`
	//   }
	TagName string

	// CustomMessages provides custom error messages for specific tags.
	// The key is the tag name (e.g., "required", "email").
	// The value supports placeholders: {field}, {value}, {param}
	//
	// Example:
	//   CustomMessages: map[string]string{
	//       "required": "{field} is required",
	//       "email": "{field} must be a valid email address",
	//       "min": "{field} must be at least {param} characters",
	//   }
	CustomMessages map[string]string
}

// DefaultOptions returns the default validator options.
func DefaultOptions() *Options {
	return &Options{
		TagName:        "validate",
		CustomMessages: nil,
	}
}
