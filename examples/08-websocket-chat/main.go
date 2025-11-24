// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
	"github.com/coregx/fursy/plugins/stream"
	"github.com/coregx/stream/websocket"
)

// ChatMessage represents a chat message.
type ChatMessage struct {
	User    string    `json:"user"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func main() {
	// Create WebSocket Hub for chat broadcasting.
	hub := websocket.NewHub()
	go hub.Run()
	defer hub.Close()

	// Create fursy router.
	router := fursy.New()

	// Middleware.
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Register WebSocket hub middleware.
	router.Use(stream.WebSocketHub(hub))

	// WebSocket endpoint - clients connect here for chat.
	router.GET("/ws", handleWebSocket)

	// Health check endpoint.
	router.GET("/health", handleHealth)

	slog.Info("WebSocket chat server starting",
		"port", 8080,
		"endpoints", []string{
			"GET  /ws      - WebSocket connection (wscat -c ws://localhost:8080/ws)",
			"GET  /health  - Health check",
		},
	)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}

// handleWebSocket handles WebSocket connections.
//
// Clients connect with:
//
//	wscat -c ws://localhost:8080/ws
//
// Or via browser WebSocket API:
//
//	const ws = new WebSocket('ws://localhost:8080/ws');
func handleWebSocket(c *fursy.Context) error {
	hub, ok := stream.GetWebSocketHub(c)
	if !ok {
		return c.Problem(fursy.InternalServerError("Hub not configured"))
	}

	slog.Info("WebSocket client connecting", "remote_addr", c.Request.RemoteAddr)

	return stream.WebSocketUpgrade(c, func(conn *websocket.Conn) error {
		// Register client to hub.
		hub.Register(conn)
		defer func() {
			hub.Unregister(conn)
			slog.Info("WebSocket client disconnected",
				"remote_addr", c.Request.RemoteAddr,
				"clients_remaining", hub.ClientCount(),
			)

			// Broadcast disconnect message.
			disconnectMsg := ChatMessage{
				User:    "System",
				Message: "A user has left the chat",
				Time:    time.Now(),
			}
			hub.BroadcastJSON(disconnectMsg)
		}()

		// Send welcome message.
		welcomeMsg := ChatMessage{
			User:    "System",
			Message: "Welcome to the chat! Type a message to start.",
			Time:    time.Now(),
		}
		if err := conn.WriteJSON(welcomeMsg); err != nil {
			return err
		}

		// Broadcast join message to others.
		joinMsg := ChatMessage{
			User:    "System",
			Message: "A new user has joined the chat",
			Time:    time.Now(),
		}
		hub.BroadcastJSON(joinMsg)

		slog.Info("WebSocket client connected",
			"remote_addr", c.Request.RemoteAddr,
			"total_clients", hub.ClientCount(),
		)

		// Read loop - receive messages and broadcast to all.
		for {
			var msg ChatMessage
			if err := conn.ReadJSON(&msg); err != nil {
				// Client disconnected or error occurred.
				return err
			}

			// Set timestamp.
			msg.Time = time.Now()

			// Broadcast to all clients (including sender).
			hub.BroadcastJSON(msg)

			slog.Info("Chat message",
				"user", msg.User,
				"message", msg.Message,
				"clients", hub.ClientCount(),
			)
		}
	}, nil)
}

// handleHealth returns server health status and connected clients count.
func handleHealth(c *fursy.Context) error {
	hub, ok := stream.GetWebSocketHub(c)
	if !ok {
		return c.Problem(fursy.InternalServerError("Hub not configured"))
	}

	return c.JSON(200, map[string]any{
		"status":  "ok",
		"clients": hub.ClientCount(),
		"time":    time.Now(),
	})
}
