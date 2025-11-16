// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coregx/fursy"
)

// TestLogger tests the default Logger middleware.
func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultLogger(&buf)

	r := fursy.New()
	r.Use(LoggerWithConfig(LoggerConfig{
		Logger: logger,
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()

	// Check log contains expected fields
	if !strings.Contains(output, "HTTP request") {
		t.Error("log should contain 'HTTP request'")
	}
	if !strings.Contains(output, "method=GET") {
		t.Error("log should contain method")
	}
	if !strings.Contains(output, "path=/test") {
		t.Error("log should contain path")
	}
	if !strings.Contains(output, "status=200") {
		t.Error("log should contain status")
	}
	if !strings.Contains(output, "latency_ms") {
		t.Error("log should contain latency")
	}
	if !strings.Contains(output, "ip=") {
		t.Error("log should contain IP")
	}
}

// TestLogger_JSONFormat tests JSON output format.
func TestLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := JSONLogger(&buf)

	r := fursy.New()
	r.Use(LoggerWithConfig(LoggerConfig{
		Logger: logger,
	}))

	r.GET("/api/users", func(c *fursy.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/users", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()

	// Check JSON format
	if !strings.Contains(output, `"msg":"HTTP request"`) {
		t.Error("JSON log should contain msg field")
	}
	if !strings.Contains(output, `"method":"GET"`) {
		t.Error("JSON log should contain method field")
	}
	if !strings.Contains(output, `"path":"/api/users"`) {
		t.Error("JSON log should contain path field")
	}
	if !strings.Contains(output, `"status":200`) {
		t.Error("JSON log should contain status field")
	}
}

// TestLogger_SkipPaths tests skipping specified paths.
func TestLogger_SkipPaths(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultLogger(&buf)

	r := fursy.New()
	r.Use(LoggerWithConfig(LoggerConfig{
		Logger:    logger,
		SkipPaths: []string{"/health", "/metrics"},
	}))

	r.GET("/health", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	r.GET("/api/users", func(c *fursy.Context) error {
		return c.String(200, "users")
	})

	// Request to skipped path
	req1 := httptest.NewRequest("GET", "/health", http.NoBody)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	output1 := buf.String()
	if strings.Contains(output1, "/health") {
		t.Error("skipped path /health should not be logged")
	}

	// Request to normal path
	buf.Reset()
	req2 := httptest.NewRequest("GET", "/api/users", http.NoBody)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	output2 := buf.String()
	if !strings.Contains(output2, "/api/users") {
		t.Error("normal path /api/users should be logged")
	}
}

// TestLogger_SkipFunc tests custom skip function.
func TestLogger_SkipFunc(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultLogger(&buf)

	r := fursy.New()
	r.Use(LoggerWithConfig(LoggerConfig{
		Logger: logger,
		SkipFunc: func(req *http.Request) bool {
			// Skip requests with X-No-Log header
			return req.Header.Get("X-No-Log") == "true"
		},
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Request with X-No-Log header
	req1 := httptest.NewRequest("GET", "/test", http.NoBody)
	req1.Header.Set("X-No-Log", "true")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	output1 := buf.String()
	if output1 != "" {
		t.Error("request with X-No-Log should not be logged")
	}

	// Normal request
	req2 := httptest.NewRequest("GET", "/test", http.NoBody)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	output2 := buf.String()
	if !strings.Contains(output2, "/test") {
		t.Error("normal request should be logged")
	}
}

// TestLogger_StatusCodes tests different status code handling.
func TestLogger_StatusCodes(t *testing.T) {
	tests := []struct {
		name          string
		status        int
		expectedLevel string
		handler       fursy.HandlerFunc
	}{
		{
			name:          "2xx success - INFO level",
			status:        200,
			expectedLevel: "INFO",
			handler: func(c *fursy.Context) error {
				return c.String(200, "OK")
			},
		},
		{
			name:          "4xx client error - WARN level",
			status:        404,
			expectedLevel: "WARN",
			handler: func(c *fursy.Context) error {
				return c.String(404, "Not Found")
			},
		},
		{
			name:          "5xx server error - ERROR level",
			status:        500,
			expectedLevel: "ERROR",
			handler: func(c *fursy.Context) error {
				return c.String(500, "Internal Server Error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := DefaultLogger(&buf)

			r := fursy.New()
			r.Use(LoggerWithConfig(LoggerConfig{
				Logger: logger,
			}))

			r.GET("/test", tt.handler)

			req := httptest.NewRequest("GET", "/test", http.NoBody)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			output := buf.String()

			if !strings.Contains(output, "level="+tt.expectedLevel) {
				t.Errorf("expected log level %s, got: %s", tt.expectedLevel, output)
			}

			if w.Code != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, w.Code)
			}
		})
	}
}

// TestLogger_ErrorLogging tests error logging.
func TestLogger_ErrorLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultLogger(&buf)

	r := fursy.New()
	r.Use(LoggerWithConfig(LoggerConfig{
		Logger: logger,
	}))

	testErr := errors.New("test error")
	r.GET("/error", func(_ *fursy.Context) error {
		return testErr
	})

	req := httptest.NewRequest("GET", "/error", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()

	if !strings.Contains(output, "error=\"test error\"") {
		t.Errorf("log should contain error message, got: %s", output)
	}

	if !strings.Contains(output, "level=ERROR") {
		t.Error("error should be logged at ERROR level")
	}
}

// TestLogger_BytesWritten tests tracking bytes written.
func TestLogger_BytesWritten(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultLogger(&buf)

	r := fursy.New()
	r.Use(LoggerWithConfig(LoggerConfig{
		Logger: logger,
	}))

	responseBody := "This is a test response body with some content"
	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, responseBody)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()

	expectedBytes := len(responseBody)

	if !strings.Contains(output, "bytes=") {
		t.Errorf("log should contain bytes written, got: %s", output)
	}

	// Verify actual bytes written
	if w.Body.Len() != expectedBytes {
		t.Errorf("expected %d bytes written, got %d", expectedBytes, w.Body.Len())
	}
}

// TestLogger_Latency tests latency measurement.
func TestLogger_Latency(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultLogger(&buf)

	r := fursy.New()
	r.Use(LoggerWithConfig(LoggerConfig{
		Logger: logger,
	}))

	r.GET("/test", func(c *fursy.Context) error {
		// Simulate some processing time
		// (In real tests, avoid time.Sleep)
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()

	if !strings.Contains(output, "latency_ms=") {
		t.Errorf("log should contain latency measurement, got: %s", output)
	}
}

// TestGetClientIP tests client IP extraction.
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		setupReq   func(*http.Request)
		expectedIP string
	}{
		{
			name: "X-Real-IP header",
			setupReq: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "1.2.3.4")
			},
			expectedIP: "1.2.3.4",
		},
		{
			name: "X-Forwarded-For single IP",
			setupReq: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "5.6.7.8")
			},
			expectedIP: "5.6.7.8",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			setupReq: func(r *http.Request) {
				r.Header.Set("X-Forwarded-For", "9.10.11.12, 13.14.15.16")
			},
			expectedIP: "9.10.11.12",
		},
		{
			name: "RemoteAddr with port",
			setupReq: func(r *http.Request) {
				r.RemoteAddr = "17.18.19.20:54321"
			},
			expectedIP: "17.18.19.20",
		},
		{
			name: "X-Real-IP takes precedence",
			setupReq: func(r *http.Request) {
				r.Header.Set("X-Real-IP", "priority.ip")
				r.Header.Set("X-Forwarded-For", "fallback.ip")
				r.RemoteAddr = "final.fallback:8080"
			},
			expectedIP: "priority.ip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", http.NoBody)
			tt.setupReq(req)

			ip := getClientIP(req)

			if ip != tt.expectedIP {
				t.Errorf("expected IP %s, got %s", tt.expectedIP, ip)
			}
		})
	}
}

// TestCleanIP tests IP cleaning function.
func TestCleanIP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"192.168.1.1", "192.168.1.1"},
		{"192.168.1.1:8080", "192.168.1.1"},
		{"  192.168.1.1  ", "192.168.1.1"},
		{"  192.168.1.1:8080  ", "192.168.1.1"},
		{"[2001:db8::1]", "2001:db8::1"},
		{"2001:db8::1", "2001:db8::1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cleanIP(tt.input)
			if result != tt.expected {
				t.Errorf("cleanIP(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestLogResponseWriter tests the response writer wrapper.
func TestLogResponseWriter(t *testing.T) {
	t.Run("captures status code", func(t *testing.T) {
		w := httptest.NewRecorder()
		lrw := &logResponseWriter{ResponseWriter: w}

		lrw.WriteHeader(404)

		if lrw.statusCode != 404 {
			t.Errorf("expected status 404, got %d", lrw.statusCode)
		}
	})

	t.Run("captures bytes written", func(t *testing.T) {
		w := httptest.NewRecorder()
		lrw := &logResponseWriter{ResponseWriter: w}

		data := []byte("test response body")
		n, err := lrw.Write(data)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if n != len(data) {
			t.Errorf("expected %d bytes written, got %d", len(data), n)
		}

		if lrw.bytesWritten != int64(len(data)) {
			t.Errorf("expected bytesWritten %d, got %d", len(data), lrw.bytesWritten)
		}
	})

	t.Run("defaults to 200 if WriteHeader not called", func(t *testing.T) {
		w := httptest.NewRecorder()
		lrw := &logResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		lrw.Write([]byte("body"))

		if lrw.statusCode != 200 {
			t.Errorf("expected default status 200, got %d", lrw.statusCode)
		}
	})

	t.Run("Unwrap returns original ResponseWriter", func(t *testing.T) {
		w := httptest.NewRecorder()
		lrw := &logResponseWriter{ResponseWriter: w}

		unwrapped := lrw.Unwrap()

		if unwrapped != w {
			t.Error("Unwrap() should return original ResponseWriter")
		}
	})
}

// TestLogger_IntegrationWithGroups tests logger with route groups.
func TestLogger_IntegrationWithGroups(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultLogger(&buf)

	r := fursy.New()
	r.Use(LoggerWithConfig(LoggerConfig{
		Logger: logger,
	}))

	api := r.Group("/api")
	v1 := api.Group("/v1")
	v1.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "users")
	})

	req := httptest.NewRequest("GET", "/api/v1/users", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()

	if !strings.Contains(output, "path=/api/v1/users") {
		t.Errorf("log should contain full path with group prefix, got: %s", output)
	}
}
