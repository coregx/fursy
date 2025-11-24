// Package chat implements the Chat bounded context.
package chat

import "time"

// Message represents a chat message.
type Message struct {
	ID        string
	UserID    string
	UserName  string
	Content   string
	CreatedAt time.Time
}

// NewMessage creates a new chat message.
func NewMessage(userID, userName, content string) *Message {
	return &Message{
		ID:        generateID(),
		UserID:    userID,
		UserName:  userName,
		Content:   content,
		CreatedAt: time.Now(),
	}
}

func generateID() string {
	return time.Now().Format("20060102150405")
}
