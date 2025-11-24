package radix

import (
	"testing"
)

// Helper function to convert []Param to map[string]string for easier testing.
func paramsToMap(params []Param) map[string]string {
	m := make(map[string]string, len(params))
	for _, p := range params {
		m[p.Key] = p.Value
	}
	return m
}

// TestTree_ParamWithMultipleStaticChildren tests the fix for the bug where
// routes with :param followed by multiple static segments would panic.
func TestTree_ParamWithMultipleStaticChildren(t *testing.T) {
	tree := New()

	routes := []struct {
		path    string
		handler string
	}{
		{"/api/users/:id", "handler-1"},
		{"/api/users/:id/activate", "handler-2"},
		{"/api/users/:id/deactivate", "handler-3"},
		{"/api/users/:id/password", "handler-4"},
		{"/api/users/:id/avatar", "handler-5"},
		{"/api/users/:id/posts", "handler-6"},
		{"/api/users/:id/settings", "handler-7"},
	}

	for i, route := range routes {
		err := tree.Insert(route.path, route.handler)
		if err != nil {
			t.Fatalf("Failed to insert route %d (%s): %v", i, route.path, err)
		}
	}

	// Verify all routes are accessible
	tests := []struct {
		path       string
		wantValue  string
		wantParams map[string]string
	}{
		{"/api/users/123", "handler-1", map[string]string{"id": "123"}},
		{"/api/users/456/activate", "handler-2", map[string]string{"id": "456"}},
		{"/api/users/789/deactivate", "handler-3", map[string]string{"id": "789"}},
		{"/api/users/abc/password", "handler-4", map[string]string{"id": "abc"}},
		{"/api/users/xyz/avatar", "handler-5", map[string]string{"id": "xyz"}},
		{"/api/users/111/posts", "handler-6", map[string]string{"id": "111"}},
		{"/api/users/222/settings", "handler-7", map[string]string{"id": "222"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			value, params, found := tree.Lookup(tt.path)
			if !found || value == nil {
				t.Fatalf("Expected route %s to be found", tt.path)
			}

			if got := value.(string); got != tt.wantValue {
				t.Errorf("Lookup(%q) value = %v, want %v", tt.path, got, tt.wantValue)
			}

			paramsMap := paramsToMap(params)
			if len(paramsMap) != len(tt.wantParams) {
				t.Errorf("Lookup(%q) params count = %d, want %d", tt.path, len(paramsMap), len(tt.wantParams))
			}

			for key, wantVal := range tt.wantParams {
				if gotVal, ok := paramsMap[key]; !ok {
					t.Errorf("Lookup(%q) missing param %q", tt.path, key)
				} else if gotVal != wantVal {
					t.Errorf("Lookup(%q) param %q = %q, want %q", tt.path, key, gotVal, wantVal)
				}
			}
		})
	}
}

// TestTree_NestedParams tests nested params with static segments.
func TestTree_NestedParams(t *testing.T) {
	tree := New()

	routes := []string{
		"/users/:user_id",
		"/users/:user_id/posts",
		"/users/:user_id/posts/:post_id",
		"/users/:user_id/posts/:post_id/comments",
		"/users/:user_id/posts/:post_id/comments/:comment_id",
	}

	for i, route := range routes {
		handler := route // Use path as handler for testing
		err := tree.Insert(route, handler)
		if err != nil {
			t.Fatalf("Failed to insert route %d (%s): %v", i, route, err)
		}
	}

	// Test all routes are accessible
	tests := []struct {
		path       string
		wantParams map[string]string
	}{
		{"/users/u1", map[string]string{"user_id": "u1"}},
		{"/users/u2/posts", map[string]string{"user_id": "u2"}},
		{"/users/u3/posts/p1", map[string]string{"user_id": "u3", "post_id": "p1"}},
		{"/users/u4/posts/p2/comments", map[string]string{"user_id": "u4", "post_id": "p2"}},
		{"/users/u5/posts/p3/comments/c1", map[string]string{"user_id": "u5", "post_id": "p3", "comment_id": "c1"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			value, params, found := tree.Lookup(tt.path)
			if !found || value == nil {
				t.Fatalf("Expected route %s to be found", tt.path)
			}

			paramsMap := paramsToMap(params)
			if len(paramsMap) != len(tt.wantParams) {
				t.Errorf("Lookup(%q) params count = %d, want %d", tt.path, len(paramsMap), len(tt.wantParams))
			}

			for key, wantVal := range tt.wantParams {
				if gotVal, ok := paramsMap[key]; !ok {
					t.Errorf("Lookup(%q) missing param %q", tt.path, key)
				} else if gotVal != wantVal {
					t.Errorf("Lookup(%q) param %q = %q, want %q", tt.path, key, gotVal, wantVal)
				}
			}
		})
	}
}

// TestTree_AlternatingParamStatic tests alternating param/static segments.
func TestTree_AlternatingParamStatic(t *testing.T) {
	tree := New()

	routes := []string{
		"/a/:b/c/:d/e",
		"/a/:b/c/:d/f",
		"/a/:b/c/:d/g",
		"/:a/b/c/d/e",
		"/:a/b/c/d/f",
	}

	for _, route := range routes {
		err := tree.Insert(route, route)
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", route, err)
		}
	}

	// Test routing works correctly
	tests := []struct {
		path       string
		wantFound  bool
		wantParams map[string]string
	}{
		{"/a/123/c/456/e", true, map[string]string{"b": "123", "d": "456"}},
		{"/a/xyz/c/abc/f", true, map[string]string{"b": "xyz", "d": "abc"}},
		{"/a/111/c/222/g", true, map[string]string{"b": "111", "d": "222"}},
		{"/foo/b/c/d/e", true, map[string]string{"a": "foo"}},
		{"/bar/b/c/d/f", true, map[string]string{"a": "bar"}},
		{"/a/123/c/456/z", false, nil}, // Not registered
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			value, params, found := tree.Lookup(tt.path)

			if tt.wantFound && (!found || value == nil) {
				t.Errorf("Lookup(%q) expected to find route, but got nil", tt.path)
				return
			}

			if !tt.wantFound && found {
				t.Errorf("Lookup(%q) expected not found, but found route", tt.path)
				return
			}

			if !tt.wantFound {
				return
			}

			paramsMap := paramsToMap(params)
			for key, wantVal := range tt.wantParams {
				if gotVal, ok := paramsMap[key]; !ok {
					t.Errorf("Lookup(%q) missing param %q", tt.path, key)
				} else if gotVal != wantVal {
					t.Errorf("Lookup(%q) param %q = %q, want %q", tt.path, key, gotVal, wantVal)
				}
			}
		})
	}
}

// TestTree_ParamWithLongStaticTail tests param followed by long static path.
func TestTree_ParamWithLongStaticTail(t *testing.T) {
	tree := New()

	routes := []string{
		"/:a/b/c/d/e/f/g",
		"/:a/b/c/d/e/f/h",
		"/:a/b/c/d/e/x/y",
	}

	for _, route := range routes {
		err := tree.Insert(route, route)
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", route, err)
		}
	}

	tests := []struct {
		path      string
		wantFound bool
		wantParam string
	}{
		{"/param1/b/c/d/e/f/g", true, "param1"},
		{"/param2/b/c/d/e/f/h", true, "param2"},
		{"/param3/b/c/d/e/x/y", true, "param3"},
		{"/param4/b/c/d/e/f/z", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			value, params, found := tree.Lookup(tt.path)

			if tt.wantFound && (!found || value == nil) {
				t.Errorf("Lookup(%q) expected to find route", tt.path)
				return
			}

			if !tt.wantFound && found {
				t.Errorf("Lookup(%q) expected not found, but found route", tt.path)
				return
			}

			if !tt.wantFound {
				return
			}

			paramsMap := paramsToMap(params)
			if got, ok := paramsMap["a"]; !ok {
				t.Errorf("Lookup(%q) missing param 'a'", tt.path)
			} else if got != tt.wantParam {
				t.Errorf("Lookup(%q) param 'a' = %q, want %q", tt.path, got, tt.wantParam)
			}
		})
	}
}

// TestTree_ConsecutiveParams tests multiple consecutive params.
func TestTree_ConsecutiveParams(t *testing.T) {
	tree := New()

	routes := []string{
		"/:a/:b/:c",
		"/:a/:b/:c/d",
		"/:a/:b/:c/e",
	}

	for _, route := range routes {
		err := tree.Insert(route, route)
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", route, err)
		}
	}

	tests := []struct {
		path       string
		wantFound  bool
		wantParams map[string]string
	}{
		{"/p1/p2/p3", true, map[string]string{"a": "p1", "b": "p2", "c": "p3"}},
		{"/p1/p2/p3/d", true, map[string]string{"a": "p1", "b": "p2", "c": "p3"}},
		{"/p1/p2/p3/e", true, map[string]string{"a": "p1", "b": "p2", "c": "p3"}},
		{"/p1/p2/p3/f", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			value, params, found := tree.Lookup(tt.path)

			if tt.wantFound && (!found || value == nil) {
				t.Errorf("Lookup(%q) expected to find route", tt.path)
				return
			}

			if !tt.wantFound && found {
				t.Errorf("Lookup(%q) expected not found, but found route", tt.path)
				return
			}

			if !tt.wantFound {
				return
			}

			paramsMap := paramsToMap(params)
			for key, wantVal := range tt.wantParams {
				if gotVal, ok := paramsMap[key]; !ok {
					t.Errorf("Lookup(%q) missing param %q", tt.path, key)
				} else if gotVal != wantVal {
					t.Errorf("Lookup(%q) param %q = %q, want %q", tt.path, key, gotVal, wantVal)
				}
			}
		})
	}
}

// TestTree_StaticVsParam tests that static routes have priority over param routes.
func TestTree_StaticVsParam(t *testing.T) {
	tree := New()

	routes := []struct {
		path  string
		value string
	}{
		{"/users/admin", "static-admin"},
		{"/users/:id", "param-id"},
		{"/users/:id/posts", "param-posts"},
		{"/users/admin/posts", "static-admin-posts"},
	}

	for _, route := range routes {
		err := tree.Insert(route.path, route.value)
		if err != nil {
			t.Fatalf("Failed to insert %s: %v", route.path, err)
		}
	}

	tests := []struct {
		path      string
		wantValue string
		wantParam string
	}{
		{"/users/admin", "static-admin", ""},             // Static has priority
		{"/users/123", "param-id", "123"},                // Param fallback
		{"/users/123/posts", "param-posts", "123"},       // Param + static
		{"/users/admin/posts", "static-admin-posts", ""}, // Static has priority
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			value, params, found := tree.Lookup(tt.path)

			if !found || value == nil {
				t.Fatalf("Lookup(%q) expected to find route", tt.path)
			}

			if got := value.(string); got != tt.wantValue {
				t.Errorf("Lookup(%q) = %q, want %q", tt.path, got, tt.wantValue)
			}

			paramsMap := paramsToMap(params)
			if tt.wantParam != "" {
				if id, ok := paramsMap["id"]; !ok {
					t.Errorf("Lookup(%q) expected param 'id'", tt.path)
				} else if id != tt.wantParam {
					t.Errorf("Lookup(%q) param 'id' = %q, want %q", tt.path, id, tt.wantParam)
				}
			}
		})
	}
}
