// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package stream provides SSE and WebSocket integration middleware for fursy router.
//
// This plugin enables seamless integration of Server-Sent Events (SSE) and WebSocket
// connections from github.com/coregx/stream library into fursy handlers.
//
// Features:
//   - SSEHub[T] middleware for sharing SSE hubs across handlers
//   - WebSocketHub middleware for sharing WebSocket hubs
//   - Context helper methods: c.SSE() and c.WebSocket()
//   - Type-safe hub retrieval with generics
//
// Example SSE usage:
//
//	hub := sse.NewHub[Notification]()
//	go hub.Run()
//
//	router := fursy.New()
//	router.Use(stream.SSEHub(hub))
//
//	router.GET("/events", func(c *fursy.Context) error {
//	    hub, _ := stream.GetSSEHub[Notification](c)
//	    return c.SSE(func(conn *sse.Conn) error {
//	        hub.Register(conn)
//	        defer hub.Unregister(conn)
//	        <-conn.Done()
//	        return nil
//	    })
//	})
package stream

import (
	"context"

	"github.com/coregx/fursy"
	"github.com/coregx/stream/sse"
)

// contextKey for storing hubs in context.
type contextKey int

const (
	sseHubKey contextKey = iota
	wsHubKey
)

// SSEHub creates a middleware that provides SSE Hub in request context.
//
// The hub is stored in the request context and can be retrieved using GetSSEHub.
// This allows sharing the same hub across multiple handlers and middleware.
//
// Type parameter T specifies the type of events that will be broadcast through the hub.
//
// Example:
//
//	type Notification struct {
//	    Type    string `json:"type"`
//	    Message string `json:"message"`
//	}
//
//	hub := sse.NewHub[Notification]()
//	go hub.Run()
//	defer hub.Close()
//
//	router := fursy.New()
//	router.Use(stream.SSEHub(hub))
//
//	// Handlers can now retrieve the hub using GetSSEHub
func SSEHub[T any](hub *sse.Hub[T]) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		ctx := context.WithValue(c.Request.Context(), sseHubKey, hub)
		c.Request = c.Request.WithContext(ctx)
		return c.Next()
	}
}

// GetSSEHub retrieves SSE hub from request context.
//
// Returns (hub, true) if hub is found and has the correct type T.
// Returns (nil, false) if hub not found or type mismatch.
//
// Type parameter T must match the type used when creating the hub with SSEHub middleware.
//
// Example:
//
//	hub, ok := stream.GetSSEHub[Notification](c)
//	if !ok {
//	    return c.Problem(fursy.InternalServerError("Hub not configured"))
//	}
//
//	// Use hub to broadcast events
//	hub.BroadcastJSON(Notification{Type: "info", Message: "Hello"})
func GetSSEHub[T any](c *fursy.Context) (*sse.Hub[T], bool) {
	hub, ok := c.Request.Context().Value(sseHubKey).(*sse.Hub[T])
	return hub, ok
}
