// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Integration tests for Box.NegotiateFormat and Box.Negotiate methods.
// Unit tests for negotiate package are in internal/negotiate/negotiate_test.go.

// Test Box.NegotiateFormat method.

func TestContext_NegotiateFormat_NoAcceptHeader(t *testing.T) {
	router := New()
	router.GET("/test", func(c *Context) error {
		format := c.NegotiateFormat(MIMEApplicationJSON, MIMEApplicationXML, MIMETextHTML)

		// No Accept header - should return first offered.
		if format != MIMEApplicationJSON {
			t.Errorf("Expected %s (first offered), got %s", MIMEApplicationJSON, format)
		}

		return c.String(200, format)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	// No Accept header set.
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestContext_NegotiateFormat_WithAcceptHeader(t *testing.T) {
	tests := []struct {
		name     string
		accept   string
		offered  []string
		expected string
	}{
		{
			name:     "JSON preferred",
			accept:   "application/json",
			offered:  []string{MIMEApplicationJSON, MIMEApplicationXML},
			expected: MIMEApplicationJSON,
		},
		{
			name:     "XML preferred with q-weighting",
			accept:   "application/xml;q=1.0, application/json;q=0.9",
			offered:  []string{MIMEApplicationJSON, MIMEApplicationXML},
			expected: MIMEApplicationXML,
		},
		{
			name:     "HTML preferred",
			accept:   "text/html, application/json;q=0.9",
			offered:  []string{MIMEApplicationJSON, MIMETextHTML},
			expected: MIMETextHTML,
		},
		{
			name:     "Wildcard matches first",
			accept:   "*/*",
			offered:  []string{MIMEApplicationXML, MIMEApplicationJSON},
			expected: MIMEApplicationXML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := New()
			router.GET("/test", func(c *Context) error {
				format := c.NegotiateFormat(tt.offered...)

				if format != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, format)
				}

				return c.String(200, format)
			})

			req := httptest.NewRequest("GET", "/test", http.NoBody)
			req.Header.Set("Accept", tt.accept)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

// Test Box.Negotiate method.

func TestContext_Negotiate_JSON(t *testing.T) {
	router := New()
	router.GET("/test", func(c *Context) error {
		data := map[string]string{"message": "hello"}
		return c.Negotiate(200, data)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	vary := w.Header().Get("Vary")
	if vary != "Accept" {
		t.Errorf("Expected Vary: Accept header, got %s", vary)
	}

	if w.Body.String() != `{"message":"hello"}`+"\n" {
		t.Errorf("Expected JSON body, got %s", w.Body.String())
	}
}

func TestContext_Negotiate_XML(t *testing.T) {
	type TestData struct {
		Message string `xml:"message"`
	}

	router := New()
	router.GET("/test", func(c *Context) error {
		data := TestData{Message: "hello"}
		return c.Negotiate(200, data)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "application/xml")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/xml; charset=utf-8" {
		t.Errorf("Expected Content-Type application/xml, got %s", contentType)
	}
}

func TestContext_Negotiate_PlainText(t *testing.T) {
	router := New()
	router.GET("/test", func(c *Context) error {
		data := "Hello, World!"
		return c.Negotiate(200, data)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "text/plain")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !contains(contentType, "text/plain") {
		t.Errorf("Expected Content-Type text/plain, got %s", contentType)
	}
}

func TestContext_Negotiate_NoAcceptableFormat(t *testing.T) {
	router := New()
	router.GET("/test", func(c *Context) error {
		data := map[string]string{"message": "hello"}
		return c.Negotiate(200, data)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("Accept", "video/mp4") // Unsupported format.
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 406 {
		t.Errorf("Expected status 406 Not Acceptable, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/problem+json; charset=utf-8" {
		t.Errorf("Expected Content-Type application/problem+json, got %s", contentType)
	}
}

func TestContext_Negotiate_QWeightingSelection(t *testing.T) {
	type TestData struct {
		Message string `json:"message" xml:"message"`
	}

	router := New()
	router.GET("/test", func(c *Context) error {
		data := TestData{Message: "hello"}
		return c.Negotiate(200, data)
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	// XML preferred over JSON.
	req.Header.Set("Accept", "application/xml;q=1.0, application/json;q=0.9")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/xml; charset=utf-8" {
		t.Errorf("Expected Content-Type application/xml (higher q), got %s", contentType)
	}
}

// Note: contains helper function is defined in validation_test.go
