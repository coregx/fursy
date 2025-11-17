package main

import (
	"log/slog"
	"sync"

	"github.com/coregx/fursy"
)

// Database simulates a simple in-memory database.
var (
	users      = make(map[int]*User)
	usersMutex sync.RWMutex
	nextUserID = 1
)

func init() {
	// Create default admin user.
	users[1] = &User{
		ID:       1,
		Email:    "admin@example.com",
		Username: "admin",
		Role:     "admin",
	}
	nextUserID = 2
}

// HandleLogin handles user login and returns JWT token.
func HandleLogin(cfg *Config) fursy.Handler[LoginRequest, LoginResponse] {
	return func(c *fursy.Box[LoginRequest, LoginResponse]) error {
		// Validate request.
		if err := c.Bind(); err != nil {
			return err
		}

		req := c.ReqBody

		// Find user by email (simplified - in real app, check password hash).
		usersMutex.RLock()
		var found *User
		for _, user := range users {
			if user.Email == req.Email {
				found = user
				break
			}
		}
		usersMutex.RUnlock()

		if found == nil {
			slog.Warn("Login failed", "email", req.Email)
			return c.Problem(fursy.Unauthorized("Invalid credentials"))
		}

		// Generate JWT token.
		token, err := GenerateToken(cfg.JWTSecret, found)
		if err != nil {
			slog.Error("Failed to generate token", "error", err)
			return c.Problem(fursy.InternalServerError("Failed to generate token"))
		}

		slog.Info("User logged in", "user_id", found.ID, "username", found.Username)

		resp := LoginResponse{
			Token: token,
			User:  *found,
		}

		return c.JSON(200, resp)
	}
}

// HandleCreateUser handles user creation (admin only).
func HandleCreateUser(c *fursy.Box[CreateUserRequest, UserResponse]) error {
	// Validate request.
	if err := c.Bind(); err != nil {
		return err
	}

	req := c.ReqBody

	// Check if user already exists.
	usersMutex.RLock()
	for _, user := range users {
		if user.Email == req.Email {
			usersMutex.RUnlock()
			return c.Problem(fursy.Conflict("User with this email already exists"))
		}
	}
	usersMutex.RUnlock()

	// Create user.
	usersMutex.Lock()
	user := &User{
		ID:       nextUserID,
		Email:    req.Email,
		Username: req.Username,
		Role:     req.Role,
	}
	users[user.ID] = user
	nextUserID++
	usersMutex.Unlock()

	slog.Info("User created", "user_id", user.ID, "username", user.Username, "role", user.Role)

	resp := UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}

	return c.Created("/api/users/"+string(rune(user.ID)), resp)
}

// HandleGetProfile returns current user's profile.
func HandleGetProfile(c *fursy.Box[fursy.Empty, UserResponse]) error {
	// Get user ID from context (set by AuthMiddleware).
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return c.Problem(fursy.Unauthorized("User not authenticated"))
	}

	// Get user from database.
	usersMutex.RLock()
	user, exists := users[userID]
	usersMutex.RUnlock()

	if !exists {
		return c.Problem(fursy.NotFound("User not found"))
	}

	resp := UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}

	return c.JSON(200, resp)
}

// HandleUpdateProfile updates current user's profile.
func HandleUpdateProfile(c *fursy.Box[UpdateProfileRequest, UserResponse]) error {
	// Validate request.
	if err := c.Bind(); err != nil {
		return err
	}

	req := c.ReqBody

	// Get user ID from context.
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return c.Problem(fursy.Unauthorized("User not authenticated"))
	}

	// Update user.
	usersMutex.Lock()
	user, exists := users[userID]
	if !exists {
		usersMutex.Unlock()
		return c.Problem(fursy.NotFound("User not found"))
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	usersMutex.Unlock()

	slog.Info("Profile updated", "user_id", user.ID, "username", user.Username)

	resp := UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
	}

	return c.JSON(200, resp)
}

// HandleListUsers returns all users (admin only).
func HandleListUsers(c *fursy.Box[fursy.Empty, UserListResponse]) error {
	usersMutex.RLock()
	defer usersMutex.RUnlock()

	resp := UserListResponse{
		Users: make([]UserResponse, 0, len(users)),
		Total: len(users),
		Page:  1,
		Limit: 100,
	}

	for _, user := range users {
		resp.Users = append(resp.Users, UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			Role:     user.Role,
		})
	}

	return c.JSON(200, resp)
}
