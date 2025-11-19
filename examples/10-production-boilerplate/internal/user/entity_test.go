package user

import (
	"testing"
	"time"
)

// TestNewEmail tests Email value object validation.
func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"valid email with dash", "user-name@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"empty email", "", true},
		{"invalid format - no @", "userexample.com", true},
		{"invalid format - no domain", "user@", true},
		{"invalid format - no TLD", "user@example", true},
		{"invalid format - spaces", "user @example.com", true},
		{"invalid format - missing username", "@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && email.String() != tt.email {
				t.Errorf("NewEmail() = %v, want %v", email.String(), tt.email)
			}
		})
	}
}

// TestNewPassword tests Password value object.
func TestNewPassword(t *testing.T) {
	tests := []struct {
		name        string
		plaintext   string
		wantErr     bool
		errContains string
	}{
		{"valid password - 8 chars", "12345678", false, ""},
		{"valid password - long", "MySecurePassword123!", false, ""},
		{"valid password - special chars", "P@ssw0rd!", false, ""},
		{"too short - 7 chars", "1234567", true, "password must be at least 8 characters"},
		{"empty password", "", true, "password must be at least 8 characters"},
		{"minimum length", "abcdefgh", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := NewPassword(tt.plaintext)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if err.Error() != tt.errContains {
					// Check if error message is as expected (exact match)
					t.Errorf("NewPassword() error = %v, want %v", err.Error(), tt.errContains)
				}
			}
			if !tt.wantErr {
				// Verify password was hashed
				if password.Hash() == tt.plaintext {
					t.Error("Password should be hashed, not plaintext")
				}
				// Verify password can be checked
				if !password.Check(tt.plaintext) {
					t.Error("Password.Check() should verify correct password")
				}
				// Verify wrong password fails
				if password.Check("wrongpassword") {
					t.Error("Password.Check() should reject wrong password")
				}
			}
		})
	}
}

// TestPasswordHashAndCheck tests password hashing and verification.
func TestPasswordHashAndCheck(t *testing.T) {
	plaintext := "MySecurePassword123"

	// Create password
	password, err := NewPassword(plaintext)
	if err != nil {
		t.Fatalf("NewPassword() error = %v", err)
	}

	// Hash should not be empty
	if password.Hash() == "" {
		t.Error("Password hash should not be empty")
	}

	// Hash should not be plaintext
	if password.Hash() == plaintext {
		t.Error("Password should be hashed, not stored as plaintext")
	}

	// Correct password should verify
	if !password.Check(plaintext) {
		t.Error("Password.Check() should verify correct password")
	}

	// Wrong password should fail
	if password.Check("WrongPassword") {
		t.Error("Password.Check() should reject wrong password")
	}

	// Reconstruct from hash
	reconstructed := FromHash(password.Hash())
	if !reconstructed.Check(plaintext) {
		t.Error("Reconstructed password should verify correct password")
	}
}

// TestNewUser tests User factory method.
func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		userName    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid user",
			email:    "user@example.com",
			password: "SecurePass123",
			userName: "John Doe",
			wantErr:  false,
		},
		{
			name:        "invalid email",
			email:       "invalid-email",
			password:    "SecurePass123",
			userName:    "John Doe",
			wantErr:     true,
			errContains: "invalid email format",
		},
		{
			name:        "short password",
			email:       "user@example.com",
			password:    "short",
			userName:    "John Doe",
			wantErr:     true,
			errContains: "password must be at least 8 characters",
		},
		{
			name:        "empty name",
			email:       "user@example.com",
			password:    "SecurePass123",
			userName:    "",
			wantErr:     true,
			errContains: "name cannot be empty",
		},
		{
			name:        "empty email",
			email:       "",
			password:    "SecurePass123",
			userName:    "John Doe",
			wantErr:     true,
			errContains: "email cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.email, tt.password, tt.userName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if err.Error() != tt.errContains {
					t.Errorf("NewUser() error = %v, want error containing %v", err.Error(), tt.errContains)
				}
			}
			if !tt.wantErr {
				// Verify user was created with correct defaults
				if user.ID == "" {
					t.Error("User ID should be generated")
				}
				if user.Email.String() != tt.email {
					t.Errorf("User email = %v, want %v", user.Email.String(), tt.email)
				}
				if user.Name != tt.userName {
					t.Errorf("User name = %v, want %v", user.Name, tt.userName)
				}
				if user.Role != RoleUser {
					t.Errorf("User role = %v, want %v", user.Role, RoleUser)
				}
				if user.Status != StatusActive {
					t.Errorf("User status = %v, want %v", user.Status, StatusActive)
				}
				if user.CreatedAt.IsZero() {
					t.Error("CreatedAt should be set")
				}
				if user.UpdatedAt.IsZero() {
					t.Error("UpdatedAt should be set")
				}
			}
		})
	}
}

// TestUser_Activate tests user activation.
func TestUser_Activate(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  Status
		wantErr        bool
		expectedStatus Status
	}{
		{"activate inactive user", StatusInactive, false, StatusActive},
		{"activate banned user", StatusBanned, false, StatusActive},
		{"activate already active user", StatusActive, true, StatusActive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createTestUser(t)
			user.Status = tt.initialStatus
			oldUpdatedAt := user.UpdatedAt
			time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

			err := user.Activate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Activate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if user.Status != tt.expectedStatus {
				t.Errorf("After Activate(), status = %v, want %v", user.Status, tt.expectedStatus)
			}
			if !tt.wantErr && !user.UpdatedAt.After(oldUpdatedAt) {
				t.Error("UpdatedAt should be updated after Activate()")
			}
		})
	}
}

// TestUser_Deactivate tests user deactivation.
func TestUser_Deactivate(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  Status
		wantErr        bool
		expectedStatus Status
	}{
		{"deactivate active user", StatusActive, false, StatusInactive},
		{"deactivate banned user", StatusBanned, false, StatusInactive},
		{"deactivate already inactive user", StatusInactive, true, StatusInactive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createTestUser(t)
			user.Status = tt.initialStatus
			oldUpdatedAt := user.UpdatedAt
			time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

			err := user.Deactivate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Deactivate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if user.Status != tt.expectedStatus {
				t.Errorf("After Deactivate(), status = %v, want %v", user.Status, tt.expectedStatus)
			}
			if !tt.wantErr && !user.UpdatedAt.After(oldUpdatedAt) {
				t.Error("UpdatedAt should be updated after Deactivate()")
			}
		})
	}
}

// TestUser_Ban tests user banning.
func TestUser_Ban(t *testing.T) {
	tests := []struct {
		name           string
		initialStatus  Status
		wantErr        bool
		expectedStatus Status
	}{
		{"ban active user", StatusActive, false, StatusBanned},
		{"ban inactive user", StatusInactive, false, StatusBanned},
		{"ban already banned user", StatusBanned, true, StatusBanned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createTestUser(t)
			user.Status = tt.initialStatus
			oldUpdatedAt := user.UpdatedAt
			time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

			err := user.Ban()
			if (err != nil) != tt.wantErr {
				t.Errorf("Ban() error = %v, wantErr %v", err, tt.wantErr)
			}
			if user.Status != tt.expectedStatus {
				t.Errorf("After Ban(), status = %v, want %v", user.Status, tt.expectedStatus)
			}
			if !tt.wantErr && !user.UpdatedAt.After(oldUpdatedAt) {
				t.Error("UpdatedAt should be updated after Ban()")
			}
		})
	}
}

// TestUser_PromoteToAdmin tests user promotion.
func TestUser_PromoteToAdmin(t *testing.T) {
	tests := []struct {
		name         string
		initialRole  Role
		wantErr      bool
		expectedRole Role
	}{
		{"promote regular user", RoleUser, false, RoleAdmin},
		{"promote already admin", RoleAdmin, true, RoleAdmin},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createTestUser(t)
			user.Role = tt.initialRole
			oldUpdatedAt := user.UpdatedAt
			time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

			err := user.PromoteToAdmin()
			if (err != nil) != tt.wantErr {
				t.Errorf("PromoteToAdmin() error = %v, wantErr %v", err, tt.wantErr)
			}
			if user.Role != tt.expectedRole {
				t.Errorf("After PromoteToAdmin(), role = %v, want %v", user.Role, tt.expectedRole)
			}
			if !tt.wantErr && !user.UpdatedAt.After(oldUpdatedAt) {
				t.Error("UpdatedAt should be updated after PromoteToAdmin()")
			}
		})
	}
}

// TestUser_DemoteToUser tests user demotion.
func TestUser_DemoteToUser(t *testing.T) {
	tests := []struct {
		name         string
		initialRole  Role
		wantErr      bool
		expectedRole Role
	}{
		{"demote admin user", RoleAdmin, false, RoleUser},
		{"demote already regular user", RoleUser, true, RoleUser},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createTestUser(t)
			user.Role = tt.initialRole
			oldUpdatedAt := user.UpdatedAt
			time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

			err := user.DemoteToUser()
			if (err != nil) != tt.wantErr {
				t.Errorf("DemoteToUser() error = %v, wantErr %v", err, tt.wantErr)
			}
			if user.Role != tt.expectedRole {
				t.Errorf("After DemoteToUser(), role = %v, want %v", user.Role, tt.expectedRole)
			}
			if !tt.wantErr && !user.UpdatedAt.After(oldUpdatedAt) {
				t.Error("UpdatedAt should be updated after DemoteToUser()")
			}
		})
	}
}

// TestUser_ChangePassword tests password change.
func TestUser_ChangePassword(t *testing.T) {
	oldPassword := "OldPassword123"
	newPassword := "NewPassword456"

	tests := []struct {
		name        string
		oldPass     string
		newPass     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid password change",
			oldPass: oldPassword,
			newPass: newPassword,
			wantErr: false,
		},
		{
			name:        "wrong old password",
			oldPass:     "WrongPassword",
			newPass:     newPassword,
			wantErr:     true,
			errContains: "invalid old password",
		},
		{
			name:        "new password too short",
			oldPass:     oldPassword,
			newPass:     "short",
			wantErr:     true,
			errContains: "password must be at least 8 characters",
		},
		{
			name:        "empty new password",
			oldPass:     oldPassword,
			newPass:     "",
			wantErr:     true,
			errContains: "password must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser("user@example.com", oldPassword, "Test User")
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			oldHash := user.Password.Hash()
			oldUpdatedAt := user.UpdatedAt
			time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

			err = user.ChangePassword(tt.oldPass, tt.newPass)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" {
				if err.Error() != tt.errContains {
					t.Errorf("ChangePassword() error = %v, want error containing %v", err.Error(), tt.errContains)
				}
			}
			if !tt.wantErr {
				// Verify password changed
				if user.Password.Hash() == oldHash {
					t.Error("Password hash should change")
				}
				// Verify new password works
				if !user.Password.Check(tt.newPass) {
					t.Error("New password should be verifiable")
				}
				// Verify old password no longer works
				if user.Password.Check(oldPassword) {
					t.Error("Old password should not work")
				}
				// Verify timestamp updated
				if !user.UpdatedAt.After(oldUpdatedAt) {
					t.Error("UpdatedAt should be updated after ChangePassword()")
				}
			}
		})
	}
}

// TestUser_UpdateProfile tests profile update.
func TestUser_UpdateProfile(t *testing.T) {
	tests := []struct {
		name        string
		newName     string
		wantErr     bool
		errContains string
	}{
		{"valid name update", "New Name", false, ""},
		{"empty name", "", true, "name cannot be empty"},
		{"long name", "A Very Long Name With Many Words And Characters", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := createTestUser(t)
			oldName := user.Name
			oldUpdatedAt := user.UpdatedAt
			time.Sleep(10 * time.Millisecond)

			err := user.UpdateProfile(tt.newName)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" {
				if err.Error() != tt.errContains {
					t.Errorf("UpdateProfile() error = %v, want error containing %v", err.Error(), tt.errContains)
				}
			}
			if !tt.wantErr {
				if user.Name != tt.newName {
					t.Errorf("After UpdateProfile(), name = %v, want %v", user.Name, tt.newName)
				}
				if !user.UpdatedAt.After(oldUpdatedAt) {
					t.Error("UpdatedAt should be updated after UpdateProfile()")
				}
			} else {
				// Verify name didn't change on error
				if user.Name != oldName {
					t.Error("Name should not change when UpdateProfile() returns error")
				}
			}
		})
	}
}

// TestUser_IsActive tests status check methods.
func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{StatusActive, true},
		{StatusInactive, false},
		{StatusBanned, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			user := createTestUser(t)
			user.Status = tt.status
			if got := user.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestUser_IsAdmin tests admin check method.
func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleAdmin, true},
		{RoleUser, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			user := createTestUser(t)
			user.Role = tt.role
			if got := user.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestUser_IsBanned tests banned check method.
func TestUser_IsBanned(t *testing.T) {
	tests := []struct {
		status Status
		want   bool
	}{
		{StatusBanned, true},
		{StatusActive, false},
		{StatusInactive, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			user := createTestUser(t)
			user.Status = tt.status
			if got := user.IsBanned(); got != tt.want {
				t.Errorf("IsBanned() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create a test user.
func createTestUser(t *testing.T) *User {
	t.Helper()
	user, err := NewUser("test@example.com", "Password123", "Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}
