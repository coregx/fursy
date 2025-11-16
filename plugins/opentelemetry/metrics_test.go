// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package opentelemetry

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coregx/fursy"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

var ctx = context.Background()

// setupTestMeter creates a test meter with in-memory reader.
func setupTestMeter() (*sdkmetric.MeterProvider, *sdkmetric.ManualReader) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)
	return mp, reader
}

func TestMetrics_RequestDuration(t *testing.T) {
	mp, reader := setupTestMeter()
	defer func() { _ = mp.Shutdown(ctx) }()

	router := fursy.New()
	router.Use(Metrics("test-service"))

	router.GET("/users", func(c *fursy.Context) error {
		time.Sleep(50 * time.Millisecond)
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Collect metrics.
	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Find http.server.request.duration histogram.
	var foundDuration bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.server.request.duration" {
				foundDuration = true

				// Check histogram data.
				histogram, ok := m.Data.(metricdata.Histogram[float64])
				if !ok {
					t.Fatalf("expected Histogram, got %T", m.Data)
				}

				if len(histogram.DataPoints) == 0 {
					t.Fatal("expected at least one data point")
				}

				dp := histogram.DataPoints[0]

				// Check count.
				if dp.Count != 1 {
					t.Errorf("expected count=1, got %d", dp.Count)
				}

				// Check that sum is approximately 50ms (0.05s).
				if dp.Sum < 0.04 || dp.Sum > 0.1 {
					t.Errorf("expected sum around 0.05s, got %f", dp.Sum)
				}

				// Check attributes.
				hasMethod := false
				hasStatus := false
				for _, attr := range dp.Attributes.ToSlice() {
					if string(attr.Key) == "http.request.method" && attr.Value.AsString() == "GET" {
						hasMethod = true
					}
					if string(attr.Key) == "http.response.status_code" && attr.Value.AsInt64() == 200 {
						hasStatus = true
					}
				}

				if !hasMethod {
					t.Error("missing http.request.method attribute")
				}
				if !hasStatus {
					t.Error("missing http.response.status_code attribute")
				}
			}
		}
	}

	if !foundDuration {
		t.Error("http.server.request.duration metric not found")
	}
}

func TestMetrics_RequestCounter(t *testing.T) {
	mp, reader := setupTestMeter()
	defer func() { _ = mp.Shutdown(ctx) }()

	router := fursy.New()
	router.Use(Metrics("test-service"))

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Make 3 requests.
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/users", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	// Collect metrics.
	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Find http.server.request.count counter.
	var foundCounter bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.server.request.count" {
				foundCounter = true

				// Check counter data.
				sum, ok := m.Data.(metricdata.Sum[int64])
				if !ok {
					t.Fatalf("expected Sum, got %T", m.Data)
				}

				if len(sum.DataPoints) == 0 {
					t.Fatal("expected at least one data point")
				}

				dp := sum.DataPoints[0]

				// Check count.
				if dp.Value != 3 {
					t.Errorf("expected count=3, got %d", dp.Value)
				}
			}
		}
	}

	if !foundCounter {
		t.Error("http.server.request.count metric not found")
	}
}

func TestMetrics_ResponseStatusCode(t *testing.T) {
	mp, reader := setupTestMeter()
	defer func() { _ = mp.Shutdown(ctx) }()

	router := fursy.New()
	router.Use(Metrics("test-service"))

	router.GET("/error", func(c *fursy.Context) error {
		return c.String(500, "Internal Server Error")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Collect metrics.
	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Find http.server.request.duration and check status code.
	var foundStatus bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.server.request.duration" {
				histogram := m.Data.(metricdata.Histogram[float64])
				dp := histogram.DataPoints[0]

				for _, attr := range dp.Attributes.ToSlice() {
					if string(attr.Key) == "http.response.status_code" && attr.Value.AsInt64() == 500 {
						foundStatus = true
					}
				}
			}
		}
	}

	if !foundStatus {
		t.Error("status code 500 not found in metrics")
	}
}

func TestMetrics_ServerName(t *testing.T) {
	mp, reader := setupTestMeter()
	defer func() { _ = mp.Shutdown(ctx) }()

	router := fursy.New()
	router.Use(MetricsWithConfig(MetricsConfig{
		ServerName: "api.example.com",
	}))

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Collect metrics.
	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Find http.server.request.duration and check server.address.
	var foundServerName bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.server.request.duration" {
				histogram := m.Data.(metricdata.Histogram[float64])
				dp := histogram.DataPoints[0]

				for _, attr := range dp.Attributes.ToSlice() {
					if string(attr.Key) == "server.address" && attr.Value.AsString() == "api.example.com" {
						foundServerName = true
					}
				}
			}
		}
	}

	if !foundServerName {
		t.Error("server.address attribute not found in metrics")
	}
}

func TestMetrics_Skipper(t *testing.T) {
	mp, reader := setupTestMeter()
	defer func() { _ = mp.Shutdown(ctx) }()

	router := fursy.New()
	router.Use(MetricsWithConfig(MetricsConfig{
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

	// Collect metrics.
	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Should have no metrics.
	metricCount := 0
	for _, sm := range rm.ScopeMetrics {
		metricCount += len(sm.Metrics)
	}

	if metricCount != 0 {
		t.Errorf("/health should not create metrics, got %d metrics", metricCount)
	}

	// Request to /users (should create metrics).
	req = httptest.NewRequest("GET", "/users", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Collect metrics again.
	rm = metricdata.ResourceMetrics{}
	err = reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Should have metrics now.
	metricCount = 0
	for _, sm := range rm.ScopeMetrics {
		metricCount += len(sm.Metrics)
	}

	if metricCount == 0 {
		t.Error("/users should create metrics")
	}
}

func TestMetrics_CustomBuckets(t *testing.T) {
	mp, reader := setupTestMeter()
	defer func() { _ = mp.Shutdown(ctx) }()

	customBuckets := []float64{0.01, 0.05, 0.1, 0.5, 1.0}

	router := fursy.New()
	router.Use(MetricsWithConfig(MetricsConfig{
		ServerName:               "test-service",
		ExplicitBucketBoundaries: customBuckets,
	}))

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Collect metrics.
	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Find http.server.request.duration and check buckets.
	var foundHistogram bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.server.request.duration" {
				foundHistogram = true

				histogram := m.Data.(metricdata.Histogram[float64])
				dp := histogram.DataPoints[0]

				// Check that buckets match custom boundaries.
				if len(dp.Bounds) != len(customBuckets) {
					t.Errorf("expected %d bucket boundaries, got %d", len(customBuckets), len(dp.Bounds))
				}

				for i, bound := range dp.Bounds {
					if bound != customBuckets[i] {
						t.Errorf("bucket[%d]: expected %f, got %f", i, customBuckets[i], bound)
					}
				}
			}
		}
	}

	if !foundHistogram {
		t.Error("http.server.request.duration histogram not found")
	}
}

func TestMetrics_ActiveRequests(t *testing.T) {
	mp, reader := setupTestMeter()
	defer func() { _ = mp.Shutdown(ctx) }()

	router := fursy.New()
	router.Use(MetricsWithConfig(MetricsConfig{
		ServerName:             "test-service",
		RecordInFlightRequests: true,
	}))

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Collect metrics.
	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Find http.server.active_requests metric.
	var foundActiveRequests bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.server.active_requests" {
				foundActiveRequests = true
			}
		}
	}

	if !foundActiveRequests {
		t.Error("http.server.active_requests metric not found (RecordInFlightRequests=true)")
	}
}

func TestMetrics_RequestResponseSize(t *testing.T) {
	mp, reader := setupTestMeter()
	defer func() { _ = mp.Shutdown(ctx) }()

	router := fursy.New()
	router.Use(Metrics("test-service"))

	router.POST("/users", func(c *fursy.Context) error {
		return c.String(200, "User created successfully")
	})

	// Create request with body.
	req := httptest.NewRequest("POST", "/users", nil)
	req.ContentLength = 100 // Simulate 100 bytes request.
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Collect metrics.
	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Find http.server.request.size.
	var foundRequestSize bool
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.server.request.size" {
				foundRequestSize = true

				histogram := m.Data.(metricdata.Histogram[int64])
				if len(histogram.DataPoints) == 0 {
					t.Fatal("expected at least one data point")
				}

				dp := histogram.DataPoints[0]

				// Check sum is 100.
				if dp.Sum != 100 {
					t.Errorf("expected request size sum=100, got %d", dp.Sum)
				}
			}

			// Check response size.
			if m.Name == "http.server.response.size" {
				histogram := m.Data.(metricdata.Histogram[int64])
				if len(histogram.DataPoints) == 0 {
					t.Fatal("expected at least one data point for response size")
				}

				dp := histogram.DataPoints[0]

				// Response is "User created successfully" = 25 bytes.
				if dp.Sum != 25 {
					t.Errorf("expected response size sum=25, got %d", dp.Sum)
				}
			}
		}
	}

	if !foundRequestSize {
		t.Error("http.server.request.size metric not found")
	}
}

func TestMetrics_HTTPMethods(t *testing.T) {
	mp, reader := setupTestMeter()
	defer func() { _ = mp.Shutdown(ctx) }()

	router := fursy.New()
	router.Use(Metrics("test-service"))

	router.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "GET")
	})

	router.POST("/users", func(c *fursy.Context) error {
		return c.String(201, "POST")
	})

	// Test GET.
	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Test POST.
	req = httptest.NewRequest("POST", "/users", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Collect metrics.
	var rm metricdata.ResourceMetrics
	err := reader.Collect(ctx, &rm)
	if err != nil {
		t.Fatal(err)
	}

	// Find http.server.request.count and check methods.
	foundGET := false
	foundPOST := false

	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.server.request.count" {
				sum := m.Data.(metricdata.Sum[int64])

				// We should have 2 data points (one for GET, one for POST).
				if len(sum.DataPoints) != 2 {
					t.Errorf("expected 2 data points, got %d", len(sum.DataPoints))
				}

				for _, dp := range sum.DataPoints {
					for _, attr := range dp.Attributes.ToSlice() {
						if string(attr.Key) == "http.request.method" {
							if attr.Value.AsString() == "GET" {
								foundGET = true
							}
							if attr.Value.AsString() == "POST" {
								foundPOST = true
							}
						}
					}
				}
			}
		}
	}

	if !foundGET {
		t.Error("GET method not found in metrics")
	}
	if !foundPOST {
		t.Error("POST method not found in metrics")
	}
}
