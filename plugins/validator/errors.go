// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validator

import (
	"fmt"
	"strings"

	"github.com/coregx/fursy"
	"github.com/go-playground/validator/v10"
)

const (
	kindString = "string"
	kindSlice  = "slice"
	kindArray  = "array"
	kindMap    = "map"
)

// convertErrors converts validator.ValidationErrors to fursy.ValidationErrors.
func (v *Validator) convertErrors(errs validator.ValidationErrors) fursy.ValidationErrors {
	var result fursy.ValidationErrors

	for _, err := range errs {
		result.AddError(fursy.ValidationError{
			Field:   err.Field(),
			Tag:     err.Tag(),
			Value:   err.Value(),
			Message: v.formatMessage(err),
		})
	}

	return result
}

// formatMessage creates a human-readable error message for a validation error.
func (v *Validator) formatMessage(err validator.FieldError) string {
	// Check if custom message exists for this tag.
	if v.options.CustomMessages != nil {
		if template, ok := v.options.CustomMessages[err.Tag()]; ok {
			return v.interpolateMessage(template, err)
		}
	}

	// Use default message based on tag.
	return v.defaultMessage(err)
}

// interpolateMessage replaces placeholders in message template.
// Supported placeholders: {field}, {value}, {param}.
func (v *Validator) interpolateMessage(template string, err validator.FieldError) string {
	msg := template
	msg = strings.ReplaceAll(msg, "{field}", err.Field())
	msg = strings.ReplaceAll(msg, "{value}", fmt.Sprintf("%v", err.Value()))
	msg = strings.ReplaceAll(msg, "{param}", err.Param())
	return msg
}

// defaultMessage provides default error messages for common validation tags.
//
//nolint:gocyclo,cyclop,funlen // Comprehensive switch for all validation tags - complexity is acceptable
func (v *Validator) defaultMessage(err validator.FieldError) string {
	field := err.Field()
	param := err.Param()

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "min":
		switch err.Kind().String() {
		case kindString:
			return fmt.Sprintf("%s must be at least %s characters long", field, param)
		case kindSlice, kindArray, kindMap:
			return fmt.Sprintf("%s must contain at least %s items", field, param)
		default:
			return fmt.Sprintf("%s must be at least %s", field, param)
		}
	case "max":
		switch err.Kind().String() {
		case kindString:
			return fmt.Sprintf("%s must be at most %s characters long", field, param)
		case kindSlice, kindArray, kindMap:
			return fmt.Sprintf("%s must contain at most %s items", field, param)
		default:
			return fmt.Sprintf("%s must be at most %s", field, param)
		}
	case "len":
		switch err.Kind().String() {
		case kindString:
			return fmt.Sprintf("%s must be exactly %s characters long", field, param)
		case kindSlice, kindArray, kindMap:
			return fmt.Sprintf("%s must contain exactly %s items", field, param)
		default:
			return fmt.Sprintf("%s must be exactly %s", field, param)
		}
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", field, param)
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a valid numeric value", field)
	case "number":
		return fmt.Sprintf("%s must be a valid number", field)
	case "hexadecimal":
		return fmt.Sprintf("%s must be a valid hexadecimal string", field)
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color code", field)
	case "rgb":
		return fmt.Sprintf("%s must be a valid RGB color", field)
	case "rgba":
		return fmt.Sprintf("%s must be a valid RGBA color", field)
	case "hsl":
		return fmt.Sprintf("%s must be a valid HSL color", field)
	case "hsla":
		return fmt.Sprintf("%s must be a valid HSLA color", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "uuid3":
		return fmt.Sprintf("%s must be a valid UUID v3", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", field)
	case "uuid5":
		return fmt.Sprintf("%s must be a valid UUID v5", field)
	case "isbn":
		return fmt.Sprintf("%s must be a valid ISBN", field)
	case "isbn10":
		return fmt.Sprintf("%s must be a valid ISBN-10", field)
	case "isbn13":
		return fmt.Sprintf("%s must be a valid ISBN-13", field)
	case "json":
		return fmt.Sprintf("%s must be valid JSON", field)
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude", field)
	case "ssn":
		return fmt.Sprintf("%s must be a valid Social Security Number", field)
	case "ipv4":
		return fmt.Sprintf("%s must be a valid IPv4 address", field)
	case "ipv6":
		return fmt.Sprintf("%s must be a valid IPv6 address", field)
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", field)
	case "mac":
		return fmt.Sprintf("%s must be a valid MAC address", field)
	case "contains":
		return fmt.Sprintf("%s must contain '%s'", field, param)
	case "containsany":
		return fmt.Sprintf("%s must contain at least one of: %s", field, param)
	case "excludes":
		return fmt.Sprintf("%s must not contain '%s'", field, param)
	case "excludesall":
		return fmt.Sprintf("%s must not contain any of: %s", field, param)
	case "startswith":
		return fmt.Sprintf("%s must start with '%s'", field, param)
	case "endswith":
		return fmt.Sprintf("%s must end with '%s'", field, param)
	case "datetime":
		return fmt.Sprintf("%s must be a valid date/time in format: %s", field, param)
	default:
		// Generic message for unknown tags.
		if param != "" {
			return fmt.Sprintf("%s failed '%s' validation (param: %s)", field, err.Tag(), param)
		}
		return fmt.Sprintf("%s failed '%s' validation", field, err.Tag())
	}
}
