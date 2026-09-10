// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- TDD: Error Pipeline Tests (write BEFORE implementation) ---
//
// These tests verify that handler errors are mapped to correct HTTP
// status codes and RFC 9457 Problem Details responses, instead of
// always returning plain-text 500.

// errorHandlerReq is a test request type with validation tags.
type errorHandlerReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// errorHandlerRes is a test response type.
type errorHandlerRes struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

// testValidator validates errorHandlerReq — name must be non-empty.
type testValidator struct{}

func (v *testValidator) Validate(i interface{}) error {
	if req, ok := i.(*errorHandlerReq); ok && req.Name == "" {
		return ValidationErrors{
			{Field: "name", Message: "name is required"},
		}
	}
	return nil
}

// TestErrorHandler_ValidationError_Returns422 verifies that a validation
// error from Bind() produces a 422 response with RFC 9457 Problem body,
// not a plain-text 500.
func TestErrorHandler_ValidationError_Returns422(t *testing.T) {
	r := New()
	r.SetValidator(&testValidator{})

	r.POST("/users", func(c *Box[errorHandlerReq, errorHandlerRes]) error {
		return c.OK(errorHandlerRes{ID: 1, Message: "created"})
	})

	body := `{"name":"","email":"test@example.com"}`
	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("validation error: want status 422, got %d; body: %s", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/problem+json") {
		t.Errorf("validation error: want Content-Type application/problem+json, got %q", ct)
	}
}

// TestErrorHandler_ProblemNotFound_Returns404 verifies that returning a
// Problem with Status 404 produces a 404 response, not 500.
func TestErrorHandler_ProblemNotFound_Returns404(t *testing.T) {
	r := New()

	r.Handle("GET", "/users/:id", func(_ *Context) error {
		return NotFound("user not found")
	})

	req := httptest.NewRequest("GET", "/users/999", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Problem NotFound: want status 404, got %d; body: %s", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/problem+json") {
		t.Errorf("Problem NotFound: want Content-Type application/problem+json, got %q", ct)
	}

	var prob map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &prob); err != nil {
		t.Fatalf("Problem NotFound: response not valid JSON: %v", err)
	}
	if prob["status"] != float64(404) {
		t.Errorf("Problem NotFound: want status=404 in body, got %v", prob["status"])
	}
}

// TestErrorHandler_ProblemBadRequest_Returns400 verifies Problem with 400.
func TestErrorHandler_ProblemBadRequest_Returns400(t *testing.T) {
	r := New()

	r.Handle("POST", "/data", func(_ *Context) error {
		return BadRequest("invalid input")
	})

	req := httptest.NewRequest("POST", "/data", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Problem BadRequest: want status 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

// TestErrorHandler_BindingMalformedJSON_Returns400 verifies that
// malformed JSON in request body produces 400, not 500.
func TestErrorHandler_BindingMalformedJSON_Returns400(t *testing.T) {
	r := New()

	r.POST("/users", func(c *Box[errorHandlerReq, errorHandlerRes]) error {
		return c.OK(errorHandlerRes{ID: 1, Message: "created"})
	})

	body := `{invalid json`
	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("malformed JSON: want status 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

// TestErrorHandler_BindingEmptyBody_Returns400 verifies that POST with
// empty body produces 400, not 500.
func TestErrorHandler_BindingEmptyBody_Returns400(t *testing.T) {
	r := New()

	r.POST("/users", func(c *Box[errorHandlerReq, errorHandlerRes]) error {
		return c.OK(errorHandlerRes{ID: 1, Message: "created"})
	})

	req := httptest.NewRequest("POST", "/users", http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = 0
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("empty body: want status 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

// TestErrorHandler_UnknownError_Returns500_NoDetails verifies that an
// unknown error produces 500 WITHOUT exposing error details.
func TestErrorHandler_UnknownError_Returns500_NoDetails(t *testing.T) {
	r := New()

	r.Handle("GET", "/crash", func(_ *Context) error {
		return errors.New("database connection failed: host=prod-db password=secret")
	})

	req := httptest.NewRequest("GET", "/crash", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("unknown error: want status 500, got %d", w.Code)
	}

	body := w.Body.String()
	if strings.Contains(body, "database") || strings.Contains(body, "password") || strings.Contains(body, "secret") {
		t.Errorf("unknown error: SECURITY — error details leaked to client: %s", body)
	}
}

// TestErrorHandler_Custom verifies that SetErrorHandler overrides the default.
func TestErrorHandler_Custom(t *testing.T) {
	r := New()
	r.SetErrorHandler(func(c *Context, err error) {
		c.SetHeader("X-Custom-Error", "true")
		_ = c.String(http.StatusTeapot, "custom: "+err.Error())
	})

	r.Handle("GET", "/test", func(_ *Context) error {
		return errors.New("test error")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTeapot {
		t.Errorf("custom handler: want status 418, got %d", w.Code)
	}
	if w.Header().Get("X-Custom-Error") != "true" {
		t.Error("custom handler: X-Custom-Error header not set")
	}
}

// TestErrorHandler_AlreadyWritten_NoDuplicateBody verifies that if a
// handler has already written headers and body, the error handler does
// not append additional content.
func TestErrorHandler_AlreadyWritten_NoDuplicateBody(t *testing.T) {
	r := New()

	r.Handle("GET", "/test", func(c *Context) error {
		_ = c.JSON(http.StatusOK, map[string]bool{"ok": true})
		return errors.New("late error")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if strings.Contains(body, "Internal Server Error") {
		t.Errorf("already written: error handler appended to body: %q", body)
	}
	if !strings.HasPrefix(body, `{"ok":true}`) {
		t.Errorf("already written: body corrupted, want prefix %q got %q", `{"ok":true}`, body)
	}
}

// TestErrorHandler_UnsupportedMediaType_Returns415 verifies that
// sending an unsupported Content-Type produces 415.
func TestErrorHandler_UnsupportedMediaType_Returns415(t *testing.T) {
	r := New()

	r.POST("/users", func(c *Box[errorHandlerReq, errorHandlerRes]) error {
		return c.OK(errorHandlerRes{ID: 1, Message: "created"})
	})

	body := `name=test`
	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("unsupported media type: want status 415, got %d; body: %s", w.Code, w.Body.String())
	}
}

// TestErrorHandler_AfterNoContent_NoDuplicateBody verifies that if a handler
// sends NoContent(204) and then returns an error, the error handler does NOT
// append "Internal Server Error" to the (empty) response body.
func TestErrorHandler_AfterNoContent_NoDuplicateBody(t *testing.T) {
	r := New()

	r.Handle("DELETE", "/items/:id", func(c *Context) error {
		_ = c.NoContent(http.StatusNoContent)
		return errors.New("post-write error")
	})

	req := httptest.NewRequest("DELETE", "/items/42", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("after NoContent: want status 204, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "" {
		t.Errorf("after NoContent: want empty body, got %q", body)
	}
}

// TestErrorHandler_AfterXML_NoDuplicateBody verifies that if a handler
// writes an XML response and then returns an error, the error handler
// does NOT append additional content to the response.
func TestErrorHandler_AfterXML_NoDuplicateBody(t *testing.T) {
	type Item struct {
		ID string `xml:"id"`
	}

	r := New()

	r.Handle("GET", "/items/:id", func(c *Context) error {
		_ = c.XML(http.StatusOK, Item{ID: "42"})
		return errors.New("post-write error")
	})

	req := httptest.NewRequest("GET", "/items/42", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if strings.Contains(body, "Internal Server Error") {
		t.Errorf("after XML: error handler appended to body: %q", body)
	}
	if !strings.Contains(body, "<Item>") || !strings.Contains(body, "<id>42</id>") {
		t.Errorf("after XML: original body corrupted: %q", body)
	}
}

// TestErrorHandler_AfterBlob_NoDuplicateBody verifies that Blob() also
// marks the response as written, preventing error handler from appending.
func TestErrorHandler_AfterBlob_NoDuplicateBody(t *testing.T) {
	r := New()

	r.Handle("GET", "/file", func(c *Context) error {
		_ = c.Blob(http.StatusOK, "application/octet-stream", []byte("binary-data"))
		return errors.New("post-write error")
	})

	req := httptest.NewRequest("GET", "/file", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if body != "binary-data" {
		t.Errorf("after Blob: want body %q, got %q", "binary-data", body)
	}
}

// TestErrorHandler_AfterMarkdown_NoDuplicateBody verifies that Markdown()
// marks the response as written.
func TestErrorHandler_AfterMarkdown_NoDuplicateBody(t *testing.T) {
	r := New()

	r.Handle("GET", "/docs", func(c *Context) error {
		_ = c.Markdown("# Title")
		return errors.New("post-write error")
	})

	req := httptest.NewRequest("GET", "/docs", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if body != "# Title" {
		t.Errorf("after Markdown: want body %q, got %q", "# Title", body)
	}
}

// TestErrorHandler_AfterStream_NoDuplicateBody verifies that Stream()
// marks the response as written.
func TestErrorHandler_AfterStream_NoDuplicateBody(t *testing.T) {
	r := New()

	r.Handle("GET", "/stream", func(c *Context) error {
		_ = c.Stream(http.StatusOK, "text/plain", strings.NewReader("streamed"))
		return errors.New("post-write error")
	})

	req := httptest.NewRequest("GET", "/stream", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if body != "streamed" {
		t.Errorf("after Stream: want body %q, got %q", "streamed", body)
	}
}

// TestResponseWriter_Flusher verifies that the responseWriter wrapper
// implements http.Flusher, which is required for SSE streaming.
func TestResponseWriter_Flusher(t *testing.T) {
	r := New()
	var flusherOK bool

	r.Handle("GET", "/sse", func(c *Context) error {
		if f, ok := c.Response.(http.Flusher); ok {
			flusherOK = true
			f.Flush()
		}
		return c.String(200, "ok")
	})

	req := httptest.NewRequest("GET", "/sse", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !flusherOK {
		t.Error("responseWriter should implement http.Flusher")
	}
}
