// Package notification implements the Notification bounded context.
package notification

import "time"

// Notification represents a notification message.
type Notification struct {
	ID        string
	UserID    string
	Message   string
	Type      NotificationType
	CreatedAt time.Time
}

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
)

// NewNotification creates a new notification.
func NewNotification(userID, message string, notifType NotificationType) *Notification {
	return &Notification{
		ID:        generateID(),
		UserID:    userID,
		Message:   message,
		Type:      notifType,
		CreatedAt: time.Now(),
	}
}

func generateID() string {
	return time.Now().Format("20060102150405")
}
