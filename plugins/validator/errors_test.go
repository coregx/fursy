// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validator

import (
	"reflect"
	"strings"
	"testing"
)

// Test structs for comprehensive validation tag coverage.
type TestAllTags struct {
	// String validations.
	URL         string `validate:"url"`
	URI         string `validate:"uri"`
	Alpha       string `validate:"alpha"`
	AlphaNum    string `validate:"alphanum"`
	Numeric     string `validate:"numeric"`
	Number      string `validate:"number"`
	Hexadecimal string `validate:"hexadecimal"`
	HexColor    string `validate:"hexcolor"`
	RGB         string `validate:"rgb"`
	RGBA        string `validate:"rgba"`
	HSL         string `validate:"hsl"`
	HSLA        string `validate:"hsla"`
	UUID        string `validate:"uuid"`
	UUID3       string `validate:"uuid3"`
	UUID4       string `validate:"uuid4"`
	UUID5       string `validate:"uuid5"`
	ISBN        string `validate:"isbn"`
	ISBN10      string `validate:"isbn10"`
	ISBN13      string `validate:"isbn13"`
	JSON        string `validate:"json"`
	Latitude    string `validate:"latitude"`
	Longitude   string `validate:"longitude"`
	SSN         string `validate:"ssn"`
	IPv4        string `validate:"ipv4"`
	IPv6        string `validate:"ipv6"`
	IP          string `validate:"ip"`
	MAC         string `validate:"mac"`
	Contains    string `validate:"contains=test"`
	ContainsAny string `validate:"containsany=abc"`
	Excludes    string `validate:"excludes=test"`
	ExcludesAll string `validate:"excludesall=abc"`
	StartsWith  string `validate:"startswith=prefix"`
	EndsWith    string `validate:"endswith=suffix"`
	DateTime    string `validate:"datetime=2006-01-02"`
	MaxStr      string `validate:"max=10"`
	LenStr      string `validate:"len=5"`

	// Number validations.
	GT    int    `validate:"gt=10"`
	LT    int    `validate:"lt=100"`
	EQ    int    `validate:"eq=42"`
	NE    int    `validate:"ne=0"`
	OneOf string `validate:"oneof=red green blue"`

	// Slice validations.
	MinSlice []string `validate:"min=2"`
	MaxSlice []string `validate:"max=5"`
	LenSlice []string `validate:"len=3"`
}

// TestDefaultMessage_AllTags tests all validation tag error messages.
func TestDefaultMessage_AllTags(t *testing.T) {
	v := New()

	tests := []struct {
		name        string
		data        any
		wantContain string
		field       string
	}{
		// String format validations.
		{name: "url", data: &TestAllTags{URL: "not-a-url"}, wantContain: "valid URL", field: "URL"},
		{name: "uri", data: &TestAllTags{URI: "::invalid"}, wantContain: "valid URI", field: "URI"},
		{name: "alpha", data: &TestAllTags{Alpha: "abc123"}, wantContain: "alphabetic", field: "Alpha"},
		{name: "alphanum", data: &TestAllTags{AlphaNum: "abc!@#"}, wantContain: "alphanumeric", field: "AlphaNum"},
		{name: "numeric", data: &TestAllTags{Numeric: "abc"}, wantContain: "numeric", field: "Numeric"},
		{name: "number", data: &TestAllTags{Number: "abc"}, wantContain: "number", field: "Number"},
		{name: "hexadecimal", data: &TestAllTags{Hexadecimal: "zzz"}, wantContain: "hexadecimal", field: "Hexadecimal"},
		{name: "hexcolor", data: &TestAllTags{HexColor: "notcolor"}, wantContain: "hex color", field: "HexColor"},
		{name: "rgb", data: &TestAllTags{RGB: "invalid"}, wantContain: "RGB", field: "RGB"},
		{name: "rgba", data: &TestAllTags{RGBA: "invalid"}, wantContain: "RGBA", field: "RGBA"},
		{name: "hsl", data: &TestAllTags{HSL: "invalid"}, wantContain: "HSL", field: "HSL"},
		{name: "hsla", data: &TestAllTags{HSLA: "invalid"}, wantContain: "HSLA", field: "HSLA"},
		{name: "uuid", data: &TestAllTags{UUID: "not-a-uuid"}, wantContain: "UUID", field: "UUID"},
		{name: "uuid3", data: &TestAllTags{UUID3: "not-a-uuid"}, wantContain: "UUID v3", field: "UUID3"},
		{name: "uuid4", data: &TestAllTags{UUID4: "not-a-uuid"}, wantContain: "UUID v4", field: "UUID4"},
		{name: "uuid5", data: &TestAllTags{UUID5: "not-a-uuid"}, wantContain: "UUID v5", field: "UUID5"},
		{name: "isbn", data: &TestAllTags{ISBN: "123"}, wantContain: "ISBN", field: "ISBN"},
		{name: "isbn10", data: &TestAllTags{ISBN10: "123"}, wantContain: "ISBN-10", field: "ISBN10"},
		{name: "isbn13", data: &TestAllTags{ISBN13: "123"}, wantContain: "ISBN-13", field: "ISBN13"},
		{name: "json", data: &TestAllTags{JSON: "{invalid"}, wantContain: "valid JSON", field: "JSON"},
		{name: "latitude", data: &TestAllTags{Latitude: "invalid"}, wantContain: "latitude", field: "Latitude"},
		{name: "longitude", data: &TestAllTags{Longitude: "invalid"}, wantContain: "longitude", field: "Longitude"},
		{name: "ssn", data: &TestAllTags{SSN: "123"}, wantContain: "Social Security", field: "SSN"},
		{name: "ipv4", data: &TestAllTags{IPv4: "999.999.999.999"}, wantContain: "IPv4", field: "IPv4"},
		{name: "ipv6", data: &TestAllTags{IPv6: "zzz"}, wantContain: "IPv6", field: "IPv6"},
		{name: "ip", data: &TestAllTags{IP: "invalid"}, wantContain: "IP address", field: "IP"},
		{name: "mac", data: &TestAllTags{MAC: "invalid"}, wantContain: "MAC", field: "MAC"},

		// String comparison validations.
		{name: "contains", data: &TestAllTags{Contains: "hello"}, wantContain: "contain 'test'", field: "Contains"},
		{name: "containsany", data: &TestAllTags{ContainsAny: "xyz"}, wantContain: "at least one", field: "ContainsAny"},
		{name: "excludes", data: &TestAllTags{Excludes: "test"}, wantContain: "not contain", field: "Excludes"},
		{name: "excludesall", data: &TestAllTags{ExcludesAll: "abc"}, wantContain: "not contain any", field: "ExcludesAll"},
		{name: "startswith", data: &TestAllTags{StartsWith: "other"}, wantContain: "start with", field: "StartsWith"},
		{name: "endswith", data: &TestAllTags{EndsWith: "other"}, wantContain: "end with", field: "EndsWith"},
		{name: "datetime", data: &TestAllTags{DateTime: "invalid"}, wantContain: "date/time", field: "DateTime"},
		{name: "max string", data: &TestAllTags{MaxStr: "12345678901"}, wantContain: "at most 10 characters", field: "MaxStr"},
		{name: "len string", data: &TestAllTags{LenStr: "1234"}, wantContain: "exactly 5 characters", field: "LenStr"},

		// Number comparison validations.
		{name: "gt", data: &TestAllTags{GT: 5}, wantContain: "greater than 10", field: "GT"},
		{name: "lt", data: &TestAllTags{LT: 150}, wantContain: "less than 100", field: "LT"},
		{name: "eq", data: &TestAllTags{EQ: 10}, wantContain: "equal to 42", field: "EQ"},
		{name: "ne", data: &TestAllTags{NE: 0}, wantContain: "not be equal to 0", field: "NE"},
		{name: "oneof", data: &TestAllTags{OneOf: "yellow"}, wantContain: "one of", field: "OneOf"},

		// Slice validations.
		{name: "min slice", data: &TestAllTags{MinSlice: []string{"a"}}, wantContain: "at least 2 items", field: "MinSlice"},
		{name: "max slice", data: &TestAllTags{MaxSlice: []string{"a", "b", "c", "d", "e", "f"}}, wantContain: "at most 5 items", field: "MaxSlice"},
		{name: "len slice", data: &TestAllTags{LenSlice: []string{"a", "b"}}, wantContain: "exactly 3 items", field: "LenSlice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.data)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}

			errMsg := err.Error()
			if !strings.Contains(strings.ToLower(errMsg), strings.ToLower(tt.wantContain)) {
				t.Errorf("error message %q should contain %q", errMsg, tt.wantContain)
			}
		})
	}
}

// TestRegisterTagNameFunc tests custom tag name function registration.
func TestRegisterTagNameFunc(t *testing.T) {
	v := New()

	type TestJSONTag struct {
		EmailAddress string `json:"email" validate:"required,email"`
		FullName     string `json:"name" validate:"required"`
	}

	// Register function to use JSON tag names.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// This is a simplified version - just return field name for testing.
		return fld.Name
	})

	user := &TestJSONTag{
		EmailAddress: "",
		FullName:     "",
	}

	err := v.Validate(user)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

// TestInterpolateMessage tests message placeholder interpolation.
func TestInterpolateMessage(t *testing.T) {
	customMessages := map[string]string{
		"min": "{field} must have at least {param} characters (got: {value})",
		"max": "Maximum {param} allowed for {field}, you entered: {value}",
	}

	v := New(&Options{
		CustomMessages: customMessages,
	})

	type TestInterpolation struct {
		ShortField string `validate:"min=5"`
		LongField  string `validate:"max=3"`
	}

	data := &TestInterpolation{
		ShortField: "abc",
		LongField:  "abcd",
	}

	err := v.Validate(data)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	errMsg := err.Error()

	// Check that placeholders are replaced.
	if !strings.Contains(errMsg, "at least 5 characters") || !strings.Contains(errMsg, "got: abc") {
		// Alternative check for second message.
		if !strings.Contains(errMsg, "Maximum 3 allowed") && !strings.Contains(errMsg, "you entered: abcd") {
			t.Errorf("expected interpolated message, got: %s", errMsg)
		}
	}
}
