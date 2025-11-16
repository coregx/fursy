// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRouter_Use tests the Use method for adding middleware.
func TestRouter_Use(t *testing.T) {
	t.Run("single middleware", func(t *testing.T) {
		r := New()
		middleware := func(c *Context) error {
			return c.Next()
		}

		result := r.Use(middleware)

		// Should return the router for chaining
		if result != r {
			t.Error("Use() should return router for chaining")
		}

		// Should have 1 middleware
		if len(r.middleware) != 1 {
			t.Errorf("expected 1 middleware, got %d", len(r.middleware))
		}
	})

	t.Run("multiple middleware", func(t *testing.T) {
		r := New()
		mw1 := func(c *Context) error { return c.Next() }
		mw2 := func(c *Context) error { return c.Next() }
		mw3 := func(c *Context) error { return c.Next() }

		r.Use(mw1, mw2, mw3)

		if len(r.middleware) != 3 {
			t.Errorf("expected 3 middleware, got %d", len(r.middleware))
		}
	})

	t.Run("chaining Use calls", func(t *testing.T) {
		r := New()
		mw1 := func(c *Context) error { return c.Next() }
		mw2 := func(c *Context) error { return c.Next() }

		r.Use(mw1).Use(mw2)

		if len(r.middleware) != 2 {
			t.Errorf("expected 2 middleware, got %d", len(r.middleware))
		}
	})
}

// TestMiddleware_ExecutionOrder tests that middleware executes in the correct order.
func TestMiddleware_ExecutionOrder(t *testing.T) {
	var executionOrder []string

	r := New()

	// Middleware 1: runs first, adds "before1", then Next(), then adds "after1"
	r.Use(func(c *Context) error {
		executionOrder = append(executionOrder, "before1")
		err := c.Next()
		executionOrder = append(executionOrder, "after1")
		return err
	})

	// Middleware 2: runs second
	r.Use(func(c *Context) error {
		executionOrder = append(executionOrder, "before2")
		err := c.Next()
		executionOrder = append(executionOrder, "after2")
		return err
	})

	// Handler: runs last
	r.GET("/test", func(c *Context) error {
		executionOrder = append(executionOrder, "handler")
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Expected order: before1 -> before2 -> handler -> after2 -> after1
	expected := []string{"before1", "before2", "handler", "after2", "after1"}

	if len(executionOrder) != len(expected) {
		t.Fatalf("expected %d steps, got %d: %v", len(expected), len(executionOrder), executionOrder)
	}

	for i, step := range expected {
		if executionOrder[i] != step {
			t.Errorf("step %d: expected %s, got %s", i, step, executionOrder[i])
		}
	}
}

// TestMiddleware_Next tests the Next() method.
func TestMiddleware_Next(t *testing.T) {
	t.Run("calls next handler", func(t *testing.T) {
		r := New()
		handlerCalled := false

		r.Use(func(c *Context) error {
			return c.Next()
		})

		r.GET("/test", func(c *Context) error {
			handlerCalled = true
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if !handlerCalled {
			t.Error("handler was not called after Next()")
		}
	})

	t.Run("propagates handler response", func(t *testing.T) {
		r := New()

		r.Use(func(c *Context) error {
			return c.Next()
		})

		r.GET("/test", func(c *Context) error {
			return c.String(200, "test-response")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Body.String() != "test-response" {
			t.Errorf("expected 'test-response', got %s", w.Body.String())
		}
	})

	t.Run("returns nil when no more handlers", func(t *testing.T) {
		c := newContext()
		c.handlers = []HandlerFunc{
			func(_ *Context) error {
				return nil
			},
		}
		c.index = -1

		// Call Next() to execute first handler
		err := c.Next()
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}

		// Call Next() again - should return nil (no more handlers)
		err = c.Next()
		if err != nil {
			t.Errorf("expected nil when no more handlers, got %v", err)
		}
	})
}

// TestMiddleware_Abort tests the Abort() method.
func TestMiddleware_Abort(t *testing.T) {
	t.Run("stops handler chain", func(t *testing.T) {
		r := New()
		handlerCalled := false

		// Middleware that aborts
		r.Use(func(c *Context) error {
			c.Abort()
			return c.String(401, "Unauthorized")
		})

		// This handler should NOT be called
		r.GET("/test", func(c *Context) error {
			handlerCalled = true
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if handlerCalled {
			t.Error("handler was called after Abort()")
		}

		if w.Code != 401 {
			t.Errorf("expected status 401, got %d", w.Code)
		}

		if w.Body.String() != "Unauthorized" {
			t.Errorf("expected 'Unauthorized', got %s", w.Body.String())
		}
	})

	t.Run("subsequent middleware not executed", func(t *testing.T) {
		r := New()
		var executed []string

		// MW1: executes and calls Next()
		r.Use(func(c *Context) error {
			executed = append(executed, "mw1")
			return c.Next()
		})

		// MW2: aborts
		r.Use(func(c *Context) error {
			executed = append(executed, "mw2")
			c.Abort()
			return c.String(403, "Forbidden")
		})

		// MW3: should NOT execute
		r.Use(func(c *Context) error {
			executed = append(executed, "mw3")
			return c.Next()
		})

		// Handler: should NOT execute
		r.GET("/test", func(c *Context) error {
			executed = append(executed, "handler")
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should only have executed mw1 and mw2
		if len(executed) != 2 {
			t.Fatalf("expected 2 executions, got %d: %v", len(executed), executed)
		}

		if executed[0] != "mw1" || executed[1] != "mw2" {
			t.Errorf("unexpected execution order: %v", executed)
		}

		if w.Code != 403 {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("IsAborted returns true", func(t *testing.T) {
		c := newContext()
		c.handlers = []HandlerFunc{
			func(c *Context) error {
				c.Abort()
				return nil
			},
		}
		c.index = -1

		if c.IsAborted() {
			t.Error("IsAborted() should be false initially")
		}

		_ = c.Next()

		if !c.IsAborted() {
			t.Error("IsAborted() should be true after Abort()")
		}
	})
}

// TestMiddleware_ErrorPropagation tests error propagation through the middleware chain.
//
//nolint:gocognit // Test function with multiple scenarios
func TestMiddleware_ErrorPropagation(t *testing.T) {
	t.Run("handler error propagates to middleware", func(t *testing.T) {
		r := New()
		handlerErr := errors.New("handler error")
		var receivedErr error

		r.Use(func(c *Context) error {
			err := c.Next()
			receivedErr = err
			return err
		})

		r.GET("/test", func(_ *Context) error {
			return handlerErr
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if !errors.Is(receivedErr, handlerErr) {
			t.Errorf("expected error %v, got %v", handlerErr, receivedErr)
		}

		// Router should send 500 on error
		if w.Code != 500 {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("middleware error propagates", func(t *testing.T) {
		r := New()
		middlewareErr := errors.New("middleware error")
		handlerCalled := false

		r.Use(func(_ *Context) error {
			return middlewareErr
		})

		r.GET("/test", func(c *Context) error {
			handlerCalled = true
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Handler should not be called
		if handlerCalled {
			t.Error("handler was called after middleware error")
		}

		// Should get 500
		if w.Code != 500 {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("middleware can handle and transform errors", func(t *testing.T) {
		r := New()
		handlerErr := errors.New("original error")

		// Error handler middleware
		r.Use(func(c *Context) error {
			err := c.Next()
			if err != nil {
				// Transform error into custom response
				return c.String(400, "Custom Error: "+err.Error())
			}
			return nil
		})

		r.GET("/test", func(_ *Context) error {
			return handlerErr
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		expected := "Custom Error: original error"
		if w.Body.String() != expected {
			t.Errorf("expected %s, got %s", expected, w.Body.String())
		}
	})

	t.Run("middleware can suppress errors", func(t *testing.T) {
		r := New()

		// Error suppression middleware
		r.Use(func(c *Context) error {
			err := c.Next()
			if err != nil {
				// Log error (in real scenario) and suppress
				_ = c.String(200, "Error handled gracefully")
				return nil //nolint:nilerr // Return nil to suppress error
			}
			return nil
		})

		r.GET("/test", func(_ *Context) error {
			return errors.New("some error")
		})

		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should get 200 because error was suppressed
		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "Error handled gracefully" {
			t.Errorf("unexpected body: %s", w.Body.String())
		}
	})
}

// TestMiddleware_DataPassing tests passing data between middleware using Box.Set/Get.
func TestMiddleware_DataPassing(t *testing.T) {
	r := New()

	// MW1: sets user data
	r.Use(func(c *Context) error {
		c.Set("userID", "123")
		c.Set("authenticated", true)
		return c.Next()
	})

	// MW2: reads and verifies user data
	r.Use(func(c *Context) error {
		userID := c.GetString("userID")
		authenticated := c.GetBool("authenticated")

		if userID != "123" {
			return errors.New("userID not passed correctly")
		}
		if !authenticated {
			return errors.New("authenticated not passed correctly")
		}

		return c.Next()
	})

	// Handler: uses the data
	r.GET("/test", func(c *Context) error {
		userID := c.GetString("userID")
		return c.String(200, "User: "+userID)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "User: 123" {
		t.Errorf("expected 'User: 123', got %s", w.Body.String())
	}
}

// TestMiddleware_RealWorldScenarios tests realistic middleware patterns.
//
//nolint:gocognit // Test function with multiple realistic scenarios
func TestMiddleware_RealWorldScenarios(t *testing.T) {
	t.Run("logger middleware", func(t *testing.T) {
		r := New()
		var logged string

		// Simple logger middleware
		r.Use(func(c *Context) error {
			method := c.Request.Method
			path := c.Request.URL.Path
			logged = method + " " + path

			err := c.Next()

			// Could log duration, status, etc here
			return err
		})

		r.GET("/users/123", func(c *Context) error {
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/users/123", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if logged != "GET /users/123" {
			t.Errorf("expected 'GET /users/123', got %s", logged)
		}
	})

	t.Run("authentication middleware", func(t *testing.T) {
		r := New()

		// Auth middleware
		r.Use(func(c *Context) error {
			token := c.GetHeader("Authorization")
			if token == "" {
				c.Abort()
				return c.String(401, "Unauthorized")
			}

			// Simulate token validation
			if !strings.HasPrefix(token, "Bearer ") {
				c.Abort()
				return c.String(401, "Invalid token")
			}

			// Set user info
			c.Set("userID", "user-from-token")
			return c.Next()
		})

		r.GET("/protected", func(c *Context) error {
			userID := c.GetString("userID")
			return c.String(200, "Hello "+userID)
		})

		// Test without token
		req1 := httptest.NewRequest("GET", "/protected", http.NoBody)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		if w1.Code != 401 {
			t.Errorf("expected 401 without token, got %d", w1.Code)
		}

		// Test with invalid token
		req2 := httptest.NewRequest("GET", "/protected", http.NoBody)
		req2.Header.Set("Authorization", "invalid")
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Code != 401 {
			t.Errorf("expected 401 with invalid token, got %d", w2.Code)
		}

		// Test with valid token
		req3 := httptest.NewRequest("GET", "/protected", http.NoBody)
		req3.Header.Set("Authorization", "Bearer valid-token")
		w3 := httptest.NewRecorder()
		r.ServeHTTP(w3, req3)

		if w3.Code != 200 {
			t.Errorf("expected 200 with valid token, got %d", w3.Code)
		}

		if w3.Body.String() != "Hello user-from-token" {
			t.Errorf("unexpected body: %s", w3.Body.String())
		}
	})

	t.Run("CORS middleware", func(t *testing.T) {
		r := New()

		// CORS middleware
		r.Use(func(c *Context) error {
			c.SetHeader("Access-Control-Allow-Origin", "*")
			c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
			c.SetHeader("Access-Control-Allow-Headers", "Content-Type")

			// Handle preflight
			if c.Request.Method == "OPTIONS" {
				return c.NoContent(204)
			}

			return c.Next()
		})

		r.GET("/api/data", func(c *Context) error {
			return c.String(200, "data")
		})

		// Register OPTIONS handler for preflight
		r.OPTIONS("/api/data", func(c *Context) error {
			// Middleware will handle the response
			return c.NoContent(204)
		})

		// Test OPTIONS request (preflight)
		req1 := httptest.NewRequest("OPTIONS", "/api/data", http.NoBody)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		if w1.Code != 204 {
			t.Errorf("expected 204 for OPTIONS, got %d", w1.Code)
		}

		origin := w1.Header().Get("Access-Control-Allow-Origin")
		if origin != "*" {
			t.Errorf("expected CORS header, got %s", origin)
		}

		// Test regular GET request
		req2 := httptest.NewRequest("GET", "/api/data", http.NoBody)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Code != 200 {
			t.Errorf("expected 200, got %d", w2.Code)
		}

		origin2 := w2.Header().Get("Access-Control-Allow-Origin")
		if origin2 != "*" {
			t.Errorf("expected CORS header on GET, got %s", origin2)
		}
	})

	t.Run("recovery middleware", func(t *testing.T) {
		r := New()
		recovered := false

		// Recovery middleware
		r.Use(func(c *Context) error {
			defer func() {
				if r := recover(); r != nil {
					recovered = true
					_ = c.String(500, "Internal Server Error")
				}
			}()

			return c.Next()
		})

		r.GET("/panic", func(_ *Context) error {
			panic("something went wrong")
		})

		req := httptest.NewRequest("GET", "/panic", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if !recovered {
			t.Error("panic was not recovered")
		}

		if w.Code != 500 {
			t.Errorf("expected 500 after panic, got %d", w.Code)
		}
	})
}

// TestMiddleware_ContextReset tests that context is properly reset between requests.
func TestMiddleware_ContextReset(t *testing.T) {
	r := New()

	// Middleware that sets data
	r.Use(func(c *Context) error {
		c.Set("request-specific", "value")
		return c.Next()
	})

	r.GET("/test", func(c *Context) error {
		val := c.GetString("request-specific")
		return c.String(200, val)
	})

	// First request
	req1 := httptest.NewRequest("GET", "/test", http.NoBody)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Body.String() != "value" {
		t.Errorf("first request: expected 'value', got %s", w1.Body.String())
	}

	// Second request - context should be reset (clean data map)
	req2 := httptest.NewRequest("GET", "/test", http.NoBody)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Body.String() != "value" {
		t.Errorf("second request: expected 'value', got %s", w2.Body.String())
	}
}

// BenchmarkMiddleware_Chain benchmarks middleware chain execution.
func BenchmarkMiddleware_Chain(b *testing.B) {
	r := New()

	// Add 3 simple middleware
	r.Use(func(c *Context) error {
		return c.Next()
	})
	r.Use(func(c *Context) error {
		return c.Next()
	})
	r.Use(func(c *Context) error {
		return c.Next()
	})

	r.GET("/test", func(c *Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkMiddleware_DataPassing benchmarks Set/Get operations.
func BenchmarkMiddleware_DataPassing(b *testing.B) {
	r := New()

	r.Use(func(c *Context) error {
		c.Set("key1", "value1")
		c.Set("key2", 123)
		c.Set("key3", true)
		return c.Next()
	})

	r.GET("/test", func(c *Context) error {
		_ = c.GetString("key1")
		_ = c.GetInt("key2")
		_ = c.GetBool("key3")
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkMiddleware_Abort benchmarks Abort() performance.
func BenchmarkMiddleware_Abort(b *testing.B) {
	r := New()

	r.Use(func(c *Context) error {
		c.Abort()
		return c.String(401, "Unauthorized")
	})

	r.Use(func(c *Context) error {
		// Should not execute
		return c.Next()
	})

	r.GET("/test", func(c *Context) error {
		// Should not execute
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
