package notification

// BroadcastRequest is the payload for broadcasting a notification.
type BroadcastRequest struct {
	Message string           `json:"message" validate:"required"`
	Type    NotificationType `json:"type"`
}

// MessageResponse is a simple message response.
type MessageResponse struct {
	Message string `json:"message"`
}
