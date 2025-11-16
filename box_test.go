// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test types for generic context.
type TestRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TestResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

// TestGenericContext_New tests creating a new generic context.
func TestGenericContext_New(t *testing.T) {
	base := newContext()
	ctx := newBox[TestRequest, TestResponse](base)

	if ctx.Context == nil {
		t.Fatal("Context should not be nil")
	}

	if ctx.ReqBody != nil {
		t.Error("ReqBody should be nil initially")
	}

	if ctx.ResBody != nil {
		t.Error("ResBody should be nil initially")
	}
}

// TestGenericContext_Bind_JSON tests JSON request binding.
func TestGenericContext_Bind_JSON(t *testing.T) {
	r := New()

	POST[TestRequest, TestResponse](r, "/test", func(c *Box[TestRequest, TestResponse]) error {
		if c.ReqBody == nil {
			return c.BadRequest(TestResponse{Message: "Request body is nil"})
		}

		if c.ReqBody.Name != "John" || c.ReqBody.Email != "john@example.com" {
			return c.BadRequest(TestResponse{Message: "Invalid request body"})
		}

		return c.OK(TestResponse{ID: 1, Message: "Success"})
	})

	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	expectedBody := `{"id":1,"message":"Success"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// TestGenericContext_Bind_EmptyType tests binding with Empty type.
func TestGenericContext_Bind_EmptyType(t *testing.T) {
	r := New()

	GET[Empty, TestResponse](r, "/test", func(c *Box[Empty, TestResponse]) error {
		// ReqBody should be nil for Empty type
		if c.ReqBody != nil {
			return c.InternalServerError(TestResponse{Message: "ReqBody should be nil for Empty type"})
		}

		return c.OK(TestResponse{ID: 1, Message: "Success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestGenericContext_OK tests OK response method.
func TestGenericContext_OK(t *testing.T) {
	r := New()

	GET[Empty, TestResponse](r, "/test", func(c *Box[Empty, TestResponse]) error {
		return c.OK(TestResponse{ID: 123, Message: "Hello"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	expectedBody := `{"id":123,"message":"Hello"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// TestGenericContext_Created tests Created response method.
func TestGenericContext_Created(t *testing.T) {
	r := New()

	POST[TestRequest, TestResponse](r, "/test", func(c *Box[TestRequest, TestResponse]) error {
		return c.Created("/test/123", TestResponse{ID: 123, Message: "Created"})
	})

	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/test/123" {
		t.Errorf("expected Location header %q, got %q", "/test/123", location)
	}

	expectedBody := `{"id":123,"message":"Created"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// TestGenericContext_Accepted tests Accepted response method.
func TestGenericContext_Accepted(t *testing.T) {
	r := New()

	POST[TestRequest, TestResponse](r, "/test", func(c *Box[TestRequest, TestResponse]) error {
		return c.Accepted(TestResponse{ID: 456, Message: "Accepted"})
	})

	body := `{"name":"Jane","email":"jane@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 202 {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	expectedBody := `{"id":456,"message":"Accepted"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// TestGenericContext_BadRequest tests BadRequest response method.
func TestGenericContext_BadRequest(t *testing.T) {
	r := New()

	POST[TestRequest, TestResponse](r, "/test", func(c *Box[TestRequest, TestResponse]) error {
		return c.BadRequest(TestResponse{Message: "Invalid input"})
	})

	body := `{"name":"","email":""}` // Empty fields
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	expectedBody := `{"id":0,"message":"Invalid input"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// TestGenericContext_NotFound tests NotFound response method.
func TestGenericContext_NotFound(t *testing.T) {
	r := New()

	GET[Empty, TestResponse](r, "/test/:id", func(c *Box[Empty, TestResponse]) error {
		return c.NotFound(TestResponse{Message: "Resource not found"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test/999", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	expectedBody := `{"id":0,"message":"Resource not found"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// TestGenericContext_InternalServerError tests InternalServerError response method.
func TestGenericContext_InternalServerError(t *testing.T) {
	r := New()

	GET[Empty, TestResponse](r, "/test", func(c *Box[Empty, TestResponse]) error {
		return c.InternalServerError(TestResponse{Message: "Database error"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	expectedBody := `{"id":0,"message":"Database error"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// TestGenericContext_WithParameters tests generic context with URL parameters.
func TestGenericContext_WithParameters(t *testing.T) {
	r := New()

	GET[Empty, TestResponse](r, "/users/:id", func(c *Box[Empty, TestResponse]) error {
		id := c.Param("id")
		return c.OK(TestResponse{Message: "User ID: " + id})
	})

	req := httptest.NewRequest(http.MethodGet, "/users/123", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	expectedBody := `{"id":0,"message":"User ID: 123"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// TestGenericContext_InvalidJSON tests binding with invalid JSON.
func TestGenericContext_InvalidJSON(t *testing.T) {
	r := New()

	POST[TestRequest, TestResponse](r, "/test", func(c *Box[TestRequest, TestResponse]) error {
		// Should not reach here due to binding error
		return c.OK(TestResponse{Message: "Should not reach here"})
	})

	body := `{invalid json}` // Invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should return 500 due to binding error (handled by router error handling)
	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// TestBox_NoContentSuccess tests the NoContentSuccess convenience method.
func TestBox_NoContentSuccess(t *testing.T) {
	r := New()

	DELETE[Empty, Empty](r, "/users/:id", func(c *Box[Empty, Empty]) error {
		// Simulate deletion
		return c.NoContentSuccess()
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/123", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should return 204 No Content
	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Body should be empty
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", w.Body.String())
	}
}

// TestBox_UpdatedOK tests the UpdatedOK convenience method.
func TestBox_UpdatedOK(t *testing.T) {
	r := New()

	PUT[TestRequest, TestResponse](r, "/users/:id", func(c *Box[TestRequest, TestResponse]) error {
		// Simulate update with response body
		return c.UpdatedOK(TestResponse{
			ID:      1,
			Message: "User updated: " + c.ReqBody.Name,
		})
	})

	body := `{"name":"Updated name"}`
	req := httptest.NewRequest(http.MethodPut, "/users/123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should return 200 OK (same as OK method)
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check body
	expectedBody := `{"id":1,"message":"User updated: Updated name"}` + "\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// TestBox_UpdatedNoContent tests the UpdatedNoContent convenience method.
func TestBox_UpdatedNoContent(t *testing.T) {
	r := New()

	PUT[TestRequest, Empty](r, "/users/:id", func(c *Box[TestRequest, Empty]) error {
		// Simulate update without response body
		return c.UpdatedNoContent()
	})

	body := `{"name":"Updated name"}`
	req := httptest.NewRequest(http.MethodPut, "/users/123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should return 204 No Content
	if w.Code != 204 {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Body should be empty
	if w.Body.Len() != 0 {
		t.Errorf("expected empty body, got %q", w.Body.String())
	}
}

// TestBox_ConvenienceMethods_RESTWorkflow tests complete REST workflow with Box convenience methods.
func TestBox_ConvenienceMethods_RESTWorkflow(t *testing.T) {
	r := New()

	// Simulated storage
	users := make(map[int]TestResponse)
	nextID := 1

	// POST - Create user (201 Created)
	POST[TestRequest, TestResponse](r, "/users", func(c *Box[TestRequest, TestResponse]) error {
		user := TestResponse{
			ID:      nextID,
			Message: c.ReqBody.Name,
		}
		users[nextID] = user
		nextID++
		return c.Created("/users/"+string(rune(user.ID+'0')), user)
	})

	// PUT - Update user with body (200 OK)
	PUT[TestRequest, TestResponse](r, "/users/:id", func(c *Box[TestRequest, TestResponse]) error {
		id := 1 // Simplified - normally would parse from params
		updated := TestResponse{
			ID:      id,
			Message: c.ReqBody.Name + " (updated)",
		}
		users[id] = updated
		return c.UpdatedOK(updated)
	})

	// PATCH - Update user without body (204 No Content)
	PATCH[TestRequest, Empty](r, "/users/:id/status", func(c *Box[TestRequest, Empty]) error {
		// Simulate status update
		return c.UpdatedNoContent()
	})

	// DELETE - Delete user (204 No Content)
	DELETE[Empty, Empty](r, "/users/:id", func(c *Box[Empty, Empty]) error {
		id := 1 // Simplified
		delete(users, id)
		return c.NoContentSuccess()
	})

	// Test POST (create user)
	body := `{"name":"John Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Errorf("POST: expected status 201, got %d", w.Code)
	}

	// Test PUT (update user with response)
	body = `{"name":"Jane Doe"}`
	req = httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("PUT: expected status 200, got %d", w.Code)
	}

	// Test PATCH (update status without response)
	body = `{"name":"active"}`
	req = httptest.NewRequest(http.MethodPatch, "/users/1/status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Errorf("PATCH: expected status 204, got %d", w.Code)
	}

	// Test DELETE (remove user)
	req = httptest.NewRequest(http.MethodDelete, "/users/1", http.NoBody)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Errorf("DELETE: expected status 204, got %d", w.Code)
	}
}
