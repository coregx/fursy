package radix

import "testing"

// FuzzRadixLookup seeds a tree with known routes and fuzzes random lookup paths.
// Must never panic regardless of input.
func FuzzRadixLookup(f *testing.F) {
	tree := New()
	routes := []string{
		"/",
		"/users",
		"/users/:id",
		"/users/:id/posts",
		"/files/*path",
		"/api/v1/documents:commit",
		"/api/v1/documents:batchGet",
		"/clients:getOrganization/:id",
	}
	for _, r := range routes {
		if err := tree.Insert(r, r); err != nil {
			f.Fatalf("Insert(%q): %v", r, err)
		}
	}

	f.Add("/users/42")
	f.Add("/files/a/b/c")
	f.Add("/api/v1/documents:commit")
	f.Add("/%00")
	f.Add("/users//empty")
	f.Add("/\\evil.com")
	f.Add("/a:b:c:d")
	f.Add("/*")
	f.Add("/:")
	f.Add("")

	f.Fuzz(func(_ *testing.T, path string) {
		buf := make([]Param, 0, 4)
		tree.Lookup(path, buf)
	})
}

// FuzzRoutePattern fuzzes route registration with random path grammar.
// Must never panic on any input — errors are acceptable, panics are not.
func FuzzRoutePattern(f *testing.F) {
	f.Add("/users/:id")
	f.Add("/docs:commit")
	f.Add("/:a/:b/:c/:d")
	f.Add("/files/*rest")
	f.Add("/a%20b/c:d")
	f.Add("/:id(\\d+)")
	f.Add("/a/b/c/d/e/f/g")
	f.Add("/")

	f.Fuzz(func(_ *testing.T, path string) {
		if path == "" || path[0] != '/' {
			return
		}
		tree := New()
		_ = tree.Insert(path, "handler")
	})
}
