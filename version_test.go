// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test Version struct methods.

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name     string
		version  Version
		expected string
	}{
		{
			name:     "major only",
			version:  Version{Major: 1},
			expected: "v1",
		},
		{
			name:     "major.minor",
			version:  Version{Major: 2, Minor: 1},
			expected: "v2.1",
		},
		{
			name:     "major.minor.patch",
			version:  Version{Major: 3, Minor: 2, Patch: 1},
			expected: "v3.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestVersion_Equal(t *testing.T) {
	v1 := Version{Major: 1, Minor: 2, Patch: 3}
	v2 := Version{Major: 1, Minor: 2, Patch: 3}
	v3 := Version{Major: 2, Minor: 0, Patch: 0}

	if !v1.Equal(v2) {
		t.Error("Expected v1 to equal v2")
	}

	if v1.Equal(v3) {
		t.Error("Expected v1 not to equal v3")
	}
}

func TestVersion_GreaterThan(t *testing.T) {
	tests := []struct {
		name     string
		v1       Version
		v2       Version
		expected bool
	}{
		{
			name:     "major greater",
			v1:       Version{Major: 2},
			v2:       Version{Major: 1},
			expected: true,
		},
		{
			name:     "minor greater",
			v1:       Version{Major: 1, Minor: 2},
			v2:       Version{Major: 1, Minor: 1},
			expected: true,
		},
		{
			name:     "patch greater",
			v1:       Version{Major: 1, Minor: 1, Patch: 2},
			v2:       Version{Major: 1, Minor: 1, Patch: 1},
			expected: true,
		},
		{
			name:     "equal",
			v1:       Version{Major: 1},
			v2:       Version{Major: 1},
			expected: false,
		},
		{
			name:     "less than",
			v1:       Version{Major: 1},
			v2:       Version{Major: 2},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v1.GreaterThan(tt.v2)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestVersion_LessThan(t *testing.T) {
	v1 := Version{Major: 1}
	v2 := Version{Major: 2}
	v3 := Version{Major: 1}

	if !v1.LessThan(v2) {
		t.Error("Expected v1 < v2")
	}

	if v2.LessThan(v1) {
		t.Error("Expected !(v2 < v1)")
	}

	if v1.LessThan(v3) {
		t.Error("Expected !(v1 < v3) when equal")
	}
}

// Test ParseVersion function.

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Version
		ok       bool
	}{
		{
			name:     "v1",
			input:    "v1",
			expected: Version{Major: 1},
			ok:       true,
		},
		{
			name:     "V2 (uppercase)",
			input:    "V2",
			expected: Version{Major: 2},
			ok:       true,
		},
		{
			name:     "1 (no prefix)",
			input:    "1",
			expected: Version{Major: 1},
			ok:       true,
		},
		{
			name:     "v2.1",
			input:    "v2.1",
			expected: Version{Major: 2, Minor: 1},
			ok:       true,
		},
		{
			name:     "v3.2.1",
			input:    "v3.2.1",
			expected: Version{Major: 3, Minor: 2, Patch: 1},
			ok:       true,
		},
		{
			name:     "2.1.0",
			input:    "2.1.0",
			expected: Version{Major: 2, Minor: 1, Patch: 0},
			ok:       true,
		},
		{
			name:  "invalid: empty",
			input: "",
			ok:    false,
		},
		{
			name:  "invalid: non-numeric",
			input: "vX",
			ok:    false,
		},
		{
			name:  "invalid: negative",
			input: "v-1",
			ok:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ParseVersion(tt.input)

			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got ok=%v", tt.ok, ok)
				return
			}

			if !ok {
				return // Expected parse failure.
			}

			if !result.Equal(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test ExtractVersionFromPath function.

func TestExtractVersionFromPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected Version
		ok       bool
	}{
		{
			name:     "/v1/users",
			path:     "/v1/users",
			expected: Version{Major: 1},
			ok:       true,
		},
		{
			name:     "/api/v2/posts",
			path:     "/api/v2/posts",
			expected: Version{Major: 2},
			ok:       true,
		},
		{
			name:     "/api/v2.1/posts",
			path:     "/api/v2.1/posts",
			expected: Version{Major: 2, Minor: 1},
			ok:       true,
		},
		{
			name:     "/api/v3.2.1/posts",
			path:     "/api/v3.2.1/posts",
			expected: Version{Major: 3, Minor: 2, Patch: 1},
			ok:       true,
		},
		{
			name: "no version",
			path: "/api/users",
			ok:   false,
		},
		{
			name:     "version in middle",
			path:     "/users/v1/posts",
			expected: Version{Major: 1},
			ok:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ExtractVersionFromPath(tt.path)

			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got ok=%v", tt.ok, ok)
				return
			}

			if !ok {
				return // Expected extraction failure.
			}

			if !result.Equal(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test DeprecationInfo.

func TestDeprecationInfo_SetDeprecationHeaders(t *testing.T) {
	router := New()

	sunsetDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	router.GET("/test", func(c *Context) error {
		info := DeprecationInfo{
			Version:    Version{Major: 1},
			SunsetDate: &sunsetDate,
			Message:    "Please migrate to v2",
			Link:       "https://api.example.com/docs/v2-migration",
		}

		info.SetDeprecationHeaders(c)

		return c.String(200, "deprecated")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check Deprecation header.
	if w.Header().Get("Deprecation") != "true" {
		t.Errorf("Expected Deprecation: true header, got %s", w.Header().Get("Deprecation"))
	}

	// Check Sunset header (RFC 8594).
	sunset := w.Header().Get("Sunset")
	if sunset == "" {
		t.Error("Expected Sunset header")
	}

	// Check Link header.
	link := w.Header().Get("Link")
	if !contains(link, "rel=\"sunset\"") {
		t.Errorf("Expected Link header with rel=sunset, got %s", link)
	}

	// Check Warning header (RFC 7234).
	warning := w.Header().Get("Warning")
	if !contains(warning, "299") || !contains(warning, "deprecated") {
		t.Errorf("Expected Warning header with deprecation message, got %s", warning)
	}
}

// Test Box.APIVersion method.

func TestContext_APIVersion_FromHeader(t *testing.T) {
	router := New()
	router.GET("/test", func(c *Context) error {
		version := c.APIVersion()

		if version.Major != 2 {
			t.Errorf("Expected major version 2, got %d", version.Major)
		}

		return c.String(200, version.String())
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Api-Version", "2")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Body.String() != "v2" {
		t.Errorf("Expected v2, got %s", w.Body.String())
	}
}

func TestContext_APIVersion_FromPath(t *testing.T) {
	router := New()
	router.GET("/api/v1/test", func(c *Context) error {
		version := c.APIVersion()

		if version.Major != 1 {
			t.Errorf("Expected major version 1, got %d", version.Major)
		}

		return c.String(200, version.String())
	})

	req := httptest.NewRequest("GET", "/api/v1/test", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Body.String() != "v1" {
		t.Errorf("Expected v1, got %s", w.Body.String())
	}
}

func TestContext_APIVersion_HeaderPriority(t *testing.T) {
	router := New()
	router.GET("/api/v1/test", func(c *Context) error {
		version := c.APIVersion()

		// Header should take priority over path.
		if version.Major != 2 {
			t.Errorf("Expected major version 2 (from header), got %d", version.Major)
		}

		return c.String(200, version.String())
	})

	req := httptest.NewRequest("GET", "/api/v1/test", http.NoBody)
	req.Header.Set("Api-Version", "2")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Body.String() != "v2" {
		t.Errorf("Expected v2, got %s", w.Body.String())
	}
}

func TestContext_APIVersion_NoVersion(t *testing.T) {
	router := New()
	router.GET("/test", func(c *Context) error {
		version := c.APIVersion()

		// No version found - should be zero.
		if version.Major != 0 {
			t.Errorf("Expected zero version, got %d", version.Major)
		}

		return c.String(200, "no version")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
}

// Test RequireVersion middleware.

func TestRequireVersion_Success(t *testing.T) {
	router := New()

	v1 := router.Group("/api/v1")
	v1.Use(RequireVersion(Version{Major: 1}))
	v1.GET("/users", func(c *Context) error {
		return c.String(200, "v1 users")
	})

	req := httptest.NewRequest("GET", "/api/v1/users", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "v1 users" {
		t.Errorf("Expected 'v1 users', got %s", w.Body.String())
	}
}

func TestRequireVersion_Mismatch(t *testing.T) {
	router := New()

	v2 := router.Group("/api/v2")
	v2.Use(RequireVersion(Version{Major: 2}))
	v2.GET("/users", func(c *Context) error {
		return c.String(200, "v2 users")
	})

	// Request with v1 header but v2 path.
	req := httptest.NewRequest("GET", "/api/v2/users", http.NoBody)
	req.Header.Set("Api-Version", "1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 Bad Request (version mismatch).
	if w.Code != 400 {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestRequireVersion_NoVersion(t *testing.T) {
	router := New()

	v1 := router.Group("/api/v1")
	v1.Use(RequireVersion(Version{Major: 1}))
	v1.GET("/users", func(_ *Context) error {
		return nil
	})

	// Request without version.
	req := httptest.NewRequest("GET", "/api/v1/users", http.NoBody)
	// Remove path version by requesting wrong path.
	req.URL.Path = "/api/users"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404 (route not found).
	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// Test DeprecateVersion middleware.

func TestDeprecateVersion_Middleware(t *testing.T) {
	router := New()

	sunsetDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	v1 := router.Group("/api/v1")
	v1.Use(DeprecateVersion(DeprecationInfo{
		Version:    Version{Major: 1},
		SunsetDate: &sunsetDate,
		Message:    "Please migrate to v2",
		Link:       "https://api.example.com/docs/v2-migration",
	}))
	v1.GET("/users", func(c *Context) error {
		return c.String(200, "v1 users")
	})

	req := httptest.NewRequest("GET", "/api/v1/users", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check deprecation headers are set.
	if w.Header().Get("Deprecation") != "true" {
		t.Error("Expected Deprecation header")
	}

	if w.Header().Get("Sunset") == "" {
		t.Error("Expected Sunset header")
	}

	if w.Header().Get("Warning") == "" {
		t.Error("Expected Warning header")
	}
}

// Test integration: multiple versions with deprecation.

func TestAPIVersioning_Integration(t *testing.T) {
	router := New()

	sunsetDate := time.Date(2025, 6, 30, 23, 59, 59, 0, time.UTC)

	// v1 - deprecated.
	v1 := router.Group("/api/v1")
	v1.Use(DeprecateVersion(DeprecationInfo{
		Version:    Version{Major: 1},
		SunsetDate: &sunsetDate,
		Message:    "Migrate to v2 by June 2025",
		Link:       "https://api.example.com/docs/migration",
	}))
	v1.GET("/users", func(c *Context) error {
		return c.JSON(200, map[string]string{"version": "v1", "status": "deprecated"})
	})

	// v2 - current.
	v2 := router.Group("/api/v2")
	v2.GET("/users", func(c *Context) error {
		return c.JSON(200, map[string]string{"version": "v2", "status": "current"})
	})

	// Test v1 (deprecated).
	req1 := httptest.NewRequest("GET", "/api/v1/users", http.NoBody)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("v1: Expected status 200, got %d", w1.Code)
	}

	if w1.Header().Get("Deprecation") != "true" {
		t.Error("v1: Expected Deprecation header")
	}

	// Test v2 (current).
	req2 := httptest.NewRequest("GET", "/api/v2/users", http.NoBody)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Errorf("v2: Expected status 200, got %d", w2.Code)
	}

	if w2.Header().Get("Deprecation") != "" {
		t.Error("v2: Should not have Deprecation header")
	}
}
