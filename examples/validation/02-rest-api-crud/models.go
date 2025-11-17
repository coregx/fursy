// Package main demonstrates REST API CRUD operations with validation.
package main

// User represents a user in the system.
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
}

// CreateUserRequest represents user creation request with validation.
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Age      int    `json:"age" validate:"required,gte=18,lte=120"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// UpdateUserRequest represents user update request with validation.
// All fields are optional for partial updates.
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=50,alphanum"`
	FullName string `json:"full_name,omitempty" validate:"omitempty,min=2,max=100"`
	Age      int    `json:"age,omitempty" validate:"omitempty,gte=18,lte=120"`
}

// UserResponse represents a single user response.
type UserResponse struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
}

// UserListResponse represents a list of users response.
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}

// EmptyResponse represents an empty response for DELETE operations.
type EmptyResponse struct{}
