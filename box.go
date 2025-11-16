// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fursy

import (
	"net/http"

	"github.com/coregx/fursy/internal/binding"
)

// Box is a type-safe context for HTTP request handling with request/response types.
//
// It embeds *Context, providing all base functionality (params, middleware, etc.),
// while adding type-safe request and response body handling.
//
// Type parameters:
//   - Req: The expected request body type (use Empty if no body)
//   - Res: The response body type (use Empty if no structured response)
//
// The context automatically binds the request body to ReqBody based on Content-Type.
// The response body (ResBody) can be set and sent using type-safe methods like OK(), Created(), etc.
//
// Example:
//
//	type CreateUserRequest struct {
//	    Name  string `json:"name" validate:"required"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	type UserResponse struct {
//	    ID    int    `json:"id"`
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	router.POST[CreateUserRequest, UserResponse]("/users", func(c *Box[CreateUserRequest, UserResponse]) error {
//	    // ReqBody is automatically bound from JSON
//	    req := c.ReqBody
//
//	    // Create user
//	    user := db.CreateUser(req.Name, req.Email)
//
//	    // Set response
//	    return c.Created("/users/"+user.ID, UserResponse{
//	        ID:    user.ID,
//	        Name:  user.Name,
//	        Email: user.Email,
//	    })
//	})
type Box[Req, Res any] struct {
	*Context

	// ReqBody is the parsed and validated request body.
	// It is automatically bound from the request based on Content-Type.
	// For handlers with no request body, use Empty type and ReqBody will be nil.
	ReqBody *Req

	// ResBody is the response body to be sent.
	// Set this field or use type-safe response methods (OK, Created, etc.)
	// For handlers with no response body, use Empty type.
	ResBody *Res
}

// newBox creates a new generic Box from a Context.
// This is used internally by the router when executing generic handlers.
func newBox[Req, Res any](base *Context) *Box[Req, Res] {
	return &Box[Req, Res]{
		Context: base,
	}
}

// OK sends a 200 OK response with the given data.
//
// Example:
//
//	return c.OK(UserResponse{ID: 1, Name: "John"})
func (c *Box[Req, Res]) OK(data Res) error {
	c.ResBody = &data
	return c.JSON(http.StatusOK, data)
}

// Created sends a 201 Created response with Location header and data.
//
// The location parameter should be the URL of the newly created resource.
//
// Example:
//
//	return c.Created("/users/123", UserResponse{ID: 123, Name: "John"})
func (c *Box[Req, Res]) Created(location string, data Res) error {
	c.ResBody = &data
	c.SetHeader("Location", location)
	return c.JSON(http.StatusCreated, data)
}

// Accepted sends a 202 Accepted response with data.
//
// Use this when a request has been accepted for processing but the processing
// has not been completed yet (e.g., async operations).
//
// Example:
//
//	return c.Accepted(TaskResponse{TaskID: "abc123", Status: "pending"})
func (c *Box[Req, Res]) Accepted(data Res) error {
	c.ResBody = &data
	return c.JSON(http.StatusAccepted, data)
}

// BadRequest sends a 400 Bad Request response with error details.
//
// Example:
//
//	return c.BadRequest(ErrorResponse{Message: "Invalid email format"})
func (c *Box[Req, Res]) BadRequest(data Res) error {
	c.ResBody = &data
	return c.JSON(http.StatusBadRequest, data)
}

// Unauthorized sends a 401 Unauthorized response with error details.
//
// Example:
//
//	return c.Unauthorized(ErrorResponse{Message: "Invalid credentials"})
func (c *Box[Req, Res]) Unauthorized(data Res) error {
	c.ResBody = &data
	return c.JSON(http.StatusUnauthorized, data)
}

// Forbidden sends a 403 Forbidden response with error details.
//
// Example:
//
//	return c.Forbidden(ErrorResponse{Message: "Insufficient permissions"})
func (c *Box[Req, Res]) Forbidden(data Res) error {
	c.ResBody = &data
	return c.JSON(http.StatusForbidden, data)
}

// NotFound sends a 404 Not Found response with error details.
//
// Example:
//
//	return c.NotFound(ErrorResponse{Message: "User not found"})
func (c *Box[Req, Res]) NotFound(data Res) error {
	c.ResBody = &data
	return c.JSON(http.StatusNotFound, data)
}

// InternalServerError sends a 500 Internal Server Error response with error details.
//
// Example:
//
//	return c.InternalServerError(ErrorResponse{Message: "Database connection failed"})
func (c *Box[Req, Res]) InternalServerError(data Res) error {
	c.ResBody = &data
	return c.JSON(http.StatusInternalServerError, data)
}

// ========================================
// Convenience Methods for Common REST Operations
// ========================================

// NoContentSuccess sends a 204 No Content response.
// This is a convenience method for successful operations with no response body.
//
// Common for DELETE operations and some PUT/PATCH updates.
// REST best practice: DELETE should return 204, not 200.
//
// Example:
//
//	router.DELETE[DeleteUserRequest, Empty]("/users/:id", func(c *Box[DeleteUserRequest, Empty]) error {
//	    deleteUser(c.ReqBody.ID)
//	    return c.NoContentSuccess()  // 204 No Content
//	})
func (c *Box[Req, Res]) NoContentSuccess() error {
	return c.NoContent(http.StatusNoContent)
}

// UpdatedOK sends a 200 OK response for successful updates with response body.
// This is an alias for OK() but provides semantic clarity for PUT/PATCH operations.
//
// Example:
//
//	router.PUT[UpdateUserRequest, UserResponse]("/users/:id", func(c *Box[UpdateUserRequest, UserResponse]) error {
//	    updated := updateUser(c.ReqBody)
//	    return c.UpdatedOK(updated)  // 200 OK - semantically clear it's an update
//	})
func (c *Box[Req, Res]) UpdatedOK(data Res) error {
	return c.OK(data)
}

// UpdatedNoContent sends a 204 No Content response for successful updates without response body.
// Use this when PUT/PATCH operations succeed but don't return data.
//
// Example:
//
//	router.PUT[UpdateUserRequest, Empty]("/users/:id", func(c *Box[UpdateUserRequest, Empty]) error {
//	    updateUser(c.ReqBody)
//	    return c.UpdatedNoContent()  // 204 No Content
//	})
func (c *Box[Req, Res]) UpdatedNoContent() error {
	return c.NoContent(http.StatusNoContent)
}

// Bind binds the request body to ReqBody based on Content-Type.
//
// Supported content types:
//   - application/json (default)
//   - application/xml, text/xml
//   - application/x-www-form-urlencoded
//   - multipart/form-data
//
// If a validator is set via Router.SetValidator(), the request body
// will be automatically validated after binding. Validation errors
// are returned as ValidationErrors.
//
// This method is automatically called by the generic handler adapter,
// so you typically don't need to call it manually.
//
// Returns error if binding or validation fails.
//
// Example:
//
//	// Manual binding (usually automatic)
//	if err := c.Bind(); err != nil {
//	    return c.BadRequest(ErrorResponse{Message: err.Error()})
//	}
func (c *Box[Req, Res]) Bind() error {
	// Check if Req is Empty type - if so, skip binding
	var zeroReq Req
	if _, ok := any(zeroReq).(Empty); ok {
		return nil
	}

	// Allocate request body
	req := new(Req)

	// Bind using the binding system
	if err := binding.Bind(c.Request, req); err != nil {
		return err
	}

	// Validate if validator is set
	if c.router != nil && c.router.validator != nil {
		if err := c.router.validator.Validate(req); err != nil {
			return err
		}
	}

	c.ReqBody = req
	return nil
}
