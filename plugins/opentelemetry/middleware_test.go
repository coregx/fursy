// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package opentelemetry

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/coregx/fursy"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

// setupTestTracer creates a test tracer with in-memory exporter.
func setupTestTracer() (*sdktrace.TracerProvider, *tracetest.InMemoryExporter) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)
	return tp, exporter
}

func TestMiddleware_BasicTracing(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	router := fursy.New()
	router.Use(Middleware("test-service"))

	router.GET("/users/:id", func(c *fursy.Context) error {
		return c.String(200, "User 123")
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Check response.
	if rec.Code != 200 {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Check span.
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]

	// Check span name.
	if span.Name != "GET /users/123" {
		t.Errorf("expected span name 'GET /users/123', got '%s'", span.Name)
	}

	// Check span kind.
	if span.SpanKind != trace.SpanKindServer {
		t.Errorf("expected span kind SERVER, got %v", span.SpanKind)
	}

	// Check HTTP attributes.
	attrs := span.Attributes
	hasMethod := false
	hasPath := false
	hasStatusCode := false

	for _, attr := range attrs {
		if attr.Key == semconv.HTTPRequestMethodKey && attr.Value.AsString() == "GET" {
			hasMethod = true
		}
		if attr.Key == "url.path" && attr.Value.AsString() == "/users/123" {
			hasPath = true
		}
		if attr.Key == semconv.HTTPResponseStatusCodeKey && attr.Value.AsInt64() == 200 {
			hasStatusCode = true
		}
	}

	if !hasMethod {
		t.Error("span missing http.request.method attribute")
	}
	if !hasPath {
		t.Error("span missing url.path attribute")
	}
	if !hasStatusCode {
		t.Error("span missing http.response.status_code attribute")
	}

	// Check span status (should be unset for 2xx).
	if span.Status.Code != codes.Unset {
		t.Errorf("expected span status Unset, got %v", span.Status.Code)
	}
}

func TestMiddleware_ErrorRecording(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	router := fursy.New()
	router.Use(Middleware("test-service"))

	testErr := errors.New("database connection failed")

	router.GET("/users", func(c *fursy.Context) error {
		return testErr
	})

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Check span.
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]

	// Check span status.
	if span.Status.Code != codes.Error {
		t.Errorf("expected span status Error, got %v", span.Status.Code)
	}

	if span.Status.Description != testErr.Error() {
		t.Errorf("expected span status description '%s', got '%s'", testErr.Error(), span.Status.Description)
	}

	// Check error event.
	if len(span.Events) == 0 {
		t.Fatal("expected span to have error event")
	}

	event := span.Events[0]
	if event.Name != "exception" {
		t.Errorf("expected event name 'exception', got '%s'", event.Name)
	}
}

func TestMiddleware_ContextPropagation(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	// Set up W3C Trace Box propagator.
	otel.SetTextMapPropagator(propagation.TraceContext{})

	router := fursy.New()
	router.Use(Middleware("test-service"))

	var extractedSpanContext trace.SpanContext

	router.GET("/users", func(c *fursy.Context) error {
		// Extract span context from request context.
		extractedSpanContext = trace.SpanContextFromContext(c.Request.Context())
		return c.String(200, "OK")
	})

	// Create a parent span and inject it into request headers.
	tracer := tp.Tracer(ScopeName)
	ctx, parentSpan := tracer.Start(context.Background(), "parent-span")
	defer parentSpan.End()

	req := httptest.NewRequest("GET", "/users", nil)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Check that child span has correct parent.
	spans := exporter.GetSpans()
	if len(spans) < 1 {
		t.Fatal("expected at least 1 span")
	}

	// Find the middleware span (not the parent span).
	var middlewareSpan *tracetest.SpanStub
	for i := range spans {
		if spans[i].Name == "GET /users" {
			middlewareSpan = &spans[i]
			break
		}
	}

	if middlewareSpan == nil {
		t.Fatal("middleware span not found")
	}

	// Check parent-child relationship.
	if middlewareSpan.Parent.TraceID() != parentSpan.SpanContext().TraceID() {
		t.Error("middleware span should have same TraceID as parent")
	}

	if middlewareSpan.Parent.SpanID() != parentSpan.SpanContext().SpanID() {
		t.Error("middleware span should have parent's SpanID as parent")
	}

	// Check that handler received the trace context.
	if !extractedSpanContext.IsValid() {
		t.Error("handler did not receive valid span context")
	}
}

func TestMiddleware_Skipper(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	router := fursy.New()
	router.Use(MiddlewareWithConfig(Config{
		ServerName: "test-service",
		Skipper: func(c *fursy.Context) bool {
			// Skip /health endpoint.
			return c.Request.URL.Path == "/health"
		},
	}))

	router.GET("/health", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "Users")
	})

	// Request to /health (should be skipped).
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if len(exporter.GetSpans()) != 0 {
		t.Error("/health should not create a span")
	}

	// Request to /users (should create span).
	req = httptest.NewRequest("GET", "/users", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if len(exporter.GetSpans()) != 1 {
		t.Errorf("expected 1 span for /users, got %d", len(exporter.GetSpans()))
	}
}

func TestMiddleware_CustomSpanNameFormatter(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	router := fursy.New()
	router.Use(MiddlewareWithConfig(Config{
		ServerName: "test-service",
		SpanNameFormatter: func(c *fursy.Context) string {
			return fmt.Sprintf("[%s] %s", c.Request.Method, c.Request.URL.Path)
		},
	}))

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	if spans[0].Name != "[GET] /users" {
		t.Errorf("expected span name '[GET] /users', got '%s'", spans[0].Name)
	}
}

func TestMiddleware_HTTPAttributes(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	router := fursy.New()
	router.Use(MiddlewareWithConfig(Config{
		ServerName:    "api.example.com",
		WithClientIP:  true,
		WithUserAgent: true,
	}))

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("User-Agent", "TestClient/1.0")
	req.Header.Set("X-Real-IP", "192.168.1.100")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	attrs := spans[0].Attributes

	// Check server.address.
	hasServerAddress := false
	hasClientAddress := false
	hasUserAgent := false

	for _, attr := range attrs {
		if attr.Key == semconv.ServerAddressKey && attr.Value.AsString() == "api.example.com" {
			hasServerAddress = true
		}
		if attr.Key == semconv.ClientAddressKey && attr.Value.AsString() == "192.168.1.100" {
			hasClientAddress = true
		}
		if attr.Key == semconv.UserAgentOriginalKey && attr.Value.AsString() == "TestClient/1.0" {
			hasUserAgent = true
		}
	}

	if !hasServerAddress {
		t.Error("span missing server.address attribute")
	}
	if !hasClientAddress {
		t.Error("span missing client.address attribute")
	}
	if !hasUserAgent {
		t.Error("span missing user_agent.original attribute")
	}
}

func TestMiddleware_RequestHeaders(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	router := fursy.New()
	router.Use(MiddlewareWithConfig(Config{
		ServerName:         "test-service",
		WithRequestHeaders: []string{"X-Request-ID", "X-Correlation-ID"},
	}))

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/users", nil)
	req.Header.Set("X-Request-ID", "req-123")
	req.Header.Set("X-Correlation-ID", "corr-456")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	attrs := spans[0].Attributes

	hasRequestID := false
	hasCorrelationID := false

	for _, attr := range attrs {
		if attr.Key == attribute.Key("http.request.header.X-Request-ID") && attr.Value.AsString() == "req-123" {
			hasRequestID = true
		}
		if attr.Key == attribute.Key("http.request.header.X-Correlation-ID") && attr.Value.AsString() == "corr-456" {
			hasCorrelationID = true
		}
	}

	if !hasRequestID {
		t.Error("span missing http.request.header.X-Request-ID attribute")
	}
	if !hasCorrelationID {
		t.Error("span missing http.request.header.X-Correlation-ID attribute")
	}
}

func TestMiddleware_ResponseHeaders(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	router := fursy.New()
	router.Use(MiddlewareWithConfig(Config{
		ServerName:          "test-service",
		WithResponseHeaders: []string{"X-Response-ID", "X-Rate-Limit"},
	}))

	router.GET("/users", func(c *fursy.Context) error {
		c.SetHeader("X-Response-ID", "resp-789")
		c.SetHeader("X-Rate-Limit", "100")
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	attrs := spans[0].Attributes

	hasResponseID := false
	hasRateLimit := false

	for _, attr := range attrs {
		if attr.Key == attribute.Key("http.response.header.X-Response-ID") && attr.Value.AsString() == "resp-789" {
			hasResponseID = true
		}
		if attr.Key == attribute.Key("http.response.header.X-Rate-Limit") && attr.Value.AsString() == "100" {
			hasRateLimit = true
		}
	}

	if !hasResponseID {
		t.Error("span missing http.response.header.X-Response-ID attribute")
	}
	if !hasRateLimit {
		t.Error("span missing http.response.header.X-Rate-Limit attribute")
	}
}

func TestMiddleware_4xxStatus(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	router := fursy.New()
	router.Use(Middleware("test-service"))

	router.GET("/users/:id", func(c *fursy.Context) error {
		return c.String(404, "Not Found")
	})

	req := httptest.NewRequest("GET", "/users/999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]

	// Check status code attribute.
	var statusCode int64
	for _, attr := range span.Attributes {
		if attr.Key == semconv.HTTPResponseStatusCodeKey {
			statusCode = attr.Value.AsInt64()
		}
	}

	if statusCode != 404 {
		t.Errorf("expected status code 404, got %d", statusCode)
	}

	// Check span status (should be Error for 4xx).
	if span.Status.Code != codes.Error {
		t.Errorf("expected span status Error for 404, got %v", span.Status.Code)
	}

	if span.Status.Description != "Not Found" {
		t.Errorf("expected span status description 'Not Found', got '%s'", span.Status.Description)
	}
}

func TestMiddleware_5xxStatus(t *testing.T) {
	tp, exporter := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	router := fursy.New()
	router.Use(Middleware("test-service"))

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(500, "Internal Server Error")
	})

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]

	// Check status code attribute.
	var statusCode int64
	for _, attr := range span.Attributes {
		if attr.Key == semconv.HTTPResponseStatusCodeKey {
			statusCode = attr.Value.AsInt64()
		}
	}

	if statusCode != 500 {
		t.Errorf("expected status code 500, got %d", statusCode)
	}

	// Check span status (should be Error for 5xx).
	if span.Status.Code != codes.Error {
		t.Errorf("expected span status Error for 500, got %v", span.Status.Code)
	}
}

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name           string
		realIP         string
		forwardedFor   string
		remoteAddr     string
		expectedResult string
	}{
		{
			name:           "X-Real-IP header",
			realIP:         "1.2.3.4",
			forwardedFor:   "",
			remoteAddr:     "5.6.7.8:12345",
			expectedResult: "1.2.3.4",
		},
		{
			name:           "X-Forwarded-For header (single IP)",
			realIP:         "",
			forwardedFor:   "1.2.3.4",
			remoteAddr:     "5.6.7.8:12345",
			expectedResult: "1.2.3.4",
		},
		{
			name:           "X-Forwarded-For header (multiple IPs)",
			realIP:         "",
			forwardedFor:   "1.2.3.4, 5.6.7.8, 9.10.11.12",
			remoteAddr:     "13.14.15.16:12345",
			expectedResult: "1.2.3.4",
		},
		{
			name:           "RemoteAddr fallback",
			realIP:         "",
			forwardedFor:   "",
			remoteAddr:     "1.2.3.4:12345",
			expectedResult: "1.2.3.4",
		},
		{
			name:           "RemoteAddr IPv6",
			realIP:         "",
			forwardedFor:   "",
			remoteAddr:     "[::1]:12345",
			expectedResult: "[::1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)

			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			if tt.forwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.forwardedFor)
			}
			req.RemoteAddr = tt.remoteAddr

			result := extractClientIP(req)

			if result != tt.expectedResult {
				t.Errorf("expected '%s', got '%s'", tt.expectedResult, result)
			}
		})
	}
}

func TestScheme(t *testing.T) {
	tests := []struct {
		name     string
		hasTLS   bool
		xProto   string
		expected string
	}{
		{"HTTPS with TLS", true, "", "https"},
		{"HTTP without TLS", false, "", "http"},
		{"X-Forwarded-Proto https", false, "https", "https"},
		{"X-Forwarded-Proto http", false, "http", "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)

			if tt.hasTLS {
				// Create a non-nil TLS connection state.
				req.TLS = &tls.ConnectionState{}
			}

			if tt.xProto != "" {
				req.Header.Set("X-Forwarded-Proto", tt.xProto)
			}

			result := scheme(req)

			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"8080", 8080},
		{"443", 443},
		{"0", 0},
		{"65535", 65535},
		{"invalid", 0},
		{"12a34", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parsePort(tt.input)
			if result != tt.expected {
				t.Errorf("parsePort(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
