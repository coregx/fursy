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

// TestRecovery tests the default Recovery middleware.
func TestRecovery(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisablePrintStack: true, // Disable stderr output in tests
	}))

	r.GET("/panic", func(_ *fursy.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return 500.
	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	// Should write error message.
	if !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Errorf("expected error message, got: %s", w.Body.String())
	}

	// Should log panic.
	output := buf.String()
	if !strings.Contains(output, "Panic recovered") {
		t.Errorf("log should contain 'Panic recovered', got: %s", output)
	}
	if !strings.Contains(output, "test panic") {
		t.Errorf("log should contain panic message, got: %s", output)
	}
	if !strings.Contains(output, "stack=") {
		t.Errorf("log should contain stack trace, got: %s", output)
	}
}

// TestRecovery_ErrorPanic tests panic with error type.
func TestRecovery_ErrorPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisablePrintStack: true,
	}))

	testErr := errors.New("test error")
	r.GET("/panic", func(_ *fursy.Context) error {
		panic(testErr)
	})

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Errorf("log should contain error message, got: %s", output)
	}
}

// TestRecovery_IntPanic tests panic with non-error type.
func TestRecovery_IntPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisablePrintStack: true,
	}))

	r.GET("/panic", func(_ *fursy.Context) error {
		panic(42)
	})

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	output := buf.String()
	if !strings.Contains(output, "42") {
		t.Errorf("log should contain panic value, got: %s", output)
	}
}

// TestRecovery_NoStackTrace tests Recovery with stack trace disabled.
func TestRecovery_NoStackTrace(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisableStackTrace: true,
		DisablePrintStack: true,
	}))

	r.GET("/panic", func(_ *fursy.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	output := buf.String()
	if strings.Contains(output, "stack=") {
		t.Errorf("log should not contain stack trace when disabled, got: %s", output)
	}
	if !strings.Contains(output, "test panic") {
		t.Errorf("log should still contain panic message, got: %s", output)
	}
}

// TestRecovery_NoPanic tests normal request without panic.
func TestRecovery_NoPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisablePrintStack: true,
	}))

	r.GET("/normal", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/normal", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("expected 'OK', got %s", w.Body.String())
	}

	// Should not log anything.
	if buf.Len() > 0 {
		t.Errorf("should not log for normal requests, got: %s", buf.String())
	}
}

// TestRecovery_JSONFormat tests JSON logger format.
func TestRecovery_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := JSONRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisablePrintStack: true,
	}))

	r.GET("/panic", func(_ *fursy.Context) error {
		panic("json panic")
	})

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	output := buf.String()

	// Check JSON format.
	if !strings.Contains(output, `"msg":"Panic recovered"`) {
		t.Errorf("JSON log should contain msg field, got: %s", output)
	}
	if !strings.Contains(output, `"panic":"json panic"`) {
		t.Errorf("JSON log should contain panic field, got: %s", output)
	}
	if !strings.Contains(output, `"method":"GET"`) {
		t.Errorf("JSON log should contain method field, got: %s", output)
	}
	if !strings.Contains(output, `"path":"/panic"`) {
		t.Errorf("JSON log should contain path field, got: %s", output)
	}
}

// TestRecovery_StackTraceSize tests custom stack trace size.
func TestRecovery_StackTraceSize(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		StackTraceSize:    512, // Small size
		DisablePrintStack: true,
	}))

	r.GET("/panic", func(_ *fursy.Context) error {
		panic("stack test")
	})

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	// Should have stack trace (though truncated).
	output := buf.String()
	if !strings.Contains(output, "stack=") {
		t.Errorf("log should contain stack trace, got: %s", output)
	}
}

// TestPanicHandler tests the simplified PanicHandler.
func TestPanicHandler(t *testing.T) {
	r := fursy.New()
	r.Use(PanicHandler())

	r.GET("/panic", func(_ *fursy.Context) error {
		panic("simple panic")
	})

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Errorf("expected error message, got: %s", w.Body.String())
	}
}

// TestPanicHandler_NoPanic tests PanicHandler with normal request.
func TestPanicHandler_NoPanic(t *testing.T) {
	r := fursy.New()
	r.Use(PanicHandler())

	r.GET("/normal", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/normal", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("expected 'OK', got %s", w.Body.String())
	}
}

// TestRecovery_WithRouteGroups tests Recovery with route groups.
func TestRecovery_WithRouteGroups(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisablePrintStack: true,
	}))

	api := r.Group("/api")
	api.GET("/users", func(_ *fursy.Context) error {
		panic("group panic")
	})

	req := httptest.NewRequest("GET", "/api/users", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	output := buf.String()
	if !strings.Contains(output, "group panic") {
		t.Errorf("log should contain panic message, got: %s", output)
	}
	if !strings.Contains(output, "path=/api/users") {
		t.Errorf("log should contain full path with group prefix, got: %s", output)
	}
}

// TestRecovery_MiddlewareChain tests Recovery in middleware chain.
func TestRecovery_MiddlewareChain(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	var executed []string

	r := fursy.New()

	// First middleware - should execute.
	r.Use(func(c *fursy.Context) error {
		executed = append(executed, "before")
		err := c.Next()
		executed = append(executed, "after")
		return err
	})

	// Recovery middleware.
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisablePrintStack: true,
	}))

	r.GET("/panic", func(_ *fursy.Context) error {
		executed = append(executed, "handler")
		panic("chain panic")
	})

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should execute before, handler (panic), then after.
	expected := []string{"before", "handler", "after"}
	if len(executed) != len(expected) {
		t.Fatalf("expected %d executions, got %d: %v", len(expected), len(executed), executed)
	}

	for i, exp := range expected {
		if executed[i] != exp {
			t.Errorf("step %d: expected %s, got %s", i, exp, executed[i])
		}
	}

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// TestRecovery_CustomType tests panic with custom error type.
func TestRecovery_CustomType(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisablePrintStack: true,
	}))

	type customError struct {
		code    int
		message string
	}

	r.GET("/panic", func(_ *fursy.Context) error {
		panic(customError{code: 999, message: "custom error"})
	})

	req := httptest.NewRequest("GET", "/panic", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	output := buf.String()
	// Should contain string representation of custom type.
	if !strings.Contains(output, "custom error") || !strings.Contains(output, "999") {
		t.Errorf("log should contain custom error details, got: %s", output)
	}
}

// TestRecovery_MultipleRequests tests Recovery handles multiple requests correctly.
func TestRecovery_MultipleRequests(t *testing.T) {
	var buf bytes.Buffer
	logger := DefaultRecoveryLogger(&buf)

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{
		Logger:            logger,
		DisablePrintStack: true,
	}))

	r.GET("/panic", func(_ *fursy.Context) error {
		panic("panic request")
	})

	r.GET("/normal", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// First request - panic.
	req1 := httptest.NewRequest("GET", "/panic", http.NoBody)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != 500 {
		t.Errorf("first request: expected status 500, got %d", w1.Code)
	}

	// Second request - normal (should not be affected).
	req2 := httptest.NewRequest("GET", "/normal", http.NoBody)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Errorf("second request: expected status 200, got %d", w2.Code)
	}

	if w2.Body.String() != "OK" {
		t.Errorf("second request: expected 'OK', got %s", w2.Body.String())
	}

	// Third request - panic again.
	req3 := httptest.NewRequest("GET", "/panic", http.NoBody)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	if w3.Code != 500 {
		t.Errorf("third request: expected status 500, got %d", w3.Code)
	}
}
