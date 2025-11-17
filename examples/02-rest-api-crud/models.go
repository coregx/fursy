package main

// User represents a user in the system.
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
}

// CreateUserRequest represents user creation request.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
}

// UpdateUserRequest represents user update request.
// All fields are optional for partial updates.
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	FullName string `json:"full_name,omitempty"`
	Age      int    `json:"age,omitempty"`
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
