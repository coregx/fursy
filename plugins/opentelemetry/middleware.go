// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package opentelemetry

import (
	"fmt"
	"net/http"

	"github.com/coregx/fursy"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// ScopeName is the instrumentation scope name.
	ScopeName = "github.com/coregx/fursy/plugins/opentelemetry"

	// Version is the instrumentation version.
	Version = "0.1.0"
)

// Config holds the OpenTelemetry middleware configuration.
type Config struct {
	// TracerProvider provides the tracer for creating spans.
	// If not set, the global TracerProvider is used.
	TracerProvider trace.TracerProvider

	// Propagators is a set of propagators to use for extracting
	// and injecting context. If not set, the global TextMapPropagator is used.
	Propagators propagation.TextMapPropagator

	// Skipper defines a function to skip middleware for certain requests.
	// If it returns true, the request will not be traced.
	// Example: skip /health and /metrics endpoints.
	Skipper func(*fursy.Context) bool

	// SpanNameFormatter allows customizing the span name.
	// Default format is "{method} {route}" (e.g., "GET /users/:id").
	// If route is not available, falls back to "{method} {path}".
	SpanNameFormatter func(c *fursy.Context) string

	// ServerName is the logical server name to use in span attributes.
	// Maps to server.address semantic convention.
	ServerName string

	// WithClientIP determines if client.address should be included.
	// Default: true.
	WithClientIP bool

	// WithUserAgent determines if user_agent.original should be included.
	// Default: true.
	WithUserAgent bool

	// WithRequestHeaders lists request headers to capture as span attributes.
	// Format: http.request.header.<name>
	// Default: empty (no headers captured).
	WithRequestHeaders []string

	// WithResponseHeaders lists response headers to capture as span attributes.
	// Format: http.response.header.<name>
	// Default: empty (no headers captured).
	WithResponseHeaders []string
}

// Middleware returns a FURSY middleware that traces HTTP requests using OpenTelemetry.
//
// The middleware creates a span for each request following the OpenTelemetry
// HTTP semantic conventions. It automatically:
//   - Extracts trace context from incoming requests (W3C Trace Box)
//   - Creates a new span for the request
//   - Records HTTP attributes (method, status code, route, etc.)
//   - Propagates trace context to downstream services
//   - Records errors and panics
//
// Example:
//
//	router := fursy.New()
//	router.Use(opentelemetry.Middleware("my-service"))
//
//	// With custom configuration:
//	router.Use(opentelemetry.MiddlewareWithConfig(opentelemetry.Config{
//	    Skipper: func(c *fursy.Context) bool {
//	        return c.Request.URL.Path == "/health"
//	    },
//	    WithClientIP: true,
//	    WithUserAgent: true,
//	}))
func Middleware(serverName string) fursy.HandlerFunc {
	return MiddlewareWithConfig(Config{
		ServerName:    serverName,
		WithClientIP:  true,
		WithUserAgent: true,
	})
}

// MiddlewareWithConfig returns a FURSY middleware with custom configuration.
func MiddlewareWithConfig(config Config) fursy.HandlerFunc {
	// Set defaults.
	if config.TracerProvider == nil {
		config.TracerProvider = otel.GetTracerProvider()
	}

	if config.Propagators == nil {
		config.Propagators = otel.GetTextMapPropagator()
	}

	if config.SpanNameFormatter == nil {
		config.SpanNameFormatter = defaultSpanNameFormatter
	}

	tracer := config.TracerProvider.Tracer(
		ScopeName,
		trace.WithInstrumentationVersion(Version),
	)

	return func(c *fursy.Context) error {
		// Skip if Skipper returns true.
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		// Extract trace context from incoming request headers.
		ctx := config.Propagators.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Generate span name.
		spanName := config.SpanNameFormatter(c)

		// Start a new span.
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(httpServerAttributes(c, config)...),
		)
		defer span.End()

		// Update request context with trace context.
		c.Request = c.Request.WithContext(ctx)

		// Wrap ResponseWriter to capture status code.
		wrapper := &responseWriter{
			ResponseWriter: c.Response,
			statusCode:     http.StatusOK, // Default to 200 OK.
		}
		c.Response = wrapper

		// Execute the handler chain.
		err := c.Next()

		// Record error if present.
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		// Set HTTP response attributes.
		status := wrapper.statusCode
		span.SetAttributes(semconv.HTTPResponseStatusCode(status))

		// Set span status based on HTTP status code.
		if status >= 400 {
			span.SetStatus(codes.Error, http.StatusText(status))
		}

		// Capture response headers if configured.
		for _, header := range config.WithResponseHeaders {
			if value := c.Response.Header().Get(header); value != "" {
				attrKey := fmt.Sprintf("http.response.header.%s", header)
				span.SetAttributes(attribute.String(attrKey, value))
			}
		}

		return err
	}
}

// defaultSpanNameFormatter returns the default span name format: "{method} {route}".
//
// If route pattern is available (e.g., "/users/:id"), use it.
// Otherwise, fall back to the actual path.
func defaultSpanNameFormatter(c *fursy.Context) string {
	method := c.Request.Method

	// Try to get route pattern from context.
	// In FURSY, the route pattern isn't currently stored, so we use the path.
	// Future improvement: store matched route pattern in context.
	route := c.Request.URL.Path

	// For known paths, use them directly.
	// This creates better span names like "GET /users/:id" vs "GET /users/123".
	// TODO: Store route pattern in context during routing.

	return fmt.Sprintf("%s %s", method, route)
}

// httpServerAttributes returns the HTTP semantic convention attributes for the server span.
func httpServerAttributes(c *fursy.Context, config Config) []attribute.KeyValue {
	req := c.Request

	attrs := []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String(req.Method),
		semconv.URLPath(req.URL.Path),
		semconv.URLScheme(scheme(req)),
		semconv.NetworkProtocolVersion(httpVersion(req)),
	}

	// Server name (server.address).
	if config.ServerName != "" {
		attrs = append(attrs, semconv.ServerAddress(config.ServerName))
	}

	// Server port (server.port).
	if port := req.URL.Port(); port != "" {
		attrs = append(attrs, semconv.ServerPort(parsePort(port)))
	}

	// Client IP (client.address).
	if config.WithClientIP {
		if clientIP := extractClientIP(req); clientIP != "" {
			attrs = append(attrs, semconv.ClientAddress(clientIP))
		}
	}

	// User-Agent (user_agent.original).
	if config.WithUserAgent {
		if ua := req.UserAgent(); ua != "" {
			attrs = append(attrs, semconv.UserAgentOriginal(ua))
		}
	}

	// Request headers.
	for _, header := range config.WithRequestHeaders {
		if value := req.Header.Get(header); value != "" {
			attrKey := fmt.Sprintf("http.request.header.%s", header)
			attrs = append(attrs, attribute.String(attrKey, value))
		}
	}

	return attrs
}

// extractClientIP extracts the client IP address from the request.
//
// Priority order:
//  1. X-Real-IP header
//  2. X-Forwarded-For header (first IP)
//  3. RemoteAddr
func extractClientIP(req *http.Request) string {
	// Try X-Real-IP header.
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Try X-Forwarded-For header (first IP).
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For: client, proxy1, proxy2
		// Take the first IP (client).
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Fall back to RemoteAddr.
	// RemoteAddr format: "IP:port" or "[IPv6]:port".
	remoteAddr := req.RemoteAddr
	for i := len(remoteAddr) - 1; i >= 0; i-- {
		if remoteAddr[i] == ':' {
			return remoteAddr[:i]
		}
	}

	return remoteAddr
}

// scheme returns the request scheme (http or https).
func scheme(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}

	// Check X-Forwarded-Proto header.
	if proto := req.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}

	return "http"
}

// httpVersion returns the HTTP protocol version (e.g., "1.1", "2", "3").
func httpVersion(req *http.Request) string {
	switch req.ProtoMajor {
	case 1:
		return "1.1"
	case 2:
		return "2"
	case 3:
		return "3"
	default:
		return "1.1"
	}
}

// parsePort parses the port number from a string.
// Returns 0 if parsing fails.
func parsePort(port string) int {
	var p int
	for i := 0; i < len(port); i++ {
		if port[i] < '0' || port[i] > '9' {
			return 0
		}
		p = p*10 + int(port[i]-'0')
	}
	return p
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

// WriteHeader captures the status code and calls the underlying WriteHeader.
func (w *responseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

// Write ensures WriteHeader is called and calls the underlying Write.
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter.
// This is useful for middleware that need to access the original ResponseWriter.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
