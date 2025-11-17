// Package main demonstrates OpenTelemetry tracing and metrics with fursy.
//
// This example shows how to instrument a fursy application with:
//   - W3C Trace Context propagation
//   - Distributed tracing with Jaeger
//   - Custom spans for business logic
//   - Error recording in traces
//   - HTTP semantic conventions
//
// Prerequisites:
//   - Docker (for Jaeger and Prometheus)
//   - Go 1.25+
//
// Start infrastructure:
//
//	docker-compose up -d
//
// Run the server:
//
//	go run main.go
//
// View traces:
//
//	http://localhost:16686 (Jaeger UI)
//
// Generate traffic:
//
//	curl http://localhost:8080/
//	curl http://localhost:8080/users/123
//	curl http://localhost:8080/users/456/orders
//	curl http://localhost:8080/error
package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName    = "fursy-otel-example"
	serviceVersion = "0.1.0"
)

func main() {
	// Initialize OpenTelemetry tracing.
	tp, err := initTracer()
	if err != nil {
		slog.Error("Failed to initialize tracer", "error", err)
		os.Exit(1)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("Failed to shutdown tracer", "error", err)
		}
	}()

	// Create fursy router.
	router := fursy.New()

	// Add OpenTelemetry middleware.
	// This middleware:
	//   - Extracts W3C Trace Context from incoming requests
	//   - Creates a span for each request
	//   - Records HTTP semantic convention attributes
	//   - Propagates trace context to downstream services
	router.Use(opentelemetry.MiddlewareWithConfig(opentelemetry.Config{
		ServerName: serviceName,

		// Skip health check endpoint to reduce noise.
		Skipper: func(c *fursy.Context) bool {
			return c.Request.URL.Path == "/health"
		},

		// Capture request ID header for correlation.
		WithRequestHeaders: []string{"X-Request-ID"},

		// Include client IP and User-Agent.
		WithClientIP:  true,
		WithUserAgent: true,
	}))

	// Define routes.
	router.GET("/", homeHandler)
	router.GET("/health", healthHandler)
	router.GET("/users/:id", getUserHandler)
	router.GET("/users/:id/orders", getUserOrdersHandler)
	router.POST("/users", createUserHandler)
	router.GET("/error", errorHandler)
	router.GET("/slow", slowHandler)

	// Start server.
	port := getEnv("PORT", "8080")
	addr := ":" + port

	slog.Info("Server starting", "port", port, "service", serviceName)
	slog.Info("Jaeger UI available at http://localhost:16686")
	slog.Info("OTLP endpoint at http://localhost:4318")
	slog.Info("")
	slog.Info("Try these endpoints:")
	slog.Info("  curl http://localhost:8080/")
	slog.Info("  curl http://localhost:8080/users/123")
	slog.Info("  curl http://localhost:8080/users/123/orders")
	slog.Info("  curl http://localhost:8080/error")
	slog.Info("  curl http://localhost:8080/slow")

	// Create HTTP server.
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	slog.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	slog.Info("Server stopped")
}

// initTracer initializes OpenTelemetry tracing with OTLP/HTTP exporter.
// Sends traces to Jaeger via OpenTelemetry Collector protocol.
func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// Create OTLP HTTP exporter (Jaeger supports OTLP).
	endpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318")
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // Use insecure for local development.
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create trace provider.
	tp := sdktrace.NewTracerProvider(
		// Use batch span processor for better performance.
		sdktrace.WithBatcher(exporter),

		// Sample all traces (use sampling in production).
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider.
	otel.SetTracerProvider(tp)

	return tp, nil
}

// homeHandler demonstrates a simple handler with automatic tracing.
func homeHandler(c *fursy.Context) error {
	return c.OK(map[string]string{
		"message": "OpenTelemetry Example for FURSY",
		"service": serviceName,
		"version": serviceVersion,
		"docs":    "http://localhost:16686",
	})
}

// healthHandler is a health check endpoint (skipped by middleware).
func healthHandler(c *fursy.Context) error {
	return c.OK(map[string]string{
		"status": "healthy",
	})
}

// getUserHandler demonstrates custom spans within a handler.
// Shows how to create child spans for specific operations.
func getUserHandler(c *fursy.Context) error {
	ctx := c.Request.Context()
	tracer := otel.Tracer(serviceName)

	userID := c.Param("id")

	// Create a custom span for database operation.
	ctx, dbSpan := tracer.Start(ctx, "database.query.user",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.operation", "SELECT"),
			attribute.String("db.table", "users"),
			attribute.String("user.id", userID),
		),
	)

	// Simulate database query.
	time.Sleep(time.Duration(10+rand.Intn(40)) * time.Millisecond)

	// Record custom events.
	dbSpan.AddEvent("user.found",
		trace.WithAttributes(
			attribute.String("user.id", userID),
			attribute.String("user.name", "User "+userID),
		),
	)

	dbSpan.End()

	// Create another span for business logic.
	_, bizSpan := tracer.Start(ctx, "business.enrich_user_data",
		trace.WithAttributes(
			attribute.String("user.id", userID),
		),
	)

	// Simulate business logic processing.
	time.Sleep(time.Duration(5+rand.Intn(15)) * time.Millisecond)

	bizSpan.End()

	// Return response.
	return c.OK(map[string]any{
		"id":    userID,
		"name":  "User " + userID,
		"email": fmt.Sprintf("user%s@example.com", userID),
		"role":  "admin",
	})
}

// getUserOrdersHandler demonstrates nested spans and distributed tracing.
// Shows how to propagate trace context to external services.
func getUserOrdersHandler(c *fursy.Context) error {
	ctx := c.Request.Context()
	tracer := otel.Tracer(serviceName)

	userID := c.Param("id")

	// Span for fetching user.
	ctx, userSpan := tracer.Start(ctx, "get_user",
		trace.WithAttributes(
			attribute.String("user.id", userID),
		),
	)
	time.Sleep(time.Duration(10+rand.Intn(20)) * time.Millisecond)
	userSpan.End()

	// Span for fetching orders.
	ctx, ordersSpan := tracer.Start(ctx, "get_user_orders",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("user.id", userID),
			attribute.String("db.operation", "SELECT"),
			attribute.String("db.table", "orders"),
		),
	)

	// Simulate database query.
	time.Sleep(time.Duration(20+rand.Intn(50)) * time.Millisecond)

	orderCount := rand.Intn(10) + 1
	ordersSpan.SetAttributes(attribute.Int("order.count", orderCount))
	ordersSpan.End()

	// Simulate external API call with trace context propagation.
	_, apiSpan := tracer.Start(ctx, "http.call.payment_service",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.method", "GET"),
			attribute.String("http.url", "http://payment-service/api/payments"),
			attribute.String("user.id", userID),
		),
	)
	time.Sleep(time.Duration(30+rand.Intn(70)) * time.Millisecond)
	apiSpan.End()

	orders := make([]map[string]any, orderCount)
	for i := 0; i < orderCount; i++ {
		orders[i] = map[string]any{
			"id":     fmt.Sprintf("order-%d", i+1),
			"amount": rand.Float64() * 1000,
			"status": "completed",
		}
	}

	return c.OK(map[string]any{
		"user_id": userID,
		"orders":  orders,
		"total":   orderCount,
	})
}

// createUserHandler demonstrates POST request tracing with request body.
func createUserHandler(c *fursy.Context) error {
	ctx := c.Request.Context()
	tracer := otel.Tracer(serviceName)

	// Create span for validation.
	_, validationSpan := tracer.Start(ctx, "validate_user_input")
	time.Sleep(5 * time.Millisecond)
	validationSpan.End()

	// Create span for database insert.
	_, dbSpan := tracer.Start(ctx, "database.insert.user",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.operation", "INSERT"),
			attribute.String("db.table", "users"),
		),
	)

	time.Sleep(time.Duration(20+rand.Intn(40)) * time.Millisecond)

	newUserID := fmt.Sprintf("user-%d", time.Now().Unix())
	dbSpan.SetAttributes(attribute.String("user.id", newUserID))
	dbSpan.End()

	return c.OK(map[string]any{
		"id":      newUserID,
		"name":    "New User",
		"created": time.Now().Format(time.RFC3339),
	})
}

// errorHandler demonstrates error recording in traces.
// Errors are automatically captured by the OpenTelemetry middleware.
func errorHandler(c *fursy.Context) error {
	ctx := c.Request.Context()
	tracer := otel.Tracer(serviceName)

	// Create a span that will fail.
	_, span := tracer.Start(ctx, "failing_operation")
	defer span.End()

	// Simulate some work before error.
	time.Sleep(10 * time.Millisecond)

	// Record error in span.
	err := fmt.Errorf("simulated error: database connection failed")
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	// Return error (will be recorded by middleware too).
	return c.Problem(fursy.InternalServerError(
		"A simulated error occurred for demonstration purposes",
	))
}

// slowHandler demonstrates slow request tracing.
// Useful for identifying performance bottlenecks.
func slowHandler(c *fursy.Context) error {
	ctx := c.Request.Context()
	tracer := otel.Tracer(serviceName)

	// Simulate multiple slow operations.
	operations := []struct {
		name     string
		duration time.Duration
	}{
		{"cache_lookup", 100 * time.Millisecond},
		{"database_query", 200 * time.Millisecond},
		{"external_api_call", 300 * time.Millisecond},
		{"data_processing", 150 * time.Millisecond},
	}

	for _, op := range operations {
		_, span := tracer.Start(ctx, op.name)
		time.Sleep(op.duration)
		span.End()
	}

	return c.OK(map[string]any{
		"message":  "Slow operation completed",
		"duration": "~750ms",
	})
}

// getEnv gets an environment variable with a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
