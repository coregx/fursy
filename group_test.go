// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRouter_Group tests basic group creation.
func TestRouter_Group(t *testing.T) {
	t.Run("creates group with prefix", func(t *testing.T) {
		r := New()
		g := r.Group("/api")

		// g cannot be nil - Group() always returns a value
		// Check fields directly
		if g.prefix != "/api" {
			t.Errorf("expected prefix '/api', got %s", g.prefix)
		}

		if g.router != r {
			t.Error("group router reference is incorrect")
		}
	})

	t.Run("creates group with middleware", func(t *testing.T) {
		r := New()
		mw1 := func(c *Context) error { return c.Next() }
		mw2 := func(c *Context) error { return c.Next() }

		g := r.Group("/api", mw1, mw2)

		if len(g.middleware) != 2 {
			t.Errorf("expected 2 middleware, got %d", len(g.middleware))
		}
	})

	t.Run("creates group without middleware", func(t *testing.T) {
		r := New()
		g := r.Group("/api")

		if len(g.middleware) != 0 {
			t.Errorf("expected 0 middleware, got %d", len(g.middleware))
		}
	})
}

// TestGroup_Use tests adding middleware to a group.
func TestGroup_Use(t *testing.T) {
	t.Run("adds single middleware", func(t *testing.T) {
		r := New()
		g := r.Group("/api")
		mw := func(c *Context) error { return c.Next() }

		result := g.Use(mw)

		if result != g {
			t.Error("Use() should return group for chaining")
		}

		if len(g.middleware) != 1 {
			t.Errorf("expected 1 middleware, got %d", len(g.middleware))
		}
	})

	t.Run("adds multiple middleware", func(t *testing.T) {
		r := New()
		g := r.Group("/api")
		mw1 := func(c *Context) error { return c.Next() }
		mw2 := func(c *Context) error { return c.Next() }

		g.Use(mw1, mw2)

		if len(g.middleware) != 2 {
			t.Errorf("expected 2 middleware, got %d", len(g.middleware))
		}
	})

	t.Run("chains Use calls", func(t *testing.T) {
		r := New()
		g := r.Group("/api")
		mw1 := func(c *Context) error { return c.Next() }
		mw2 := func(c *Context) error { return c.Next() }

		g.Use(mw1).Use(mw2)

		if len(g.middleware) != 2 {
			t.Errorf("expected 2 middleware, got %d", len(g.middleware))
		}
	})
}

// TestGroup_Routes tests registering routes on a group.
func TestGroup_Routes(t *testing.T) {
	t.Run("registers GET route with prefix", func(t *testing.T) {
		r := New()
		g := r.Group("/api")

		g.GET("/users", func(c *Context) error {
			return c.String(200, "users")
		})

		req := httptest.NewRequest("GET", "/api/users", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "users" {
			t.Errorf("expected 'users', got %s", w.Body.String())
		}
	})

	t.Run("registers all HTTP methods", func(t *testing.T) {
		r := New()
		g := r.Group("/api")

		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

		for _, method := range methods {
			switch method {
			case "GET":
				g.GET("/test", func(c *Context) error { return c.String(200, method) })
			case "POST":
				g.POST("/test", func(c *Context) error { return c.String(200, method) })
			case "PUT":
				g.PUT("/test", func(c *Context) error { return c.String(200, method) })
			case "DELETE":
				g.DELETE("/test", func(c *Context) error { return c.String(200, method) })
			case "PATCH":
				g.PATCH("/test", func(c *Context) error { return c.String(200, method) })
			case "HEAD":
				g.HEAD("/test", func(c *Context) error { return c.NoContent(200) })
			case "OPTIONS":
				g.OPTIONS("/test", func(c *Context) error { return c.NoContent(200) })
			}
		}

		for _, method := range methods {
			req := httptest.NewRequest(method, "/api/test", http.NoBody)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("%s: expected status 200, got %d", method, w.Code)
			}
		}
	})

	t.Run("path concatenation works correctly", func(t *testing.T) {
		r := New()
		g := r.Group("/api/v1")

		g.GET("/users/:id", func(c *Context) error {
			id := c.Param("id")
			return c.String(200, "user:"+id)
		})

		req := httptest.NewRequest("GET", "/api/v1/users/123", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "user:123" {
			t.Errorf("expected 'user:123', got %s", w.Body.String())
		}
	})
}

// TestGroup_NestedGroups tests nested group functionality.
func TestGroup_NestedGroups(t *testing.T) {
	t.Run("2-level nesting", func(t *testing.T) {
		r := New()

		api := r.Group("/api")
		v1 := api.Group("/v1")

		v1.GET("/users", func(c *Context) error {
			return c.String(200, "v1-users")
		})

		req := httptest.NewRequest("GET", "/api/v1/users", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "v1-users" {
			t.Errorf("expected 'v1-users', got %s", w.Body.String())
		}
	})

	t.Run("3-level nesting", func(t *testing.T) {
		r := New()

		api := r.Group("/api")
		v1 := api.Group("/v1")
		admin := v1.Group("/admin")

		admin.GET("/settings", func(c *Context) error {
			return c.String(200, "admin-settings")
		})

		req := httptest.NewRequest("GET", "/api/v1/admin/settings", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "admin-settings" {
			t.Errorf("expected 'admin-settings', got %s", w.Body.String())
		}
	})

	t.Run("middleware inheritance in nested groups", func(t *testing.T) {
		r := New()
		var executed []string

		api := r.Group("/api")
		api.Use(func(c *Context) error {
			executed = append(executed, "api-mw")
			return c.Next()
		})

		v1 := api.Group("/v1") // Should inherit api middleware
		v1.GET("/users", func(c *Context) error {
			executed = append(executed, "handler")
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/api/v1/users", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if len(executed) != 2 {
			t.Fatalf("expected 2 executions, got %d: %v", len(executed), executed)
		}

		if executed[0] != "api-mw" || executed[1] != "handler" {
			t.Errorf("unexpected execution order: %v", executed)
		}
	})

	t.Run("nested group with custom middleware", func(t *testing.T) {
		r := New()
		var executed []string

		api := r.Group("/api")
		api.Use(func(c *Context) error {
			executed = append(executed, "api-mw")
			return c.Next()
		})

		// Create v1 group with custom middleware (does NOT inherit api-mw)
		v1 := api.Group("/v1", func(c *Context) error {
			executed = append(executed, "v1-mw")
			return c.Next()
		})

		v1.GET("/users", func(c *Context) error {
			executed = append(executed, "handler")
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/api/v1/users", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should only execute v1-mw + handler (not api-mw)
		if len(executed) != 2 {
			t.Fatalf("expected 2 executions, got %d: %v", len(executed), executed)
		}

		if executed[0] != "v1-mw" || executed[1] != "handler" {
			t.Errorf("unexpected execution order: %v", executed)
		}
	})
}

// TestGroup_MiddlewareOrder tests middleware execution order.
func TestGroup_MiddlewareOrder(t *testing.T) {
	t.Run("router → group → handler", func(t *testing.T) {
		r := New()
		var executed []string

		r.Use(func(c *Context) error {
			executed = append(executed, "router-mw")
			return c.Next()
		})

		g := r.Group("/api")
		g.Use(func(c *Context) error {
			executed = append(executed, "group-mw")
			return c.Next()
		})

		g.GET("/users", func(c *Context) error {
			executed = append(executed, "handler")
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/api/users", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		expected := []string{"router-mw", "group-mw", "handler"}
		if len(executed) != len(expected) {
			t.Fatalf("expected %d executions, got %d: %v", len(expected), len(executed), executed)
		}

		for i, exp := range expected {
			if executed[i] != exp {
				t.Errorf("step %d: expected %s, got %s", i, exp, executed[i])
			}
		}
	})

	t.Run("multiple router middleware + group middleware", func(t *testing.T) {
		r := New()
		var executed []string

		r.Use(func(c *Context) error {
			executed = append(executed, "router-mw1")
			return c.Next()
		})
		r.Use(func(c *Context) error {
			executed = append(executed, "router-mw2")
			return c.Next()
		})

		g := r.Group("/api")
		g.Use(func(c *Context) error {
			executed = append(executed, "group-mw1")
			return c.Next()
		})
		g.Use(func(c *Context) error {
			executed = append(executed, "group-mw2")
			return c.Next()
		})

		g.GET("/users", func(c *Context) error {
			executed = append(executed, "handler")
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/api/users", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		expected := []string{"router-mw1", "router-mw2", "group-mw1", "group-mw2", "handler"}
		if len(executed) != len(expected) {
			t.Fatalf("expected %d executions, got %d: %v", len(expected), len(executed), executed)
		}

		for i, exp := range expected {
			if executed[i] != exp {
				t.Errorf("step %d: expected %s, got %s", i, exp, executed[i])
			}
		}
	})
}

// TestGroup_EdgeCases tests edge cases in path handling.
func TestGroup_EdgeCases(t *testing.T) {
	t.Run("empty prefix", func(t *testing.T) {
		r := New()
		g := r.Group("")

		g.GET("/users", func(c *Context) error {
			return c.String(200, "users")
		})

		req := httptest.NewRequest("GET", "/users", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("prefix without leading slash causes panic", func(t *testing.T) {
		// Documented behavior: group prefix should have leading slash
		// radix tree requires paths to start with '/'
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for prefix without leading slash")
			}
		}()

		r := New()
		g := r.Group("api") // Missing leading slash - should panic

		g.GET("/users", func(c *Context) error {
			return c.String(200, "users")
		})
	})

	t.Run("multiple groups on same router", func(t *testing.T) {
		r := New()

		api := r.Group("/api")
		api.GET("/users", func(c *Context) error {
			return c.String(200, "api-users")
		})

		admin := r.Group("/admin")
		admin.GET("/users", func(c *Context) error {
			return c.String(200, "admin-users")
		})

		// Test /api/users
		req1 := httptest.NewRequest("GET", "/api/users", http.NoBody)
		w1 := httptest.NewRecorder()
		r.ServeHTTP(w1, req1)

		if w1.Body.String() != "api-users" {
			t.Errorf("expected 'api-users', got %s", w1.Body.String())
		}

		// Test /admin/users
		req2 := httptest.NewRequest("GET", "/admin/users", http.NoBody)
		w2 := httptest.NewRecorder()
		r.ServeHTTP(w2, req2)

		if w2.Body.String() != "admin-users" {
			t.Errorf("expected 'admin-users', got %s", w2.Body.String())
		}
	})
}

// TestGroup_DataPassing tests passing data through group middleware.
func TestGroup_DataPassing(t *testing.T) {
	r := New()

	g := r.Group("/api")
	g.Use(func(c *Context) error {
		c.Set("groupData", "from-group")
		return c.Next()
	})

	g.GET("/users", func(c *Context) error {
		data := c.GetString("groupData")
		return c.String(200, "data:"+data)
	})

	req := httptest.NewRequest("GET", "/api/users", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Body.String() != "data:from-group" {
		t.Errorf("expected 'data:from-group', got %s", w.Body.String())
	}
}

// TestGroup_ErrorHandling tests error handling in group middleware.
func TestGroup_ErrorHandling(t *testing.T) {
	t.Run("group middleware error stops execution", func(t *testing.T) {
		r := New()
		handlerCalled := false

		g := r.Group("/api")
		g.Use(func(_ *Context) error {
			// Middleware returns error - handler should not be called
			return &testError{message: "group error"}
		})

		g.GET("/users", func(c *Context) error {
			handlerCalled = true
			return c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/api/users", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if handlerCalled {
			t.Error("handler was called after middleware error")
		}

		if w.Code != 500 {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

// testError is a test error type.
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
