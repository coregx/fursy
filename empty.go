// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package fursy provides the Empty type for type-safe handlers without request/response bodies.
package fursy

// Empty represents the absence of a request or response body in type-safe handlers.
//
// Use Empty when a handler doesn't need to bind a request body or send a response body.
// This maintains type safety while indicating "no data" intent.
//
// Examples:
//
//	// Handler with no request body, returns string
//	router.GET[Empty, string]("/hello", func(c *Box[Empty, string]) error {
//	    c.ResBody = new(string)
//	    *c.ResBody = "Hello, World!"
//	    return c.OK(*c.ResBody)
//	})
//
//	// Handler with request body, no response body
//	router.POST[CreateUserRequest, Empty]("/users", func(c *Box[CreateUserRequest, Empty]) error {
//	    user := c.ReqBody
//	    db.CreateUser(user)
//	    return c.NoContent(201)
//	})
//
//	// Handler with neither request nor response body
//	router.DELETE[Empty, Empty]("/cache", func(c *Box[Empty, Empty]) error {
//	    cache.Clear()
//	    return c.NoContent(204)
//	})
type Empty struct{}
