package main

import (
	"strings"
	"time"

	"github.com/coregx/fursy"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates JWT tokens.
func AuthMiddleware(secret string) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		// Get Authorization header.
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			return c.Problem(fursy.Unauthorized("Missing authorization header"))
		}

		// Extract token from "Bearer <token>".
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Problem(fursy.Unauthorized("Invalid authorization header format"))
		}

		tokenString := parts[1]

		// Parse and validate token.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Problem(fursy.Unauthorized("Invalid or expired token"))
		}

		// Extract claims.
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Store user info in context.
			c.Set("user_id", int(claims["user_id"].(float64)))
			c.Set("username", claims["username"].(string))
			c.Set("role", claims["role"].(string))
		}

		return c.Next()
	}
}

// RequireRole checks if user has required role.
func RequireRole(role string) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		userRole, ok := c.Get("role").(string)
		if !ok || userRole != role {
			return c.Problem(fursy.Forbidden("Insufficient permissions"))
		}
		return c.Next()
	}
}

// GenerateToken generates a JWT token for user.
func GenerateToken(secret string, user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
