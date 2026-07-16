# 🔥 FURSY
> **F**ast **U**niversal **R**outing **Sy**stem

Next-generation HTTP router for Go with blazing performance, type-safe handlers, and minimal dependencies.

[![Go Reference](https://pkg.go.dev/badge/github.com/coregx/fursy.svg)](https://pkg.go.dev/github.com/coregx/fursy)
[![Go Report Card](https://goreportcard.com/badge/github.com/coregx/fursy)](https://goreportcard.com/report/github.com/coregx/fursy)
[![Tests](https://github.com/coregx/fursy/actions/workflows/test.yml/badge.svg)](https://github.com/coregx/fursy/actions)
[![codecov](https://codecov.io/gh/coregx/fursy/branch/main/graph/badge.svg)](https://codecov.io/gh/coregx/fursy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/coregx/fursy)](https://github.com/coregx/fursy/releases)

---

## ⚡ Quick Start

```go
package main

import (
    "log"
    "net/http"

    "github.com/coregx/fursy"
)

func main() {
    router := fursy.New()

    // Optional: Set validator for automatic validation
    // router.SetValidator(validator.New())

    // Simple text response with convenience method
    router.GET("/", func(c *fursy.Context) error {
        return c.Text("Welcome to FURSY!")  // 200 OK
    })

    // GET with convenience method (200 OK)
    router.GET("/users/:id", func(c *fursy.Context) error {
        id := c.Param("id")
        return c.OK(map[string]string{
            "id":   id,
            "name": "User " + id,
        })
    })

    // POST with convenience method (201 Created)
    router.POST("/users", func(c *fursy.Context) error {
        username := c.Form("username")
        email := c.Form("email")

        user := map[string]string{
            "id":    "123",
            "name":  username,
            "email": email,
        }
        return c.Created(user)  // 201 Created - REST best practice!
    })

    // DELETE with convenience method (204 No Content)
    router.DELETE("/users/:id", func(c *fursy.Context) error {
        // Delete user...
        return c.NoContentSuccess()  // 204 No Content
    })

    // Query parameters
    router.GET("/search", func(c *fursy.Context) error {
        query := c.Query("q")
        page := c.QueryDefault("page", "1")
        return c.OK(map[string]string{
            "query": query,
            "page":  page,
        })
    })

    log.Println("Server starting on :8080...")
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

> **Note**: The examples above use simple handlers with `*Context`. For type-safe generic handlers `Box[Req, Res]`, see the [Type-Safe Handlers](#type-safe-handlers-first-in-go) section below!

---

## 🌟 Why FURSY?

### Type-Safe Handlers (First in Go!)

```go
func Handler(box *fursy.Box[Request, Response]) error {
    // Compile-time type safety
    // Automatic validation
    // Zero boilerplate
}
```

### Native RFC 9457 Problem Details

```json
{
  "type": "https://fursy.coregx.dev/problems/validation-error",
  "title": "Validation Failed",
  "status": 400,
  "errors": [...]
}
```

### Built-in OpenAPI 3.1 Generation

```go
spec := r.OpenAPI(fursy.OpenAPIConfig{
    Title: "My API",
    Version: "1.0.0",
})
// Complete OpenAPI 3.1 spec from code!
```

### Minimal Dependencies

- **Core Routing**: Zero external dependencies (stdlib only)
- **Middleware**: Minimal deps (JWT: golang-jwt/jwt, RateLimit: x/time)
- **Plugins**: Optional extensions (OpenTelemetry, validators)
- Predictable, minimal security surface

### Production-Ready Performance

- **256 ns/op** static routes, **326 ns/op** parametric routes
- **1 allocation/op** (routing hot path)
- **~10M req/s** throughput (simple routes)
- Zero-allocation radix tree routing
- Efficient context pooling

---

## 📦 Installation

```bash
go get github.com/coregx/fursy
```

**Requirements**: Go 1.25+

---

## 🚀 Features

- ✅ **High Performance Routing** - 256-326 ns/op, 1 alloc/op
- ✅ **Type-Safe Generic Handlers** - Box[Req, Res] with compile-time safety
- ✅ **Automatic Validation** - Set once, validate everywhere with 100+ tags
- ✅ **Content Negotiation** - RFC 9110 compliant, AI agent support
- ✅ **RFC 9457 Problem Details** - Standardized error responses
- ✅ **Minimal Dependencies** - Core routing: stdlib only, middleware: minimal deps
- ✅ **Middleware Pipeline** - Next/Abort pattern, pre-allocated buffers
- ✅ **Route Groups** - Nested groups with middleware inheritance
- ✅ **JWT Authentication** - Token validation, claims extraction
- ✅ **Rate Limiting** - Token bucket algorithm, per-IP/per-user
- ✅ **Security Headers** - OWASP 2025 compliant (CSP, HSTS, etc.)
- ✅ **Circuit Breaker** - Failure threshold, auto-recovery
- ✅ **Trailing Slash Handling** - Strip or redirect, bidirectional, HTTP-correct (301/308)
- ✅ **Graceful Shutdown** - Connection draining, Kubernetes-ready
- ✅ **Context Pooling** - Memory-efficient, prevents leaks
- ✅ **Convenience Methods** - REST-friendly shortcuts (OK, Created, NoContentSuccess)
- ✅ **Real-Time Communications** - SSE + WebSocket via stream library
- ✅ **Database Integration** - dbcontext pattern with transaction support
- ✅ **Production Boilerplate** - Complete DDD example with real-time features

---

## 🎛️ Middleware

FURSY includes **8 production-ready middleware** with minimal dependencies. Core middleware have **zero external dependencies** (stdlib only), with only 2 exceptions: JWT (golang-jwt/jwt) and RateLimit (x/time).

### Core Middleware (Zero Dependencies)

#### Logger

Structured logging with `log/slog` for comprehensive request tracking.

```go
import (
    "log/slog"
    "github.com/coregx/fursy/middleware"
)

// Default configuration
router.Use(middleware.Logger())

// With configuration
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Logger: logger,
    SkipPaths: []string{"/health", "/metrics"},
}))
```

**Features**:
- ✅ Structured logging with `log/slog` (stdlib)
- ✅ Request method, path, status, latency, bytes written
- ✅ Client IP extraction (X-Real-IP, X-Forwarded-For)
- ✅ Skip paths or custom skip function
- ✅ JSON or text format support
- ✅ Zero external dependencies

---

#### Recovery

Panic recovery with stack traces and RFC 9457 Problem Details.

```go
router.Use(middleware.Recovery())

// With stack traces (development)
router.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
    IncludeStackTrace: true,
}))
```

**Features**:
- ✅ Automatic panic recovery
- ✅ Stack trace logging
- ✅ RFC 9457 error responses
- ✅ Custom error handler
- ✅ Production-safe (no stack traces by default)
- ✅ Zero external dependencies

---

#### CORS

Cross-Origin Resource Sharing (RFC-compliant, OWASP recommended).

```go
router.Use(middleware.CORS())

// With custom config
router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: "https://example.com,https://foo.com",
    AllowMethods: "GET,POST,PUT,DELETE",
    AllowHeaders: "Content-Type,Authorization",
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
}))
```

**Features**:
- ✅ Wildcard origins (`*`) support
- ✅ Preflight requests (OPTIONS) handling
- ✅ Credentials support
- ✅ Expose headers configuration
- ✅ MaxAge caching
- ✅ Zero external dependencies

---

#### BasicAuth

HTTP Basic Authentication with constant-time comparison.

```go
router.Use(middleware.BasicAuth(middleware.BasicAuthConfig{
    Username: "admin",
    Password: "secret",
}))

// With custom validator
router.Use(middleware.BasicAuth(middleware.BasicAuthConfig{
    Validator: func(username, password string) bool {
        return checkDatabase(username, password)
    },
}))
```

**Features**:
- ✅ Simple username/password validation
- ✅ Custom validator function
- ✅ Realm configuration
- ✅ WWW-Authenticate header
- ✅ Constant-time comparison (timing attack protection)
- ✅ Zero external dependencies

---

#### Secure

OWASP 2025 security headers for production hardening.

```go
router.Use(middleware.Secure())

// With custom config
router.Use(middleware.SecureWithConfig(middleware.SecureConfig{
    ContentSecurityPolicy:   "default-src 'self'; script-src 'self' 'unsafe-inline'",
    HSTSMaxAge:             31536000, // 1 year
    HSTSExcludeSubdomains:  false,
    XFrameOptions:          "DENY",
    ContentTypeNosniff:     "nosniff",
    ReferrerPolicy:         "strict-origin-when-cross-origin",
}))
```

**Features (OWASP 2025)**:
- ✅ Content-Security-Policy (CSP)
- ✅ Strict-Transport-Security (HSTS)
- ✅ X-Frame-Options
- ✅ X-Content-Type-Options: nosniff
- ✅ X-XSS-Protection (deprecated, not set by default)
- ✅ Referrer-Policy
- ✅ Cross-Origin-Embedder-Policy
- ✅ Cross-Origin-Opener-Policy
- ✅ Cross-Origin-Resource-Policy
- ✅ Permissions-Policy

**Coverage**: 100%
**Dependencies**: Zero (stdlib only)

---

### Authentication & Rate Limiting

#### JWT

JWT token validation with algorithm confusion prevention.

```go
// Simple usage with signing key
router.Use(middleware.JWT([]byte("your-secret-key")))

// With full configuration
router.Use(middleware.JWTWithConfig(middleware.JWTConfig{
    SigningKey:    []byte("your-secret-key"),
    SigningMethod: "HS256",
    TokenLookup:  "header:Authorization",
}))

// With issuer and audience validation
router.Use(middleware.JWTWithConfig(middleware.JWTConfig{
    SigningKey:       []byte("secret"),
    SigningMethod:    "HS256",
    ValidateIssuer:   "my-app",
    ValidateAudience: "api",
}))
```

**Features**:
- ✅ Algorithms: HS256, HS384, HS512, RS256, ES256
- ✅ Token from Header/Query/Cookie
- ✅ Issuer/Audience validation
- ✅ Algorithm confusion prevention (forbids "none")
- ✅ Custom claims support
- ✅ Expiration time validation

**Dependency**: `github.com/golang-jwt/jwt/v5`

---

#### RateLimit

Token bucket rate limiting with RFC-compliant headers.

```go
// Simple usage (100 req/s, burst of 200)
router.Use(middleware.RateLimit(100, 200))

// With full configuration and custom key function
router.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
    Rate:  10,
    Burst: 20,
    KeyFunc: func(c *fursy.Context) string {
        // Rate limit by user ID
        return c.GetString("user_id")
    },
}))
```

**Features**:
- ✅ Token bucket algorithm (`golang.org/x/time/rate`)
- ✅ Per-IP or custom key function
- ✅ RFC headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`)
- ✅ In-memory store with automatic cleanup
- ✅ Custom error handler
- ✅ Configurable retry-after header

**Dependency**: `golang.org/x/time/rate`

---

### Resilience

#### CircuitBreaker

Zero-dependency circuit breaker for fault tolerance.

```go
// Simple usage with defaults (5 consecutive failures, 60s timeout)
router.Use(middleware.CircuitBreaker())

// With consecutive failures threshold
router.Use(middleware.CircuitBreakerConsecutive(5, 30*time.Second))

// With ratio-based threshold (3 failures out of 10 requests)
router.Use(middleware.CircuitBreakerRatio(3, 10, 30*time.Second))

// With full configuration
router.Use(middleware.CircuitBreakerWithConfig(middleware.CircuitBreakerConfig{
    ConsecutiveFailures: 5,
    Timeout:             30 * time.Second,
    MaxRequests:         2, // Allow 2 requests in half-open state
}))
```

**Features**:
- ✅ Zero external dependencies (pure Go)
- ✅ Consecutive failures threshold
- ✅ Ratio-based threshold
- ✅ Time-window threshold
- ✅ Half-open state with max requests
- ✅ States: Closed → Open → Half-Open → Closed
- ✅ Custom error handler
- ✅ Thread-safe (concurrent request handling)

**Coverage**: 95.6%
**Dependencies**: Zero (stdlib only)

---

### Middleware Comparison

| Middleware | FURSY | Gin | Echo | Fiber |
|------------|-------|-----|------|-------|
| **Logger** | ✅ `log/slog` | ✅ Custom | ✅ Custom | ✅ Custom |
| **Recovery** | ✅ RFC 9457 | ✅ Basic | ✅ Basic | ✅ Basic |
| **CORS** | ✅ Built-in (zero deps) | 🔧 Plugin | 🔧 Plugin | ✅ Built-in |
| **BasicAuth** | ✅ Built-in | ✅ Built-in | ✅ Built-in | ✅ Built-in |
| **JWT** | ✅ Built-in | 🔧 Plugin | 🔧 Plugin | ✅ Built-in |
| **Rate Limit** | ✅ Built-in (RFC headers) | 🔧 Plugin | 🔧 Plugin | ✅ Built-in |
| **Security Headers** | ✅ OWASP 2025 | ❌ | 🔧 Plugin | ✅ Basic |
| **Circuit Breaker** | ✅ Zero deps | ❌ | ❌ | ❌ |
| **Test Coverage** | **94.6%** | ? | ? | ? |
| **Dependencies** | **Core: 0, JWT: 1, RateLimit: 1** | Multiple | Multiple | Multiple |

**Legend**:
- ✅ = Built-in with high quality implementation
- 🔧 = Plugin/third-party required
- ❌ = Not available

**FURSY advantage**: Production-ready middleware with minimal dependencies, OWASP 2025 compliance, RFC 9457 error responses, and comprehensive test coverage.

---

### Learn More

- **[Middleware Examples](examples/05-middleware/)** - Complete examples for all 8 middleware
- **[Middleware Source](middleware/)** - Middleware implementations with tests

---

## 🔀 Trailing Slash Handling

By default, `/users` and `/users/` are treated as different routes. Configure trailing slash handling to match your needs:

```go
// Strip trailing slashes silently (recommended for APIs behind reverse proxies)
router := fursy.New()
router.WithTrailingSlash(fursy.StripTrailingSlash)

router.GET("/api/companies", handler)
// GET /api/companies  → 200 (exact match)
// GET /api/companies/ → 200 (trailing slash stripped silently)
```

```go
// Redirect to canonical URL (SEO-friendly for public APIs)
router := fursy.New()
router.WithTrailingSlash(fursy.RedirectTrailingSlash)

router.GET("/api/companies", handler)
// GET  /api/companies/ → 301 redirect to /api/companies
// POST /api/companies/ → 308 redirect (preserves HTTP method)
```

**Bidirectional**: If `/files/` is the registered route and `/files` is requested, the router handles it too — redirecting or serving the `/files/` handler.

**Three modes**:
| Mode | Behavior |
|------|----------|
| `IgnoreTrailingSlash` (default) | `/users` and `/users/` are different routes |
| `StripTrailingSlash` | Silently serves handler without redirect |
| `RedirectTrailingSlash` | 301 (GET) or 308 (other methods) redirect |

---

## 🎯 Convenience Methods (REST Best Practices)

FURSY provides convenient shortcuts for common HTTP response patterns, following REST best practices:

### Context Convenience Methods

```go
// GET - 200 OK (most common)
router.GET("/users", func(c *fursy.Context) error {
    users := getAllUsers()
    return c.OK(users)  // Short for c.JSON(200, users)
})

// POST - 201 Created (resource creation)
router.POST("/users", func(c *fursy.Context) error {
    user := createUser(c)
    return c.Created(user)  // 201, not 200!
})

// DELETE - 204 No Content (successful deletion)
router.DELETE("/users/:id", func(c *fursy.Context) error {
    deleteUser(c.Param("id"))
    return c.NoContentSuccess()  // 204, no body
})

// Async operations - 202 Accepted
router.POST("/jobs", func(c *fursy.Context) error {
    jobID := startAsyncJob(c)
    return c.Accepted(map[string]string{"jobId": jobID})
})

// Simple text - 200 OK
router.GET("/ping", func(c *fursy.Context) error {
    return c.Text("pong")  // text/plain, 200
})
```

**Why use convenience methods?**
- ✅ **Less boilerplate** - `c.OK(data)` vs `c.JSON(200, data)`
- ✅ **REST semantics** - `Created()` clearly indicates 201, preventing mistakes
- ✅ **Self-documenting** - Code intent is clear from method name
- ✅ **Flexibility** - Original methods still available for custom status codes

**For custom status codes**, use explicit methods:
```go
// Partial content - 206
return c.JSON(206, partialData)

// Custom redirect - 307
return c.Redirect(307, "/new-location")
```

### Box Convenience Methods (Type-Safe)

```go
// GET - 200 OK
router.GET[GetUserRequest, UserResponse]("/users/:id", func(b *fursy.Box[GetUserRequest, UserResponse]) error {
    user := getUser(b.ReqBody.ID)
    return b.OK(user)  // Type-safe 200 OK
})

// POST - 201 Created with Location header
router.POST[CreateUserRequest, UserResponse]("/users", func(b *fursy.Box[CreateUserRequest, UserResponse]) error {
    user := createUser(b.ReqBody)
    return b.Created("/users/"+user.ID, user)  // 201 + Location
})

// PUT - 200 OK with body
router.PUT[UpdateUserRequest, UserResponse]("/users/:id", func(b *fursy.Box[UpdateUserRequest, UserResponse]) error {
    updated := updateUser(b.ReqBody)
    return b.UpdatedOK(updated)  // Semantic clarity
})

// PUT - 204 No Content (no response body)
router.PUT[UpdateUserRequest, Empty]("/users/:id", func(b *fursy.Box[UpdateUserRequest, Empty]) error {
    updateUser(b.ReqBody)
    return b.UpdatedNoContent()  // 204, no body
})

// DELETE - 204 No Content
router.DELETE[Empty, Empty]("/users/:id", func(b *fursy.Box[Empty, Empty]) error {
    deleteUser(c.Param("id"))
    return b.NoContentSuccess()  // 204
})
```

### Plugin Integration Methods

FURSY provides seamless integration with plugins through convenient Context methods:

#### Database Access

```go
import (
    "github.com/coregx/fursy"
    "github.com/coregx/fursy/plugins/database"
)

// Setup database
sqlDB, _ := sql.Open("postgres", dsn)
db := database.NewDB(sqlDB)

router := fursy.New()
router.Use(database.Middleware(db))

// Access database in handlers
router.GET("/users/:id", func(c *fursy.Context) error {
    db := c.DB().(*database.DB)  // Type assertion

    var user User
    err := db.QueryRow(c.Request.Context(),
        "SELECT id, name FROM users WHERE id = $1", c.Param("id")).
        Scan(&user.ID, &user.Name)

    if err == sql.ErrNoRows {
        return c.Problem(fursy.NotFound("User not found"))
    }
    return c.JSON(200, user)
})
```

**Type-safe helper** (recommended):

```go
router.GET("/users/:id", func(c *fursy.Context) error {
    db, ok := database.GetDB(c)  // Type-safe retrieval
    if !ok {
        return c.Problem(fursy.InternalServerError("Database not configured"))
    }

    // Use db...
})
```

#### Server-Sent Events (SSE)

```go
import (
    "github.com/coregx/fursy/plugins/stream"
    "github.com/coregx/stream/sse"
)

// Setup SSE hub
hub := sse.NewHub[Notification]()
go hub.Run()
defer hub.Close()

router.Use(stream.SSEHub(hub))

// SSE endpoint
router.GET("/events", func(c *fursy.Context) error {
    hub, _ := stream.GetSSEHub[Notification](c)

    return stream.SSEUpgrade(c, func(conn *sse.Conn) error {
        hub.Register(conn)
        defer hub.Unregister(conn)

        <-conn.Done()  // Wait for client disconnect
        return nil
    })
})
```

#### WebSocket

```go
import (
    "github.com/coregx/fursy/plugins/stream"
    "github.com/coregx/stream/websocket"
)

// Setup WebSocket hub
hub := websocket.NewHub()
go hub.Run()
defer hub.Close()

router.Use(stream.WebSocketHub(hub))

// WebSocket endpoint
router.GET("/ws", func(c *fursy.Context) error {
    hub, _ := stream.GetWebSocketHub(c)

    return stream.WebSocketUpgrade(c, func(conn *websocket.Conn) error {
        hub.Register(conn)
        defer hub.Unregister(conn)

        for {
            msgType, data, err := conn.Read()
            if err != nil {
                return err
            }
            hub.Broadcast(data)  // Echo to all clients
        }
    }, nil)
})
```

**See also**:
- **[plugins/database](plugins/database/)** - Database integration with transactions
- **[plugins/stream](plugins/stream/)** - SSE and WebSocket real-time communication
- **[examples/07-sse-notifications](examples/07-sse-notifications/)** - SSE example
- **[examples/08-websocket-chat](examples/08-websocket-chat/)** - WebSocket example

---

## 🎯 Automatic Validation

FURSY provides **type-safe automatic validation** through the validator plugin, giving you compile-time type safety combined with runtime validation - a unique combination in the Go ecosystem.

### Why FURSY Validation is Different

Traditional routers require **manual validation** on every handler:

```go
// ❌ Manual validation (Gin, Echo, Fiber)
func CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {  // No validation!
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Manual validation needed
    if req.Email == "" || !isValidEmail(req.Email) {
        c.JSON(400, gin.H{"error": "invalid email"})
        return
    }
    // ... repeat for every field
}
```

With FURSY's **type-safe handlers**, validation is **automatic and guaranteed**:

```go
// ✅ Automatic validation (FURSY)
router.POST[CreateUserRequest, UserResponse]("/users",
    func(c *fursy.Box[CreateUserRequest, UserResponse]) error {
        if err := c.Bind(); err != nil {
            return err  // Automatic RFC 9457 error response
        }

        // c.ReqBody is ALREADY validated! ✅
        user := createUser(c.ReqBody)
        return c.Created("/users/"+user.ID, user)
    })
```

**Key advantages:**
- ✅ **Set once, validate everywhere** - No manual checks per handler
- ✅ **Compile-time type safety** - Generics ensure request/response types match
- ✅ **RFC 9457 compliant** - Standard error format with field-level details
- ✅ **100+ validation tags** - email, URL, UUID, min/max, and more

### Quick Example

```go
package main

import (
    "github.com/coregx/fursy"
    "github.com/coregx/fursy/plugins/validator"
)

type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=50"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,gte=18,lte=120"`
}

type UserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    router := fursy.New()

    // Set validator once - applies to ALL handlers
    router.SetValidator(validator.New())

    // Type-safe handler with automatic validation
    router.POST[CreateUserRequest, UserResponse]("/users",
        func(c *fursy.Box[CreateUserRequest, UserResponse]) error {
            if err := c.Bind(); err != nil {
                return err  // Automatic RFC 9457 response
            }

            // c.ReqBody is validated and type-safe!
            user := createUser(c.ReqBody)
            return c.Created("/users/"+user.ID, user)
        })

    log.Fatal(http.ListenAndServe(":8080", router))
}
```

### Validation Error Response

When validation fails, FURSY returns **RFC 9457 Problem Details** with field-level errors:

**Request:**
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jo","email":"invalid","age":15}'
```

**Response (422 Unprocessable Entity):**
```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "3 field(s) failed validation",
  "errors": {
    "name": "Name must be at least 3 characters",
    "email": "Email must be a valid email address",
    "age": "Age must be 18 or greater"
  }
}
```

### Comparison with Other Routers

| Feature | FURSY | Gin | Echo | Fiber |
|---------|-------|-----|------|-------|
| Type Safety | ✅ Compile-time (`Box[Req, Res]`) | ❌ Runtime only | ❌ Runtime only | ❌ Runtime only |
| Auto Validation | ✅ Set once, validate all | ❌ Manual per handler | ❌ Manual per handler | ❌ Manual per handler |
| Error Format | ✅ RFC 9457 (standard) | ❌ Custom JSON | ❌ Custom JSON | ❌ Custom JSON |
| Setup Complexity | ✅ One line (`SetValidator`) | ❌ Validator + binding per route | ❌ Validator + binding per route | ❌ Validator + binding per route |
| Field-Level Errors | ✅ Automatic | 🔧 Manual mapping | 🔧 Manual mapping | 🔧 Manual mapping |

**Learn more**: See [Validator Plugin Documentation](plugins/validator/README.md) for custom validators, nested structs, and advanced features.

---

## 🌐 Content Negotiation

FURSY provides **RFC 9110 compliant content negotiation**, enabling your API to respond in multiple formats based on the client's `Accept` header. This is essential for building modern APIs that serve both humans (HTML/Markdown) and machines (JSON/XML).

### Why Content Negotiation Matters

Modern APIs need to support multiple clients:
- **Web Browsers** → HTML
- **API Clients** → JSON
- **AI Agents** → Markdown (for better understanding)
- **Legacy Systems** → XML

FURSY handles this automatically using **RFC 9110** standards with quality values (q-parameters).

### Automatic Format Selection

The simplest approach - FURSY picks the best format automatically:

```go
router.GET("/users/:id", func(c *fursy.Context) error {
    user := getUser(c.Param("id"))

    // Automatically selects format based on Accept header
    // Supports: JSON, HTML, XML, Text, Markdown
    return c.Negotiate(200, user)
})
```

**Client requests:**
```bash
# JSON (default)
curl http://localhost:8080/users/123
# → Content-Type: application/json

# HTML
curl -H "Accept: text/html" http://localhost:8080/users/123
# → Content-Type: text/html

# XML
curl -H "Accept: application/xml" http://localhost:8080/users/123
# → Content-Type: application/xml
```

### Explicit Format Control

For finer control, check what the client accepts:

```go
router.GET("/docs", func(c *fursy.Context) error {
    // Check if client accepts markdown
    if c.Accepts(fursy.MIMETextMarkdown) {
        docs := generateMarkdownDocs()
        return c.Markdown(docs)  // AI-friendly format
    }

    // Fallback to JSON
    return c.OK(map[string]string{"message": "Use Accept: text/markdown for docs"})
})
```

### Quality Values (q-parameter)

RFC 9110 defines quality values to prioritize formats:

```go
router.GET("/api/data", func(c *fursy.Context) error {
    data := getData()

    // Client sends: Accept: text/html;q=0.9, application/json;q=1.0
    // FURSY automatically picks JSON (higher q-value)
    format := c.AcceptsAny(
        fursy.MIMEApplicationJSON,  // q=1.0
        fursy.MIMETextHTML,          // q=0.9
        fursy.MIMETextMarkdown,      // fallback
    )

    switch format {
    case fursy.MIMEApplicationJSON:
        return c.JSON(200, data)
    case fursy.MIMETextHTML:
        return c.HTML(200, renderHTML(data))
    case fursy.MIMETextMarkdown:
        return c.Markdown(formatMarkdown(data))
    default:
        return c.OK(data)  // Default to JSON
    }
})
```

### Supported Formats

| Format | MIME Type | Constant | Use Case |
|--------|-----------|----------|----------|
| JSON | `application/json` | `MIMEApplicationJSON` | API responses (default) |
| HTML | `text/html` | `MIMETextHTML` | Web browsers |
| XML | `application/xml` | `MIMEApplicationXML` | Legacy systems |
| Plain Text | `text/plain` | `MIMETextPlain` | Simple data |
| Markdown | `text/markdown` | `MIMETextMarkdown` | AI agents, documentation |

### AI Agent Support

FURSY has first-class support for AI agents via Markdown responses:

```go
router.GET("/api/schema", func(c *fursy.Context) error {
    // AI agents prefer markdown for better understanding
    if c.Accepts(fursy.MIMETextMarkdown) {
        schema := `
# API Schema

## Users Endpoint
- **GET** /users - List all users
- **POST** /users - Create new user
  - Required: name (string), email (string)

## Authentication
All endpoints require Bearer token in Authorization header.
`
        return c.Markdown(schema)
    }

    // Regular clients get JSON
    return c.JSON(200, getOpenAPISchema())
})
```

**Why Markdown for AI?**
- ✅ Better semantic understanding than JSON
- ✅ Preserves structure (headers, lists, code blocks)
- ✅ More context for LLMs to understand API behavior
- ✅ Human-readable for debugging

### Comparison with Other Routers

| Feature | FURSY | Gin | Echo | Fiber |
|---------|-------|-----|------|-------|
| RFC 9110 Compliance | ✅ Full | 🔧 Partial | 🔧 Partial | 🔧 Partial |
| Automatic Negotiation | ✅ `Negotiate()` | ❌ Manual | 🔧 `c.Format()` | ❌ Manual |
| Quality Values (q) | ✅ Automatic | ❌ No | ❌ No | ❌ No |
| Accept Helpers | ✅ `Accepts()`, `AcceptsAny()` | ❌ No | ❌ No | ✅ `c.Accepts()` |
| Markdown Support | ✅ Built-in | ❌ Manual | ❌ Manual | ❌ Manual |
| AI Agent Ready | ✅ Yes | ❌ No | ❌ No | ❌ No |

**FURSY advantage**: Only router with full RFC 9110 compliance, automatic q-value handling, and built-in AI agent support.

**Learn more**: See [RFC 9110 - HTTP Semantics (Content Negotiation)](https://datatracker.ietf.org/doc/html/rfc9110#section-12) for the complete specification.

---

## 📊 Observability

FURSY provides **production-ready observability** through the OpenTelemetry plugin, giving you complete visibility into your HTTP services with distributed tracing and metrics.

### Why Observability Matters

Modern distributed systems require:
- **Distributed Tracing** → Track requests across microservices
- **Performance Metrics** → Monitor latency, throughput, errors
- **Error Tracking** → Automatic error recording and status tracking
- **Production Debugging** → Understand behavior in real-time

FURSY's OpenTelemetry plugin provides all of this with **zero boilerplate** - just add middleware.

### Distributed Tracing

Track every request with W3C Trace Context propagation:

```go
import (
    "context"
    "github.com/coregx/fursy"
    "github.com/coregx/fursy/plugins/opentelemetry"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
    // Initialize OpenTelemetry tracer
    exporter, _ := jaeger.New(jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint("http://localhost:14268/api/traces"),
    ))
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
    )
    otel.SetTracerProvider(tp)
    defer tp.Shutdown(context.Background())

    // Add tracing middleware - that's it!
    router := fursy.New()
    router.Use(opentelemetry.Middleware("my-service"))

    router.GET("/users/:id", func(c *fursy.Context) error {
        // Automatically traced! Span includes:
        // - HTTP method, path, status
        // - Request/response headers
        // - Duration
        // - Errors (if any)
        user := getUser(c.Param("id"))
        return c.OK(user)
    })

    http.ListenAndServe(":8080", router)
}
```

**Features**:
- ✅ **W3C Trace Context** - Automatic propagation across services
- ✅ **HTTP Semantic Conventions** - Full OpenTelemetry compliance
- ✅ **Error Recording** - Automatic error and status tracking
- ✅ **Zero Overhead Filtering** - Skip health checks and metrics endpoints

### Metrics Collection

Track HTTP performance with Prometheus-compatible metrics:

```go
import (
    "github.com/coregx/fursy"
    "github.com/coregx/fursy/plugins/opentelemetry"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/prometheus"
    sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func main() {
    // Initialize Prometheus exporter
    exporter, _ := prometheus.New()
    mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
    otel.SetMeterProvider(mp)

    // Add metrics middleware
    router := fursy.New()
    router.Use(opentelemetry.Metrics("my-service"))

    // Metrics automatically collected:
    // - http.server.request.duration (histogram)
    // - http.server.request.count (counter)
    // - http.server.request.size (histogram)
    // - http.server.response.size (histogram)

    router.GET("/users", func(c *fursy.Context) error {
        users := getAllUsers()
        return c.OK(users)
    })

    // Expose metrics at /metrics
    router.GET("/metrics", promhttp.Handler())

    http.ListenAndServe(":8080", router)
}
```

**Available Metrics**:

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `http.server.request.duration` | Histogram | Request latency | method, status, server |
| `http.server.request.count` | Counter | Total requests | method, status |
| `http.server.request.size` | Histogram | Request body size | method |
| `http.server.response.size` | Histogram | Response body size | method, status |

**Cardinality Management**: All metrics use low-cardinality labels (method, status, server) to prevent metrics explosion.

### Custom Spans for Business Logic

Add custom spans to trace specific operations:

```go
import "go.opentelemetry.io/otel"

router.GET("/users/:id", func(c *fursy.Context) error {
    // HTTP request span is created automatically by middleware

    // Add custom span for database query
    ctx := c.Request.Context()
    tracer := otel.Tracer("my-service")
    ctx, span := tracer.Start(ctx, "database.get_user")
    defer span.End()

    user := db.GetUser(ctx, c.Param("id"))

    // Add custom span for external API call
    _, apiSpan := tracer.Start(ctx, "api.enrich_user_data")
    enrichedData := api.Enrich(user)
    apiSpan.End()

    return c.OK(enrichedData)
})
```

### Jaeger Integration Example

Complete setup with Jaeger for local development:

```bash
# Start Jaeger all-in-one (includes UI)
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 14268:14268 \
  jaegertracing/all-in-one:latest

# View traces at http://localhost:16686
```

Your fursy application will automatically send traces to Jaeger. No configuration changes needed!

### Comparison with Other Routers

| Feature | FURSY | Gin | Echo | Fiber |
|---------|-------|-----|------|-------|
| OpenTelemetry Built-in | ✅ Plugin | 🔧 Third-party | 🔧 Third-party | 🔧 Third-party |
| HTTP Semantic Conventions | ✅ Full | 🔧 Partial | 🔧 Partial | 🔧 Partial |
| Metrics API | ✅ OpenTelemetry | 🔧 Prometheus only | 🔧 Prometheus only | ✅ Built-in |
| Distributed Tracing | ✅ W3C Trace Context | 🔧 Manual | 🔧 Manual | 🔧 Manual |
| Cardinality Management | ✅ Automatic | ❌ Manual | ❌ Manual | ✅ Automatic |
| Zero-config | ✅ One line | ❌ Multiple steps | ❌ Multiple steps | ✅ One line |

**FURSY advantage**: Official OpenTelemetry plugin with full HTTP semantic conventions compliance and zero-config setup.

**Learn more**: See [OpenTelemetry Plugin Documentation](plugins/opentelemetry/README.md) for advanced configuration, custom spans, and production patterns.

---

## 📖 Documentation

**Status**: 🟡 In Development

- [Getting Started](#quick-start) (above)
- [Validator Plugin](plugins/validator/README.md) - Type-safe validation
- [OpenTelemetry Plugin](plugins/opentelemetry/README.md) - Distributed tracing and metrics
- API Reference (coming soon)
- Examples (coming soon)
- Migration Guides (coming soon)

---

## 🎯 Comparison

| Feature | FURSY | Gin | Echo | Chi | Fiber |
|---------|--------|-----|------|-----|-------|
| Type-Safe Handlers | ✅ | ❌ | ❌ | ❌ | ❌ |
| Auto Validation | ✅ | 🔧 Manual | 🔧 Manual | 🔧 Manual | 🔧 Manual |
| Content Negotiation | ✅ RFC 9110 | 🔧 Partial | 🔧 Partial | ❌ | 🔧 Partial |
| Zero Deps (core) | ✅ | ❌ | ❌ | ✅ | ❌ |
| OpenAPI Built-in | ✅ | 🔧 Plugin | 🔧 Plugin | 🔧 Plugin | 🔧 Plugin |
| RFC 9457 Errors | ✅ | ❌ | ❌ | ❌ | ❌ |
| Performance | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| Go Version | 1.25+ | 1.13+ | 1.17+ | 1.16+ | 1.17+ |

**FURSY is unique**: Only router combining furious performance, type-safe generics, automatic validation, RFC 9110 content negotiation, OpenAPI, and RFC 9457 with minimal dependencies.

---

## 📈 Status

**Current Version**: v0.3.3 (Production Ready)

**Status**: Production Ready - Complete ecosystem with real-time, database, and production examples

**Coverage**: 94.6% test coverage (total), 94.6% core, 97.7% binding, 95.6% middleware

**Performance**: 256 ns/op (static), 326 ns/op (parametric), 1 alloc/op

**Roadmap**:

```
✅ v0.1.0          ✅ v0.2.0          ✅ v0.3.0             🎯 v1.0.0 LTS
(Foundation)     (Docs+Examples)  (Real-time+DB)         (TBD - After Full
                                                          API Stabilization)
    │                  │                 │                       │
    ▼                  ▼                 ▼                       ▼
Core Router        Documentation    Real-Time+DB            Stable API
Middleware         11 Examples      Production Ready        Long-Term Support
Production         Validation       2 Plugins               (NOT Rushing!)
Features           OpenAPI          DDD Boilerplate
```

**Current Status**: v0.3.3 Production Ready ✅
**Ecosystem**: stream v0.1.0 (SSE + WebSocket), 2 production plugins, 10 examples
**Next**: v0.x.x feature releases as needed (Cache, more plugins, community tools)
**v1.0.0 LTS**: After 6-12 months of production usage and full API stabilization

---

## 🤝 Contributing

We welcome contributions! Please see:

- [CONTRIBUTING.md](CONTRIBUTING.md) - Development workflow and guidelines
- [RELEASE_GUIDE.md](RELEASE_GUIDE.md) - Release process
- [SECURITY.md](SECURITY.md) - Security policy

**Development Requirements**:
- Go 1.25+
- golangci-lint
- Follow git-flow branching model

**Want to help?**
- ⭐ Star the repo
- 📢 Share with others
- 🐛 Report bugs or request features
- 💬 Join discussions (coming soon)

---

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details

---

## 🔗 Links

- **GitHub**: [github.com/coregx/fursy](https://github.com/coregx/fursy)
- **Organization**: [github.com/coregx](https://github.com/coregx)
- **Sister Project**: [Relica](https://github.com/coregx/relica) - Database query builder

---

## 💡 Inspiration

FURSY stands on the shoulders of giants:

**Technical**:
- [ozzo-routing](https://github.com/go-ozzo/ozzo-routing) - Middleware pipeline
- [httprouter](https://github.com/julienschmidt/httprouter) - Radix tree routing
- [fiber](https://github.com/gofiber/fiber) - Performance inspiration
- [FastAPI](https://fastapi.tiangolo.com/) (Python) - Type hints + OpenAPI

**Philosophy**:
- [Relica](https://github.com/coregx/relica) - Zero deps, type safety, quality

### Special Thanks

**Professor Ancha Baranova** - This project would not have been possible without her invaluable help and support. Her assistance was crucial in making all coregx projects a reality.

---

## 📞 Contact

**Questions?** Check back soon for:
- GitHub Discussions
- Discord server
- Documentation site

---

<div align="center">

### 🔥 **FURSY** - Unleash Routing Fursy

**Blazing Fast** • **Minimal Dependencies** • **Type-Safe** • **Furious**

---

*Built with ❤️ by the coregx team*

**Version**: v0.3.3 - Production Ready
**Ecosystem**: stream v0.1.0 + 2 plugins + 10 examples + DDD boilerplate
**Next**: v1.0.0 LTS (after full API stabilization)

</div>

---
