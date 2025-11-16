// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/coregx/fursy"
	"golang.org/x/time/rate"
)

func TestRateLimit_BasicLimiting(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimit(5, 10)) // 5 req/s, burst 10

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Make 10 requests (should all pass due to burst).
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("request %d: expected status 200, got %d", i+1, rec.Code)
		}
	}

	// 11th request should be rate limited.
	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 429 {
		t.Errorf("expected status 429 after burst, got %d", rec.Code)
	}

	// Check Retry-After header.
	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header on 429 response")
	}
}

func TestRateLimit_Headers(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:    10,
		Burst:   20,
		Headers: true,
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Check X-RateLimit-* headers.
	if rec.Header().Get("X-RateLimit-Limit") != "10" {
		t.Errorf("expected X-RateLimit-Limit=10, got %s", rec.Header().Get("X-RateLimit-Limit"))
	}

	remaining := rec.Header().Get("X-RateLimit-Remaining")
	if remaining == "" {
		t.Error("expected X-RateLimit-Remaining header")
	}

	reset := rec.Header().Get("X-RateLimit-Reset")
	if reset == "" {
		t.Error("expected X-RateLimit-Reset header")
	}

	// Verify reset is a unix timestamp in the future.
	resetTime, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		t.Errorf("X-RateLimit-Reset should be unix timestamp, got %s", reset)
	}

	if resetTime <= time.Now().Unix() {
		t.Error("X-RateLimit-Reset should be in the future")
	}
}

func TestRateLimit_PerIP(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:  2,
		Burst: 3,
		KeyFunc: func(c *fursy.Context) string {
			return getClientIP(c.Request)
		},
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// IP 1: Make 3 requests (burst).
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", http.NoBody)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("IP1 request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// IP 1: 4th request should be rate limited.
	req := httptest.NewRequest("GET", "/", http.NoBody)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 429 {
		t.Errorf("IP1 expected 429 after burst, got %d", rec.Code)
	}

	// IP 2: Should have separate limit.
	req = httptest.NewRequest("GET", "/", http.NoBody)
	req.RemoteAddr = "192.168.1.2:1234"
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("IP2 expected 200 (separate limit), got %d", rec.Code)
	}
}

func TestRateLimit_PerUser(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:  5,
		Burst: 10,
		KeyFunc: func(c *fursy.Context) string {
			// Simulate user-based limiting.
			return c.Request.Header.Get("X-User-ID")
		},
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// User 1: Make 10 requests (burst).
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", http.NoBody)
		req.Header.Set("X-User-ID", "user1")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("user1 request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// User 1: 11th request should be rate limited.
	req := httptest.NewRequest("GET", "/", http.NoBody)
	req.Header.Set("X-User-ID", "user1")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 429 {
		t.Errorf("user1 expected 429 after burst, got %d", rec.Code)
	}

	// User 2: Should have separate limit.
	req = httptest.NewRequest("GET", "/", http.NoBody)
	req.Header.Set("X-User-ID", "user2")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("user2 expected 200 (separate limit), got %d", rec.Code)
	}
}

func TestRateLimit_GlobalLimiter(t *testing.T) {
	// Create a global limiter (shared across all requests).
	limiter := rate.NewLimiter(5, 10)

	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Limiter: limiter,
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Make 10 requests from different IPs (should all share the same limiter).
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", http.NoBody)
		req.RemoteAddr = "192.168.1." + strconv.Itoa(i) + ":1234"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// 11th request should be rate limited (regardless of IP).
	req := httptest.NewRequest("GET", "/", http.NoBody)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 429 {
		t.Errorf("expected 429 for global limiter, got %d", rec.Code)
	}
}

func TestRateLimit_Skipper(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:  1,
		Burst: 1,
		Skipper: func(c *fursy.Context) bool {
			// Skip /health endpoint.
			return c.Request.URL.Path == "/health"
		},
	}))

	router.GET("/health", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	router.GET("/api", func(c *fursy.Context) error {
		return c.String(200, "API")
	})

	// /health should not be rate limited.
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/health", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("/health request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// /api should be rate limited after 1 request.
	req := httptest.NewRequest("GET", "/api", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("/api first request: expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest("GET", "/api", http.NoBody)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 429 {
		t.Errorf("/api second request: expected 429, got %d", rec.Code)
	}
}

func TestRateLimit_CustomErrorHandler(t *testing.T) {
	customErrorCalled := false

	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:  1,
		Burst: 1,
		ErrorHandler: func(c *fursy.Context, _ time.Duration) error {
			customErrorCalled = true
			return c.String(503, "Custom rate limit error")
		},
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// First request OK.
	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Second request should trigger custom error handler.
	req = httptest.NewRequest("GET", "/", http.NoBody)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if !customErrorCalled {
		t.Error("expected custom error handler to be called")
	}

	if rec.Code != 503 {
		t.Errorf("expected status 503 from custom error handler, got %d", rec.Code)
	}

	if rec.Body.String() != "Custom rate limit error" {
		t.Errorf("expected custom error message, got %q", rec.Body.String())
	}
}

func TestRateLimit_SuccessHandler(t *testing.T) {
	successHandlerCalled := false
	var remainingTokens int

	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:  10,
		Burst: 20,
		SuccessHandler: func(_ *fursy.Context, remaining int) error {
			successHandlerCalled = true
			remainingTokens = remaining
			return nil
		},
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if !successHandlerCalled {
		t.Error("expected success handler to be called")
	}

	if remainingTokens < 0 {
		t.Errorf("expected non-negative remaining tokens, got %d", remainingTokens)
	}
}

func TestRateLimit_SuccessHandler_ReturnsError(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:  10,
		Burst: 20,
		SuccessHandler: func(c *fursy.Context, _ int) error {
			// Return error from success handler.
			return c.String(500, "Success handler error")
		},
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 500 {
		t.Errorf("expected status 500 when success handler returns error, got %d", rec.Code)
	}
}

func TestRateLimit_TokenRefill(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimit(10, 2)) // 10 tokens/second, burst 2

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Consume burst (2 requests).
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// 3rd request should be rate limited.
	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 429 {
		t.Errorf("expected 429 after burst, got %d", rec.Code)
	}

	// Wait for token refill (200ms = 2 tokens at 10/sec).
	time.Sleep(200 * time.Millisecond)

	// Should be able to make 2 more requests.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("after refill request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestRateLimit_ConcurrentRequests(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimit(100, 200)) // High limits

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Make 100 concurrent requests.
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/", http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			mu.Lock()
			if rec.Code == 200 {
				successCount++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	// All 100 requests should succeed (within burst).
	if successCount != 100 {
		t.Errorf("expected 100 successful requests, got %d", successCount)
	}
}

func TestRateLimit_MaxKeys_LRU(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:    10,
		Burst:   20,
		MaxKeys: 5, // Small limit for testing
		KeyFunc: func(c *fursy.Context) string {
			return c.Request.Header.Get("X-Key")
		},
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Create 10 different keys (should evict oldest when > 5).
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", http.NoBody)
		req.Header.Set("X-Key", "key"+strconv.Itoa(i))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("key%d: expected 200, got %d", i, rec.Code)
		}
	}

	// Oldest keys should have been evicted, but we can't easily verify internal state.
	// This test ensures no panic/deadlock occurs during eviction.
}

func TestRateLimit_InMemoryStore(t *testing.T) {
	store := newInMemoryStore(100)

	// Get limiter for key1.
	limiter1 := store.GetLimiter("key1", 10, 20)
	if limiter1 == nil {
		t.Fatal("expected non-nil limiter")
	}

	// Get limiter for key1 again (should return same instance).
	limiter1Again := store.GetLimiter("key1", 10, 20)
	if limiter1 != limiter1Again {
		t.Error("expected same limiter instance for same key")
	}

	// Get limiter for key2 (should be different).
	limiter2 := store.GetLimiter("key2", 10, 20)
	if limiter1 == limiter2 {
		t.Error("expected different limiter instances for different keys")
	}
}

func TestRateLimit_Cleanup(t *testing.T) {
	store := newInMemoryStore(100)

	// Create limiters with old access times.
	store.GetLimiter("old1", 10, 20)
	store.GetLimiter("old2", 10, 20)

	// Manually set old access times.
	store.mu.Lock()
	for _, entry := range store.limiters {
		entry.lastAccess = time.Now().Add(-10 * time.Minute)
	}
	store.mu.Unlock()

	// Create a recent limiter.
	store.GetLimiter("recent", 10, 20)

	// Cleanup with 5 minute expiry.
	store.Cleanup(5 * time.Minute)

	// Check that old limiters were removed.
	store.mu.RLock()
	count := len(store.limiters)
	hasOld1 := false
	hasOld2 := false
	hasRecent := false
	for key := range store.limiters {
		if key == "old1" {
			hasOld1 = true
		}
		if key == "old2" {
			hasOld2 = true
		}
		if key == "recent" {
			hasRecent = true
		}
	}
	store.mu.RUnlock()

	if count != 1 {
		t.Errorf("expected 1 limiter after cleanup, got %d", count)
	}

	if hasOld1 || hasOld2 {
		t.Error("expected old limiters to be removed")
	}

	if !hasRecent {
		t.Error("expected recent limiter to remain")
	}
}

// Note: Headers are always enabled per RFC draft-ietf-httpapi-ratelimit-headers.
// This test has been removed as disabling headers is no longer supported.

func TestRateLimit_RemainingDecreases(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:    10,
		Burst:   5,
		Headers: true,
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	var prevRemaining = 999

	// Make 5 requests and verify remaining decreases.
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		remainingStr := rec.Header().Get("X-RateLimit-Remaining")
		remaining, err := strconv.Atoi(remainingStr)
		if err != nil {
			t.Fatalf("request %d: failed to parse remaining: %v", i+1, err)
		}

		if i == 0 {
			prevRemaining = remaining
		} else {
			if remaining >= prevRemaining {
				t.Errorf("request %d: expected remaining to decrease, prev=%d curr=%d", i+1, prevRemaining, remaining)
			}
			prevRemaining = remaining
		}
	}
}

func TestRateLimit_DefaultRate(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		// No Rate specified - should use default (10).
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Should succeed with default rate.
	if rec.Code != 200 {
		t.Errorf("expected status 200 with default rate, got %d", rec.Code)
	}

	// Check default rate header.
	limit := rec.Header().Get("X-RateLimit-Limit")
	if limit != "10" {
		t.Errorf("expected default X-RateLimit-Limit=10, got %s", limit)
	}
}

func TestRateLimit_DefaultBurst(t *testing.T) {
	router := fursy.New()
	router.Use(RateLimitWithConfig(RateLimitConfig{
		Rate: 5,
		// No Burst specified - should use default (2x rate = 10).
	}))

	router.GET("/", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Should allow burst of 10 requests (2x rate).
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", http.NoBody)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != 200 {
			t.Errorf("request %d with default burst: expected 200, got %d", i+1, rec.Code)
		}
	}

	// 11th request should be rate limited.
	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != 429 {
		t.Errorf("expected 429 after default burst, got %d", rec.Code)
	}
}
