// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"fmt"
	"strings"
)

// Validator is the interface for request validation.
//
// Implementations can use any validation library (validator/v10, ozzo-validation, custom, etc.)
// or implement custom validation logic.
//
// Example implementations:
//   - plugins/validator: go-playground/validator/v10 integration
//   - Custom validator with business rules
type Validator interface {
	// Validate validates the given struct and returns validation errors.
	// Returns nil if validation passes.
	// Returns ValidationErrors for field-level errors.
	Validate(any) error
}

// ValidationError represents a single field validation error.
//
// It provides structured information about what field failed validation,
// which rule was violated, and a human-readable message.
type ValidationError struct {
	// Field is the name of the field that failed validation.
	// For nested structs, uses dot notation (e.g., "Address.City").
	Field string `json:"field"`

	// Tag is the validation rule that failed (e.g., "required", "email", "min").
	Tag string `json:"tag"`

	// Value is the actual value that failed validation.
	// Omitted from JSON if nil to avoid exposing sensitive data.
	Value any `json:"value,omitempty"`

	// Message is a human-readable error message.
	Message string `json:"message"`
}

// Error implements the error interface.
func (ve *ValidationError) Error() string {
	if ve.Message != "" {
		return ve.Message
	}
	return fmt.Sprintf("field '%s' failed '%s' validation", ve.Field, ve.Tag)
}

// ValidationErrors is a collection of validation errors.
//
// It implements the error interface and provides methods for error formatting.
type ValidationErrors []ValidationError

// Error implements the error interface.
// Returns a concatenated string of all validation errors.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}

	if len(ve) == 1 {
		return ve[0].Error()
	}

	var sb strings.Builder
	sb.WriteString("validation failed: ")
	for i, err := range ve {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Fields returns a map of field names to their error messages.
// Useful for API responses with per-field error details.
//
// Example:
//
//	{
//	  "email": "must be a valid email address",
//	  "age": "must be at least 18"
//	}
func (ve ValidationErrors) Fields() map[string]string {
	fields := make(map[string]string, len(ve))
	for _, err := range ve {
		fields[err.Field] = err.Message
	}
	return fields
}

// IsEmpty returns true if there are no validation errors.
func (ve ValidationErrors) IsEmpty() bool {
	return len(ve) == 0
}

// Add adds a validation error to the collection.
func (ve *ValidationErrors) Add(field, tag, message string) {
	*ve = append(*ve, ValidationError{
		Field:   field,
		Tag:     tag,
		Message: message,
	})
}

// AddError adds a validation error struct to the collection.
func (ve *ValidationErrors) AddError(err ValidationError) {
	*ve = append(*ve, err)
}
