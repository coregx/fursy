// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package binding

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Test structs for binding.
type BindTestStruct struct {
	Name  string `json:"name" xml:"name" form:"name"`
	Email string `json:"email" xml:"email" form:"email"`
	Age   int    `json:"age" xml:"age" form:"age"`
}

// TestJSONBinder tests JSON binding.
func TestJSONBinder(t *testing.T) {
	body := `{"name":"John","email":"john@example.com","age":30}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	var result BindTestStruct
	if err := jsonBinding.Bind(req, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "John" {
		t.Errorf("Name = %q, want %q", result.Name, "John")
	}
	if result.Email != "john@example.com" {
		t.Errorf("Email = %q, want %q", result.Email, "john@example.com")
	}
	if result.Age != 30 {
		t.Errorf("Age = %d, want %d", result.Age, 30)
	}
}

// TestJSONBinder_EmptyBody tests JSON binding with empty body.
func TestJSONBinder_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
	req.Header.Set("Content-Type", "application/json")

	var result BindTestStruct
	err := jsonBinding.Bind(req, &result)
	if !errors.Is(err, ErrEmptyRequestBody) {
		t.Errorf("expected ErrEmptyRequestBody, got %v", err)
	}
}

// TestJSONBinder_InvalidJSON tests JSON binding with invalid JSON.
func TestJSONBinder_InvalidJSON(t *testing.T) {
	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	var result BindTestStruct
	err := jsonBinding.Bind(req, &result)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestXMLBinder tests XML binding.
func TestXMLBinder(t *testing.T) {
	body := `<BindTestStruct><name>Jane</name><email>jane@example.com</email><age>25</age></BindTestStruct>`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/xml")

	var result BindTestStruct
	if err := xmlBinding.Bind(req, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Jane" {
		t.Errorf("Name = %q, want %q", result.Name, "Jane")
	}
	if result.Email != "jane@example.com" {
		t.Errorf("Email = %q, want %q", result.Email, "jane@example.com")
	}
	if result.Age != 25 {
		t.Errorf("Age = %d, want %d", result.Age, 25)
	}
}

// TestXMLBinder_EmptyBody tests XML binding with empty body.
func TestXMLBinder_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
	req.Header.Set("Content-Type", "application/xml")

	var result BindTestStruct
	err := xmlBinding.Bind(req, &result)
	if !errors.Is(err, ErrEmptyRequestBody) {
		t.Errorf("expected ErrEmptyRequestBody, got %v", err)
	}
}

// TestFormBinder tests form binding.
func TestFormBinder(t *testing.T) {
	form := url.Values{}
	form.Set("name", "Bob")
	form.Set("email", "bob@example.com")
	form.Set("age", "35")

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result BindTestStruct
	if err := formBinding.Bind(req, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Bob" {
		t.Errorf("Name = %q, want %q", result.Name, "Bob")
	}
	if result.Email != "bob@example.com" {
		t.Errorf("Email = %q, want %q", result.Email, "bob@example.com")
	}
	if result.Age != 35 {
		t.Errorf("Age = %d, want %d", result.Age, 35)
	}
}

// TestFormBinder_EmptyForm tests form binding with empty form.
func TestFormBinder_EmptyForm(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result BindTestStruct
	err := formBinding.Bind(req, &result)
	if !errors.Is(err, ErrEmptyRequestBody) {
		t.Errorf("expected ErrEmptyRequestBody, got %v", err)
	}
}

// TestMultipartBinder tests multipart form binding.
func TestMultipartBinder(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	writer.WriteField("name", "Alice")
	writer.WriteField("email", "alice@example.com")
	writer.WriteField("age", "28")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/test", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var result BindTestStruct
	if err := multipartBinding.Bind(req, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "Alice" {
		t.Errorf("Name = %q, want %q", result.Name, "Alice")
	}
	if result.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", result.Email, "alice@example.com")
	}
	if result.Age != 28 {
		t.Errorf("Age = %d, want %d", result.Age, 28)
	}
}

// TestGetBinder tests binder selection based on Content-Type.
func TestGetBinder(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		wantErr     bool
		wantType    Binder
	}{
		{"empty defaults to JSON", "", false, jsonBinding},
		{"application/json", "application/json", false, jsonBinding},
		{"application/json with charset", "application/json; charset=utf-8", false, jsonBinding},
		{"application/xml", "application/xml", false, xmlBinding},
		{"text/xml", "text/xml", false, xmlBinding},
		{"application/x-www-form-urlencoded", "application/x-www-form-urlencoded", false, formBinding},
		{"multipart/form-data", "multipart/form-data", false, multipartBinding},
		{"multipart/form-data with boundary", "multipart/form-data; boundary=----WebKitFormBoundary", false, multipartBinding},
		{"unsupported type", "application/pdf", true, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binder, err := GetBinder(tt.contentType)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if binder != tt.wantType {
				t.Errorf("got binder type %T, want %T", binder, tt.wantType)
			}
		})
	}
}

// TestXMLBinder_InvalidXML tests XML binding with invalid XML.
func TestXMLBinder_InvalidXML(t *testing.T) {
	body := `<invalid xml`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/xml")

	var result BindTestStruct
	err := xmlBinding.Bind(req, &result)
	if err == nil {
		t.Error("expected error for invalid XML, got nil")
	}
}

// TestXMLBinder_NilBody tests XML binding with nil body.
func TestXMLBinder_NilBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
	req.Body = nil
	req.ContentLength = 0
	req.Header.Set("Content-Type", "application/xml")

	var result BindTestStruct
	err := xmlBinding.Bind(req, &result)
	if !errors.Is(err, ErrEmptyRequestBody) {
		t.Errorf("expected ErrEmptyRequestBody, got %v", err)
	}
}

// TestJSONBinder_NilBody tests JSON binding with nil body.
func TestJSONBinder_NilBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", http.NoBody)
	req.Body = nil
	req.ContentLength = 0
	req.Header.Set("Content-Type", "application/json")

	var result BindTestStruct
	err := jsonBinding.Bind(req, &result)
	if !errors.Is(err, ErrEmptyRequestBody) {
		t.Errorf("expected ErrEmptyRequestBody, got %v", err)
	}
}

// TestMultipartBinder_EmptyForm tests multipart binding with empty multipart form.
func TestMultipartBinder_EmptyForm(t *testing.T) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/test", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var result BindTestStruct
	err := multipartBinding.Bind(req, &result)
	if !errors.Is(err, ErrEmptyRequestBody) {
		t.Errorf("expected ErrEmptyRequestBody, got %v", err)
	}
}

// TestMultipartBinder_InvalidContentType tests multipart binding with invalid content type.
func TestMultipartBinder_InvalidContentType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("not multipart"))
	req.Header.Set("Content-Type", "multipart/form-data") // Missing boundary

	var result BindTestStruct
	err := multipartBinding.Bind(req, &result)
	if err == nil {
		t.Error("expected error for missing boundary, got nil")
	}
}

// BindAllTypesStruct is used to test setField for all supported types.
type BindAllTypesStruct struct {
	Str      string  `form:"str"`
	Int      int     `form:"int"`
	Int8     int8    `form:"int8"`
	Int16    int16   `form:"int16"`
	Int32    int32   `form:"int32"`
	Int64    int64   `form:"int64"`
	Uint     uint    `form:"uint"`
	Uint8    uint8   `form:"uint8"`
	Uint16   uint16  `form:"uint16"`
	Uint32   uint32  `form:"uint32"`
	Uint64   uint64  `form:"uint64"`
	Float32  float32 `form:"float32"`
	Float64  float64 `form:"float64"`
	Bool     bool    `form:"bool"`
	Ignored  string  `form:"-"`
	NoTag    string
}

// TestSetField_AllTypes tests setField with every supported field type.
func TestSetField_AllTypes(t *testing.T) {
	form := url.Values{}
	form.Set("str", "hello")
	form.Set("int", "-42")
	form.Set("int8", "-8")
	form.Set("int16", "-16")
	form.Set("int32", "-32")
	form.Set("int64", "-64")
	form.Set("uint", "42")
	form.Set("uint8", "8")
	form.Set("uint16", "16")
	form.Set("uint32", "32")
	form.Set("uint64", "64")
	form.Set("float32", "3.14")
	form.Set("float64", "2.718281828")
	form.Set("bool", "true")
	form.Set("Ignored", "should be ignored")
	form.Set("NoTag", "no tag value")

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result BindAllTypesStruct
	if err := formBinding.Bind(req, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Str != "hello" {
		t.Errorf("Str = %q, want %q", result.Str, "hello")
	}
	if result.Int != -42 {
		t.Errorf("Int = %d, want %d", result.Int, -42)
	}
	if result.Int8 != -8 {
		t.Errorf("Int8 = %d, want %d", result.Int8, -8)
	}
	if result.Int16 != -16 {
		t.Errorf("Int16 = %d, want %d", result.Int16, -16)
	}
	if result.Int32 != -32 {
		t.Errorf("Int32 = %d, want %d", result.Int32, -32)
	}
	if result.Int64 != -64 {
		t.Errorf("Int64 = %d, want %d", result.Int64, -64)
	}
	if result.Uint != 42 {
		t.Errorf("Uint = %d, want %d", result.Uint, 42)
	}
	if result.Uint8 != 8 {
		t.Errorf("Uint8 = %d, want %d", result.Uint8, 8)
	}
	if result.Uint16 != 16 {
		t.Errorf("Uint16 = %d, want %d", result.Uint16, 16)
	}
	if result.Uint32 != 32 {
		t.Errorf("Uint32 = %d, want %d", result.Uint32, 32)
	}
	if result.Uint64 != 64 {
		t.Errorf("Uint64 = %d, want %d", result.Uint64, 64)
	}
	if result.Float32 < 3.13 || result.Float32 > 3.15 {
		t.Errorf("Float32 = %f, want ~3.14", result.Float32)
	}
	if result.Float64 < 2.71 || result.Float64 > 2.72 {
		t.Errorf("Float64 = %f, want ~2.718", result.Float64)
	}
	if !result.Bool {
		t.Error("Bool = false, want true")
	}
	if result.Ignored != "" {
		t.Errorf("Ignored = %q, want empty (tag is -)", result.Ignored)
	}
	if result.NoTag != "no tag value" {
		t.Errorf("NoTag = %q, want %q", result.NoTag, "no tag value")
	}
}

// TestSetField_InvalidValues tests setField error handling for invalid type conversions.
func TestSetField_InvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		form  url.Values
		field string
	}{
		{
			name:  "invalid int",
			form:  url.Values{"int": {"not_a_number"}},
			field: "Int",
		},
		{
			name:  "invalid uint",
			form:  url.Values{"uint": {"not_a_number"}},
			field: "Uint",
		},
		{
			name:  "invalid float",
			form:  url.Values{"float64": {"not_a_float"}},
			field: "Float64",
		},
		{
			name:  "invalid bool",
			form:  url.Values{"bool": {"not_a_bool"}},
			field: "Bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			var result BindAllTypesStruct
			err := formBinding.Bind(req, &result)
			if err == nil {
				t.Error("expected error for invalid value, got nil")
			}
		})
	}
}

// TestSetField_UnsupportedType tests setField with an unsupported field type.
func TestSetField_UnsupportedType(t *testing.T) {
	type withComplex struct {
		Data complex128 `form:"data"`
	}

	form := url.Values{"data": {"1+2i"}}
	var result withComplex
	err := mapForm(&result, form)
	if err == nil {
		t.Error("expected error for unsupported field type, got nil")
	}
}

// TestMapForm_NonPointer tests mapForm with non-pointer argument.
func TestMapForm_NonPointer(t *testing.T) {
	form := url.Values{"name": {"test"}}
	var s BindTestStruct
	err := mapForm(s, form)
	if err == nil {
		t.Error("expected error for non-pointer, got nil")
	}
}

// TestMapForm_NonStruct tests mapForm with pointer to non-struct.
func TestMapForm_NonStruct(t *testing.T) {
	form := url.Values{"name": {"test"}}
	s := "not a struct"
	err := mapForm(&s, form)
	if err == nil {
		t.Error("expected error for non-struct pointer, got nil")
	}
}

// TestMapForm_MissingFormValues tests mapForm when form values are missing.
func TestMapForm_MissingFormValues(t *testing.T) {
	form := url.Values{} // No values
	var result BindTestStruct
	err := mapForm(&result, form)
	if err != nil {
		t.Errorf("expected nil error for missing values, got %v", err)
	}
	// All fields should have zero values.
	if result.Name != "" {
		t.Errorf("Name = %q, want empty", result.Name)
	}
	if result.Age != 0 {
		t.Errorf("Age = %d, want 0", result.Age)
	}
}

// TestMapForm_UnexportedFields tests that unexported fields are skipped.
func TestMapForm_UnexportedFields(t *testing.T) {
	type withUnexported struct {
		Name   string `form:"name"`
		hidden string `form:"hidden"` //nolint:unused // testing unexported field behavior
	}

	form := url.Values{
		"name":   {"visible"},
		"hidden": {"should be ignored"},
	}
	var result withUnexported
	err := mapForm(&result, form)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Name != "visible" {
		t.Errorf("Name = %q, want %q", result.Name, "visible")
	}
}

// TestFormBinder_InvalidForm tests form binding with invalid request body.
func TestFormBinder_InvalidForm(_ *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Force ContentLength to indicate there is a body.
	var result BindTestStruct
	err := formBinding.Bind(req, &result)
	// ParseForm may or may not error on partially invalid URL-encoded data.
	// The important thing is it doesn't panic.
	_ = err
}

// TestBind_Integration tests the Bind convenience function.
func TestBind_Integration(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantName    string
		wantEmail   string
		wantAge     int
		wantErr     bool
	}{
		{
			name:        "JSON",
			contentType: "application/json",
			body:        `{"name":"Test","email":"test@example.com","age":20}`,
			wantName:    "Test",
			wantEmail:   "test@example.com",
			wantAge:     20,
		},
		{
			name:        "XML",
			contentType: "application/xml",
			body:        `<BindTestStruct><name>Test</name><email>test@example.com</email><age>20</age></BindTestStruct>`,
			wantName:    "Test",
			wantEmail:   "test@example.com",
			wantAge:     20,
		},
		{
			name:        "unsupported",
			contentType: "text/plain",
			body:        "plain text",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			var result BindTestStruct
			err := Bind(req, &result)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", result.Name, tt.wantName)
			}
			if result.Email != tt.wantEmail {
				t.Errorf("Email = %q, want %q", result.Email, tt.wantEmail)
			}
			if result.Age != tt.wantAge {
				t.Errorf("Age = %d, want %d", result.Age, tt.wantAge)
			}
		})
	}
}
