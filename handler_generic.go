// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

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
//	router.POST[CreateUserRequest, UserResponse]("/users", createUser)
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
