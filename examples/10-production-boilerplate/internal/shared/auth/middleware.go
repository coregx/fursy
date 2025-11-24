package auth

import (
	"strings"

	"github.com/coregx/fursy"
)

// Middleware creates JWT authentication middleware.
func Middleware(jwtSvc *JWTService) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		// Get Authorization header
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			return c.Problem(fursy.Unauthorized("Missing Authorization header"))
		}

		// Check Bearer format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Problem(fursy.Unauthorized("Invalid Authorization header format"))
		}

		// Validate token
		claims, err := jwtSvc.ValidateToken(parts[1])
		if err != nil {
			return c.Problem(fursy.Unauthorized("Invalid or expired token"))
		}

		// Store user info in context
		ctx := WithUserID(c.Request.Context(), claims.UserID)
		ctx = WithUserRole(ctx, claims.Role)
		c.Request = c.Request.WithContext(ctx)

		return c.Next()
	}
}

// RequireRole creates a middleware that checks user role.
func RequireRole(role string) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		userRole := GetUserRole(c.Request.Context())
		if userRole != role {
			return c.Problem(fursy.Forbidden("You don't have permission to access this resource"))
		}
		return c.Next()
	}
}
