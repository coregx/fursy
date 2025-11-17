package main

import (
	"errors"
	"strconv"

	"github.com/coregx/fursy"
)

// Handlers contains all HTTP handlers for user operations.
type Handlers struct {
	db *Database
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(db *Database) *Handlers {
	return &Handlers{db: db}
}

// CreateUser handles POST /users - create a new user.
func (h *Handlers) CreateUser(c *fursy.Box[CreateUserRequest, UserResponse]) error {
	// Bind and validate request body automatically.
	if err := c.Bind(); err != nil {
		return err
	}

	// Create user in database.
	user, err := h.db.Create(c.ReqBody)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			return c.Problem(fursy.Conflict("Username already exists"))
		}
		return c.Problem(fursy.InternalServerError("Failed to create user"))
	}

	// Convert to response.
	resp := UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		FullName: user.FullName,
		Age:      user.Age,
	}

	// Return 201 Created with Location header.
	return c.Created("/users/"+strconv.Itoa(user.ID), resp)
}

// GetUser handles GET /users/:id - retrieve a user by ID.
func (h *Handlers) GetUser(c *fursy.Box[fursy.Empty, UserResponse]) error {
	// Get ID from URL parameter.
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Problem(fursy.BadRequest("Invalid user ID"))
	}

	// Get user from database.
	user, err := h.db.GetByID(id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return c.Problem(fursy.NotFound("User not found"))
		}
		return c.Problem(fursy.InternalServerError("Failed to retrieve user"))
	}

	// Convert to response.
	resp := UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		FullName: user.FullName,
		Age:      user.Age,
	}

	return c.JSON(200, resp)
}

// ListUsers handles GET /users - retrieve all users.
func (h *Handlers) ListUsers(c *fursy.Box[fursy.Empty, UserListResponse]) error {
	// Get all users from database.
	users := h.db.GetAll()

	// Convert to response.
	resp := UserListResponse{
		Users: make([]UserResponse, 0, len(users)),
		Total: len(users),
	}

	for _, user := range users {
		resp.Users = append(resp.Users, UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			FullName: user.FullName,
			Age:      user.Age,
		})
	}

	return c.JSON(200, resp)
}

// UpdateUser handles PUT /users/:id - update a user.
func (h *Handlers) UpdateUser(c *fursy.Box[UpdateUserRequest, UserResponse]) error {
	// Get ID from URL parameter.
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Problem(fursy.BadRequest("Invalid user ID"))
	}

	// Bind and validate request body automatically.
	if err := c.Bind(); err != nil {
		return err
	}

	// Update user in database.
	user, err := h.db.Update(id, c.ReqBody)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return c.Problem(fursy.NotFound("User not found"))
		}
		return c.Problem(fursy.InternalServerError("Failed to update user"))
	}

	// Convert to response.
	resp := UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		FullName: user.FullName,
		Age:      user.Age,
	}

	return c.JSON(200, resp)
}

// DeleteUser handles DELETE /users/:id - delete a user.
func (h *Handlers) DeleteUser(c *fursy.Box[fursy.Empty, EmptyResponse]) error {
	// Get ID from URL parameter.
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Problem(fursy.BadRequest("Invalid user ID"))
	}

	// Delete user from database.
	if err := h.db.Delete(id); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return c.Problem(fursy.NotFound("User not found"))
		}
		return c.Problem(fursy.InternalServerError("Failed to delete user"))
	}

	// Return 204 No Content.
	return c.NoContentSuccess()
}
