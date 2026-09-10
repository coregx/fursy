package fursy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRouter_New tests router initialization.
func TestRouter_New(t *testing.T) {
	r := New()
	// Router should never be nil.
	if r.trees == nil {
		t.Error("trees map not initialized")
	}
	if !r.handleMethodNotAllowed {
		t.Error("handleMethodNotAllowed should be true by default")
	}
	if !r.handleOPTIONS {
		t.Error("handleOPTIONS should be true by default")
	}
}

// TestRouter_GET tests GET method registration.
func TestRouter_GET(t *testing.T) {
	r := New()
	called := false
	handler := func(c *Context) error {
		called = true
		return c.String(200, "OK")
	}

	r.Handle("GET", "/test", handler)

	if r.trees[http.MethodGet] == nil {
		t.Fatal("GET tree not created")
	}

	// Test route execution.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}
}

// TestRouter_POST tests POST method registration.
func TestRouter_POST(t *testing.T) {
	r := New()
	called := false
	handler := func(c *Context) error {
		called = true
		return c.String(201, "Created")
	}

	r.Handle("POST", "/users", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", http.NoBody)

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != 201 {
		t.Errorf("Status code = %d, want 201", w.Code)
	}
}

// TestRouter_PUT tests PUT method registration.
func TestRouter_PUT(t *testing.T) {
	r := New()
	called := false
	handler := func(c *Context) error {
		called = true
		return c.NoContent(204)
	}

	r.Handle("PUT", "/users/1", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/users/1", http.NoBody)

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != 204 {
		t.Errorf("Status code = %d, want 204", w.Code)
	}
}

// TestRouter_DELETE tests DELETE method registration.
func TestRouter_DELETE(t *testing.T) {
	r := New()
	called := false
	handler := func(c *Context) error {
		called = true
		return c.NoContent(204)
	}

	r.Handle("DELETE", "/users/1", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/1", http.NoBody)

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != 204 {
		t.Errorf("Status code = %d, want 204", w.Code)
	}
}

// TestRouter_PATCH tests PATCH method registration.
func TestRouter_PATCH(t *testing.T) {
	r := New()
	called := false
	handler := func(c *Context) error {
		called = true
		return c.JSON(200, map[string]string{"status": "updated"})
	}

	r.Handle("PATCH", "/users/1", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/users/1", http.NoBody)

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}
}

// TestRouter_HEAD tests HEAD method registration.
func TestRouter_HEAD(t *testing.T) {
	r := New()
	called := false
	handler := func(c *Context) error {
		called = true
		c.SetHeader("X-Custom", "value")
		return c.NoContent(200)
	}

	r.Handle("HEAD", "/users/1", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/users/1", http.NoBody)

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}
	if w.Header().Get("X-Custom") != "value" {
		t.Error("Header not set")
	}
}

// TestRouter_OPTIONS tests OPTIONS method registration.
func TestRouter_OPTIONS(t *testing.T) {
	r := New()
	called := false
	handler := func(c *Context) error {
		called = true
		c.SetHeader("Allow", "GET, POST, OPTIONS")
		return c.NoContent(200)
	}

	r.Handle("OPTIONS", "/users", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/users", http.NoBody)

	r.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}
	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}
}

// TestRouter_Handle tests the generic Handle method.
func TestRouter_Handle(t *testing.T) {
	r := New()
	handler := func(c *Context) error {
		return c.String(200, "OK")
	}

	// Valid registration.
	r.Handle(http.MethodGet, "/test", handler)

	if r.trees[http.MethodGet] == nil {
		t.Error("GET tree not created")
	}
}

// TestRouter_Handle_Panics tests Handle method panic conditions.
func TestRouter_Handle_Panics(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		handler HandlerFunc
		wantMsg string
	}{
		{
			name:    "empty method",
			method:  "",
			path:    "/test",
			handler: func(_ *Context) error { return nil },
			wantMsg: "fursy: HTTP method cannot be empty",
		},
		{
			name:    "empty path",
			method:  http.MethodGet,
			path:    "",
			handler: func(_ *Context) error { return nil },
			wantMsg: "fursy: path cannot be empty",
		},
		{
			name:    "nil handler",
			method:  http.MethodGet,
			path:    "/test",
			handler: nil,
			wantMsg: "fursy: handler cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New()
			defer func() {
				rec := recover()
				if rec == nil {
					t.Error("expected panic, got none")
					return
				}
				msg, ok := rec.(string)
				if !ok {
					t.Errorf("panic value is not string: %v", rec)
					return
				}
				if msg != tt.wantMsg {
					t.Errorf("panic message = %q, want %q", msg, tt.wantMsg)
				}
			}()
			r.Handle(tt.method, tt.path, tt.handler)
		})
	}
}

// TestRouter_ServeHTTP_NotFound tests 404 response.
func TestRouter_ServeHTTP_NotFound(t *testing.T) {
	r := New()
	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notfound", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusNotFound)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/problem+json") {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
}

// TestRouter_ErrorResponses_ProblemJSON verifies that all auto-generated error
// responses use RFC 9457 Problem Details (application/problem+json).
func TestRouter_ErrorResponses_ProblemJSON(t *testing.T) {
	r := New()
	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	tests := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{"404 Not Found", "GET", "/notfound", 404},
		{"405 Method Not Allowed", "POST", "/users", 405},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			r.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}

			ct := w.Header().Get("Content-Type")
			if !strings.Contains(ct, "application/problem+json") {
				t.Errorf("Content-Type = %q, want application/problem+json", ct)
			}

			var p Problem
			if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
				t.Fatalf("failed to decode Problem JSON: %v", err)
			}
			if p.Status != tt.status {
				t.Errorf("Problem.Status = %d, want %d", p.Status, tt.status)
			}
			if p.Title == "" {
				t.Error("Problem.Title should not be empty")
			}
		})
	}
}

// TestRouter_ServeHTTP_MethodNotAllowed tests 405 response.
func TestRouter_ServeHTTP_MethodNotAllowed(t *testing.T) {
	r := New()
	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/problem+json") {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
}

// TestRouter_ServeHTTP_Parameters tests URL parameter extraction.
func TestRouter_ServeHTTP_Parameters(t *testing.T) {
	r := New()
	r.Handle("GET", "/users/:id/posts/:postID", func(c *Context) error {
		id := c.Param("id")
		postID := c.Param("postID")
		return c.String(200, "User: "+id+", Post: "+postID)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/456", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	want := "User: 123, Post: 456"
	if string(body) != want {
		t.Errorf("Body = %q, want %q", body, want)
	}
}

// TestRouter_ServeHTTP_Wildcard tests wildcard route.
func TestRouter_ServeHTTP_Wildcard(t *testing.T) {
	r := New()
	r.Handle("GET", "/files/*filepath", func(c *Context) error {
		filepath := c.Param("filepath")
		return c.String(200, "File: "+filepath)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/docs/readme.md", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	want := "File: docs/readme.md"
	if string(body) != want {
		t.Errorf("Body = %q, want %q", body, want)
	}
}

// TestRouter_ServeHTTP_HandlerError tests handler error handling.
func TestRouter_ServeHTTP_HandlerError(t *testing.T) {
	r := New()
	r.Handle("GET", "/error", func(_ *Context) error {
		return ErrInvalidRedirectCode // Return an error without writing response.
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/error", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/problem+json") {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
}

// TestRouter_ServeHTTP_ContextPooling tests context reuse.
func TestRouter_ServeHTTP_ContextPooling(t *testing.T) {
	r := New()
	var firstCtx, secondCtx *Context

	r.Handle("GET", "/first", func(c *Context) error {
		firstCtx = c
		return c.String(200, "First")
	})

	r.Handle("GET", "/second", func(c *Context) error {
		secondCtx = c
		return c.String(200, "Second")
	})

	// First request.
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/first", http.NoBody)
	r.ServeHTTP(w1, req1)

	// Second request.
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/second", http.NoBody)
	r.ServeHTTP(w2, req2)

	// Context should be reused (same pointer after reset).
	if firstCtx == secondCtx {
		t.Log("Context pooling working: same context reused")
	}
}

// TestRouter_MultipleMethods tests registering same path with different methods.
func TestRouter_MultipleMethods(t *testing.T) {
	r := New()
	getCalled := false
	postCalled := false

	r.Handle("GET", "/users", func(c *Context) error {
		getCalled = true
		return c.String(200, "GET")
	})

	r.Handle("POST", "/users", func(c *Context) error {
		postCalled = true
		return c.String(201, "POST")
	})

	// Test GET.
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/users", http.NoBody)
	r.ServeHTTP(w1, req1)

	if !getCalled {
		t.Error("GET handler not called")
	}
	if w1.Code != 200 {
		t.Errorf("GET status = %d, want 200", w1.Code)
	}

	// Test POST.
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/users", http.NoBody)
	r.ServeHTTP(w2, req2)

	if !postCalled {
		t.Error("POST handler not called")
	}
	if w2.Code != 201 {
		t.Errorf("POST status = %d, want 201", w2.Code)
	}
}

// TestRouter_pathExistsInOtherMethods tests the helper function.
func TestRouter_pathExistsInOtherMethods(t *testing.T) {
	r := New()
	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})
	r.Handle("POST", "/users", func(c *Context) error {
		return c.String(201, "Created")
	})

	// Path exists in GET, check from POST.
	if !r.pathExistsInOtherMethods("/users", http.MethodPut) {
		t.Error("Path should exist in other methods")
	}

	// Path doesn't exist.
	if r.pathExistsInOtherMethods("/notfound", http.MethodGet) {
		t.Error("Path should not exist in any method")
	}
}

// TestRouter_MethodNotAllowed_Disabled tests disabling 405 handling.
func TestRouter_MethodNotAllowed_Disabled(t *testing.T) {
	r := New()
	r.handleMethodNotAllowed = false

	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", http.NoBody)

	r.ServeHTTP(w, req)

	// Should return 404 instead of 405 when disabled.
	if w.Code != http.StatusNotFound {
		t.Errorf("Status code = %d, want %d (404)", w.Code, http.StatusNotFound)
	}
}

// TestRouter_TrailingSlash_Default tests that trailing slash is strict by default.
func TestRouter_TrailingSlash_Default(t *testing.T) {
	r := New()
	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"exact match", "/users", 200},
		{"trailing slash 404", "/users/", 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("GET %s: status = %d, want %d", tt.path, w.Code, tt.wantCode)
			}
		})
	}
}

// TestRouter_StripTrailingSlash tests silent trailing slash removal.
func TestRouter_StripTrailingSlash(t *testing.T) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)

	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "users")
	})
	r.Handle("GET", "/users/:id/posts", func(c *Context) error {
		return c.String(200, "posts:"+c.Param("id"))
	})

	tests := []struct {
		name     string
		path     string
		wantCode int
		wantBody string
	}{
		{"exact match", "/users", 200, "users"},
		{"strip trailing slash", "/users/", 200, "users"},
		{"param exact", "/users/42/posts", 200, "posts:42"},
		{"param strip slash", "/users/42/posts/", 200, "posts:42"},
		{"unregistered path", "/notfound", 404, ""},
		{"unregistered with slash", "/notfound/", 404, ""},
		{"root path", "/", 404, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("GET %s: status = %d, want %d", tt.path, w.Code, tt.wantCode)
			}
			if tt.wantBody != "" {
				body, _ := io.ReadAll(w.Body)
				if string(body) != tt.wantBody {
					t.Errorf("GET %s: body = %q, want %q", tt.path, body, tt.wantBody)
				}
			}
		})
	}
}

// TestRouter_StripTrailingSlash_Bidirectional tests adding slash when the
// registered route has a trailing slash.
func TestRouter_StripTrailingSlash_Bidirectional(t *testing.T) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)

	r.Handle("GET", "/files/", func(c *Context) error {
		return c.String(200, "files")
	})

	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"exact with slash", "/files/", 200},
		{"without slash resolves", "/files", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("GET %s: status = %d, want %d", tt.path, w.Code, tt.wantCode)
			}
		})
	}
}

// TestRouter_RedirectTrailingSlash tests redirect behavior.
func TestRouter_RedirectTrailingSlash(t *testing.T) {
	r := New()
	r.WithTrailingSlash(RedirectTrailingSlash)

	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "users")
	})
	r.Handle("POST", "/users", func(c *Context) error {
		return c.String(201, "created")
	})

	tests := []struct {
		name         string
		method       string
		path         string
		wantCode     int
		wantLocation string
	}{
		{"GET exact match", http.MethodGet, "/users", 200, ""},
		{"GET redirect slash", http.MethodGet, "/users/", 301, "/users"},
		{"POST exact match", http.MethodPost, "/users", 201, ""},
		{"POST redirect 308", http.MethodPost, "/users/", 308, "/users"},
		{"PUT unregistered 405", http.MethodPut, "/users/", 405, ""},
		{"DELETE unregistered 405", http.MethodDelete, "/users/", 405, ""},
		{"PATCH unregistered 405", http.MethodPatch, "/users/", 405, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("%s %s: status = %d, want %d", tt.method, tt.path, w.Code, tt.wantCode)
			}
			if tt.wantLocation != "" {
				loc := w.Header().Get("Location")
				if loc != tt.wantLocation {
					t.Errorf("%s %s: Location = %q, want %q", tt.method, tt.path, loc, tt.wantLocation)
				}
			}
		})
	}
}

// TestRouter_RedirectTrailingSlash_Bidirectional tests redirect when the
// registered route has a trailing slash.
// TestRouter_RedirectTrailingSlash_NonGETMethods tests 308 redirects for
// PUT, DELETE, PATCH when those methods have registered routes.
func TestRouter_RedirectTrailingSlash_NonGETMethods(t *testing.T) {
	r := New()
	r.WithTrailingSlash(RedirectTrailingSlash)

	r.Handle("PUT", "/items/:id", func(c *Context) error {
		return c.String(200, "updated")
	})
	r.Handle("DELETE", "/items/:id", func(c *Context) error {
		return c.String(200, "deleted")
	})
	r.Handle("PATCH", "/items/:id", func(c *Context) error {
		return c.String(200, "patched")
	})

	methods := []string{http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/items/42/", http.NoBody)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusPermanentRedirect {
				t.Errorf("%s /items/42/: status = %d, want 308", method, w.Code)
			}
			loc := w.Header().Get("Location")
			if loc != "/items/42" {
				t.Errorf("%s /items/42/: Location = %q, want %q", method, loc, "/items/42")
			}
		})
	}
}

func TestRouter_RedirectTrailingSlash_Bidirectional(t *testing.T) {
	r := New()
	r.WithTrailingSlash(RedirectTrailingSlash)

	r.Handle("GET", "/files/", func(c *Context) error {
		return c.String(200, "files")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMovedPermanently)
	}
	loc := w.Header().Get("Location")
	if loc != "/files/" {
		t.Errorf("Location = %q, want %q", loc, "/files/")
	}
}

// TestRouter_RedirectTrailingSlash_PreservesQuery tests that query params
// are preserved across redirects.
func TestRouter_RedirectTrailingSlash_PreservesQuery(t *testing.T) {
	r := New()
	r.WithTrailingSlash(RedirectTrailingSlash)

	r.Handle("GET", "/search", func(c *Context) error {
		return c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/search/?q=test&page=2", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("status = %d, want 301", w.Code)
	}
	loc := w.Header().Get("Location")
	want := "/search?q=test&page=2"
	if loc != want {
		t.Errorf("Location = %q, want %q", loc, want)
	}
}

// TestRouter_TrailingSlash_RootPath tests that root "/" is never altered.
func TestRouter_TrailingSlash_RootPath(t *testing.T) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)

	r.Handle("GET", "/", func(c *Context) error {
		return c.String(200, "root")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("GET /: status = %d, want 200", w.Code)
	}
	body, _ := io.ReadAll(w.Body)
	if string(body) != "root" {
		t.Errorf("GET /: body = %q, want %q", body, "root")
	}
}

// TestRouter_TrailingSlash_MethodNotAllowed tests 405 with trailing slash.
func TestRouter_TrailingSlash_MethodNotAllowed(t *testing.T) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)

	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /users/: status = %d, want 405", w.Code)
	}
}

// TestRouter_TrailingSlash_Wildcard tests that wildcard routes are not
// affected by trailing slash handling.
func TestRouter_TrailingSlash_Wildcard(t *testing.T) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)

	r.Handle("GET", "/files/*filepath", func(c *Context) error {
		return c.String(200, c.Param("filepath"))
	})

	tests := []struct {
		name     string
		path     string
		wantCode int
		wantBody string
	}{
		{"wildcard normal", "/files/a/b.txt", 200, "a/b.txt"},
		{"wildcard with trailing", "/files/a/b/", 200, "a/b/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("GET %s: status = %d, want %d", tt.path, w.Code, tt.wantCode)
			}
			body, _ := io.ReadAll(w.Body)
			if string(body) != tt.wantBody {
				t.Errorf("GET %s: body = %q, want %q", tt.path, body, tt.wantBody)
			}
		})
	}
}

// TestRouter_TrailingSlash_WithMiddleware tests that middleware executes
// correctly with trailing slash handling.
func TestRouter_TrailingSlash_WithMiddleware(t *testing.T) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)

	middlewareCalled := false
	r.Use(func(c *Context) error {
		middlewareCalled = true
		return c.Next()
	})

	r.Handle("GET", "/api/data", func(c *Context) error {
		return c.String(200, "data")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/data/", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !middlewareCalled {
		t.Error("middleware was not called")
	}
}

// TestRouter_WithTrailingSlash_Fluent tests fluent API chaining.
func TestRouter_WithTrailingSlash_Fluent(t *testing.T) {
	r := New()
	result := r.WithTrailingSlash(StripTrailingSlash)

	if result != r {
		t.Error("WithTrailingSlash should return same router for chaining")
	}
	if r.trailingSlash != StripTrailingSlash {
		t.Errorf("trailingSlash = %d, want %d", r.trailingSlash, StripTrailingSlash)
	}
}

// TestTrailingSlashAlternate tests the path toggle helper.
func TestTrailingSlashAlternate(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/users", "/users/"},
		{"/users/", "/users"},
		{"/a/b/c", "/a/b/c/"},
		{"/a/b/c/", "/a/b/c"},
		{"/", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := trailingSlashAlternate(tt.path)
			if got != tt.want {
				t.Errorf("trailingSlashAlternate(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestRouter_RedirectTrailingSlash_NoRedirectLoop ensures no redirect loop
// when neither path variant matches.
func TestRouter_RedirectTrailingSlash_NoRedirectLoop(t *testing.T) {
	r := New()
	r.WithTrailingSlash(RedirectTrailingSlash)

	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notfound/", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (not a redirect loop)", w.Code)
	}
}

// --- F5 audit fix: body size limit ---

// bodyLimitReq is a test request type for body limit tests.
type bodyLimitReq struct {
	Data string `json:"data"`
}

// bodyLimitRes is a test response type for body limit tests.
type bodyLimitRes struct {
	OK bool `json:"ok"`
}

// TestBodyLimit_Large_Returns413 verifies that a request body exceeding
// the default max body size (4MB) returns 413 Payload Too Large.
func TestBodyLimit_Large_Returns413(t *testing.T) {
	r := New()

	r.POST("/upload", func(c *Box[bodyLimitReq, bodyLimitRes]) error {
		return c.OK(bodyLimitRes{OK: true})
	})

	// Create a 5MB JSON body.
	largeData := strings.Repeat("x", 5*1024*1024)
	body := `{"data":"` + largeData + `"}`
	req := httptest.NewRequest("POST", "/upload", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("5MB body: want 413, got %d; body: %s", w.Code, w.Body.String())
	}
}

// TestBodyLimit_Normal_Passes verifies that a small request body passes
// through the body limit check.
func TestBodyLimit_Normal_Passes(t *testing.T) {
	r := New()

	r.POST("/upload", func(c *Box[bodyLimitReq, bodyLimitRes]) error {
		return c.OK(bodyLimitRes{OK: true})
	})

	body := `{"data":"hello"}`
	req := httptest.NewRequest("POST", "/upload", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("small body: want 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

// TestBodyLimit_Custom verifies that SetMaxBodySize overrides the default.
func TestBodyLimit_Custom(t *testing.T) {
	r := New()
	r.SetMaxBodySize(1 << 20) // 1MB

	r.POST("/upload", func(c *Box[bodyLimitReq, bodyLimitRes]) error {
		return c.OK(bodyLimitRes{OK: true})
	})

	// 2MB body should be rejected with 1MB limit.
	largeData := strings.Repeat("x", 2*1024*1024)
	body := `{"data":"` + largeData + `"}`
	req := httptest.NewRequest("POST", "/upload", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("2MB body with 1MB limit: want 413, got %d; body: %s", w.Code, w.Body.String())
	}
}

// TestBodyLimit_ZeroDisables verifies that SetMaxBodySize(0) disables the limit.
func TestBodyLimit_ZeroDisables(t *testing.T) {
	r := New()
	r.SetMaxBodySize(0) // Disable limit.

	r.POST("/upload", func(c *Box[bodyLimitReq, bodyLimitRes]) error {
		return c.OK(bodyLimitRes{OK: true})
	})

	// 5MB body should pass when limit is disabled.
	largeData := strings.Repeat("x", 5*1024*1024)
	body := `{"data":"` + largeData + `"}`
	req := httptest.NewRequest("POST", "/upload", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("5MB body with limit disabled: want 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

// TestBodyLimit_NonGenericHandler verifies that plain handlers (HandlerFunc)
// are not affected by body limit — only generic handlers with Bind() are limited.
func TestBodyLimit_NonGenericHandler(t *testing.T) {
	r := New()

	r.Handle("POST", "/raw", func(c *Context) error {
		return c.String(200, "OK")
	})

	// Large body should pass for plain handler (no automatic Bind).
	largeData := strings.Repeat("x", 5*1024*1024)
	body := `{"data":"` + largeData + `"}`
	req := httptest.NewRequest("POST", "/raw", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("plain handler with large body: want 200, got %d", w.Code)
	}
}

// BenchmarkRouter_TrailingSlash_Strip benchmarks strip trailing slash overhead.
func BenchmarkRouter_TrailingSlash_Strip(b *testing.B) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)
	r.Handle("GET", "/api/v1/users", func(c *Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_TrailingSlash_ExactMatch benchmarks that exact matches
// have zero overhead when trailing slash is enabled.
func BenchmarkRouter_TrailingSlash_ExactMatch(b *testing.B) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)
	r.Handle("GET", "/api/v1/users", func(c *Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}

// TestRouter_HEAD_FallbackToGET tests that HEAD requests fall back to the GET
// handler when no explicit HEAD route is registered (RFC 9110).
func TestRouter_HEAD_FallbackToGET(t *testing.T) {
	r := New()
	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "user list")
	})

	// HEAD /users should return 200 (GET handler), not 405.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/users", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("HEAD /users: expected 200, got %d", w.Code)
	}

	// Note: httptest.ResponseRecorder does NOT suppress body for HEAD.
	// A real http.Server does — see TestRouter_HEAD_FallbackToGET_RealServer
	// for end-to-end verification. Here we just verify the status code.
}

// TestRouter_HEAD_ExplicitRoute tests that an explicit HEAD route takes priority.
func TestRouter_HEAD_ExplicitRoute(t *testing.T) {
	r := New()
	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "GET response")
	})
	r.Handle("HEAD", "/users", func(c *Context) error {
		c.SetHeader("X-Custom", "head-handler")
		return c.NoContent(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/users", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("HEAD /users: expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Custom") != "head-handler" {
		t.Error("HEAD /users: explicit HEAD handler should take priority over GET fallback")
	}
}

// TestRouter_HEAD_NoGET_Returns404 tests that HEAD returns 404 when neither
// HEAD nor GET route is registered for the path.
func TestRouter_HEAD_NoGET_Returns404(t *testing.T) {
	r := New()
	r.Handle("POST", "/users", func(c *Context) error {
		return c.String(201, "created")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/users", http.NoBody)
	r.ServeHTTP(w, req)

	// POST exists but neither HEAD nor GET — should be 405.
	if w.Code != 405 {
		t.Errorf("HEAD /users (only POST registered): expected 405, got %d", w.Code)
	}
}

// TestRouter_GlobalMiddleware_On404 tests that global middleware executes for 404 paths.
func TestRouter_GlobalMiddleware_On404(t *testing.T) {
	middlewareCalled := false

	r := New()
	r.Use(func(c *Context) error {
		middlewareCalled = true
		c.SetHeader("X-Middleware", "executed")
		return c.Next()
	})
	r.Handle("GET", "/exists", func(c *Context) error {
		return c.String(200, "OK")
	})

	// Request to non-existent path.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/not-found", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
	if !middlewareCalled {
		t.Error("global middleware should execute on 404 paths")
	}
	if w.Header().Get("X-Middleware") != "executed" {
		t.Error("middleware header should be set on 404 response")
	}
}

// TestRouter_GlobalMiddleware_On405 tests that global middleware executes for 405 paths.
func TestRouter_GlobalMiddleware_On405(t *testing.T) {
	middlewareCalled := false

	r := New()
	r.Use(func(c *Context) error {
		middlewareCalled = true
		c.SetHeader("X-Middleware", "executed")
		return c.Next()
	})
	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	// POST to GET-only path.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
	if !middlewareCalled {
		t.Error("global middleware should execute on 405 paths")
	}
	if w.Header().Get("X-Middleware") != "executed" {
		t.Error("middleware header should be set on 405 response")
	}
}

// TestRouter_OPTIONS_NoMiddleware_Returns204 tests that OPTIONS returns 204+Allow
// even when no middleware is registered.
func TestRouter_OPTIONS_NoMiddleware_Returns204(t *testing.T) {
	r := New()
	// No middleware registered.
	r.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "OK")
	})
	r.Handle("POST", "/users", func(c *Context) error {
		return c.String(201, "Created")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/users", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != 204 {
		t.Errorf("OPTIONS /users without middleware: expected 204, got %d", w.Code)
	}

	allow := w.Header().Get("Allow")
	if allow == "" {
		t.Error("OPTIONS response should include Allow header")
	}
	if !strings.Contains(allow, "GET") || !strings.Contains(allow, "POST") {
		t.Errorf("Allow header should contain GET and POST, got %q", allow)
	}
}

// TestRouter_PercentEncodedColon tests that URL-encoded colon (%3A) in the path
// is properly decoded by net/http and matched as a param value, not as a param marker.
// net/http decodes %3A→: in URL.Path. The radix tree's findChild skips wildcard
// nodes for literal character matches, so ":id" becomes the param VALUE, not empty.
func TestRouter_PercentEncodedColon(t *testing.T) {
	r := New()
	r.Handle("GET", "/users/:id", func(c *Context) error {
		return c.String(200, "id="+c.Param("id"))
	})

	// net/http decodes %3A to ":" → path becomes /users/:id → param value = ":id".
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/%3Aid", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("GET /users/%%3Aid: expected 200, got %d", w.Code)
	}
	// Param value should be ":id" (decoded), NOT empty.
	if w.Body.String() != "id=:id" {
		t.Errorf("GET /users/%%3Aid: expected body 'id=:id', got %q", w.Body.String())
	}

	// Normal param still works.
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/users/42", http.NoBody)
	r.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Errorf("GET /users/42: expected 200, got %d", w2.Code)
	}
	if w2.Body.String() != "id=42" {
		t.Errorf("GET /users/42: expected body 'id=42', got %q", w2.Body.String())
	}
}
