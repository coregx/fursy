// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// Test types for OpenAPI generation.
type testUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

func TestOpenAPI_GenerateBasic(t *testing.T) {
	router := New()

	router.GET("/users", func(c *Context) error {
		return c.String(200, "users")
	})

	router.POST("/users", func(c *Context) error {
		return c.String(201, "created")
	})

	doc, err := router.GenerateOpenAPI(Info{
		Title:   "Test API",
		Version: "1.0.0",
	})

	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	if doc.OpenAPI != "3.1.0" {
		t.Errorf("Expected OpenAPI version 3.1.0, got %s", doc.OpenAPI)
	}

	if doc.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got %s", doc.Info.Title)
	}

	if doc.Info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", doc.Info.Version)
	}

	// Check paths - GET and POST to /users should be in one PathItem.
	if len(doc.Paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(doc.Paths))
	}

	// Check /users GET and POST.
	usersPath, exists := doc.Paths["/users"]
	if !exists {
		t.Fatal("/users path not found")
	}

	if usersPath.Get == nil {
		t.Error("/users GET operation not found")
	}

	if usersPath.Post == nil {
		t.Error("/users POST operation not found")
	}
}

func TestOpenAPI_WithInfo(t *testing.T) {
	router := New()
	router.WithInfo(Info{
		Title:       "My API",
		Version:     "2.0.0",
		Description: "Test description",
	})

	router.GET("/test", func(_ *Context) error {
		return nil
	})

	doc, err := router.GenerateOpenAPI(Info{
		Title:   "Ignored",
		Version: "0.0.0",
	})

	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	// Should use router info, not parameter.
	if doc.Info.Title != "My API" {
		t.Errorf("Expected title 'My API', got %s", doc.Info.Title)
	}

	if doc.Info.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got %s", doc.Info.Version)
	}

	if doc.Info.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got %s", doc.Info.Description)
	}
}

func TestOpenAPI_WithServer(t *testing.T) {
	router := New()
	router.WithServer(Server{
		URL:         "https://api.example.com",
		Description: "Production server",
	})
	router.WithServer(Server{
		URL:         "https://staging.example.com",
		Description: "Staging server",
	})

	router.GET("/test", func(_ *Context) error {
		return nil
	})

	doc, err := router.GenerateOpenAPI(Info{
		Title:   "Test",
		Version: "1.0.0",
	})

	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	if len(doc.Servers) != 2 {
		t.Fatalf("Expected 2 servers, got %d", len(doc.Servers))
	}

	if doc.Servers[0].URL != "https://api.example.com" {
		t.Errorf("Expected first server URL 'https://api.example.com', got %s", doc.Servers[0].URL)
	}

	if doc.Servers[1].URL != "https://staging.example.com" {
		t.Errorf("Expected second server URL 'https://staging.example.com', got %s", doc.Servers[1].URL)
	}
}

func TestOpenAPI_HandleWithOptions(t *testing.T) {
	router := New()

	router.HandleWithOptions(http.MethodGet, "/users/:id", func(_ *Context) error {
		return nil
	}, &RouteOptions{
		Summary:     "Get user by ID",
		Description: "Returns a single user",
		Tags:        []string{"users"},
		OperationID: "getUserByID",
	})

	doc, err := router.GenerateOpenAPI(Info{
		Title:   "Test",
		Version: "1.0.0",
	})

	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	// Path should be converted to OpenAPI format.
	usersPath, exists := doc.Paths["/users/{id}"]
	if !exists {
		t.Fatal("/users/{id} path not found")
	}

	if usersPath.Get == nil {
		t.Fatal("/users/{id} GET operation not found")
	}

	op := usersPath.Get

	if op.Summary != "Get user by ID" {
		t.Errorf("Expected summary 'Get user by ID', got %s", op.Summary)
	}

	if op.Description != "Returns a single user" {
		t.Errorf("Expected description 'Returns a single user', got %s", op.Description)
	}

	if len(op.Tags) != 1 || op.Tags[0] != "users" {
		t.Errorf("Expected tags ['users'], got %v", op.Tags)
	}

	if op.OperationID != "getUserByID" {
		t.Errorf("Expected operationId 'getUserByID', got %s", op.OperationID)
	}
}

func TestOpenAPI_PathConversion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "static path",
			input:    "/users",
			expected: "/users",
		},
		{
			name:     "single parameter",
			input:    "/users/:id",
			expected: "/users/{id}",
		},
		{
			name:     "multiple parameters",
			input:    "/users/:userId/posts/:postId",
			expected: "/users/{userId}/posts/{postId}",
		},
		{
			name:     "wildcard",
			input:    "/files/*path",
			expected: "/files/{path}",
		},
		{
			name:     "mixed",
			input:    "/api/v1/users/:id",
			expected: "/api/v1/users/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertPathToOpenAPI(tt.input)
			if result != tt.expected {
				t.Errorf("convertPathToOpenAPI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestOpenAPI_SchemaGeneration(t *testing.T) {
	tests := []struct {
		name     string
		typ      reflect.Type
		expected *Schema
	}{
		{
			name: "string",
			typ:  reflect.TypeOf(""),
			expected: &Schema{
				Type: "string",
			},
		},
		{
			name: "int",
			typ:  reflect.TypeOf(0),
			expected: &Schema{
				Type: "integer",
			},
		},
		{
			name: "bool",
			typ:  reflect.TypeOf(false),
			expected: &Schema{
				Type: "boolean",
			},
		},
		{
			name: "float64",
			typ:  reflect.TypeOf(0.0),
			expected: &Schema{
				Type: "number",
			},
		},
		{
			name: "slice",
			typ:  reflect.TypeOf([]string{}),
			expected: &Schema{
				Type: "array",
				Items: &Schema{
					Type: "string",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSchema(tt.typ)

			if result.Type != tt.expected.Type {
				t.Errorf("Type = %q, want %q", result.Type, tt.expected.Type)
			}

			if tt.expected.Items != nil {
				if result.Items == nil {
					t.Error("Expected Items to be set")
				} else if result.Items.Type != tt.expected.Items.Type {
					t.Errorf("Items.Type = %q, want %q", result.Items.Type, tt.expected.Items.Type)
				}
			}
		})
	}
}

func TestOpenAPI_SchemaGeneration_Struct(t *testing.T) {
	schema := generateSchema(reflect.TypeOf(testUser{}))

	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %s", schema.Type)
	}

	if schema.Properties == nil {
		t.Fatal("Expected Properties to be set")
	}

	// Check id field.
	idSchema, exists := schema.Properties["id"]
	if !exists {
		t.Error("Expected 'id' property")
	} else if idSchema.Type != "integer" {
		t.Errorf("Expected 'id' type 'integer', got %s", idSchema.Type)
	}

	// Check name field.
	nameSchema, exists := schema.Properties["name"]
	if !exists {
		t.Error("Expected 'name' property")
	} else if nameSchema.Type != "string" {
		t.Errorf("Expected 'name' type 'string', got %s", nameSchema.Type)
	}

	// Check email field (omitempty).
	emailSchema, exists := schema.Properties["email"]
	if !exists {
		t.Error("Expected 'email' property")
	} else if emailSchema.Type != "string" {
		t.Errorf("Expected 'email' type 'string', got %s", emailSchema.Type)
	}

	// Check required fields (should not include email with omitempty).
	expectedRequired := []string{"id", "name"}
	if !reflect.DeepEqual(schema.Required, expectedRequired) {
		t.Errorf("Expected required %v, got %v", expectedRequired, schema.Required)
	}
}

func TestOpenAPI_ProblemDetailsSchema(t *testing.T) {
	router := New()
	router.GET("/test", func(_ *Context) error {
		return nil
	})

	doc, err := router.GenerateOpenAPI(Info{
		Title:   "Test",
		Version: "1.0.0",
	})

	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	// Check that Problem schema is included.
	problemSchema, exists := doc.Components.Schemas["Problem"]
	if !exists {
		t.Fatal("Problem schema not found in components")
	}

	if problemSchema.Type != "object" {
		t.Errorf("Expected Problem type 'object', got %s", problemSchema.Type)
	}

	// Check required fields.
	expectedRequired := []string{"type", "title", "status"}
	if !reflect.DeepEqual(problemSchema.Required, expectedRequired) {
		t.Errorf("Expected Problem required %v, got %v", expectedRequired, problemSchema.Required)
	}

	// Check that Problem response is included.
	problemResponse, exists := doc.Components.Responses["Problem"]
	if !exists {
		t.Fatal("Problem response not found in components")
	}

	if problemResponse.Description != "RFC 9457 Problem Details" {
		t.Errorf("Expected Problem response description 'RFC 9457 Problem Details', got %s", problemResponse.Description)
	}
}

//nolint:nestif,gocritic // Test validation requires nested checks.
func TestOpenAPI_DefaultErrorResponses(t *testing.T) {
	router := New()
	router.GET("/users", func(_ *Context) error {
		return nil
	})

	doc, err := router.GenerateOpenAPI(Info{
		Title:   "Test",
		Version: "1.0.0",
	})

	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	usersPath := doc.Paths["/users"]
	if usersPath.Get == nil {
		t.Fatal("GET /users not found")
	}

	op := usersPath.Get

	// Check 400 response.
	resp400, exists := op.Responses["400"]
	if !exists {
		t.Error("Expected 400 response")
	} else {
		if resp400.Description != "Bad Request" {
			t.Errorf("Expected 400 description 'Bad Request', got %s", resp400.Description)
		}

		content, exists := resp400.Content["application/problem+json"]
		if !exists {
			t.Error("Expected application/problem+json content type for 400")
		} else if content.Schema.Ref != "#/components/schemas/Problem" {
			t.Errorf("Expected 400 schema ref '#/components/schemas/Problem', got %s", content.Schema.Ref)
		}
	}

	// Check 500 response.
	resp500, exists := op.Responses["500"]
	if !exists {
		t.Error("Expected 500 response")
	} else {
		if resp500.Description != "Internal Server Error" {
			t.Errorf("Expected 500 description 'Internal Server Error', got %s", resp500.Description)
		}
	}
}

func TestOpenAPI_AllHTTPMethods(t *testing.T) {
	router := New()

	router.GET("/test", func(_ *Context) error { return nil })
	router.POST("/test", func(_ *Context) error { return nil })
	router.PUT("/test", func(_ *Context) error { return nil })
	router.DELETE("/test", func(_ *Context) error { return nil })
	router.PATCH("/test", func(_ *Context) error { return nil })
	router.HEAD("/test", func(_ *Context) error { return nil })
	router.OPTIONS("/test", func(_ *Context) error { return nil })

	doc, err := router.GenerateOpenAPI(Info{
		Title:   "Test",
		Version: "1.0.0",
	})

	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	testPath := doc.Paths["/test"]

	if testPath.Get == nil {
		t.Error("GET operation not found")
	}
	if testPath.Post == nil {
		t.Error("POST operation not found")
	}
	if testPath.Put == nil {
		t.Error("PUT operation not found")
	}
	if testPath.Delete == nil {
		t.Error("DELETE operation not found")
	}
	if testPath.Patch == nil {
		t.Error("PATCH operation not found")
	}
	if testPath.Head == nil {
		t.Error("HEAD operation not found")
	}
	if testPath.Options == nil {
		t.Error("OPTIONS operation not found")
	}
}

func TestOpenAPI_JSONMarshaling(t *testing.T) {
	router := New()
	router.WithInfo(Info{
		Title:   "Test API",
		Version: "1.0.0",
	})

	router.GET("/users", func(_ *Context) error {
		return nil
	})

	doc, err := router.GenerateOpenAPI(Info{})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	// Test JSON marshaling.
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty JSON")
	}

	// Check that it's valid JSON.
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if result["openapi"] != "3.1.0" {
		t.Errorf("Expected openapi '3.1.0', got %v", result["openapi"])
	}
}

func TestRouter_ServeOpenAPI(t *testing.T) {
	router := New()
	router.WithInfo(Info{
		Title:       "Test API",
		Version:     "2.0.0",
		Description: "API for testing",
	})

	router.GET("/users", func(c *Context) error {
		return c.String(200, "users")
	})

	router.POST("/users", func(c *Context) error {
		return c.String(201, "created")
	})

	// Serve OpenAPI spec.
	router.ServeOpenAPI("/openapi.json")

	// Test the endpoint.
	req := httptest.NewRequest("GET", "/openapi.json", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Expected Content-Type 'application/json; charset=utf-8', got %s", contentType)
	}

	// Verify it's valid OpenAPI JSON.
	var doc OpenAPI
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatalf("Failed to unmarshal OpenAPI document: %v", err)
	}

	if doc.OpenAPI != "3.1.0" {
		t.Errorf("Expected OpenAPI version '3.1.0', got %s", doc.OpenAPI)
	}

	if doc.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got %s", doc.Info.Title)
	}

	if doc.Info.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got %s", doc.Info.Version)
	}

	if doc.Info.Description != "API for testing" {
		t.Errorf("Expected description 'API for testing', got %s", doc.Info.Description)
	}

	// Verify paths are included (/users + /openapi.json).
	if len(doc.Paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(doc.Paths))
	}

	usersPath, exists := doc.Paths["/users"]
	if !exists {
		t.Error("/users path not found")
	}

	if usersPath.Get == nil {
		t.Error("GET operation not found")
	}

	if usersPath.Post == nil {
		t.Error("POST operation not found")
	}

	// Verify OpenAPI endpoint itself is documented.
	_, exists = doc.Paths["/openapi.json"]
	if !exists {
		t.Error("/openapi.json path not found in spec")
	}
}

func TestRouter_ServeOpenAPI_DefaultInfo(t *testing.T) {
	router := New()

	router.GET("/test", func(_ *Context) error {
		return nil
	})

	// Serve OpenAPI spec without configuring Info.
	router.ServeOpenAPI("/api-spec.json")

	// Test the endpoint.
	req := httptest.NewRequest("GET", "/api-spec.json", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify it uses default info.
	var doc OpenAPI
	if err := json.Unmarshal(w.Body.Bytes(), &doc); err != nil {
		t.Fatalf("Failed to unmarshal OpenAPI document: %v", err)
	}

	if doc.Info.Title != "API Documentation" {
		t.Errorf("Expected default title 'API Documentation', got %s", doc.Info.Title)
	}

	if doc.Info.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got %s", doc.Info.Version)
	}
}
