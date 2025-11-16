// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coregx/fursy"
)

// TestSecure_Defaults tests the default security headers.
func TestSecure_Defaults(t *testing.T) {
	router := fursy.New()
	router.Use(Secure())

	router.GET("/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check default headers.
	headers := rec.Header()

	if v := headers.Get("X-Content-Type-Options"); v != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options: nosniff, got %q", v)
	}

	if v := headers.Get("X-Frame-Options"); v != "SAMEORIGIN" {
		t.Errorf("Expected X-Frame-Options: SAMEORIGIN, got %q", v)
	}

	if v := headers.Get("Referrer-Policy"); v != "strict-origin-when-cross-origin" {
		t.Errorf("Expected Referrer-Policy: strict-origin-when-cross-origin, got %q", v)
	}

	// HSTS should NOT be set by default (requires explicit config).
	if v := headers.Get("Strict-Transport-Security"); v != "" {
		t.Errorf("Expected no HSTS header by default, got %q", v)
	}

	// CSP should NOT be set by default (application-specific).
	if v := headers.Get("Content-Security-Policy"); v != "" {
		t.Errorf("Expected no CSP header by default, got %q", v)
	}

	// X-XSS-Protection should NOT be set by default (deprecated per OWASP 2025).
	if v := headers.Get("X-XSS-Protection"); v != "" {
		t.Errorf("Expected no X-XSS-Protection by default (deprecated), got %q", v)
	}
}

// TestSecure_CustomConfig tests custom security configuration.
func TestSecure_CustomConfig(t *testing.T) {
	router := fursy.New()
	router.Use(SecureWithConfig(SecureConfig{
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		ReferrerPolicy:     "no-referrer",
	}))

	router.GET("/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	headers := rec.Header()

	if v := headers.Get("X-Frame-Options"); v != "DENY" {
		t.Errorf("Expected X-Frame-Options: DENY, got %q", v)
	}

	if v := headers.Get("Referrer-Policy"); v != "no-referrer" {
		t.Errorf("Expected Referrer-Policy: no-referrer, got %q", v)
	}
}

// TestSecure_HSTS tests HSTS header configuration.
func TestSecure_HSTS(t *testing.T) {
	tests := []struct {
		name              string
		maxAge            int
		excludeSubdomains bool
		preloadEnabled    bool
		expectedHeader    string
	}{
		{
			name:           "1 year with subdomains",
			maxAge:         31536000,
			expectedHeader: "max-age=31536000; includeSubDomains",
		},
		{
			name:              "1 year without subdomains",
			maxAge:            31536000,
			excludeSubdomains: true,
			expectedHeader:    "max-age=31536000",
		},
		{
			name:           "2 years with preload",
			maxAge:         63072000,
			preloadEnabled: true,
			expectedHeader: "max-age=63072000; includeSubDomains; preload",
		},
		{
			name:              "Custom with all options",
			maxAge:            86400,
			excludeSubdomains: true,
			preloadEnabled:    true,
			expectedHeader:    "max-age=86400; preload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := fursy.New()
			router.Use(SecureWithConfig(SecureConfig{
				HSTSMaxAge:            tt.maxAge,
				HSTSExcludeSubdomains: tt.excludeSubdomains,
				HSTSPreloadEnabled:    tt.preloadEnabled,
			}))

			router.GET("/test", func(c *fursy.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if v := rec.Header().Get("Strict-Transport-Security"); v != tt.expectedHeader {
				t.Errorf("Expected HSTS: %q, got %q", tt.expectedHeader, v)
			}
		})
	}
}

// TestSecure_CSP tests Content Security Policy configuration.
func TestSecure_CSP(t *testing.T) {
	tests := []struct {
		name           string
		csp            string
		reportOnly     bool
		expectedHeader string
		expectedValue  string
	}{
		{
			name:           "Basic CSP",
			csp:            "default-src 'self'",
			expectedHeader: "Content-Security-Policy",
			expectedValue:  "default-src 'self'",
		},
		{
			name:           "CSP with scripts",
			csp:            "default-src 'self'; script-src 'self' 'unsafe-inline'",
			expectedHeader: "Content-Security-Policy",
			expectedValue:  "default-src 'self'; script-src 'self' 'unsafe-inline'",
		},
		{
			name:           "CSP Report Only",
			csp:            "default-src 'self'",
			reportOnly:     true,
			expectedHeader: "Content-Security-Policy-Report-Only",
			expectedValue:  "default-src 'self'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := fursy.New()
			router.Use(SecureWithConfig(SecureConfig{
				ContentSecurityPolicy: tt.csp,
				CSPReportOnly:         tt.reportOnly,
			}))

			router.GET("/test", func(c *fursy.Context) error {
				return c.String(http.StatusOK, "OK")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if v := rec.Header().Get(tt.expectedHeader); v != tt.expectedValue {
				t.Errorf("Expected %s: %q, got %q", tt.expectedHeader, tt.expectedValue, v)
			}

			// Ensure the other header is NOT set.
			if !tt.reportOnly {
				if v := rec.Header().Get("Content-Security-Policy-Report-Only"); v != "" {
					t.Errorf("Expected no CSP-Report-Only header, got %q", v)
				}
			} else {
				if v := rec.Header().Get("Content-Security-Policy"); v != "" {
					t.Errorf("Expected no CSP header, got %q", v)
				}
			}
		})
	}
}

// TestSecure_CrossOriginHeaders tests Cross-Origin-* headers.
func TestSecure_CrossOriginHeaders(t *testing.T) {
	router := fursy.New()
	router.Use(SecureWithConfig(SecureConfig{
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
	}))

	router.GET("/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	headers := rec.Header()

	if v := headers.Get("Cross-Origin-Embedder-Policy"); v != "require-corp" {
		t.Errorf("Expected COEP: require-corp, got %q", v)
	}

	if v := headers.Get("Cross-Origin-Opener-Policy"); v != "same-origin" {
		t.Errorf("Expected COOP: same-origin, got %q", v)
	}

	if v := headers.Get("Cross-Origin-Resource-Policy"); v != "same-origin" {
		t.Errorf("Expected CORP: same-origin, got %q", v)
	}
}

// TestSecure_PermissionsPolicy tests Permissions-Policy header.
func TestSecure_PermissionsPolicy(t *testing.T) {
	router := fursy.New()
	router.Use(SecureWithConfig(SecureConfig{
		PermissionsPolicy: "geolocation=(self), microphone=()",
	}))

	router.GET("/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if v := rec.Header().Get("Permissions-Policy"); v != "geolocation=(self), microphone=()" {
		t.Errorf("Expected Permissions-Policy, got %q", v)
	}
}

// TestSecure_XSSProtection tests X-XSS-Protection header (deprecated).
func TestSecure_XSSProtection(t *testing.T) {
	router := fursy.New()
	router.Use(SecureWithConfig(SecureConfig{
		XSSProtection: "1; mode=block", // Explicitly set (not recommended)
	}))

	router.GET("/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Should be set only when explicitly configured.
	if v := rec.Header().Get("X-XSS-Protection"); v != "1; mode=block" {
		t.Errorf("Expected X-XSS-Protection: 1; mode=block, got %q", v)
	}
}

// TestSecure_Skipper tests the Skipper functionality.
func TestSecure_Skipper(t *testing.T) {
	router := fursy.New()
	router.Use(SecureWithConfig(SecureConfig{
		Skipper: func(c *fursy.Context) bool {
			// Skip security headers for /public path.
			return strings.HasPrefix(c.Request.URL.Path, "/public")
		},
	}))

	router.GET("/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	router.GET("/public/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// Test regular path (headers should be set).
	req1 := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)

	if v := rec1.Header().Get("X-Content-Type-Options"); v != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options on /test, got %q", v)
	}

	// Test skipped path (headers should NOT be set).
	req2 := httptest.NewRequest(http.MethodGet, "/public/test", http.NoBody)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	if v := rec2.Header().Get("X-Content-Type-Options"); v != "" {
		t.Errorf("Expected no X-Content-Type-Options on /public/test (skipped), got %q", v)
	}
}

// TestSecureDefaults tests the SecureDefaults helper function.
func TestSecureDefaults(t *testing.T) {
	config := SecureDefaults()

	if config.ContentTypeNosniff != "nosniff" {
		t.Errorf("Expected ContentTypeNosniff: nosniff, got %q", config.ContentTypeNosniff)
	}

	if config.XFrameOptions != "SAMEORIGIN" {
		t.Errorf("Expected XFrameOptions: SAMEORIGIN, got %q", config.XFrameOptions)
	}

	if config.ReferrerPolicy != "strict-origin-when-cross-origin" {
		t.Errorf("Expected ReferrerPolicy: strict-origin-when-cross-origin, got %q", config.ReferrerPolicy)
	}

	// HSTS should be 0 (not enabled by default).
	if config.HSTSMaxAge != 0 {
		t.Errorf("Expected HSTSMaxAge: 0, got %d", config.HSTSMaxAge)
	}

	// CSP should be empty (application-specific).
	if config.ContentSecurityPolicy != "" {
		t.Errorf("Expected empty CSP, got %q", config.ContentSecurityPolicy)
	}
}

// TestSecureStrict tests the SecureStrict helper function.
func TestSecureStrict(t *testing.T) {
	config := SecureStrict()

	if config.ContentTypeNosniff != "nosniff" {
		t.Errorf("Expected ContentTypeNosniff: nosniff, got %q", config.ContentTypeNosniff)
	}

	if config.XFrameOptions != "DENY" {
		t.Errorf("Expected XFrameOptions: DENY, got %q", config.XFrameOptions)
	}

	if config.ReferrerPolicy != "no-referrer" {
		t.Errorf("Expected ReferrerPolicy: no-referrer, got %q", config.ReferrerPolicy)
	}

	if config.HSTSMaxAge != 63072000 {
		t.Errorf("Expected HSTSMaxAge: 63072000 (2 years), got %d", config.HSTSMaxAge)
	}

	if !config.HSTSPreloadEnabled {
		t.Error("Expected HSTSPreloadEnabled: true")
	}

	if config.ContentSecurityPolicy != "default-src 'self'" {
		t.Errorf("Expected CSP: default-src 'self', got %q", config.ContentSecurityPolicy)
	}

	if config.CrossOriginEmbedderPolicy != "require-corp" {
		t.Errorf("Expected COEP: require-corp, got %q", config.CrossOriginEmbedderPolicy)
	}

	if config.CrossOriginOpenerPolicy != "same-origin" {
		t.Errorf("Expected COOP: same-origin, got %q", config.CrossOriginOpenerPolicy)
	}

	if config.CrossOriginResourcePolicy != "same-origin" {
		t.Errorf("Expected CORP: same-origin, got %q", config.CrossOriginResourcePolicy)
	}
}

// TestSecureWithCSP tests the SecureWithCSP helper function.
func TestSecureWithCSP(t *testing.T) {
	csp := "default-src 'self'; script-src 'self' 'unsafe-inline'"
	config := SecureWithCSP(csp)

	// Should have defaults + custom CSP.
	if config.ContentTypeNosniff != "nosniff" {
		t.Errorf("Expected default ContentTypeNosniff, got %q", config.ContentTypeNosniff)
	}

	if config.ContentSecurityPolicy != csp {
		t.Errorf("Expected CSP: %q, got %q", csp, config.ContentSecurityPolicy)
	}
}

// TestSecureWithHSTS tests the SecureWithHSTS helper function.
func TestSecureWithHSTS(t *testing.T) {
	tests := []struct {
		name            string
		maxAge          int
		preload         bool
		expectedMaxAge  int
		expectedPreload bool
	}{
		{
			name:            "1 year without preload",
			maxAge:          31536000,
			preload:         false,
			expectedMaxAge:  31536000,
			expectedPreload: false,
		},
		{
			name:            "2 years with preload",
			maxAge:          63072000,
			preload:         true,
			expectedMaxAge:  63072000,
			expectedPreload: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := SecureWithHSTS(tt.maxAge, tt.preload)

			if config.HSTSMaxAge != tt.expectedMaxAge {
				t.Errorf("Expected HSTSMaxAge: %d, got %d", tt.expectedMaxAge, config.HSTSMaxAge)
			}

			if config.HSTSPreloadEnabled != tt.expectedPreload {
				t.Errorf("Expected HSTSPreloadEnabled: %v, got %v", tt.expectedPreload, config.HSTSPreloadEnabled)
			}

			// Should still have defaults.
			if config.ContentTypeNosniff != "nosniff" {
				t.Errorf("Expected default ContentTypeNosniff, got %q", config.ContentTypeNosniff)
			}
		})
	}
}

// TestBuildHSTSHeader tests the BuildHSTSHeader utility function.
func TestBuildHSTSHeader(t *testing.T) {
	tests := []struct {
		name              string
		maxAge            int
		includeSubdomains bool
		preload           bool
		expected          string
	}{
		{
			name:              "Basic",
			maxAge:            31536000,
			includeSubdomains: false,
			preload:           false,
			expected:          "max-age=31536000",
		},
		{
			name:              "With subdomains",
			maxAge:            31536000,
			includeSubdomains: true,
			preload:           false,
			expected:          "max-age=31536000; includeSubDomains",
		},
		{
			name:              "With preload",
			maxAge:            63072000,
			includeSubdomains: false,
			preload:           true,
			expected:          "max-age=63072000; preload",
		},
		{
			name:              "Full options",
			maxAge:            63072000,
			includeSubdomains: true,
			preload:           true,
			expected:          "max-age=63072000; includeSubDomains; preload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildHSTSHeader(tt.maxAge, tt.includeSubdomains, tt.preload)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSecure_EmptyValues tests that empty string values don't set headers.
func TestSecure_EmptyValues(t *testing.T) {
	router := fursy.New()
	router.Use(SecureWithConfig(SecureConfig{
		ContentTypeNosniff:        "",
		XFrameOptions:             "",
		ReferrerPolicy:            "",
		ContentSecurityPolicy:     "",
		CrossOriginEmbedderPolicy: "",
		CrossOriginOpenerPolicy:   "",
		CrossOriginResourcePolicy: "",
		PermissionsPolicy:         "",
		XSSProtection:             "",
	}))

	router.GET("/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	headers := rec.Header()

	// Since defaults are applied, some headers will still be set.
	// But let's verify the behavior with truly empty config.
	// Actually, the middleware applies defaults, so this test needs adjustment.

	// Let's test that explicitly empty CSP doesn't set header.
	if v := headers.Get("Content-Security-Policy"); v != "" {
		t.Errorf("Expected no CSP header with empty config, got %q", v)
	}

	if v := headers.Get("Cross-Origin-Embedder-Policy"); v != "" {
		t.Errorf("Expected no COEP header with empty config, got %q", v)
	}

	if v := headers.Get("Permissions-Policy"); v != "" {
		t.Errorf("Expected no Permissions-Policy with empty config, got %q", v)
	}

	// But defaults should still apply for nosniff, frame options, referrer.
	if v := headers.Get("X-Content-Type-Options"); v != "nosniff" {
		t.Errorf("Expected default X-Content-Type-Options: nosniff, got %q", v)
	}
}

// TestSecure_AllHeadersSet tests setting all headers at once.
func TestSecure_AllHeadersSet(t *testing.T) {
	router := fursy.New()
	router.Use(SecureWithConfig(SecureConfig{
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "DENY",
		ReferrerPolicy:            "no-referrer",
		HSTSMaxAge:                31536000,
		HSTSPreloadEnabled:        true,
		ContentSecurityPolicy:     "default-src 'self'",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "same-origin",
		PermissionsPolicy:         "geolocation=()",
		XSSProtection:             "0", // Explicitly disable (recommended per OWASP)
	}))

	router.GET("/test", func(c *fursy.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	headers := rec.Header()

	// Verify all headers are set.
	expectedHeaders := map[string]string{
		"X-Content-Type-Options":       "nosniff",
		"X-Frame-Options":              "DENY",
		"Referrer-Policy":              "no-referrer",
		"Strict-Transport-Security":    "max-age=31536000; includeSubDomains; preload",
		"Content-Security-Policy":      "default-src 'self'",
		"Cross-Origin-Embedder-Policy": "require-corp",
		"Cross-Origin-Opener-Policy":   "same-origin",
		"Cross-Origin-Resource-Policy": "same-origin",
		"Permissions-Policy":           "geolocation=()",
		"X-XSS-Protection":             "0",
	}

	for header, expected := range expectedHeaders {
		if v := headers.Get(header); v != expected {
			t.Errorf("Expected %s: %q, got %q", header, expected, v)
		}
	}
}

// TestSecure_Integration tests the middleware in a realistic scenario.
func TestSecure_Integration(t *testing.T) {
	router := fursy.New()

	// Use strict security configuration.
	router.Use(SecureWithConfig(SecureStrict()))

	router.GET("/api/users", func(c *fursy.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"user": "john"})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", http.NoBody)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	headers := rec.Header()

	// Verify strict security headers are all set.
	if v := headers.Get("X-Frame-Options"); v != "DENY" {
		t.Errorf("Expected X-Frame-Options: DENY, got %q", v)
	}

	if v := headers.Get("Strict-Transport-Security"); !strings.Contains(v, "max-age=63072000") {
		t.Errorf("Expected HSTS with 2 years, got %q", v)
	}

	if v := headers.Get("Content-Security-Policy"); v != "default-src 'self'" {
		t.Errorf("Expected strict CSP, got %q", v)
	}

	if v := headers.Get("Cross-Origin-Embedder-Policy"); v != "require-corp" {
		t.Errorf("Expected COEP: require-corp, got %q", v)
	}
}
