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
//	v1.GET("/users", listUsers)      // GET /api/v1/users
//	v1.POST("/users", createUser)    // POST /api/v1/users
//
//	v2 := api.Group("/v2")
//	v2.GET("/users", listUsersV2)    // GET /api/v2/users
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
//	v1.GET("/users", handler)           // GET /api/v1/users (logger + auth)
//
//	v2 := api.Group("/v2", RateLimitMiddleware())  // Custom middleware
//	v2.GET("/users", handler)           // GET /api/v2/users (ratelimit only)
func (g *RouteGroup) Group(prefix string, middleware ...HandlerFunc) *RouteGroup {
	// If no middleware provided, inherit from parent group
	var groupMiddleware []HandlerFunc
	if len(middleware) == 0 {
		// Copy parent middleware to avoid shared slice issues
		groupMiddleware = make([]HandlerFunc, len(g.middleware))
		copy(groupMiddleware, g.middleware)
	} else {
		groupMiddleware = middleware
	}

	return &RouteGroup{
		prefix:     g.prefix + prefix,
		router:     g.router,
		middleware: groupMiddleware,
	}
}

// GET registers a GET route on the group.
//
// Example:
//
//	api := router.Group("/api")
//	api.GET("/users", func(c *Box) error {
//	    return c.JSON(200, users)
//	})
func (g *RouteGroup) GET(path string, handler HandlerFunc) {
	g.Handle("GET", path, handler)
}

// POST registers a POST route on the group.
//
// Example:
//
//	api := router.Group("/api")
//	api.POST("/users", func(c *Box) error {
//	    return c.JSON(201, newUser)
//	})
func (g *RouteGroup) POST(path string, handler HandlerFunc) {
	g.Handle("POST", path, handler)
}

// PUT registers a PUT route on the group.
//
// Example:
//
//	api := router.Group("/api")
//	api.PUT("/users/:id", func(c *Box) error {
//	    return c.JSON(200, updatedUser)
//	})
func (g *RouteGroup) PUT(path string, handler HandlerFunc) {
	g.Handle("PUT", path, handler)
}

// DELETE registers a DELETE route on the group.
//
// Example:
//
//	api := router.Group("/api")
//	api.DELETE("/users/:id", func(c *Box) error {
//	    return c.NoContent(204)
//	})
func (g *RouteGroup) DELETE(path string, handler HandlerFunc) {
	g.Handle("DELETE", path, handler)
}

// PATCH registers a PATCH route on the group.
//
// Example:
//
//	api := router.Group("/api")
//	api.PATCH("/users/:id", func(c *Box) error {
//	    return c.JSON(200, patchedUser)
//	})
func (g *RouteGroup) PATCH(path string, handler HandlerFunc) {
	g.Handle("PATCH", path, handler)
}

// HEAD registers a HEAD route on the group.
//
// Example:
//
//	api := router.Group("/api")
//	api.HEAD("/users/:id", func(c *Box) error {
//	    return c.NoContent(200)
//	})
func (g *RouteGroup) HEAD(path string, handler HandlerFunc) {
	g.Handle("HEAD", path, handler)
}

// OPTIONS registers an OPTIONS route on the group.
//
// Example:
//
//	api := router.Group("/api")
//	api.OPTIONS("/users", func(c *Box) error {
//	    c.SetHeader("Allow", "GET, POST")
//	    return c.NoContent(200)
//	})
func (g *RouteGroup) OPTIONS(path string, handler HandlerFunc) {
	g.Handle("OPTIONS", path, handler)
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
	// Combine group prefix with route path
	fullPath := g.prefix + path

	// Combine group middleware + handler into a slice
	groupHandlers := g.combineMiddleware(handler)

	// Register route on parent router with group handlers
	// The router will combine its own middleware with these handlers in ServeHTTP
	g.router.handleWithGroupMiddleware(method, fullPath, groupHandlers)
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
