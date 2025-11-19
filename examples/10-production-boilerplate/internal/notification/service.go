package notification

import (
	"github.com/coregx/stream/sse"
)

// Service encapsulates notification business logic.
type Service interface {
	// Broadcast sends notification to all connected users.
	Broadcast(message string, notifType NotificationType) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	hub *sse.Hub[*Notification]
}

// NewService creates a new Notification service.
func NewService(hub *sse.Hub[*Notification]) Service {
	return &serviceImpl{hub: hub}
}

// Broadcast sends notification to all connected users.
func (s *serviceImpl) Broadcast(message string, notifType NotificationType) error {
	notification := NewNotification("system", message, notifType)
	return s.hub.Broadcast(notification)
}
