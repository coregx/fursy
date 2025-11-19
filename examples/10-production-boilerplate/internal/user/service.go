package user

import (
	"context"
	"errors"

	"example.com/production-boilerplate/internal/shared/auth"
)

// Service encapsulates user business logic / use cases.
type Service interface {
	// Register creates a new user account.
	Register(ctx context.Context, req RegisterRequest) (*User, error)

	// Login authenticates user and returns JWT token.
	Login(ctx context.Context, req LoginRequest) (string, error)

	// GetProfile retrieves user profile.
	GetProfile(ctx context.Context, userID string) (*User, error)

	// UpdateProfile updates user profile.
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*User, error)

	// ChangePassword changes user password.
	ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error

	// ListUsers lists all users (admin only).
	ListUsers(ctx context.Context, offset, limit int) ([]*User, int, error)

	// BanUser bans user account (admin only).
	BanUser(ctx context.Context, userID string) error

	// PromoteToAdmin promotes user to admin (admin only).
	PromoteToAdmin(ctx context.Context, userID string) error
}

// serviceImpl implements Service.
type serviceImpl struct {
	repo   Repository
	jwtSvc *auth.JWTService
}

// NewService creates a new User service.
func NewService(repo Repository, jwtSvc *auth.JWTService) Service {
	return &serviceImpl{
		repo:   repo,
		jwtSvc: jwtSvc,
	}
}

// Register creates a new user account.
func (s *serviceImpl) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if email already exists
	email, _ := NewEmail(req.Email)
	_, err := s.repo.GetByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	// Create new user (rich model validates itself)
	user, err := NewUser(req.Email, req.Password, req.Name)
	if err != nil {
		return nil, err
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates user and returns JWT token.
func (s *serviceImpl) Login(ctx context.Context, req LoginRequest) (string, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return "", err
	}

	// Get user by email
	email, _ := NewEmail(req.Email)
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// Check password (business logic in entity!)
	if !user.Password.Check(req.Password) {
		return "", ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive() {
		return "", ErrUserInactive
	}

	// Generate JWT token
	token, err := s.jwtSvc.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetProfile retrieves user profile.
func (s *serviceImpl) GetProfile(ctx context.Context, userID string) (*User, error) {
	return s.repo.Get(ctx, userID)
}

// UpdateProfile updates user profile.
func (s *serviceImpl) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Get user
	user, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update profile (business logic in entity!)
	if err := user.UpdateProfile(req.Name); err != nil {
		return nil, err
	}

	// Save changes
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword changes user password.
func (s *serviceImpl) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}

	// Get user
	user, err := s.repo.Get(ctx, userID)
	if err != nil {
		return err
	}

	// Change password (business logic in entity!)
	if err := user.ChangePassword(req.OldPassword, req.NewPassword); err != nil {
		return err
	}

	// Save changes
	return s.repo.Update(ctx, user)
}

// ListUsers lists all users (admin only).
func (s *serviceImpl) ListUsers(ctx context.Context, offset, limit int) ([]*User, int, error) {
	users, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// BanUser bans user account (admin only).
func (s *serviceImpl) BanUser(ctx context.Context, userID string) error {
	// Get user
	user, err := s.repo.Get(ctx, userID)
	if err != nil {
		return err
	}

	// Ban user (business logic in entity!)
	if err := user.Ban(); err != nil {
		return err
	}

	// Save changes
	return s.repo.Update(ctx, user)
}

// PromoteToAdmin promotes user to admin (admin only).
func (s *serviceImpl) PromoteToAdmin(ctx context.Context, userID string) error {
	// Get user
	user, err := s.repo.Get(ctx, userID)
	if err != nil {
		return err
	}

	// Promote to admin (business logic in entity!)
	if err := user.PromoteToAdmin(); err != nil {
		return err
	}

	// Save changes
	return s.repo.Update(ctx, user)
}

// Domain errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user account is inactive")
)
