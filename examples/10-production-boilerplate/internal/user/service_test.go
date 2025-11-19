package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/production-boilerplate/internal/shared/auth"
)

// mockRepository is a mock implementation of Repository for testing.
type mockRepository struct {
	users         map[string]*User
	usersByEmail  map[string]*User
	createError   error
	updateError   error
	getError      error
	getByEmailErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:        make(map[string]*User),
		usersByEmail: make(map[string]*User),
	}
}

func (m *mockRepository) Get(ctx context.Context, id string) (*User, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	user, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *mockRepository) GetByEmail(ctx context.Context, email Email) (*User, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	user, ok := m.usersByEmail[email.String()]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *mockRepository) List(ctx context.Context, offset, limit int) ([]*User, error) {
	users := make([]*User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	// Simple pagination
	start := offset
	end := offset + limit
	if start > len(users) {
		return []*User{}, nil
	}
	if end > len(users) {
		end = len(users)
	}
	return users[start:end], nil
}

func (m *mockRepository) Count(ctx context.Context) (int, error) {
	return len(m.users), nil
}

func (m *mockRepository) Create(ctx context.Context, user *User) error {
	if m.createError != nil {
		return m.createError
	}
	// Check for duplicate email
	if _, exists := m.usersByEmail[user.Email.String()]; exists {
		return ErrUserAlreadyExists
	}
	m.users[user.ID] = user
	m.usersByEmail[user.Email.String()] = user
	return nil
}

func (m *mockRepository) Update(ctx context.Context, user *User) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, exists := m.users[user.ID]; !exists {
		return ErrUserNotFound
	}
	m.users[user.ID] = user
	m.usersByEmail[user.Email.String()] = user
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	user, ok := m.users[id]
	if !ok {
		return ErrUserNotFound
	}
	delete(m.users, id)
	delete(m.usersByEmail, user.Email.String())
	return nil
}

// TestService_Register tests user registration use case.
func TestService_Register(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		setup   func(*mockRepository)
		wantErr error
	}{
		{
			name: "successful registration",
			req: RegisterRequest{
				Email:    "newuser@example.com",
				Password: "SecurePass123",
				Name:     "New User",
			},
			setup:   func(m *mockRepository) {},
			wantErr: nil,
		},
		{
			name: "duplicate email",
			req: RegisterRequest{
				Email:    "existing@example.com",
				Password: "SecurePass123",
				Name:     "Existing User",
			},
			setup: func(m *mockRepository) {
				user, _ := NewUser("existing@example.com", "Password123", "Existing")
				m.users[user.ID] = user
				m.usersByEmail[user.Email.String()] = user
			},
			wantErr: ErrUserAlreadyExists,
		},
		{
			name: "invalid email",
			req: RegisterRequest{
				Email:    "invalid-email",
				Password: "SecurePass123",
				Name:     "User",
			},
			wantErr: errors.New("invalid email format"),
		},
		{
			name: "short password",
			req: RegisterRequest{
				Email:    "user@example.com",
				Password: "short",
				Name:     "User",
			},
			wantErr: errors.New("password must be at least 8 characters"),
		},
		{
			name: "empty name",
			req: RegisterRequest{
				Email:    "user@example.com",
				Password: "SecurePass123",
				Name:     "",
			},
			wantErr: errors.New("name is required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)
			svc := NewService(repo, jwtSvc)

			if tt.setup != nil {
				tt.setup(repo)
			}

			user, err := svc.Register(context.Background(), tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("Register() expected error, got nil")
				} else if err.Error() != tt.wantErr.Error() {
					t.Errorf("Register() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Register() unexpected error = %v", err)
				return
			}

			// Verify user was created correctly
			if user.Email.String() != tt.req.Email {
				t.Errorf("User email = %v, want %v", user.Email.String(), tt.req.Email)
			}
			if user.Name != tt.req.Name {
				t.Errorf("User name = %v, want %v", user.Name, tt.req.Name)
			}
			if user.Role != RoleUser {
				t.Errorf("User role = %v, want %v", user.Role, RoleUser)
			}
			if user.Status != StatusActive {
				t.Errorf("User status = %v, want %v", user.Status, StatusActive)
			}
		})
	}
}

// TestService_Login tests user login use case.
func TestService_Login(t *testing.T) {
	validEmail := "user@example.com"
	validPassword := "SecurePass123"
	validUser, _ := NewUser(validEmail, validPassword, "Test User")

	tests := []struct {
		name    string
		req     LoginRequest
		setup   func(*mockRepository)
		wantErr error
	}{
		{
			name: "successful login",
			req: LoginRequest{
				Email:    validEmail,
				Password: validPassword,
			},
			setup: func(m *mockRepository) {
				m.users[validUser.ID] = validUser
				m.usersByEmail[validUser.Email.String()] = validUser
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			req: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: validPassword,
			},
			setup:   func(m *mockRepository) {},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "wrong password",
			req: LoginRequest{
				Email:    validEmail,
				Password: "WrongPassword",
			},
			setup: func(m *mockRepository) {
				m.users[validUser.ID] = validUser
				m.usersByEmail[validUser.Email.String()] = validUser
			},
			wantErr: ErrInvalidCredentials,
		},
		{
			name: "inactive user",
			req: LoginRequest{
				Email:    validEmail,
				Password: validPassword,
			},
			setup: func(m *mockRepository) {
				inactiveUser, _ := NewUser(validEmail, validPassword, "Inactive User")
				inactiveUser.Status = StatusInactive
				m.users[inactiveUser.ID] = inactiveUser
				m.usersByEmail[inactiveUser.Email.String()] = inactiveUser
			},
			wantErr: ErrUserInactive,
		},
		{
			name: "banned user",
			req: LoginRequest{
				Email:    validEmail,
				Password: validPassword,
			},
			setup: func(m *mockRepository) {
				bannedUser, _ := NewUser(validEmail, validPassword, "Banned User")
				bannedUser.Status = StatusBanned
				m.users[bannedUser.ID] = bannedUser
				m.usersByEmail[bannedUser.Email.String()] = bannedUser
			},
			wantErr: ErrUserInactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)
			svc := NewService(repo, jwtSvc)

			if tt.setup != nil {
				tt.setup(repo)
			}

			token, err := svc.Login(context.Background(), tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Error("Login() expected error, got nil")
				} else if err != tt.wantErr {
					t.Errorf("Login() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Login() unexpected error = %v", err)
				return
			}

			// Verify token was generated
			if token == "" {
				t.Error("Login() should return non-empty token")
			}
		})
	}
}

// TestService_GetProfile tests get profile use case.
func TestService_GetProfile(t *testing.T) {
	validUser, _ := NewUser("user@example.com", "Password123", "Test User")

	tests := []struct {
		name    string
		userID  string
		setup   func(*mockRepository)
		wantErr error
	}{
		{
			name:   "user found",
			userID: validUser.ID,
			setup: func(m *mockRepository) {
				m.users[validUser.ID] = validUser
			},
			wantErr: nil,
		},
		{
			name:    "user not found",
			userID:  "nonexistent-id",
			setup:   func(m *mockRepository) {},
			wantErr: ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)
			svc := NewService(repo, jwtSvc)

			if tt.setup != nil {
				tt.setup(repo)
			}

			user, err := svc.GetProfile(context.Background(), tt.userID)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("GetProfile() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("GetProfile() unexpected error = %v", err)
				return
			}

			if user.ID != tt.userID {
				t.Errorf("GetProfile() user ID = %v, want %v", user.ID, tt.userID)
			}
		})
	}
}

// TestService_UpdateProfile tests update profile use case.
func TestService_UpdateProfile(t *testing.T) {
	validUser, _ := NewUser("user@example.com", "Password123", "Old Name")

	tests := []struct {
		name    string
		userID  string
		req     UpdateProfileRequest
		setup   func(*mockRepository)
		wantErr bool
	}{
		{
			name:   "successful update",
			userID: validUser.ID,
			req:    UpdateProfileRequest{Name: "New Name"},
			setup: func(m *mockRepository) {
				m.users[validUser.ID] = validUser
			},
			wantErr: false,
		},
		{
			name:    "user not found",
			userID:  "nonexistent-id",
			req:     UpdateProfileRequest{Name: "New Name"},
			setup:   func(m *mockRepository) {},
			wantErr: true,
		},
		{
			name:   "empty name",
			userID: validUser.ID,
			req:    UpdateProfileRequest{Name: ""},
			setup: func(m *mockRepository) {
				m.users[validUser.ID] = validUser
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)
			svc := NewService(repo, jwtSvc)

			if tt.setup != nil {
				tt.setup(repo)
			}

			user, err := svc.UpdateProfile(context.Background(), tt.userID, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("UpdateProfile() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("UpdateProfile() unexpected error = %v", err)
				return
			}

			if user.Name != tt.req.Name {
				t.Errorf("UpdateProfile() user name = %v, want %v", user.Name, tt.req.Name)
			}
		})
	}
}

// TestService_ChangePassword tests change password use case.
func TestService_ChangePassword(t *testing.T) {
	oldPassword := "OldPassword123"
	newPassword := "NewPassword456"
	validUser, _ := NewUser("user@example.com", oldPassword, "Test User")

	tests := []struct {
		name    string
		userID  string
		req     ChangePasswordRequest
		setup   func(*mockRepository)
		wantErr bool
	}{
		{
			name:   "successful password change",
			userID: validUser.ID,
			req: ChangePasswordRequest{
				OldPassword: oldPassword,
				NewPassword: newPassword,
			},
			setup: func(m *mockRepository) {
				m.users[validUser.ID] = validUser
			},
			wantErr: false,
		},
		{
			name:   "wrong old password",
			userID: validUser.ID,
			req: ChangePasswordRequest{
				OldPassword: "WrongPassword",
				NewPassword: newPassword,
			},
			setup: func(m *mockRepository) {
				m.users[validUser.ID] = validUser
			},
			wantErr: true,
		},
		{
			name:   "short new password",
			userID: validUser.ID,
			req: ChangePasswordRequest{
				OldPassword: oldPassword,
				NewPassword: "short",
			},
			setup: func(m *mockRepository) {
				m.users[validUser.ID] = validUser
			},
			wantErr: true,
		},
		{
			name:   "user not found",
			userID: "nonexistent-id",
			req: ChangePasswordRequest{
				OldPassword: oldPassword,
				NewPassword: newPassword,
			},
			setup:   func(m *mockRepository) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)
			svc := NewService(repo, jwtSvc)

			if tt.setup != nil {
				tt.setup(repo)
			}

			err := svc.ChangePassword(context.Background(), tt.userID, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Error("ChangePassword() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ChangePassword() unexpected error = %v", err)
			}
		})
	}
}

// TestService_ListUsers tests list users use case.
func TestService_ListUsers(t *testing.T) {
	// Create test users
	user1, _ := NewUser("user1@example.com", "Password123", "User 1")
	user2, _ := NewUser("user2@example.com", "Password123", "User 2")
	user3, _ := NewUser("user3@example.com", "Password123", "User 3")

	tests := []struct {
		name       string
		offset     int
		limit      int
		setup      func(*mockRepository)
		wantCount  int
		wantTotal  int
	}{
		{
			name:   "all users",
			offset: 0,
			limit:  10,
			setup: func(m *mockRepository) {
				m.users[user1.ID] = user1
				m.users[user2.ID] = user2
				m.users[user3.ID] = user3
			},
			wantCount: 3,
			wantTotal: 3,
		},
		{
			name:   "pagination - first page",
			offset: 0,
			limit:  2,
			setup: func(m *mockRepository) {
				m.users[user1.ID] = user1
				m.users[user2.ID] = user2
				m.users[user3.ID] = user3
			},
			wantCount: 2,
			wantTotal: 3,
		},
		{
			name:   "pagination - second page",
			offset: 2,
			limit:  2,
			setup: func(m *mockRepository) {
				m.users[user1.ID] = user1
				m.users[user2.ID] = user2
				m.users[user3.ID] = user3
			},
			wantCount: 1,
			wantTotal: 3,
		},
		{
			name:      "no users",
			offset:    0,
			limit:     10,
			setup:     func(m *mockRepository) {},
			wantCount: 0,
			wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)
			svc := NewService(repo, jwtSvc)

			if tt.setup != nil {
				tt.setup(repo)
			}

			users, total, err := svc.ListUsers(context.Background(), tt.offset, tt.limit)
			if err != nil {
				t.Errorf("ListUsers() unexpected error = %v", err)
				return
			}

			if len(users) != tt.wantCount {
				t.Errorf("ListUsers() returned %d users, want %d", len(users), tt.wantCount)
			}

			if total != tt.wantTotal {
				t.Errorf("ListUsers() total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

// TestService_BanUser tests ban user use case.
func TestService_BanUser(t *testing.T) {
	activeUser, _ := NewUser("user@example.com", "Password123", "Test User")
	bannedUser, _ := NewUser("banned@example.com", "Password123", "Banned User")
	bannedUser.Status = StatusBanned

	tests := []struct {
		name    string
		userID  string
		setup   func(*mockRepository)
		wantErr bool
	}{
		{
			name:   "ban active user",
			userID: activeUser.ID,
			setup: func(m *mockRepository) {
				m.users[activeUser.ID] = activeUser
			},
			wantErr: false,
		},
		{
			name:   "ban already banned user",
			userID: bannedUser.ID,
			setup: func(m *mockRepository) {
				m.users[bannedUser.ID] = bannedUser
			},
			wantErr: true,
		},
		{
			name:    "user not found",
			userID:  "nonexistent-id",
			setup:   func(m *mockRepository) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)
			svc := NewService(repo, jwtSvc)

			if tt.setup != nil {
				tt.setup(repo)
			}

			err := svc.BanUser(context.Background(), tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Error("BanUser() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("BanUser() unexpected error = %v", err)
			}
		})
	}
}

// TestService_PromoteToAdmin tests promote to admin use case.
func TestService_PromoteToAdmin(t *testing.T) {
	regularUser, _ := NewUser("user@example.com", "Password123", "Regular User")
	adminUser, _ := NewUser("admin@example.com", "Password123", "Admin User")
	adminUser.Role = RoleAdmin

	tests := []struct {
		name    string
		userID  string
		setup   func(*mockRepository)
		wantErr bool
	}{
		{
			name:   "promote regular user",
			userID: regularUser.ID,
			setup: func(m *mockRepository) {
				m.users[regularUser.ID] = regularUser
			},
			wantErr: false,
		},
		{
			name:   "promote already admin",
			userID: adminUser.ID,
			setup: func(m *mockRepository) {
				m.users[adminUser.ID] = adminUser
			},
			wantErr: true,
		},
		{
			name:    "user not found",
			userID:  "nonexistent-id",
			setup:   func(m *mockRepository) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepository()
			jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)
			svc := NewService(repo, jwtSvc)

			if tt.setup != nil {
				tt.setup(repo)
			}

			err := svc.PromoteToAdmin(context.Background(), tt.userID)

			if tt.wantErr {
				if err == nil {
					t.Error("PromoteToAdmin() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("PromoteToAdmin() unexpected error = %v", err)
			}
		})
	}
}
