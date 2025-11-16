# OpenTelemetry Plugin for FURSY

Production-ready OpenTelemetry instrumentation for FURSY HTTP router with full HTTP semantic conventions support.

## Features

### Tracing
- ✅ **Automatic HTTP span creation** - W3C Trace Context propagation
- ✅ **HTTP Semantic Conventions** - Full compliance with OpenTelemetry HTTP spec
- ✅ **Error recording** - Automatic error and status tracking
- ✅ **Context propagation** - Seamless distributed tracing

### Metrics
- ✅ **Request duration histogram** - HTTP latency tracking with configurable buckets
- ✅ **Request counter** - Total number of requests by method and status
- ✅ **Request/Response size** - Body size histograms
- ✅ **Active requests** - In-flight request tracking (optional)
- ✅ **Cardinality management** - Low-cardinality labels (method, status, server)

### Common
- ✅ **Flexible configuration** - Skipper, custom formatters, header capture
- ✅ **Zero-overhead filtering** - Skip health checks and metrics endpoints
- ✅ **90% test coverage** - Production-ready quality

## Installation

```bash
go get github.com/coregx/fursy/plugins/opentelemetry
```

## Quick Start

```go
package main

import (
	"context"
	"log"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	// Initialize OpenTelemetry.
	exporter, _ := stdouttrace.New()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	defer tp.Shutdown(context.Background())

	// Create FURSY router with OpenTelemetry middleware.
	router := fursy.New()
	router.Use(opentelemetry.Middleware("my-service"))

	router.GET("/users/:id", func(c *fursy.Context) error {
		return c.String(200, "User 123")
	})

	router.Start(":8080")
}
```

## Metrics Middleware

Track HTTP metrics with OpenTelemetry Metrics API:

```go
package main

import (
	"context"
	"log"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	// Initialize Prometheus exporter.
	exporter, _ := prometheus.New()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(mp)

	// Create FURSY router with Metrics middleware.
	router := fursy.New()
	router.Use(opentelemetry.Metrics("my-service"))

	router.GET("/users/:id", func(c *fursy.Context) error {
		return c.String(200, "User 123")
	})

	// Metrics available:
	// - http.server.request.duration (Histogram)
	// - http.server.request.count (Counter)
	// - http.server.request.size (Histogram)
	// - http.server.response.size (Histogram)

	router.Start(":8080")
}
```

### Metrics Configuration

```go
router.Use(opentelemetry.MetricsWithConfig(opentelemetry.MetricsConfig{
	ServerName: "api.example.com",

	// Skip health checks.
	Skipper: func(c *fursy.Context) bool {
		return c.Request.URL.Path == "/health"
	},

	// Custom histogram buckets (seconds).
	ExplicitBucketBoundaries: []float64{0.01, 0.05, 0.1, 0.5, 1.0},

	// Track in-flight requests (optional, increases cardinality).
	RecordInFlightRequests: true,
}))
```

## Configuration

### Basic Usage

```go
// Simple middleware with default configuration.
router.Use(opentelemetry.Middleware("my-service"))
```

### Advanced Configuration

```go
router.Use(opentelemetry.MiddlewareWithConfig(opentelemetry.Config{
	ServerName: "api.example.com",

	// Skip health checks and metrics.
	Skipper: func(c *fursy.Context) bool {
		path := c.Request.URL.Path
		return path == "/health" || path == "/metrics"
	},

	// Custom span name formatter.
	SpanNameFormatter: func(c *fursy.Context) string {
		return fmt.Sprintf("[%s] %s", c.Request.Method, c.Request.URL.Path)
	},

	// Capture specific headers.
	WithRequestHeaders: []string{"X-Request-ID", "X-Correlation-ID"},
	WithResponseHeaders: []string{"X-Response-ID"},

	// Client IP and User-Agent tracking.
	WithClientIP: true,
	WithUserAgent: true,
}))
```

## HTTP Semantic Conventions

The middleware automatically records the following OpenTelemetry HTTP semantic convention attributes:

### Request Attributes

- `http.request.method` - HTTP method (GET, POST, etc.)
- `url.path` - Request path
- `url.scheme` - HTTP or HTTPS
- `network.protocol.version` - HTTP/1.1, HTTP/2, HTTP/3
- `server.address` - Server name (from config)
- `server.port` - Server port
- `client.address` - Client IP (from X-Real-IP, X-Forwarded-For, or RemoteAddr)
- `user_agent.original` - User-Agent header

### Response Attributes

- `http.response.status_code` - HTTP status code

### Custom Attributes

- `http.request.header.<name>` - Request headers (if configured)
- `http.response.header.<name>` - Response headers (if configured)

## Context Propagation

The middleware automatically handles W3C Trace Context propagation:

- **Incoming requests**: Extracts `traceparent` and `tracestate` headers
- **Outgoing requests**: Context is available in handlers for downstream calls

Example:

```go
router.GET("/users", func(c *fursy.Context) error {
	// Get trace context from request.
	ctx := c.Request.Context()

	// Make downstream HTTP call with trace context.
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://user-service/users", nil)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, _ := http.DefaultClient.Do(req)
	// ... handle response
})
```

## Error Recording

Errors are automatically recorded as span events and status:

```go
router.GET("/users/:id", func(c *fursy.Context) error {
	user, err := getUserByID(id)
	if err != nil {
		// Error will be recorded in span with status=Error.
		return err
	}
	return c.JSON(200, user)
})
```

## Span Naming

Default format: `{method} {path}`

Example span names:
- `GET /users/123`
- `POST /users`
- `DELETE /users/456`

Custom formatter:

```go
SpanNameFormatter: func(c *fursy.Context) string {
	return fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
}
```

## Filtering Requests

Skip health checks, metrics, and other endpoints:

```go
Skipper: func(c *fursy.Context) bool {
	path := c.Request.URL.Path
	return path == "/health" ||
	       path == "/metrics" ||
	       path == "/ready" ||
	       path == "/alive"
}
```

## Best Practices

### 1. Skip Non-Business Endpoints

Health checks and metrics endpoints generate high volumes of spans with little value:

```go
Skipper: func(c *fursy.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/health") ||
	       strings.HasPrefix(c.Request.URL.Path, "/metrics")
}
```

### 2. Capture Correlation IDs

For request tracing across services:

```go
WithRequestHeaders: []string{"X-Request-ID", "X-Correlation-ID", "X-B3-TraceId"}
```

### 3. Use Appropriate Exporters

Development:
```go
exporter, _ := stdouttrace.New()
```

Production:
```go
exporter, _ := otlptracehttp.New(ctx,
	otlptracehttp.WithEndpoint("collector.example.com:4318"),
)
```

### 4. Set Server Name

Helps identify services in distributed traces:

```go
ServerName: "api.example.com"
```

### 5. Sampling

For high-traffic services, use sampling to reduce overhead:

```go
tp := sdktrace.NewTracerProvider(
	sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)), // 10% sampling
	sdktrace.WithBatcher(exporter),
)
```

## Integration Examples

### With Jaeger

```go
exporter, _ := jaeger.New(jaeger.WithCollectorEndpoint(
	jaeger.WithEndpoint("http://localhost:14268/api/traces"),
))
tp := sdktrace.NewTracerProvider(
	sdktrace.WithBatcher(exporter),
)
```

### With Zipkin

```go
exporter, _ := zipkin.New("http://localhost:9411/api/v2/spans")
tp := sdktrace.NewTracerProvider(
	sdktrace.WithBatcher(exporter),
)
```

### With OTLP (OpenTelemetry Collector)

```go
exporter, _ := otlptracehttp.New(ctx,
	otlptracehttp.WithEndpoint("localhost:4318"),
	otlptracehttp.WithInsecure(),
)
tp := sdktrace.NewTracerProvider(
	sdktrace.WithBatcher(exporter),
)
```

## Performance

- **Overhead**: ~100-200ns per request (with sampling)
- **Memory**: Minimal - spans are batched and exported asynchronously
- **Zero allocations**: When request is skipped via Skipper

## License

MIT License - see [LICENSE](../../LICENSE) for details.

## Contributing

Contributions welcome! This is a plugin for the [FURSY](https://github.com/coregx/fursy) HTTP router.

## See Also

- [FURSY Router](https://github.com/coregx/fursy)
- [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
- [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/http/http-spans/)
