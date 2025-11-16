// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware provides CORS support.
package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coregx/fursy"
)

// CORS header constants.
const (
	headerOrigin = "Origin"

	headerRequestMethod  = "Access-Control-Request-Method"
	headerRequestHeaders = "Access-Control-Request-Headers"

	headerAllowOrigin      = "Access-Control-Allow-Origin"
	headerAllowCredentials = "Access-Control-Allow-Credentials"
	headerAllowHeaders     = "Access-Control-Allow-Headers"
	headerAllowMethods     = "Access-Control-Allow-Methods"
	headerExposeHeaders    = "Access-Control-Expose-Headers"
	headerMaxAge           = "Access-Control-Max-Age"
)

// CORSConfig defines the configuration for the CORS middleware.
type CORSConfig struct {
	// AllowOrigins is a comma-separated list of origins that may access the resource.
	// Use "*" to allow any origin, or "null" to disallow all origins.
	// Default: "*"
	AllowOrigins string

	// AllowMethods is a comma-separated list of HTTP methods allowed when accessing the resource.
	// Use "*" to allow any method.
	// Default: "GET,HEAD,PUT,POST,DELETE,PATCH"
	AllowMethods string

	// AllowHeaders is a comma-separated list of request headers allowed when accessing the resource.
	// Use "*" to allow any header.
	// Default: ""
	AllowHeaders string

	// ExposeHeaders is a comma-separated list of response headers that browsers are allowed to access.
	// Default: ""
	ExposeHeaders string

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication, or client-side SSL certificates.
	// Default: false
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached.
	// Default: 0 (no caching)
	MaxAge time.Duration

	// Internal maps for efficient lookup.
	allowOriginMap map[string]bool
	allowMethodMap map[string]bool
	allowHeaderMap map[string]bool
}

// AllowAll is a predefined CORS config that allows all origins, methods, and headers.
var AllowAll = CORSConfig{
	AllowOrigins: "*",
	AllowMethods: "*",
	AllowHeaders: "*",
}

// CORS returns a middleware that handles Cross-Origin Resource Sharing (CORS).
//
// The middleware handles:
//   - Preflight OPTIONS requests with appropriate CORS headers
//   - Actual requests by adding CORS headers to responses
//   - Origin validation based on configuration
//
// Example:
//
//	router := fursy.New()
//	router.Use(middleware.CORS())
//
// With custom config:
//
//	config := middleware.CORSConfig{
//	    AllowOrigins: "https://example.com,https://foo.com",
//	    AllowMethods: "GET,POST,PUT",
//	    AllowHeaders: "Content-Type,Authorization",
//	    AllowCredentials: true,
//	    MaxAge: 12 * time.Hour,
//	}
//	router.Use(middleware.CORSWithConfig(config))
func CORS() fursy.HandlerFunc {
	return CORSWithConfig(CORSConfig{})
}

// CORSWithConfig returns a middleware with custom CORS configuration.
//
// Example:
//
//	config := middleware.CORSConfig{
//	    AllowOrigins: "https://example.com",
//	    AllowMethods: "GET,POST,PUT,DELETE",
//	    AllowHeaders: "Content-Type,Authorization",
//	    ExposeHeaders: "X-Request-ID",
//	    AllowCredentials: true,
//	    MaxAge: 12 * time.Hour,
//	}
//	router.Use(middleware.CORSWithConfig(config))
func CORSWithConfig(config CORSConfig) fursy.HandlerFunc {
	// Set defaults.
	if config.AllowOrigins == "" {
		config.AllowOrigins = "*"
	}
	if config.AllowMethods == "" {
		config.AllowMethods = "GET,HEAD,PUT,POST,DELETE,PATCH"
	}

	// Initialize lookup maps.
	config.init()

	return func(c *fursy.Context) error {
		origin := c.Request.Header.Get(headerOrigin)
		if origin == "" {
			// Not a CORS request.
			return c.Next()
		}

		// Check if this is a preflight request.
		if c.Request.Method == http.MethodOptions {
			method := c.Request.Header.Get(headerRequestMethod)
			if method == "" {
				// Not a preflight request - no Access-Control-Request-Method.
				return c.Next()
			}

			// Handle preflight request.
			headers := c.Request.Header.Get(headerRequestHeaders)
			config.setPreflightHeaders(origin, method, headers, c.Response.Header())

			// Write 204 response and stop processing.
			c.Abort()
			return c.NoContent(http.StatusNoContent)
		}

		// Handle actual request.
		config.setActualHeaders(origin, c.Response.Header())
		return c.Next()
	}
}

// init initializes the internal lookup maps for efficient origin/method/header checking.
func (cfg *CORSConfig) init() {
	cfg.allowOriginMap = buildAllowMap(cfg.AllowOrigins, true)
	cfg.allowMethodMap = buildAllowMap(cfg.AllowMethods, true)
	cfg.allowHeaderMap = buildAllowMap(cfg.AllowHeaders, false)
}

// isOriginAllowed checks if the given origin is allowed.
func (cfg *CORSConfig) isOriginAllowed(origin string) bool {
	if cfg.AllowOrigins == "null" {
		return false
	}
	return cfg.AllowOrigins == "*" || cfg.allowOriginMap[origin]
}

// setActualHeaders sets CORS headers for actual requests.
func (cfg *CORSConfig) setActualHeaders(origin string, headers http.Header) {
	if !cfg.isOriginAllowed(origin) {
		return
	}

	cfg.setOriginHeader(origin, headers)

	if cfg.ExposeHeaders != "" {
		headers.Set(headerExposeHeaders, cfg.ExposeHeaders)
	}
}

// setPreflightHeaders sets CORS headers for preflight OPTIONS requests.
func (cfg *CORSConfig) setPreflightHeaders(origin, method, reqHeaders string, headers http.Header) {
	allowed, allowedHeaders := cfg.isPreflightAllowed(origin, method, reqHeaders)
	if !allowed {
		return
	}

	cfg.setOriginHeader(origin, headers)

	if cfg.MaxAge > 0 {
		headers.Set(headerMaxAge, strconv.FormatInt(int64(cfg.MaxAge/time.Second), 10))
	}

	if cfg.AllowMethods == "*" {
		headers.Set(headerAllowMethods, method)
	} else if cfg.allowMethodMap[method] {
		headers.Set(headerAllowMethods, cfg.AllowMethods)
	}

	if allowedHeaders != "" {
		headers.Set(headerAllowHeaders, reqHeaders)
	}
}

// isPreflightAllowed checks if the preflight request is allowed.
// Returns (allowed, allowedHeaders).
func (cfg *CORSConfig) isPreflightAllowed(origin, method, reqHeaders string) (allowed bool, allowedHeaders string) {
	if !cfg.isOriginAllowed(origin) {
		return false, ""
	}

	if cfg.AllowMethods != "*" && !cfg.allowMethodMap[method] {
		return false, ""
	}

	if cfg.AllowHeaders == "*" || reqHeaders == "" {
		return true, reqHeaders
	}

	// Check each requested header.
	headers := []string{}
	for _, header := range strings.Split(reqHeaders, ",") {
		header = strings.TrimSpace(header)
		if cfg.allowHeaderMap[strings.ToUpper(header)] {
			headers = append(headers, header)
		}
	}

	if len(headers) > 0 {
		return true, strings.Join(headers, ",")
	}

	return false, ""
}

// setOriginHeader sets the Access-Control-Allow-Origin header.
func (cfg *CORSConfig) setOriginHeader(origin string, headers http.Header) {
	if cfg.AllowCredentials {
		// When credentials are allowed, must use specific origin (not "*").
		headers.Set(headerAllowOrigin, origin)
		headers.Set(headerAllowCredentials, "true")
	} else {
		if cfg.AllowOrigins == "*" {
			headers.Set(headerAllowOrigin, "*")
		} else {
			headers.Set(headerAllowOrigin, origin)
		}
	}
}

// buildAllowMap builds a lookup map from a comma-separated string.
// If caseSensitive is false, keys are converted to uppercase.
func buildAllowMap(s string, caseSensitive bool) map[string]bool {
	m := make(map[string]bool)
	if s == "" {
		return m
	}

	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if caseSensitive {
			m[p] = true
		} else {
			m[strings.ToUpper(p)] = true
		}
	}

	return m
}
