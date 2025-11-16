// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coregx/fursy"
)

// TestCircuitBreaker_DefaultConsecutiveFailures tests default consecutive failures threshold.
func TestCircuitBreaker_DefaultConsecutiveFailures(t *testing.T) {
	router := fursy.New()

	var failCount int32
	cb := CircuitBreakerWithConfig(CircuitBreakerConfig{
		ConsecutiveFailures: 3, // Trip after 3 consecutive failures
		Timeout:             1 * time.Second,
	})
	router.Use(cb)

	router.GET("/test", func(c *fursy.Context) error {
		count := atomic.AddInt32(&failCount, 1)
		if count <= 3 {
			return errors.New("simulated failure")
		}
		return c.String(http.StatusOK, "OK")
	})

	// First 3 requests should fail (circuit still closed).
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		// Handler errors, but circuit still allows requests
	}

	// 4th request should be blocked (circuit open).
	req4 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec4 := httptest.NewRecorder()
	router.ServeHTTP(rec4, req4)

	if rec4.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected circuit open (503), got %d", rec4.Code)
	}

	if !contains(rec4.Body.String(), "circuit breaker open") {
		t.Errorf("Expected circuit breaker error message, got %s", rec4.Body.String())
	}
}

// TestCircuitBreaker_StateTransitions tests Closed → Open → Half-Open → Closed transitions.
func TestCircuitBreaker_StateTransitions(t *testing.T) {
	router := fursy.New()

	var failCount int32
	var states []State
	var mu sync.Mutex

	cb := CircuitBreakerWithConfig(CircuitBreakerConfig{
		ConsecutiveFailures: 2,
		Timeout:             100 * time.Millisecond,
		MaxRequests:         1,
		OnStateChange: func(_ /* from */, to State) {
			mu.Lock()
			states = append(states, to)
			mu.Unlock()
		},
	})
	router.Use(cb)

	router.GET("/test", func(c *fursy.Context) error {
		count := atomic.LoadInt32(&failCount)
		if count < 2 {
			atomic.AddInt32(&failCount, 1)
			return errors.New("fail")
		}
		// Success after 2 failures
		return c.String(http.StatusOK, "OK")
	})

	// 1. Closed state: 2 failures → Open
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	// Verify Open state
	req3 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec3 := httptest.NewRecorder()
	router.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected circuit open, got %d", rec3.Code)
	}

	// 2. Wait for timeout → Half-Open
	time.Sleep(150 * time.Millisecond)

	// 3. Half-Open state: success → Closed
	req4 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec4 := httptest.NewRecorder()
	router.ServeHTTP(rec4, req4)

	if rec4.Code != http.StatusOK {
		t.Errorf("Expected success in half-open, got %d", rec4.Code)
	}

	// Verify state transitions
	time.Sleep(50 * time.Millisecond) // Wait for callbacks
	mu.Lock()
	defer mu.Unlock()

	if len(states) < 2 {
		t.Fatalf("Expected at least 2 state transitions, got %d: %v", len(states), states)
	}

	if states[0] != StateOpen {
		t.Errorf("First transition should be to Open, got %s", states[0])
	}

	// Note: Half-Open transition happens before request, Closed after success
	hasHalfOpen := false
	hasClosed := false
	for _, s := range states {
		if s == StateHalfOpen {
			hasHalfOpen = true
		}
		if s == StateClosed {
			hasClosed = true
		}
	}

	if !hasHalfOpen {
		t.Error("Expected Half-Open state transition")
	}
	if !hasClosed {
		t.Error("Expected Closed state transition")
	}
}

// TestCircuitBreaker_WindowBasedRatio tests count-based window threshold.
func TestCircuitBreaker_WindowBasedRatio(t *testing.T) {
	router := fursy.New()

	var requestCount int32

	cb := CircuitBreakerWithConfig(CircuitBreakerConfig{
		FailureThreshold: 3, // 3 failures
		RequestWindow:    5, // out of 5 requests (60% failure rate)
		Timeout:          1 * time.Second,
	})
	router.Use(cb)

	router.GET("/test", func(c *fursy.Context) error {
		count := atomic.AddInt32(&requestCount, 1)
		// Fail requests 1, 3, 5 (3 failures out of 5)
		if count == 1 || count == 3 || count == 5 {
			return errors.New("fail")
		}
		return c.String(http.StatusOK, "OK")
	})

	// Send 5 requests (3 failures, 2 successes)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	// 6th request should be blocked (circuit open after 5th request failed)
	req6 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec6 := httptest.NewRecorder()
	router.ServeHTTP(rec6, req6)

	if rec6.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected circuit open (503) after ratio threshold, got %d", rec6.Code)
	}
}

// TestCircuitBreaker_TimeBasedWindow tests time-based window threshold.
func TestCircuitBreaker_TimeBasedWindow(t *testing.T) {
	router := fursy.New()

	cb := CircuitBreakerWithConfig(CircuitBreakerConfig{
		FailureThreshold: 3,
		TimeWindow:       500 * time.Millisecond,
		Timeout:          1 * time.Second,
	})
	router.Use(cb)

	router.GET("/test", func(_ *fursy.Context) error {
		return errors.New("fail")
	})

	// Send 3 failures within 500ms
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		time.Sleep(50 * time.Millisecond)
	}

	// Next request should be blocked (circuit open)
	req4 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec4 := httptest.NewRecorder()
	router.ServeHTTP(rec4, req4)

	if rec4.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected circuit open after time window threshold, got %d", rec4.Code)
	}
}

// TestCircuitBreaker_HalfOpenMaxRequests tests MaxRequests limit in Half-Open state.
func TestCircuitBreaker_HalfOpenMaxRequests(t *testing.T) {
	router := fursy.New()

	var requestCount int32

	cb := CircuitBreakerWithConfig(CircuitBreakerConfig{
		ConsecutiveFailures: 2,
		Timeout:             100 * time.Millisecond,
		MaxRequests:         2, // Allow 2 requests in Half-Open
	})
	router.Use(cb)

	router.GET("/test", func(c *fursy.Context) error {
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 2 {
			return errors.New("fail")
		}
		return c.String(http.StatusOK, "OK")
	})

	// Trigger circuit open (2 failures)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	// Wait for timeout → Half-Open
	time.Sleep(150 * time.Millisecond)

	// Send MaxRequests (2) successful requests in Half-Open
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Request %d in half-open should succeed, got %d", i+1, rec.Code)
		}
	}

	// Circuit should now be Closed, next request should succeed
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected success after half-open recovery, got %d", rec.Code)
	}
}

// TestCircuitBreaker_CustomReadyToTrip tests custom ReadyToTrip callback.
func TestCircuitBreaker_CustomReadyToTrip(t *testing.T) {
	router := fursy.New()

	var failCount int32

	cb := CircuitBreakerWithConfig(CircuitBreakerConfig{
		ReadyToTrip: func(counts Counts) bool {
			// Custom logic: trip if failure rate > 70%
			if counts.Requests < 10 {
				return false
			}
			failureRate := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRate > 0.7
		},
		Timeout: 1 * time.Second,
	})
	router.Use(cb)

	router.GET("/test", func(c *fursy.Context) error {
		count := atomic.AddInt32(&failCount, 1)
		// 8 failures out of 10 requests (80% failure rate)
		if count <= 8 {
			return errors.New("fail")
		}
		return c.String(http.StatusOK, "OK")
	})

	// Send 10 requests (8 failures, 2 successes)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	// 11th request should be blocked (circuit open after 8th failure exceeded 70% threshold)
	req11 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec11 := httptest.NewRecorder()
	router.ServeHTTP(rec11, req11)

	if rec11.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected circuit open with custom ReadyToTrip, got %d", rec11.Code)
	}
}

// TestCircuitBreaker_CustomIsSuccessful tests custom IsSuccessful callback.
func TestCircuitBreaker_CustomIsSuccessful(t *testing.T) {
	router := fursy.New()

	cb := CircuitBreakerWithConfig(CircuitBreakerConfig{
		ConsecutiveFailures: 2,
		IsSuccessful: func(c *fursy.Context) bool {
			// Custom logic: 4xx errors are also failures
			if rw, ok := c.Response.(interface{ Status() int }); ok {
				status := rw.Status()
				return status >= 200 && status < 400
			}
			return true
		},
		Timeout: 1 * time.Second,
	})
	router.Use(cb)

	router.GET("/test", func(_ *fursy.Context) error {
		// Return 400 (will be considered failure with custom IsSuccessful)
		return errors.New("bad request")
	})

	// Send 2 requests with 400 status
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	// 3rd request should be blocked (circuit open)
	req3 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec3 := httptest.NewRecorder()
	router.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected circuit open with custom IsSuccessful, got %d", rec3.Code)
	}
}

// TestCircuitBreaker_CustomErrorHandler tests custom error handler.
func TestCircuitBreaker_CustomErrorHandler(t *testing.T) {
	router := fursy.New()

	cb := CircuitBreakerWithConfig(CircuitBreakerConfig{
		ConsecutiveFailures: 1,
		ErrorHandler: func(c *fursy.Context) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error":       "circuit_breaker_open",
				"retry_after": "60s",
			})
		},
		Timeout: 1 * time.Second,
	})
	router.Use(cb)

	router.GET("/test", func(_ *fursy.Context) error {
		return errors.New("fail")
	})

	// Trigger circuit open
	req1 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)

	// Next request should use custom error handler
	req2 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected custom error status 429, got %d", rec2.Code)
	}

	if !contains(rec2.Body.String(), "circuit_breaker_open") {
		t.Errorf("Expected custom error message, got %s", rec2.Body.String())
	}
}

// TestCircuitBreaker_Skipper tests the Skipper functionality.
func TestCircuitBreaker_Skipper(t *testing.T) {
	router := fursy.New()

	cb := CircuitBreakerWithConfig(CircuitBreakerConfig{
		ConsecutiveFailures: 1,
		Skipper: func(c *fursy.Context) bool {
			// Skip circuit breaker for /health endpoint
			return c.Request.URL.Path == "/health"
		},
		Timeout: 1 * time.Second,
	})
	router.Use(cb)

	router.GET("/test", func(_ *fursy.Context) error {
		return errors.New("fail")
	})

	router.GET("/health", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "healthy")
	})

	// Trigger circuit open on /test
	req1 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)

	// /test should be blocked
	req2 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected circuit open on /test, got %d", rec2.Code)
	}

	// /health should still work (skipped)
	req3 := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	rec3 := httptest.NewRecorder()
	router.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusOK {
		t.Errorf("Expected /health to work (skipped), got %d", rec3.Code)
	}
}

// TestCircuitBreaker_HelperFunctions tests helper constructors.
func TestCircuitBreaker_HelperFunctions(t *testing.T) {
	t.Run("CircuitBreakerConsecutive", func(t *testing.T) {
		router := fursy.New()
		router.Use(CircuitBreakerConsecutive(2, 1*time.Second))

		router.GET("/test", func(_ *fursy.Context) error {
			return errors.New("fail")
		})

		// 2 failures → circuit open
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
		}

		req3 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec3 := httptest.NewRecorder()
		router.ServeHTTP(rec3, req3)

		if rec3.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected circuit open, got %d", rec3.Code)
		}
	})

	t.Run("CircuitBreakerRatio", func(t *testing.T) {
		router := fursy.New()
		router.Use(CircuitBreakerRatio(2, 3, 1*time.Second)) // 2 failures out of 3 requests

		var count int32
		router.GET("/test", func(c *fursy.Context) error {
			n := atomic.AddInt32(&count, 1)
			if n <= 2 {
				return errors.New("fail")
			}
			return c.String(http.StatusOK, "OK")
		})

		// 3 requests (2 failures, 1 success) → circuit open
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
		}

		req4 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec4 := httptest.NewRecorder()
		router.ServeHTTP(rec4, req4)

		if rec4.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected circuit open with ratio, got %d", rec4.Code)
		}
	})

	t.Run("CircuitBreakerTimeWindow", func(t *testing.T) {
		router := fursy.New()
		router.Use(CircuitBreakerTimeWindow(2, 500*time.Millisecond, 1*time.Second))

		router.GET("/test", func(_ *fursy.Context) error {
			return errors.New("fail")
		})

		// 2 failures within 500ms → circuit open
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			time.Sleep(50 * time.Millisecond)
		}

		req3 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec3 := httptest.NewRecorder()
		router.ServeHTTP(rec3, req3)

		if rec3.Code != http.StatusServiceUnavailable {
			t.Errorf("Expected circuit open with time window, got %d", rec3.Code)
		}
	})
}

// TestCircuitBreaker_ConcurrentRequests tests thread-safety.
func TestCircuitBreaker_ConcurrentRequests(t *testing.T) {
	router := fursy.New()

	router.Use(CircuitBreakerWithConfig(CircuitBreakerConfig{
		ConsecutiveFailures: 10,
		Timeout:             1 * time.Second,
	}))

	router.GET("/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Send 100 concurrent requests
	var wg sync.WaitGroup
	successCount := atomic.Int32{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				successCount.Add(1)
			}
		}()
	}

	wg.Wait()

	if successCount.Load() != 100 {
		t.Errorf("Expected 100 successful requests, got %d", successCount.Load())
	}
}

// TestCircuitBreaker_StateString tests State.String() method.
func TestCircuitBreaker_StateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("State.String() = %s, want %s", got, tt.expected)
			}
		})
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
