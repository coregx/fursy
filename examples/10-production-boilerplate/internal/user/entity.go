// Package user implements the User bounded context with rich domain models.
package user

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system (RICH MODEL).
// Business logic lives HERE, not in service layer.
type User struct {
	ID        string
	Email     Email    // Value Object
	Password  Password // Value Object
	Name      string
	Role      Role   // Value Object
	Status    Status // Value Object
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Email is a Value Object that ensures email validity.
type Email struct {
	value string
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail creates a validated Email.
func NewEmail(email string) (Email, error) {
	if email == "" {
		return Email{}, errors.New("email cannot be empty")
	}
	if !emailRegex.MatchString(email) {
		return Email{}, errors.New("invalid email format")
	}
	return Email{value: email}, nil
}

// String returns the email string.
func (e Email) String() string {
	return e.value
}

// Password is a Value Object for password handling.
type Password struct {
	hash string
}

// NewPassword creates a hashed password.
func NewPassword(plaintext string) (Password, error) {
	if len(plaintext) < 8 {
		return Password{}, errors.New("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, err
	}

	return Password{hash: string(hash)}, nil
}

// Check verifies password against hash.
func (p Password) Check(plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintext)) == nil
}

// Hash returns the password hash (for persistence).
func (p Password) Hash() string {
	return p.hash
}

// FromHash creates Password from existing hash (for loading from DB).
func FromHash(hash string) Password {
	return Password{hash: hash}
}

// Role represents user role (Value Object).
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// Status represents user account status (Value Object).
type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
	StatusBanned   Status = "banned"
)

// NewUser creates a new User with validation (Factory Method).
func NewUser(email, password, name string) (*User, error) {
	// Validate email
	emailVO, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	// Validate password
	passwordVO, err := NewPassword(password)
	if err != nil {
		return nil, err
	}

	// Validate name
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	now := time.Now()
	return &User{
		ID:        uuid.New().String(),
		Email:     emailVO,
		Password:  passwordVO,
		Name:      name,
		Role:      RoleUser, // Default role
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Activate activates user account (Business Logic in Entity!).
func (u *User) Activate() error {
	if u.Status == StatusActive {
		return errors.New("user is already active")
	}
	u.Status = StatusActive
	u.UpdatedAt = time.Now()
	return nil
}

// Deactivate deactivates user account (Business Logic in Entity!).
func (u *User) Deactivate() error {
	if u.Status == StatusInactive {
		return errors.New("user is already inactive")
	}
	u.Status = StatusInactive
	u.UpdatedAt = time.Now()
	return nil
}

// Ban bans user account (Business Logic in Entity!).
func (u *User) Ban() error {
	if u.Status == StatusBanned {
		return errors.New("user is already banned")
	}
	u.Status = StatusBanned
	u.UpdatedAt = time.Now()
	return nil
}

// PromoteToAdmin promotes user to admin (Business Logic in Entity!).
func (u *User) PromoteToAdmin() error {
	if u.Role == RoleAdmin {
		return errors.New("user is already admin")
	}
	u.Role = RoleAdmin
	u.UpdatedAt = time.Now()
	return nil
}

// DemoteToUser demotes admin to regular user (Business Logic in Entity!).
func (u *User) DemoteToUser() error {
	if u.Role == RoleUser {
		return errors.New("user is already a regular user")
	}
	u.Role = RoleUser
	u.UpdatedAt = time.Now()
	return nil
}

// ChangePassword changes user password (Business Logic in Entity!).
func (u *User) ChangePassword(oldPassword, newPassword string) error {
	// Verify old password
	if !u.Password.Check(oldPassword) {
		return errors.New("invalid old password")
	}

	// Create new password
	newPasswordVO, err := NewPassword(newPassword)
	if err != nil {
		return err
	}

	u.Password = newPasswordVO
	u.UpdatedAt = time.Now()
	return nil
}

// UpdateProfile updates user profile information (Business Logic in Entity!).
func (u *User) UpdateProfile(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	u.Name = name
	u.UpdatedAt = time.Now()
	return nil
}

// IsActive checks if user is active.
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsAdmin checks if user is admin.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsBanned checks if user is banned.
func (u *User) IsBanned() bool {
	return u.Status == StatusBanned
}
