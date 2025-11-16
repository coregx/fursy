// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package negotiate

import (
	"testing"
)

// Test parseAccept function.

func TestParseAccept_Empty(t *testing.T) {
	ranges := parseAccept("")

	if len(ranges) != 1 {
		t.Fatalf("Expected 1 range for empty header, got %d", len(ranges))
	}

	if ranges[0].Type != "*" || ranges[0].Subtype != "*" {
		t.Errorf("Expected */* for empty header, got %s/%s", ranges[0].Type, ranges[0].Subtype)
	}

	if ranges[0].Quality != 1.0 {
		t.Errorf("Expected quality 1.0, got %f", ranges[0].Quality)
	}
}

func TestParseAccept_Single(t *testing.T) {
	ranges := parseAccept("application/json")

	if len(ranges) != 1 {
		t.Fatalf("Expected 1 range, got %d", len(ranges))
	}

	mr := ranges[0]
	if mr.Type != "application" || mr.Subtype != "json" {
		t.Errorf("Expected application/json, got %s/%s", mr.Type, mr.Subtype)
	}

	if mr.Quality != 1.0 {
		t.Errorf("Expected quality 1.0 (default), got %f", mr.Quality)
	}
}

func TestParseAccept_Multiple(t *testing.T) {
	ranges := parseAccept("text/html, application/json, application/xml")

	if len(ranges) != 3 {
		t.Fatalf("Expected 3 ranges, got %d", len(ranges))
	}

	// All should have default quality 1.0, order preserved.
	expected := []struct{ typ, subtype string }{
		{"text", "html"},
		{"application", "json"},
		{"application", "xml"},
	}

	for i, exp := range expected {
		if ranges[i].Type != exp.typ || ranges[i].Subtype != exp.subtype {
			t.Errorf("Range %d: expected %s/%s, got %s/%s", i, exp.typ, exp.subtype, ranges[i].Type, ranges[i].Subtype)
		}
	}
}

func TestParseAccept_QWeighting(t *testing.T) {
	ranges := parseAccept("text/html, application/json;q=0.9, application/xml;q=0.8")

	if len(ranges) != 3 {
		t.Fatalf("Expected 3 ranges, got %d", len(ranges))
	}

	// Should be sorted by quality (highest first).
	expected := []struct {
		typ, subtype string
		quality      float64
	}{
		{"text", "html", 1.0},
		{"application", "json", 0.9},
		{"application", "xml", 0.8},
	}

	for i, exp := range expected {
		mr := ranges[i]
		if mr.Type != exp.typ || mr.Subtype != exp.subtype {
			t.Errorf("Range %d: expected %s/%s, got %s/%s", i, exp.typ, exp.subtype, mr.Type, mr.Subtype)
		}
		if mr.Quality != exp.quality {
			t.Errorf("Range %d: expected quality %f, got %f", i, exp.quality, mr.Quality)
		}
	}
}

func TestParseAccept_Wildcards(t *testing.T) {
	tests := []struct {
		name            string
		accept          string
		expectedFirst   string
		expectedQuality float64
	}{
		{
			name:            "full wildcard",
			accept:          "*/*",
			expectedFirst:   "*/*",
			expectedQuality: 1.0,
		},
		{
			name:            "type wildcard",
			accept:          "text/*",
			expectedFirst:   "text/*",
			expectedQuality: 1.0,
		},
		{
			name:            "mixed with wildcards",
			accept:          "text/html, text/*, */*;q=0.1",
			expectedFirst:   "text/html",
			expectedQuality: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranges := parseAccept(tt.accept)

			if len(ranges) == 0 {
				t.Fatal("Expected at least 1 range")
			}

			mr := ranges[0]
			got := mr.Type + "/" + mr.Subtype

			if got != tt.expectedFirst {
				t.Errorf("Expected first range %s, got %s", tt.expectedFirst, got)
			}

			if mr.Quality != tt.expectedQuality {
				t.Errorf("Expected quality %f, got %f", tt.expectedQuality, mr.Quality)
			}
		})
	}
}

func TestParseAccept_Parameters(t *testing.T) {
	ranges := parseAccept("application/json;q=0.9;charset=utf-8")

	if len(ranges) != 1 {
		t.Fatalf("Expected 1 range, got %d", len(ranges))
	}

	mr := ranges[0]

	if mr.Quality != 0.9 {
		t.Errorf("Expected quality 0.9, got %f", mr.Quality)
	}

	if charset, ok := mr.Parameters["charset"]; !ok || charset != "utf-8" {
		t.Errorf("Expected charset=utf-8, got %v", charset)
	}
}

func TestParseAccept_RFC9110_Example(t *testing.T) {
	// RFC 9110 example: client prefers JSON, but accepts XML and anything else as fallback.
	ranges := parseAccept("application/json, application/xml;q=0.9, */*;q=0.1")

	if len(ranges) != 3 {
		t.Fatalf("Expected 3 ranges, got %d", len(ranges))
	}

	// Should be sorted: JSON (q=1.0) > XML (q=0.9) > wildcard (q=0.1).
	expected := []struct {
		typ, subtype string
		quality      float64
	}{
		{"application", "json", 1.0},
		{"application", "xml", 0.9},
		{"*", "*", 0.1},
	}

	for i, exp := range expected {
		mr := ranges[i]
		if mr.Type != exp.typ || mr.Subtype != exp.subtype {
			t.Errorf("Range %d: expected %s/%s, got %s/%s", i, exp.typ, exp.subtype, mr.Type, mr.Subtype)
		}
		if mr.Quality != exp.quality {
			t.Errorf("Range %d: expected quality %f, got %f", i, exp.quality, mr.Quality)
		}
	}
}

func TestParseAccept_SpecificitySorting(t *testing.T) {
	// When quality is same, specificity matters: explicit > type wildcard > full wildcard.
	ranges := parseAccept("*/*;q=0.5, text/*;q=0.5, text/html;q=0.5")

	if len(ranges) != 3 {
		t.Fatalf("Expected 3 ranges, got %d", len(ranges))
	}

	// Should be sorted: text/html (specific) > text/* (type wildcard) > */* (full wildcard).
	expected := []string{"text/html", "text/*", "*/*"}

	for i, exp := range expected {
		got := ranges[i].Type + "/" + ranges[i].Subtype
		if got != exp {
			t.Errorf("Range %d: expected %s, got %s", i, exp, got)
		}
	}
}

// Test ContentType function.

func TestContentType_ExactMatch(t *testing.T) {
	offered := []string{"application/json", "text/html", "application/xml"}
	accept := "application/json"

	result := ContentType(accept, offered)

	if result != "application/json" {
		t.Errorf("Expected application/json, got %s", result)
	}
}

func TestContentType_QWeighting(t *testing.T) {
	offered := []string{"application/json", "application/xml", "text/html"}
	accept := "text/html, application/json;q=0.9, application/xml;q=0.8"

	result := ContentType(accept, offered)

	// text/html has highest quality (implicit q=1.0).
	if result != "text/html" {
		t.Errorf("Expected text/html, got %s", result)
	}
}

func TestContentType_WildcardMatch(t *testing.T) {
	offered := []string{"application/json", "text/html"}
	accept := "*/*"

	result := ContentType(accept, offered)

	// Should match first offered (application/json).
	if result != "application/json" {
		t.Errorf("Expected application/json (first offered), got %s", result)
	}
}

func TestContentType_TypeWildcard(t *testing.T) {
	offered := []string{"application/json", "text/html", "text/plain"}
	accept := "text/*"

	result := ContentType(accept, offered)

	// Should match first text/* type (text/html).
	if result != "text/html" {
		t.Errorf("Expected text/html, got %s", result)
	}
}

func TestContentType_NoMatch(t *testing.T) {
	offered := []string{"application/json", "application/xml"}
	accept := "text/html"

	result := ContentType(accept, offered)

	if result != "" {
		t.Errorf("Expected empty string (no match), got %s", result)
	}
}

func TestContentType_EmptyOffered(t *testing.T) {
	offered := []string{}
	accept := "application/json"

	result := ContentType(accept, offered)

	if result != "" {
		t.Errorf("Expected empty string (no offered types), got %s", result)
	}
}

func TestContentType_EmptyAccept(t *testing.T) {
	offered := []string{"application/json", "text/html"}
	accept := ""

	result := ContentType(accept, offered)

	// Empty Accept header should match */* which matches first offered.
	if result != "application/json" {
		t.Errorf("Expected application/json (first offered), got %s", result)
	}
}
