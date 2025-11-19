// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stream_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/stream"
	"github.com/coregx/stream/sse"
	"github.com/coregx/stream/websocket"
)

// Test 1: SSEHub middleware stores hub in context.
func TestSSEHub_Middleware(t *testing.T) {
	t.Helper()

	hub := sse.NewHub[string]()
	defer hub.Close()

	router := fursy.New()
	router.Use(stream.SSEHub(hub))

	router.GET("/test", func(c *fursy.Context) error {
		retrievedHub, ok := stream.GetSSEHub[string](c)
		if !ok {
			t.Error("hub not found in context")
		}
		if retrievedHub != hub {
			t.Error("wrong hub retrieved")
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

// Test 2: WebSocketHub middleware stores hub in context.
func TestWebSocketHub_Middleware(t *testing.T) {
	t.Helper()

	hub := websocket.NewHub()
	defer hub.Close()

	router := fursy.New()
	router.Use(stream.WebSocketHub(hub))

	router.GET("/test", func(c *fursy.Context) error {
		retrievedHub, ok := stream.GetWebSocketHub(c)
		if !ok {
			t.Error("hub not found in context")
		}
		if retrievedHub != hub {
			t.Error("wrong hub retrieved")
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

// Test 3: GetSSEHub returns false if hub not in context.
func TestGetSSEHub_NotFound(t *testing.T) {
	t.Helper()

	router := fursy.New()

	router.GET("/test", func(c *fursy.Context) error {
		_, ok := stream.GetSSEHub[string](c)
		if ok {
			t.Error("expected hub not found, but got ok=true")
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

// Test 4: GetSSEHub returns false if wrong type.
func TestGetSSEHub_WrongType(t *testing.T) {
	t.Helper()

	hub := sse.NewHub[string]()
	defer hub.Close()

	router := fursy.New()
	router.Use(stream.SSEHub(hub))

	router.GET("/test", func(c *fursy.Context) error {
		// Try to get hub with wrong type (int instead of string).
		_, ok := stream.GetSSEHub[int](c)
		if ok {
			t.Error("expected type mismatch, but got ok=true")
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

// Test 5: GetWebSocketHub returns false if hub not in context.
func TestGetWebSocketHub_NotFound(t *testing.T) {
	t.Helper()

	router := fursy.New()

	router.GET("/test", func(c *fursy.Context) error {
		_, ok := stream.GetWebSocketHub(c)
		if ok {
			t.Error("expected hub not found, but got ok=true")
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

// Test 6: SSEHub middleware with multiple handlers.
func TestSSEHub_MultipleHandlers(t *testing.T) {
	t.Helper()

	hub := sse.NewHub[string]()
	defer hub.Close()

	router := fursy.New()
	router.Use(stream.SSEHub(hub))

	// Handler 1: Check hub exists.
	router.GET("/handler1", func(c *fursy.Context) error {
		_, ok := stream.GetSSEHub[string](c)
		if !ok {
			t.Error("hub not found in handler1")
		}
		return c.JSON(200, map[string]string{"handler": "1"})
	})

	// Handler 2: Check hub exists.
	router.GET("/handler2", func(c *fursy.Context) error {
		_, ok := stream.GetSSEHub[string](c)
		if !ok {
			t.Error("hub not found in handler2")
		}
		return c.JSON(200, map[string]string{"handler": "2"})
	})

	// Test handler1.
	req1 := httptest.NewRequest("GET", "/handler1", http.NoBody)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Errorf("handler1: expected 200, got %d", w1.Code)
	}

	// Test handler2.
	req2 := httptest.NewRequest("GET", "/handler2", http.NoBody)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Errorf("handler2: expected 200, got %d", w2.Code)
	}
}

// Test 7: WebSocketHub middleware with multiple handlers.
func TestWebSocketHub_MultipleHandlers(t *testing.T) {
	t.Helper()

	hub := websocket.NewHub()
	defer hub.Close()

	router := fursy.New()
	router.Use(stream.WebSocketHub(hub))

	// Handler 1: Check hub exists.
	router.GET("/handler1", func(c *fursy.Context) error {
		_, ok := stream.GetWebSocketHub(c)
		if !ok {
			t.Error("hub not found in handler1")
		}
		return c.JSON(200, map[string]string{"handler": "1"})
	})

	// Handler 2: Check hub exists.
	router.GET("/handler2", func(c *fursy.Context) error {
		_, ok := stream.GetWebSocketHub(c)
		if !ok {
			t.Error("hub not found in handler2")
		}
		return c.JSON(200, map[string]string{"handler": "2"})
	})

	// Test handler1.
	req1 := httptest.NewRequest("GET", "/handler1", http.NoBody)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Errorf("handler1: expected 200, got %d", w1.Code)
	}

	// Test handler2.
	req2 := httptest.NewRequest("GET", "/handler2", http.NoBody)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Errorf("handler2: expected 200, got %d", w2.Code)
	}
}

// Test 8: SSEHub with nested route groups.
func TestSSEHub_NestedGroups(t *testing.T) {
	t.Helper()

	hub := sse.NewHub[string]()
	defer hub.Close()

	router := fursy.New()
	router.Use(stream.SSEHub(hub))

	// Group /api.
	api := router.Group("/api")
	api.GET("/events", func(c *fursy.Context) error {
		_, ok := stream.GetSSEHub[string](c)
		if !ok {
			t.Error("hub not found in /api/events")
		}
		return c.JSON(200, map[string]string{"path": "/api/events"})
	})

	// Nested group /api/v1.
	v1 := api.Group("/v1")
	v1.GET("/notifications", func(c *fursy.Context) error {
		_, ok := stream.GetSSEHub[string](c)
		if !ok {
			t.Error("hub not found in /api/v1/notifications")
		}
		return c.JSON(200, map[string]string{"path": "/api/v1/notifications"})
	})

	// Test /api/events.
	req1 := httptest.NewRequest("GET", "/api/events", http.NoBody)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Errorf("/api/events: expected 200, got %d", w1.Code)
	}

	// Test /api/v1/notifications.
	req2 := httptest.NewRequest("GET", "/api/v1/notifications", http.NoBody)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != 200 {
		t.Errorf("/api/v1/notifications: expected 200, got %d", w2.Code)
	}
}

// Test 9: SSEUpgrade works with different HTTP methods.
// Note: SSE technically works with any HTTP method, though GET is conventional.
func TestSSEUpgrade_DifferentMethods(t *testing.T) {
	t.Helper()

	router := fursy.New()

	// SSE endpoint that accepts POST (unconventional but valid).
	router.POST("/events", func(c *fursy.Context) error {
		return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
			// Send a test event and close.
			return conn.SendData("test")
		})
	})

	req := httptest.NewRequest("POST", "/events", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed (SSE doesn't restrict HTTP methods).
	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify SSE content-type header.
	if ct := w.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("expected text/event-stream, got %q", ct)
	}
}

// Test 10: WebSocketUpgrade with invalid request (no Upgrade header).
func TestWebSocketUpgrade_InvalidRequest(t *testing.T) {
	t.Helper()

	router := fursy.New()

	router.GET("/ws", func(c *fursy.Context) error {
		// WebSocket requires Upgrade header.
		return stream.WebSocketUpgrade(c, func(_ *websocket.Conn) error {
			return nil
		}, nil)
	})

	req := httptest.NewRequest("GET", "/ws", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return error (WebSocket upgrade failed).
	if w.Code == 200 {
		t.Error("expected error for request without Upgrade header, but got 200")
	}
}
