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
// It performs the SSE upgrade and calls the user's handler with the connection.
//
// The connection is automatically closed when the handler returns.
//
// Example:
//
//	return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
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
// It performs the WebSocket upgrade and calls the user's handler with the connection.
//
// The connection is automatically closed when the handler returns.
//
// Example:
//
//	return stream.WebSocketUpgrade(c, func(conn *websocket.Conn) error {
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

// init is kept for potential future registration hooks.
//
// Users should call plugins/stream helpers directly:
//   - return stream.SSEUpgrade(c, handler)
//   - return stream.WebSocketUpgrade(c, handler, opts)
func init() {
}
