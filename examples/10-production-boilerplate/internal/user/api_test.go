package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/production-boilerplate/internal/shared/auth"
	"github.com/coregx/fursy"
)

// mockService is a mock implementation of Service for API testing.
type mockService struct {
	registerFunc      func(ctx context.Context, req RegisterRequest) (*User, error)
	loginFunc         func(ctx context.Context, req LoginRequest) (string, error)
	getProfileFunc    func(ctx context.Context, userID string) (*User, error)
	updateProfileFunc func(ctx context.Context, userID string, req UpdateProfileRequest) (*User, error)
	changePasswordFunc func(ctx context.Context, userID string, req ChangePasswordRequest) error
	listUsersFunc     func(ctx context.Context, offset, limit int) ([]*User, int, error)
	banUserFunc       func(ctx context.Context, userID string) error
	promoteFunc       func(ctx context.Context, userID string) error
}

func (m *mockService) Register(ctx context.Context, req RegisterRequest) (*User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockService) Login(ctx context.Context, req LoginRequest) (string, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return "", nil
}

func (m *mockService) GetProfile(ctx context.Context, userID string) (*User, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*User, error) {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockService) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	if m.changePasswordFunc != nil {
		return m.changePasswordFunc(ctx, userID, req)
	}
	return nil
}

func (m *mockService) ListUsers(ctx context.Context, offset, limit int) ([]*User, int, error) {
	if m.listUsersFunc != nil {
		return m.listUsersFunc(ctx, offset, limit)
	}
	return []*User{}, 0, nil
}

func (m *mockService) BanUser(ctx context.Context, userID string) error {
	if m.banUserFunc != nil {
		return m.banUserFunc(ctx, userID)
	}
	return nil
}

func (m *mockService) PromoteToAdmin(ctx context.Context, userID string) error {
	if m.promoteFunc != nil {
		return m.promoteFunc(ctx, userID)
	}
	return nil
}

// setupTestRouter creates a router with basic test routes (skip admin to avoid route conflicts in tests).
func setupTestRouter(api *API) *fursy.Router {
	router := fursy.New()
	jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)

	// Auth middleware
	authMiddleware := func(c *fursy.Context) error {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			return c.Problem(fursy.Unauthorized("Missing authorization header"))
		}

		// Extract token (format: "Bearer <token>")
		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			return c.Problem(fursy.Unauthorized("Invalid authorization format"))
		}
		token := authHeader[7:]

		claims, err := jwtSvc.ValidateToken(token)
		if err != nil {
			return c.Problem(fursy.Unauthorized("Invalid token"))
		}

		// Add user ID and role to context
		ctx := auth.WithUserID(c.Request.Context(), claims.UserID)
		ctx = auth.WithUserRole(ctx, claims.Role)
		c.Request = c.Request.WithContext(ctx)

		return c.Next()
	}

	// Register only public and basic protected routes (skip admin routes due to route conflicts)
	router.POST("/api/auth/register", api.register)
	router.POST("/api/auth/login", api.login)

	// Protected routes
	protected := router.Group("/api")
	protected.Use(authMiddleware)
	{
		protected.GET("/users/me", api.getProfile)
		protected.PUT("/users/me", api.updateProfile)
		protected.POST("/users/me/password", api.changePassword)
	}

	return router
}

// TestAPI_Register tests user registration endpoint.
func TestAPI_Register(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		mockFunc       func(ctx context.Context, req RegisterRequest) (*User, error)
		wantStatusCode int
		checkBody      func(t *testing.T, body string)
	}{
		{
			name: "successful registration",
			payload: map[string]string{
				"email":    "newuser@example.com",
				"password": "SecurePass123",
				"name":     "New User",
			},
			mockFunc: func(ctx context.Context, req RegisterRequest) (*User, error) {
				user, _ := NewUser(req.Email, req.Password, req.Name)
				return user, nil
			},
			wantStatusCode: http.StatusCreated,
			checkBody: func(t *testing.T, body string) {
				var resp UserResponse
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if resp.Email != "newuser@example.com" {
					t.Errorf("Response email = %v, want newuser@example.com", resp.Email)
				}
			},
		},
		{
			name: "duplicate email",
			payload: map[string]string{
				"email":    "existing@example.com",
				"password": "SecurePass123",
				"name":     "User",
			},
			mockFunc: func(ctx context.Context, req RegisterRequest) (*User, error) {
				return nil, ErrUserAlreadyExists
			},
			wantStatusCode: http.StatusConflict,
			checkBody: func(t *testing.T, body string) {
				if !contains(body, "User already exists") {
					t.Errorf("Response should contain 'User already exists', got: %s", body)
				}
			},
		},
		{
			name:           "invalid JSON",
			payload:        "invalid json",
			wantStatusCode: http.StatusBadRequest,
			checkBody: func(t *testing.T, body string) {
				if !contains(body, "Invalid Request") {
					t.Errorf("Response should contain 'Invalid Request', got: %s", body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				registerFunc: tt.mockFunc,
			}
			api := NewAPI(mockSvc)
			router := setupTestRouter(api)

			// Create request
			var body []byte
			if str, ok := tt.payload.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.payload)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert status code
			if w.Code != tt.wantStatusCode {
				t.Errorf("Status code = %d, want %d", w.Code, tt.wantStatusCode)
			}

			// Check body
			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.String())
			}
		})
	}
}

// TestAPI_Login tests user login endpoint.
func TestAPI_Login(t *testing.T) {
	tests := []struct {
		name           string
		payload        interface{}
		mockFunc       func(ctx context.Context, req LoginRequest) (string, error)
		wantStatusCode int
		checkBody      func(t *testing.T, body string)
	}{
		{
			name: "successful login",
			payload: map[string]string{
				"email":    "user@example.com",
				"password": "SecurePass123",
			},
			mockFunc: func(ctx context.Context, req LoginRequest) (string, error) {
				return "test-jwt-token", nil
			},
			wantStatusCode: http.StatusOK,
			checkBody: func(t *testing.T, body string) {
				var resp map[string]string
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if resp["token"] != "test-jwt-token" {
					t.Errorf("Token = %v, want test-jwt-token", resp["token"])
				}
			},
		},
		{
			name: "invalid credentials",
			payload: map[string]string{
				"email":    "user@example.com",
				"password": "WrongPassword",
			},
			mockFunc: func(ctx context.Context, req LoginRequest) (string, error) {
				return "", ErrInvalidCredentials
			},
			wantStatusCode: http.StatusUnauthorized,
			checkBody: func(t *testing.T, body string) {
				if !contains(body, "Invalid credentials") {
					t.Errorf("Response should contain 'Invalid credentials', got: %s", body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				loginFunc: tt.mockFunc,
			}
			api := NewAPI(mockSvc)
			router := setupTestRouter(api)

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Status code = %d, want %d", w.Code, tt.wantStatusCode)
			}

			if tt.checkBody != nil {
				tt.checkBody(t, w.Body.String())
			}
		})
	}
}

// TestAPI_GetProfile tests get profile endpoint (protected).
func TestAPI_GetProfile(t *testing.T) {
	testUser, _ := NewUser("user@example.com", "Password123", "Test User")

	tests := []struct {
		name           string
		authHeader     string
		mockFunc       func(ctx context.Context, userID string) (*User, error)
		wantStatusCode int
	}{
		{
			name:       "successful get profile",
			authHeader: createTestJWT(t, testUser.ID, "user"),
			mockFunc: func(ctx context.Context, userID string) (*User, error) {
				return testUser, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "invalid token format",
			authHeader:     "InvalidFormat",
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockService{
				getProfileFunc: tt.mockFunc,
			}
			api := NewAPI(mockSvc)
			router := setupTestRouter(api)

			req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Status code = %d, want %d (body: %s)", w.Code, tt.wantStatusCode, w.Body.String())
			}
		})
	}
}

// TestAPI_UpdateProfile tests update profile endpoint.
func TestAPI_UpdateProfile(t *testing.T) {
	testUser, _ := NewUser("user@example.com", "Password123", "Old Name")

	mockSvc := &mockService{
		updateProfileFunc: func(ctx context.Context, userID string, req UpdateProfileRequest) (*User, error) {
			testUser.Name = req.Name
			return testUser, nil
		},
	}

	api := NewAPI(mockSvc)
	router := setupTestRouter(api)

	payload := map[string]string{"name": "New Name"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/api/users/me", bytes.NewReader(body))
	req.Header.Set("Authorization", createTestJWT(t, testUser.ID, "user"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}

	var resp UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err == nil {
		if resp.Name != "New Name" {
			t.Errorf("Updated name = %v, want New Name", resp.Name)
		}
	}
}

// TestAPI_ChangePassword tests change password endpoint.
func TestAPI_ChangePassword(t *testing.T) {
	testUser, _ := NewUser("user@example.com", "Password123", "Test User")

	mockSvc := &mockService{
		changePasswordFunc: func(ctx context.Context, userID string, req ChangePasswordRequest) error {
			return nil
		},
	}

	api := NewAPI(mockSvc)
	router := setupTestRouter(api)

	payload := map[string]string{
		"old_password": "Password123",
		"new_password": "NewPassword456",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/users/me/password", bytes.NewReader(body))
	req.Header.Set("Authorization", createTestJWT(t, testUser.ID, "user"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusOK)
	}
}

// Note: Admin routes (ListUsers, BanUser, PromoteToAdmin) are tested via integration tests
// to avoid routing conflicts in API unit tests.

// Helper functions

func createTestJWT(t *testing.T, userID, role string) string {
	t.Helper()
	jwtSvc := auth.NewJWTService([]byte("test-secret"), 24*time.Hour)
	token, err := jwtSvc.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("Failed to create test JWT: %v", err)
	}
	return "Bearer " + token
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
