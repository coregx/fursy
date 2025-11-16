// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewProblem tests creating a new Problem.
func TestNewProblem(t *testing.T) {
	p := NewProblem(404, "Not Found", "The requested resource was not found")

	if p.Type != "about:blank" {
		t.Errorf("Type = %q, want %q", p.Type, "about:blank")
	}
	if p.Title != "Not Found" {
		t.Errorf("Title = %q, want %q", p.Title, "Not Found")
	}
	if p.Status != 404 {
		t.Errorf("Status = %d, want %d", p.Status, 404)
	}
	if p.Detail != "The requested resource was not found" {
		t.Errorf("Detail = %q, want %q", p.Detail, "The requested resource was not found")
	}
	if p.Instance != "" {
		t.Errorf("Instance = %q, want empty", p.Instance)
	}
}

// TestProblem_Error tests Problem.Error() implementation.
func TestProblem_Error(t *testing.T) {
	tests := []struct {
		name    string
		problem Problem
		want    string
	}{
		{
			name: "with detail",
			problem: Problem{
				Title:  "Not Found",
				Detail: "User not found",
			},
			want: "User not found",
		},
		{
			name: "without detail",
			problem: Problem{
				Title: "Internal Server Error",
			},
			want: "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.problem.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestProblem_WithType tests WithType fluent API.
func TestProblem_WithType(t *testing.T) {
	p := NewProblem(400, "Bad Request", "Invalid input").
		WithType("https://example.com/probs/invalid-input")

	if p.Type != "https://example.com/probs/invalid-input" {
		t.Errorf("Type = %q, want %q", p.Type, "https://example.com/probs/invalid-input")
	}
}

// TestProblem_WithInstance tests WithInstance fluent API.
func TestProblem_WithInstance(t *testing.T) {
	p := NewProblem(404, "Not Found", "User not found").
		WithInstance("/users/123")

	if p.Instance != "/users/123" {
		t.Errorf("Instance = %q, want %q", p.Instance, "/users/123")
	}
}

// TestProblem_WithExtension tests WithExtension fluent API.
func TestProblem_WithExtension(t *testing.T) {
	p := NewProblem(403, "Forbidden", "Insufficient balance").
		WithExtension("balance", 30).
		WithExtension("cost", 50)

	if len(p.Extensions) != 2 {
		t.Errorf("Extensions length = %d, want 2", len(p.Extensions))
	}
	if p.Extensions["balance"] != 30 {
		t.Errorf("Extensions[balance] = %v, want 30", p.Extensions["balance"])
	}
	if p.Extensions["cost"] != 50 {
		t.Errorf("Extensions[cost] = %v, want 50", p.Extensions["cost"])
	}
}

// TestProblem_WithExtensions tests WithExtensions fluent API.
func TestProblem_WithExtensions(t *testing.T) {
	p := NewProblem(400, "Bad Request", "Validation failed").
		WithExtensions(map[string]any{
			"field":  "email",
			"reason": "invalid format",
		})

	if len(p.Extensions) != 2 {
		t.Errorf("Extensions length = %d, want 2", len(p.Extensions))
	}
	if p.Extensions["field"] != "email" {
		t.Errorf("Extensions[field] = %v, want email", p.Extensions["field"])
	}
}

// TestProblem_MarshalJSON tests JSON marshaling.
func TestProblem_MarshalJSON(t *testing.T) {
	p := Problem{
		Type:     "https://example.com/probs/out-of-credit",
		Title:    "You do not have enough credit",
		Status:   403,
		Detail:   "Your current balance is 30, but that costs 50",
		Instance: "/account/12345/msgs/abc",
		Extensions: map[string]any{
			"balance": 30,
			"cost":    50,
		},
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Check standard fields.
	if result["type"] != "https://example.com/probs/out-of-credit" {
		t.Errorf("type = %v, want https://example.com/probs/out-of-credit", result["type"])
	}
	if result["title"] != "You do not have enough credit" {
		t.Errorf("title = %v, want You do not have enough credit", result["title"])
	}
	if result["status"] != float64(403) {
		t.Errorf("status = %v, want 403", result["status"])
	}
	if result["detail"] != "Your current balance is 30, but that costs 50" {
		t.Errorf("detail = %v, want Your current balance is 30, but that costs 50", result["detail"])
	}
	if result["instance"] != "/account/12345/msgs/abc" {
		t.Errorf("instance = %v, want /account/12345/msgs/abc", result["instance"])
	}

	// Check flattened extensions.
	if result["balance"] != float64(30) {
		t.Errorf("balance = %v, want 30", result["balance"])
	}
	if result["cost"] != float64(50) {
		t.Errorf("cost = %v, want 50", result["cost"])
	}
}

// TestProblem_MarshalJSON_OmitEmpty tests that detail and instance are omitted when empty.
func TestProblem_MarshalJSON_OmitEmpty(t *testing.T) {
	p := Problem{
		Type:   "about:blank",
		Title:  "Not Found",
		Status: 404,
		// Detail and Instance are empty.
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, ok := result["detail"]; ok {
		t.Error("detail should be omitted when empty")
	}
	if _, ok := result["instance"]; ok {
		t.Error("instance should be omitted when empty")
	}
}

// TestProblem_MarshalJSON_NoStandardFieldOverwrite tests that extensions don't overwrite standard fields.
func TestProblem_MarshalJSON_NoStandardFieldOverwrite(t *testing.T) {
	p := Problem{
		Type:   "https://example.com/probs/test",
		Title:  "Test Problem",
		Status: 400,
		Extensions: map[string]any{
			"type":   "should not overwrite",
			"title":  "should not overwrite",
			"status": 999,
			"custom": "this is ok",
		},
	}

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Standard fields should NOT be overwritten.
	if result["type"] != "https://example.com/probs/test" {
		t.Errorf("type was overwritten: %v", result["type"])
	}
	if result["title"] != "Test Problem" {
		t.Errorf("title was overwritten: %v", result["title"])
	}
	if result["status"] != float64(400) {
		t.Errorf("status was overwritten: %v", result["status"])
	}

	// Custom extension should be present.
	if result["custom"] != "this is ok" {
		t.Errorf("custom = %v, want this is ok", result["custom"])
	}
}

// TestStandardProblems tests all standard problem constructors.
func TestStandardProblems(t *testing.T) {
	tests := []struct {
		name   string
		create func(string) Problem
		status int
		title  string
	}{
		{"BadRequest", BadRequest, 400, "Bad Request"},
		{"Unauthorized", Unauthorized, 401, "Unauthorized"},
		{"Forbidden", Forbidden, 403, "Forbidden"},
		{"NotFound", NotFound, 404, "Not Found"},
		{"MethodNotAllowed", MethodNotAllowed, 405, "Method Not Allowed"},
		{"Conflict", Conflict, 409, "Conflict"},
		{"UnprocessableEntity", UnprocessableEntity, 422, "Unprocessable Entity"},
		{"TooManyRequests", TooManyRequests, 429, "Too Many Requests"},
		{"InternalServerError", InternalServerError, 500, "Internal Server Error"},
		{"ServiceUnavailable", ServiceUnavailable, 503, "Service Unavailable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.create("test detail")

			if p.Status != tt.status {
				t.Errorf("Status = %d, want %d", p.Status, tt.status)
			}
			if p.Title != tt.title {
				t.Errorf("Title = %q, want %q", p.Title, tt.title)
			}
			if p.Detail != "test detail" {
				t.Errorf("Detail = %q, want %q", p.Detail, "test detail")
			}
			if p.Type != "about:blank" {
				t.Errorf("Type = %q, want %q", p.Type, "about:blank")
			}
		})
	}
}

// TestValidationProblem tests ValidationProblem constructor.
func TestValidationProblem(t *testing.T) {
	errs := ValidationErrors{
		{Field: "email", Tag: "required", Message: "email is required"},
		{Field: "age", Tag: "min", Message: "age must be at least 18"},
	}

	p := ValidationProblem(errs)

	if p.Status != 422 {
		t.Errorf("Status = %d, want 422", p.Status)
	}
	if p.Title != "Validation Failed" {
		t.Errorf("Title = %q, want %q", p.Title, "Validation Failed")
	}

	// Check extensions contain errors field.
	errorsField, ok := p.Extensions["errors"].(map[string]string)
	if !ok {
		t.Fatalf("Extensions[errors] should be map[string]string, got %T", p.Extensions["errors"])
	}

	if len(errorsField) != 2 {
		t.Errorf("errors length = %d, want 2", len(errorsField))
	}
	if errorsField["email"] != "email is required" {
		t.Errorf("errors[email] = %q, want %q", errorsField["email"], "email is required")
	}
	if errorsField["age"] != "age must be at least 18" {
		t.Errorf("errors[age] = %q, want %q", errorsField["age"], "age must be at least 18")
	}
}

// TestValidationProblem_SingleError tests ValidationProblem with single error.
func TestValidationProblem_SingleError(t *testing.T) {
	errs := ValidationErrors{
		{Field: "email", Tag: "required", Message: "email is required"},
	}

	p := ValidationProblem(errs)

	// Detail should be the single error message.
	if p.Detail != "email is required" {
		t.Errorf("Detail = %q, want %q", p.Detail, "email is required")
	}
}

// TestValidationProblem_Empty tests ValidationProblem with empty errors.
func TestValidationProblem_Empty(t *testing.T) {
	errs := ValidationErrors{}

	p := ValidationProblem(errs)

	if p.Detail != "validation failed" {
		t.Errorf("Detail = %q, want %q", p.Detail, "validation failed")
	}
}

// TestContext_Problem tests Box.Problem() method.
func TestContext_Problem(t *testing.T) {
	r := New()

	r.GET("/test", func(c *Context) error {
		return c.Problem(NotFound("Resource not found"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check status code.
	if w.Code != 404 {
		t.Errorf("Status = %d, want 404", w.Code)
	}

	// Check Content-Type.
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/problem+json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/problem+json; charset=utf-8")
	}

	// Check response body.
	var result Problem
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if result.Status != 404 {
		t.Errorf("Status = %d, want 404", result.Status)
	}
	if result.Title != "Not Found" {
		t.Errorf("Title = %q, want %q", result.Title, "Not Found")
	}
	if result.Detail != "Resource not found" {
		t.Errorf("Detail = %q, want %q", result.Detail, "Resource not found")
	}
}

// TestContext_Problem_WithExtensions tests Problem with extensions.
func TestContext_Problem_WithExtensions(t *testing.T) {
	r := New()

	r.GET("/test", func(c *Context) error {
		return c.Problem(
			Forbidden("Insufficient balance").
				WithExtension("balance", 30).
				WithExtension("required", 50),
		)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if result["balance"] != float64(30) {
		t.Errorf("balance = %v, want 30", result["balance"])
	}
	if result["required"] != float64(50) {
		t.Errorf("required = %v, want 50", result["required"])
	}
}

// TestContext_Problem_ValidationErrors tests Problem with validation errors.
func TestContext_Problem_ValidationErrors(t *testing.T) {
	r := New()

	r.GET("/test", func(c *Context) error {
		errs := ValidationErrors{
			{Field: "email", Tag: "email", Message: "must be a valid email"},
			{Field: "age", Tag: "min", Message: "must be at least 18"},
		}
		return c.Problem(ValidationProblem(errs))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 422 {
		t.Errorf("Status = %d, want 422", w.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Check errors field.
	errorsField, ok := result["errors"].(map[string]any)
	if !ok {
		t.Fatalf("errors should be map[string]any, got %T", result["errors"])
	}

	if errorsField["email"] != "must be a valid email" {
		t.Errorf("errors[email] = %v, want must be a valid email", errorsField["email"])
	}
	if errorsField["age"] != "must be at least 18" {
		t.Errorf("errors[age] = %v, want must be at least 18", errorsField["age"])
	}
}

// TestProblem_Integration tests full RFC 9457 compliance.
func TestProblem_Integration(t *testing.T) {
	r := New()

	r.GET("/users/:id", func(c *Context) error {
		id := c.Param("id")
		if id != "123" {
			return c.Problem(
				NotFound("User not found").
					WithType("https://api.example.com/errors/user-not-found").
					WithInstance("/users/"+id).
					WithExtension("user_id", id),
			)
		}
		return c.String(200, "User found")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/999", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Check all RFC 9457 fields.
	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expectedFields := map[string]any{
		"type":     "https://api.example.com/errors/user-not-found",
		"title":    "Not Found",
		"status":   float64(404),
		"detail":   "User not found",
		"instance": "/users/999",
		"user_id":  "999",
	}

	for field, expectedValue := range expectedFields {
		if result[field] != expectedValue {
			t.Errorf("%s = %v, want %v", field, result[field], expectedValue)
		}
	}
}
