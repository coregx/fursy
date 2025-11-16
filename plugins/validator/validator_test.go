// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package validator

import (
	"errors"
	"strings"
	"testing"

	"github.com/coregx/fursy"
	"github.com/go-playground/validator/v10"
)

// Test structs.
type TestUser struct {
	Email    string `validate:"required,email"`
	Username string `validate:"required,min=3,max=50"`
	Age      int    `validate:"gte=18,lte=120"`
	Password string `validate:"required,min=8"`
}

type TestProduct struct {
	Name  string  `validate:"required,min=3"`
	Price float64 `validate:"required,gt=0"`
	SKU   string  `validate:"required,alphanum"`
}

type TestAddress struct {
	Street  string `validate:"required"`
	City    string `validate:"required"`
	ZipCode string `validate:"required,numeric,len=5"`
}

type TestUserWithAddress struct {
	Name    string       `validate:"required"`
	Email   string       `validate:"required,email"`
	Address *TestAddress `validate:"required"`
}

type TestCustomTag struct {
	Domain string `validate:"custom_domain"`
}

// TestNew tests validator creation.
func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts *Options
	}{
		{
			name: "default options",
			opts: nil,
		},
		{
			name: "custom tag name",
			opts: &Options{
				TagName: "valid",
			},
		},
		{
			name: "custom messages",
			opts: &Options{
				CustomMessages: map[string]string{
					"required": "{field} is mandatory",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New(tt.opts)
			if v == nil {
				t.Fatal("expected validator, got nil")
			}
			if v.validate == nil {
				t.Error("expected validate instance, got nil")
			}
			if v.options == nil {
				t.Error("expected options, got nil")
			}
		})
	}
}

// TestValidate_Success tests successful validation.
func TestValidate_Success(t *testing.T) {
	v := New()

	tests := []struct {
		name string
		data any
	}{
		{
			name: "valid user",
			data: &TestUser{
				Email:    "user@example.com",
				Username: "johndoe",
				Age:      25,
				Password: "password123",
			},
		},
		{
			name: "valid product",
			data: &TestProduct{
				Name:  "Widget",
				Price: 19.99,
				SKU:   "ABC123",
			},
		},
		{
			name: "nil data",
			data: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.data)
			if err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

// TestValidate_Errors tests validation errors.
//
//nolint:gocognit // Comprehensive test scenarios - complexity is acceptable
func TestValidate_Errors(t *testing.T) {
	v := New()

	tests := []struct {
		name          string
		data          any
		wantErrCount  int
		wantErrFields []string
	}{
		{
			name: "missing required fields",
			data: &TestUser{
				Email:    "",
				Username: "",
				Age:      25,
				Password: "",
			},
			wantErrCount:  3,
			wantErrFields: []string{"Email", "Username", "Password"},
		},
		{
			name: "invalid email",
			data: &TestUser{
				Email:    "not-an-email",
				Username: "johndoe",
				Age:      25,
				Password: "password123",
			},
			wantErrCount:  1,
			wantErrFields: []string{"Email"},
		},
		{
			name: "username too short",
			data: &TestUser{
				Email:    "user@example.com",
				Username: "ab",
				Age:      25,
				Password: "password123",
			},
			wantErrCount:  1,
			wantErrFields: []string{"Username"},
		},
		{
			name: "age below minimum",
			data: &TestUser{
				Email:    "user@example.com",
				Username: "johndoe",
				Age:      17,
				Password: "password123",
			},
			wantErrCount:  1,
			wantErrFields: []string{"Age"},
		},
		{
			name: "password too short",
			data: &TestUser{
				Email:    "user@example.com",
				Username: "johndoe",
				Age:      25,
				Password: "pass",
			},
			wantErrCount:  1,
			wantErrFields: []string{"Password"},
		},
		{
			name: "multiple errors",
			data: &TestUser{
				Email:    "invalid-email",
				Username: "ab",
				Age:      17,
				Password: "short",
			},
			wantErrCount:  4,
			wantErrFields: []string{"Email", "Username", "Age", "Password"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.data)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var validationErrs fursy.ValidationErrors
			if !errors.As(err, &validationErrs) {
				t.Fatalf("expected fursy.ValidationErrors, got: %T", err)
			}

			if len(validationErrs) != tt.wantErrCount {
				t.Errorf("expected %d errors, got %d", tt.wantErrCount, len(validationErrs))
			}

			// Check that expected fields have errors.
			for _, field := range tt.wantErrFields {
				found := false
				for _, verr := range validationErrs {
					if verr.Field == field {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error for field %s, but not found", field)
				}
			}
		})
	}
}

// TestValidate_NestedStructs tests validation of nested structs.
//
//nolint:gocognit,nestif,gocritic // Comprehensive test scenarios - complexity is acceptable
func TestValidate_NestedStructs(t *testing.T) {
	v := New()

	tests := []struct {
		name         string
		data         any
		wantErr      bool
		wantErrField string
	}{
		{
			name: "valid nested struct",
			data: &TestUserWithAddress{
				Name:  "John Doe",
				Email: "john@example.com",
				Address: &TestAddress{
					Street:  "123 Main St",
					City:    "Springfield",
					ZipCode: "12345",
				},
			},
			wantErr: false,
		},
		{
			name: "missing nested required field",
			data: &TestUserWithAddress{
				Name:  "John Doe",
				Email: "john@example.com",
				Address: &TestAddress{
					Street:  "",
					City:    "Springfield",
					ZipCode: "12345",
				},
			},
			wantErr:      true,
			wantErrField: "Street",
		},
		{
			name: "invalid zipcode",
			data: &TestUserWithAddress{
				Name:  "John Doe",
				Email: "john@example.com",
				Address: &TestAddress{
					Street:  "123 Main St",
					City:    "Springfield",
					ZipCode: "ABC",
				},
			},
			wantErr:      true,
			wantErrField: "ZipCode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				var validationErrs fursy.ValidationErrors
				if !errors.As(err, &validationErrs) {
					t.Fatalf("expected fursy.ValidationErrors, got: %T", err)
				}

				if tt.wantErrField != "" {
					found := false
					for _, verr := range validationErrs {
						if verr.Field == tt.wantErrField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected error for field %s, but not found", tt.wantErrField)
					}
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestValidate_CustomMessages tests custom error messages.
func TestValidate_CustomMessages(t *testing.T) {
	customMessages := map[string]string{
		"required": "{field} is mandatory",
		"email":    "Please provide a valid email for {field}",
		"min":      "{field} must have at least {param} characters",
	}

	v := New(&Options{
		CustomMessages: customMessages,
	})

	user := &TestUser{
		Email:    "",
		Username: "ab",
		Age:      25,
		Password: "password123",
	}

	err := v.Validate(user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var validationErrs fursy.ValidationErrors
	if !errors.As(err, &validationErrs) {
		t.Fatalf("expected fursy.ValidationErrors, got: %T", err)
	}

	// Check custom messages are used.
	for _, verr := range validationErrs {
		switch verr.Field {
		case "Email":
			if !strings.Contains(verr.Message, "mandatory") {
				t.Errorf("expected custom 'required' message, got: %s", verr.Message)
			}
		case "Username":
			if !strings.Contains(verr.Message, "at least 3") {
				t.Errorf("expected custom 'min' message, got: %s", verr.Message)
			}
		}
	}
}

// TestRegisterCustomValidator tests custom validator registration.
func TestRegisterCustomValidator(t *testing.T) {
	v := New()

	// Register custom validator.
	err := v.RegisterCustomValidator("custom_domain", func(fl validator.FieldLevel) bool {
		domain := fl.Field().String()
		return strings.HasSuffix(domain, ".example.com")
	})
	if err != nil {
		t.Fatalf("failed to register custom validator: %v", err)
	}

	tests := []struct {
		name    string
		data    *TestCustomTag
		wantErr bool
	}{
		{
			name:    "valid custom domain",
			data:    &TestCustomTag{Domain: "api.example.com"},
			wantErr: false,
		},
		{
			name:    "invalid custom domain",
			data:    &TestCustomTag{Domain: "api.other.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestRegisterAlias tests alias registration.
func TestRegisterAlias(t *testing.T) {
	v := New()

	// Register alias.
	v.RegisterAlias("password", "required,min=8,max=72")

	type TestPassword struct {
		Pass string `validate:"password"`
	}

	tests := []struct {
		name    string
		data    *TestPassword
		wantErr bool
	}{
		{
			name:    "valid password",
			data:    &TestPassword{Pass: "password123"},
			wantErr: false,
		},
		{
			name:    "password too short",
			data:    &TestPassword{Pass: "pass"},
			wantErr: true,
		},
		{
			name:    "missing password",
			data:    &TestPassword{Pass: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestVar tests single variable validation.
func TestVar(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		value   any
		tag     string
		wantErr bool
	}{
		{
			name:    "valid email",
			value:   "user@example.com",
			tag:     "email",
			wantErr: false,
		},
		{
			name:    "invalid email",
			value:   "not-an-email",
			tag:     "email",
			wantErr: true,
		},
		{
			name:    "valid min length",
			value:   "hello",
			tag:     "min=3",
			wantErr: false,
		},
		{
			name:    "invalid min length",
			value:   "hi",
			tag:     "min=3",
			wantErr: true,
		},
		{
			name:    "valid number range",
			value:   25,
			tag:     "gte=18,lte=120",
			wantErr: false,
		},
		{
			name:    "invalid number range",
			value:   17,
			tag:     "gte=18",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Var(tt.value, tt.tag)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// TestStruct tests struct validation convenience method.
func TestStruct(t *testing.T) {
	v := New()

	user := &TestUser{
		Email:    "user@example.com",
		Username: "johndoe",
		Age:      25,
		Password: "password123",
	}

	err := v.Struct(user)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	invalidUser := &TestUser{
		Email:    "invalid",
		Username: "ab",
		Age:      17,
		Password: "short",
	}

	err = v.Struct(invalidUser)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// TestDefaultMessage tests default error message generation.
func TestDefaultMessage(t *testing.T) {
	v := New()

	// Test all major validation tags to ensure coverage.
	tests := []struct {
		data        any
		wantContain string
	}{
		{
			data:        &TestUser{Email: "", Username: "johndoe", Age: 25, Password: "password123"},
			wantContain: "required",
		},
		{
			data:        &TestUser{Email: "invalid", Username: "johndoe", Age: 25, Password: "password123"},
			wantContain: "valid email",
		},
		{
			data:        &TestUser{Email: "user@example.com", Username: "ab", Age: 25, Password: "password123"},
			wantContain: "at least 3",
		},
		{
			data:        &TestUser{Email: "user@example.com", Username: "johndoe", Age: 17, Password: "password123"},
			wantContain: "greater than or equal",
		},
	}

	for i, tt := range tests {
		t.Run(tt.wantContain, func(t *testing.T) {
			err := v.Validate(tt.data)
			if err == nil {
				t.Fatalf("test %d: expected error, got nil", i)
			}

			errMsg := err.Error()
			if !strings.Contains(strings.ToLower(errMsg), strings.ToLower(tt.wantContain)) {
				t.Errorf("test %d: error message %q should contain %q", i, errMsg, tt.wantContain)
			}
		})
	}
}

// TestErrorConversion tests conversion to fursy.ValidationErrors.
func TestErrorConversion(t *testing.T) {
	v := New()

	user := &TestUser{
		Email:    "invalid",
		Username: "ab",
		Age:      17,
		Password: "short",
	}

	err := v.Validate(user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var validationErrs fursy.ValidationErrors
	if !errors.As(err, &validationErrs) {
		t.Fatalf("expected fursy.ValidationErrors, got: %T", err)
	}

	// Check all errors have required fields.
	for _, verr := range validationErrs {
		if verr.Field == "" {
			t.Error("Field should not be empty")
		}
		if verr.Tag == "" {
			t.Error("Tag should not be empty")
		}
		if verr.Message == "" {
			t.Error("Message should not be empty")
		}
	}

	// Check Fields() method works.
	fields := validationErrs.Fields()
	if len(fields) != len(validationErrs) {
		t.Errorf("expected %d field errors, got %d", len(validationErrs), len(fields))
	}
}

// TestInterfaceCompliance tests that Validator implements fursy.Validator.
func TestInterfaceCompliance(_ *testing.T) {
	var _ fursy.Validator = (*Validator)(nil)
}
