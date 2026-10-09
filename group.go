// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

// RouteGroup represents a group of routes that share the same path prefix and middleware.
// Groups allow organizing routes hierarchically and applying middleware to specific route sets.
//
// Example:
//
//	api := router.Group("/api")
//	api.Use(AuthMiddleware())
//
//	v1 := api.Group("/v1")
//	v1.Handle("GET", "/users", listUsers)      // GET /api/v1/users
//	v1.Handle("POST", "/users", createUser)    // POST /api/v1/users
//
//	v2 := api.Group("/v2")
//	v2.Handle("GET", "/users", listUsersV2)    // GET /api/v2/users
type RouteGroup struct {
	// prefix is the path prefix for all routes in this group.
	prefix string

	// router is a reference to the parent router.
	router *Router

	// middleware stores group-specific middleware.
	// These are combined with router middleware when registering routes.
	middleware []HandlerFunc
}

// Use registers middleware to the route group.
// Group middleware is executed after router middleware but before route handlers.
//
// Middleware order: Router.Use() → Group.Use() → Handler
//
// Example:
//
//	api := router.Group("/api")
//	api.Use(LoggerMiddleware())
//	api.Use(AuthMiddleware())
//
// Can be chained:
//
//	api.Use(Logger()).Use(Auth())
func (g *RouteGroup) Use(middleware ...HandlerFunc) *RouteGroup {
	g.middleware = append(g.middleware, middleware...)
	return g
}

// Group creates a new nested route group with the given prefix and optional middleware.
// The new group's prefix is the combination of the parent prefix and the new prefix.
// If no middleware is provided, the new group inherits the parent's middleware.
//
// Example:
//
//	api := router.Group("/api")
//	api.Use(LoggerMiddleware())
//
//	v1 := api.Group("/v1")              // Inherits logger
//	v1.Use(AuthMiddleware())            // Adds auth
//	v1.Handle("GET", "/users", handler)           // GET /api/v1/users (logger + auth)
//
//	v2 := api.Group("/v2", RateLimitMiddleware())  // Custom middleware
//	v2.Handle("GET", "/users", handler)           // GET /api/v2/users (ratelimit only)
func (g *RouteGroup) Group(prefix string, middleware ...HandlerFunc) *RouteGroup {
	// If no middleware provided, inherit from parent group
	// Always inherit parent middleware, then append child-specific.
	groupMiddleware := make([]HandlerFunc, len(g.middleware), len(g.middleware)+len(middleware))
	copy(groupMiddleware, g.middleware)
	groupMiddleware = append(groupMiddleware, middleware...)

	return &RouteGroup{
		prefix:     g.prefix + prefix,
		router:     g.router,
		middleware: groupMiddleware,
	}
}

// GET registers a type-safe GET route on the group.
// Type parameters are inferred from the handler signature.
//
// An optional *RouteOptions may be supplied to document the route for OpenAPI.
func (g *RouteGroup) GET[Req, Res any](path string, handler Handler[Req, Res], opts ...*RouteOptions) {
	g.registerGeneric("GET", path, handler, firstRouteOptions(opts))
}

// POST registers a type-safe POST route on the group.
//
// An optional *RouteOptions may be supplied to document the route for OpenAPI.
func (g *RouteGroup) POST[Req, Res any](path string, handler Handler[Req, Res], opts ...*RouteOptions) {
	g.registerGeneric("POST", path, handler, firstRouteOptions(opts))
}

// PUT registers a type-safe PUT route on the group.
//
// An optional *RouteOptions may be supplied to document the route for OpenAPI.
func (g *RouteGroup) PUT[Req, Res any](path string, handler Handler[Req, Res], opts ...*RouteOptions) {
	g.registerGeneric("PUT", path, handler, firstRouteOptions(opts))
}

// DELETE registers a type-safe DELETE route on the group.
//
// An optional *RouteOptions may be supplied to document the route for OpenAPI.
func (g *RouteGroup) DELETE[Req, Res any](path string, handler Handler[Req, Res], opts ...*RouteOptions) {
	g.registerGeneric("DELETE", path, handler, firstRouteOptions(opts))
}

// PATCH registers a type-safe PATCH route on the group.
//
// An optional *RouteOptions may be supplied to document the route for OpenAPI.
func (g *RouteGroup) PATCH[Req, Res any](path string, handler Handler[Req, Res], opts ...*RouteOptions) {
	g.registerGeneric("PATCH", path, handler, firstRouteOptions(opts))
}

// HEAD registers a type-safe HEAD route on the group.
//
// An optional *RouteOptions may be supplied to document the route for OpenAPI.
func (g *RouteGroup) HEAD[Req, Res any](path string, handler Handler[Req, Res], opts ...*RouteOptions) {
	g.registerGeneric("HEAD", path, handler, firstRouteOptions(opts))
}

// OPTIONS registers a type-safe OPTIONS route on the group.
//
// An optional *RouteOptions may be supplied to document the route for OpenAPI.
func (g *RouteGroup) OPTIONS[Req, Res any](path string, handler Handler[Req, Res], opts ...*RouteOptions) {
	g.registerGeneric("OPTIONS", path, handler, firstRouteOptions(opts))
}

// registerGeneric registers a type-safe handler on the group, recording the
// Req/Res body types as route metadata for OpenAPI generation.
func (g *RouteGroup) registerGeneric[Req, Res any](method, path string, handler Handler[Req, Res], opts *RouteOptions) {
	fullPath := g.prefix + path
	groupHandlers := g.combineMiddleware(adaptGenericHandler(handler))
	g.router.handleWithGroupMiddleware(method, fullPath, groupHandlers, opts,
		genericBodyType[Req](), genericBodyType[Res]())
}

// Handle registers a route with the given HTTP method, path, and handler.
// This is the core method used by all HTTP method shortcuts (GET, POST, etc.).
//
// The final route path is: group.prefix + path
// The final middleware chain is: router.middleware + group.middleware + handler
//
// Example:
//
//	api := router.Group("/api")
//	api.Handle("GET", "/users", handler)  // Registers GET /api/users
func (g *RouteGroup) Handle(method, path string, handler HandlerFunc) {
	g.HandleWithOptions(method, path, handler, nil)
}

// HandleWithOptions registers a route with documentation metadata for OpenAPI
// generation, in addition to the group prefix and middleware.
//
// Example:
//
//	api := router.Group("/api")
//	api.HandleWithOptions("GET", "/users/:id", handler, &RouteOptions{
//	    Summary: "Get user by ID",
//	    Tags:    []string{"users"},
//	})
func (g *RouteGroup) HandleWithOptions(method, path string, handler HandlerFunc, opts *RouteOptions) {
	// Combine group prefix with route path
	fullPath := g.prefix + path

	// Combine group middleware + handler into a slice
	groupHandlers := g.combineMiddleware(handler)

	// Register route on parent router with group handlers
	// The router will combine its own middleware with these handlers in ServeHTTP
	g.router.handleWithGroupMiddleware(method, fullPath, groupHandlers, opts, nil, nil)
}

// combineMiddleware combines group middleware and the handler.
// Returns a slice of handlers ready to be merged with router middleware.
//
// Order: group.middleware → handler.
func (g *RouteGroup) combineMiddleware(handler HandlerFunc) []HandlerFunc {
	// If group has no middleware, return just the handler
	if len(g.middleware) == 0 {
		return []HandlerFunc{handler}
	}

	// Combine group middleware + handler
	combined := make([]HandlerFunc, len(g.middleware)+1)
	copy(combined, g.middleware)
	combined[len(g.middleware)] = handler

	return combined
}
