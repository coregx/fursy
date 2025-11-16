// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package fursy provides Context types for HTTP request handling.
package fursy

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/coregx/fursy/internal/negotiate"
)

// NOTE: This package provides two context types:
//   - Context: For simple non-generic handlers (HandlerFunc)
//   - Box[Req, Res]: For type-safe generic handlers (Handler[Req, Res])
//
// Example simple handler:
//
//	func GetUser(c *fursy.Context) error {
//		id := c.Param("id")
//		return c.JSON(200, map[string]string{"id": id})
//	}
//
// Example generic handler:
//
//	type UserRequest struct { Name string }
//	type UserResponse struct { ID int; Name string }
//	func CreateUser(c *fursy.Box[UserRequest, UserResponse]) error {
//		c.ResBody = &UserResponse{ID: 1, Name: c.ReqBody.Name}
//		return c.OK()
//	}

// Param represents a URL parameter extracted from the path.
//
// Example:
//
//	For route "/users/:id" and path "/users/123":
//	Param{Key: "id", Value: "123"}
type Param struct {
	Key   string // Parameter name (e.g., "id" from /:id)
	Value string // Parameter value extracted from path
}

// Context is the base context for all handlers and middleware.
// It provides access to request/response, routing info, and middleware chain execution.
//
// Context is designed to be embedded in higher-level context types (like generic Box[Req, Res])
// while providing all the core functionality needed for middleware and handlers.
//
// The context is pooled and reused across requests for zero allocations.
// Do not store Context references - all operations must complete within the handler's execution.
type Context struct {
	// Request is the current HTTP request.
	Request *http.Request

	// Response is the response writer.
	Response http.ResponseWriter

	// router reference for accessing router configuration.
	router *Router

	// params stores URL parameters extracted from the route.
	// Pre-allocated with capacity 8 to avoid allocations for typical routes.
	params []Param

	// query is a lazy-loaded cache of parsed query parameters.
	query map[string][]string

	// data stores arbitrary values for passing data between middleware.
	data map[string]any

	// Middleware chain execution.
	// Pre-allocated with capacity 16 to avoid allocations for typical middleware chains.
	handlers []HandlerFunc
	index    int
	aborted  bool
}

const (
	// maxParamsCapacity is the maximum capacity for params slice before reallocation.
	// If params grow beyond this, we create a new slice to avoid holding large buffers.
	maxParamsCapacity = 32

	// maxHandlersCapacity is the maximum capacity for handlers slice before reallocation.
	// If handlers grow beyond this, we create a new slice to avoid holding large buffers.
	maxHandlersCapacity = 64
)

// newContext creates a new Context instance.
// This is called by the Router's sync.Pool.
func newContext() *Context {
	return &Context{
		data:     make(map[string]any),
		params:   make([]Param, 0, 8),        // Pre-allocate params buffer (typical: 1-4 params).
		handlers: make([]HandlerFunc, 0, 16), // Pre-allocate handlers buffer (typical: 3-8 middleware).
	}
}

// init initializes the context with request/response for a new request.
// This is called by Router.ServeHTTP before executing the handler chain.
func (c *Context) init(w http.ResponseWriter, r *http.Request, router *Router, params []Param) {
	c.Request = r
	c.Response = w
	c.router = router
	c.params = params
	c.query = nil // reset query cache
	// data map is reused (cleared in reset)
}

// reset clears the context for reuse in the pool.
// This is called by Router.ServeHTTP after the handler chain completes.
func (c *Context) reset() {
	c.Request = nil
	c.Response = nil
	c.router = nil
	c.query = nil

	// Reset params slice: keep capacity if reasonable, otherwise reallocate.
	// This prevents memory leaks from holding large backing arrays.
	if cap(c.params) > maxParamsCapacity {
		// Capacity grew too large, allocate new buffer.
		c.params = make([]Param, 0, 8)
	} else {
		// Reuse buffer, reset length only (keep capacity).
		c.params = c.params[:0]
	}

	// Clear data map but keep allocation.
	for k := range c.data {
		delete(c.data, k)
	}

	// Reset handlers slice: keep capacity if reasonable, otherwise reallocate.
	if cap(c.handlers) > maxHandlersCapacity {
		// Capacity grew too large, allocate new buffer.
		c.handlers = make([]HandlerFunc, 0, 16)
	} else {
		// Reuse buffer, reset length only (keep capacity).
		c.handlers = c.handlers[:0]
	}

	c.index = -1
	c.aborted = false
}

// Next executes the next handler in the middleware chain.
// It returns the error from the handler, allowing middleware to handle or transform errors.
//
// Example middleware:
//
//	func Logger() HandlerFunc {
//	    return func(c *Context) error {
//	        start := time.Now()
//	        err := c.Next()  // Call next handler
//	        log.Printf("%s - %v", c.Request.URL.Path, time.Since(start))
//	        return err
//	    }
//	}
func (c *Context) Next() error {
	c.index++
	if c.index < len(c.handlers) && !c.aborted {
		return c.handlers[c.index](c)
	}
	return nil
}

// Abort prevents pending handlers from being called.
// Note that this does not stop the current handler - it only prevents subsequent handlers in the chain.
//
// This is useful when a middleware wants to stop the chain (e.g., authentication failure).
//
// Example:
//
//	func RequireAuth() HandlerFunc {
//	    return func(c *Context) error {
//	        if !isAuthenticated(c) {
//	            c.Abort()
//	            return c.JSON(401, map[string]string{"error": "unauthorized"})
//	        }
//	        return c.Next()
//	    }
//	}
func (c *Context) Abort() {
	c.aborted = true
}

// IsAborted returns true if the context was aborted.
// This can be used to check if a middleware called Abort().
func (c *Context) IsAborted() bool {
	return c.aborted
}

// Router returns the router instance that is handling this request.
// This can be used to access router configuration or state.
func (c *Context) Router() *Router {
	return c.router
}

// Param returns the value of the URL parameter by name.
// Returns empty string if the parameter doesn't exist.
//
// Example:
//
//	// Route: /users/:id
//	// Request: /users/123
//	id := c.Param("id") // "123"
func (c *Context) Param(name string) string {
	for _, p := range c.params {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

// Query returns the first value for the named query parameter.
// Returns empty string if the parameter doesn't exist.
//
// The query string is parsed lazily on first access and cached for subsequent calls.
//
// Example:
//
//	// Request: /users?page=2&limit=10
//	page := c.Query("page")   // "2"
//	limit := c.Query("limit") // "10"
func (c *Context) Query(name string) string {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	values := c.query[name]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// QueryDefault returns the query parameter value or a default value.
// If the parameter doesn't exist or is empty, returns defaultValue.
//
// Example:
//
//	// Request: /users?page=2
//	page := c.QueryDefault("page", "1")   // "2"
//	limit := c.QueryDefault("limit", "10") // "10" (default)
func (c *Context) QueryDefault(name, defaultValue string) string {
	value := c.Query(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// QueryValues returns all values for the named query parameter.
// Returns nil if the parameter doesn't exist.
//
// Example:
//
//	// Request: /search?tag=go&tag=web&tag=api
//	tags := c.QueryValues("tag") // []string{"go", "web", "api"}
func (c *Context) QueryValues(name string) []string {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query[name]
}

// Form returns the first value for the named form parameter.
// It checks both POST/PUT body parameters and URL query parameters.
// Form parameters take precedence over query parameters.
//
// For multipart forms, it parses up to 32MB of data.
//
// Example:
//
//	// POST /login with body: username=john&password=secret
//	username := c.Form("username") // "john"
func (c *Context) Form(name string) string {
	if c.Request.Form == nil {
		_ = c.Request.ParseMultipartForm(32 << 20) // 32 MB - error ignored as FormValue handles it
	}
	return c.Request.FormValue(name)
}

// FormDefault returns the form parameter value or a default value.
// If the parameter doesn't exist or is empty, returns defaultValue.
//
// Example:
//
//	role := c.FormDefault("role", "user") // "user" if not provided
func (c *Context) FormDefault(name, defaultValue string) string {
	value := c.Form(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// PostForm returns the form value from POST/PUT body only (not URL query).
// Unlike Form(), this does not fall back to query parameters.
//
// Example:
//
//	// POST /update?id=123 with body: name=john
//	name := c.PostForm("name") // "john"
//	id := c.PostForm("id")     // "" (not in POST body)
func (c *Context) PostForm(name string) string {
	if c.Request.PostForm == nil {
		_ = c.Request.ParseMultipartForm(32 << 20) // 32 MB - error ignored as PostFormValue handles it
	}
	return c.Request.PostFormValue(name)
}

// String sends a plain text response.
//
// Example:
//
//	return c.String(200, "Hello, World!")
func (c *Context) String(code int, s string) error {
	c.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response.WriteHeader(code)
	_, err := c.Response.Write([]byte(s))
	return err
}

// JSON sends a JSON response.
// The obj is encoded using encoding/json and sent with application/json content type.
//
// Example:
//
//	return c.JSON(200, map[string]string{"message": "success"})
func (c *Context) JSON(code int, obj any) error {
	c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response.WriteHeader(code)
	encoder := json.NewEncoder(c.Response)
	return encoder.Encode(obj)
}

// JSONIndent sends a JSON response with indentation for pretty-printing.
// This is useful for debugging or human-readable responses.
//
// Example:
//
//	return c.JSONIndent(200, data, "  ") // 2-space indent
func (c *Context) JSONIndent(code int, obj any, indent string) error {
	c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response.WriteHeader(code)
	encoder := json.NewEncoder(c.Response)
	encoder.SetIndent("", indent)
	return encoder.Encode(obj)
}

// XML sends an XML response.
// The obj is encoded using encoding/xml and sent with application/xml content type.
//
// Example:
//
//	type User struct {
//		XMLName xml.Name `xml:"user"`
//		ID      string   `xml:"id"`
//		Name    string   `xml:"name"`
//	}
//	return c.XML(200, User{ID: "123", Name: "John"})
func (c *Context) XML(code int, obj any) error {
	c.Response.Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Response.WriteHeader(code)
	encoder := xml.NewEncoder(c.Response)
	return encoder.Encode(obj)
}

// NoContent sends a response with no body.
// This is commonly used for 204 No Content responses.
//
// Example:
//
//	return c.NoContent(204) // Successful deletion
func (c *Context) NoContent(code int) error {
	c.Response.WriteHeader(code)
	return nil
}

// Redirect sends an HTTP redirect response.
// The code must be in the 3xx range (300-308).
//
// Common redirect codes:
//   - 301: Moved Permanently
//   - 302: Found (temporary redirect)
//   - 303: See Other
//   - 307: Temporary Redirect (preserves method)
//   - 308: Permanent Redirect (preserves method)
//
// Example:
//
//	return c.Redirect(302, "/login")
func (c *Context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	http.Redirect(c.Response, c.Request, url, code)
	return nil
}

// Blob sends a binary response with custom content type.
// This is useful for sending raw binary data like images or files.
//
// Example:
//
//	imageData := []byte{...}
//	return c.Blob(200, "image/png", imageData)
func (c *Context) Blob(code int, contentType string, data []byte) error {
	c.Response.Header().Set("Content-Type", contentType)
	c.Response.WriteHeader(code)
	_, err := c.Response.Write(data)
	return err
}

// Stream sends a response from an io.Reader.
// This is useful for streaming large files or data without loading everything into memory.
//
// Example:
//
//	file, _ := os.Open("large-file.pdf")
//	defer file.Close()
//	return c.Stream(200, "application/pdf", file)
func (c *Context) Stream(code int, contentType string, r io.Reader) error {
	c.Response.Header().Set("Content-Type", contentType)
	c.Response.WriteHeader(code)
	_, err := io.Copy(c.Response, r)
	return err
}

// ========================================
// Convenience Response Methods
// ========================================
//
// These methods provide shortcuts for the most common HTTP response patterns,
// reducing boilerplate while maintaining clarity about the response status code.
// For custom status codes, use the explicit methods above (JSON, String, etc.).

// OK sends a 200 OK JSON response.
// This is a convenience method for the most common success case.
//
// Use this for successful GET requests or operations that return data.
//
// Example:
//
//	router.GET("/users", func(c *fursy.Context) error {
//	    users := getAllUsers()
//	    return c.OK(users)  // 200 OK
//	})
func (c *Context) OK(obj any) error {
	return c.JSON(200, obj)
}

// Created sends a 201 Created JSON response.
// Use this for successful POST requests that create a new resource.
//
// REST best practice: POST operations that create resources should return 201, not 200.
//
// Example:
//
//	router.POST("/users", func(c *fursy.Context) error {
//	    newUser := createUser(c)
//	    return c.Created(newUser)  // 201 Created
//	})
func (c *Context) Created(obj any) error {
	return c.JSON(201, obj)
}

// Accepted sends a 202 Accepted JSON response.
// Use this when the request has been accepted for processing but not completed.
//
// Common for async operations, background jobs, or queued tasks.
//
// Example:
//
//	router.POST("/jobs", func(c *fursy.Context) error {
//	    jobID := startAsyncJob(c)
//	    return c.Accepted(map[string]string{"jobId": jobID})  // 202 Accepted
//	})
func (c *Context) Accepted(obj any) error {
	return c.JSON(202, obj)
}

// NoContentSuccess sends a 204 No Content response.
// This is a convenience method for successful operations with no response body.
//
// Common for DELETE operations and some PUT/PATCH updates.
// REST best practice: DELETE should return 204, not 200.
//
// Example:
//
//	router.DELETE("/users/:id", func(c *fursy.Context) error {
//	    deleteUser(c.Param("id"))
//	    return c.NoContentSuccess()  // 204 No Content
//	})
func (c *Context) NoContentSuccess() error {
	return c.NoContent(204)
}

// Text sends a 200 OK plain text response.
// This is a convenience method for simple text responses.
//
// Example:
//
//	router.GET("/ping", func(c *fursy.Context) error {
//	    return c.Text("pong")  // 200 OK, text/plain
//	})
func (c *Context) Text(s string) error {
	return c.String(200, s)
}

// SetHeader sets a response header.
// This must be called before writing the response body.
//
// Example:
//
//	c.SetHeader("X-Request-ID", requestID)
//	return c.String(200, "OK")
func (c *Context) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

// GetHeader returns a request header value.
// Returns empty string if the header doesn't exist.
//
// Example:
//
//	userAgent := c.GetHeader("User-Agent")
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// Get retrieves data from the context.
// Returns nil if the key doesn't exist.
//
// This is useful for passing data between middleware and handlers.
//
// Example:
//
//	// In authentication middleware:
//	c.Set("userID", "123")
//
//	// In handler:
//	userID := c.Get("userID").(string)
func (c *Context) Get(key string) any {
	return c.data[key]
}

// Set stores data in the context.
// This is useful for passing data between middleware and handlers.
//
// Example:
//
//	c.Set("userID", "123")
//	c.Set("authenticated", true)
func (c *Context) Set(key string, value any) {
	c.data[key] = value
}

// GetString retrieves a string value from the context.
// Returns empty string if the key doesn't exist or value is not a string.
//
// Example:
//
//	userID := c.GetString("userID")
func (c *Context) GetString(key string) string {
	if v, ok := c.data[key].(string); ok {
		return v
	}
	return ""
}

// GetInt retrieves an int value from the context.
// Returns 0 if the key doesn't exist or value is not an int.
//
// Example:
//
//	page := c.GetInt("page")
func (c *Context) GetInt(key string) int {
	if v, ok := c.data[key].(int); ok {
		return v
	}
	return 0
}

// GetBool retrieves a bool value from the context.
// Returns false if the key doesn't exist or value is not a bool.
//
// Example:
//
//	authenticated := c.GetBool("authenticated")
func (c *Context) GetBool(key string) bool {
	if v, ok := c.data[key].(bool); ok {
		return v
	}
	return false
}

// Problem sends an RFC 9457 Problem Details response.
//
// Problem Details (RFC 9457) provides a standard way to carry machine-readable
// details of errors in HTTP responses, with Content-Type: application/problem+json.
//
// Example:
//
//	return c.Problem(fursy.NotFound("User not found"))
//
//	return c.Problem(fursy.UnprocessableEntity("Invalid input").
//	    WithExtension("field", "email").
//	    WithExtension("reason", "already exists"))
//
// For validation errors, use ValidationProblem:
//
//	if err := c.Bind(); err != nil {
//	    if verr, ok := err.(ValidationErrors); ok {
//	        return c.Problem(ValidationProblem(verr))
//	    }
//	    return c.Problem(BadRequest(err.Error()))
//	}
func (c *Context) Problem(p Problem) error {
	// Set proper Content-Type for RFC 9457.
	c.Response.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	c.Response.WriteHeader(p.Status)
	encoder := json.NewEncoder(c.Response)
	return encoder.Encode(p)
}

// NegotiateFormat returns the best offered content type based on the Accept header.
//
// This method performs RFC 9110 compliant content negotiation, including:
//   - Quality value (q) weighting
//   - Specificity matching (explicit > wildcard)
//   - Parameter precedence
//
// Returns the selected content type, or an empty string if no match found.
//
// Example:
//
//	format := c.NegotiateFormat(fursy.MIMEApplicationJSON, fursy.MIMETextHTML, fursy.MIMEApplicationXML)
//	switch format {
//	case fursy.MIMEApplicationJSON:
//	    return c.JSON(200, data)
//	case fursy.MIMETextHTML:
//	    return c.HTML(200, "template", data)
//	case fursy.MIMEApplicationXML:
//	    return c.XML(200, data)
//	default:
//	    return c.Problem(NotAcceptable("No acceptable format found"))
//	}
func (c *Context) NegotiateFormat(offered ...string) string {
	if len(offered) == 0 {
		return ""
	}

	accept := c.Request.Header.Get("Accept")
	if accept == "" {
		// No Accept header - return first offered.
		return offered[0]
	}

	return negotiate.ContentType(accept, offered)
}

// Negotiate performs content negotiation and sends the response in the best format.
//
// This is a convenience method that combines NegotiateFormat with automatic response
// rendering. It automatically sets the Vary: Accept header for proper HTTP caching.
//
// Supported formats:
//   - application/json (JSON)
//   - application/xml, text/xml (XML)
//   - text/html (HTML - requires HTMLData and HTMLTemplate)
//   - text/plain (Plain text)
//
// Returns ErrNotAcceptable if no acceptable format is found.
//
// Example:
//
//	type User struct {
//	    ID   int    `json:"id" xml:"id"`
//	    Name string `json:"name" xml:"name"`
//	}
//
//	user := User{ID: 1, Name: "John"}
//	return c.Negotiate(200, user)
//	// Client with "Accept: application/json" receives JSON
//	// Client with "Accept: application/xml" receives XML
func (c *Context) Negotiate(status int, data any) error {
	// Set Vary: Accept for proper caching.
	c.SetHeader("Vary", "Accept")

	// Determine offered formats (common formats).
	offered := []string{MIMEApplicationJSON, MIMEApplicationXML, MIMETextXML, MIMETextPlain}

	format := c.NegotiateFormat(offered...)
	if format == "" {
		return c.Problem(NotAcceptable("No acceptable content type available"))
	}

	// Render based on negotiated format.
	switch format {
	case MIMEApplicationJSON:
		return c.JSON(status, data)
	case MIMEApplicationXML, MIMETextXML:
		return c.XML(status, data)
	case MIMETextPlain:
		// Plain text - use string representation.
		return c.String(status, fmt.Sprintf("%v", data))
	default:
		return c.Problem(InternalServerError("Unsupported content type: " + format))
	}
}

// NotAcceptable creates a 406 Not Acceptable Problem.
//
// RFC 9110 Section 15.5.7: The 406 Not Acceptable status code indicates that
// the target resource does not have a current representation that would be
// acceptable to the user agent, according to the proactive negotiation header
// fields received in the request, and the server is unwilling to supply a
// default representation.
func NotAcceptable(detail string) Problem {
	return NewProblem(406, "Not Acceptable", detail)
}

// ErrInvalidRedirectCode is returned when redirect code is not 3xx.
var ErrInvalidRedirectCode = errors.New("fursy: invalid redirect code (must be 3xx)")
