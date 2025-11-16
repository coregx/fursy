// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockValidator is a mock validator for testing.
type mockValidator struct {
	shouldFail bool
	errors     ValidationErrors
}

func (m *mockValidator) Validate(_ any) error {
	if m.shouldFail {
		if m.errors != nil {
			return m.errors
		}
		return ValidationErrors{
			{Field: "email", Tag: "required", Message: "email is required"},
			{Field: "age", Tag: "min", Message: "age must be at least 18"},
		}
	}
	return nil
}

// TestValidationError tests ValidationError.Error().
func TestValidationError(t *testing.T) {
	tests := []struct {
		name    string
		err     ValidationError
		wantMsg string
	}{
		{
			name: "with message",
			err: ValidationError{
				Field:   "email",
				Tag:     "required",
				Message: "email is required",
			},
			wantMsg: "email is required",
		},
		{
			name: "without message",
			err: ValidationError{
				Field: "age",
				Tag:   "min",
			},
			wantMsg: "field 'age' failed 'min' validation",
		},
		{
			name: "nested field",
			err: ValidationError{
				Field:   "Address.City",
				Tag:     "required",
				Message: "city is required",
			},
			wantMsg: "city is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

// TestValidationErrors_Error tests ValidationErrors.Error().
func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name    string
		errors  ValidationErrors
		wantMsg string
	}{
		{
			name:    "empty",
			errors:  ValidationErrors{},
			wantMsg: "validation failed",
		},
		{
			name: "single error",
			errors: ValidationErrors{
				{Field: "email", Tag: "required", Message: "email is required"},
			},
			wantMsg: "email is required",
		},
		{
			name: "multiple errors",
			errors: ValidationErrors{
				{Field: "email", Tag: "required", Message: "email is required"},
				{Field: "age", Tag: "min", Message: "age must be at least 18"},
			},
			wantMsg: "validation failed: email is required; age must be at least 18",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errors.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

// TestValidationErrors_Fields tests ValidationErrors.Fields().
func TestValidationErrors_Fields(t *testing.T) {
	errs := ValidationErrors{
		{Field: "email", Tag: "required", Message: "email is required"},
		{Field: "age", Tag: "min", Message: "age must be at least 18"},
		{Field: "password", Tag: "min", Message: "password must be at least 8 characters"},
	}

	fields := errs.Fields()

	if len(fields) != 3 {
		t.Errorf("Fields() returned %d fields, want 3", len(fields))
	}

	expectedFields := map[string]string{
		"email":    "email is required",
		"age":      "age must be at least 18",
		"password": "password must be at least 8 characters",
	}

	for field, expectedMsg := range expectedFields {
		if msg, ok := fields[field]; !ok {
			t.Errorf("Fields() missing field %q", field)
		} else if msg != expectedMsg {
			t.Errorf("Fields()[%q] = %q, want %q", field, msg, expectedMsg)
		}
	}
}

// TestValidationErrors_IsEmpty tests ValidationErrors.IsEmpty().
func TestValidationErrors_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		errors ValidationErrors
		want   bool
	}{
		{
			name:   "empty",
			errors: ValidationErrors{},
			want:   true,
		},
		{
			name:   "nil",
			errors: nil,
			want:   true,
		},
		{
			name: "not empty",
			errors: ValidationErrors{
				{Field: "email", Tag: "required", Message: "email is required"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errors.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestValidationErrors_Add tests ValidationErrors.Add().
func TestValidationErrors_Add(t *testing.T) {
	var errs ValidationErrors

	if !errs.IsEmpty() {
		t.Error("errs should be empty initially")
	}

	errs.Add("email", "required", "email is required")

	if errs.IsEmpty() {
		t.Error("errs should not be empty after Add()")
	}

	if len(errs) != 1 {
		t.Errorf("len(errs) = %d, want 1", len(errs))
	}

	if errs[0].Field != "email" {
		t.Errorf("Field = %q, want %q", errs[0].Field, "email")
	}
	if errs[0].Tag != "required" {
		t.Errorf("Tag = %q, want %q", errs[0].Tag, "required")
	}
	if errs[0].Message != "email is required" {
		t.Errorf("Message = %q, want %q", errs[0].Message, "email is required")
	}

	// Add more.
	errs.Add("age", "min", "age must be at least 18")

	if len(errs) != 2 {
		t.Errorf("len(errs) = %d, want 2", len(errs))
	}
}

// TestValidationErrors_AddError tests ValidationErrors.AddError().
func TestValidationErrors_AddError(t *testing.T) {
	var errs ValidationErrors

	err := ValidationError{
		Field:   "email",
		Tag:     "required",
		Value:   "",
		Message: "email is required",
	}

	errs.AddError(err)

	if len(errs) != 1 {
		t.Errorf("len(errs) = %d, want 1", len(errs))
	}

	if errs[0].Field != "email" {
		t.Errorf("Field = %q, want %q", errs[0].Field, "email")
	}
}

// TestRouter_SetValidator tests Router.SetValidator().
func TestRouter_SetValidator(t *testing.T) {
	r := New()

	if r.validator != nil {
		t.Error("validator should be nil by default")
	}

	validator := &mockValidator{}
	r.SetValidator(validator)

	if r.validator == nil {
		t.Error("validator should be set")
	}

	// Test chainability.
	result := r.SetValidator(validator)
	if result != r {
		t.Error("SetValidator() should return the router for chaining")
	}
}

// TestContext_Bind_NoValidator tests binding without validator.
func TestContext_Bind_NoValidator(t *testing.T) {
	r := New()

	type Request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	POST[Request, Request](r, "/test", func(c *Box[Request, Request]) error {
		if c.ReqBody == nil {
			return c.BadRequest(Request{Name: "Request body is nil"})
		}

		return c.OK(*c.ReqBody)
	})

	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Should succeed even with invalid email (no validator).
	body = `{"name":"John","email":"invalid"}`
	req = httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200 (no validation), got %d", w.Code)
	}
}

// TestContext_Bind_WithValidator tests binding with validator.
func TestContext_Bind_WithValidator(t *testing.T) {
	r := New()

	type Request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Set failing validator.
	r.SetValidator(&mockValidator{shouldFail: true})

	POST[Request, Request](r, "/test", func(c *Box[Request, Request]) error {
		// This should not be reached due to validation failure.
		return c.OK(*c.ReqBody)
	})

	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should return 500 (handler returned error).
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 (validation failed), got %d", w.Code)
	}
}

// TestContext_Bind_ValidatorSuccess tests binding with passing validator.
func TestContext_Bind_ValidatorSuccess(t *testing.T) {
	r := New()

	type Request struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Set passing validator.
	r.SetValidator(&mockValidator{shouldFail: false})

	POST[Request, Request](r, "/test", func(c *Box[Request, Request]) error {
		if c.ReqBody == nil {
			return c.BadRequest(Request{Name: "Request body is nil"})
		}

		return c.OK(*c.ReqBody)
	})

	body := `{"name":"John","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// TestContext_Bind_EmptyType tests binding with Empty type (no validation).
func TestContext_Bind_EmptyType(t *testing.T) {
	r := New()

	// Set failing validator - should not be called for Empty type.
	r.SetValidator(&mockValidator{shouldFail: true})

	GET[Empty, string](r, "/test", func(c *Box[Empty, string]) error {
		return c.OK("success")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should succeed even with failing validator (Empty type skips binding/validation).
	if w.Code != 200 {
		t.Errorf("expected status 200 (Empty type), got %d", w.Code)
	}
}

// TestContext_Bind_CustomValidationErrors tests custom validation errors.
func TestContext_Bind_CustomValidationErrors(t *testing.T) {
	r := New()

	type Request struct {
		Email string `json:"email"`
		Age   int    `json:"age"`
	}

	// Set validator with custom errors.
	customErrors := ValidationErrors{
		{Field: "email", Tag: "email", Message: "must be a valid email address"},
		{Field: "age", Tag: "gte", Message: "must be at least 18 years old"},
	}
	r.SetValidator(&mockValidator{shouldFail: true, errors: customErrors})

	POST[Request, Request](r, "/test", func(c *Box[Request, Request]) error {
		return c.OK(*c.ReqBody)
	})

	body := `{"email":"invalid","age":15}`
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Validation should fail.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 (validation failed), got %d", w.Code)
	}
}

// TestValidationErrors_AsError tests using ValidationErrors as error.
func TestValidationErrors_AsError(t *testing.T) {
	validationErrs := ValidationErrors{
		{Field: "email", Tag: "required", Message: "email is required"},
	}

	// Should be usable as error.
	var err error = validationErrs

	if validationErrs.IsEmpty() {
		t.Error("ValidationErrors should not be empty")
	}

	// Should be detectable with errors.As.
	var ve ValidationErrors
	if !errors.As(err, &ve) {
		t.Error("should be able to use errors.As with ValidationErrors")
	}

	if len(ve) != 1 {
		t.Errorf("expected 1 validation error, got %d", len(ve))
	}
}

// CreateUserRequest is a test struct for integration tests.
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// UserResponse is a test struct for integration tests.
type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// emailValidator is a mock validator that checks email contains "@".
type emailValidator struct{}

func (v *emailValidator) Validate(obj any) error {
	req, ok := obj.(*CreateUserRequest)
	if !ok {
		return nil
	}

	var errs ValidationErrors
	if req.Email == "" {
		errs.Add("email", "required", "email is required")
	} else if !contains(req.Email, "@") {
		errs.Add("email", "email", "must be a valid email address")
	}
	if req.Age < 18 {
		errs.Add("age", "gte", "must be at least 18 years old")
	}

	if !errs.IsEmpty() {
		return errs
	}
	return nil
}

// Helper function for integration test.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" || findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestValidation_Integration tests full integration scenario.
func TestValidation_Integration(t *testing.T) {
	r := New()

	r.SetValidator(&emailValidator{})

	POST[CreateUserRequest, UserResponse](r, "/users", func(c *Box[CreateUserRequest, UserResponse]) error {
		// If we get here, validation passed.
		return c.Created("/users/1", UserResponse{
			ID:    1,
			Name:  c.ReqBody.Name,
			Email: c.ReqBody.Email,
		})
	})

	// Test 1: Valid request.
	t.Run("valid request", func(t *testing.T) {
		body := `{"name":"John","email":"john@example.com","age":25}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}
	})

	// Test 2: Invalid email.
	t.Run("invalid email", func(t *testing.T) {
		body := `{"name":"John","email":"invalid","age":25}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 (validation failed), got %d", w.Code)
		}
	})

	// Test 3: Age too young.
	t.Run("age too young", func(t *testing.T) {
		body := `{"name":"John","email":"john@example.com","age":15}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 (validation failed), got %d", w.Code)
		}
	})

	// Test 4: Multiple validation errors.
	t.Run("multiple errors", func(t *testing.T) {
		body := `{"name":"John","email":"invalid","age":15}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 (validation failed), got %d", w.Code)
		}
	})
}
