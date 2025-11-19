// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package stream

import (
	"context"

	"github.com/coregx/fursy"
	"github.com/coregx/stream/websocket"
)

// WebSocketHub creates a middleware that provides WebSocket Hub in request context.
//
// The hub is stored in the request context and can be retrieved using GetWebSocketHub.
// This allows sharing the same hub across multiple handlers and middleware.
//
// Example:
//
//	hub := websocket.NewHub()
//	go hub.Run()
//	defer hub.Close()
//
//	router := fursy.New()
//	router.Use(stream.WebSocketHub(hub))
//
//	// Handlers can now retrieve the hub using GetWebSocketHub
func WebSocketHub(hub *websocket.Hub) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		ctx := context.WithValue(c.Request.Context(), wsHubKey, hub)
		c.Request = c.Request.WithContext(ctx)
		return c.Next()
	}
}

// GetWebSocketHub retrieves WebSocket hub from request context.
//
// Returns (hub, true) if hub is found.
// Returns (nil, false) if hub not found.
//
// Example:
//
//	hub, ok := stream.GetWebSocketHub(c)
//	if !ok {
//	    return c.Problem(fursy.InternalServerError("Hub not configured"))
//	}
//
//	// Use hub to broadcast messages
//	hub.Broadcast([]byte("Hello, clients!"))
func GetWebSocketHub(c *fursy.Context) (*websocket.Hub, bool) {
	hub, ok := c.Request.Context().Value(wsHubKey).(*websocket.Hub)
	return hub, ok
}
