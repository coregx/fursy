package fursy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkRouter_StaticRoute benchmarks routing for a simple static route.
func BenchmarkRouter_StaticRoute(b *testing.B) {
	router := New()
	router.GET("/users", func(c *Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/users", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_ParameterRoute benchmarks routing with URL parameters.
func BenchmarkRouter_ParameterRoute(b *testing.B) {
	router := New()
	router.GET("/users/:id", func(c *Context) error {
		_ = c.Param("id")
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/users/123", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_ParameterRoute_MultipleParams benchmarks multiple parameters.
func BenchmarkRouter_ParameterRoute_MultipleParams(b *testing.B) {
	router := New()
	router.GET("/posts/:category/:postID", func(c *Context) error {
		_ = c.Param("category")
		_ = c.Param("postID")
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/posts/tech/456", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_WildcardRoute benchmarks catch-all wildcard routing.
func BenchmarkRouter_WildcardRoute(b *testing.B) {
	router := New()
	router.GET("/files/*path", func(c *Context) error {
		_ = c.Param("path")
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/files/documents/report.pdf", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_MultipleRoutes benchmarks routing with many registered routes.
func BenchmarkRouter_MultipleRoutes(b *testing.B) {
	router := New()

	// Register multiple routes to simulate realistic scenario.
	routes := []string{
		"/",
		"/users",
		"/users/:id",
		"/posts",
		"/posts/:category/:id",
		"/api/v1/users",
		"/api/v1/users/:id/profile",
		"/api/v1/posts",
		"/api/v1/posts/:id",
		"/files/*path",
		"/static/*path",
		"/admin/users",
		"/admin/posts",
	}

	handler := func(c *Context) error {
		return c.NoContent(http.StatusOK)
	}

	for _, route := range routes {
		router.GET(route, handler)
	}

	req := httptest.NewRequest("GET", "/users/123", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_DeepNesting benchmarks deeply nested routes.
func BenchmarkRouter_DeepNesting(b *testing.B) {
	router := New()
	router.GET("/api/v1/organizations/:orgID/projects/:projectID/issues/:issueID/comments/:commentID",
		func(c *Context) error {
			_ = c.Param("orgID")
			_ = c.Param("projectID")
			_ = c.Param("issueID")
			_ = c.Param("commentID")
			return c.NoContent(http.StatusOK)
		})

	req := httptest.NewRequest("GET", "/api/v1/organizations/123/projects/456/issues/789/comments/101", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_NotFound benchmarks 404 lookup performance.
func BenchmarkRouter_NotFound(b *testing.B) {
	router := New()
	router.GET("/users", func(c *Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/posts", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_MethodNotAllowed benchmarks 405 lookup performance.
func BenchmarkRouter_MethodNotAllowed(b *testing.B) {
	router := New()
	router.GET("/users", func(c *Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("POST", "/users", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkContext_Param benchmarks Box.Param() extraction.
func BenchmarkContext_Param(b *testing.B) {
	c := newContext()
	c.params = []Param{
		{Key: "id", Value: "123"},
		{Key: "category", Value: "tech"},
		{Key: "postID", Value: "456"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = c.Param("id")
	}
}

// BenchmarkContext_Query benchmarks Box.Query() extraction.
func BenchmarkContext_Query(b *testing.B) {
	req := httptest.NewRequest("GET", "/test?page=1&limit=10&sort=asc", http.NoBody)
	c := newContext()
	c.Request = req

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = c.Query("page")
	}
}

// BenchmarkRouter_RootPath benchmarks root path routing.
func BenchmarkRouter_RootPath(b *testing.B) {
	router := New()
	router.GET("/", func(c *Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_LongStaticPath benchmarks long static paths.
func BenchmarkRouter_LongStaticPath(b *testing.B) {
	router := New()
	router.GET("/api/v1/organizations/settings/security/authentication/providers", func(c *Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/organizations/settings/security/authentication/providers", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkRouter_MixedRoutes benchmarks mix of static and parametric routes.
func BenchmarkRouter_MixedRoutes(b *testing.B) {
	router := New()

	// Mix of static and parametric routes.
	router.GET("/api/users", func(_ *Context) error { return nil })
	router.GET("/api/users/:id", func(_ *Context) error { return nil })
	router.GET("/api/posts", func(_ *Context) error { return nil })
	router.GET("/api/posts/:category/:id", func(_ *Context) error { return nil })
	router.GET("/api/comments/:id", func(_ *Context) error { return nil })

	req := httptest.NewRequest("GET", "/api/posts/tech/42", http.NoBody)
	w := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

// BenchmarkContext_JSON benchmarks JSON response encoding.
func BenchmarkContext_JSON(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	data := map[string]string{
		"id":      "123",
		"name":    "John Doe",
		"email":   "john@example.com",
		"status":  "active",
		"role":    "admin",
		"created": "2024-01-01",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = c.JSON(200, data)
	}
}

// BenchmarkContext_String benchmarks String response.
func BenchmarkContext_String(b *testing.B) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", http.NoBody)

	c := newContext()
	c.Response = w
	c.Request = req

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = c.String(200, "Hello, World!")
	}
}

// BenchmarkContext_Pooling benchmarks Context pooling efficiency.
func BenchmarkContext_Pooling(b *testing.B) {
	router := New()
	router.GET("/test", func(c *Context) error {
		return c.String(200, "OK")
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
