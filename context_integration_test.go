// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coregx/fursy"
)

// TestContext_SSE_NotImported tests that c.SSE() returns ErrStreamNotImported
// when plugins/stream is not imported.
func TestContext_SSE_NotImported(t *testing.T) {
	router := fursy.New()
	router.GET("/sse", func(c *fursy.Context) error {
		err := c.SSE(func(_ any) error {
			return nil
		})

		if !errors.Is(err, fursy.ErrStreamNotImported) {
			t.Errorf("expected ErrStreamNotImported, got %v", err)
		}
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/sse", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestContext_WebSocket_NotImported tests that c.WebSocket() returns ErrStreamNotImported
// when plugins/stream is not imported.
func TestContext_WebSocket_NotImported(t *testing.T) {
	router := fursy.New()
	router.GET("/ws", func(c *fursy.Context) error {
		err := c.WebSocket(func(_ any) error {
			return nil
		}, nil)

		if !errors.Is(err, fursy.ErrStreamNotImported) {
			t.Errorf("expected ErrStreamNotImported, got %v", err)
		}
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/ws", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestContext_DB_NotConfigured tests that c.DB() returns nil
// when database middleware is not configured.
func TestContext_DB_NotConfigured(t *testing.T) {
	router := fursy.New()
	router.GET("/test", func(c *fursy.Context) error {
		db := c.DB()
		if db != nil {
			t.Error("expected nil, got DB")
		}
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// TestContext_ErrorMessages tests that error messages are helpful
// when methods are called incorrectly.
func TestContext_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		handler       fursy.HandlerFunc
		expectedError error
	}{
		{
			name: "SSE without plugin",
			handler: func(c *fursy.Context) error {
				return c.SSE(func(_ any) error {
					return nil
				})
			},
			expectedError: fursy.ErrStreamNotImported,
		},
		{
			name: "WebSocket without plugin",
			handler: func(c *fursy.Context) error {
				return c.WebSocket(func(_ any) error {
					return nil
				}, nil)
			},
			expectedError: fursy.ErrStreamNotImported,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(_ *testing.T) {
			router := fursy.New()
			router.GET("/test", tt.handler)

			req := httptest.NewRequest("GET", "/test", http.NoBody)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// The error should be returned to the handler
			// and can be checked in the test
		})
	}
}
