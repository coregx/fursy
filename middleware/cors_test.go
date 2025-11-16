// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coregx/fursy"
)

// TestCORS tests the default CORS middleware.
func TestCORS(t *testing.T) {
	r := fursy.New()
	r.Use(CORS())

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Actual request with Origin.
	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Should have Access-Control-Allow-Origin header.
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Allow-Origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

// TestCORS_NoOrigin tests request without Origin header.
func TestCORS_NoOrigin(t *testing.T) {
	r := fursy.New()
	r.Use(CORS())

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Should not have CORS headers.
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("should not have Allow-Origin header for non-CORS request")
	}
}

// TestCORS_Preflight tests OPTIONS preflight requests.
func TestCORS_Preflight(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(CORSConfig{
		AllowOrigins: "https://example.com,https://foo.com",
		AllowMethods: "GET,POST,PUT",
		AllowHeaders: "Content-Type,Authorization",
		MaxAge:       12 * time.Hour,
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Register OPTIONS handler for preflight (CORS middleware will handle it).
	r.OPTIONS("/test", func(c *fursy.Context) error {
		return c.NoContent(204)
	})

	// Preflight request.
	req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected Allow-Origin https://example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	if w.Header().Get("Access-Control-Allow-Methods") != "GET,POST,PUT" {
		t.Errorf("expected Allow-Methods GET,POST,PUT, got %s", w.Header().Get("Access-Control-Allow-Methods"))
	}

	if w.Header().Get("Access-Control-Allow-Headers") != "Content-Type" {
		t.Errorf("expected Allow-Headers Content-Type, got %s", w.Header().Get("Access-Control-Allow-Headers"))
	}

	maxAge := w.Header().Get("Access-Control-Max-Age")
	expected := "43200" // 12 hours in seconds
	if maxAge != expected {
		t.Errorf("expected Max-Age %s, got %s", expected, maxAge)
	}
}

// TestCORS_PreflightWildcard tests preflight with wildcard config.
func TestCORS_PreflightWildcard(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(AllowAll))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	// Register OPTIONS handler for preflight (CORS middleware will handle it).
	r.OPTIONS("/test", func(c *fursy.Context) error {
		return c.NoContent(204)
	})

	// Preflight request.
	req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "DELETE")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Allow-Origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// With AllowMethods="*", should echo back the requested method.
	if w.Header().Get("Access-Control-Allow-Methods") != "DELETE" {
		t.Errorf("expected Allow-Methods DELETE, got %s", w.Header().Get("Access-Control-Allow-Methods"))
	}

	// With AllowHeaders="*", should echo back the requested headers.
	if w.Header().Get("Access-Control-Allow-Headers") != "X-Custom-Header" {
		t.Errorf("expected Allow-Headers X-Custom-Header, got %s", w.Header().Get("Access-Control-Allow-Headers"))
	}
}

// TestCORS_AllowCredentials tests AllowCredentials configuration.
func TestCORS_AllowCredentials(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(CORSConfig{
		AllowOrigins:     "https://example.com",
		AllowCredentials: true,
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("with credentials, should use specific origin")
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Allow-Credentials true, got %s", w.Header().Get("Access-Control-Allow-Credentials"))
	}
}

// TestCORS_AllowCredentialsWithWildcard tests that credentials forces specific origin.
func TestCORS_AllowCredentialsWithWildcard(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(CORSConfig{
		AllowOrigins:     "*",
		AllowCredentials: true,
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Even with AllowOrigins="*", when credentials=true, should use specific origin.
	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("with credentials, should use specific origin even with wildcard")
	}

	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Allow-Credentials true")
	}
}

// TestCORS_ExposeHeaders tests ExposeHeaders configuration.
func TestCORS_ExposeHeaders(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(CORSConfig{
		AllowOrigins:  "*",
		ExposeHeaders: "X-Request-ID,X-Response-Time",
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Expose-Headers") != "X-Request-ID,X-Response-Time" {
		t.Errorf("expected Expose-Headers, got %s", w.Header().Get("Access-Control-Expose-Headers"))
	}
}

// TestCORS_DisallowedOrigin tests rejection of non-allowed origins.
func TestCORS_DisallowedOrigin(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(CORSConfig{
		AllowOrigins: "https://example.com,https://foo.com",
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", "https://bar.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should not have CORS headers for disallowed origin.
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("should not have Allow-Origin for disallowed origin")
	}
}

// TestCORS_DisallowedMethod tests rejection of non-allowed methods in preflight.
func TestCORS_DisallowedMethod(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(CORSConfig{
		AllowOrigins: "https://example.com",
		AllowMethods: "GET,POST",
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "DELETE")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should not have CORS headers for disallowed method.
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("should not have Allow-Origin for disallowed method")
	}
}

// TestCORS_DisallowedHeaders tests rejection of non-allowed headers in preflight.
func TestCORS_DisallowedHeaders(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(CORSConfig{
		AllowOrigins: "https://example.com",
		AllowMethods: "GET,POST",
		AllowHeaders: "Content-Type",
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should not have CORS headers for disallowed headers.
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("should not have Allow-Origin for disallowed headers")
	}
}

// TestCORS_NullOrigin tests "null" origin configuration (disallow all).
func TestCORS_NullOrigin(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(CORSConfig{
		AllowOrigins: "null",
	}))

	r.GET("/test", func(c *fursy.Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should not have CORS headers when AllowOrigins="null".
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("should not have Allow-Origin when AllowOrigins=null")
	}
}

// TestCORS_OPTIONSWithoutRequestMethod tests OPTIONS without Access-Control-Request-Method.
func TestCORS_OPTIONSWithoutRequestMethod(t *testing.T) {
	r := fursy.New()
	r.Use(CORS())

	handlerCalled := false
	r.OPTIONS("/test", func(c *fursy.Context) error {
		handlerCalled = true
		return c.String(200, "OK")
	})

	// OPTIONS request without Access-Control-Request-Method (not a preflight).
	req := httptest.NewRequest("OPTIONS", "/test", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should call the handler (not treated as preflight).
	if !handlerCalled {
		t.Error("handler should be called for OPTIONS without Request-Method")
	}
}

// TestCORS_WithRouteGroups tests CORS with route groups.
func TestCORS_WithRouteGroups(t *testing.T) {
	r := fursy.New()
	r.Use(CORSWithConfig(CORSConfig{
		AllowOrigins: "https://example.com",
	}))

	api := r.Group("/api")
	api.GET("/users", func(c *fursy.Context) error {
		return c.String(200, "users")
	})

	req := httptest.NewRequest("GET", "/api/users", http.NoBody)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected Allow-Origin header in group route")
	}
}

// TestBuildAllowMap tests the buildAllowMap helper function.
func TestBuildAllowMap(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		m := buildAllowMap("", false)
		if len(m) != 0 {
			t.Errorf("expected empty map, got %d items", len(m))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		m := buildAllowMap("GET, put", false)
		if len(m) != 2 {
			t.Errorf("expected 2 items, got %d", len(m))
		}
		if !m["GET"] {
			t.Error("should have GET")
		}
		if !m["PUT"] {
			t.Error("should have PUT (uppercased)")
		}
		if m["put"] {
			t.Error("should not have lowercase put")
		}
	})

	t.Run("case sensitive", func(t *testing.T) {
		m := buildAllowMap("GET, put", true)
		if len(m) != 2 {
			t.Errorf("expected 2 items, got %d", len(m))
		}
		if !m["GET"] {
			t.Error("should have GET")
		}
		if m["PUT"] {
			t.Error("should not have PUT")
		}
		if !m["put"] {
			t.Error("should have put (original case)")
		}
	})
}

// TestCORSConfig_IsOriginAllowed tests the isOriginAllowed method.
func TestCORSConfig_IsOriginAllowed(t *testing.T) {
	tests := []struct {
		name         string
		allowOrigins string
		origin       string
		expected     bool
	}{
		{"wildcard allows all", "*", "https://example.com", true},
		{"null disallows all", "null", "https://example.com", false},
		{"specific origin not allowed", "https://foo.com", "https://example.com", false},
		{"specific origin allowed", "https://example.com", "https://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &CORSConfig{AllowOrigins: tt.allowOrigins}
			cfg.init()
			result := cfg.isOriginAllowed(tt.origin)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCORSConfig_IsPreflightAllowed tests the isPreflightAllowed method.
func TestCORSConfig_IsPreflightAllowed(t *testing.T) {
	t.Run("allowed origin and method", func(t *testing.T) {
		cfg := &CORSConfig{
			AllowOrigins: "https://example.com",
			AllowMethods: "GET,POST",
		}
		cfg.init()
		allowed, headers := cfg.isPreflightAllowed("https://example.com", "POST", "")
		if !allowed {
			t.Error("should be allowed")
		}
		if headers != "" {
			t.Errorf("expected empty headers, got %s", headers)
		}
	})

	t.Run("disallowed origin", func(t *testing.T) {
		cfg := &CORSConfig{
			AllowOrigins: "https://example.com",
			AllowMethods: "GET,POST",
		}
		cfg.init()
		allowed, _ := cfg.isPreflightAllowed("https://foo.com", "POST", "")
		if allowed {
			t.Error("should not be allowed")
		}
	})

	t.Run("disallowed method", func(t *testing.T) {
		cfg := &CORSConfig{
			AllowOrigins: "https://example.com",
			AllowMethods: "GET,POST",
		}
		cfg.init()
		allowed, _ := cfg.isPreflightAllowed("https://example.com", "DELETE", "")
		if allowed {
			t.Error("should not be allowed")
		}
	})

	t.Run("disallowed headers", func(t *testing.T) {
		cfg := &CORSConfig{
			AllowOrigins: "https://example.com",
			AllowMethods: "GET,POST",
			AllowHeaders: "Content-Type",
		}
		cfg.init()
		allowed, _ := cfg.isPreflightAllowed("https://example.com", "POST", "X-Custom")
		if allowed {
			t.Error("should not be allowed")
		}
	})
}
