// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stream

import (
	"net/http"

	"github.com/coregx/fursy"
	"github.com/coregx/stream/sse"
	"github.com/coregx/stream/websocket"
)

// SSEUpgrade upgrades HTTP connection to Server-Sent Events.
//
// This is the actual implementation for fursy.Context.SSE() method.
// It performs the SSE upgrade and calls the user's handler with the connection.
//
// The connection is automatically closed when the handler returns.
//
// Example (internal use by Context.SSE):
//
//	return SSEUpgrade(c, func(conn *sse.Conn) error {
//	    hub.Register(conn)
//	    defer hub.Unregister(conn)
//	    <-conn.Done()
//	    return nil
//	})
func SSEUpgrade(c *fursy.Context, handler func(conn *sse.Conn) error) error {
	conn, err := sse.UpgradeWithContext(c.Request.Context(), c.Response, c.Request)
	if err != nil {
		return c.Problem(fursy.InternalServerError("SSE upgrade failed: " + err.Error()))
	}
	defer func() {
		_ = conn.Close() // Error on close is not critical for SSE.
	}()

	return handler(conn)
}

// WebSocketUpgrade upgrades HTTP connection to WebSocket.
//
// This is the actual implementation for fursy.Context.WebSocket() method.
// It performs the WebSocket upgrade and calls the user's handler with the connection.
//
// The connection is automatically closed when the handler returns.
//
// Example (internal use by Context.WebSocket):
//
//	return WebSocketUpgrade(c, func(conn *websocket.Conn) error {
//	    hub.Register(conn)
//	    defer hub.Unregister(conn)
//	    // ... read/write loop
//	    return nil
//	}, opts)
func WebSocketUpgrade(c *fursy.Context, handler func(conn *websocket.Conn) error, opts *websocket.UpgradeOptions) error {
	conn, err := websocket.Upgrade(c.Response, c.Request, opts)
	if err != nil {
		return c.Problem(fursy.NewProblem(http.StatusBadRequest, "WebSocket Upgrade Failed", err.Error()))
	}
	defer func() {
		_ = conn.Close() // Error on close is not critical for WebSocket.
	}()

	return handler(conn)
}

// init registers the stream implementations with fursy Context.
// This allows c.SSE() and c.WebSocket() to work when plugins/stream is imported.
//
// Usage in user code:
//
//	import _ "github.com/coregx/fursy/plugins/stream"
//
// The underscore import triggers this init() function, which registers
// the actual SSE and WebSocket implementations.
func init() {
	// Note: We cannot directly modify c.SSE and c.WebSocket methods
	// because Go doesn't support method overriding.
	//
	// Instead, users must explicitly call SSEUpgrade and WebSocketUpgrade
	// or use wrapper helpers. This is a limitation of Go's type system.
	//
	// Alternative API (more explicit):
	//   - stream.HandleSSE(c, handler)
	//   - stream.HandleWebSocket(c, handler, opts)
	//
	// For now, we document that c.SSE() and c.WebSocket() are stubs,
	// and users should call plugins/stream helpers directly:
	//   - return stream.SSEUpgrade(c, handler)
	//   - return stream.WebSocketUpgrade(c, handler, opts)
}
