// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// OpenAPI schema type constants.
const (
	schemaTypeString  = "string"
	schemaTypeInteger = "integer"
	schemaTypeObject  = "object"
)

// OpenAPI represents an OpenAPI 3.1 document.
//
// This is the root object of the OpenAPI Description.
// It provides metadata about the API and describes the available endpoints.
//
// Spec: https://spec.openapis.org/oas/v3.1.1.html
type OpenAPI struct {
	// OpenAPI version string (e.g., "3.1.0").
	OpenAPI string `json:"openapi"`

	// Info provides metadata about the API.
	Info Info `json:"info"`

	// Servers is an array of Server objects providing connectivity information.
	Servers []Server `json:"servers,omitempty"`

	// Paths holds the available paths and operations for the API.
	Paths map[string]PathItem `json:"paths"`

	// Components holds reusable objects for different aspects of the OAS.
	Components *Components `json:"components,omitempty"`

	// Security is a declaration of which security mechanisms can be used across the API.
	Security []SecurityRequirement `json:"security,omitempty"`

	// Tags is a list of tags used by the document with additional metadata.
	Tags []Tag `json:"tags,omitempty"`
}

// Info provides metadata about the API.
type Info struct {
	// Title of the API (required).
	Title string `json:"title"`

	// Version of the OpenAPI document (required).
	Version string `json:"version"`

	// Summary is a short summary of the API.
	Summary string `json:"summary,omitempty"`

	// Description is a description of the API (supports CommonMark).
	Description string `json:"description,omitempty"`

	// Contact information for the exposed API.
	Contact *Contact `json:"contact,omitempty"`

	// License information for the exposed API.
	License *License `json:"license,omitempty"`

	// TermsOfService is a URL to the Terms of Service for the API.
	TermsOfService string `json:"termsOfService,omitempty"`
}

// Contact information for the exposed API.
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License information for the exposed API.
type License struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier,omitempty"`
	URL        string `json:"url,omitempty"`
}

// Server represents a server.
type Server struct {
	// URL to the target host.
	URL string `json:"url"`

	// Description is an optional string describing the host.
	Description string `json:"description,omitempty"`

	// Variables for server URL template substitution.
	Variables map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable represents a server variable for template substitution.
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// PathItem describes operations available on a single path.
type PathItem struct {
	Summary     string      `json:"summary,omitempty"`
	Description string      `json:"description,omitempty"`
	Get         *Operation  `json:"get,omitempty"`
	Post        *Operation  `json:"post,omitempty"`
	Put         *Operation  `json:"put,omitempty"`
	Delete      *Operation  `json:"delete,omitempty"`
	Patch       *Operation  `json:"patch,omitempty"`
	Head        *Operation  `json:"head,omitempty"`
	Options     *Operation  `json:"options,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty"`
}

// Operation describes a single API operation on a path.
type Operation struct {
	// Tags for API documentation control.
	Tags []string `json:"tags,omitempty"`

	// Summary is a short summary of what the operation does.
	Summary string `json:"summary,omitempty"`

	// Description is a verbose explanation of the operation behavior.
	Description string `json:"description,omitempty"`

	// OperationID is a unique string used to identify the operation.
	OperationID string `json:"operationId,omitempty"`

	// Parameters that are applicable for this operation.
	Parameters []Parameter `json:"parameters,omitempty"`

	// RequestBody applicable for this operation.
	RequestBody *RequestBody `json:"requestBody,omitempty"`

	// Responses is the list of possible responses.
	Responses map[string]Response `json:"responses"`

	// Deprecated declares this operation to be deprecated.
	Deprecated bool `json:"deprecated,omitempty"`

	// Security is a declaration of which security mechanisms can be used.
	Security []SecurityRequirement `json:"security,omitempty"`
}

// Parameter describes a single operation parameter.
type Parameter struct {
	// Name of the parameter (required).
	Name string `json:"name"`

	// In is the location of the parameter (required).
	// Possible values: "query", "header", "path", "cookie".
	In string `json:"in"`

	// Description of the parameter.
	Description string `json:"description,omitempty"`

	// Required determines whether this parameter is mandatory.
	// Must be true if the parameter location is "path".
	Required bool `json:"required,omitempty"`

	// Deprecated specifies that a parameter is deprecated.
	Deprecated bool `json:"deprecated,omitempty"`

	// Schema defining the type used for the parameter.
	Schema *Schema `json:"schema,omitempty"`
}

// RequestBody describes a single request body.
type RequestBody struct {
	// Description of the request body.
	Description string `json:"description,omitempty"`

	// Content is a map of media types to media type objects.
	Content map[string]MediaType `json:"content"`

	// Required determines if the request body is required.
	Required bool `json:"required,omitempty"`
}

// Response describes a single response from an API operation.
type Response struct {
	// Description of the response (required).
	Description string `json:"description"`

	// Content is a map of media types to media type objects.
	Content map[string]MediaType `json:"content,omitempty"`

	// Headers is a map of response headers.
	Headers map[string]Header `json:"headers,omitempty"`
}

// MediaType provides schema and examples for the media type.
type MediaType struct {
	// Schema defining the content of the request, response, or parameter.
	Schema *Schema `json:"schema,omitempty"`

	// Example of the media type.
	Example any `json:"example,omitempty"`

	// Examples of the media type.
	Examples map[string]Example `json:"examples,omitempty"`
}

// Example represents an example value.
type Example struct {
	Summary       string `json:"summary,omitempty"`
	Description   string `json:"description,omitempty"`
	Value         any    `json:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty"`
}

// Header represents a header parameter.
type Header struct {
	Description string  `json:"description,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Deprecated  bool    `json:"deprecated,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// Schema represents a data type schema.
// This is compatible with JSON Schema Draft 2020-12.
type Schema struct {
	// Type specifies the data type.
	Type string `json:"type,omitempty"`

	// Format provides additional type information.
	Format string `json:"format,omitempty"`

	// Title of the schema.
	Title string `json:"title,omitempty"`

	// Description of the schema.
	Description string `json:"description,omitempty"`

	// Properties for object types (property name -> schema).
	Properties map[string]*Schema `json:"properties,omitempty"`

	// Required lists required properties for object types.
	Required []string `json:"required,omitempty"`

	// Items schema for array types.
	Items *Schema `json:"items,omitempty"`

	// Enum restricts values to a specific set.
	Enum []any `json:"enum,omitempty"`

	// Default value.
	Default any `json:"default,omitempty"`

	// Example value.
	Example any `json:"example,omitempty"`

	// Nullable indicates if the value can be null.
	Nullable bool `json:"nullable,omitempty"`

	// ReadOnly indicates the property is read-only.
	ReadOnly bool `json:"readOnly,omitempty"`

	// WriteOnly indicates the property is write-only.
	WriteOnly bool `json:"writeOnly,omitempty"`

	// Ref is a reference to another schema.
	Ref string `json:"$ref,omitempty"`

	// AdditionalProperties for object types.
	AdditionalProperties any `json:"additionalProperties,omitempty"`

	// OneOf specifies that the value must match exactly one schema.
	OneOf []*Schema `json:"oneOf,omitempty"`

	// AnyOf specifies that the value must match at least one schema.
	AnyOf []*Schema `json:"anyOf,omitempty"`

	// AllOf specifies that the value must match all schemas.
	AllOf []*Schema `json:"allOf,omitempty"`
}

// Components holds reusable objects.
type Components struct {
	// Schemas is a map of reusable Schema objects.
	Schemas map[string]*Schema `json:"schemas,omitempty"`

	// Responses is a map of reusable Response objects.
	Responses map[string]Response `json:"responses,omitempty"`

	// Parameters is a map of reusable Parameter objects.
	Parameters map[string]Parameter `json:"parameters,omitempty"`

	// RequestBodies is a map of reusable RequestBody objects.
	RequestBodies map[string]RequestBody `json:"requestBodies,omitempty"`

	// Headers is a map of reusable Header objects.
	Headers map[string]Header `json:"headers,omitempty"`

	// SecuritySchemes is a map of reusable SecurityScheme objects.
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme defines a security scheme.
type SecurityScheme struct {
	Type             string      `json:"type"`
	Description      string      `json:"description,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows represents OAuth flows.
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow represents an OAuth flow.
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// SecurityRequirement represents a security requirement.
type SecurityRequirement map[string][]string

// Tag represents a tag with metadata.
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// generateSchema generates a JSON Schema from a Go type using reflection.
//
//nolint:gocognit,gocyclo,cyclop // Schema generation requires complex type introspection.
func generateSchema(t reflect.Type) *Schema {
	// Handle pointer types.
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := &Schema{}

	switch t.Kind() {
	case reflect.String:
		schema.Type = schemaTypeString
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema.Type = schemaTypeInteger
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = schemaTypeInteger
		schema.Format = "uint"
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
	case reflect.Bool:
		schema.Type = "boolean"
	case reflect.Slice, reflect.Array:
		schema.Type = "array"
		schema.Items = generateSchema(t.Elem())
	case reflect.Map:
		schema.Type = schemaTypeObject
		schema.AdditionalProperties = generateSchema(t.Elem())
	case reflect.Struct:
		schema.Type = schemaTypeObject
		schema.Properties = make(map[string]*Schema)
		required := []string{}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Skip unexported fields.
			if !field.IsExported() {
				continue
			}

			// Get JSON tag.
			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			// Parse JSON tag.
			fieldName := field.Name
			omitempty := false
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" {
					fieldName = parts[0]
				}
				for _, opt := range parts[1:] {
					if opt == "omitempty" {
						omitempty = true
					}
				}
			}

			// Generate schema for field.
			fieldSchema := generateSchema(field.Type)

			// Add description from comment (if available).
			// Note: We can't easily get comments via reflection.

			schema.Properties[fieldName] = fieldSchema

			// Check if required.
			if !omitempty && field.Type.Kind() != reflect.Ptr {
				required = append(required, fieldName)
			}
		}

		if len(required) > 0 {
			schema.Required = required
		}
	default:
		// Unknown type - use generic object.
		schema.Type = schemaTypeObject
	}

	return schema
}

// GenerateOpenAPI generates an OpenAPI 3.1 document from the router.
//
// This method introspects all registered routes and generates a complete
// OpenAPI 3.1 specification including paths, schemas, and components.
//
// If info is not provided via WithInfo(), the info parameter is used.
//
// Example:
//
//	doc, err := router.GenerateOpenAPI(Info{
//	    Title:   "My API",
//	    Version: "1.0.0",
//	})
//
//nolint:gocognit,gocyclo,cyclop,gocritic,funlen // OpenAPI generation requires complex route introspection.
func (r *Router) GenerateOpenAPI(info Info) (*OpenAPI, error) {
	// Use router info if set, otherwise use parameter.
	if r.info != nil {
		info = *r.info
	}

	doc := &OpenAPI{
		OpenAPI: "3.1.0",
		Info:    info,
		Paths:   make(map[string]PathItem),
		Components: &Components{
			Schemas: make(map[string]*Schema),
			Responses: map[string]Response{
				"Problem": {
					Description: "RFC 9457 Problem Details",
					Content: map[string]MediaType{
						"application/problem+json": {
							Schema: &Schema{
								Ref: "#/components/schemas/Problem",
							},
						},
					},
				},
			},
		},
	}

	// Add servers if configured.
	if len(r.servers) > 0 {
		doc.Servers = r.servers
	}

	// Add RFC 9457 Problem Details schema.
	doc.Components.Schemas["Problem"] = &Schema{
		Type:        "object",
		Title:       "Problem Details",
		Description: "RFC 9457 Problem Details for HTTP APIs",
		Properties: map[string]*Schema{
			"type": {
				Type:        "string",
				Description: "URI reference identifying the problem type",
				Default:     "about:blank",
			},
			"title": {
				Type:        "string",
				Description: "Short, human-readable summary of the problem type",
			},
			"status": {
				Type:        "integer",
				Description: "HTTP status code",
			},
			"detail": {
				Type:        "string",
				Description: "Human-readable explanation specific to this occurrence",
			},
			"instance": {
				Type:        "string",
				Description: "URI reference identifying the specific occurrence",
			},
		},
		Required: []string{"type", "title", "status"},
	}

	// Process all registered routes.
	for _, route := range r.routes {
		// Convert FURSY path format to OpenAPI format.
		// /users/:id -> /users/{id}
		openAPIPath := convertPathToOpenAPI(route.Path)

		// Get or create PathItem for this path.
		pathItem, exists := doc.Paths[openAPIPath]
		if !exists {
			pathItem = PathItem{}
		}

		// Create operation.
		operation := &Operation{
			Summary:     route.Summary,
			Description: route.Description,
			Tags:        route.Tags,
			OperationID: route.OperationID,
			Deprecated:  route.Deprecated,
			Responses:   make(map[string]Response),
		}

		// Add parameters.
		if len(route.Parameters) > 0 {
			for _, param := range route.Parameters {
				operation.Parameters = append(operation.Parameters, Parameter{
					Name:        param.Name,
					In:          param.In,
					Description: param.Description,
					Required:    param.Required,
					Schema:      generateSchema(param.Type),
				})
			}
		}

		// Add request body if RequestType is set.
		if route.RequestType != nil {
			schema := generateSchema(route.RequestType)
			operation.RequestBody = &RequestBody{
				Required: true,
				Content: map[string]MediaType{
					"application/json": {
						Schema: schema,
					},
				},
			}
		}

		// Add responses.
		if len(route.Responses) > 0 {
			for status, resp := range route.Responses {
				statusStr := fmt.Sprintf("%d", status)
				operation.Responses[statusStr] = Response{
					Description: resp.Description,
					Content: map[string]MediaType{
						resp.ContentType: {
							Schema: generateSchema(resp.Type),
						},
					},
				}
			}
		} else {
			// Default responses.
			if route.ResponseType != nil {
				operation.Responses["200"] = Response{
					Description: "Success",
					Content: map[string]MediaType{
						"application/json": {
							Schema: generateSchema(route.ResponseType),
						},
					},
				}
			} else {
				operation.Responses["200"] = Response{
					Description: "Success",
				}
			}
		}

		// Add default error responses.
		operation.Responses["400"] = Response{
			Description: "Bad Request",
			Content: map[string]MediaType{
				"application/problem+json": {
					Schema: &Schema{Ref: "#/components/schemas/Problem"},
				},
			},
		}
		operation.Responses["500"] = Response{
			Description: "Internal Server Error",
			Content: map[string]MediaType{
				"application/problem+json": {
					Schema: &Schema{Ref: "#/components/schemas/Problem"},
				},
			},
		}

		// Assign operation to correct HTTP method.
		switch route.Method {
		case http.MethodGet:
			pathItem.Get = operation
		case http.MethodPost:
			pathItem.Post = operation
		case http.MethodPut:
			pathItem.Put = operation
		case http.MethodDelete:
			pathItem.Delete = operation
		case http.MethodPatch:
			pathItem.Patch = operation
		case http.MethodHead:
			pathItem.Head = operation
		case http.MethodOptions:
			pathItem.Options = operation
		}

		doc.Paths[openAPIPath] = pathItem
	}

	return doc, nil
}

// convertPathToOpenAPI converts FURSY path format to OpenAPI format.
// /users/:id -> /users/{id}
// /files/*path -> /files/{path}.
//
//nolint:gocritic,staticcheck // if-else chain is clearer than switch for path parsing.
func convertPathToOpenAPI(path string) string {
	result := strings.Builder{}
	i := 0
	for i < len(path) {
		if path[i] == ':' {
			// Named parameter: :id -> {id}
			result.WriteByte('{')
			i++
			start := i
			for i < len(path) && path[i] != '/' {
				i++
			}
			result.WriteString(path[start:i])
			result.WriteByte('}')
		} else if path[i] == '*' {
			// Wildcard parameter: *path -> {path}
			result.WriteByte('{')
			i++
			start := i
			for i < len(path) && path[i] != '/' {
				i++
			}
			result.WriteString(path[start:i])
			result.WriteByte('}')
		} else {
			result.WriteByte(path[i])
			i++
		}
	}
	return result.String()
}

// MarshalJSON is implemented by the default json/v2 marshaler.
// This comment exists to document that custom marshaling is not needed.

// WriteJSON writes the OpenAPI document as JSON.
func (doc *OpenAPI) WriteJSON(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Marshal using json/v2 API.
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// WriteYAML writes the OpenAPI document as YAML.
// Note: Requires YAML library (not implemented yet as it's not stdlib).
//
//nolint:revive // Parameter w reserved for future YAML implementation.
func (doc *OpenAPI) WriteYAML(w http.ResponseWriter) error {
	return fmt.Errorf("YAML output not yet implemented (requires external dependency)")
}
