// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware provides circuit breaker middleware for resilience.
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coregx/fursy"
)

// State represents the circuit breaker state.
type State int

const (
	// StateClosed means requests pass through normally.
	StateClosed State = iota
	// StateOpen means requests are blocked (fail fast).
	StateOpen
	// StateHalfOpen means limited requests are allowed for testing recovery.
	StateHalfOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// Counts tracks circuit breaker statistics.
type Counts struct {
	Requests             int
	TotalSuccesses       int
	TotalFailures        int
	ConsecutiveSuccesses int
	ConsecutiveFailures  int
}

// CircuitBreakerConfig defines the configuration for the CircuitBreaker middleware.
type CircuitBreakerConfig struct {
	// ConsecutiveFailures is the number of consecutive failures before opening circuit.
	// Used when ReadyToTrip is not set.
	// Default: 5
	ConsecutiveFailures int

	// FailureThreshold is the number of failures in the window before opening circuit.
	// Used with RequestWindow or TimeWindow for ratio-based threshold.
	// Default: 0 (disabled, use consecutive failures instead)
	FailureThreshold int

	// RequestWindow is the number of requests in the sliding window.
	// Used with FailureThreshold for count-based ratio threshold.
	// Example: FailureThreshold=5, RequestWindow=10 means 50% failure rate triggers Open.
	// Default: 0 (disabled, use consecutive failures instead)
	RequestWindow int

	// TimeWindow is the duration of the sliding time window.
	// Used with FailureThreshold for time-based ratio threshold.
	// Example: FailureThreshold=5, TimeWindow=10s means 5 failures in 10s triggers Open.
	// Default: 0 (disabled, use count-based window or consecutive failures)
	TimeWindow time.Duration

	// Timeout is the duration to stay in Open state before transitioning to Half-Open.
	// Default: 60 seconds
	Timeout time.Duration

	// MaxRequests is the maximum number of requests allowed in Half-Open state.
	// If MaxRequests is 0, circuit breaker allows only 1 request in Half-Open.
	// Default: 1
	MaxRequests int

	// ReadyToTrip is called with a copy of Counts whenever a request fails in Closed state.
	// If ReadyToTrip returns true, the circuit breaker will transition to Open state.
	// If ReadyToTrip is nil, default behavior is used (ConsecutiveFailures >= threshold).
	ReadyToTrip func(counts Counts) bool

	// OnStateChange is called whenever the circuit breaker changes state.
	// Can be used for logging, metrics, etc.
	// Default: nil
	OnStateChange func(from, to State)

	// Skipper defines a function to skip the middleware.
	// Default: nil (middleware always executes)
	Skipper func(c *fursy.Context) bool

	// IsSuccessful determines if a request is considered successful.
	// Default: status < 500 (5xx errors are failures)
	IsSuccessful func(c *fursy.Context) bool

	// ErrorHandler is called when the circuit breaker is open.
	// Default: returns 503 Service Unavailable
	ErrorHandler func(c *fursy.Context) error

	// Name is the circuit breaker instance name (for logging/metrics).
	// Default: "default"
	Name string
}

// circuitBreaker implements the circuit breaker state machine.
type circuitBreaker struct {
	config CircuitBreakerConfig
	state  State
	counts Counts
	expiry time.Time
	mu     sync.RWMutex

	// For time-based window tracking
	requests []requestRecord

	// isCustomReadyToTrip indicates if ReadyToTrip was customized by user
	isCustomReadyToTrip bool
}

// cbResponseWriter wraps http.ResponseWriter to capture status code.
type cbResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *cbResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *cbResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *cbResponseWriter) Status() int {
	if !w.written {
		return 0
	}
	return w.statusCode
}

// requestRecord tracks individual request outcomes for time-based windows.
type requestRecord struct {
	timestamp time.Time
	success   bool
}

// CircuitBreaker returns a middleware that implements the circuit breaker pattern.
//
// The middleware protects your application from cascading failures by monitoring
// request failures and opening the circuit when failures exceed thresholds.
//
// Circuit Breaker States:
//   - Closed: Requests pass through normally. Failures are tracked.
//   - Open: Requests are blocked immediately (fail fast). After Timeout, transitions to Half-Open.
//   - Half-Open: Limited requests are allowed to test if service recovered.
//
// Features:
//   - Consecutive failures threshold (default: 5 consecutive failures)
//   - Window-based ratio threshold (count-based or time-based)
//   - Configurable timeout for Open → Half-Open transition
//   - MaxRequests in Half-Open state (default: 1)
//   - Custom ReadyToTrip callback for advanced logic
//   - OnStateChange callback for logging/metrics
//   - Thread-safe implementation
//
// Example (simple - consecutive failures):
//
//	router := fursy.New()
//	router.Use(middleware.CircuitBreaker())  // Default: 5 consecutive failures
//
// Example (ratio-based - count window):
//
//	router.Use(middleware.CircuitBreakerWithConfig(middleware.CircuitBreakerConfig{
//	    FailureThreshold: 5,    // 5 failures
//	    RequestWindow:    10,   // out of last 10 requests
//	    Timeout:          30 * time.Second,
//	}))
//
// Example (ratio-based - time window):
//
//	router.Use(middleware.CircuitBreakerWithConfig(middleware.CircuitBreakerConfig{
//	    FailureThreshold: 10,   // 10 failures
//	    TimeWindow:       60 * time.Second,  // in last 60 seconds
//	    Timeout:          30 * time.Second,
//	}))
//
// Example (with callbacks):
//
//	router.Use(middleware.CircuitBreakerWithConfig(middleware.CircuitBreakerConfig{
//	    ConsecutiveFailures: 3,
//	    Timeout:             60 * time.Second,
//	    OnStateChange: func(from, to State) {
//	        log.Printf("Circuit breaker: %s -> %s", from, to)
//	    },
//	}))
func CircuitBreaker() fursy.HandlerFunc {
	return CircuitBreakerWithConfig(CircuitBreakerConfig{})
}

// CircuitBreakerWithConfig returns a middleware with custom circuit breaker configuration.
//
//nolint:gocognit,gocyclo,cyclop // Circuit breaker has natural complexity due to state machine logic.
func CircuitBreakerWithConfig(config CircuitBreakerConfig) fursy.HandlerFunc {
	// Set defaults.
	if config.ConsecutiveFailures == 0 {
		config.ConsecutiveFailures = 5
	}

	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	if config.MaxRequests == 0 {
		config.MaxRequests = 1
	}

	// Track if ReadyToTrip is custom (set by user) or default.
	isCustomReadyToTrip := config.ReadyToTrip != nil

	//nolint:nestif // Configuration setup has natural nesting.
	if config.ReadyToTrip == nil {
		// Default: open circuit after consecutive failures threshold.
		config.ReadyToTrip = func(counts Counts) bool {
			// If using count-based window threshold
			if config.FailureThreshold > 0 && config.RequestWindow > 0 {
				// Count-based window: check failure ratio
				if counts.Requests >= config.RequestWindow {
					return counts.TotalFailures >= config.FailureThreshold
				}
				return false
			}
			// If using time-based window threshold
			if config.FailureThreshold > 0 && config.TimeWindow > 0 {
				// Time-based window: check if failures exceed threshold
				return counts.TotalFailures >= config.FailureThreshold
			}
			// Otherwise use consecutive failures
			return counts.ConsecutiveFailures >= config.ConsecutiveFailures
		}
	}

	if config.IsSuccessful == nil {
		// Default: 5xx status codes are failures.
		config.IsSuccessful = func(c *fursy.Context) bool {
			// Get status from response
			if rw, ok := c.Response.(interface{ Status() int }); ok {
				status := rw.Status()
				return status > 0 && status < 500
			}
			return true
		}
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultCircuitBreakerErrorHandler
	}

	if config.Name == "" {
		config.Name = "default"
	}

	// Create circuit breaker instance.
	cb := &circuitBreaker{
		config:              config,
		state:               StateClosed,
		counts:              Counts{},
		isCustomReadyToTrip: isCustomReadyToTrip,
	}

	// If using time-based window, initialize requests slice.
	if config.FailureThreshold > 0 && config.TimeWindow > 0 {
		cb.requests = make([]requestRecord, 0)
	}

	return func(c *fursy.Context) error {
		// Skip if Skipper returns true.
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		// Check if circuit breaker allows request.
		if err := cb.beforeRequest(); err != nil {
			return config.ErrorHandler(c)
		}

		// Wrap response writer to capture status code.
		// c.Response is of type any, so we check if it implements http.ResponseWriter.
		originalResponse := c.Response
		//nolint:staticcheck // c.Response is any, type assertion is necessary.
		rw, ok := originalResponse.(http.ResponseWriter)
		if ok {
			wrapper := &cbResponseWriter{
				ResponseWriter: rw,
				statusCode:     0,
				written:        false,
			}
			c.Response = wrapper
			defer func() {
				c.Response = originalResponse // Restore original
			}()
		}

		// Execute request.
		err := c.Next()

		// Record result.
		success := err == nil && config.IsSuccessful(c)
		cb.afterRequest(success)

		return err
	}
}

// beforeRequest checks if the circuit breaker allows the request.
func (cb *circuitBreaker) beforeRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.state

	switch state {
	case StateClosed:
		// Request allowed in Closed state.
		return nil

	case StateOpen:
		// Check if timeout has expired.
		if now.After(cb.expiry) {
			// Transition to Half-Open.
			cb.setState(StateHalfOpen)
			// Reset counts for Half-Open testing.
			cb.counts = Counts{}
			return nil
		}
		// Still open, block request.
		return errors.New("circuit breaker is open")

	case StateHalfOpen:
		// Check if we've reached MaxRequests.
		if cb.counts.Requests >= cb.config.MaxRequests {
			// Too many requests in Half-Open, block this one.
			return errors.New("circuit breaker is half-open (max requests reached)")
		}
		// Allow request in Half-Open (for testing recovery).
		return nil

	default:
		return errors.New("unknown circuit breaker state")
	}
}

// afterRequest records the request result and updates state.
//
//nolint:gocognit,gocyclo,cyclop // State machine logic has natural complexity.
func (cb *circuitBreaker) afterRequest(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.state

	// Update counts.
	cb.counts.Requests++

	if success {
		cb.counts.TotalSuccesses++
		cb.counts.ConsecutiveSuccesses++
		cb.counts.ConsecutiveFailures = 0
	} else {
		cb.counts.TotalFailures++
		cb.counts.ConsecutiveFailures++
		cb.counts.ConsecutiveSuccesses = 0
	}

	// For time-based window, track individual requests.
	if cb.config.FailureThreshold > 0 && cb.config.TimeWindow > 0 {
		// Add current request.
		cb.requests = append(cb.requests, requestRecord{
			timestamp: now,
			success:   success,
		})

		// Remove expired requests outside time window.
		cutoff := now.Add(-cb.config.TimeWindow)
		validIdx := 0
		for i, req := range cb.requests {
			if req.timestamp.After(cutoff) {
				validIdx = i
				break
			}
		}
		cb.requests = cb.requests[validIdx:]

		// Recalculate counts from window.
		windowSuccesses := 0
		windowFailures := 0
		for _, req := range cb.requests {
			if req.success {
				windowSuccesses++
			} else {
				windowFailures++
			}
		}

		// Update counts to reflect window.
		cb.counts.TotalSuccesses = windowSuccesses
		cb.counts.TotalFailures = windowFailures
		cb.counts.Requests = len(cb.requests)
	}

	// State transitions based on current state and result.
	switch state {
	case StateClosed:
		// Determine when to check ReadyToTrip:
		// 1. Always check if custom ReadyToTrip provided
		// 2. For window-based strategies, check after every request
		// 3. For default consecutive failures, check only on failures
		shouldCheck := !success // Default: check on failure

		if cb.isCustomReadyToTrip {
			// Custom ReadyToTrip: always check (user may have complex logic)
			shouldCheck = true
		} else if cb.config.FailureThreshold > 0 && (cb.config.RequestWindow > 0 || cb.config.TimeWindow > 0) {
			// Window-based: always check (threshold calculated from window)
			shouldCheck = true
		}

		if shouldCheck {
			// Check if we should trip to Open.
			if cb.config.ReadyToTrip(cb.counts) {
				cb.setState(StateOpen)
				cb.expiry = now.Add(cb.config.Timeout)
			}
		}

	case StateHalfOpen:
		if success {
			// Check if we've had enough successes to close circuit.
			if cb.counts.ConsecutiveSuccesses >= cb.config.MaxRequests {
				// All test requests succeeded, close circuit.
				cb.setState(StateClosed)
				cb.counts = Counts{}
				cb.requests = nil
			}
		} else {
			// Failure in Half-Open, reopen circuit.
			cb.setState(StateOpen)
			cb.expiry = now.Add(cb.config.Timeout)
		}
	}
}

// setState transitions to a new state and calls OnStateChange callback.
func (cb *circuitBreaker) setState(newState State) {
	oldState := cb.state

	if oldState == newState {
		return
	}

	cb.state = newState

	// Call state change callback if configured.
	if cb.config.OnStateChange != nil {
		// Call callback without holding lock (avoid deadlock).
		go cb.config.OnStateChange(oldState, newState)
	}
}

// GetState returns the current state (for testing/monitoring).
func (cb *circuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetCounts returns a copy of current counts (for testing/monitoring).
func (cb *circuitBreaker) GetCounts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.counts
}

// defaultCircuitBreakerErrorHandler is the default error handler for open circuit.
func defaultCircuitBreakerErrorHandler(c *fursy.Context) error {
	return c.String(http.StatusServiceUnavailable, "Service temporarily unavailable (circuit breaker open)")
}

// Reset manually resets the circuit breaker to Closed state (for testing).
func (cb *circuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.counts = Counts{}
	cb.expiry = time.Time{}
	cb.requests = nil
}

// CircuitBreakerWithName returns a circuit breaker with a specific name.
// Useful for having multiple circuit breakers with different configurations.
func CircuitBreakerWithName(name string, config CircuitBreakerConfig) fursy.HandlerFunc {
	config.Name = name
	return CircuitBreakerWithConfig(config)
}

// CircuitBreakerConsecutive is a helper that creates a circuit breaker
// with consecutive failures threshold.
func CircuitBreakerConsecutive(failures int, timeout time.Duration) fursy.HandlerFunc {
	return CircuitBreakerWithConfig(CircuitBreakerConfig{
		ConsecutiveFailures: failures,
		Timeout:             timeout,
	})
}

// CircuitBreakerRatio is a helper that creates a circuit breaker
// with ratio-based threshold (count window).
func CircuitBreakerRatio(failures, requests int, timeout time.Duration) fursy.HandlerFunc {
	return CircuitBreakerWithConfig(CircuitBreakerConfig{
		FailureThreshold: failures,
		RequestWindow:    requests,
		Timeout:          timeout,
	})
}

// CircuitBreakerTimeWindow is a helper that creates a circuit breaker
// with time-based window threshold.
func CircuitBreakerTimeWindow(failures int, window, timeout time.Duration) fursy.HandlerFunc {
	return CircuitBreakerWithConfig(CircuitBreakerConfig{
		FailureThreshold: failures,
		TimeWindow:       window,
		Timeout:          timeout,
	})
}

// FormatState returns a formatted string with circuit breaker state and counts.
func FormatState(cb *circuitBreaker) string {
	state := cb.GetState()
	counts := cb.GetCounts()

	return fmt.Sprintf(
		"State: %s, Requests: %d, Successes: %d, Failures: %d, Consecutive Failures: %d",
		state, counts.Requests, counts.TotalSuccesses, counts.TotalFailures, counts.ConsecutiveFailures,
	)
}
