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
