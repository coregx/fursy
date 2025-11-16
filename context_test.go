package fursy

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestContext_Param tests URL parameter extraction.
func TestContext_Param(t *testing.T) {
	c := newContext()
	c.params = []Param{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "john"},
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing param id", "id", "123"},
		{"existing param name", "name", "john"},
		{"non-existing param", "unknown", ""},
		{"empty key", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.Param(tt.key)
			if got != tt.want {
				t.Errorf("Param(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// TestContext_Query tests query parameter extraction.
func TestContext_Query(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=2&limit=10&tags=go&tags=web", http.NoBody)
	c := newContext()
	c.Request = req

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing param page", "page", "2"},
		{"existing param limit", "limit", "10"},
		{"multiple values (first)", "tags", "go"},
		{"non-existing param", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.Query(tt.key)
			if got != tt.want {
				t.Errorf("Query(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// TestContext_QueryDefault tests query parameter with default value.
func TestContext_QueryDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=2", http.NoBody)
	c := newContext()
	c.Request = req

	tests := []struct {
		name         string
		key          string
		defaultValue string
		want         string
	}{
		{"existing param", "page", "1", "2"},
		{"non-existing param", "limit", "10", "10"},
		{"empty param value", "empty", "default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.QueryDefault(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("QueryDefault(%q, %q) = %q, want %q", tt.key, tt.defaultValue, got, tt.want)
			}
		})
	}
}

// TestContext_QueryValues tests multiple query parameter values.
func TestContext_QueryValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?tags=go&tags=web&tags=api", http.NoBody)
	c := newContext()
	c.Request = req

	tags := c.QueryValues("tags")
	want := []string{"go", "web", "api"}

	if len(tags) != len(want) {
		t.Fatalf("QueryValues(tags) returned %d values, want %d", len(tags), len(want))
	}

	for i, tag := range tags {
		if tag != want[i] {
			t.Errorf("QueryValues(tags)[%d] = %q, want %q", i, tag, want[i])
		}
	}

	// Test non-existing parameter.
	nonExisting := c.QueryValues("unknown")
	if nonExisting != nil {
		t.Errorf("QueryValues(unknown) = %v, want nil", nonExisting)
	}
}

// TestContext_Form tests form parameter extraction.
func TestContext_Form(t *testing.T) {
	form := url.Values{}
	form.Add("username", "john")
	form.Add("password", "secret")

	req := httptest.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := newContext()
	c.Request = req

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing username", "username", "john"},
		{"existing password", "password", "secret"},
		{"non-existing", "email", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.Form(tt.key)
			if got != tt.want {
				t.Errorf("Form(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// TestContext_FormDefault tests form parameter with default value.
func TestContext_FormDefault(t *testing.T) {
	form := url.Values{}
	form.Add("username", "john")

	req := httptest.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := newContext()
	c.Request = req

	tests := []struct {
		name         string
		key          string
		defaultValue string
		want         string
	}{
		{"existing param", "username", "guest", "john"},
		{"non-existing param", "role", "user", "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.FormDefault(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("FormDefault(%q, %q) = %q, want %q", tt.key, tt.defaultValue, got, tt.want)
			}
		})
	}
}

// TestContext_PostForm tests POST body form extraction.
func TestContext_PostForm(t *testing.T) {
	form := url.Values{}
	form.Add("name", "john")

	req := httptest.NewRequest("POST", "/update?id=123", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	c := newContext()
	c.Request = req

	// PostForm should get value from body.
	name := c.PostForm("name")
	if name != "john" {
		t.Errorf("PostForm(name) = %q, want %q", name, "john")
	}

	// PostForm should NOT get value from query string.
	id := c.PostForm("id")
	if id != "" {
		t.Errorf("PostForm(id) = %q, want empty (query param should be ignored)", id)
	}
}

// TestContext_String tests plain text response.
func TestContext_String(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	err := c.String(200, "Hello, World!")
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "text/plain; charset=utf-8")
	}

	body := w.Body.String()
	if body != "Hello, World!" {
		t.Errorf("Body = %q, want %q", body, "Hello, World!")
	}
}

// TestContext_JSON tests JSON response.
func TestContext_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	data := map[string]string{"message": "success", "id": "123"}
	err := c.JSON(200, data)
	if err != nil {
		t.Fatalf("JSON() error = %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json; charset=utf-8")
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["message"] != "success" {
		t.Errorf("message = %q, want %q", result["message"], "success")
	}
	if result["id"] != "123" {
		t.Errorf("id = %q, want %q", result["id"], "123")
	}
}

// TestContext_JSONIndent tests indented JSON response.
func TestContext_JSONIndent(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	data := map[string]string{"message": "success"}
	err := c.JSONIndent(200, data, "  ")
	if err != nil {
		t.Fatalf("JSONIndent() error = %v", err)
	}

	body := w.Body.String()
	// Indented JSON should contain newlines.
	if !strings.Contains(body, "\n") {
		t.Errorf("JSONIndent() body doesn't contain newlines: %q", body)
	}
}

// TestContext_XML tests XML response.
func TestContext_XML(t *testing.T) {
	type User struct {
		XMLName xml.Name `xml:"user"`
		ID      string   `xml:"id"`
		Name    string   `xml:"name"`
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	user := User{ID: "123", Name: "John"}
	err := c.XML(200, user)
	if err != nil {
		t.Fatalf("XML() error = %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/xml; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/xml; charset=utf-8")
	}

	body := w.Body.String()
	if !strings.Contains(body, "<user>") {
		t.Errorf("XML body doesn't contain <user> tag: %q", body)
	}
}

// TestContext_NoContent tests no content response.
func TestContext_NoContent(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	err := c.NoContent(204)
	if err != nil {
		t.Fatalf("NoContent() error = %v", err)
	}

	if w.Code != 204 {
		t.Errorf("Status code = %d, want 204", w.Code)
	}

	if w.Body.Len() != 0 {
		t.Errorf("Body length = %d, want 0", w.Body.Len())
	}
}

// TestContext_Redirect tests HTTP redirect.
func TestContext_Redirect(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		url     string
		wantErr bool
	}{
		{"valid 302", 302, "/login", false},
		{"valid 301", 301, "/new-location", false},
		{"valid 307", 307, "/temp", false},
		{"invalid code 200", 200, "/login", true},
		{"invalid code 404", 404, "/login", true},
		{"invalid code 500", 500, "/login", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/old", http.NoBody)

			c := newContext()
			c.Response = w
			c.Request = req

			err := c.Redirect(tt.code, tt.url)

			if tt.wantErr {
				assertRedirectError(t, err, tt.code, tt.url)
			} else {
				assertRedirectSuccess(t, err, w, tt.code, tt.url)
			}
		})
	}
}

// assertRedirectError verifies redirect error cases.
func assertRedirectError(t *testing.T, err error, code int, urlPath string) {
	t.Helper()
	if err == nil {
		t.Errorf("Redirect(%d, %q) expected error, got nil", code, urlPath)
		return
	}
	if !errors.Is(err, ErrInvalidRedirectCode) {
		t.Errorf("Redirect() error = %v, want ErrInvalidRedirectCode", err)
	}
}

// assertRedirectSuccess verifies successful redirect.
func assertRedirectSuccess(t *testing.T, err error, w *httptest.ResponseRecorder, wantCode int, wantURL string) {
	t.Helper()
	if err != nil {
		t.Errorf("Redirect(%d, %q) unexpected error: %v", wantCode, wantURL, err)
		return
	}
	if w.Code != wantCode {
		t.Errorf("Status code = %d, want %d", w.Code, wantCode)
	}
	location := w.Header().Get("Location")
	if location != wantURL {
		t.Errorf("Location header = %q, want %q", location, wantURL)
	}
}

// TestContext_Blob tests binary response.
func TestContext_Blob(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	data := []byte{0x89, 0x50, 0x4E, 0x47} // PNG signature
	err := c.Blob(200, "image/png", data)
	if err != nil {
		t.Fatalf("Blob() error = %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "image/png" {
		t.Errorf("Content-Type = %q, want %q", contentType, "image/png")
	}

	body := w.Body.Bytes()
	if !bytes.Equal(body, data) {
		t.Errorf("Body = %v, want %v", body, data)
	}
}

// TestContext_Stream tests streaming response.
func TestContext_Stream(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	data := strings.NewReader("streaming content")
	err := c.Stream(200, "text/plain", data)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	body := w.Body.String()
	if body != "streaming content" {
		t.Errorf("Body = %q, want %q", body, "streaming content")
	}
}

// TestContext_SetHeader tests setting response headers.
func TestContext_SetHeader(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	c.SetHeader("X-Request-ID", "12345")
	c.SetHeader("X-Custom-Header", "value")

	if got := w.Header().Get("X-Request-ID"); got != "12345" {
		t.Errorf("X-Request-ID = %q, want %q", got, "12345")
	}
	if got := w.Header().Get("X-Custom-Header"); got != "value" {
		t.Errorf("X-Custom-Header = %q, want %q", got, "value")
	}
}

// TestContext_GetHeader tests getting request headers.
func TestContext_GetHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", http.NoBody)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Custom", "custom-value")

	c := newContext()
	c.Request = req

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing User-Agent", "User-Agent", "test-agent"},
		{"existing custom header", "X-Custom", "custom-value"},
		{"non-existing header", "X-Unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.GetHeader(tt.key)
			if got != tt.want {
				t.Errorf("GetHeader(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// TestContext_GetSet tests data storage.
func TestContext_GetSet(t *testing.T) {
	c := newContext()

	// Test Set and Get.
	c.Set("userID", "123")
	c.Set("authenticated", true)
	c.Set("count", 42)

	if got := c.Get("userID"); got != "123" {
		t.Errorf("Get(userID) = %v, want %q", got, "123")
	}
	if got := c.Get("authenticated"); got != true {
		t.Errorf("Get(authenticated) = %v, want true", got)
	}
	if got := c.Get("count"); got != 42 {
		t.Errorf("Get(count) = %v, want 42", got)
	}

	// Test non-existing key.
	if got := c.Get("unknown"); got != nil {
		t.Errorf("Get(unknown) = %v, want nil", got)
	}
}

// TestContext_GetString tests typed string retrieval.
func TestContext_GetString(t *testing.T) {
	c := newContext()
	c.Set("userID", "123")
	c.Set("count", 42) // wrong type

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing string", "userID", "123"},
		{"wrong type", "count", ""},
		{"non-existing", "unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.GetString(tt.key)
			if got != tt.want {
				t.Errorf("GetString(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// TestContext_GetInt tests typed int retrieval.
func TestContext_GetInt(t *testing.T) {
	c := newContext()
	c.Set("count", 42)
	c.Set("userID", "123") // wrong type

	tests := []struct {
		name string
		key  string
		want int
	}{
		{"existing int", "count", 42},
		{"wrong type", "userID", 0},
		{"non-existing", "unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.GetInt(tt.key)
			if got != tt.want {
				t.Errorf("GetInt(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

// TestContext_GetBool tests typed bool retrieval.
func TestContext_GetBool(t *testing.T) {
	c := newContext()
	c.Set("authenticated", true)
	c.Set("disabled", false)
	c.Set("count", 42) // wrong type

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"existing true", "authenticated", true},
		{"existing false", "disabled", false},
		{"wrong type", "count", false},
		{"non-existing", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.GetBool(tt.key)
			if got != tt.want {
				t.Errorf("GetBool(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// TestContext_Reset tests context reset for pooling.
func TestContext_Reset(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)
	router := New()

	c := newContext()
	c.init(w, req, router, []Param{{Key: "id", Value: "123"}})
	c.Set("userID", "456")
	c.query = make(map[string][]string) // simulate lazy load

	// Reset should clear everything.
	c.reset()

	if c.Request != nil {
		t.Error("Request should be nil after reset")
	}
	if c.Response != nil {
		t.Error("Response should be nil after reset")
	}
	if c.router != nil {
		t.Error("router should be nil after reset")
	}
	// params slice is reused (capacity preserved), check length only.
	if len(c.params) != 0 {
		t.Errorf("params should be empty after reset, got len=%d", len(c.params))
	}
	// handlers slice is reused (capacity preserved), check length only.
	if len(c.handlers) != 0 {
		t.Errorf("handlers should be empty after reset, got len=%d", len(c.handlers))
	}
	if c.query != nil {
		t.Error("query should be nil after reset")
	}
	if len(c.data) != 0 {
		t.Errorf("data map should be empty after reset, got %d items", len(c.data))
	}
}

// TestContext_Router tests Router accessor.
func TestContext_Router(t *testing.T) {
	router := New()
	c := newContext()
	c.router = router

	if got := c.Router(); got != router {
		t.Error("Router() returned wrong router instance")
	}
}

// TestRouter_ContextIntegration tests Box integration with Router.
func TestRouter_ContextIntegration(t *testing.T) {
	router := New()

	// Register a handler that uses Box methods.
	router.GET("/users/:id", func(c *Context) error {
		id := c.Param("id")
		return c.JSON(200, map[string]string{"id": id, "name": "User " + id})
	})

	req := httptest.NewRequest("GET", "/users/123", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["id"] != "123" {
		t.Errorf("id = %q, want %q", result["id"], "123")
	}
	if result["name"] != "User 123" {
		t.Errorf("name = %q, want %q", result["name"], "User 123")
	}
}

// TestRouter_ContextPooling tests that Context pooling works correctly.
func TestRouter_ContextPooling(t *testing.T) {
	router := New()

	callCount := 0
	router.GET("/test", func(c *Context) error {
		callCount++
		c.Set("call", callCount)
		return c.String(200, "OK")
	})

	// Make multiple requests to verify pooling.
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Request %d: Status code = %d, want 200", i, w.Code)
		}
	}

	if callCount != 10 {
		t.Errorf("Handler called %d times, want 10", callCount)
	}
}

// TestRouter_ContextQuery tests query parameter extraction through Box.
func TestRouter_ContextQuery(t *testing.T) {
	router := New()

	router.GET("/search", func(c *Context) error {
		q := c.Query("q")
		page := c.QueryDefault("page", "1")
		return c.JSON(200, map[string]string{"query": q, "page": page})
	})

	req := httptest.NewRequest("GET", "/search?q=golang", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)

	if result["query"] != "golang" {
		t.Errorf("query = %q, want %q", result["query"], "golang")
	}
	if result["page"] != "1" {
		t.Errorf("page = %q, want %q (default)", result["page"], "1")
	}
}

// TestRouter_ContextErrorHandling tests error handling in handlers.
func TestRouter_ContextErrorHandling(t *testing.T) {
	router := New()

	// Handler that returns an error.
	router.GET("/error", func(_ *Context) error {
		return io.ErrUnexpectedEOF
	})

	req := httptest.NewRequest("GET", "/error", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 500 Internal Server Error.
	if w.Code != 500 {
		t.Errorf("Status code = %d, want 500", w.Code)
	}

	body := w.Body.String()
	if body != "Internal Server Error" {
		t.Errorf("Body = %q, want %q", body, "Internal Server Error")
	}
}

// TestContext_DataStorageMiddleware simulates middleware data passing.
func TestContext_DataStorageMiddleware(t *testing.T) {
	router := New()

	router.GET("/protected", func(c *Context) error {
		// Simulate middleware setting user data.
		c.Set("userID", "123")
		c.Set("authenticated", true)

		// Handler reads middleware data.
		userID := c.GetString("userID")
		authenticated := c.GetBool("authenticated")

		return c.JSON(200, map[string]any{
			"userID":        userID,
			"authenticated": authenticated,
		})
	})

	req := httptest.NewRequest("GET", "/protected", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	var result map[string]any
	json.Unmarshal(w.Body.Bytes(), &result)

	if result["userID"] != "123" {
		t.Errorf("userID = %v, want %q", result["userID"], "123")
	}
	if result["authenticated"] != true {
		t.Errorf("authenticated = %v, want true", result["authenticated"])
	}
}

// TestContext_OK tests the OK convenience method.
func TestContext_OK(t *testing.T) {
	router := New()
	router.GET("/users", func(c *Context) error {
		return c.OK(map[string]string{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/users", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	// Check body
	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["status"] != "success" {
		t.Errorf("status = %q, want %q", result["status"], "success")
	}
}

// TestContext_Created tests the Created convenience method.
func TestContext_Created(t *testing.T) {
	router := New()
	router.POST("/users", func(c *Context) error {
		return c.Created(map[string]any{"id": 123, "name": "John"})
	})

	req := httptest.NewRequest("POST", "/users", strings.NewReader(`{"name":"John"}`))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check status code - should be 201 Created
	if w.Code != 201 {
		t.Errorf("Status code = %d, want 201", w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	// Check body
	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if int(result["id"].(float64)) != 123 {
		t.Errorf("id = %v, want 123", result["id"])
	}
}

// TestContext_Accepted tests the Accepted convenience method.
func TestContext_Accepted(t *testing.T) {
	router := New()
	router.POST("/jobs", func(c *Context) error {
		return c.Accepted(map[string]string{"jobId": "abc123", "status": "pending"})
	})

	req := httptest.NewRequest("POST", "/jobs", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check status code - should be 202 Accepted
	if w.Code != 202 {
		t.Errorf("Status code = %d, want 202", w.Code)
	}

	// Check body
	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if result["jobId"] != "abc123" {
		t.Errorf("jobId = %q, want %q", result["jobId"], "abc123")
	}
}

// TestContext_NoContentSuccess tests the NoContentSuccess convenience method.
func TestContext_NoContentSuccess(t *testing.T) {
	router := New()
	router.DELETE("/users/:id", func(c *Context) error {
		// Simulate deletion
		return c.NoContentSuccess()
	})

	req := httptest.NewRequest("DELETE", "/users/123", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check status code - should be 204 No Content
	if w.Code != 204 {
		t.Errorf("Status code = %d, want 204", w.Code)
	}

	// Check body is empty
	if w.Body.Len() != 0 {
		t.Errorf("Body length = %d, want 0 (empty body for 204)", w.Body.Len())
	}
}

// TestContext_Text tests the Text convenience method.
func TestContext_Text(t *testing.T) {
	router := New()
	router.GET("/ping", func(c *Context) error {
		return c.Text("pong")
	})

	req := httptest.NewRequest("GET", "/ping", http.NoBody)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check status code - should be 200 OK
	if w.Code != 200 {
		t.Errorf("Status code = %d, want 200", w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", contentType)
	}

	// Check body
	if w.Body.String() != "pong" {
		t.Errorf("Body = %q, want %q", w.Body.String(), "pong")
	}
}

// TestContext_ConvenienceMethods_RESTWorkflow tests complete REST workflow with convenience methods.
func TestContext_ConvenienceMethods_RESTWorkflow(t *testing.T) {
	router := New()

	// Simulate in-memory storage
	users := make(map[string]map[string]any)
	nextID := 1

	// GET - list all users (200 OK)
	router.GET("/users", func(c *Context) error {
		userList := make([]map[string]any, 0, len(users))
		for _, user := range users {
			userList = append(userList, user)
		}
		return c.OK(userList)
	})

	// POST - create user (201 Created)
	router.POST("/users", func(c *Context) error {
		id := string(rune(nextID + '0'))
		user := map[string]any{"id": id, "name": "User" + id}
		users[id] = user
		nextID++
		return c.Created(user)
	})

	// DELETE - delete user (204 No Content)
	router.DELETE("/users/:id", func(c *Context) error {
		id := c.Param("id")
		delete(users, id)
		return c.NoContentSuccess()
	})

	// Test GET (empty list initially)
	req := httptest.NewRequest("GET", "/users", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("GET status = %d, want 200", w.Code)
	}

	// Test POST (create user)
	req = httptest.NewRequest("POST", "/users", http.NoBody)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 201 {
		t.Errorf("POST status = %d, want 201", w.Code)
	}

	// Test DELETE (remove user)
	req = httptest.NewRequest("DELETE", "/users/1", http.NoBody)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Errorf("DELETE status = %d, want 204", w.Code)
	}
}

// TestContext_Accepts tests Accept header checking.
func TestContext_Accepts(t *testing.T) {
	tests := []struct {
		name      string
		accept    string
		mediaType string
		want      bool
	}{
		{
			name:      "exact match",
			accept:    "text/markdown",
			mediaType: MIMETextMarkdown,
			want:      true,
		},
		{
			name:      "wildcard match",
			accept:    "*/*",
			mediaType: MIMETextMarkdown,
			want:      true,
		},
		{
			name:      "no match",
			accept:    "application/json",
			mediaType: MIMETextMarkdown,
			want:      false,
		},
		{
			name:      "multiple with match",
			accept:    "text/html, text/markdown;q=0.8",
			mediaType: MIMETextMarkdown,
			want:      true,
		},
		{
			name:      "multiple no match",
			accept:    "text/html, application/xml",
			mediaType: MIMETextMarkdown,
			want:      false,
		},
		{
			name:      "empty accept header",
			accept:    "",
			mediaType: MIMETextMarkdown,
			want:      true, // No Accept header = accept first offered
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			c := newContext()
			c.Request = req

			got := c.Accepts(tt.mediaType)
			if got != tt.want {
				t.Errorf("Accepts(%q) with Accept=%q = %v, want %v",
					tt.mediaType, tt.accept, got, tt.want)
			}
		})
	}
}

// TestContext_AcceptsAny tests multiple media type checking.
func TestContext_AcceptsAny(t *testing.T) {
	tests := []struct {
		name       string
		accept     string
		mediaTypes []string
		want       string
	}{
		{
			name:       "first match",
			accept:     "text/markdown",
			mediaTypes: []string{MIMETextMarkdown, MIMETextHTML, MIMEApplicationJSON},
			want:       MIMETextMarkdown,
		},
		{
			name:       "second match",
			accept:     "text/html",
			mediaTypes: []string{MIMETextMarkdown, MIMETextHTML, MIMEApplicationJSON},
			want:       MIMETextHTML,
		},
		{
			name:       "q-value priority",
			accept:     "text/html;q=0.9, text/markdown;q=1.0",
			mediaTypes: []string{MIMETextMarkdown, MIMETextHTML, MIMEApplicationJSON},
			want:       MIMETextMarkdown, // Higher q-value
		},
		{
			name:       "no match",
			accept:     "application/xml",
			mediaTypes: []string{MIMETextMarkdown, MIMETextHTML, MIMEApplicationJSON},
			want:       "",
		},
		{
			name:       "wildcard",
			accept:     "*/*",
			mediaTypes: []string{MIMETextMarkdown, MIMETextHTML},
			want:       MIMETextMarkdown, // First offered
		},
		{
			name:       "empty accept",
			accept:     "",
			mediaTypes: []string{MIMETextMarkdown, MIMETextHTML},
			want:       MIMETextMarkdown, // First offered
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			c := newContext()
			c.Request = req

			got := c.AcceptsAny(tt.mediaTypes...)
			if got != tt.want {
				t.Errorf("AcceptsAny(%v) with Accept=%q = %q, want %q",
					tt.mediaTypes, tt.accept, got, tt.want)
			}
		})
	}
}

// TestContext_Markdown tests markdown response.
func TestContext_Markdown(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "simple markdown",
			content: "# Hello World",
		},
		{
			name: "multiline markdown",
			content: `# API Documentation

## Endpoints
- GET /users
- POST /users`,
		},
		{
			name:    "empty content",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", http.NoBody)
			w := httptest.NewRecorder()
			c := newContext()
			c.init(w, req, nil, nil)

			err := c.Markdown(tt.content)
			if err != nil {
				t.Fatalf("Markdown() error = %v", err)
			}

			// Check status code
			if w.Code != 200 {
				t.Errorf("status code = %d, want 200", w.Code)
			}

			// Check Content-Type
			contentType := w.Header().Get("Content-Type")
			want := MIMETextMarkdown + "; charset=utf-8"
			if contentType != want {
				t.Errorf("Content-Type = %q, want %q", contentType, want)
			}

			// Check body
			body := w.Body.String()
			if body != tt.content {
				t.Errorf("body = %q, want %q", body, tt.content)
			}
		})
	}
}
