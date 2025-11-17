package main

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// validateStrongPassword checks if password meets strength requirements.
// Requirements: min 8 chars, uppercase, lowercase, digit, special char.
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasDigit   = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]`).MatchString(password)
	)

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// validatePhone checks if phone number is in valid format.
// Accepts: +1234567890, (123) 456-7890, 123-456-7890.
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// Remove all non-digit characters for validation.
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Check if we have 10-15 digits (international format).
	if len(digits) < 10 || len(digits) > 15 {
		return false
	}

	// Phone number should contain only allowed characters.
	validPattern := regexp.MustCompile(`^[\d\s()+\-]+$`)
	return validPattern.MatchString(phone)
}

// validateCompanyDomain checks if email ends with company domain.
func validateCompanyDomain(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	// List of allowed company domains.
	allowedDomains := []string{
		"@example.com",
		"@company.org",
		"@enterprise.net",
	}

	for _, domain := range allowedDomains {
		if strings.HasSuffix(strings.ToLower(email), domain) {
			return true
		}
	}

	return false
}
