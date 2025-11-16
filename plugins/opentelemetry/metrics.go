// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package opentelemetry

import (
	"net/http"
	"time"

	"github.com/coregx/fursy"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// MetricsConfig holds the OpenTelemetry metrics middleware configuration.
type MetricsConfig struct {
	// MeterProvider provides the meter for creating metrics.
	// If not set, the global MeterProvider is used.
	MeterProvider metric.MeterProvider

	// Skipper defines a function to skip middleware for certain requests.
	// If it returns true, the request will not be measured.
	// Example: skip /health and /metrics endpoints.
	Skipper func(*fursy.Context) bool

	// ServerName is the logical server name to use in metric attributes.
	// Maps to server.address semantic convention.
	ServerName string

	// ExplicitBucketBoundaries defines custom histogram buckets for latency.
	// Default (nil): uses OpenTelemetry recommended buckets for HTTP latency:
	// [0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10] seconds.
	ExplicitBucketBoundaries []float64

	// RecordInFlightRequests enables tracking of concurrent requests.
	// Default: false (to reduce cardinality).
	RecordInFlightRequests bool
}

// defaultBuckets are the OpenTelemetry recommended histogram buckets for HTTP server latency.
// Based on OpenTelemetry HTTP Metrics semantic conventions.
// Units: seconds.
var defaultBuckets = []float64{
	0.005, // 5ms
	0.01,  // 10ms
	0.025, // 25ms
	0.05,  // 50ms
	0.075, // 75ms
	0.1,   // 100ms
	0.25,  // 250ms
	0.5,   // 500ms
	0.75,  // 750ms
	1,     // 1s
	2.5,   // 2.5s
	5,     // 5s
	7.5,   // 7.5s
	10,    // 10s
}

// metricsInstruments holds the OpenTelemetry metric instruments.
type metricsInstruments struct {
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	requestSize     metric.Int64Histogram
	responseSize    metric.Int64Histogram
	activeRequests  metric.Int64UpDownCounter
	measureInflight bool
	serverName      string
	skipper         func(*fursy.Context) bool
}

// Metrics returns a FURSY middleware that records HTTP metrics using OpenTelemetry.
//
// The middleware records the following metrics following OpenTelemetry HTTP semantic conventions:
//
//   - http.server.request.duration (Histogram) - Request duration in seconds
//   - http.server.request.count (Counter) - Total number of requests
//   - http.server.request.size (Histogram) - Request body size in bytes
//   - http.server.response.size (Histogram) - Response body size in bytes
//   - http.server.active_requests (UpDownCounter) - Number of active requests (optional)
//
// Attributes recorded:
//   - http.request.method - HTTP method (GET, POST, etc.)
//   - http.response.status_code - HTTP status code
//   - server.address - Server name (from config)
//
// Example:
//
//	router := fursy.New()
//	router.Use(opentelemetry.Metrics("my-service"))
//
//	// With custom configuration:
//	router.Use(opentelemetry.MetricsWithConfig(opentelemetry.MetricsConfig{
//	    ServerName: "api.example.com",
//	    Skipper: func(c *fursy.Context) bool {
//	        return c.Request.URL.Path == "/health"
//	    },
//	    RecordInFlightRequests: true,
//	}))
func Metrics(serverName string) fursy.HandlerFunc {
	return MetricsWithConfig(MetricsConfig{
		ServerName: serverName,
	})
}

// MetricsWithConfig returns a FURSY middleware with custom metrics configuration.
func MetricsWithConfig(config MetricsConfig) fursy.HandlerFunc {
	// Set defaults.
	if config.MeterProvider == nil {
		config.MeterProvider = otel.GetMeterProvider()
	}

	if config.ExplicitBucketBoundaries == nil {
		config.ExplicitBucketBoundaries = defaultBuckets
	}

	meter := config.MeterProvider.Meter(
		ScopeName,
		metric.WithInstrumentationVersion(Version),
	)

	// Create metric instruments.
	instruments := &metricsInstruments{
		measureInflight: config.RecordInFlightRequests,
		serverName:      config.ServerName,
		skipper:         config.Skipper,
	}

	// http.server.request.duration - REQUIRED by OpenTelemetry spec.
	var err error
	instruments.requestDuration, err = meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Measures the duration of inbound HTTP requests"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(config.ExplicitBucketBoundaries...),
	)
	if err != nil {
		panic(err) // Should never happen unless meter is misconfigured.
	}

	// http.server.request.count - Total request counter.
	instruments.requestCounter, err = meter.Int64Counter(
		"http.server.request.count",
		metric.WithDescription("Measures the number of inbound HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		panic(err)
	}

	// http.server.request.size - Request body size.
	instruments.requestSize, err = meter.Int64Histogram(
		"http.server.request.size",
		metric.WithDescription("Measures the size of HTTP request messages"),
		metric.WithUnit("By"),
	)
	if err != nil {
		panic(err)
	}

	// http.server.response.size - Response body size.
	instruments.responseSize, err = meter.Int64Histogram(
		"http.server.response.size",
		metric.WithDescription("Measures the size of HTTP response messages"),
		metric.WithUnit("By"),
	)
	if err != nil {
		panic(err)
	}

	// http.server.active_requests - In-flight requests (optional).
	if config.RecordInFlightRequests {
		instruments.activeRequests, err = meter.Int64UpDownCounter(
			"http.server.active_requests",
			metric.WithDescription("Measures the number of concurrent HTTP requests that are currently in-flight"),
			metric.WithUnit("{request}"),
		)
		if err != nil {
			panic(err)
		}
	}

	return func(c *fursy.Context) error {
		// Skip if Skipper returns true.
		if instruments.skipper != nil && instruments.skipper(c) {
			return c.Next()
		}

		start := time.Now()
		ctx := c.Request.Context()

		// Wrap ResponseWriter to capture status code and bytes written.
		wrapper := &metricsResponseWriter{
			ResponseWriter: c.Response,
			statusCode:     200, // Default to 200 OK.
		}
		c.Response = wrapper

		// Record in-flight request start.
		if instruments.measureInflight {
			instruments.activeRequests.Add(ctx, 1, metric.WithAttributes(
				serverAddressAttribute(instruments.serverName)...),
			)
		}

		// Execute the handler chain.
		err := c.Next()

		// Calculate duration.
		duration := time.Since(start).Seconds()

		// Build metric attributes.
		attrs := httpMetricAttributes(c, instruments.serverName, wrapper.statusCode)

		// Record metrics.
		instruments.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
		instruments.requestDuration.Record(ctx, duration, metric.WithAttributes(attrs...))

		// Record request/response sizes.
		if c.Request.ContentLength > 0 {
			instruments.requestSize.Record(ctx, c.Request.ContentLength, metric.WithAttributes(attrs...))
		}

		if wrapper.bytesWritten > 0 {
			instruments.responseSize.Record(ctx, wrapper.bytesWritten, metric.WithAttributes(attrs...))
		}

		// Record in-flight request end.
		if instruments.measureInflight {
			instruments.activeRequests.Add(ctx, -1, metric.WithAttributes(
				serverAddressAttribute(instruments.serverName)...),
			)
		}

		return err
	}
}

// httpMetricAttributes returns the HTTP semantic convention attributes for metrics.
func httpMetricAttributes(c *fursy.Context, serverName string, statusCode int) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String(c.Request.Method),
		semconv.HTTPResponseStatusCode(statusCode),
	}

	// Server name (server.address).
	if serverName != "" {
		attrs = append(attrs, semconv.ServerAddress(serverName))
	}

	return attrs
}

// serverAddressAttribute returns server.address attribute if serverName is set.
func serverAddressAttribute(serverName string) []attribute.KeyValue {
	if serverName == "" {
		return nil
	}
	return []attribute.KeyValue{semconv.ServerAddress(serverName)}
}

// metricsResponseWriter wraps http.ResponseWriter to capture status code and bytes written.
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	wroteHeader  bool
}

// WriteHeader captures the status code and calls the underlying WriteHeader.
func (w *metricsResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
		w.ResponseWriter.WriteHeader(code)
	}
}

// Write ensures WriteHeader is called and captures bytes written.
func (w *metricsResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(200)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += int64(n)
	return n, err
}

// Unwrap returns the underlying ResponseWriter.
// This is useful for middleware that need to access the original ResponseWriter.
func (w *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
