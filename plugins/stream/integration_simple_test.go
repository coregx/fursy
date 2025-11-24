// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stream_test

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/stream"
	"github.com/coregx/stream/sse"
	"github.com/coregx/stream/websocket"
)

// Integration Test 1: Real SSE endpoint with Hub.
func TestSSE_Integration_WithHub(t *testing.T) {
	t.Helper()

	// Setup hub.
	hub := sse.NewHub[string]()
	go hub.Run()
	defer hub.Close()

	// Setup router.
	router := fursy.New()
	router.Use(stream.SSEHub(hub))

	// SSE endpoint.
	router.GET("/events", func(c *fursy.Context) error {
		hub, ok := stream.GetSSEHub[string](c)
		if !ok {
			return c.Problem(fursy.InternalServerError("Hub not configured"))
		}

		return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
			hub.Register(conn)
			defer hub.Unregister(conn)
			<-conn.Done()
			return nil
		})
	})

	// Create test server.
	server := httptest.NewServer(router)
	defer server.Close()

	// Client connects.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/events", http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Verify SSE headers.
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("wrong content-type: %s", resp.Header.Get("Content-Type"))
	}

	// Broadcast message.
	time.Sleep(100 * time.Millisecond)
	hub.Broadcast("test message")
	time.Sleep(100 * time.Millisecond)

	// Read event.
	scanner := bufio.NewScanner(resp.Body)
	var receivedData string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			receivedData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			break
		}
	}

	if receivedData != "test message" {
		t.Errorf("expected 'test message', got %q", receivedData)
	}
}

// Integration Test 2: SSE with JSON events.
func TestSSE_Integration_JSON(t *testing.T) {
	t.Helper()

	type Notification struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}

	hub := sse.NewHub[Notification]()
	go hub.Run()
	defer hub.Close()

	router := fursy.New()
	router.Use(stream.SSEHub(hub))

	router.GET("/events", func(c *fursy.Context) error {
		hub, ok := stream.GetSSEHub[Notification](c)
		if !ok {
			return c.Problem(fursy.InternalServerError("Hub not configured"))
		}

		return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
			hub.Register(conn)
			defer hub.Unregister(conn)
			<-conn.Done()
			return nil
		})
	})

	server := httptest.NewServer(router)
	defer server.Close()

	// Client connects.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/events", http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Broadcast JSON message.
	time.Sleep(100 * time.Millisecond)
	hub.BroadcastJSON(Notification{Type: "info", Message: "hello"})
	time.Sleep(100 * time.Millisecond)

	// Read event.
	scanner := bufio.NewScanner(resp.Body)
	var receivedData string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			receivedData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			break
		}
	}

	expectedJSON := `{"type":"info","message":"hello"}`
	if receivedData != expectedJSON {
		t.Errorf("expected %q, got %q", expectedJSON, receivedData)
	}
}

// Integration Test 3: WebSocket Hub availability (no actual WS client - just test hub access).
func TestWebSocket_Integration_HubAvailability(t *testing.T) {
	t.Helper()

	hub := websocket.NewHub()
	go hub.Run()
	defer hub.Close()

	router := fursy.New()
	router.Use(stream.WebSocketHub(hub))

	// Simple health endpoint that checks hub availability.
	router.GET("/health", func(c *fursy.Context) error {
		hub, ok := stream.GetWebSocketHub(c)
		if !ok {
			return c.Problem(fursy.InternalServerError("Hub not configured"))
		}

		return c.JSON(200, map[string]any{
			"status":  "ok",
			"clients": hub.ClientCount(),
		})
	})

	// Test health endpoint.
	req := httptest.NewRequest("GET", "/health", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Verify JSON response contains "status": "ok".
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("response missing status: %s", w.Body.String())
	}
}
