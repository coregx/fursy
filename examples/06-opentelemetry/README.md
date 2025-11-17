# OpenTelemetry Observability Example

This example demonstrates comprehensive OpenTelemetry instrumentation for FURSY HTTP router, including:

- **Distributed Tracing** with W3C Trace Context
- **Jaeger Integration** for trace visualization
- **Custom Spans** for business operations
- **Error Recording** in traces
- **HTTP Semantic Conventions**
- **Context Propagation** across services
- **Performance Monitoring** for slow requests

## What You'll Learn

1. How to add OpenTelemetry tracing to a FURSY application
2. How to create custom spans for database queries and external calls
3. How to record errors and performance metrics in traces
4. How to use Jaeger UI to visualize distributed traces
5. Best practices for production observability

## Architecture

```
┌─────────────────┐
│   Your Client   │
│  (curl/browser) │
└────────┬────────┘
         │ HTTP Request
         ▼
┌─────────────────────────────────┐
│   FURSY Application             │
│                                 │
│  ┌──────────────────────────┐  │
│  │ OpenTelemetry Middleware │  │
│  │   (Trace Context Extract)│  │
│  └──────────┬───────────────┘  │
│             │                   │
│             ▼                   │
│  ┌──────────────────────────┐  │
│  │   Route Handlers         │  │
│  │   + Custom Spans         │  │
│  └──────────┬───────────────┘  │
│             │                   │
└─────────────┼───────────────────┘
              │ OTLP/HTTP
              ▼
┌─────────────────────────────────┐
│   Jaeger (All-in-One)           │
│                                 │
│  - OTLP Receiver (4318)         │
│  - Trace Storage                │
│  - Query Service                │
│  - UI (16686)                   │
└─────────────────────────────────┘
```

## Prerequisites

- **Go 1.25+**
- **Docker** and **Docker Compose**
- **curl** (for testing)

## Quick Start

### 1. Start Infrastructure

Start Jaeger and Prometheus using Docker Compose:

```bash
docker-compose up -d
```

Verify containers are running:

```bash
docker-compose ps
```

You should see:
- `fursy-jaeger` on ports 16686, 4318, 4317
- `fursy-prometheus` on port 9090

### 2. Install Dependencies

```bash
go mod download
```

### 3. Run the Application

```bash
go run main.go
```

You should see:

```
INFO Server starting port=8080 service=fursy-otel-example
INFO Jaeger UI available at http://localhost:16686
INFO OTLP endpoint at http://localhost:4318

Try these endpoints:
  curl http://localhost:8080/
  curl http://localhost:8080/users/123
  curl http://localhost:8080/users/123/orders
  curl http://localhost:8080/error
  curl http://localhost:8080/slow
```

### 4. Generate Traffic

Open a new terminal and run:

```bash
# Simple request
curl http://localhost:8080/

# User lookup with custom spans
curl http://localhost:8080/users/123

# Complex operation with nested spans
curl http://localhost:8080/users/456/orders

# Error recording
curl http://localhost:8080/error

# Slow request tracing
curl http://localhost:8080/slow

# POST request
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe"}'
```

### 5. View Traces in Jaeger

1. Open Jaeger UI: http://localhost:16686
2. Select **Service**: `fursy-otel-example`
3. Click **Find Traces**
4. Click on any trace to see details

## Understanding the Traces

### Trace Structure

Each HTTP request creates a trace with the following structure:

```
Trace ID: abc123...
│
├─ Span: GET /users/123 (HTTP Server)
│  ├─ Attributes:
│  │  - http.request.method: GET
│  │  - url.path: /users/123
│  │  - http.response.status_code: 200
│  │  - client.address: 127.0.0.1
│  │
│  ├─ Child Span: database.query.user (Client)
│  │  ├─ Attributes:
│  │  │  - db.operation: SELECT
│  │  │  - db.table: users
│  │  │  - user.id: 123
│  │  │
│  │  └─ Event: user.found
│  │
│  └─ Child Span: business.enrich_user_data
│     └─ Attributes:
│        - user.id: 123
```

### Trace Details

Click on a trace to see:

1. **Timeline View**: Visual representation of span execution
2. **Span Attributes**: HTTP semantic conventions
3. **Events**: Custom events within spans
4. **Tags**: Key-value metadata
5. **Logs**: Error messages and stack traces

### Finding Specific Traces

**By Operation Name:**
```
Operation: GET /users/:id
```

**By HTTP Status Code:**
```
Tags: http.response.status_code=500
```

**By User ID:**
```
Tags: user.id=123
```

**By Duration:**
```
Min Duration: 100ms
Max Duration: 1s
```

## Example Endpoints

### 1. Home (`GET /`)

Simple request with automatic tracing.

**Spans:**
- `GET /` (HTTP server span)

**Try it:**
```bash
curl http://localhost:8080/
```

### 2. Get User (`GET /users/:id`)

Demonstrates custom spans for database operations.

**Spans:**
- `GET /users/:id` (HTTP server)
  - `database.query.user` (DB query)
  - `business.enrich_user_data` (Business logic)

**Try it:**
```bash
curl http://localhost:8080/users/123
```

**In Jaeger:**
- See database query duration
- See custom event "user.found"
- See enrichment processing time

### 3. Get User Orders (`GET /users/:id/orders`)

Demonstrates nested spans and external service calls.

**Spans:**
- `GET /users/:id/orders` (HTTP server)
  - `get_user` (User lookup)
  - `get_user_orders` (DB query)
  - `http.call.payment_service` (External API)

**Try it:**
```bash
curl http://localhost:8080/users/456/orders
```

**In Jaeger:**
- See the full request timeline
- Identify which operation is slowest
- See parallel vs sequential execution

### 4. Error Recording (`GET /error`)

Demonstrates error capture in traces.

**Spans:**
- `GET /error` (HTTP server, status=Error)
  - `failing_operation` (status=Error)

**Try it:**
```bash
curl http://localhost:8080/error
```

**In Jaeger:**
- Trace will be marked with error icon (⚠️)
- Error message visible in span details
- HTTP status code: 500
- Span status: Error

### 5. Slow Request (`GET /slow`)

Demonstrates performance monitoring.

**Spans:**
- `GET /slow` (HTTP server)
  - `cache_lookup` (100ms)
  - `database_query` (200ms)
  - `external_api_call` (300ms)
  - `data_processing` (150ms)

**Try it:**
```bash
curl http://localhost:8080/slow
```

**In Jaeger:**
- Total duration: ~750ms
- See which operation is the bottleneck
- Timeline view shows sequential execution

### 6. Create User (`POST /users`)

Demonstrates POST request tracing.

**Spans:**
- `POST /users` (HTTP server)
  - `validate_user_input` (Validation)
  - `database.insert.user` (DB insert)

**Try it:**
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice"}'
```

## Custom Spans

### Creating Custom Spans

```go
func myHandler(c *fursy.Context) error {
    ctx := c.Request.Context()
    tracer := otel.Tracer("my-service")

    // Create a custom span
    ctx, span := tracer.Start(ctx, "my_operation",
        trace.WithSpanKind(trace.SpanKindClient),
        trace.WithAttributes(
            attribute.String("operation.type", "database"),
            attribute.String("query.id", "123"),
        ),
    )
    defer span.End()

    // Do work...
    result, err := doWork(ctx)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    // Add custom events
    span.AddEvent("work.completed",
        trace.WithAttributes(
            attribute.Int("result.count", len(result)),
        ),
    )

    return c.OK(result)
}
```

### Span Attributes

Common attributes to use:

**Database Operations:**
```go
attribute.String("db.operation", "SELECT"),
attribute.String("db.table", "users"),
attribute.String("db.statement", "SELECT * FROM users WHERE id = ?"),
```

**HTTP Calls:**
```go
attribute.String("http.method", "GET"),
attribute.String("http.url", "https://api.example.com/data"),
attribute.Int("http.status_code", 200),
```

**Business Logic:**
```go
attribute.String("user.id", userID),
attribute.String("operation.type", "payment"),
attribute.Float64("amount", 99.99),
```

## HTTP Semantic Conventions

The OpenTelemetry middleware automatically records these attributes:

### Request Attributes

- `http.request.method` - HTTP method (GET, POST, etc.)
- `url.path` - Request path
- `url.scheme` - http or https
- `network.protocol.version` - HTTP/1.1, HTTP/2, HTTP/3
- `server.address` - Service name
- `client.address` - Client IP
- `user_agent.original` - User-Agent header

### Response Attributes

- `http.response.status_code` - HTTP status code (200, 404, 500, etc.)

### Custom Headers

Configure in middleware:

```go
router.Use(opentelemetry.MiddlewareWithConfig(opentelemetry.Config{
    WithRequestHeaders: []string{"X-Request-ID", "X-Correlation-ID"},
}))
```

## Error Tracing

### Automatic Error Recording

Errors returned from handlers are automatically recorded:

```go
func handler(c *fursy.Context) error {
    // This error will be recorded in the span
    return c.Error(500, fursy.Problem{
        Type:   "database_error",
        Title:  "Database Connection Failed",
        Status: 500,
        Detail: "Could not connect to database",
    })
}
```

**In Jaeger:**
- Span marked with error status
- Error message in span details
- Stack trace (if available)

### Manual Error Recording

```go
func handler(c *fursy.Context) error {
    ctx := c.Request.Context()
    span := trace.SpanFromContext(ctx)

    if err := validate(data); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "validation failed")
        return c.Error(400, ...)
    }

    return c.OK(data)
}
```

## Performance Optimization

### Sampling

For high-traffic production services, use sampling:

```go
// Sample 10% of traces
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)),
    sdktrace.WithBatcher(exporter),
)
```

### Skip Health Checks

Reduce noise by skipping health check endpoints:

```go
router.Use(opentelemetry.MiddlewareWithConfig(opentelemetry.Config{
    Skipper: func(c *fursy.Context) bool {
        path := c.Request.URL.Path
        return path == "/health" || path == "/metrics" || path == "/ready"
    },
}))
```

### Batch Span Processor

Use batch processor for better performance (already configured in this example):

```go
sdktrace.WithBatcher(exporter)
```

## Troubleshooting

### Traces Not Appearing in Jaeger

1. **Check Jaeger is running:**
   ```bash
   docker-compose ps
   curl http://localhost:16686
   ```

2. **Check OTLP endpoint:**
   ```bash
   curl http://localhost:4318/v1/traces
   ```

3. **Check application logs:**
   Look for tracer initialization errors

4. **Verify service name:**
   In Jaeger UI, check if `fursy-otel-example` appears in service list

### Connection Refused

If you see "connection refused" errors:

1. **Check OTLP endpoint in code matches docker-compose:**
   - Code: `localhost:4318`
   - Docker: Jaeger exposes port 4318

2. **On Linux, use container IP instead of localhost**

### No Custom Spans

If custom spans don't appear:

1. **Ensure context propagation:**
   ```go
   ctx := c.Request.Context()  // Get context from request
   ctx, span := tracer.Start(ctx, "my_span")  // Use context
   defer span.End()  // Always end the span
   ```

2. **Check span is child of request span:**
   Use the context from `c.Request.Context()`

### Slow Trace Upload

If traces appear slowly in Jaeger:

1. **Check batch processor settings:**
   ```go
   sdktrace.WithBatcher(exporter,
       sdktrace.WithBatchTimeout(time.Second),
   )
   ```

2. **Flush on shutdown:**
   The example already calls `tp.Shutdown()` which flushes traces

## Production Best Practices

### 1. Use Sampling

Don't trace every request in production:

```go
// Probabilistic sampling (10%)
sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1))

// Or parent-based with fallback to 10%
sdktrace.WithSampler(sdktrace.ParentBased(
    sdktrace.TraceIDRatioBased(0.1),
))
```

### 2. Configure Resource Attributes

Add service metadata:

```go
import "go.opentelemetry.io/otel/sdk/resource"
import semconv "go.opentelemetry.io/otel/semconv/v1.27.0"

res := resource.NewWithAttributes(
    semconv.SchemaURL,
    semconv.ServiceName("my-service"),
    semconv.ServiceVersion("1.0.0"),
    semconv.DeploymentEnvironment("production"),
)

tp := sdktrace.NewTracerProvider(
    sdktrace.WithResource(res),
    sdktrace.WithBatcher(exporter),
)
```

### 3. Use OTLP Collector

For production, use OpenTelemetry Collector instead of direct Jaeger:

```go
exporter, _ := otlptracehttp.New(ctx,
    otlptracehttp.WithEndpoint("otel-collector.example.com:4318"),
    otlptracehttp.WithHeaders(map[string]string{
        "Authorization": "Bearer " + token,
    }),
)
```

### 4. Set Timeouts

Configure timeouts for tracer shutdown:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := tp.Shutdown(ctx); err != nil {
    log.Printf("Failed to shutdown tracer: %v", err)
}
```

### 5. Monitor Exporter Health

Check exporter errors:

```go
// Use a logger that captures errors
logger := logr.New(...)
otel.SetLogger(logger)
```

## Next Steps

1. **Explore Metrics:** Try the `opentelemetry.Metrics()` middleware for HTTP metrics
2. **Add More Examples:** Create custom spans for your business logic
3. **Production Setup:** Deploy OpenTelemetry Collector
4. **Alerting:** Set up alerts based on trace data
5. **Distributed Tracing:** Propagate context to downstream services

## Additional Resources

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [HTTP Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/http/http-spans/)
- [FURSY OpenTelemetry Plugin](../../plugins/opentelemetry/README.md)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)

## Cleanup

Stop and remove containers:

```bash
docker-compose down
```

Remove volumes:

```bash
docker-compose down -v
```

## License

MIT License - see [LICENSE](../../LICENSE) for details.
