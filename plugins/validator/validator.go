// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validator

import (
	"errors"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator/v10 for use with fursy.
//
// It implements the fursy.Validator interface and automatically converts
// validation errors to fursy.ValidationErrors with RFC 9457 compliance.
type Validator struct {
	validate *validator.Validate
	options  *Options
}

// New creates a new Validator with optional configuration.
//
// Example:
//
//	// Default configuration
//	v := validator.New()
//
//	// With custom options
//	v := validator.New(&validator.Options{
//	    TagName: "validate",
//	    CustomMessages: map[string]string{
//	        "required": "{field} is required",
//	        "email": "{field} must be a valid email",
//	    },
//	})
func New(opts ...*Options) *Validator {
	var options *Options
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	} else {
		options = DefaultOptions()
	}

	validate := validator.New()

	// Use custom tag name if specified.
	if options.TagName != "" {
		validate.SetTagName(options.TagName)
	}

	// Use field names from JSON tags for better error messages.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Name
	})

	return &Validator{
		validate: validate,
		options:  options,
	}
}

// Validate validates the given data and returns validation errors.
//
// It implements the fursy.Validator interface.
//
// Returns nil if validation passes.
// Returns fursy.ValidationErrors if validation fails.
// Returns error for other types of errors (e.g., invalid input type).
func (v *Validator) Validate(data any) error {
	if data == nil {
		return nil
	}

	err := v.validate.Struct(data)
	if err == nil {
		return nil
	}

	// Convert validator.ValidationErrors to fursy.ValidationErrors.
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return v.convertErrors(validationErrs)
	}

	// Return other errors as-is (e.g., invalid type).
	return err
}

// RegisterCustomValidator registers a custom validation function.
//
// The tag parameter is the validation tag name (e.g., "custom_email").
// The fn parameter is the validation function that returns true if valid.
//
// Example:
//
//	v := validator.New()
//	v.RegisterCustomValidator("custom_email", func(fl validator.FieldLevel) bool {
//	    email := fl.Field().String()
//	    return strings.HasSuffix(email, "@example.com")
//	})
//
//	// Then use it in struct tags:
//	type User struct {
//	    Email string `validate:"required,custom_email"`
//	}
func (v *Validator) RegisterCustomValidator(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// RegisterTagNameFunc registers a function to get custom field names for errors.
//
// This is useful for using JSON tag names instead of struct field names in error messages.
//
// Example:
//
//	v := validator.New()
//	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
//	    name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
//	    if name == "-" {
//	        return ""
//	    }
//	    return name
//	})
func (v *Validator) RegisterTagNameFunc(fn validator.TagNameFunc) {
	v.validate.RegisterTagNameFunc(fn)
}

// RegisterAlias registers an alias for validation tags.
//
// This allows creating shorthand tags for complex validation rules.
//
// Example:
//
//	v := validator.New()
//	v.RegisterAlias("password", "required,min=8,max=72")
//
//	// Then use it in struct tags:
//	type User struct {
//	    Password string `validate:"password"`
//	}
func (v *Validator) RegisterAlias(alias, tags string) {
	v.validate.RegisterAlias(alias, tags)
}

// Struct validates a struct and returns validation errors.
//
// This is a convenience method that calls Validate internally.
func (v *Validator) Struct(data any) error {
	return v.Validate(data)
}

// Var validates a single variable value against validation tags.
//
// Example:
//
//	v := validator.New()
//	email := "invalid-email"
//	err := v.Var(email, "required,email")
//	// Returns validation error
func (v *Validator) Var(field any, tag string) error {
	err := v.validate.Var(field, tag)
	if err == nil {
		return nil
	}

	// Convert validator.ValidationErrors to fursy.ValidationErrors.
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		return v.convertErrors(validationErrs)
	}

	return err
}
