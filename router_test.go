package fursy

import (
	"io"
	"net/http"
	"net/http/httptest"
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

	r.GET("/test", handler)

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

	r.POST("/users", handler)

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

	r.PUT("/users/1", handler)

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

	r.DELETE("/users/1", handler)

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

	r.PATCH("/users/1", handler)

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

	r.HEAD("/users/1", handler)

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

	r.OPTIONS("/users", handler)

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
	r.GET("/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notfound", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusNotFound)
	}
	body, _ := io.ReadAll(w.Body)
	if string(body) != "Not Found" {
		t.Errorf("Body = %q, want %q", body, "Not Found")
	}
}

// TestRouter_ServeHTTP_MethodNotAllowed tests 405 response.
func TestRouter_ServeHTTP_MethodNotAllowed(t *testing.T) {
	r := New()
	r.GET("/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
	body, _ := io.ReadAll(w.Body)
	if string(body) != "Method Not Allowed" {
		t.Errorf("Body = %q, want %q", body, "Method Not Allowed")
	}
}

// TestRouter_ServeHTTP_Parameters tests URL parameter extraction.
func TestRouter_ServeHTTP_Parameters(t *testing.T) {
	r := New()
	r.GET("/users/:id/posts/:postID", func(c *Context) error {
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
	r.GET("/files/*filepath", func(c *Context) error {
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
	r.GET("/error", func(_ *Context) error {
		return ErrInvalidRedirectCode // Return an error without writing response.
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/error", http.NoBody)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %d, want %d", w.Code, http.StatusInternalServerError)
	}
	body, _ := io.ReadAll(w.Body)
	if string(body) != "Internal Server Error" {
		t.Errorf("Body = %q, want %q", body, "Internal Server Error")
	}
}

// TestRouter_ServeHTTP_ContextPooling tests context reuse.
func TestRouter_ServeHTTP_ContextPooling(t *testing.T) {
	r := New()
	var firstCtx, secondCtx *Context

	r.GET("/first", func(c *Context) error {
		firstCtx = c
		return c.String(200, "First")
	})

	r.GET("/second", func(c *Context) error {
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

	r.GET("/users", func(c *Context) error {
		getCalled = true
		return c.String(200, "GET")
	})

	r.POST("/users", func(c *Context) error {
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
	r.GET("/users", func(c *Context) error {
		return c.String(200, "OK")
	})
	r.POST("/users", func(c *Context) error {
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

	r.GET("/users", func(c *Context) error {
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
