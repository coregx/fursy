// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"net/http"
	"reflect"
)

// Handler is a type-safe handler function for HTTP requests with typed request/response bodies.
//
// Type parameters:
//   - Req: The expected request body type (use Empty if no body)
//   - Res: The response body type (use Empty if no structured response)
//
// The handler receives a Box[Req, Res] with automatically bound request body (ReqBody)
// and provides type-safe methods for sending responses.
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	type UserResponse struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	}
//
//	func createUser(c *Box[CreateUserRequest, UserResponse]) error {
//	    req := c.ReqBody  // Type: *CreateUserRequest
//	    user := db.Create(req.Name, req.Email)
//	    return c.Created("/users/"+user.ID, UserResponse{
//	        ID:   user.ID,
//	        Name: user.Name,
//	    })
//	}
//
//	router.POST("/users", createUser)  // Go 1.27: type parameters inferred from handler
type Handler[Req, Res any] func(*Box[Req, Res]) error

// adaptGenericHandler converts a generic Handler[Req, Res] to a non-generic HandlerFunc.
//
// This adapter:
//  1. Creates a generic Box[Req, Res] from Context
//  2. Binds the request body to Box.ReqBody (based on Content-Type)
//  3. Calls the generic handler
//  4. Returns any error from binding or handler execution
//
// This is used internally by Router.GET, Router.POST, etc. to support generic handlers.
func adaptGenericHandler[Req, Res any](handler Handler[Req, Res]) HandlerFunc {
	return func(base *Context) error {
		// Enforce body size limit if configured.
		// MaxBytesReader wraps the body so that reading beyond the limit
		// returns http.MaxBytesError, which defaultErrorHandler maps to 413.
		if base.router != nil && base.router.maxBodySize > 0 && base.Request.Body != nil {
			base.Request.Body = http.MaxBytesReader(base.Response, base.Request.Body, base.router.maxBodySize)
		}

		// Create generic context
		ctx := newBox[Req, Res](base)

		// Bind request body
		if err := ctx.Bind(); err != nil {
			return err
		}

		// Call generic handler
		return handler(ctx)
	}
}

// genericBodyType returns the reflect.Type for the generic parameter T, or nil
// if T is the Empty sentinel (meaning "no body").
//
// It mirrors the Empty detection performed in Box.Bind so that the route
// metadata recorded at registration matches runtime binding behavior.
func genericBodyType[T any]() reflect.Type {
	var zero T
	if _, ok := any(zero).(Empty); ok {
		return nil
	}
	return reflect.TypeFor[T]()
}
