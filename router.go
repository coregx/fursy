// Package fursy provides a high-performance HTTP router for Go 1.25+.
//
// FURSY combines type-safe routing with modern Go features like generics,
// providing fast URL matching (<100ns), zero dependencies, and clean API.
//
// # Quick Start
//
//	router := fursy.New()
//
//	router.GET("/users/:id", func(c *fursy.Box) error {
//		id := c.Param("id")
//		return c.String(200, "User ID: "+id)
//	})
//
//	http.ListenAndServe(":8080", router)
//
// # Route Types
//
// FURSY supports three types of routes:
//
//   - Static: /users
//   - Parameters: /users/:id
//   - Wildcards: /files/*path
//
// # Performance
//
// FURSY uses a radix tree for routing, providing <100ns lookups
// and zero allocations for simple routes.
//
// # HTTP Methods
//
// All standard HTTP methods are supported:
// GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS.
//
// # URL Parameters
//
// Extract parameters using Box methods:
//
//	id := c.Param("id")
//	page := c.Query("page")
//	username := c.Form("username")
//
// # Error Handling
//
//   - 404 Not Found: Automatic for unregistered routes
//   - 405 Method Not Allowed: Automatic when route exists but method differs
//
// See Router documentation for more details.
package fursy

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/coregx/fursy/internal/radix"
)

// Router is the main HTTP router for FURSY.
// It provides fast URL routing with support for static paths,
// parameters (:id), and wildcards (*path).
//
// Router implements http.Handler and can be used directly with http.ListenAndServe.
//
// Example:
//
//	router := fursy.New()
//	router.GET("/users/:id", func(c *fursy.Box) error {
//		id := c.Param("id")
//		return c.String(200, "User ID: "+id)
//	})
//	http.ListenAndServe(":8080", router)

// Router is the main HTTP router for FURSY.
// It provides fast URL routing with support for static paths,
// parameters (:id), and wildcards (*path).
//
// Router implements http.Handler and can be used directly with http.ListenAndServe.
type Router struct {
	// trees stores one radix tree per HTTP method for efficient routing.
	trees map[string]*radix.Tree

	// pool reuses Context instances across requests for zero allocations.
	pool sync.Pool

	// middleware stores global middleware that executes for all routes.
	middleware []HandlerFunc

	// validator is an optional validator for automatic request validation.
	// If set, Box.Bind() will automatically validate request bodies.
	// Set using Router.SetValidator().
	validator Validator

	// handleMethodNotAllowed enables automatic 405 responses when a route
	// exists for a path but not for the requested HTTP method.
	handleMethodNotAllowed bool

	// handleOPTIONS enables automatic handling of OPTIONS requests.
	handleOPTIONS bool

	// routes stores metadata about all registered routes for OpenAPI generation.
	routes []RouteInfo

	// info stores API metadata for OpenAPI generation.
	info *Info

	// servers stores server information for OpenAPI generation.
	servers []Server

	// server stores reference to http.Server for graceful shutdown.
	// Set by ListenAndServeWithShutdown or manually via SetServer.
	server *http.Server

	// shutdownCallbacks stores functions to call during graceful shutdown.
	// Register callbacks using OnShutdown().
	shutdownCallbacks []func()

	// shutdownMu protects shutdown callbacks from concurrent access.
	shutdownMu sync.Mutex
}

// New creates a new Router instance with default configuration.
//
// The router is created with:
//   - Context pooling enabled for zero allocations
//   - Method Not Allowed handling enabled
//   - OPTIONS handling enabled
//   - Empty routing tables (trees are created on first route registration)
func New() *Router {
	r := &Router{
		trees:                  make(map[string]*radix.Tree),
		handleMethodNotAllowed: true,
		handleOPTIONS:          true,
	}

	// Initialize context pool.
	r.pool.New = func() any {
		c := newContext()
		c.router = r
		return c
	}

	return r
}

// Use registers global middleware that executes for all routes.
// Middleware is executed in the order it is registered.
//
// Middleware can:
//   - Modify the request/response
//   - Call c.Next() to continue the chain
//   - Call c.Abort() to stop the chain
//   - Return an error to propagate to error handler
//
// Example:
//
//	router.Use(Logger())
//	router.Use(Recovery())
//	router.Use(CORS())
//
//	func Logger() HandlerFunc {
//	    return func(c *Context) error {
//	        start := time.Now()
//	        err := c.Next()
//	        log.Printf("%s - %v", c.Request.URL.Path, time.Since(start))
//	        return err
//	    }
//	}
func (r *Router) Use(middleware ...HandlerFunc) *Router {
	r.middleware = append(r.middleware, middleware...)
	return r
}

// SetValidator sets the validator for automatic request validation.
//
// When a validator is set, Box.Bind() will automatically validate
// request bodies after binding. If validation fails, Bind() returns
// a ValidationErrors error.
//
// Validator is optional. If not set, binding works without validation.
//
// Example:
//
//	// Using validator/v10 (requires plugin)
//	import "github.com/coregx/fursy/plugins/validator"
//
//	router := fursy.New()
//	router.SetValidator(validator.New())
//
//	// Now all POST/PUT/PATCH requests will be validated
//	type CreateUserRequest struct {
//	    Email string `json:"email" validate:"required,email"`
//	    Age   int    `json:"age" validate:"gte=18,lte=120"`
//	}
//
//	POST[CreateUserRequest, UserResponse](router, "/users", func(c *Box[CreateUserRequest, UserResponse]) error {
//	    // c.ReqBody is already validated here!
//	    // ...
//	})
func (r *Router) SetValidator(v Validator) *Router {
	r.validator = v
	return r
}

// WithInfo sets the API metadata for OpenAPI generation.
//
// This configures the info section of the generated OpenAPI document.
//
// Example:
//
//	router.WithInfo(Info{
//	    Title:       "My API",
//	    Version:     "1.0.0",
//	    Description: "A sample API built with FURSY",
//	})
func (r *Router) WithInfo(info Info) *Router {
	r.info = &info
	return r
}

// WithServer adds a server to the OpenAPI document.
//
// Servers define the base URLs where the API is deployed.
//
// Example:
//
//	router.WithServer(Server{
//	    URL:         "https://api.example.com",
//	    Description: "Production server",
//	})
func (r *Router) WithServer(server Server) *Router {
	r.servers = append(r.servers, server)
	return r
}

// ServeOpenAPI registers a route that serves the OpenAPI 3.1 specification as JSON.
//
// This is a convenience method that automatically generates and serves the OpenAPI
// document at the specified path. The document is generated on each request, so it
// always reflects the current state of registered routes.
//
// For production use, consider caching the generated document or serving a
// pre-generated specification file.
//
// Example:
//
//	router := fursy.New()
//	router.WithInfo(fursy.Info{
//	    Title:   "My API",
//	    Version: "1.0.0",
//	})
//
//	// Register your routes
//	router.GET("/users", handler)
//	router.POST("/users", handler)
//
//	// Serve OpenAPI specification
//	router.ServeOpenAPI("/openapi.json")
//
//	// Now GET /openapi.json returns the OpenAPI 3.1 document
func (r *Router) ServeOpenAPI(path string) {
	r.GET(path, func(c *Context) error {
		// Use router info if configured, otherwise use minimal defaults.
		info := Info{
			Title:   "API Documentation",
			Version: "1.0.0",
		}
		if r.info != nil {
			info = *r.info
		}

		// Generate OpenAPI document.
		doc, err := r.GenerateOpenAPI(info)
		if err != nil {
			return err
		}

		// Serve as JSON.
		return doc.WriteJSON(c.Response)
	})
}

// Group creates a new route group with the given path prefix and optional middleware.
// Groups allow organizing routes hierarchically and applying middleware to specific route sets.
//
// The group inherits router middleware but can have its own middleware stack.
// Middleware order: Router.Use() → Group.Use() → Handler
//
// Example:
//
//	router := fursy.New()
//	router.Use(LoggerMiddleware())  // Global
//
//	api := router.Group("/api")
//	api.Use(AuthMiddleware())       // API-specific
//
//	v1 := api.Group("/v1")
//	v1.GET("/users", handler)       // GET /api/v1/users (logger + auth)
//
//	admin := api.Group("/admin", AdminMiddleware())  // Custom middleware
//	admin.GET("/settings", handler)  // GET /api/admin/settings (admin only)
func (r *Router) Group(prefix string, middleware ...HandlerFunc) *RouteGroup {
	return &RouteGroup{
		prefix:     prefix,
		router:     r,
		middleware: middleware,
	}
}

// GET registers a handler for GET requests to the specified path.
//
// Example:
//
//	router.GET("/users", func(c *fursy.Box) error {
//		return c.JSON(200, users)
//	})
func (r *Router) GET(path string, handler HandlerFunc) {
	r.Handle(http.MethodGet, path, handler)
}

// POST registers a handler for POST requests to the specified path.
//
// Example:
//
//	router.POST("/users", func(c *fursy.Box) error {
//		return c.JSON(201, newUser)
//	})
func (r *Router) POST(path string, handler HandlerFunc) {
	r.Handle(http.MethodPost, path, handler)
}

// PUT registers a handler for PUT requests to the specified path.
//
// Example:
//
//	router.PUT("/users/:id", func(c *fursy.Box) error {
//		id := c.Param("id")
//		return c.NoContent(204)
//	})
func (r *Router) PUT(path string, handler HandlerFunc) {
	r.Handle(http.MethodPut, path, handler)
}

// DELETE registers a handler for DELETE requests to the specified path.
//
// Example:
//
//	router.DELETE("/users/:id", func(c *fursy.Box) error {
//		id := c.Param("id")
//		return c.NoContent(204)
//	})
func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.Handle(http.MethodDelete, path, handler)
}

// PATCH registers a handler for PATCH requests to the specified path.
//
// Example:
//
//	router.PATCH("/users/:id", func(c *fursy.Box) error {
//		id := c.Param("id")
//		return c.JSON(200, updatedUser)
//	})
func (r *Router) PATCH(path string, handler HandlerFunc) {
	r.Handle(http.MethodPatch, path, handler)
}

// HEAD registers a handler for HEAD requests to the specified path.
//
// Example:
//
//	router.HEAD("/users/:id", func(c *fursy.Box) error {
//		return c.NoContent(200)
//	})
func (r *Router) HEAD(path string, handler HandlerFunc) {
	r.Handle(http.MethodHead, path, handler)
}

// OPTIONS registers a handler for OPTIONS requests to the specified path.
//
// Example:
//
//	router.OPTIONS("/users", func(c *fursy.Box) error {
//		c.SetHeader("Allow", "GET, POST, PUT, DELETE")
//		return c.NoContent(200)
//	})
func (r *Router) OPTIONS(path string, handler HandlerFunc) {
	r.Handle(http.MethodOptions, path, handler)
}

// Handle registers a handler for the given HTTP method and path.
//
// The method must be a valid HTTP method (GET, POST, etc.).
// The path must start with a '/' and can contain:
//   - Static segments: /users
//   - Named parameters: /users/:id
//   - Catch-all parameters: /files/*path
//
// Panics if method or path is empty, or if handler is nil.
//
// Example:
//
//	router.Handle("GET", "/users/:id", func(c *fursy.Box) error {
//		id := c.Param("id")
//		return c.String(200, "User ID: "+id)
//	})
func (r *Router) Handle(method, path string, handler HandlerFunc) {
	r.HandleWithOptions(method, path, handler, nil)
}

// HandleWithOptions registers a handler with route metadata for OpenAPI generation.
//
// This method extends Handle() with support for route documentation.
//
// Example:
//
//	router.HandleWithOptions("GET", "/users/:id", handler, &RouteOptions{
//	    Summary:     "Get user by ID",
//	    Description: "Returns a single user",
//	    Tags:        []string{"users"},
//	})
func (r *Router) HandleWithOptions(method, path string, handler HandlerFunc, opts *RouteOptions) {
	if method == "" {
		panic("fursy: HTTP method cannot be empty")
	}
	if path == "" {
		panic("fursy: path cannot be empty")
	}
	if handler == nil {
		panic("fursy: handler cannot be nil")
	}

	// Get or create tree for this method.
	tree := r.trees[method]
	if tree == nil {
		tree = radix.New()
		r.trees[method] = tree
	}

	// Insert route into radix tree.
	if err := tree.Insert(path, handler); err != nil {
		panic("fursy: " + err.Error())
	}

	// Store route metadata for OpenAPI generation.
	routeInfo := RouteInfo{
		Method: method,
		Path:   path,
	}

	if opts != nil {
		routeInfo.Summary = opts.Summary
		routeInfo.Description = opts.Description
		routeInfo.Tags = opts.Tags
		routeInfo.OperationID = opts.OperationID
		routeInfo.Deprecated = opts.Deprecated
		routeInfo.Parameters = opts.Parameters
		routeInfo.Responses = opts.Responses
	}

	r.routes = append(r.routes, routeInfo)
}

// handleWithGroupMiddleware registers a route with group middleware.
// This is called by RouteGroup.Handle() to register routes with group-specific middleware.
//
// The groupHandlers slice contains: group.middleware + handler
// These will be combined with router.middleware in ServeHTTP.
func (r *Router) handleWithGroupMiddleware(method, path string, groupHandlers []HandlerFunc) {
	if method == "" {
		panic("fursy: HTTP method cannot be empty")
	}
	if path == "" {
		panic("fursy: path cannot be empty")
	}
	if len(groupHandlers) == 0 {
		panic("fursy: groupHandlers cannot be empty")
	}

	// Create a wrapper handler that executes group middleware + handler
	wrapper := r.createGroupHandlerWrapper(groupHandlers)

	// Get or create tree for this method.
	tree := r.trees[method]
	if tree == nil {
		tree = radix.New()
		r.trees[method] = tree
	}

	// Insert route into radix tree with the wrapper.
	if err := tree.Insert(path, wrapper); err != nil {
		panic("fursy: " + err.Error())
	}
}

// createGroupHandlerWrapper creates a handler that executes group middleware + handler.
// This wrapper will be called as part of the router middleware chain.
//
// Execution order in ServeHTTP: router.middleware → wrapper (group.middleware → handler).
func (r *Router) createGroupHandlerWrapper(groupHandlers []HandlerFunc) HandlerFunc {
	return func(c *Context) error {
		// Save current middleware chain state
		savedHandlers := c.handlers
		savedIndex := c.index
		savedAborted := c.aborted

		// Build group middleware chain
		c.handlers = groupHandlers
		c.index = -1
		c.aborted = false

		// Execute group middleware chain
		err := c.Next()

		// Restore router middleware chain state
		c.handlers = savedHandlers
		c.index = savedIndex
		c.aborted = savedAborted

		return err
	}
}

// ServeHTTP implements http.Handler interface, making Router compatible
// with the standard library's http.Server.
//
// It performs the following steps:
//  1. Gets a Context from the pool (zero allocation)
//  2. Looks up the appropriate routing tree for the HTTP method
//  3. Searches for a matching route in the radix tree
//  4. Extracts URL parameters if the route contains wildcards
//  5. Initializes the Context with request/response/params
//  6. Builds middleware chain (global middleware + route handler)
//  7. Executes the middleware chain via c.Next()
//  8. Resets and returns Context to the pool
//
// Returns 404 Not Found if no route matches the path.
// Returns 405 Method Not Allowed if the path exists but for a different method
// (when handleMethodNotAllowed is enabled).
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Get context from pool.
	c := r.pool.Get().(*Context)
	defer func() {
		c.reset()
		r.pool.Put(c)
	}()

	path := req.URL.Path

	// Get tree for this HTTP method.
	tree := r.trees[req.Method]
	if tree == nil {
		if r.handleMethodNotAllowed {
			// Check if path exists in other methods.
			if r.pathExistsInOtherMethods(path, req.Method) {
				c.init(w, req, r, nil)
				_ = c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
				return
			}
		}
		c.init(w, req, r, nil)
		_ = c.String(http.StatusNotFound, "Not Found")
		return
	}

	// Lookup route in radix tree.
	handler, params, found := tree.Lookup(path)
	if !found {
		c.init(w, req, r, nil)
		_ = c.String(http.StatusNotFound, "Not Found")
		return
	}

	// Convert internal params to public Param slice.
	// Reuse pre-allocated params buffer from context (zero allocation).
	c.params = c.params[:0] // Reset length, keep capacity.
	for _, p := range params {
		c.params = append(c.params, Param{Key: p.Key, Value: p.Value})
	}

	// Initialize context.
	c.init(w, req, r, c.params)

	// Build handler chain: middleware + route handler.
	// Reuse pre-allocated handlers buffer from context (zero allocation).
	routeHandler := handler.(HandlerFunc)
	c.handlers = c.handlers[:0] // Reset length, keep capacity.
	c.handlers = append(c.handlers, r.middleware...)
	c.handlers = append(c.handlers, routeHandler)
	c.index = -1
	c.aborted = false

	// Execute middleware chain.
	if err := c.Next(); err != nil {
		// If handler returned an error and hasn't written a response,
		// send a 500 Internal Server Error.
		// In the future, this will call custom ErrorHandler.
		_ = c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

// pathExistsInOtherMethods checks if a path exists in other HTTP methods.
// Used for 405 Method Not Allowed responses.
func (r *Router) pathExistsInOtherMethods(path, method string) bool {
	for m, tree := range r.trees {
		if m != method {
			_, _, found := tree.Lookup(path)
			if found {
				return true
			}
		}
	}
	return false
}

// OnShutdown registers a function to be called during graceful shutdown.
//
// Callbacks are executed in reverse order (last registered, first called)
// before the HTTP server stops accepting new connections.
//
// Use this for cleanup tasks like:
//   - Closing database connections
//   - Flushing logs or metrics
//   - Saving in-memory data
//   - Releasing external resources
//
// OnShutdown is safe for concurrent use.
//
// Example:
//
//	router := fursy.New()
//
//	// Register cleanup callbacks
//	router.OnShutdown(func() {
//	    log.Println("Closing database connections...")
//	    db.Close()
//	})
//
//	router.OnShutdown(func() {
//	    log.Println("Flushing metrics...")
//	    metrics.Flush()
//	})
//
//	// Start server with graceful shutdown
//	if err := router.ListenAndServeWithShutdown(":8080"); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) OnShutdown(f func()) {
	if f == nil {
		return
	}
	r.shutdownMu.Lock()
	defer r.shutdownMu.Unlock()
	r.shutdownCallbacks = append(r.shutdownCallbacks, f)
}

// Shutdown gracefully shuts down the HTTP server and executes registered callbacks.
//
// Shutdown works in two phases:
//  1. Calls all registered OnShutdown callbacks in reverse order
//  2. Calls http.Server.Shutdown() to gracefully stop the server
//
// The server shutdown process:
//   - Immediately closes all listeners (stops accepting new connections)
//   - Waits for active requests to complete (respects context timeout)
//   - Returns context error if timeout is exceeded
//
// Shutdown does NOT forcefully close active connections after timeout.
// If you need to force close, cancel the context.
//
// Safe to call multiple times (subsequent calls are no-ops).
//
// Example:
//
//	router := fursy.New()
//	router.OnShutdown(func() {
//	    log.Println("Cleanup...")
//	})
//
//	// Start server in goroutine
//	srv := &http.Server{Addr: ":8080", Handler: router}
//	router.SetServer(srv)
//
//	go func() {
//	    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
//	        log.Fatal(err)
//	    }
//	}()
//
//	// Handle shutdown signal
//	sigChan := make(chan os.Signal, 1)
//	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
//	<-sigChan
//
//	// Graceful shutdown with 30s timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	if err := router.Shutdown(ctx); err != nil {
//	    log.Printf("Shutdown error: %v", err)
//	}
func (r *Router) Shutdown(ctx context.Context) error {
	// Call shutdown callbacks in reverse order (last registered, first called).
	r.shutdownMu.Lock()
	callbacks := make([]func(), len(r.shutdownCallbacks))
	copy(callbacks, r.shutdownCallbacks)
	r.shutdownMu.Unlock()

	for i := len(callbacks) - 1; i >= 0; i-- {
		callbacks[i]()
	}

	// Shutdown http.Server if configured.
	if r.server != nil {
		return r.server.Shutdown(ctx)
	}

	return nil
}

// SetServer sets the http.Server for graceful shutdown.
//
// This is typically called by ListenAndServeWithShutdown, but can be
// used manually if you create the server yourself.
//
// Example:
//
//	router := fursy.New()
//	srv := &http.Server{
//	    Addr:         ":8080",
//	    Handler:      router,
//	    ReadTimeout:  10 * time.Second,
//	    WriteTimeout: 10 * time.Second,
//	}
//
//	router.SetServer(srv)
//	router.OnShutdown(func() {
//	    log.Println("Server shutting down...")
//	})
//
//	// Now router.Shutdown() will shutdown srv
func (r *Router) SetServer(srv *http.Server) {
	r.server = srv
}

// ListenAndServeWithShutdown starts the HTTP server with automatic graceful shutdown.
//
// This is a convenience method that:
//  1. Creates an http.Server with the given address
//  2. Listens for SIGTERM and SIGINT signals (Kubernetes/Docker compatible)
//  3. Starts the server in a goroutine
//  4. Blocks until shutdown signal is received
//  5. Calls Shutdown() with the specified timeout (default: 30s)
//
// The timeout must be less than Kubernetes terminationGracePeriodSeconds (default 30s)
// to allow time for preStop hooks and connection draining.
//
// Returns:
//   - nil if shutdown completed successfully
//   - http.ErrServerClosed is treated as successful shutdown (not returned)
//   - Error from server startup (e.g., address in use)
//   - Error from Shutdown if timeout exceeded
//
// Example (simple):
//
//	router := fursy.New()
//	router.GET("/health", healthHandler)
//
//	router.OnShutdown(func() {
//	    log.Println("Closing database...")
//	    db.Close()
//	})
//
//	// Blocks until SIGTERM/SIGINT, then graceful shutdown with 30s timeout
//	if err := router.ListenAndServeWithShutdown(":8080"); err != nil {
//	    log.Fatal(err)
//	}
//
// Example (custom timeout):
//
//	// 10 second shutdown timeout
//	if err := router.ListenAndServeWithShutdown(":8080", 10*time.Second); err != nil {
//	    log.Fatal(err)
//	}
//
// Example (Kubernetes-ready):
//
//	router := fursy.New()
//
//	// Health check for readiness probe
//	router.GET("/health", func(c *fursy.Context) error {
//	    return c.String(200, "OK")
//	})
//
//	// Cleanup on shutdown
//	router.OnShutdown(func() {
//	    log.Println("Shutdown initiated...")
//	    db.Close()
//	    cache.Close()
//	    log.Println("Cleanup complete")
//	})
//
//	// Kubernetes sends SIGTERM before killing pod
//	// terminationGracePeriodSeconds: 30s (default)
//	// Our shutdown timeout: 25s (leaves 5s buffer)
//	if err := router.ListenAndServeWithShutdown(":8080", 25*time.Second); err != nil {
//	    log.Fatal(err)
//	}
func (r *Router) ListenAndServeWithShutdown(addr string, timeout ...time.Duration) error {
	// Default timeout: 30s (Kubernetes-compatible).
	shutdownTimeout := 30 * time.Second
	if len(timeout) > 0 && timeout[0] > 0 {
		shutdownTimeout = timeout[0]
	}

	// Create HTTP server.
	srv := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second, // Protection against Slowloris attacks.
	}
	r.SetServer(srv)

	// Create context that cancels on SIGTERM or SIGINT.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Channel to receive server startup errors.
	serverErr := make(chan error, 1)

	// Start server in goroutine.
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal or server error.
	select {
	case err := <-serverErr:
		// Server failed to start.
		return err
	case <-ctx.Done():
		// Shutdown signal received.
		stop() // Stop signal notification.
	}

	// Create shutdown context with timeout.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Gracefully shutdown (calls callbacks + server.Shutdown).
	return r.Shutdown(shutdownCtx)
}
