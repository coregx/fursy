// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGenericHEAD tests the generic HEAD() function registration.
func TestGenericHEAD(t *testing.T) {
	r := New()

	HEAD[Empty, Empty](r, "/resources/:id", func(c *Box[Empty, Empty]) error {
		c.SetHeader("X-Resource-ID", c.Param("id"))
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodHead, "/resources/42", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("X-Resource-ID") != "42" {
		t.Errorf("expected X-Resource-ID header '42', got %q", w.Header().Get("X-Resource-ID"))
	}

	// HEAD response must have empty body per HTTP spec.
	if w.Body.Len() != 0 {
		t.Errorf("HEAD response body must be empty, got %d bytes", w.Body.Len())
	}
}

// TestGenericHEAD_NotFound tests HEAD returning 404 when resource is absent.
func TestGenericHEAD_NotFound(t *testing.T) {
	r := New()

	HEAD[Empty, Empty](r, "/items/:id", func(c *Box[Empty, Empty]) error {
		id := c.Param("id")
		if id == "0" {
			return c.NoContent(http.StatusNotFound)
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodHead, "/items/0", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// TestGenericOPTIONS tests the generic OPTIONS() function registration.
func TestGenericOPTIONS(t *testing.T) {
	r := New()

	OPTIONS[Empty, Empty](r, "/users", func(c *Box[Empty, Empty]) error {
		c.SetHeader("Allow", "GET, POST, PUT, DELETE, OPTIONS")
		c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		c.SetHeader("Access-Control-Allow-Origin", "*")
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/users", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	allow := w.Header().Get("Allow")
	if allow != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("expected Allow header 'GET, POST, PUT, DELETE, OPTIONS', got %q", allow)
	}

	acam := w.Header().Get("Access-Control-Allow-Methods")
	if acam != "GET, POST, PUT, DELETE" {
		t.Errorf("expected Access-Control-Allow-Methods 'GET, POST, PUT, DELETE', got %q", acam)
	}
}

// TestGenericOPTIONS_WithParam tests OPTIONS on a parameterized path.
func TestGenericOPTIONS_WithParam(t *testing.T) {
	r := New()

	OPTIONS[Empty, Empty](r, "/users/:id", func(c *Box[Empty, Empty]) error {
		c.SetHeader("Allow", "GET, PUT, DELETE, OPTIONS")
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/users/123", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if w.Header().Get("Allow") != "GET, PUT, DELETE, OPTIONS" {
		t.Errorf("expected Allow header 'GET, PUT, DELETE, OPTIONS', got %q", w.Header().Get("Allow"))
	}
}
