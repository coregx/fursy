// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"encoding/json"
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

// recursiveNode is a self-referential type used to test schema cycle detection.
type recursiveNode struct {
	Value    string          `json:"value"`
	Children []recursiveNode `json:"children,omitempty"`
	Parent   *recursiveNode  `json:"parent,omitempty"`
}

// mutualA and mutualB form a mutually-recursive type cycle.
type mutualA struct {
	Name string   `json:"name"`
	B    *mutualB `json:"b,omitempty"`
}

type mutualB struct {
	ID int      `json:"id"`
	A  *mutualA `json:"a,omitempty"`
}

func TestOpenAPI_GenerateBasic(t *testing.T) {
	router := New()

	router.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "users")
	})

	router.Handle("POST", "/users", func(c *Context) error {
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

	router.Handle("GET", "/test", func(_ *Context) error {
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

	router.Handle("GET", "/test", func(_ *Context) error {
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
	router.Handle("GET", "/test", func(_ *Context) error {
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
	router.Handle("GET", "/users", func(_ *Context) error {
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

	router.Handle("GET", "/test", func(_ *Context) error { return nil })
	router.Handle("POST", "/test", func(_ *Context) error { return nil })
	router.Handle("PUT", "/test", func(_ *Context) error { return nil })
	router.Handle("DELETE", "/test", func(_ *Context) error { return nil })
	router.Handle("PATCH", "/test", func(_ *Context) error { return nil })
	router.Handle("HEAD", "/test", func(_ *Context) error { return nil })
	router.Handle("OPTIONS", "/test", func(_ *Context) error { return nil })

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

	router.Handle("GET", "/users", func(_ *Context) error {
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

	router.Handle("GET", "/users", func(c *Context) error {
		return c.String(200, "users")
	})

	router.Handle("POST", "/users", func(c *Context) error {
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

func TestOpenAPI_WriteJSON(t *testing.T) {
	router := New()
	router.WithInfo(Info{
		Title:   "Write JSON Test",
		Version: "1.0.0",
	})
	router.Handle("GET", "/ping", func(_ *Context) error {
		return nil
	})

	doc, err := router.GenerateOpenAPI(Info{})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	w := httptest.NewRecorder()
	if err := doc.WriteJSON(w); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("expected Content-Type 'application/json; charset=utf-8', got %q", contentType)
	}

	if w.Body.Len() == 0 {
		t.Error("expected non-empty body")
	}

	// Verify valid JSON with expected openapi version.
	var result map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}

	if result["openapi"] != "3.1.0" {
		t.Errorf("expected openapi '3.1.0', got %v", result["openapi"])
	}

	info, ok := result["info"].(map[string]any)
	if !ok {
		t.Fatal("expected info object in JSON")
	}

	if info["title"] != "Write JSON Test" {
		t.Errorf("expected title 'Write JSON Test', got %v", info["title"])
	}
}

func TestOpenAPI_WriteYAML(t *testing.T) {
	router := New()
	router.Handle("GET", "/test", func(_ *Context) error {
		return nil
	})

	doc, err := router.GenerateOpenAPI(Info{
		Title:   "YAML Test",
		Version: "1.0.0",
	})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	w := httptest.NewRecorder()
	err = doc.WriteYAML(w)

	// WriteYAML is not yet implemented and must return an error.
	if err == nil {
		t.Fatal("expected error from WriteYAML, got nil")
	}

	if err.Error() != "YAML output not yet implemented (requires external dependency)" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestRouter_ServeOpenAPI_DefaultInfo(t *testing.T) {
	router := New()

	router.Handle("GET", "/test", func(_ *Context) error {
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

// TestOpenAPI_CycleDetection verifies that recursive types terminate during
// schema generation instead of recursing infinitely.
func TestOpenAPI_CycleDetection(t *testing.T) {
	// Self-referential via slice and pointer.
	schema := generateSchema(reflect.TypeOf(recursiveNode{}))
	if schema.Type != "object" {
		t.Fatalf("expected object schema, got %q", schema.Type)
	}
	if _, ok := schema.Properties["value"]; !ok {
		t.Error("expected 'value' property")
	}

	children := schema.Properties["children"]
	if children == nil || children.Type != "array" || children.Items == nil {
		t.Fatalf("expected 'children' array with items, got %+v", children)
	}
	if children.Items.Type != "object" {
		t.Errorf("expected cycle placeholder object for children items, got %q", children.Items.Type)
	}

	// Mutually-recursive types must also terminate.
	mutual := generateSchema(reflect.TypeOf(mutualA{}))
	if mutual.Type != "object" {
		t.Fatalf("expected object schema, got %q", mutual.Type)
	}
	if _, ok := mutual.Properties["b"]; !ok {
		t.Error("expected 'b' property")
	}
}

// TestOpenAPI_AutoPathParameters verifies that :name path templates are
// declared as required path parameters.
func TestOpenAPI_AutoPathParameters(t *testing.T) {
	router := New()
	router.Handle("GET", "/users/:id", func(_ *Context) error { return nil })

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	op := doc.Paths["/users/{id}"].Get
	if op == nil {
		t.Fatal("GET /users/{id} not found")
	}

	id := findOpenAPIParameter(op.Parameters, "id", "path")
	if id == nil {
		t.Fatal("expected auto-declared 'id' path parameter")
	}
	if !id.Required {
		t.Error("expected 'id' path parameter to be required")
	}
	if id.Schema == nil || id.Schema.Type != "string" {
		t.Errorf("expected 'id' schema type 'string', got %+v", id.Schema)
	}
}

// TestOpenAPI_PathParameterOverride verifies that explicitly declared path
// parameters take precedence over auto-declaration.
func TestOpenAPI_PathParameterOverride(t *testing.T) {
	router := New()
	router.HandleWithOptions("GET", "/users/:id", func(_ *Context) error { return nil }, &RouteOptions{
		Parameters: []RouteParameter{
			{Name: "id", In: "path", Required: true, Description: "User ID", Type: reflect.TypeOf(int64(0))},
		},
	})

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	op := doc.Paths["/users/{id}"].Get
	count := 0
	for _, p := range op.Parameters {
		if p.Name == "id" && p.In == "path" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one 'id' path parameter, got %d", count)
	}

	id := findOpenAPIParameter(op.Parameters, "id", "path")
	if id.Description != "User ID" {
		t.Errorf("expected user-supplied description, got %q", id.Description)
	}
	if id.Schema == nil || id.Schema.Type != "integer" {
		t.Errorf("expected user-supplied integer schema, got %+v", id.Schema)
	}
}

// TestOpenAPI_UserResponsesNotClobbered verifies that user-supplied error
// responses survive generation and defaults are only added when absent.
func TestOpenAPI_UserResponsesNotClobbered(t *testing.T) {
	router := New()
	router.HandleWithOptions("GET", "/users", func(_ *Context) error { return nil }, &RouteOptions{
		Responses: map[int]RouteResponse{
			400: {
				Description: "Custom Bad Request",
				ContentType: MIMEApplicationJSON,
				Type:        reflect.TypeOf(testUser{}),
			},
		},
	})

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	op := doc.Paths["/users"].Get
	resp400, ok := op.Responses["400"]
	if !ok {
		t.Fatal("expected 400 response")
	}
	if resp400.Description != "Custom Bad Request" {
		t.Errorf("user-supplied 400 was clobbered: got %q", resp400.Description)
	}

	if _, ok := op.Responses["500"]; !ok {
		t.Error("expected default 500 response to still be added")
	}
}

// findOpenAPIParameter returns the first parameter matching name and location.
func findOpenAPIParameter(params []Parameter, name, in string) *Parameter {
	for i := range params {
		if params[i].Name == name && params[i].In == in {
			return &params[i]
		}
	}
	return nil
}

// TestGenericBodyType verifies Empty maps to nil and other types are preserved.
func TestGenericBodyType(t *testing.T) {
	if got := genericBodyType[Empty](); got != nil {
		t.Errorf("genericBodyType[Empty]() = %v, want nil", got)
	}
	if got := genericBodyType[testUser](); got != reflect.TypeOf(testUser{}) {
		t.Errorf("genericBodyType[testUser]() = %v, want %v", got, reflect.TypeOf(testUser{}))
	}
	if got := genericBodyType[*testUser](); got != reflect.TypeOf((*testUser)(nil)) {
		t.Errorf("genericBodyType[*testUser]() = %v, want %v", got, reflect.TypeOf((*testUser)(nil)))
	}
	if got := genericBodyType[string](); got != reflect.TypeOf("") {
		t.Errorf("genericBodyType[string]() = %v, want %v", got, reflect.TypeOf(""))
	}
	if got := genericBodyType[[]testUser](); got == nil {
		t.Error("genericBodyType[[]testUser]() = nil, want non-nil")
	}
}

// TestRegisterGeneric_RecordsTypes verifies that type-safe handlers record
// their Req/Res body types as RouteInfo metadata.
func TestRegisterGeneric_RecordsTypes(t *testing.T) {
	router := New()
	router.POST[testUser, testUser]("/users", func(c *Box[testUser, testUser]) error {
		return c.Created("/users/1", *c.ReqBody)
	})
	router.GET[Empty, testUser]("/users/:id", func(_ *Box[Empty, testUser]) error {
		return nil
	})
	router.DELETE[Empty, Empty]("/users/:id", func(_ *Box[Empty, Empty]) error {
		return nil
	})

	if len(router.routes) != 3 {
		t.Fatalf("expected 3 recorded routes, got %d", len(router.routes))
	}

	post := router.routes[0]
	if post.RequestType != reflect.TypeOf(testUser{}) {
		t.Errorf("POST RequestType = %v, want %v", post.RequestType, reflect.TypeOf(testUser{}))
	}
	if post.ResponseType != reflect.TypeOf(testUser{}) {
		t.Errorf("POST ResponseType = %v, want %v", post.ResponseType, reflect.TypeOf(testUser{}))
	}

	get := router.routes[1]
	if get.RequestType != nil {
		t.Errorf("GET RequestType = %v, want nil (Empty)", get.RequestType)
	}
	if get.ResponseType != reflect.TypeOf(testUser{}) {
		t.Errorf("GET ResponseType = %v, want %v", get.ResponseType, reflect.TypeOf(testUser{}))
	}

	del := router.routes[2]
	if del.RequestType != nil || del.ResponseType != nil {
		t.Errorf("DELETE types = (%v, %v), want (nil, nil)", del.RequestType, del.ResponseType)
	}
}

// TestRegisterGeneric_WithOptions verifies that variadic RouteOptions are
// recorded by type-safe handlers.
func TestRegisterGeneric_WithOptions(t *testing.T) {
	router := New()
	router.GET[Empty, testUser]("/users/:id", func(_ *Box[Empty, testUser]) error { return nil },
		&RouteOptions{Summary: "Get user", Tags: []string{"users"}})

	if len(router.routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(router.routes))
	}
	if router.routes[0].Summary != "Get user" {
		t.Errorf("Summary = %q, want %q", router.routes[0].Summary, "Get user")
	}
	if len(router.routes[0].Tags) != 1 || router.routes[0].Tags[0] != "users" {
		t.Errorf("Tags = %v, want [users]", router.routes[0].Tags)
	}
}

// TestOpenAPI_GenericHandlerSchemas verifies that type-safe handlers produce
// request and response schemas in the generated document.
func TestOpenAPI_GenericHandlerSchemas(t *testing.T) {
	router := New()
	router.POST[testUser, testUser]("/users", func(c *Box[testUser, testUser]) error {
		return c.Created("/users/1", *c.ReqBody)
	})
	router.GET[Empty, []testUser]("/users", func(_ *Box[Empty, []testUser]) error {
		return nil
	})

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	post := doc.Paths["/users"].Post
	if post == nil {
		t.Fatal("POST /users not found")
	}
	if post.RequestBody == nil {
		t.Fatal("expected POST request body schema")
	}
	reqMedia, ok := post.RequestBody.Content[MIMEApplicationJSON]
	if !ok || reqMedia.Schema == nil {
		t.Fatal("expected application/json request schema")
	}
	if reqMedia.Schema.Ref != "#/components/schemas/testUser" {
		t.Errorf("request schema ref = %q, want %q", reqMedia.Schema.Ref, "#/components/schemas/testUser")
	}

	respMedia, ok := post.Responses["200"].Content[MIMEApplicationJSON]
	if !ok || respMedia.Schema == nil {
		t.Fatal("expected application/json response schema")
	}
	if respMedia.Schema.Ref != "#/components/schemas/testUser" {
		t.Errorf("response schema ref = %q, want %q", respMedia.Schema.Ref, "#/components/schemas/testUser")
	}

	// The shared type is registered once as a named component.
	userSchema, ok := doc.Components.Schemas["testUser"]
	if !ok {
		t.Fatal("expected testUser component schema")
	}
	if _, ok := userSchema.Properties["id"]; !ok {
		t.Error("expected testUser schema to include 'id' property")
	}
	if _, ok := userSchema.Properties["name"]; !ok {
		t.Error("expected testUser schema to include 'name' property")
	}
	if len(userSchema.Required) != 2 {
		t.Errorf("expected 2 required properties (id, name), got %v", userSchema.Required)
	}

	// GET uses Empty request type: no request body should be generated.
	get := doc.Paths["/users"].Get
	if get == nil {
		t.Fatal("GET /users not found")
	}
	if get.RequestBody != nil {
		t.Error("expected no request body for Empty request type")
	}

	// Array response type should produce an array schema whose items $ref the
	// named component.
	getResp, ok := get.Responses["200"].Content[MIMEApplicationJSON]
	if !ok || getResp.Schema == nil {
		t.Fatal("expected GET response schema")
	}
	if getResp.Schema.Type != "array" {
		t.Errorf("expected GET response schema type 'array', got %q", getResp.Schema.Type)
	}
	if getResp.Schema.Items == nil || getResp.Schema.Items.Ref != "#/components/schemas/testUser" {
		t.Errorf("expected array items $ref to testUser, got %+v", getResp.Schema.Items)
	}
}

// TestOpenAPI_SuccessStatus verifies that SuccessStatus controls the inferred
// success response, including 204 with no content.
func TestOpenAPI_SuccessStatus(t *testing.T) {
	router := New()
	router.POST[testUser, testUser]("/users", func(c *Box[testUser, testUser]) error {
		return c.Created("/users/1", *c.ReqBody)
	}, &RouteOptions{SuccessStatus: http.StatusCreated})
	router.DELETE[Empty, Empty]("/users/:id", func(c *Box[Empty, Empty]) error {
		return c.NoContentSuccess()
	}, &RouteOptions{SuccessStatus: http.StatusNoContent})

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	post := doc.Paths["/users"].Post
	if _, ok := post.Responses["200"]; ok {
		t.Error("did not expect a 200 response when SuccessStatus is 201")
	}
	created, ok := post.Responses["201"]
	if !ok {
		t.Fatal("expected 201 response")
	}
	if created.Description != descCreated {
		t.Errorf("201 description = %q, want %q", created.Description, descCreated)
	}
	if media, ok := created.Content[MIMEApplicationJSON]; !ok || media.Schema == nil {
		t.Error("expected 201 response content schema")
	}

	del := doc.Paths["/users/{id}"].Delete
	noContent, ok := del.Responses["204"]
	if !ok {
		t.Fatal("expected 204 response")
	}
	if noContent.Description != descNoContent {
		t.Errorf("204 description = %q, want %q", noContent.Description, descNoContent)
	}
	if len(noContent.Content) != 0 {
		t.Errorf("204 response must not have content, got %d media types", len(noContent.Content))
	}
	if _, ok := del.Responses["400"]; !ok {
		t.Error("expected default 400 response alongside 204")
	}
}

// TestOpenAPI_OptionalRequestBody verifies OptionalRequestBody relaxes the
// inferred request body's required flag.
func TestOpenAPI_OptionalRequestBody(t *testing.T) {
	router := New()
	router.POST[testUser, testUser]("/users", func(_ *Box[testUser, testUser]) error {
		return nil
	}, &RouteOptions{OptionalRequestBody: true})

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	post := doc.Paths["/users"].Post
	if post.RequestBody == nil {
		t.Fatal("expected request body")
	}
	if post.RequestBody.Required {
		t.Error("expected optional (Required=false) request body")
	}
}

// TestOpenAPI_ExplicitResponsesOverrideInference verifies that explicit
// RouteOptions.Responses suppress the inferred success response.
func TestOpenAPI_ExplicitResponsesOverrideInference(t *testing.T) {
	router := New()
	router.POST[testUser, testUser]("/users", func(_ *Box[testUser, testUser]) error {
		return nil
	}, &RouteOptions{
		Responses: map[int]RouteResponse{
			http.StatusCreated: {
				Description: "Custom Created",
				ContentType: MIMEApplicationJSON,
				Type:        reflect.TypeOf(testUser{}),
			},
		},
	})

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	post := doc.Paths["/users"].Post
	if _, ok := post.Responses["200"]; ok {
		t.Error("inference should not add 200 when explicit Responses are provided")
	}
	if _, ok := post.Responses["201"]; !ok {
		t.Error("expected user-provided 201 response")
	}
}

// TestOpenAPI_GroupGenericRoute verifies that group generic routes carry the
// prefixed path, options, and Req/Res types.
func TestOpenAPI_GroupGenericRoute(t *testing.T) {
	router := New()
	api := router.Group("/api")
	api.GET[Empty, testUser]("/users/:id", func(_ *Box[Empty, testUser]) error {
		return nil
	}, &RouteOptions{Summary: "Get user", Tags: []string{"users"}})

	if len(router.routes) != 1 {
		t.Fatalf("expected 1 recorded route, got %d", len(router.routes))
	}
	if router.routes[0].ResponseType != reflect.TypeOf(testUser{}) {
		t.Errorf("group ResponseType = %v, want %v", router.routes[0].ResponseType, reflect.TypeOf(testUser{}))
	}

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	op := doc.Paths["/api/users/{id}"].Get
	if op == nil {
		t.Fatal("expected GET /api/users/{id} from group registration")
	}
	if op.Summary != "Get user" {
		t.Errorf("Summary = %q, want %q", op.Summary, "Get user")
	}
	if len(op.Tags) != 1 || op.Tags[0] != "users" {
		t.Errorf("Tags = %v, want [users]", op.Tags)
	}
	if findOpenAPIParameter(op.Parameters, "id", "path") == nil {
		t.Error("expected auto-declared 'id' path parameter on group route")
	}
}

// TestOpenAPI_AutoOperationID verifies deterministic operationId generation
// from method + path, with global uniqueness across the document.
func TestOpenAPI_AutoOperationID(t *testing.T) {
	router := New()
	router.GET[Empty, testUser]("/users", func(_ *Box[Empty, testUser]) error { return nil })
	router.POST[testUser, testUser]("/users", func(_ *Box[testUser, testUser]) error { return nil })
	router.GET[Empty, testUser]("/users/:id", func(_ *Box[Empty, testUser]) error { return nil })
	router.DELETE[Empty, Empty]("/files/*path", func(_ *Box[Empty, Empty]) error { return nil })

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	got := map[string]string{
		"get /users":           doc.Paths["/users"].Get.OperationID,
		"post /users":          doc.Paths["/users"].Post.OperationID,
		"get /users/{id}":      doc.Paths["/users/{id}"].Get.OperationID,
		"delete /files/{path}": doc.Paths["/files/{path}"].Delete.OperationID,
	}
	want := map[string]string{
		"get /users":           "getUsers",
		"post /users":          "postUsers",
		"get /users/{id}":      "getUsersById",
		"delete /files/{path}": "deleteFilesByPath",
	}
	for route, wantID := range want {
		if got[route] != wantID {
			t.Errorf("%s operationId = %q, want %q", route, got[route], wantID)
		}
	}

	// Every operation must have a unique, non-empty operationId.
	seen := make(map[string]string)
	for path, item := range doc.Paths {
		for method, op := range operationsOf(item) {
			if op.OperationID == "" {
				t.Errorf("%s %s: empty operationId", method, path)
				continue
			}
			if prev, dup := seen[op.OperationID]; dup {
				t.Errorf("duplicate operationId %q (%s and %s %s)", op.OperationID, prev, method, path)
			}
			seen[op.OperationID] = method + " " + path
		}
	}
}

// TestOpenAPI_OperationIDUniqueness verifies explicit operationIds are
// preserved and generated ones avoid collisions with them.
func TestOpenAPI_OperationIDUniqueness(t *testing.T) {
	router := New()
	router.HandleWithOptions("GET", "/x", func(_ *Context) error { return nil }, &RouteOptions{
		OperationID: "getUsersById",
	})
	router.GET[Empty, testUser]("/users/:id", func(_ *Box[Empty, testUser]) error { return nil })

	doc, err := router.GenerateOpenAPI(Info{Title: "Test", Version: "1.0.0"})
	if err != nil {
		t.Fatalf("GenerateOpenAPI failed: %v", err)
	}

	if got := doc.Paths["/x"].Get.OperationID; got != "getUsersById" {
		t.Errorf("explicit operationId = %q, want %q", got, "getUsersById")
	}
	if got := doc.Paths["/users/{id}"].Get.OperationID; got != "getUsersById_2" {
		t.Errorf("colliding generated operationId = %q, want %q", got, "getUsersById_2")
	}
}

// operationsOf returns the non-nil operations of a PathItem keyed by method.
func operationsOf(item PathItem) map[string]*Operation {
	all := map[string]*Operation{
		"GET":     item.Get,
		"POST":    item.Post,
		"PUT":     item.Put,
		"DELETE":  item.Delete,
		"PATCH":   item.Patch,
		"HEAD":    item.Head,
		"OPTIONS": item.Options,
	}
	ops := make(map[string]*Operation, len(all))
	for method, op := range all {
		if op != nil {
			ops[method] = op
		}
	}
	return ops
}
