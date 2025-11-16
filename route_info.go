// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import "reflect"

// RouteInfo stores metadata about a registered route.
// This information is used for OpenAPI generation, documentation, and introspection.
type RouteInfo struct {
	// Method is the HTTP method (GET, POST, etc.).
	Method string

	// Path is the route path (e.g., "/users/:id").
	Path string

	// Summary is a short description of the operation.
	Summary string

	// Description is a detailed description of the operation.
	Description string

	// Tags are categories for documentation grouping.
	Tags []string

	// OperationID is a unique identifier for the operation.
	OperationID string

	// Deprecated: indicates if this route is deprecated.
	Deprecated bool

	// RequestType is the Go type for the request body (if any).
	RequestType reflect.Type

	// ResponseType is the Go type for the response body (if any).
	ResponseType reflect.Type

	// Parameters stores metadata about path/query/header parameters.
	Parameters []RouteParameter

	// Responses stores metadata about possible responses.
	Responses map[int]RouteResponse
}

// RouteParameter stores metadata about a route parameter.
type RouteParameter struct {
	// Name of the parameter.
	Name string

	// In specifies the location: "path", "query", "header", "cookie".
	In string

	// Description of the parameter.
	Description string

	// Required indicates if the parameter is required.
	Required bool

	// Type is the Go type of the parameter.
	Type reflect.Type
}

// RouteResponse stores metadata about a response.
type RouteResponse struct {
	// Description of the response.
	Description string

	// Type is the Go type of the response body.
	Type reflect.Type

	// ContentType is the media type (e.g., "application/json").
	ContentType string
}

// RouteOptions allows configuring route metadata when registering a route.
type RouteOptions struct {
	// Summary is a short description of the operation.
	Summary string

	// Description is a detailed description of the operation.
	Description string

	// Tags are categories for documentation grouping.
	Tags []string

	// OperationID is a unique identifier for the operation.
	OperationID string

	// Deprecated: indicates if this route is deprecated.
	Deprecated bool

	// Parameters stores metadata about path/query/header parameters.
	Parameters []RouteParameter

	// Responses stores metadata about possible responses.
	Responses map[int]RouteResponse
}
