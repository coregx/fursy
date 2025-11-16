package radix

import (
	"fmt"
	"testing"
)

// TestTree_New tests tree creation.
func TestTree_New(t *testing.T) {
	tree := New()
	if tree == nil || tree.root == nil {
		t.Fatal("New() returned nil tree or nil root")
	}
	if tree.root.nType != root {
		t.Errorf("root node type = %v, want %v", tree.root.nType, root)
	}
}

// TestTree_InsertStatic tests inserting static routes.
func TestTree_InsertStatic(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid simple", "/users", false},
		{"valid nested", "/api/v1/users", false},
		{"valid root", "/", false},
		{"empty path", "", true},
		{"no leading slash", "users", true},
		{"duplicate route", "/users", true}, // Second insert of same path
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := New()

			// First insert (always succeeds unless invalid path)
			if tt.name != "duplicate route" {
				err := tree.Insert(tt.path, "handler1")
				if (err != nil) != tt.wantErr {
					t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			// Test duplicate
			tree.Insert("/users", "handler1")
			err := tree.Insert("/users", "handler2")
			if err == nil {
				t.Error("Insert() should return error for duplicate route")
			}
		})
	}
}

// TestTree_LookupStatic tests looking up static routes.
func TestTree_LookupStatic(t *testing.T) {
	tree := New()
	tree.Insert("/users", "handler1")
	tree.Insert("/api/v1/posts", "handler2")
	tree.Insert("/api/v1/users", "handler3")

	tests := []struct {
		name        string
		path        string
		wantHandler interface{}
		wantFound   bool
	}{
		{"exact match", "/users", "handler1", true},
		{"nested match", "/api/v1/posts", "handler2", true},
		{"another nested", "/api/v1/users", "handler3", true},
		{"not found", "/posts", nil, false},
		{"partial match", "/api", nil, false},
		{"extra segment", "/users/123", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, params, found := tree.Lookup(tt.path)

			if found != tt.wantFound {
				t.Errorf("Lookup() found = %v, want %v", found, tt.wantFound)
			}

			if tt.wantFound {
				if handler != tt.wantHandler {
					t.Errorf("Lookup() handler = %v, want %v", handler, tt.wantHandler)
				}
				if len(params) != 0 {
					t.Errorf("Lookup() params = %v, want empty", params)
				}
			}
		})
	}
}

// TestTree_InsertParameter tests inserting routes with parameters.
func TestTree_InsertParameter(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"single param", "/users/:id", false},
		{"multiple params", "/posts/:category/:id", false},
		{"param after static", "/api/users/:id", false},
		{"param invalid name", "/users/:", true},
		{"param empty name", "/users/:/posts", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := New()
			err := tree.Insert(tt.path, "handler")
			if (err != nil) != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTree_LookupParameter tests looking up parametric routes.
func TestTree_LookupParameter(t *testing.T) {
	tree := New()
	tree.Insert("/users/:id", "handler1")
	tree.Insert("/posts/:category/:id", "handler2")
	tree.Insert("/api/users/:id/posts", "handler3")

	tests := []struct {
		name        string
		path        string
		wantHandler interface{}
		wantParams  []Param
		wantFound   bool
	}{
		{
			name:        "single param",
			path:        "/users/123",
			wantHandler: "handler1",
			wantParams:  []Param{{Key: "id", Value: "123"}},
			wantFound:   true,
		},
		{
			name:        "two params",
			path:        "/posts/tech/456",
			wantHandler: "handler2",
			wantParams: []Param{
				{Key: "category", Value: "tech"},
				{Key: "id", Value: "456"},
			},
			wantFound: true,
		},
		{
			name:        "param with nested",
			path:        "/api/users/789/posts",
			wantHandler: "handler3",
			wantParams:  []Param{{Key: "id", Value: "789"}},
			wantFound:   true,
		},
		{
			name:      "not found",
			path:      "/products/123",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, params, found := tree.Lookup(tt.path)

			if found != tt.wantFound {
				t.Errorf("Lookup() found = %v, want %v", found, tt.wantFound)
			}

			if !tt.wantFound {
				return
			}

			if handler != tt.wantHandler {
				t.Errorf("Lookup() handler = %v, want %v", handler, tt.wantHandler)
			}

			if len(params) != len(tt.wantParams) {
				t.Fatalf("Lookup() params count = %d, want %d", len(params), len(tt.wantParams))
			}

			for i, param := range params {
				if param.Key != tt.wantParams[i].Key {
					t.Errorf("param[%d].Key = %s, want %s", i, param.Key, tt.wantParams[i].Key)
				}
				if param.Value != tt.wantParams[i].Value {
					t.Errorf("param[%d].Value = %s, want %s", i, param.Value, tt.wantParams[i].Value)
				}
			}
		})
	}
}

// TestTree_InsertWildcard tests inserting catch-all routes.
func TestTree_InsertWildcard(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid wildcard", "/files/*filepath", false},
		{"wildcard at root", "/*path", false},
		{"wildcard invalid", "/files/*", true},       // No name
		{"wildcard not last", "/files/*/docs", true}, // Must be last
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := New()
			err := tree.Insert(tt.path, "handler")
			if (err != nil) != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTree_LookupWildcard tests looking up catch-all routes.
func TestTree_LookupWildcard(t *testing.T) {
	tree := New()
	tree.Insert("/files/*filepath", "handler1")
	tree.Insert("/*path", "handler2")

	tests := []struct {
		name        string
		path        string
		wantHandler interface{}
		wantParams  []Param
		wantFound   bool
	}{
		{
			name:        "wildcard simple",
			path:        "/files/readme.md",
			wantHandler: "handler1",
			wantParams:  []Param{{Key: "filepath", Value: "readme.md"}},
			wantFound:   true,
		},
		{
			name:        "wildcard nested",
			path:        "/files/docs/api/v1/README.md",
			wantHandler: "handler1",
			wantParams:  []Param{{Key: "filepath", Value: "docs/api/v1/README.md"}},
			wantFound:   true,
		},
		{
			name:        "root wildcard",
			path:        "/anything/goes/here",
			wantHandler: "handler2",
			wantParams:  []Param{{Key: "path", Value: "anything/goes/here"}},
			wantFound:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, params, found := tree.Lookup(tt.path)

			if found != tt.wantFound {
				t.Errorf("Lookup() found = %v, want %v", found, tt.wantFound)
			}

			if !tt.wantFound {
				return
			}

			if handler != tt.wantHandler {
				t.Errorf("Lookup() handler = %v, want %v", handler, tt.wantHandler)
			}

			if len(params) != len(tt.wantParams) {
				t.Fatalf("Lookup() params count = %d, want %d", len(params), len(tt.wantParams))
			}

			for i, param := range params {
				if param.Key != tt.wantParams[i].Key {
					t.Errorf("param[%d].Key = %s, want %s", i, param.Key, tt.wantParams[i].Key)
				}
				if param.Value != tt.wantParams[i].Value {
					t.Errorf("param[%d].Value = %s, want %s", i, param.Value, tt.wantParams[i].Value)
				}
			}
		})
	}
}

// TestTree_ConflictingRoutes tests route conflicts.
func TestTree_ConflictingRoutes(t *testing.T) {
	tests := []struct {
		name   string
		routes []string
		errIdx int // Index of route that should error
	}{
		{
			name:   "conflicting params",
			routes: []string{"/users/:id", "/users/:name"},
			errIdx: 1,
		},
		{
			name:   "param vs static priority",
			routes: []string{"/users/:id", "/users/new"},
			errIdx: -1, // Should both succeed (static has priority)
		},
		{
			name:   "duplicate static",
			routes: []string{"/users", "/users"},
			errIdx: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := New()

			for i, route := range tt.routes {
				err := tree.Insert(route, "handler")

				switch i {
				case tt.errIdx:
					if err == nil {
						t.Errorf("Insert(%s) should have returned error", route)
					}
				default:
					if tt.errIdx == -1 {
						if err != nil {
							t.Errorf("Insert(%s) unexpected error: %v", route, err)
						}
					} else if err != nil {
						t.Errorf("Insert(%s) unexpected error before conflict: %v", route, err)
					}
				}
			}
		})
	}
}

// TestTree_EdgeCases tests edge cases.
func TestTree_EdgeCases(t *testing.T) {
	tree := New()
	tree.Insert("/", "root-handler")
	tree.Insert("/users", "users-handler")
	tree.Insert("/users/:id", "user-handler")
	tree.Insert("/users/:id/posts", "user-posts-handler")

	tests := []struct {
		name      string
		path      string
		wantFound bool
	}{
		{"root path", "/", true},
		{"trailing slash", "/users/", false}, // Strict matching
		{"double slash", "//users", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, found := tree.Lookup(tt.path)
			if found != tt.wantFound {
				t.Errorf("Lookup(%s) found = %v, want %v", tt.path, found, tt.wantFound)
			}
		})
	}
}

// TestTree_PriorityOrdering tests that high-priority routes are checked first.
func TestTree_PriorityOrdering(t *testing.T) {
	tree := New()

	// Insert in order: less common to more common
	routes := []struct {
		path     string
		priority int // Simulated access frequency
	}{
		{"/api/rare", 1},
		{"/api/common", 100},
		{"/api/medium", 10},
	}

	for _, r := range routes {
		tree.Insert(r.path, r.path) // Use path as handler for testing
	}

	// After insertions, check that all routes are findable
	for _, r := range routes {
		handler, _, found := tree.Lookup(r.path)
		if !found {
			t.Errorf("Route %s not found after insertion", r.path)
		}
		if handler != r.path {
			t.Errorf("Route %s returned wrong handler", r.path)
		}
	}

	// TODO: Add priority increment logic and verify ordering
	// This test validates correctness; actual priority ordering
	// would be validated through benchmarks showing faster access
	// to frequently-accessed routes.
}

// TestTree_LargeRouteSet tests with a realistic number of routes.
func TestTree_LargeRouteSet(t *testing.T) {
	tree := New()

	// Insert 100 routes
	routes := make([]string, 100)
	for i := 0; i < 100; i++ {
		// Use string formatting instead of rune() to avoid control characters
		routes[i] = fmt.Sprintf("/api/v1/resource%d/:id", i)
		if err := tree.Insert(routes[i], i); err != nil {
			t.Fatalf("Insert(%s) failed: %v", routes[i], err)
		}
	}

	// Lookup all routes
	for i := range routes {
		// Replace :id with actual value
		testPath := fmt.Sprintf("/api/v1/resource%d/123", i)
		handler, params, found := tree.Lookup(testPath)

		if !found {
			t.Errorf("Route %s not found", testPath)
			continue
		}

		if handler.(int) != i {
			t.Errorf("Wrong handler for %s: got %v, want %d", testPath, handler, i)
		}

		if len(params) != 1 || params[0].Value != "123" {
			t.Errorf("Wrong params for %s: %v", testPath, params)
		}
	}
}
