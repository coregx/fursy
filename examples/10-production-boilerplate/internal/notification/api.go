package notification

import (
	"encoding/json"
	"net/http"

	"github.com/coregx/fursy"
	"github.com/coregx/stream/sse"
)

// API handles HTTP requests for notification endpoints.
type API struct {
	service Service
	hub     *sse.Hub[*Notification]
}

// NewAPI creates a new Notification API handler.
func NewAPI(service Service, hub *sse.Hub[*Notification]) *API {
	return &API{
		service: service,
		hub:     hub,
	}
}

// RegisterRoutes registers notification routes.
func (api *API) RegisterRoutes(r *fursy.Router, authMiddleware fursy.HandlerFunc) {
	protected := r.Group("/api")
	protected.Use(authMiddleware)
	{
		protected.GET("/notifications/stream", api.stream)
		protected.POST("/notifications/broadcast", api.broadcast)
	}
}

// stream handles SSE notification stream.
func (api *API) stream(c *fursy.Context) error {
	// Upgrade to SSE connection
	conn, err := sse.Upgrade(c.Response, c.Request)
	if err != nil {
		return err
	}

	// Register connection with hub
	api.hub.Register(conn)
	defer api.hub.Unregister(conn)

	// Keep connection alive
	<-c.Request.Context().Done()
	return nil
}

// broadcast broadcasts notification to all connected clients (admin only).
func (api *API) broadcast(c *fursy.Context) error {
	var req struct {
		Message string           `json:"message"`
		Type    NotificationType `json:"type"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		return c.Problem(fursy.BadRequest("Invalid request: " + err.Error()))
	}

	if req.Message == "" {
		return c.Problem(fursy.BadRequest("Message is required"))
	}

	if req.Type == "" {
		req.Type = NotificationTypeInfo
	}

	if err := api.service.Broadcast(req.Message, req.Type); err != nil {
		return c.Problem(fursy.InternalServerError("Broadcast failed: " + err.Error()))
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Notification broadcasted successfully",
	})
}
