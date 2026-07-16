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

// TestRouter_TrailingSlash_Default tests that trailing slash is strict by default.
func TestRouter_TrailingSlash_Default(t *testing.T) {
	r := New()
	r.GET("/users", func(c *Context) error {
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

	r.GET("/users", func(c *Context) error {
		return c.String(200, "users")
	})
	r.GET("/users/:id/posts", func(c *Context) error {
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
		{"unregistered path", "/notfound", 404, "Not Found"},
		{"unregistered with slash", "/notfound/", 404, "Not Found"},
		{"root path", "/", 404, "Not Found"},
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

// TestRouter_StripTrailingSlash_Bidirectional tests adding slash when the
// registered route has a trailing slash.
func TestRouter_StripTrailingSlash_Bidirectional(t *testing.T) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)

	r.GET("/files/", func(c *Context) error {
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

	r.GET("/users", func(c *Context) error {
		return c.String(200, "users")
	})
	r.POST("/users", func(c *Context) error {
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

	r.PUT("/items/:id", func(c *Context) error {
		return c.String(200, "updated")
	})
	r.DELETE("/items/:id", func(c *Context) error {
		return c.String(200, "deleted")
	})
	r.PATCH("/items/:id", func(c *Context) error {
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

	r.GET("/files/", func(c *Context) error {
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

	r.GET("/search", func(c *Context) error {
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

	r.GET("/", func(c *Context) error {
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

	r.GET("/users", func(c *Context) error {
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

	r.GET("/files/*filepath", func(c *Context) error {
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

	r.GET("/api/data", func(c *Context) error {
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

	r.GET("/users", func(c *Context) error {
		return c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/notfound/", http.NoBody)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (not a redirect loop)", w.Code)
	}
}

// BenchmarkRouter_TrailingSlash_Strip benchmarks strip trailing slash overhead.
func BenchmarkRouter_TrailingSlash_Strip(b *testing.B) {
	r := New()
	r.WithTrailingSlash(StripTrailingSlash)
	r.GET("/api/v1/users", func(c *Context) error {
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
	r.GET("/api/v1/users", func(c *Context) error {
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
