package auth

import "context"

// Context keys for storing user info.
type contextKey int

const (
	userIDKey contextKey = iota
	userRoleKey
)

// WithUserID stores user ID in context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID retrieves user ID from context.
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// WithUserRole stores user role in context.
func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, userRoleKey, role)
}

// GetUserRole retrieves user role from context.
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(userRoleKey).(string); ok {
		return role
	}
	return ""
}
