// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware provides rate limiting middleware.
package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coregx/fursy"
	"golang.org/x/time/rate"
)

// RateLimitConfig defines the configuration for the RateLimit middleware.
type RateLimitConfig struct {
	// Limiter is the rate limiter instance.
	// If not set, uses Rate and Burst to create a new limiter.
	Limiter *rate.Limiter

	// Rate is the number of requests allowed per second.
	// Default: 10 requests/second
	Rate float64

	// Burst is the maximum burst size (bucket capacity).
	// Allows short bursts exceeding Rate.
	// Default: 20 (allows burst of 2x rate)
	Burst int

	// KeyFunc extracts the rate limit key from the request.
	// Common strategies:
	//   - IP-based: func(c) string { return c.RealIP() }
	//   - User-based: func(c) string { return c.GetString("user_id") }
	//   - API key: func(c) string { return c.Request.Header.Get("X-API-Key") }
	//   - Global: func(c) string { return "global" }
	// Default: IP-based (c.RealIP())
	KeyFunc func(c *fursy.Context) string

	// Skipper defines a function to skip the middleware.
	// Default: nil (middleware always executes)
	Skipper func(c *fursy.Context) bool

	// Store is the storage for per-key limiters.
	// If nil, uses in-memory map with cleanup goroutine.
	// For distributed systems, use external store (Redis, etc.)
	Store RateLimitStore

	// ErrorHandler is called when rate limit is exceeded.
	// Default: returns 429 Too Many Requests with Retry-After header
	ErrorHandler func(c *fursy.Context, retryAfter time.Duration) error

	// SuccessHandler is called after successful rate limit check.
	// Can be used for logging, metrics, etc.
	// Default: nil
	SuccessHandler func(c *fursy.Context, remaining int) error

	// Headers enables setting X-RateLimit-* headers.
	// Following RFC draft-ietf-httpapi-ratelimit-headers.
	// Default: true (always enabled, recommended by RFC)
	Headers bool

	// MaxKeys is the maximum number of keys to store in memory.
	// Prevents memory exhaustion from key explosion.
	// When exceeded, oldest keys are evicted (LRU).
	// Default: 10000
	// Set to 0 for unlimited (not recommended)
	MaxKeys int

	// CleanupInterval is the interval for cleaning up expired limiters.
	// Default: 1 minute
	CleanupInterval time.Duration

	// ExpireAfter is the duration after which inactive limiters are removed.
	// Default: 3 minutes
	ExpireAfter time.Duration
}

// RateLimitStore is an interface for storing rate limiters.
// Allows custom implementations (Redis, Memcached, etc.) for distributed systems.
type RateLimitStore interface {
	// GetLimiter returns the rate limiter for the given key.
	// If limiter doesn't exist, creates a new one with the given rate and burst.
	GetLimiter(key string, r rate.Limit, burst int) *rate.Limiter

	// Cleanup removes expired limiters (optional, for memory management).
	Cleanup(expireAfter time.Duration)
}

// inMemoryStore is the default in-memory store for rate limiters.
type inMemoryStore struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	maxKeys  int
}

// limiterEntry stores a rate limiter with its last access time.
type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// newInMemoryStore creates a new in-memory rate limiter store.
func newInMemoryStore(maxKeys int) *inMemoryStore {
	if maxKeys == 0 {
		maxKeys = 10000 // Default
	}

	return &inMemoryStore{
		limiters: make(map[string]*limiterEntry),
		maxKeys:  maxKeys,
	}
}

// GetLimiter returns the rate limiter for the given key.
func (s *inMemoryStore) GetLimiter(key string, r rate.Limit, burst int) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if limiter exists.
	if entry, ok := s.limiters[key]; ok {
		entry.lastAccess = time.Now()
		return entry.limiter
	}

	// Check if we need to evict (LRU).
	if s.maxKeys > 0 && len(s.limiters) >= s.maxKeys {
		s.evictOldest()
	}

	// Create new limiter.
	limiter := rate.NewLimiter(r, burst)
	s.limiters[key] = &limiterEntry{
		limiter:    limiter,
		lastAccess: time.Now(),
	}

	return limiter
}

// evictOldest removes the oldest limiter (LRU eviction).
func (s *inMemoryStore) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	// Find oldest entry.
	for key, entry := range s.limiters {
		if oldestKey == "" || entry.lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.lastAccess
		}
	}

	// Remove oldest.
	if oldestKey != "" {
		delete(s.limiters, oldestKey)
	}
}

// Cleanup removes expired limiters.
func (s *inMemoryStore) Cleanup(expireAfter time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, entry := range s.limiters {
		if now.Sub(entry.lastAccess) > expireAfter {
			delete(s.limiters, key)
		}
	}
}

// RateLimit returns a middleware that limits request rate using the token bucket algorithm.
//
// The middleware:
//   - Uses token bucket algorithm (golang.org/x/time/rate)
//   - Allows bursts up to Burst size
//   - Refills at Rate per second
//   - Sets RFC-draft standard headers (X-RateLimit-*)
//   - Returns 429 Too Many Requests when limit exceeded
//   - Supports multiple strategies (IP, user, API key, global)
//   - Automatic cleanup of expired limiters
//
// Features:
//   - Per-key rate limiting (default: IP-based)
//   - Burst handling (allows short traffic spikes)
//   - Standard headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
//   - Retry-After header on 429 response
//   - LRU eviction (prevents memory exhaustion)
//   - Configurable cleanup and expiration
//
// Example (global rate limit):
//
//	router := fursy.New()
//	router.Use(middleware.RateLimit(10, 20)) // 10 req/s, burst 20
//
// Example (per-IP rate limit):
//
//	router.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
//	    Rate:  5,
//	    Burst: 10,
//	    KeyFunc: func(c *fursy.Context) string {
//	        return c.RealIP()
//	    },
//	}))
//
// Example (per-user rate limit):
//
//	router.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
//	    Rate:  100,
//	    Burst: 200,
//	    KeyFunc: func(c *fursy.Context) string {
//	        // From JWT claims
//	        claims := c.Get(middleware.JWTContextKey).(jwt.MapClaims)
//	        return claims["sub"].(string)
//	    },
//	}))
//
// Example (layered defense - IP + user):
//
//	router.Use(middleware.RateLimit(1000, 2000)) // Global IP limit
//	router.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
//	    Rate:  10,
//	    Burst: 20,
//	    KeyFunc: func(c *fursy.Context) string {
//	        return c.GetString("user_id")
//	    },
//	})) // Per-user limit
func RateLimit(r float64, burst int) fursy.HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{
		Rate:  r,
		Burst: burst,
	})
}

// RateLimitWithConfig returns a middleware with custom rate limit configuration.
//
//nolint:gocognit,gocyclo,cyclop // Rate limiting logic requires multiple checks and branches
func RateLimitWithConfig(config RateLimitConfig) fursy.HandlerFunc {
	// Set defaults.
	if config.Rate == 0 {
		config.Rate = 10 // 10 requests/second
	}

	if config.Burst == 0 {
		config.Burst = int(config.Rate * 2) // Allow 2x burst
	}

	if config.KeyFunc == nil {
		// Default: IP-based rate limiting.
		config.KeyFunc = func(c *fursy.Context) string {
			return getClientIP(c.Request)
		}
	}

	if config.Store == nil {
		config.Store = newInMemoryStore(config.MaxKeys)
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultRateLimitErrorHandler
	}

	// Headers always enabled (RFC-compliant) unless explicitly disabled.
	if !config.Headers {
		// Enable headers by default.
		config.Headers = true
	}

	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Minute
	}

	if config.ExpireAfter == 0 {
		config.ExpireAfter = 3 * time.Minute
	}

	// Start cleanup goroutine.
	go func() {
		ticker := time.NewTicker(config.CleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			config.Store.Cleanup(config.ExpireAfter)
		}
	}()

	// Create rate limit from config.
	rateLimit := rate.Limit(config.Rate)

	return func(c *fursy.Context) error {
		// Skip if Skipper returns true.
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		// Get rate limit key.
		key := config.KeyFunc(c)

		// Get or create limiter for this key.
		var limiter *rate.Limiter
		if config.Limiter != nil {
			// Use provided limiter (global rate limit).
			limiter = config.Limiter
		} else {
			// Use per-key limiter.
			limiter = config.Store.GetLimiter(key, rateLimit, config.Burst)
		}

		// Try to consume a token.
		reservation := limiter.Reserve()
		if !reservation.OK() {
			// Rate limit exceeded (should not happen with valid config).
			return config.ErrorHandler(c, time.Second)
		}

		delay := reservation.Delay()
		if delay > 0 {
			// Rate limit exceeded - cancel reservation.
			reservation.Cancel()

			// Set Retry-After header.
			retryAfter := delay
			return config.ErrorHandler(c, retryAfter)
		}

		// Set X-RateLimit-* headers.
		if config.Headers {
			setRateLimitHeaders(c, limiter, int(config.Rate), config.Burst)
		}

		// Call success handler if configured.
		if config.SuccessHandler != nil {
			remaining := int(limiter.Tokens())
			if err := config.SuccessHandler(c, remaining); err != nil {
				return err
			}
		}

		return c.Next()
	}
}

// setRateLimitHeaders sets the standard rate limit headers.
func setRateLimitHeaders(c *fursy.Context, limiter *rate.Limiter, limit, _ int) {
	// X-RateLimit-Limit: The maximum number of requests allowed per window.
	c.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", limit))

	// X-RateLimit-Remaining: The number of requests remaining in the current window.
	remaining := int(limiter.Tokens())
	if remaining < 0 {
		remaining = 0
	}
	c.SetHeader("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

	// X-RateLimit-Reset: Unix timestamp when the rate limit resets.
	// Token bucket refills continuously, but we can estimate next full bucket.
	reset := time.Now().Add(time.Second).Unix()
	c.SetHeader("X-RateLimit-Reset", fmt.Sprintf("%d", reset))
}

// defaultRateLimitErrorHandler is the default error handler for rate limit exceeded.
func defaultRateLimitErrorHandler(c *fursy.Context, retryAfter time.Duration) error {
	// Set Retry-After header (RFC 6585).
	c.SetHeader("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())+1))

	// Set rate limit headers showing zero remaining.
	c.SetHeader("X-RateLimit-Remaining", "0")

	// Return 429 Too Many Requests.
	return c.String(http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
}
