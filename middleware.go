// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

// HandlerFunc is the function signature for simple non-generic handlers and middleware.
// It receives a Context and returns an error.
//
// Example handler:
//
//	func GetUser(c *Context) error {
//	    id := c.Param("id")
//	    return c.JSON(200, map[string]string{"id": id})
//	}
//
// Example middleware:
//
//	func Logger() HandlerFunc {
//	    return func(c *Context) error {
//	        start := time.Now()
//	        err := c.Next() // Call next handler
//	        log.Printf("%s %s - %v", c.Request.Method, c.Request.URL.Path, time.Since(start))
//	        return err
//	    }
//	}
//
// For type-safe generic handlers with automatic request/response binding,
// use Handler[Req, Res] instead, which receives Box[Req, Res].
type HandlerFunc func(*Context) error

// Middleware is a function that wraps a HandlerFunc.
// This is an optional pattern - middleware can also be written as HandlerFunc directly.
//
// Example:
//
//	func MyMiddleware() Middleware {
//	    return func(next HandlerFunc) HandlerFunc {
//	        return func(c *Context) error {
//	            // Before handler
//	            err := next(c)
//	            // After handler
//	            return err
//	        }
//	    }
//	}
type Middleware func(HandlerFunc) HandlerFunc
