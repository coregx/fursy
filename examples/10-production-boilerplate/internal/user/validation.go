package user

import (
	"errors"
	"time"
)

// RegisterRequest represents user registration request.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// Validate validates RegisterRequest.
func (r RegisterRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// LoginRequest represents login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate validates LoginRequest.
func (r LoginRequest) Validate() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// UpdateProfileRequest represents profile update request.
type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// Validate validates UpdateProfileRequest.
func (r UpdateProfileRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// ChangePasswordRequest represents password change request.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Validate validates ChangePasswordRequest.
func (r ChangePasswordRequest) Validate() error {
	if r.OldPassword == "" {
		return errors.New("old password is required")
	}
	if r.NewPassword == "" {
		return errors.New("new password is required")
	}
	if len(r.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}
	return nil
}

// UserResponse represents user in API response (separate from domain entity!).
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ToResponse converts User entity to UserResponse.
func ToResponse(user *User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Email:     user.Email.String(),
		Name:      user.Name,
		Role:      string(user.Role),
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}
