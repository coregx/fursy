// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware provides HTTP Basic Authentication middleware.
package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/coregx/fursy"
)

// UserContextKey is the key used to store authenticated user identity in the context.
const UserContextKey = "User"

// DefaultRealm is the default realm name for HTTP Basic Authentication.
var DefaultRealm = "Restricted"

// ValidatorFunc is the function that validates username and password.
// It should return the user identity (any type) on success, or an error on failure.
type ValidatorFunc func(c *fursy.Context, username, password string) (interface{}, error)

// BasicAuthConfig defines the configuration for the BasicAuth middleware.
type BasicAuthConfig struct {
	// Validator is the function that validates username and password.
	// Required.
	Validator ValidatorFunc

	// Realm is the authentication realm shown in the browser's login prompt.
	// Default: "Restricted"
	Realm string

	// Skipper defines a function to skip the middleware.
	// Default: nil (middleware always executes)
	Skipper func(c *fursy.Context) bool
}

// BasicAuth returns a middleware that provides HTTP Basic Authentication.
//
// The middleware:
//   - Parses the Authorization header (Basic scheme)
//   - Decodes base64 credentials
//   - Validates username:password using the validator function
//   - Stores user identity in context on success
//   - Returns 401 Unauthorized on failure with WWW-Authenticate header
//
// Example:
//
//	validator := func(c *fursy.Context, username, password string) (interface{}, error) {
//	    if username == "admin" && password == "secret" {
//	        return username, nil // Return user identity
//	    }
//	    return nil, errors.New("invalid credentials")
//	}
//
//	router := fursy.New()
//	router.Use(middleware.BasicAuth(validator))
//
// Access user identity in handlers:
//
//	router.GET("/protected", func(c *fursy.Context) error {
//	    user := c.GetString(middleware.UserContextKey)
//	    return c.String(200, "Hello, "+user)
//	})
func BasicAuth(validator ValidatorFunc) fursy.HandlerFunc {
	return BasicAuthWithConfig(BasicAuthConfig{
		Validator: validator,
	})
}

// BasicAuthWithConfig returns a middleware with custom configuration.
//
// Example:
//
//	config := middleware.BasicAuthConfig{
//	    Validator: validateUser,
//	    Realm: "Admin Area",
//	    Skipper: func(c *fursy.Context) bool {
//	        // Skip auth for health check endpoint
//	        return c.Request.URL.Path == "/health"
//	    },
//	}
//	router.Use(middleware.BasicAuthWithConfig(config))
func BasicAuthWithConfig(config BasicAuthConfig) fursy.HandlerFunc {
	// Validate config.
	if config.Validator == nil {
		panic("fursy/middleware: BasicAuth validator cannot be nil")
	}

	// Set defaults.
	if config.Realm == "" {
		config.Realm = DefaultRealm
	}

	return func(c *fursy.Context) error {
		// Skip if Skipper returns true.
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		// Parse Authorization header.
		auth := c.Request.Header.Get("Authorization")
		username, password := parseBasicAuth(auth)

		// Validate credentials.
		if username != "" || password != "" {
			identity, err := config.Validator(c, username, password)
			if err == nil && identity != nil {
				// Store user identity in context.
				c.Set(UserContextKey, identity)
				return c.Next()
			}
		}

		// Authentication failed - send WWW-Authenticate header.
		c.SetHeader("WWW-Authenticate", `Basic realm="`+config.Realm+`"`)
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
}

// parseBasicAuth parses the Authorization header and extracts username and password.
// Returns empty strings if parsing fails.
func parseBasicAuth(auth string) (username, password string) {
	// Check for "Basic " prefix.
	if !strings.HasPrefix(auth, "Basic ") {
		return "", ""
	}

	// Decode base64 credentials.
	encoded := auth[6:] // Skip "Basic "
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ""
	}

	// Split username:password.
	credentials := string(decoded)
	colonIndex := strings.IndexByte(credentials, ':')
	if colonIndex < 0 {
		return "", ""
	}

	return credentials[:colonIndex], credentials[colonIndex+1:]
}

// BasicAuthAccounts creates a validator that checks credentials against a map.
// This is a convenience function for simple username:password authentication.
//
// Example:
//
//	accounts := map[string]string{
//	    "admin": "secret",
//	    "user":  "pass123",
//	}
//	router.Use(middleware.BasicAuth(middleware.BasicAuthAccounts(accounts)))
func BasicAuthAccounts(accounts map[string]string) ValidatorFunc {
	return func(_ *fursy.Context, username, password string) (interface{}, error) {
		if expectedPassword, ok := accounts[username]; ok && expectedPassword == password {
			return username, nil
		}
		return nil, fursy.ErrUnauthorized
	}
}
