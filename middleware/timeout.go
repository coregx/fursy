// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package middleware provides request timeout middleware.
package middleware

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/coregx/fursy"
)

// DefaultTimeout is the request timeout used when TimeoutConfig.Timeout is not positive.
const DefaultTimeout = 30 * time.Second

// ErrRequestTimeout is the cancellation cause attached to the request context
// when the Timeout middleware deadline expires.
//
// Handlers can distinguish a middleware timeout from other cancellations
// (client disconnect, server shutdown, upstream deadlines) with:
//
//	if errors.Is(context.Cause(c.Request.Context()), middleware.ErrRequestTimeout) {
//	    // request exceeded the configured timeout
//	}
var ErrRequestTimeout = errors.New("request timed out")

// TimeoutConfig defines the configuration for the Timeout middleware.
type TimeoutConfig struct {
	// Skipper defines a function to skip the middleware.
	// Use it to exclude long-lived routes such as SSE or WebSocket endpoints.
	// Default: nil (middleware always executes)
	Skipper func(c *fursy.Context) bool

	// ErrorHandler is called when the deadline expires, the handler chain
	// returned an error wrapping context.DeadlineExceeded or ErrRequestTimeout,
	// and no response was written. It receives that (non-nil) error and returns
	// the error to propagate to the router's error handler. Response headers set
	// by the handler chain have already been discarded.
	// Default: returns fursy.ServiceUnavailable (503, RFC 9457)
	ErrorHandler func(c *fursy.Context, err error) error

	// Timeout is the maximum duration allowed for the rest of the handler chain.
	// Default: DefaultTimeout (30s) when zero or negative
	Timeout time.Duration
}

// Timeout returns a middleware that bounds request processing time to d.
//
// The middleware is cooperative: it attaches a deadline to c.Request.Context()
// and runs the handler chain synchronously. Handlers must honor the context
// (database drivers, HTTP clients, select on ctx.Done()) for the timeout to
// take effect. This design is required because fursy pools Context objects:
// running the handler in a separate goroutine would let it touch a Context
// that has already been recycled for another request.
//
// When the deadline expires, the handler chain returns an error wrapping
// context.DeadlineExceeded or ErrRequestTimeout (net/http clients return the
// context cause), and nothing has been written, the middleware discards the
// response headers set by the chain and returns a 503 Service Unavailable
// RFC 9457 Problem. Written responses, nil results and unrelated errors are
// passed through unchanged, as are cancellations that did not originate from
// this middleware (client disconnect, shorter parent deadline, outer Timeout).
//
// Timeout does not replace http.Server timeouts: request body reads and
// response writes are not bounded by the request context. Keep
// ReadHeaderTimeout, ReadTimeout and WriteTimeout configured, with
// WriteTimeout longer than d so the timeout response can be delivered.
//
// Context deadlines can only be shortened, so a nested Timeout cannot extend
// an outer one. Apply Timeout per route group, or exclude long-running routes
// with Skipper, when routes need different limits.
//
// Example:
//
//	router := fursy.New()
//	router.Use(middleware.Recovery(), middleware.Logger())
//	router.Use(middleware.Timeout(5 * time.Second))
func Timeout(d time.Duration) fursy.HandlerFunc {
	return TimeoutWithConfig(TimeoutConfig{Timeout: d})
}

// TimeoutWithConfig returns a Timeout middleware with custom configuration.
//
// Example:
//
//	router.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
//	    Timeout: 2 * time.Second,
//	    Skipper: func(c *fursy.Context) bool {
//	        return strings.HasPrefix(c.Request.URL.Path, "/events")
//	    },
//	    ErrorHandler: func(_ *fursy.Context, _ error) error {
//	        return fursy.NewProblem(http.StatusGatewayTimeout, "Gateway Timeout", "")
//	    },
//	}))
func TimeoutWithConfig(config TimeoutConfig) fursy.HandlerFunc {
	if config.Timeout <= 0 {
		config.Timeout = DefaultTimeout
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultTimeoutErrorHandler
	}

	return func(c *fursy.Context) error {
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		req, res := c.Request, c.Response
		parent := req.Context()
		ctx, cancel := context.WithTimeoutCause(parent, config.Timeout, ErrRequestTimeout)
		defer cancel()

		st := acquireTimeoutState(res)
		defer func() {
			// Also runs on panic so outer middleware (e.g. Recovery) sees the original request.
			c.Request, c.Response = req, res
			releaseTimeoutState(st)
		}()
		c.Request = req.WithContext(ctx)
		c.Response = &st.writer

		err := c.Next()
		c.Request, c.Response = req, res

		if st.writer.written || !isOwnTimeout(ctx, parent, err) {
			return err
		}
		st.restoreHeaders(res.Header())
		return config.ErrorHandler(c, err)
	}
}

// isOwnTimeout reports whether err was caused by this middleware's deadline:
// the derived context expired while the parent was still alive.
func isOwnTimeout(ctx, parent context.Context, err error) bool {
	if err == nil || ctx.Err() == nil || parent.Err() != nil {
		return false
	}
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ErrRequestTimeout)
}

// defaultTimeoutErrorHandler returns a 503 Service Unavailable problem.
func defaultTimeoutErrorHandler(_ *fursy.Context, _ error) error {
	return fursy.ServiceUnavailable(ErrRequestTimeout.Error())
}

// timeoutState is the pooled per-request state of the Timeout middleware.
type timeoutState struct {
	writer timeoutResponseWriter
	// headers is a shallow snapshot of the response headers before the chain ran.
	headers http.Header
}

var timeoutStatePool = sync.Pool{
	New: func() any { return &timeoutState{headers: make(http.Header)} },
}

func acquireTimeoutState(res http.ResponseWriter) *timeoutState {
	st, _ := timeoutStatePool.Get().(*timeoutState)
	st.writer.ResponseWriter = res
	// Header.Set/Del replace or remove entries and Header.Add appends past the
	// snapshot's length, so sharing the value slices is safe.
	for k, v := range res.Header() {
		st.headers[k] = v
	}
	return st
}

func releaseTimeoutState(st *timeoutState) {
	st.writer = timeoutResponseWriter{}
	clear(st.headers)
	timeoutStatePool.Put(st)
}

// restoreHeaders resets h to the snapshot taken before the handler chain ran.
func (st *timeoutState) restoreHeaders(h http.Header) {
	clear(h)
	for k, v := range st.headers {
		h[k] = v
	}
}

// timeoutResponseWriter tracks whether the handler chain committed a response.
type timeoutResponseWriter struct {
	http.ResponseWriter
	written bool
}

// WriteHeader marks the response as written and delegates to the underlying writer.
// Informational (1xx) statuses other than 101 Switching Protocols do not commit the response.
func (w *timeoutResponseWriter) WriteHeader(code int) {
	if code >= http.StatusOK || code == http.StatusSwitchingProtocols {
		w.written = true
	}
	w.ResponseWriter.WriteHeader(code)
}

// Write marks the response as written and delegates to the underlying writer.
func (w *timeoutResponseWriter) Write(b []byte) (int, error) {
	w.written = true
	return w.ResponseWriter.Write(b)
}

// Flush implements http.Flusher by delegating to the underlying ResponseWriter.
func (w *timeoutResponseWriter) Flush() {
	if http.NewResponseController(w.ResponseWriter).Flush() == nil {
		w.written = true
	}
}

// Hijack implements http.Hijacker by delegating to the underlying ResponseWriter.
func (w *timeoutResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, rw, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err == nil {
		w.written = true
	}
	return conn, rw, err
}

// Unwrap returns the underlying ResponseWriter.
func (w *timeoutResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
