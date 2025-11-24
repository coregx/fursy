// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json/v2"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
	"github.com/coregx/fursy/plugins/stream"
	"github.com/coregx/stream/sse"
)

// Notification represents a server-sent event.
type Notification struct {
	Type    string    `json:"type"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func main() {
	// Create SSE Hub for broadcasting notifications.
	hub := sse.NewHub[Notification]()
	go hub.Run()
	defer hub.Close()

	// Create fursy router.
	router := fursy.New()

	// Middleware.
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Register SSE hub middleware.
	router.Use(stream.SSEHub(hub))

	// SSE endpoint - clients connect here to receive events.
	router.GET("/events", handleSSE)

	// POST endpoint - send notification to all connected clients.
	router.POST("/notify", handleNotify)

	// Start periodic notifications (every 5 seconds).
	go periodicNotifications(hub)

	slog.Info("SSE server starting",
		"port", 8080,
		"endpoints", []string{
			"GET  /events  - SSE connection (curl -N http://localhost:8080/events)",
			"POST /notify  - Send notification",
		},
	)

	log.Fatal(http.ListenAndServe(":8080", router))
}

// handleSSE handles SSE connections.
//
// Clients connect with:
//
//	curl -N http://localhost:8080/events
func handleSSE(c *fursy.Context) error {
	hub, ok := stream.GetSSEHub[Notification](c)
	if !ok {
		return c.Problem(fursy.InternalServerError("Hub not configured"))
	}

	slog.Info("SSE client connected", "remote_addr", c.Request.RemoteAddr)

	return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
		// Register client to hub.
		hub.Register(conn)
		defer func() {
			hub.Unregister(conn)
			slog.Info("SSE client disconnected",
				"remote_addr", c.Request.RemoteAddr,
				"clients_remaining", hub.Clients(),
			)
		}()

		// Send welcome notification.
		if err := conn.SendJSON(Notification{
			Type:    "info",
			Message: "Connected to SSE notification service",
			Time:    time.Now(),
		}); err != nil {
			return err
		}

		// Wait until client disconnects.
		<-conn.Done()
		return nil
	})
}

// handleNotify handles POST /notify - broadcast notification to all clients.
//
// Example:
//
//	curl -X POST http://localhost:8080/notify \
//	  -H "Content-Type: application/json" \
//	  -d '{"type":"alert","message":"Important update!"}'
func handleNotify(c *fursy.Context) error {
	hub, ok := stream.GetSSEHub[Notification](c)
	if !ok {
		return c.Problem(fursy.InternalServerError("Hub not configured"))
	}

	// Parse notification from request body.
	var notification Notification
	if err := json.UnmarshalRead(c.Request.Body, &notification); err != nil {
		return c.Problem(fursy.BadRequest("Invalid request body: " + err.Error()))
	}

	// Set timestamp.
	notification.Time = time.Now()

	// Broadcast to all connected clients.
	hub.BroadcastJSON(notification)

	slog.Info("Notification sent",
		"type", notification.Type,
		"message", notification.Message,
		"clients", hub.Clients(),
	)

	return c.JSON(200, map[string]any{
		"status":  "sent",
		"clients": hub.Clients(),
	})
}

// periodicNotifications sends periodic notifications every 5 seconds.
func periodicNotifications(hub *sse.Hub[Notification]) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	count := 1
	for range ticker.C {
		hub.BroadcastJSON(Notification{
			Type:    "info",
			Message: "Periodic update #" + string(rune('0'+count)),
			Time:    time.Now(),
		})
		slog.Info("Periodic notification sent", "count", count, "clients", hub.Clients())
		count++
	}
}
