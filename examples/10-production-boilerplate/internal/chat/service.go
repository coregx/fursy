package chat

import (
	"context"

	"github.com/coregx/stream/websocket"
)

// Service encapsulates chat business logic.
type Service interface {
	// Broadcast sends message to all connected users.
	Broadcast(ctx context.Context, userID, userName, content string) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	hub *websocket.Hub
}

// NewService creates a new Chat service.
func NewService(hub *websocket.Hub) Service {
	return &serviceImpl{hub: hub}
}

// Broadcast sends message to all connected users.
func (s *serviceImpl) Broadcast(ctx context.Context, userID, userName, content string) error {
	message := NewMessage(userID, userName, content)
	s.hub.Broadcast([]byte(message.Content))
	return nil
}
