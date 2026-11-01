// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coregx/fursy"
)

const testTimeout = 20 * time.Millisecond

// waitForDeadline blocks until the request context is done and returns its error.
func waitForDeadline(c *fursy.Context) error {
	<-c.Request.Context().Done()
	return c.Request.Context().Err()
}

func serveTimeout(t *testing.T, mw, h fursy.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	r := fursy.New()
	r.Use(mw)
	r.Handle("GET", "/", h)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", http.NoBody))
	return w
}

// TestTimeout_FastHandler tests that handlers finishing in time are untouched.
func TestTimeout_FastHandler(t *testing.T) {
	w := serveTimeout(t, Timeout(time.Second), func(c *fursy.Context) error {
		c.SetHeader("Cache-Control", "no-cache")
		return c.String(http.StatusOK, "ok")
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", w.Body.String())
	}
	if got := w.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("expected handler header to be kept, got %q", got)
	}
}

// TestTimeout_HandlerReturnsContextError tests the 503 RFC 9457 response and
// the ErrRequestTimeout cause for direct and wrapped deadline errors.
func TestTimeout_HandlerReturnsContextError(t *testing.T) {
	tests := []struct {
		name string
		wrap func(err error) error
	}{
		{"direct", func(err error) error { return err }},
		{"wrapped", func(err error) error { return fmt.Errorf("query failed: %w", err) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cause error
			w := serveTimeout(t, Timeout(testTimeout), func(c *fursy.Context) error {
				err := waitForDeadline(c)
				cause = context.Cause(c.Request.Context())
				return tt.wrap(err)
			})

			if w.Code != http.StatusServiceUnavailable {
				t.Errorf("expected status 503, got %d", w.Code)
			}
			if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
				t.Errorf("expected problem+json content type, got %q", ct)
			}
			body := w.Body.String()
			if !strings.Contains(body, "Service Unavailable") || !strings.Contains(body, "request timed out") {
				t.Errorf("unexpected body: %s", body)
			}
			if !errors.Is(cause, ErrRequestTimeout) {
				t.Errorf("expected cause ErrRequestTimeout, got %v", cause)
			}
		})
	}
}

// TestTimeout_HTTPClientError tests that outbound net/http errors, which wrap
// the context cause rather than context.DeadlineExceeded, are converted.
func TestTimeout_HTTPClientError(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer upstream.Close()

	var clientErr error
	w := serveTimeout(t, Timeout(testTimeout), func(c *fursy.Context) error {
		req, err := http.NewRequestWithContext(c.Request.Context(), "GET", upstream.URL, http.NoBody)
		if err != nil {
			return err
		}
		resp, err := upstream.Client().Do(req)
		if err != nil {
			clientErr = err
			return fmt.Errorf("upstream: %w", err)
		}
		_ = resp.Body.Close()
		return c.NoContent(http.StatusNoContent)
	})

	if !errors.Is(clientErr, ErrRequestTimeout) {
		t.Errorf("expected client error to wrap ErrRequestTimeout, got %v", clientErr)
	}
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

// TestTimeout_Nested tests that nested Timeout middleware only handle their own deadline.
func TestTimeout_Nested(t *testing.T) {
	gatewayTimeout := func(_ *fursy.Context, _ error) error {
		return fursy.NewProblem(http.StatusGatewayTimeout, "Gateway Timeout", "")
	}
	newRouter := func(outer, inner time.Duration) *fursy.Router {
		r := fursy.New()
		r.Use(Timeout(outer))
		r.Use(TimeoutWithConfig(TimeoutConfig{Timeout: inner, ErrorHandler: gatewayTimeout}))
		r.Handle("GET", "/", func(c *fursy.Context) error {
			_ = waitForDeadline(c)
			return fmt.Errorf("wrapped: %w", context.Cause(c.Request.Context()))
		})
		return r
	}

	tests := []struct {
		name         string
		outer, inner time.Duration
		want         int
	}{
		{"outer fires first", testTimeout, time.Second, http.StatusServiceUnavailable},
		{"inner fires first", time.Second, testTimeout, http.StatusGatewayTimeout},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			newRouter(tt.outer, tt.inner).ServeHTTP(w, httptest.NewRequest("GET", "/", http.NoBody))
			if w.Code != tt.want {
				t.Errorf("expected status %d, got %d", tt.want, w.Code)
			}
		})
	}
}

// TestTimeout_NilResultNotConverted tests that a handler returning nil after the
// deadline keeps its result instead of being reported as a timeout.
func TestTimeout_NilResultNotConverted(t *testing.T) {
	var called bool
	mw := TimeoutWithConfig(TimeoutConfig{
		Timeout: testTimeout,
		ErrorHandler: func(_ *fursy.Context, err error) error {
			called = true
			return err
		},
	})
	w := serveTimeout(t, mw, func(c *fursy.Context) error {
		_ = waitForDeadline(c)
		return nil
	})

	if called {
		t.Error("ErrorHandler must not be called for a nil result")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestTimeout_RestoresHeaders tests that headers set by the handler chain are
// discarded on timeout while headers set by outer middleware are kept.
func TestTimeout_RestoresHeaders(t *testing.T) {
	r := fursy.New()
	r.Use(func(c *fursy.Context) error {
		c.SetHeader("X-Request-ID", c.Query("id"))
		c.SetHeader("X-Frame-Options", "DENY")
		return c.Next()
	})
	r.Use(Timeout(testTimeout))
	r.Handle("GET", "/", func(c *fursy.Context) error {
		h := c.Response.Header()
		h.Set("Cache-Control", "public, max-age=3600")
		h.Set("Set-Cookie", "session=abc")
		h.Set("Content-Length", "2")
		h.Set("X-Frame-Options", "ALLOWALL")
		h.Add("X-Request-ID", "handler")
		return waitForDeadline(c)
	})

	// Two sequential requests also verify that pooled snapshots do not leak.
	for _, id := range []string{"first", "second"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/?id="+id, http.NoBody))

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s: expected status 503, got %d", id, w.Code)
		}
		h := w.Header()
		for _, k := range []string{"Cache-Control", "Set-Cookie", "Content-Length"} {
			if v := h.Get(k); v != "" {
				t.Errorf("%s: handler header %s leaked into timeout response: %q", id, k, v)
			}
		}
		if got := h.Values("X-Request-ID"); len(got) != 1 || got[0] != id {
			t.Errorf("%s: expected X-Request-ID [%s], got %v", id, id, got)
		}
		if got := h.Get("X-Frame-Options"); got != "DENY" {
			t.Errorf("%s: expected outer X-Frame-Options DENY, got %q", id, got)
		}
		if ct := h.Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
			t.Errorf("%s: expected problem+json content type, got %q", id, ct)
		}
	}
}

// TestTimeout_RestoresOnPanic tests that outer middleware sees the original
// request and response when the handler chain panics.
func TestTimeout_RestoresOnPanic(t *testing.T) {
	var (
		hasDeadline bool
		wrapped     bool
	)
	r := fursy.New()
	r.Use(func(c *fursy.Context) (err error) {
		defer func() {
			if recover() != nil {
				_, hasDeadline = c.Request.Context().Deadline()
				_, wrapped = c.Response.(*timeoutResponseWriter)
				err = fursy.InternalServerError("")
			}
		}()
		return c.Next()
	})
	r.Use(Timeout(time.Second))
	r.Handle("GET", "/", func(_ *fursy.Context) error {
		panic("boom")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", http.NoBody))

	if hasDeadline {
		t.Error("outer middleware should see the original request context after panic")
	}
	if wrapped {
		t.Error("outer middleware should see the original response writer after panic")
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// TestTimeout_ResponseAlreadyWritten tests that a written response is preserved.
func TestTimeout_ResponseAlreadyWritten(t *testing.T) {
	var gotErr error
	r := fursy.New()
	r.Use(func(c *fursy.Context) error {
		gotErr = c.Next()
		return gotErr
	})
	r.Use(Timeout(testTimeout))
	r.Handle("GET", "/", func(c *fursy.Context) error {
		if err := c.String(http.StatusOK, "partial"); err != nil {
			return err
		}
		<-c.Request.Context().Done()
		return nil
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", http.NoBody))

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "partial" {
		t.Errorf("expected body 'partial', got %q", w.Body.String())
	}
	if gotErr != nil {
		t.Errorf("expected nil error after written response, got %v", gotErr)
	}
}

// TestTimeout_UnrelatedError tests that unrelated errors pass through.
func TestTimeout_UnrelatedError(t *testing.T) {
	w := serveTimeout(t, Timeout(testTimeout), func(c *fursy.Context) error {
		_ = waitForDeadline(c)
		return fursy.BadRequest("bad input")
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestTimeout_OtherDeadlinesNotConverted tests that cancellations not caused
// by this middleware's deadline are passed through unchanged.
func TestTimeout_OtherDeadlinesNotConverted(t *testing.T) {
	canceled := func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}
	shortDeadline := func() context.Context {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		t.Cleanup(cancel)
		return ctx
	}

	tests := []struct {
		name    string
		parent  func() context.Context
		handler fursy.HandlerFunc
	}{
		{"client disconnect", canceled, waitForDeadline},
		{"shorter parent deadline", shortDeadline, waitForDeadline},
		{"handler's own shorter deadline", context.Background, func(c *fursy.Context) error {
			ctx, cancel := context.WithTimeout(c.Request.Context(), testTimeout)
			defer cancel()
			<-ctx.Done()
			return fmt.Errorf("query: %w", ctx.Err())
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			r := fursy.New()
			r.Use(TimeoutWithConfig(TimeoutConfig{
				Timeout: time.Second,
				ErrorHandler: func(_ *fursy.Context, err error) error {
					called = true
					return err
				},
			}))
			r.Handle("GET", "/", tt.handler)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", "/", http.NoBody).WithContext(tt.parent()))

			if called {
				t.Error("ErrorHandler must not be called")
			}
			if w.Code == http.StatusServiceUnavailable {
				t.Error("expected no 503")
			}
		})
	}
}

// TestTimeout_Deadline tests the deadline seen by the handler, including the
// DefaultTimeout fallback for zero or negative durations.
func TestTimeout_Deadline(t *testing.T) {
	tests := []struct {
		d, want time.Duration
	}{
		{5 * time.Second, 5 * time.Second},
		{0, DefaultTimeout},
		{-time.Second, DefaultTimeout},
	}
	for _, tt := range tests {
		t.Run(tt.d.String(), func(t *testing.T) {
			var (
				remaining time.Duration
				ok        bool
			)
			serveTimeout(t, Timeout(tt.d), func(c *fursy.Context) error {
				var deadline time.Time
				deadline, ok = c.Request.Context().Deadline()
				remaining = time.Until(deadline)
				return c.NoContent(http.StatusNoContent)
			})

			if !ok {
				t.Fatal("expected deadline on request context")
			}
			if remaining <= tt.want-time.Second || remaining > tt.want {
				t.Errorf("expected deadline ~%v, got %v", tt.want, remaining)
			}
		})
	}
}

// TestTimeout_Skipper tests that only skipped requests get no deadline.
func TestTimeout_Skipper(t *testing.T) {
	mw := TimeoutWithConfig(TimeoutConfig{
		Timeout: time.Second,
		Skipper: func(c *fursy.Context) bool { return c.Request.URL.Path == "/events" },
	})
	r := fursy.New()
	r.Use(mw)
	for _, path := range []string{"/events", "/api"} {
		r.Handle("GET", path, func(c *fursy.Context) error {
			_, hasDeadline := c.Request.Context().Deadline()
			return c.String(http.StatusOK, fmt.Sprint(hasDeadline))
		})
	}

	for path, want := range map[string]string{"/events": "false", "/api": "true"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", path, http.NoBody))
		if got := w.Body.String(); got != want {
			t.Errorf("%s: has deadline = %s, want %s", path, got, want)
		}
	}
}

// TestTimeout_CustomErrorHandler tests a custom 504 error handler.
func TestTimeout_CustomErrorHandler(t *testing.T) {
	var handlerErr error
	mw := TimeoutWithConfig(TimeoutConfig{
		Timeout: testTimeout,
		ErrorHandler: func(_ *fursy.Context, err error) error {
			handlerErr = err
			return fursy.NewProblem(http.StatusGatewayTimeout, "Gateway Timeout", "")
		},
	})
	w := serveTimeout(t, mw, waitForDeadline)

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("expected status 504, got %d", w.Code)
	}
	if !errors.Is(handlerErr, context.DeadlineExceeded) {
		t.Errorf("expected ErrorHandler to receive DeadlineExceeded, got %v", handlerErr)
	}
}

// TestTimeout_RouteGroup tests the middleware on a route group alongside Recovery and Logger.
func TestTimeout_RouteGroup(t *testing.T) {
	var logBuf bytes.Buffer
	var outerCtxErr error

	r := fursy.New()
	r.Use(RecoveryWithConfig(RecoveryConfig{DisablePrintStack: true, Logger: DefaultRecoveryLogger(&logBuf)}))
	r.Use(LoggerWithConfig(LoggerConfig{Logger: slog.New(slog.NewTextHandler(&logBuf, nil))}))
	r.Use(func(c *fursy.Context) error {
		err := c.Next()
		outerCtxErr = c.Request.Context().Err()
		return err
	})

	api := r.Group("/api")
	api.Use(Timeout(testTimeout))
	api.Handle("GET", "/slow", waitForDeadline)
	var otherHasDeadline bool
	r.Handle("GET", "/other", func(c *fursy.Context) error {
		_, otherHasDeadline = c.Request.Context().Deadline()
		return c.NoContent(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/slow", http.NoBody))

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
	if outerCtxErr != nil {
		t.Errorf("outer middleware should see original context, got %v", outerCtxErr)
	}
	if !strings.Contains(logBuf.String(), "request timed out") {
		t.Errorf("logger should record timeout error, got: %s", logBuf.String())
	}

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/other", http.NoBody))
	if otherHasDeadline {
		t.Error("route outside the group should not have a deadline")
	}
}

// hijackRecorder is an httptest.ResponseRecorder that supports hijacking.
type hijackRecorder struct {
	*httptest.ResponseRecorder
	hijacked bool
}

func (h *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijacked = true
	server, client := net.Pipe()
	_ = client.Close()
	return server, bufio.NewReadWriter(bufio.NewReader(server), bufio.NewWriter(server)), nil
}

// TestTimeout_ResponseWriterPassthrough tests Flush, Hijack and Unwrap on the wrapper.
func TestTimeout_ResponseWriterPassthrough(t *testing.T) {
	t.Run("flush", func(t *testing.T) {
		w := serveTimeout(t, Timeout(testTimeout), func(c *fursy.Context) error {
			if err := http.NewResponseController(c.Response).Flush(); err != nil {
				return err
			}
			_ = waitForDeadline(c)
			return nil
		})
		if !w.Flushed {
			t.Error("expected underlying writer to be flushed")
		}
		if w.Code == http.StatusServiceUnavailable {
			t.Error("flushed response must not be replaced by 503")
		}
	})

	t.Run("hijack", func(t *testing.T) {
		r := fursy.New()
		r.Use(Timeout(testTimeout))
		r.Handle("GET", "/", func(c *fursy.Context) error {
			conn, _, err := http.NewResponseController(c.Response).Hijack()
			if err != nil {
				return err
			}
			_ = conn.Close()
			return waitForDeadline(c)
		})

		rec := &hijackRecorder{ResponseRecorder: httptest.NewRecorder()}
		r.ServeHTTP(rec, httptest.NewRequest("GET", "/", http.NoBody))
		if !rec.hijacked {
			t.Error("expected underlying writer to be hijacked")
		}
		if rec.Code == http.StatusServiceUnavailable {
			t.Error("hijacked connection must not receive 503")
		}
	})

	t.Run("hijack unsupported", func(t *testing.T) {
		var hijackErr error
		w := serveTimeout(t, Timeout(time.Second), func(c *fursy.Context) error {
			_, _, hijackErr = c.Response.(http.Hijacker).Hijack()
			return c.NoContent(http.StatusNoContent)
		})
		if hijackErr == nil {
			t.Error("expected hijack error for non-hijackable writer")
		}
		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		var unwrapped http.ResponseWriter
		serveTimeout(t, Timeout(time.Second), func(c *fursy.Context) error {
			unwrapped = c.Response.(interface{ Unwrap() http.ResponseWriter }).Unwrap()
			return c.NoContent(http.StatusNoContent)
		})
		if unwrapped == nil {
			t.Error("expected non-nil unwrapped writer")
		}
	})
}

// plainResponseWriter hides optional interfaces such as http.Flusher.
type plainResponseWriter struct{ http.ResponseWriter }

// TestTimeoutResponseWriter_Committed tests which operations commit the response.
func TestTimeoutResponseWriter_Committed(t *testing.T) {
	tests := []struct {
		name string
		rw   http.ResponseWriter
		op   func(w *timeoutResponseWriter)
		want bool
	}{
		{"103 early hints", httptest.NewRecorder(), func(w *timeoutResponseWriter) { w.WriteHeader(http.StatusEarlyHints) }, false},
		{"101 switching protocols", httptest.NewRecorder(), func(w *timeoutResponseWriter) { w.WriteHeader(http.StatusSwitchingProtocols) }, true},
		{"200", httptest.NewRecorder(), func(w *timeoutResponseWriter) { w.WriteHeader(http.StatusOK) }, true},
		{"write", httptest.NewRecorder(), func(w *timeoutResponseWriter) { _, _ = w.Write([]byte("x")) }, true},
		{"flush supported", httptest.NewRecorder(), func(w *timeoutResponseWriter) { w.Flush() }, true},
		{"flush unsupported", plainResponseWriter{httptest.NewRecorder()}, func(w *timeoutResponseWriter) { w.Flush() }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &timeoutResponseWriter{ResponseWriter: tt.rw}
			tt.op(w)
			if w.written != tt.want {
				t.Errorf("written = %v, want %v", w.written, tt.want)
			}
		})
	}
}

// BenchmarkTimeout measures the overhead of the Timeout middleware against a bare router.
func BenchmarkTimeout(b *testing.B) {
	handler := func(c *fursy.Context) error {
		return c.NoContent(http.StatusNoContent)
	}
	run := func(b *testing.B, mw ...fursy.HandlerFunc) {
		b.Helper()
		r := fursy.New()
		r.Use(mw...)
		r.Handle("GET", "/", handler)
		req := httptest.NewRequest("GET", "/", http.NoBody)
		b.ReportAllocs()
		for b.Loop() {
			r.ServeHTTP(httptest.NewRecorder(), req)
		}
	}

	b.Run("baseline", func(b *testing.B) { run(b) })
	b.Run("timeout", func(b *testing.B) { run(b, Timeout(time.Second)) })
}
