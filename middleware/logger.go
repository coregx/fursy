// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware provides common HTTP middleware for the FURSY router.
package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coregx/fursy"
)

// LoggerConfig defines the configuration for the Logger middleware.
type LoggerConfig struct {
	// Logger is the slog.Logger instance to use for logging.
	// If nil, a default logger writing to os.Stdout will be created.
	Logger *slog.Logger

	// SkipPaths is a list of URL paths to skip logging.
	// Useful for health checks, metrics endpoints, etc.
	SkipPaths []string

	// SkipFunc is a custom function to determine if a request should be skipped.
	// If both SkipPaths and SkipFunc are provided, a request is skipped if either matches.
	SkipFunc func(*http.Request) bool
}

// Logger returns a middleware that logs HTTP requests using structured logging (slog).
//
// The middleware logs:
//   - Request method and path
//   - Response status code
//   - Request latency in milliseconds
//   - Client IP address
//   - Bytes written
//   - Error (if any)
//
// Example:
//
//	router := fursy.New()
//	router.Use(middleware.Logger())
//
// Default output format (JSON):
//
//	{"time":"2025-01-13T10:00:00Z","level":"INFO","msg":"HTTP request",
//	 "method":"GET","path":"/api/users","status":200,"latency_ms":12.5,
//	 "ip":"192.168.1.1","bytes":1234}
func Logger() fursy.HandlerFunc {
	return LoggerWithConfig(LoggerConfig{})
}

// LoggerWithConfig returns a middleware with custom configuration.
//
// Example:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	config := middleware.LoggerConfig{
//	    Logger: logger,
//	    SkipPaths: []string{"/health", "/metrics"},
//	}
//	router.Use(middleware.LoggerWithConfig(config))
func LoggerWithConfig(config LoggerConfig) fursy.HandlerFunc {
	// Use provided logger or create default
	logger := config.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	// Pre-compile skip paths map for O(1) lookup
	skipPaths := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *fursy.Context) error {
		// Check if request should be skipped
		if skipPaths[c.Request.URL.Path] {
			return c.Next()
		}

		if config.SkipFunc != nil && config.SkipFunc(c.Request) {
			return c.Next()
		}

		// Record start time
		start := time.Now()

		// Wrap response writer to capture status and bytes
		lrw := &logResponseWriter{
			ResponseWriter: c.Response,
			statusCode:     http.StatusOK, // Default status
		}
		c.Response = lrw

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)
		latencyMS := float64(latency.Nanoseconds()) / 1e6

		// Get client IP
		clientIP := getClientIP(c.Request)

		// Build log attributes
		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", lrw.statusCode),
			slog.Float64("latency_ms", latencyMS),
			slog.String("ip", clientIP),
			slog.Int64("bytes", lrw.bytesWritten),
		}

		// Add error if present
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		// Determine log level based on status code and error
		level := slog.LevelInfo
		if err != nil || lrw.statusCode >= 500 {
			level = slog.LevelError
		} else if lrw.statusCode >= 400 {
			level = slog.LevelWarn
		}

		// Log the request
		logger.LogAttrs(c.Request.Context(), level, "HTTP request", attrs...)

		return err
	}
}

// logResponseWriter wraps http.ResponseWriter to capture status code and bytes written.
type logResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	wroteHeader  bool
}

// WriteHeader captures the status code and calls the underlying WriteHeader.
func (w *logResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

// Write captures bytes written and calls the underlying Write.
func (w *logResponseWriter) Write(b []byte) (int, error) {
	// WriteHeader is called implicitly with 200 if not called explicitly
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// Unwrap returns the underlying ResponseWriter.
// This is useful for middleware that need to access the original ResponseWriter.
func (w *logResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// getClientIP extracts the client IP address from the request.
// It checks X-Real-IP, X-Forwarded-For headers, and falls back to RemoteAddr.
func getClientIP(r *http.Request) string {
	// Check X-Real-IP header (used by some proxies)
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return cleanIP(ip)
	}

	// Check X-Forwarded-For header (used by most proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// We want the first one (the original client)
		if idx := strings.Index(xff, ","); idx != -1 {
			return cleanIP(xff[:idx])
		}
		return cleanIP(xff)
	}

	// Fall back to RemoteAddr
	return cleanIP(r.RemoteAddr)
}

// cleanIP removes port from IP address if present.
func cleanIP(ip string) string {
	ip = strings.TrimSpace(ip)

	// Find last colon (potential port separator)
	idx := strings.LastIndex(ip, ":")
	if idx == -1 {
		return ip // No colon, return as-is
	}

	// Check if IPv6 (multiple colons)
	if strings.Count(ip, ":") > 1 {
		return cleanIPv6(ip)
	}

	// IPv4 with port - remove port
	return ip[:idx]
}

// cleanIPv6 handles IPv6 address cleaning.
func cleanIPv6(ip string) string {
	// IPv6 with port uses brackets: [2001:db8::1]:8080
	if !strings.HasPrefix(ip, "[") {
		return ip // Pure IPv6 without port
	}

	// Extract IP from brackets
	if idx := strings.Index(ip, "]"); idx != -1 {
		return ip[1:idx]
	}

	return ip
}

// DefaultLogger creates a logger that writes to the given writer.
// This is a convenience function for creating custom loggers.
//
// Example:
//
//	var buf bytes.Buffer
//	logger := middleware.DefaultLogger(&buf)
//	router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
//	    Logger: logger,
//	}))
func DefaultLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// JSONLogger creates a JSON logger that writes to the given writer.
// Useful for production environments where structured logs are parsed.
//
// Example:
//
//	logger := middleware.JSONLogger(os.Stdout)
//	router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
//	    Logger: logger,
//	}))
func JSONLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
