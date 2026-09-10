// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package radix

import "testing"

// Edge case tests from Fable 5.1 audit (R1-R4).

// TestLookup_EmptyParamSegment tests behavior with empty path segments.
func TestLookup_EmptyParamSegment(t *testing.T) {
	tree := New()
	_ = tree.Insert("/a/:id/b", "handler")

	tests := []struct {
		name      string
		path      string
		wantFound bool
		wantID    string
	}{
		{"normal", "/a/123/b", true, "123"},
		{"empty segment not allowed", "/a//b", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, params, found := tree.Lookup(tt.path, nil)
			if found != tt.wantFound {
				t.Errorf("Lookup(%q): found=%v, want %v", tt.path, found, tt.wantFound)
				return
			}
			if found && len(params) > 0 && params[0].Value != tt.wantID {
				t.Errorf("Lookup(%q): param=%q, want %q", tt.path, params[0].Value, tt.wantID)
			}
		})
	}
}

// TestLookup_TrailingSlashParam tests params with trailing slash.
func TestLookup_TrailingSlashParam(t *testing.T) {
	tree := New()
	_ = tree.Insert("/users/:id", "handler")

	tests := []struct {
		name      string
		path      string
		wantFound bool
	}{
		{"normal", "/users/abc", true},
		{"with trailing slash", "/users/abc/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, found := tree.Lookup(tt.path, nil)
			if found != tt.wantFound {
				t.Errorf("Lookup(%q): found=%v, want %v", tt.path, found, tt.wantFound)
			}
		})
	}
}

// TestContains_BasicPaths verifies Contains works correctly.
func TestContains_BasicPaths(t *testing.T) {
	tree := New()
	_ = tree.Insert("/users", "h1")
	_ = tree.Insert("/users/:id", "h2")
	_ = tree.Insert("/files/*path", "h3")

	tests := []struct {
		path string
		want bool
	}{
		{"/users", true},
		{"/users/123", true},
		{"/files/a/b/c", true},
		{"/notfound", false},
		{"/users/123/extra", false},
	}

	for _, tt := range tests {
		if got := tree.Contains(tt.path); got != tt.want {
			t.Errorf("Contains(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
