package chat

import (
	"log/slog"

	"example.com/production-boilerplate/internal/shared/auth"
	"github.com/coregx/fursy"
	"github.com/coregx/stream/websocket"
)

// API handles HTTP requests for chat endpoints.
type API struct {
	service Service
	hub     *websocket.Hub
}

// NewAPI creates a new Chat API handler.
func NewAPI(service Service, hub *websocket.Hub) *API {
	return &API{
		service: service,
		hub:     hub,
	}
}

// RegisterRoutes registers chat routes.
func (api *API) RegisterRoutes(r *fursy.Router, authMiddleware fursy.HandlerFunc) {
	protected := r.Group("/api")
	protected.Use(authMiddleware)
	{
		protected.GET("/chat/ws", api.websocket)
	}
}

// websocket handles WebSocket chat connection.
func (api *API) websocket(c *fursy.Context) error {
	// Get user info from context
	userID := auth.GetUserID(c.Request.Context())

	// Upgrade to WebSocket connection
	conn, err := websocket.Upgrade(c.Response, c.Request, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return err
	}
	defer conn.Close()

	// Register connection with hub
	api.hub.Register(conn)
	defer api.hub.Unregister(conn)

	// Handle incoming messages
	for {
		_, message, err := conn.Read()
		if err != nil {
			slog.Error("WebSocket read error", "error", err)
			break
		}

		// Broadcast message to all clients
		api.hub.Broadcast(message)
	}

	slog.Info("WebSocket connection closed", "user_id", userID)
	return nil
}
