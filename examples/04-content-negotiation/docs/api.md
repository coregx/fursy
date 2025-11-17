# fursy HTTP Router API Reference

**Version**: v0.1.0
**Package**: github.com/coregx/fursy

---

## Overview

**fursy** (Fast Universal Routing SYstem) is a production-ready HTTP router for Go 1.25+ with type-safe handlers, RFC 9457 Problem Details, and built-in OpenAPI 3.1 generation.

## Key Features

- **Type-safe handlers** via Go generics - `Context[Req, Res]`
- **RFC 9457 Problem Details** for standardized error responses
- **Content negotiation** supporting JSON, XML, HTML, Markdown, and plain text
- **Zero dependencies** (core = stdlib only)
- **High performance** (~256 ns/op route lookup)

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/coregx/fursy"
)

func main() {
    router := fursy.New()

    router.GET("/", func(c *fursy.Context) error {
        return c.OK(map[string]string{"message": "Hello, World!"})
    })

    http.ListenAndServe(":8080", router)
}
```

## Content Negotiation Methods

### Accepts(mediaType string) bool

Check if the client accepts a specific media type.

**Example:**
```go
if c.Accepts(fursy.MIMETextMarkdown) {
    return c.Markdown(renderMarkdown(docs))
}
return c.JSON(200, data)
```

### AcceptsAny(mediaTypes ...string) string

Returns the best matching media type based on Accept header and q-values.

**Example:**
```go
switch c.AcceptsAny(fursy.MIMETextMarkdown, fursy.MIMETextHTML, fursy.MIMEApplicationJSON) {
case fursy.MIMETextMarkdown:
    return c.Markdown(content)
case fursy.MIMETextHTML:
    c.SetHeader("Content-Type", fursy.MIMETextHTML+"; charset=utf-8")
    return c.String(200, htmlContent)
default:
    return c.JSON(200, data)
}
```

### Negotiate(status int, data any) error

Automatically negotiate and respond in the best format.

**Example:**
```go
user := User{ID: 1, Name: "John"}
return c.Negotiate(200, user)
// Responds in JSON, XML, or plain text based on Accept header
```

## Response Methods

| Method | Content-Type | Use Case |
|--------|-------------|----------|
| `JSON(status, data)` | application/json | API responses (default) |
| `XML(status, data)` | application/xml | Legacy APIs, RSS/Atom feeds |
| `HTML(status, html)` | text/html | Web pages, browser rendering |
| `Markdown(content)` | text/markdown | AI agents, documentation |
| `Text(status, text)` | text/plain | Simple text responses |
| `Negotiate(status, data)` | Auto | Automatic format selection |

## RFC 9457 Problem Details

All errors use standardized Problem Details format:

```go
// 404 Not Found
return c.Problem(fursy.NotFound("User not found"))

// 400 Bad Request
return c.Problem(fursy.BadRequest("Invalid email format"))

// Custom problem
return c.Problem(fursy.NewProblem(429, "Too Many Requests", "Rate limit exceeded"))
```

## Middleware

Built-in middleware available:

```go
import "github.com/coregx/fursy/middleware"

router.Use(middleware.Logger())       // Request logging
router.Use(middleware.Recovery())     // Panic recovery
router.Use(middleware.CORS())         // CORS headers
router.Use(middleware.JWT(secret))    // JWT authentication
router.Use(middleware.RateLimit())    // Rate limiting
```

## Route Groups

```go
api := router.Group("/api/v1")
api.Use(middleware.JWT(secret))

api.GET("/users", listUsers)
api.POST("/users", createUser)
```

## Type-Safe Generic Handlers

```go
type CreateUserRequest struct {
    Email string `json:"email"`
    Name  string `json:"name"`
}

type UserResponse struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}

func createUser(c *fursy.Box[CreateUserRequest, UserResponse]) error {
    // c.ReqBody is *CreateUserRequest - type-safe!
    req := c.ReqBody

    user := &UserResponse{
        ID:    1,
        Email: req.Email,
        Name:  req.Name,
    }

    return c.Created("/users/1", user)
}

// Register with POST[Req, Res]
router.POST("/users", createUser)
```

## Performance

- Route lookup: **256 ns/op** (static routes)
- Allocations: **1 alloc/op** (near-zero allocation routing)
- Throughput: **~10M req/s** (simple routes)
- Test coverage: **88.9%**

## Support

- **GitHub**: https://github.com/coregx/fursy
- **Documentation**: https://pkg.go.dev/github.com/coregx/fursy
- **Issues**: https://github.com/coregx/fursy/issues

---

*This API reference is best viewed in markdown format by AI agents like Claude, ChatGPT, and other LLMs.*
