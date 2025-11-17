// Package main provides custom middleware implementations.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/coregx/fursy"
)

// =============================================
// RequestID Middleware
// =============================================

// RequestIDMiddleware adds a unique request ID to each request.
// The request ID is stored in the context and added to the response header.
//
// Usage:
//
//	router.Use(RequestIDMiddleware())
//
// Access in handlers:
//
//	requestID := c.GetString("request_id")
func RequestIDMiddleware() fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		// Check if request already has an ID (from client or proxy)
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = generateRequestID()
		}

		// Store in context for handlers
		c.Set("request_id", requestID)

		// Add to response headers
		c.SetHeader("X-Request-ID", requestID)

		return c.Next()
	}
}

// generateRequestID generates a random request ID.
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if random fails
		return time.Now().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}

// =============================================
// Timing Middleware
// =============================================

// TimingMiddleware measures request processing time.
// The duration is added to the response header and stored in context.
//
// Usage:
//
//	router.Use(TimingMiddleware())
//
// Access in handlers:
//
//	duration := c.Get("response_time").(time.Duration)
func TimingMiddleware() fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		// Record start time
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Store in context
		c.Set("response_time", duration)

		// Add to response header (in milliseconds)
		durationMS := float64(duration.Nanoseconds()) / 1e6
		c.SetHeader("X-Response-Time", formatDuration(durationMS))

		return err
	}
}

// formatDuration formats duration in a human-readable format.
func formatDuration(ms float64) string {
	if ms < 1.0 {
		return "< 1ms"
	}
	if ms < 1000.0 {
		return formatFloat(ms) + "ms"
	}
	return formatFloat(ms/1000.0) + "s"
}

// formatFloat formats float with appropriate precision.
func formatFloat(f float64) string {
	if f < 10 {
		return formatFloatPrec(f, 2)
	}
	if f < 100 {
		return formatFloatPrec(f, 1)
	}
	return formatFloatPrec(f, 0)
}

// formatFloatPrec formats float with specified precision.
func formatFloatPrec(f float64, prec int) string {
	switch prec {
	case 0:
		return formatInt(int(f + 0.5))
	case 1:
		return formatInt(int(f*10+0.5)) + "." + formatInt(int(f*10+0.5)%10)
	case 2:
		whole := int(f)
		frac := int((f-float64(whole))*100 + 0.5)
		return formatInt(whole) + "." + formatTwoDigits(frac)
	default:
		return formatInt(int(f + 0.5))
	}
}

// formatInt converts int to string without using fmt.Sprintf.
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + formatInt(-n)
	}

	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// formatTwoDigits formats int as two digits (00-99).
func formatTwoDigits(n int) string {
	if n < 10 {
		return "0" + formatInt(n)
	}
	return formatInt(n)
}

// =============================================
// Conditional Middleware (Skipper Pattern)
// =============================================

// SkipperMiddleware is a middleware that can be conditionally skipped.
// This demonstrates the Skipper pattern used by built-in middleware.
//
// Usage:
//
//	// Skip middleware for health checks
//	router.Use(SkipperMiddleware(func(c *fursy.Context) bool {
//	    return c.Request.URL.Path == "/health"
//	}))
func SkipperMiddleware(skipper func(*fursy.Context) bool) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		// Skip middleware if skipper returns true
		if skipper != nil && skipper(c) {
			return c.Next()
		}

		// Middleware logic here
		// Example: Add custom header
		c.SetHeader("X-Custom-Middleware", "executed")

		return c.Next()
	}
}

// =============================================
// Authentication Middleware (Custom)
// =============================================

// APIKeyMiddleware validates an API key from the X-API-Key header.
// This demonstrates a simple custom authentication middleware.
//
// Usage:
//
//	validKeys := map[string]bool{
//	    "key123": true,
//	    "key456": true,
//	}
//	router.Use(APIKeyMiddleware(validKeys))
func APIKeyMiddleware(validKeys map[string]bool) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		// Extract API key from header
		apiKey := c.Request.Header.Get("X-API-Key")

		// Validate API key
		if apiKey == "" {
			return c.Problem(fursy.Problem{
				Type:   "https://example.com/errors/missing-api-key",
				Title:  "Missing API Key",
				Status: 401,
				Detail: "X-API-Key header is required",
			})
		}

		if !validKeys[apiKey] {
			return c.Problem(fursy.Problem{
				Type:   "https://example.com/errors/invalid-api-key",
				Title:  "Invalid API Key",
				Status: 401,
				Detail: "The provided API key is not valid",
			})
		}

		// Store API key in context
		c.Set("api_key", apiKey)

		return c.Next()
	}
}

// =============================================
// Cache Control Middleware
// =============================================

// CacheControlMiddleware adds cache control headers to responses.
// Useful for controlling browser and CDN caching behavior.
//
// Usage:
//
//	// No cache for API endpoints
//	router.Use(CacheControlMiddleware("no-cache, no-store, must-revalidate"))
//
//	// Cache for 1 hour
//	router.Use(CacheControlMiddleware("public, max-age=3600"))
func CacheControlMiddleware(cacheControl string) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		// Set Cache-Control header
		c.SetHeader("Cache-Control", cacheControl)

		// Add Pragma for HTTP/1.0 compatibility
		if cacheControl == "no-cache, no-store, must-revalidate" {
			c.SetHeader("Pragma", "no-cache")
			c.SetHeader("Expires", "0")
		}

		return c.Next()
	}
}

// =============================================
// Compression Hint Middleware
// =============================================

// CompressionHintMiddleware adds Vary: Accept-Encoding header.
// This hints to proxies and CDNs that responses vary based on compression support.
//
// Usage:
//
//	router.Use(CompressionHintMiddleware())
func CompressionHintMiddleware() fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		// Add Vary header for compression
		c.SetHeader("Vary", "Accept-Encoding")
		return c.Next()
	}
}

// =============================================
// Version Middleware
// =============================================

// VersionMiddleware adds API version information to response headers.
// Useful for API versioning and debugging.
//
// Usage:
//
//	router.Use(VersionMiddleware("v1.2.3"))
func VersionMiddleware(version string) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		c.SetHeader("X-API-Version", version)
		return c.Next()
	}
}

// =============================================
// Method Override Middleware
// =============================================

// MethodOverrideMiddleware allows HTTP method override via header or query parameter.
// Useful for clients that can't send PUT/DELETE requests (e.g., HTML forms).
//
// Checks in order:
//  1. X-HTTP-Method-Override header
//  2. _method query parameter
//
// Usage:
//
//	router.Use(MethodOverrideMiddleware())
func MethodOverrideMiddleware() fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		// Only override POST requests
		if c.Request.Method != "POST" {
			return c.Next()
		}

		// Check header first
		method := c.Request.Header.Get("X-HTTP-Method-Override")
		if method == "" {
			// Check query parameter
			method = c.Query("_method")
		}

		// Override method if valid
		if method != "" {
			method = normalizeMethod(method)
			if isValidMethod(method) {
				c.Request.Method = method
			}
		}

		return c.Next()
	}
}

// normalizeMethod normalizes HTTP method to uppercase.
func normalizeMethod(method string) string {
	// Simple uppercase conversion
	result := make([]byte, len(method))
	for i := 0; i < len(method); i++ {
		b := method[i]
		if b >= 'a' && b <= 'z' {
			b -= 32 // Convert to uppercase
		}
		result[i] = b
	}
	return string(result)
}

// isValidMethod checks if method is a valid HTTP method.
func isValidMethod(method string) bool {
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}
	return validMethods[method]
}
