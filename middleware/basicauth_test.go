// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coregx/fursy"
)

// TestBasicAuth tests the default BasicAuth middleware.
func TestBasicAuth(t *testing.T) {
	validator := func(_ *fursy.Context, username, password string) (interface{}, error) {
		if username == "admin" && password == "secret" {
			return username, nil
		}
		return nil, errors.New("invalid credentials")
	}

	r := fursy.New()
	r.Use(BasicAuth(validator))

	r.GET("/test", func(c *fursy.Context) error {
		user := c.GetString(UserContextKey)
		return c.String(200, "Hello, "+user)
	})

	// Test with valid credentials.
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:secret")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Hello, admin" {
		t.Errorf("expected 'Hello, admin', got %s", w.Body.String())
	}
}

// TestBasicAuth_NoAuth tests request without Authorization header.
func TestBasicAuth_NoAuth(t *testing.T) {
	validator := func(_ *fursy.Context, username, _ string) (interface{}, error) {
		return username, nil
	}

	r := fursy.New()
	r.Use(BasicAuth(validator))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != `Basic realm="Restricted"` {
		t.Errorf("expected WWW-Authenticate header, got %s", wwwAuth)
	}

	if w.Body.String() != "Unauthorized" {
		t.Errorf("expected 'Unauthorized', got %s", w.Body.String())
	}
}

// TestBasicAuth_InvalidCredentials tests invalid username/password.
func TestBasicAuth_InvalidCredentials(t *testing.T) {
	validator := func(_ *fursy.Context, username, password string) (interface{}, error) {
		if username == "admin" && password == "secret" {
			return username, nil
		}
		return nil, errors.New("invalid credentials")
	}

	r := fursy.New()
	r.Use(BasicAuth(validator))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Test with invalid credentials.
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:wrong")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	if w.Header().Get("WWW-Authenticate") == "" {
		t.Error("expected WWW-Authenticate header")
	}
}

// TestBasicAuth_CustomRealm tests custom realm configuration.
func TestBasicAuth_CustomRealm(t *testing.T) {
	validator := func(_ *fursy.Context, _, _ string) (interface{}, error) {
		return nil, errors.New("always fail")
	}

	r := fursy.New()
	r.Use(BasicAuthWithConfig(BasicAuthConfig{
		Validator: validator,
		Realm:     "Admin Area",
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	wwwAuth := w.Header().Get("WWW-Authenticate")
	if wwwAuth != `Basic realm="Admin Area"` {
		t.Errorf("expected custom realm, got %s", wwwAuth)
	}
}

// TestBasicAuth_Skipper tests the Skipper function.
func TestBasicAuth_Skipper(t *testing.T) {
	validator := func(_ *fursy.Context, _, _ string) (interface{}, error) {
		return nil, errors.New("should not be called")
	}

	r := fursy.New()
	r.Use(BasicAuthWithConfig(BasicAuthConfig{
		Validator: validator,
		Skipper: func(c *fursy.Context) bool {
			// Skip auth for /health endpoint.
			return c.Request.URL.Path == "/health"
		},
	}))

	r.GET("/health", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	r.GET("/protected", func(c *fursy.Context) error {
		return c.String(200, "Secret")
	})

	// Test skipped endpoint.
	req := httptest.NewRequest("GET", "/health", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("skipped endpoint: expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("skipped endpoint: expected 'OK', got %s", w.Body.String())
	}

	// Test protected endpoint.
	req2 := httptest.NewRequest("GET", "/protected", http.NoBody)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != 401 {
		t.Errorf("protected endpoint: expected status 401, got %d", w2.Code)
	}
}

// TestBasicAuth_UserIdentity tests storing user identity in context.
func TestBasicAuth_UserIdentity(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	validator := func(_ *fursy.Context, username, password string) (interface{}, error) {
		if username == "admin" && password == "secret" {
			return &User{ID: 1, Name: "Admin"}, nil
		}
		return nil, errors.New("invalid credentials")
	}

	r := fursy.New()
	r.Use(BasicAuth(validator))

	r.GET("/test", func(c *fursy.Context) error {
		user := c.Get(UserContextKey).(*User)
		return c.String(200, user.Name)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:secret")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Admin" {
		t.Errorf("expected 'Admin', got %s", w.Body.String())
	}
}

// TestBasicAuth_WithRouteGroups tests BasicAuth with route groups.
func TestBasicAuth_WithRouteGroups(t *testing.T) {
	validator := func(_ *fursy.Context, username, password string) (interface{}, error) {
		if username == "admin" && password == "secret" {
			return username, nil
		}
		return nil, errors.New("invalid credentials")
	}

	r := fursy.New()
	r.Use(BasicAuth(validator))

	api := r.Group("/api")
	api.GET("/users", func(c *fursy.Context) error {
		user := c.GetString(UserContextKey)
		return c.String(200, "user: "+user)
	})

	req := httptest.NewRequest("GET", "/api/users", http.NoBody)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:secret")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "user: admin" {
		t.Errorf("expected 'user: admin', got %s", w.Body.String())
	}
}

// TestParseBasicAuth tests the parseBasicAuth helper function.
func TestParseBasicAuth(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		username string
		password string
	}{
		{
			name:     "empty header",
			header:   "",
			username: "",
			password: "",
		},
		{
			name:     "valid credentials",
			header:   "Basic " + base64.StdEncoding.EncodeToString([]byte("Aladdin:open sesame")),
			username: "Aladdin",
			password: "open sesame",
		},
		{
			name:     "invalid base64",
			header:   "Basic xyz",
			username: "",
			password: "",
		},
		{
			name:     "no colon separator",
			header:   "Basic " + base64.StdEncoding.EncodeToString([]byte("invalidformat")),
			username: "",
			password: "",
		},
		{
			name:     "wrong scheme",
			header:   "Bearer token123",
			username: "",
			password: "",
		},
		{
			name:     "password with colon",
			header:   "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass:word")),
			username: "user",
			password: "pass:word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, password := parseBasicAuth(tt.header)
			if username != tt.username {
				t.Errorf("expected username %q, got %q", tt.username, username)
			}
			if password != tt.password {
				t.Errorf("expected password %q, got %q", tt.password, password)
			}
		})
	}
}

// TestBasicAuthAccounts tests the BasicAuthAccounts helper.
func TestBasicAuthAccounts(t *testing.T) {
	accounts := map[string]string{
		"admin": "secret",
		"user":  "pass123",
	}

	r := fursy.New()
	r.Use(BasicAuth(BasicAuthAccounts(accounts)))

	r.GET("/test", func(c *fursy.Context) error {
		user := c.GetString(UserContextKey)
		return c.String(200, "Hello, "+user)
	})

	// Test valid user.
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:secret")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Hello, admin" {
		t.Errorf("expected 'Hello, admin', got %s", w.Body.String())
	}

	// Test invalid password.
	req2 := httptest.NewRequest("GET", "/test", http.NoBody)
	req2.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:wrong")))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != 401 {
		t.Errorf("expected status 401, got %d", w2.Code)
	}

	// Test unknown user.
	req3 := httptest.NewRequest("GET", "/test", http.NoBody)
	req3.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("unknown:pass")))
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	if w3.Code != 401 {
		t.Errorf("expected status 401, got %d", w3.Code)
	}
}

// TestBasicAuth_NilValidatorPanics tests that nil validator causes panic.
func TestBasicAuth_NilValidatorPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil validator")
		}
	}()

	_ = BasicAuthWithConfig(BasicAuthConfig{
		Validator: nil,
	})
}

// TestBasicAuth_EmptyCredentials tests empty username/password.
func TestBasicAuth_EmptyCredentials(t *testing.T) {
	validator := func(_ *fursy.Context, username, password string) (interface{}, error) {
		// Empty credentials should still be validated.
		if username == "" && password == "" {
			return nil, errors.New("empty credentials")
		}
		return username, nil
	}

	r := fursy.New()
	r.Use(BasicAuth(validator))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Test with empty credentials (just colon).
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// TestBasicAuth_MultipleRequests tests that auth works across multiple requests.
func TestBasicAuth_MultipleRequests(t *testing.T) {
	callCount := 0
	validator := func(_ *fursy.Context, username, password string) (interface{}, error) {
		callCount++
		if username == "admin" && password == "secret" {
			return username, nil
		}
		return nil, errors.New("invalid")
	}

	r := fursy.New()
	r.Use(BasicAuth(validator))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Request 1 - valid.
	req1 := httptest.NewRequest("GET", "/test", http.NoBody)
	req1.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:secret")))
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("request 1: expected status 200, got %d", w1.Code)
	}

	// Request 2 - invalid.
	req2 := httptest.NewRequest("GET", "/test", http.NoBody)
	req2.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:wrong")))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != 401 {
		t.Errorf("request 2: expected status 401, got %d", w2.Code)
	}

	// Request 3 - valid again.
	req3 := httptest.NewRequest("GET", "/test", http.NoBody)
	req3.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("admin:secret")))
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	if w3.Code != 200 {
		t.Errorf("request 3: expected status 200, got %d", w3.Code)
	}

	if callCount != 3 {
		t.Errorf("expected validator called 3 times, got %d", callCount)
	}
}
