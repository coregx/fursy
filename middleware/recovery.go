// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware provides panic recovery middleware.
package middleware

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"github.com/coregx/fursy"
)

// RecoveryConfig defines the configuration for the Recovery middleware.
type RecoveryConfig struct {
	// Logger is the slog.Logger instance to use for logging panics.
	// If nil, a default logger writing to os.Stderr will be created.
	Logger *slog.Logger

	// DisableStackTrace disables printing stack trace in logs.
	// Default: false (stack trace is printed).
	DisableStackTrace bool

	// DisablePrintStack disables printing stack to stderr.
	// Default: false (stack is printed to stderr).
	DisablePrintStack bool

	// StackTraceSize is the maximum size of the stack trace buffer in bytes.
	// Default: 4KB (4096 bytes).
	StackTraceSize int
}

// Recovery returns a middleware that recovers from panics in request handlers.
//
// When a panic occurs:
//   - The panic is recovered and converted to an error
//   - Stack trace is logged using structured logging (slog)
//   - HTTP 500 Internal Server Error is sent to the client
//   - Request processing continues normally after recovery
//
// Example:
//
//	router := fursy.New()
//	router.Use(middleware.Recovery())
//
// With custom logger:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
//	router.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
//	    Logger: logger,
//	}))
func Recovery() fursy.HandlerFunc {
	return RecoveryWithConfig(RecoveryConfig{})
}

// RecoveryWithConfig returns a middleware with custom configuration.
//
// Example:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
//	config := middleware.RecoveryConfig{
//	    Logger: logger,
//	    DisableStackTrace: false,
//	    StackTraceSize: 8192, // 8KB stack trace
//	}
//	router.Use(middleware.RecoveryWithConfig(config))
func RecoveryWithConfig(config RecoveryConfig) fursy.HandlerFunc {
	// Use provided logger or create default.
	logger := config.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelError,
		}))
	}

	// Set default stack trace size.
	stackTraceSize := config.StackTraceSize
	if stackTraceSize == 0 {
		stackTraceSize = 4096 // 4KB default
	}

	return func(c *fursy.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = handlePanic(r, c, logger, config, stackTraceSize)
			}
		}()

		return c.Next()
	}
}

// handlePanic handles a recovered panic by logging and sending error response.
func handlePanic(r interface{}, c *fursy.Context, logger *slog.Logger, config RecoveryConfig, stackTraceSize int) error {
	// Get stack trace.
	stack := getStackTrace(config.DisableStackTrace, stackTraceSize)

	// Convert panic to error.
	panicErr := convertPanicToError(r)

	// Log panic.
	logPanic(c, logger, panicErr, stack, config.DisableStackTrace)

	// Print stack to stderr for visibility.
	printStackToStderr(panicErr, stack, config)

	// Send 500 response.
	return c.String(http.StatusInternalServerError, "Internal Server Error")
}

// getStackTrace gets the current stack trace if not disabled.
func getStackTrace(disableStackTrace bool, stackTraceSize int) []byte {
	if disableStackTrace {
		return nil
	}

	stack := make([]byte, stackTraceSize)
	return stack[:runtime.Stack(stack, false)]
}

// convertPanicToError converts a panic value to an error.
func convertPanicToError(r interface{}) error {
	if e, ok := r.(error); ok {
		return e
	}
	return fmt.Errorf("%v", r)
}

// logPanic logs the panic with structured fields.
func logPanic(c *fursy.Context, logger *slog.Logger, panicErr error, stack []byte, disableStackTrace bool) {
	attrs := []slog.Attr{
		slog.String("panic", panicErr.Error()),
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
		slog.String("remote_addr", c.Request.RemoteAddr),
	}

	if !disableStackTrace && len(stack) > 0 {
		attrs = append(attrs, slog.String("stack", string(stack)))
	}

	logger.LogAttrs(c.Request.Context(), slog.LevelError, "Panic recovered", attrs...)
}

// printStackToStderr prints stack trace to stderr if enabled.
func printStackToStderr(panicErr error, stack []byte, config RecoveryConfig) {
	if config.DisablePrintStack || config.DisableStackTrace || len(stack) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "PANIC: %v\n%s\n", panicErr, stack)
}

// PanicHandler is a simplified version of Recovery that only recovers panics
// without any logging or configuration.
//
// Useful for testing or when you want minimal overhead.
//
// Example:
//
//	router := fursy.New()
//	router.Use(middleware.PanicHandler())
func PanicHandler() fursy.HandlerFunc {
	return func(c *fursy.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				// Convert panic to error.
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("%v", r)
				}

				// Send 500 response.
				_ = c.String(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		return c.Next()
	}
}

// DefaultRecoveryLogger creates a recovery logger that writes to the given writer.
// This is a convenience function for creating custom loggers.
//
// Example:
//
//	var buf bytes.Buffer
//	logger := middleware.DefaultRecoveryLogger(&buf)
//	router.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
//	    Logger: logger,
//	}))
func DefaultRecoveryLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}

// JSONRecoveryLogger creates a JSON recovery logger that writes to the given writer.
// Useful for production environments where structured logs are parsed.
//
// Example:
//
//	logger := middleware.JSONRecoveryLogger(os.Stderr)
//	router.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
//	    Logger: logger,
//	}))
func JSONRecoveryLogger(w io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
}
