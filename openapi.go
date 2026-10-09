// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// OpenAPI schema type constants.
const (
	schemaTypeString  = "string"
	schemaTypeInteger = "integer"
	schemaTypeObject  = "object"
)

// OpenAPI document constants.
const (
	openapiVersion             = "3.1.0"
	mimeApplicationProblemJSON = "application/problem+json"
	schemaRefPrefix            = "#/components/schemas/"
	refProblemSchema           = schemaRefPrefix + "Problem"
	descSuccess                = "Success"
	descCreated                = "Created"
	descAccepted               = "Accepted"
	descNoContent              = "No Content"
	descBadRequest             = "Bad Request"
	descInternalServerError    = "Internal Server Error"
	paramInPath                = "path"
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

// schemaRegistry builds JSON Schemas for a single OpenAPI document.
//
// Named package types are registered once in components.schemas and referenced
// with $ref, which removes duplication and lets recursive types terminate.
// When defs is nil, components are disabled and every schema is inlined (used
// by generateSchema for standalone introspection).
type schemaRegistry struct {
	// defs maps component name -> schema. Nil disables component generation.
	defs map[string]*Schema
	// nameByType caches the component name assigned to each type.
	nameByType map[reflect.Type]string
	// nameOwner tracks which type owns each component name, for disambiguation.
	nameOwner map[string]reflect.Type
	// seen tracks structs currently being inlined, to break inline cycles.
	seen map[reflect.Type]bool
}

// newSchemaRegistry creates a registry that emits named component schemas.
func newSchemaRegistry() *schemaRegistry {
	reg := &schemaRegistry{
		defs:       make(map[string]*Schema),
		nameByType: make(map[reflect.Type]string),
		nameOwner:  make(map[string]reflect.Type),
		seen:       make(map[reflect.Type]bool),
	}
	// Reserve the built-in schema name so a user type named "Problem" is
	// disambiguated rather than overwriting it.
	reg.nameOwner["Problem"] = nil
	return reg
}

// generateSchema generates an inline JSON Schema from a Go type using reflection.
//
// Named types are expanded inline (no components); use a schemaRegistry for
// $ref-based generation.
func generateSchema(t reflect.Type) *Schema {
	return (&schemaRegistry{seen: make(map[reflect.Type]bool)}).schemaFor(t)
}

// schemaFor returns a schema for t. Named package types yield a $ref to a
// component (registering the definition on first use); all other types are
// expanded inline, recursing through schemaFor so nested named types become
// refs as well.
func (r *schemaRegistry) schemaFor(t reflect.Type) *Schema {
	if t == nil {
		// A nil type carries no shape information; an unconstrained schema is
		// more accurate than asserting an object.
		return &Schema{}
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// Special well-known types must be handled before the component branch,
	// otherwise they would be registered as named components with the wrong body.
	if t == reflect.TypeFor[time.Time]() {
		return &Schema{Type: schemaTypeString, Format: "date-time"}
	}
	if t == reflect.TypeFor[json.RawMessage]() {
		// Raw JSON carries no fixed shape; an unconstrained schema is correct.
		return &Schema{}
	}

	if r.isComponentType(t) {
		name := r.nameFor(t)
		if _, defined := r.defs[name]; !defined {
			// Reserve the name before building so recursive references resolve
			// to this component instead of recursing forever.
			r.defs[name] = &Schema{Type: schemaTypeObject}
			r.defs[name] = r.build(t)
		}
		return &Schema{Ref: schemaRefPrefix + name}
	}

	return r.build(t)
}

// isComponentType reports whether t should be emitted as a named component.
//
// Only user-defined (package-qualified) named types qualify: predeclared types
// such as int or string (empty package path) and anonymous types (empty name)
// are inlined.
func (r *schemaRegistry) isComponentType(t reflect.Type) bool {
	return r.defs != nil && t.Name() != "" && t.PkgPath() != ""
}

// nameFor returns the component name for t, disambiguating same-named types
// from different packages with a package qualifier.
func (r *schemaRegistry) nameFor(t reflect.Type) string {
	if name, ok := r.nameByType[t]; ok {
		return name
	}

	name := t.Name()
	if owner, taken := r.nameOwner[name]; taken && owner != t {
		pkg := t.PkgPath()
		if i := strings.LastIndexByte(pkg, '/'); i >= 0 {
			pkg = pkg[i+1:]
		}
		base := pkg + "." + t.Name()
		name = base
		for i := 2; ; i++ {
			if owner, taken := r.nameOwner[name]; !taken || owner == t {
				break
			}
			name = fmt.Sprintf("%s.%d", base, i)
		}
	}

	r.nameOwner[name] = t
	r.nameByType[t] = name
	return name
}

// build constructs the schema body for t without applying a top-level $ref.
func (r *schemaRegistry) build(t reflect.Type) *Schema {
	// Inline cycle guard: if this type is already being expanded on the current
	// path (possible when components are disabled), stop with a placeholder.
	if r.seen[t] {
		return &Schema{Type: schemaTypeObject}
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
		// []byte (and named types like `type Blob []byte`) encode as base64 strings.
		if t.Elem().Kind() == reflect.Uint8 {
			schema.Type = schemaTypeString
			schema.Format = "byte"
			break
		}
		schema.Type = "array"
		schema.Items = r.schemaFor(t.Elem())
	case reflect.Map:
		schema.Type = schemaTypeObject
		schema.AdditionalProperties = r.schemaFor(t.Elem())
	case reflect.Interface:
		// any / interface{} accepts any JSON value, so no constraint is correct.
		// Leave the schema empty rather than claiming a fixed object type.
		return schema
	case reflect.Struct:
		schema.Type = schemaTypeObject
		schema.Properties = make(map[string]*Schema)

		// Mark this struct as in-progress so self-referential fields terminate.
		r.seen[t] = true
		defer delete(r.seen, t)

		if required := r.buildStructFields(schema.Properties, t); len(required) > 0 {
			schema.Required = required
		}
	default:
		// Unknown type - use generic object.
		schema.Type = schemaTypeObject
	}

	return schema
}

// buildStructFields populates props with the JSON properties of struct type t
// and returns the names of fields that are required.
//
// Embedded fields are promoted (flattened) unless they carry an explicit JSON
// name; promoted fields are handled before the exported check because an
// embedded field reports IsExported()==false when its own type name is
// unexported, while encoding/json still promotes its exported fields.
//
//nolint:gocognit // Field promotion and JSON tag parsing are inherently branchy.
func (r *schemaRegistry) buildStructFields(props map[string]*Schema, t reflect.Type) []string {
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		if field.Anonymous {
			embeddedType := field.Type
			for embeddedType.Kind() == reflect.Pointer {
				embeddedType = embeddedType.Elem()
			}
			// A tag such as json:"base" keeps the field as a nested object,
			// matching encoding/json; otherwise the fields are promoted.
			if hasJSONName(jsonTag) {
				name := jsonFieldName(jsonTag, embeddedType.Name())
				props[name] = r.schemaFor(field.Type)
				if !jsonOmitempty(jsonTag) && field.Type.Kind() != reflect.Pointer {
					required = append(required, name)
				}
				continue
			}
			promoted := r.build(embeddedType)
			for name, prop := range promoted.Properties {
				props[name] = prop
			}
			required = append(required, promoted.Required...)
			continue
		}

		if !field.IsExported() {
			continue
		}

		name := jsonFieldName(jsonTag, field.Name)
		props[name] = r.schemaFor(field.Type)
		if !jsonOmitempty(jsonTag) && field.Type.Kind() != reflect.Pointer {
			required = append(required, name)
		}
	}

	return required
}

// hasJSONName reports whether a JSON tag specifies an explicit name.
//
// A tag of "-" or "" has no name; "name" or "name,omitempty" does. This
// matches when encoding/json promotes an anonymous field: only untagged
// (or option-only) embedded fields are flattened into the parent object.
func hasJSONName(tag string) bool {
	if tag == "" || tag == "-" {
		return false
	}
	name, _, _ := strings.Cut(tag, ",")
	return name != ""
}

// jsonFieldName returns the effective JSON name for a tagged field, falling
// back to fallback when the tag carries no name (e.g. only options).
func jsonFieldName(tag, fallback string) string {
	if name, _, _ := strings.Cut(tag, ","); name != "" {
		return name
	}
	return fallback
}

// jsonOmitempty reports whether a JSON tag includes the omitempty option.
func jsonOmitempty(tag string) bool {
	_, opts, _ := strings.Cut(tag, ",")
	for _, opt := range strings.Split(opts, ",") {
		if opt == "omitempty" {
			return true
		}
	}
	return false
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
		OpenAPI: openapiVersion,
		Info:    info,
		Paths:   make(map[string]PathItem),
		Components: &Components{
			Schemas: make(map[string]*Schema),
			Responses: map[string]Response{
				"Problem": {
					Description: "RFC 9457 Problem Details",
					Content: map[string]MediaType{
						mimeApplicationProblemJSON: {
							Schema: &Schema{
								Ref: refProblemSchema,
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
		Type:        schemaTypeObject,
		Title:       "Problem Details",
		Description: "RFC 9457 Problem Details for HTTP APIs",
		Properties: map[string]*Schema{
			fieldType: {
				Type:        schemaTypeString,
				Description: "URI reference identifying the problem type",
				Default:     defaultProblemType,
			},
			fieldTitle: {
				Type:        schemaTypeString,
				Description: "Short, human-readable summary of the problem type",
			},
			fieldStatus: {
				Type:        schemaTypeInteger,
				Description: "HTTP status code",
			},
			fieldDetail: {
				Type:        schemaTypeString,
				Description: "Human-readable explanation specific to this occurrence",
			},
			fieldInstance: {
				Type:        schemaTypeString,
				Description: "URI reference identifying the specific occurrence",
			},
		},
		Required: []string{fieldType, fieldTitle, fieldStatus},
	}

	// Schema registry collects named component schemas ($ref targets).
	reg := newSchemaRegistry()

	// usedOperationIDs tracks operationIds already assigned so generated ones
	// stay globally unique within the document.
	usedOperationIDs := make(map[string]bool)

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

		// Assign an operationId: use the explicit value if provided, otherwise
		// derive a unique one from the method and path.
		if operation.OperationID == "" {
			operation.OperationID = uniqueOperationID(usedOperationIDs, route.Method, route.Path)
		} else {
			usedOperationIDs[operation.OperationID] = true
		}

		// Add parameters declared via RouteOptions.
		declaredPathParams := make(map[string]bool)
		for _, param := range route.Parameters {
			operation.Parameters = append(operation.Parameters, Parameter{
				Name:        param.Name,
				In:          param.In,
				Description: param.Description,
				Required:    param.Required,
				Schema:      reg.schemaFor(param.Type),
			})
			if param.In == paramInPath {
				declaredPathParams[param.Name] = true
			}
		}

		// Auto-declare path template parameters (e.g. /users/{id}) that were
		// not explicitly provided, so the document is valid OpenAPI.
		for _, name := range extractPathParams(openAPIPath) {
			if declaredPathParams[name] {
				continue
			}
			operation.Parameters = append(operation.Parameters, Parameter{
				Name:     name,
				In:       "path",
				Required: true,
				Schema:   &Schema{Type: schemaTypeString},
			})
		}

		// Add request body if RequestType is set.
		if route.RequestType != nil {
			schema := reg.schemaFor(route.RequestType)
			operation.RequestBody = &RequestBody{
				Required: !route.OptionalRequestBody,
				Content: map[string]MediaType{
					MIMEApplicationJSON: {
						Schema: schema,
					},
				},
			}
		}

		// Always add the inferred success response, using SuccessStatus (0 means 200).
		status := route.SuccessStatus
		if status == 0 {
			status = http.StatusOK
		}
		response := Response{Description: successDescription(status)}
		if status != http.StatusNoContent && route.ResponseType != nil {
			response.Content = map[string]MediaType{
				MIMEApplicationJSON: {
					Schema: reg.schemaFor(route.ResponseType),
				},
			}
		}
		operation.Responses[fmt.Sprintf("%d", status)] = response

		// Explicit RouteOptions.Responses overwrite the inferred response by status
		// code, so a custom error keeps the inferred success schema intact.
		if len(route.Responses) > 0 {
			addExplicitResponses(operation.Responses, route.Responses, reg)
		}

		// Add default error responses, unless the user already supplied them.
		if _, exists := operation.Responses["400"]; !exists {
			operation.Responses["400"] = Response{
				Description: descBadRequest,
				Content: map[string]MediaType{
					mimeApplicationProblemJSON: {
						Schema: &Schema{Ref: refProblemSchema},
					},
				},
			}
		}
		if _, exists := operation.Responses["500"]; !exists {
			operation.Responses["500"] = Response{
				Description: descInternalServerError,
				Content: map[string]MediaType{
					mimeApplicationProblemJSON: {
						Schema: &Schema{Ref: refProblemSchema},
					},
				},
			}
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

	// Register inferred component schemas collected while processing routes.
	for name, schema := range reg.defs {
		doc.Components.Schemas[name] = schema
	}

	return doc, nil
}

// successDescription returns the standard OpenAPI response description for a
// success status code.
func successDescription(status int) string {
	switch status {
	case http.StatusCreated:
		return descCreated
	case http.StatusAccepted:
		return descAccepted
	case http.StatusNoContent:
		return descNoContent
	default:
		return descSuccess
	}
}

// responseContentType resolves the media type for an explicit RouteResponse.
//
// An empty ContentType would otherwise produce an invalid OpenAPI document with
// an empty media-type key, so errors default to application/problem+json and
// everything else to application/json.
func responseContentType(status int, contentType string) string {
	if contentType != "" {
		return contentType
	}
	if status >= http.StatusBadRequest {
		return mimeApplicationProblemJSON
	}
	return MIMEApplicationJSON
}

// addExplicitResponses records user-supplied responses verbatim: their content
// is emitted only when the response declares a body type or media type.
func addExplicitResponses(dst map[string]Response, responses map[int]RouteResponse, reg *schemaRegistry) {
	for status, resp := range responses {
		response := Response{Description: resp.Description}
		if resp.Type != nil || resp.ContentType != "" {
			response.Content = map[string]MediaType{
				responseContentType(status, resp.ContentType): {
					Schema: reg.schemaFor(resp.Type),
				},
			}
		}
		dst[fmt.Sprintf("%d", status)] = response
	}
}

// buildOperationID derives a deterministic operationId from the HTTP method
// and route path, e.g. GET /users/:id -> getUsersById.
func buildOperationID(method, path string) string {
	var b strings.Builder
	b.WriteString(strings.ToLower(method))

	for _, seg := range strings.Split(path, "/") {
		if seg == "" {
			continue
		}
		switch seg[0] {
		case ':', '*':
			b.WriteString("By")
			b.WriteString(exportableName(seg[1:]))
		default:
			b.WriteString(exportableName(seg))
		}
	}

	if b.Len() == 0 {
		return strings.ToLower(method)
	}
	return b.String()
}

// uniqueOperationID returns a unique operationId for the route, appending a
// numeric suffix on collision and recording the result in used.
func uniqueOperationID(used map[string]bool, method, path string) string {
	base := buildOperationID(method, path)
	id := base
	for i := 2; used[id]; i++ {
		id = fmt.Sprintf("%s_%d", base, i)
	}
	used[id] = true
	return id
}

// exportableName converts a path segment into an exported-style identifier,
// upper-casing the first letter and any letter following a separator.
//
// Examples: "users" -> "Users"; "user-names" -> "UserNames"; "id" -> "Id".
func exportableName(s string) string {
	var b strings.Builder
	upperNext := true
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if upperNext {
				b.WriteRune(unicode.ToUpper(r))
				upperNext = false
			} else {
				b.WriteRune(r)
			}
		default:
			upperNext = true
		}
	}
	return b.String()
}

// extractPathParams returns the names of path template parameters in an
// OpenAPI path, in order of appearance.
//
// Example: "/users/{id}/posts/{postID}" -> ["id", "postID"].
func extractPathParams(path string) []string {
	var params []string
	for {
		start := strings.IndexByte(path, '{')
		if start == -1 {
			break
		}
		end := strings.IndexByte(path[start+1:], '}')
		if end == -1 {
			break
		}
		name := path[start+1 : start+1+end]
		if name != "" {
			params = append(params, name)
		}
		path = path[start+1+end+1:]
	}
	return params
}

// convertPathToOpenAPI converts FURSY path format to OpenAPI format.
// /users/:id -> /users/{id}
// /files/*path -> /files/{path}.
//
//nolint:gocritic // if-else chain is clearer than switch for path parsing.
func convertPathToOpenAPI(path string) string {
	result := strings.Builder{}
	i := 0
	for i < len(path) {
		if path[i] == ':' && (i == 0 || path[i-1] == '/') {
			// Named parameter at segment start: :id -> {id}
			// Mid-segment colons are literal (Google AIP-136 custom verbs).
			result.WriteByte('{')
			i++
			start := i
			for i < len(path) && path[i] != '/' {
				i++
			}
			result.WriteString(path[start:i])
			result.WriteByte('}')
		} else if path[i] == '*' && (i == 0 || path[i-1] == '/') {
			// Wildcard parameter at segment start: *path -> {path}
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

// convertPathToOpenAPI converts FURSY path format to OpenAPI format.
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
