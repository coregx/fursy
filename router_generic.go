// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import "net/http"

// GET registers a type-safe handler for GET requests to the specified path.
//
// Since Go doesn't support generic methods, we use top-level functions instead.
// This provides clean type-safe routing with automatic request binding.
//
// Type parameters:
//   - Req: The expected request body type (use Empty for GET requests)
//   - Res: The response body type
//
// Example:
//
//	type UserResponse struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	}
//
//	fursy.GET[fursy.Empty, UserResponse](router, "/users/:id", func(c *fursy.Box[fursy.Empty, UserResponse]) error {
//	    id := c.Param("id")
//	    user := db.GetUser(id)
//	    return c.OK(UserResponse{ID: user.ID, Name: user.Name})
//	})
func GET[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodGet, path, adaptGenericHandler(handler))
}

// POST registers a type-safe handler for POST requests to the specified path.
//
// The handler will receive a Box[Req, Res] with automatically bound request body.
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `json:"name" validate:"required"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	type UserResponse struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	}
//
//	fursy.POST[CreateUserRequest, UserResponse](router, "/users", func(c *fursy.Box[CreateUserRequest, UserResponse]) error {
//	    req := c.ReqBody
//	    user := db.CreateUser(req.Name, req.Email)
//	    return c.Created("/users/"+user.ID, UserResponse{ID: user.ID, Name: user.Name})
//	})
func POST[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodPost, path, adaptGenericHandler(handler))
}

// PUT registers a type-safe handler for PUT requests to the specified path.
//
// The handler will receive a Box[Req, Res] with automatically bound request body.
//
// Example:
//
//	type UpdateUserRequest struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	fursy.PUT[UpdateUserRequest, UserResponse](router, "/users/:id", func(c *fursy.Box[UpdateUserRequest, UserResponse]) error {
//	    id := c.Param("id")
//	    req := c.ReqBody
//	    user := db.UpdateUser(id, req.Name, req.Email)
//	    return c.OK(UserResponse{ID: user.ID, Name: user.Name})
//	})
func PUT[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodPut, path, adaptGenericHandler(handler))
}

// DELETE registers a type-safe handler for DELETE requests to the specified path.
//
// The handler will receive a Box[Req, Res] with automatically bound request body.
//
// Example:
//
//	fursy.DELETE[fursy.Empty, fursy.Empty](router, "/users/:id", func(c *fursy.Box[fursy.Empty, fursy.Empty]) error {
//	    id := c.Param("id")
//	    db.DeleteUser(id)
//	    return c.NoContent(204)
//	})
func DELETE[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodDelete, path, adaptGenericHandler(handler))
}

// PATCH registers a type-safe handler for PATCH requests to the specified path.
//
// The handler will receive a Box[Req, Res] with automatically bound request body.
//
// Example:
//
//	type PatchUserRequest struct {
//	    Name  *string `json:"name,omitempty"`
//	    Email *string `json:"email,omitempty"`
//	}
//
//	fursy.PATCH[PatchUserRequest, UserResponse](router, "/users/:id", func(c *fursy.Box[PatchUserRequest, UserResponse]) error {
//	    id := c.Param("id")
//	    req := c.ReqBody
//	    user := db.PatchUser(id, req)
//	    return c.OK(UserResponse{ID: user.ID, Name: user.Name})
//	})
func PATCH[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodPatch, path, adaptGenericHandler(handler))
}

// HEAD registers a type-safe handler for HEAD requests to the specified path.
//
// The handler will receive a Box[Req, Res] with automatically bound request body.
// HEAD requests should not return a body, only headers.
//
// Example:
//
//	fursy.HEAD[fursy.Empty, fursy.Empty](router, "/users/:id", func(c *fursy.Box[fursy.Empty, fursy.Empty]) error {
//	    id := c.Param("id")
//	    if db.UserExists(id) {
//	        return c.NoContent(200)
//	    }
//	    return c.NoContent(404)
//	})
func HEAD[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodHead, path, adaptGenericHandler(handler))
}

// OPTIONS registers a type-safe handler for OPTIONS requests to the specified path.
//
// The handler will receive a Box[Req, Res] with automatically bound request body.
// OPTIONS requests are typically used for CORS preflight checks.
//
// Example:
//
//	fursy.OPTIONS[fursy.Empty, fursy.Empty](router, "/users", func(c *fursy.Box[fursy.Empty, fursy.Empty]) error {
//	    c.SetHeader("Allow", "GET, POST, PUT, DELETE")
//	    c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
//	    return c.NoContent(200)
//	})
func OPTIONS[Req, Res any](r *Router, path string, handler Handler[Req, Res]) {
	r.Handle(http.MethodOptions, path, adaptGenericHandler(handler))
}
