# fursy - AI Agent Guide

> **DEFINITIVE GUIDE** for AI agents (Claude, ChatGPT, Copilot, etc.) working on fursy HTTP router project.
>
> **Project**: fursy - Fast Universal Routing SYstem
> **Version**: v0.1.0 (Production Ready)
> **Go**: 1.25+
> **Status**: Phase 3 Complete → Phase 4 (Ecosystem)

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Critical Files to Read First](#critical-files-to-read-first)
3. [Architecture](#architecture)
4. [Project Structure](#project-structure)
5. [Core Concepts](#core-concepts)
6. [Quick Start for AI Agents](#quick-start-for-ai-agents)
7. [Development Standards](#development-standards)
8. [Testing Requirements](#testing-requirements)
9. [Git Workflow](#git-workflow)
10. [Middleware](#middleware)
11. [Examples](#examples)
12. [Comparison with Other Routers](#comparison-with-other-routers)
13. [Key Files Reference](#key-files-reference)
14. [Performance Metrics](#performance-metrics)
15. [Common Gotchas](#common-gotchas)
16. [Contributing Guidelines](#contributing-guidelines)

---

## Project Overview

**FURSY** is a next-generation HTTP router for Go 1.25+ that combines blazing performance with modern features:

### Key Features

- **Type-Safe Generic Handlers**: `Context[Req, Res]` with compile-time type safety (FIRST in Go!)
- **RFC 9457 Problem Details**: Standardized error responses built-in
- **OpenAPI 3.1 Generation**: Automatic API documentation from code
- **Minimal Dependencies**: Core = stdlib only, middleware = 2 dependencies (JWT, RateLimit)
- **Zero-Allocation Routing**: 256 ns/op, 1 alloc/op (production-ready performance)
- **Production Middleware**: 8 built-in middleware (Logger, Recovery, CORS, BasicAuth, JWT, RateLimit, CircuitBreaker, Secure)
- **Content Negotiation**: RFC 9110 compliant with AI agent support (Markdown responses)

### Current Status (2025-11-17)

- **Version**: v0.1.0 (after rebranding from FURY)
- **Phase**: Phase 3 Complete ✅ (Foundation, API Excellence, Production Features)
- **Next Phase**: Phase 4 - Ecosystem (documentation, community, plugins)
- **Coverage**: 88.9% core (exceeds >85% target)
- **Linter**: 0 issues (golangci-lint strict mode)
- **Tests**: 150+ test functions, 19 benchmarks
- **Performance**: 256 ns/op static, 326 ns/op parametric, ~10M req/s throughput

### Production Ready

- All core features complete
- Comprehensive test coverage
- Zero linter issues
- Performance validated
- Ready for production use

---

## Critical Files to Read First

**IMPORTANT**: Always read these files in this order when starting work:

### 1. `.claude/STATUS.md` ⭐ HIGHEST PRIORITY

**Why**: Contains current project status, active tasks, metrics, and known issues.

**What's inside**:
- Current phase and progress (Phase 3 Complete, Phase 4 Ready)
- Active tasks (currently: documentation and examples)
- Test coverage (88.9%)
- Performance metrics (256 ns/op)
- Recent updates (rebranding FURY → fursy)
- Kanban status (25 done, 32 in backlog)

**Read it**: `D:\projects\coregx\fursy\.claude\STATUS.md`

### 2. `.claude/LINTER_RULES.md` 🔍 BEFORE WRITING CODE

**Why**: Prevents repeating common linter errors. Code MUST pass `golangci-lint run` with 0 issues!

**What's inside**:
- 8-point checklist for pre-commit validation
- 31 common linter errors with fixes
- Real examples from radix tree implementation
- Quick reference for gocritic, godot, revive, staticcheck, unparam

**Critical rules**:
- Use `s == ""` instead of `len(s) == 0` (gocritic)
- All comments must end with period (godot)
- Add package comment to every file (revive)
- No if-else after return (indent-error-flow)
- Use `errors.Is()` for error comparison (errorlint)

**Read it**: `D:\projects\coregx\fursy\.claude\LINTER_RULES.md`

### 3. `.claude/CLAUDE.md` 📖 PROJECT CONFIGURATION

**Why**: Complete project configuration and standards for AI agents.

**What's inside**:
- Project vision and goals
- Architecture decisions
- Technical requirements
- Workflow and processes
- Kanban organization

**Read it**: `D:\projects\coregx\fursy\.claude\CLAUDE.md`

### 4. `README.md` 📚 PUBLIC DOCUMENTATION

**Why**: User-facing documentation showing what fursy does and how to use it.

**Read it**: `D:\projects\coregx\fursy\README.md`

---

## Architecture

### Radix Tree Routing Engine

**Location**: `internal/radix/`

**What it does**: High-performance routing using radix tree algorithm (inspired by httprouter).

**Features**:
- Zero-allocation parameter extraction
- Supports: static paths, named params (`:id`), wildcards (`*path`)
- Optional regex constraints: `/users/:id(\\d+)`
- O(log n) lookup complexity
- 87.9% test coverage

**Performance**: 256 ns/op (static), 326 ns/op (parametric), 1 alloc/op

### Generic Type-Safe Handlers

**Innovation**: First Go router with compile-time type safety for request/response.

**Traditional approach** (Gin, Echo, Fiber):
```go
func CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    // Manual validation needed for every field!
}
```

**fursy approach** (type-safe):
```go
func CreateUser(box *fursy.Box[CreateUserRequest, UserResponse]) error {
    if err := box.Bind(); err != nil {
        return err  // Automatic RFC 9457 error
    }
    // box.ReqBody is VALIDATED and type-safe!
    user := createUser(box.ReqBody)
    return box.Created("/users/"+user.ID, user)
}
```

**Key types**:
```go
// Generic context
type Box[Req, Res any] struct {
    *Context          // Embedded base context
    ReqBody *Req      // Validated request (after Bind())
    ResBody *Res      // Response to serialize
}

// Generic handler
type Handler[Req, Res any] func(*Box[Req, Res]) error
```

### Middleware Pipeline

**Pattern**: Express-like `Next()/Abort()` pattern.

**Architecture**:
```go
type HandlerFunc func(*Context) error

func Middleware() HandlerFunc {
    return func(c *Context) error {
        // Before handler
        start := time.Now()

        // Call next middleware/handler
        err := c.Next()

        // After handler (can access response)
        duration := time.Since(start)

        return err
    }
}
```

**Execution flow**:
1. Router receives request
2. Middleware chain executes (pre-processing)
3. Route handler executes
4. Middleware chain unwinds (post-processing)
5. Response sent

**Optimization**: Pre-allocated middleware chain buffer (no allocations).

### RFC 9457 Problem Details

**Standard error format** built into fursy:

```go
type Problem struct {
    Type       string         `json:"type"`
    Title      string         `json:"title"`
    Status     int            `json:"status"`
    Detail     string         `json:"detail,omitempty"`
    Instance   string         `json:"instance,omitempty"`
    Extensions map[string]any `json:"-"`  // Additional fields
}
```

**Example response**:
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

**Why RFC 9457?**
- Standardized across all APIs
- Machine-readable error types
- Consistent structure
- Extension support for custom fields

### OpenAPI 3.1 Generation

**Automatic generation** from code:

```go
spec := router.OpenAPI(fursy.OpenAPIConfig{
    Title:       "My API",
    Version:     "1.0.0",
    Description: "API description",
})
// Returns complete OpenAPI 3.1 spec
```

**How it works**:
- Introspects route definitions
- Extracts types from generic handlers `Box[Req, Res]`
- Generates schemas from struct tags
- Includes validation rules
- Produces JSON/YAML spec

### Content Negotiation (RFC 9110)

**Automatic format selection** based on `Accept` header:

```go
router.GET("/users/:id", func(c *fursy.Context) error {
    user := getUser(c.Param("id"))
    // Automatically picks JSON/HTML/XML/Markdown based on Accept header
    return c.Negotiate(200, user)
})
```

**Supported formats**:
- JSON (`application/json`) - default
- HTML (`text/html`) - web browsers
- XML (`application/xml`) - legacy systems
- Plain text (`text/plain`)
- Markdown (`text/markdown`) - AI agents

**Quality values** (q-parameter):
```
Accept: text/html;q=0.9, application/json;q=1.0
// fursy automatically picks JSON (higher q-value)
```

---

## Project Structure

```
fursy/
├── .claude/                        # AI configuration (PRIVATE)
│   ├── CLAUDE.md                  # Project config ⭐
│   ├── STATUS.md                  # Current status ⭐⭐
│   └── LINTER_RULES.md            # Linter checklist ⭐
│
├── docs/
│   ├── PERFORMANCE.md             # Performance benchmarks
│   └── dev/                       # PRIVATE (in .gitignore)
│       ├── 00_PROJECT_VISION.md
│       ├── 02_TECHNICAL_ARCHITECTURE.md
│       ├── 03_API_SPECIFICATION.md
│       ├── 04_IMPLEMENTATION_PLAN.md
│       └── kanban/                # Task management
│
├── internal/                       # Internal implementation (NOT in Go Docs)
│   ├── radix/                     # Radix tree routing (87.9% coverage)
│   │   ├── tree.go               # Main tree implementation
│   │   ├── node.go               # Tree nodes
│   │   └── tree_test.go
│   ├── binding/                   # Request binding utilities
│   └── negotiate/                 # Content negotiation
│
├── middleware/                     # Built-in middleware (8 total)
│   ├── logger.go                  # log/slog logging (zero deps)
│   ├── recovery.go                # Panic recovery (zero deps)
│   ├── cors.go                    # CORS headers (zero deps)
│   ├── basicauth.go               # Basic Auth (zero deps)
│   ├── jwt.go                     # JWT auth (golang-jwt/jwt dep)
│   ├── ratelimit.go               # Rate limiting (x/time dep)
│   ├── circuitbreaker.go          # Circuit breaker (zero deps, 95.5%)
│   └── secure.go                  # Security headers (zero deps, 100%)
│
├── plugins/                        # Optional plugins (separate modules)
│   ├── opentelemetry/             # OpenTelemetry tracing + metrics
│   │   ├── middleware.go         # Tracing middleware (92% coverage)
│   │   ├── metrics.go            # Metrics middleware (89.9% coverage)
│   │   ├── go.mod                # Separate module
│   │   └── README.md
│   └── validator/                 # go-playground/validator integration
│       ├── validator.go          # Validator plugin (94.3% coverage)
│       ├── go.mod                # Separate module
│       └── README.md
│
├── examples/                       # Usage examples
│   ├── 01-hello-world/           # Basic routing
│   ├── 02-rest-api-crud/         # Full CRUD API
│   ├── 04-content-negotiation/   # Multi-format responses
│   ├── 05-middleware/            # Middleware usage
│   └── validation/               # 6 validation examples
│       ├── 01-basic/
│       ├── 02-rest-api-crud/
│       ├── 03-custom-validator/
│       ├── 04-nested-structs/
│       ├── 05-custom-messages/
│       └── 06-production/
│
├── Core files (public API - what users see in Go Docs)
├── router.go                      # Main Router type
├── context.go, context_base.go   # Non-generic Context
├── context_generic.go             # Generic Box[Req, Res]
├── handler.go, handler_generic.go # Handler types
├── group.go                       # Route groups
├── error.go, problem.go           # RFC 9457 errors
├── openapi.go                     # OpenAPI generation
├── validation.go                  # Request validation
├── negotiate.go                   # Content negotiation
├── version.go                     # API versioning
│
├── README.md                      # Public documentation
├── go.mod                         # Module definition
├── .gitignore
└── llms.md                        # This file
```

### Key Directories

**`.claude/`** (PRIVATE):
- AI agent configuration
- Current status and metrics
- Linter rules and checklist
- NOT committed to git

**`internal/`**:
- Implementation details
- NOT visible in Go Docs
- NOT importable by external packages
- Contains radix tree, binding, negotiation logic

**`middleware/`**:
- 8 built-in middleware
- 4 zero-dependency (logger, recovery, cors, basicauth)
- 2 minimal deps (jwt, ratelimit)
- 2 advanced (circuitbreaker, secure)

**`plugins/`**:
- Optional extensions
- Separate Go modules
- Can have dependencies
- OpenTelemetry, Validator

**`examples/`**:
- 11 complete examples
- Progressive complexity
- Runnable code
- Full documentation

---

## Core Concepts

### 1. Type-Safe Handlers with Box[Req, Res]

**Problem**: Traditional routers have no compile-time type safety.

**Solution**: Generic handlers with `Box[Req, Res]`.

```go
// Define request/response types
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=3"`
    Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Type-safe handler
router.POST[CreateUserRequest, UserResponse]("/users",
    func(box *fursy.Box[CreateUserRequest, UserResponse]) error {
        // Bind and validate request
        if err := box.Bind(); err != nil {
            return err  // Automatic RFC 9457 response
        }

        // box.ReqBody is type-safe and validated!
        user := createUser(box.ReqBody)

        // Type-safe response
        return box.Created("/users/"+user.ID, user)
    })
```

**Benefits**:
- Compile-time type checking
- Automatic validation
- Clear API contracts
- Self-documenting code

### 2. Middleware Architecture

**Express-like pattern** with `Next()/Abort()`:

```go
// Logger middleware example
func Logger() fursy.HandlerFunc {
    return func(c *fursy.Context) error {
        start := time.Now()

        // Call next middleware/handler
        err := c.Next()

        // Log after handler executes
        slog.Info("request",
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "status", c.Response.Status(),
            "duration", time.Since(start),
        )

        return err
    }
}

// Use middleware
router.Use(Logger())
```

**Key methods**:
- `c.Next()` - Execute next middleware/handler
- `c.Abort()` - Stop execution (e.g., auth failure)
- `c.IsAborted()` - Check if aborted

### 3. Route Groups

**Organize routes** with shared middleware and prefixes:

```go
router := fursy.New()

// Public routes
public := router.Group("/api/v1")
public.GET("/health", healthCheck)

// Protected routes with JWT
auth := public.Group("/admin")
auth.Use(middleware.JWT(jwtSecret))
auth.GET("/users", listUsers)      // /api/v1/admin/users
auth.POST("/users", createUser)    // /api/v1/admin/users
```

**Nesting**: Groups can nest infinitely, inheriting all parent middleware.

### 4. Error Handling (RFC 9457)

**Standard error responses** everywhere:

```go
// Built-in helpers
return fursy.NotFound("User not found")
return fursy.BadRequest("Invalid email format")
return fursy.Unauthorized("Invalid credentials")
return fursy.ValidationError("Name is required", map[string]string{
    "name": "Name must be at least 3 characters",
})

// Custom Problem
return fursy.Problem{
    Type:   "https://api.example.com/problems/rate-limit",
    Title:  "Rate Limit Exceeded",
    Status: 429,
    Detail: "You have exceeded 100 requests per hour",
    Extensions: map[string]any{
        "retryAfter": 3600,
    },
}
```

### 5. Validation

**Automatic validation** through validator plugin:

```go
// Setup (once)
router.SetValidator(validator.New())

// Request struct with validation tags
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
    Password string `json:"password" validate:"required,min=8"`
}

// Handler - validation is automatic!
router.POST[CreateUserRequest, UserResponse]("/users",
    func(box *fursy.Box[CreateUserRequest, UserResponse]) error {
        if err := box.Bind(); err != nil {
            return err  // Returns 422 with RFC 9457 error details
        }
        // box.ReqBody is validated! ✅
    })
```

**100+ validation tags**: required, email, url, uuid, min, max, gte, lte, len, etc.

### 6. Content Negotiation

**Multi-format responses** based on client preferences:

```go
router.GET("/users/:id", func(c *fursy.Context) error {
    user := getUser(c.Param("id"))

    // Option 1: Automatic selection
    return c.Negotiate(200, user)  // Picks JSON/HTML/XML/Markdown

    // Option 2: Manual check
    if c.Accepts(fursy.MIMETextMarkdown) {
        return c.Markdown(formatMarkdown(user))  // AI-friendly
    }
    return c.JSON(200, user)  // Default
})
```

**AI agent support**: Markdown responses for better LLM understanding.

---

## Quick Start for AI Agents

### Minimal Example

```go
package main

import (
    "log"
    "net/http"
    "github.com/coregx/fursy"
)

func main() {
    router := fursy.New()

    // Simple handler
    router.GET("/", func(c *fursy.Context) error {
        return c.Text("Hello, fursy!")
    })

    // Parametric route
    router.GET("/users/:id", func(c *fursy.Context) error {
        id := c.Param("id")
        return c.OK(map[string]string{"id": id})
    })

    log.Fatal(http.ListenAndServe(":8080", router))
}
```

### Type-Safe Example

```go
type GetUserRequest struct {
    ID string `param:"id" validate:"required,uuid"`
}

type UserResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

router.GET[GetUserRequest, UserResponse]("/users/:id",
    func(box *fursy.Box[GetUserRequest, UserResponse]) error {
        if err := box.Bind(); err != nil {
            return err
        }

        user := getUserByID(box.ReqBody.ID)
        return box.OK(user)
    })
```

### With Middleware

```go
router := fursy.New()

// Global middleware
router.Use(middleware.Logger())
router.Use(middleware.Recovery())

// CORS
router.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
}))

// Protected routes
admin := router.Group("/admin")
admin.Use(middleware.JWT(jwtSecret))
admin.GET("/users", listUsers)
```

---

## Development Standards

### 1. JSON: encoding/json/v2 ⚠️ CRITICAL

**MUST use** the new `encoding/json/v2` package:

```go
// ✅ CORRECT:
import "encoding/json/v2"

// ❌ WRONG:
import "encoding/json"  // Old version, don't use!
```

**Why**: Go 1.25+ has new JSON API with better performance and features.

### 2. Logging: log/slog

**MUST use** structured logging:

```go
import "log/slog"

// ✅ CORRECT:
slog.Info("request processed",
    "method", req.Method,
    "path", req.URL.Path,
    "duration", duration,
)

// ❌ WRONG:
log.Printf("Request %s %s took %v", req.Method, req.URL.Path, duration)
```

### 3. Minimal Dependencies Policy

**Core package** (`github.com/coregx/fursy`):
- Routing, Context, Groups: ONLY stdlib (0 external deps)
- Middleware exceptions (justified):
  - `middleware/jwt.go`: `golang-jwt/jwt/v5` (JWT standard)
  - `middleware/ratelimit.go`: `golang.org/x/time` (token bucket)
- All other middleware: ONLY stdlib

**Plugins** (`plugins/`):
- Can have dependencies (isolated)
- OpenTelemetry, validators, etc.

**Rule**: NEVER add dependencies to core without discussion!

### 4. Code Quality Standards

**Follow Relica philosophy** (sister project):
1. **Zero dependencies** (core)
2. **Type safety** (generics where appropriate)
3. **Comprehensive testing** (>85% coverage)
4. **Clear API** (intuitive, well-documented)
5. **Performance** (benchmarks for everything)

### 5. Naming Conventions

```go
// Types: PascalCase
type Router struct {}
type Context[Req, Res any] struct {}

// Functions/Methods: PascalCase (exported), camelCase (internal)
func New() *Router
func (r *Router) ServeHTTP(...)
func internalHelper() {}

// Constants: PascalCase or UPPER_SNAKE
const StatusOK = 200
const DefaultTimeout = 30 * time.Second
```

### 6. Error Handling

**Always use RFC 9457**:

```go
// Built-in helpers
return fursy.NotFound("User not found")
return fursy.BadRequest("Invalid input")

// Custom errors
return fursy.Problem{
    Type:   "https://example.com/problems/custom",
    Title:  "Custom Error",
    Status: 400,
}
```

### 7. API Architecture Pattern

**Public API = Wrapper** (clean Go Docs):

```go
// ✅ CORRECT: router.go (root package)
package fursy

import "github.com/coregx/fursy/internal/radix"

type Router struct {
    tree *radix.Tree  // Internal implementation hidden
}

func (r *Router) GET(path string, handler Handler) {
    // Delegate to internal
    r.tree.Insert(path, handler)
}

// ❌ WRONG: Exposing internal types
type Router struct {
    Tree *radix.Tree  // Don't export internal!
}
```

**Why**:
- Root package = clean, simple API for users
- `internal/` = complex implementation, NOT in Go Docs
- Can change implementation without breaking changes

---

## Testing Requirements

### Coverage Targets

| Phase | Target | Status |
|-------|--------|--------|
| Phase 1 (Foundation) | >85% | ✅ 88.9% |
| Phase 2 (API Excellence) | >88% | ✅ 88.9% |
| Phase 3 (Production) | >90% | ✅ 88.9% core + 95.5% circuitbreaker + 100% secure |
| Ongoing | Maintain >85% | ✅ 88.9% |

### Required Commands

```bash
# Run all tests
go test ./...

# Run with race detector (REQUIRED before commit!)
go test -race ./...

# Coverage report
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt

# Run specific test
go test -v -run TestRouter_GET ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Lint (MUST be 0 issues!)
golangci-lint run

# Pre-release check (all validations)
bash scripts/pre-release-check.sh
```

### Test Structure

**Table-driven tests**:

```go
func TestRouter_GET(t *testing.T) {
    tests := []struct {
        name       string
        path       string
        wantStatus int
        wantBody   string
    }{
        {
            name:       "simple route",
            path:       "/users",
            wantStatus: 200,
            wantBody:   `{"users":[]}`,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### Benchmark Requirements

**Always benchmark critical paths**:

```go
func BenchmarkRouter_StaticRoute(b *testing.B) {
    r := fursy.New()
    r.GET("/users", handler)

    req := httptest.NewRequest("GET", "/users", nil)
    w := httptest.NewRecorder()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        r.ServeHTTP(w, req)
    }
}
```

**Current performance targets** (achieved):
- Static routes: <500 ns/op ✅ (256 ns/op)
- Parametric routes: <500 ns/op ✅ (326 ns/op)
- Allocations: 1 alloc/op ✅
- Throughput: >100k req/s ✅ (~10M req/s)

---

## Git Workflow

### Branching Model

**Git-flow pattern**:

```bash
main          # Production releases (v1.0.0, v1.1.0, etc.)
  ├── develop # Development branch (default)
      ├── feature/TASK-XXX-description
      ├── bugfix/issue-123
      └── hotfix/critical-fix
```

### Commit Convention

**Format**: `<type>(<scope>): <subject>`

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `test`: Tests
- `refactor`: Code refactoring
- `perf`: Performance improvement
- `chore`: Maintenance

**Examples**:
```bash
feat(router): add generic Context[Req, Res]
fix(radix): handle edge case with wildcards
docs: update STATUS.md with Phase 3 progress
test(router): add benchmark for parametric routes
refactor(context): simplify error handling
perf(radix): optimize parameter extraction
```

### Important Rules

**❌ NEVER**:
- Add `Co-Authored-By: Claude <noreply@anthropic.com>` to commits
- Commits must look professional and from a real developer
- Claude is already in contributors

**✅ ALWAYS**:
- Run `go test -race ./...` before commit
- Run `golangci-lint run` (must be 0 issues)
- Update `.claude/STATUS.md` for significant changes
- Use clean, concise commit messages

### Merge Strategy

**feature → develop**: Squash merge
**bugfix → develop**: Squash merge
**develop → main**: --no-ff merge (preserve history)

---

## Middleware

**8 built-in middleware** (production-ready):

### 1. Logger

**Package**: `middleware.Logger()`
**Dependencies**: Zero (uses `log/slog`)
**Coverage**: High

**Usage**:
```go
router.Use(middleware.Logger())
```

**Features**:
- Structured logging with slog
- Request/response details
- Duration tracking
- Error logging

### 2. Recovery

**Package**: `middleware.Recovery()`
**Dependencies**: Zero
**Coverage**: High

**Usage**:
```go
router.Use(middleware.Recovery())
```

**Features**:
- Panic recovery
- Stack trace logging
- RFC 9457 error response
- Prevents server crashes

### 3. CORS

**Package**: `middleware.CORS(config)`
**Dependencies**: Zero
**Coverage**: High

**Usage**:
```go
router.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Content-Type", "Authorization"},
}))
```

**Features**:
- Configurable origins, methods, headers
- Credentials support
- Max age control

### 4. BasicAuth

**Package**: `middleware.BasicAuth(users)`
**Dependencies**: Zero
**Coverage**: High

**Usage**:
```go
users := map[string]string{
    "admin": "secret",
    "user":  "password",
}
router.Use(middleware.BasicAuth(users))
```

**Features**:
- HTTP Basic Authentication
- User/password validation
- Realm support

### 5. JWT

**Package**: `middleware.JWT(secret)`
**Dependencies**: `golang-jwt/jwt/v5`
**Coverage**: 94.2%

**Usage**:
```go
router.Use(middleware.JWT([]byte("secret")))
```

**Features**:
- JWT token validation
- Claims extraction
- Custom validation
- Bearer token support

### 6. RateLimit

**Package**: `middleware.RateLimit(config)`
**Dependencies**: `golang.org/x/time`
**Coverage**: 94.4%

**Usage**:
```go
router.Use(middleware.RateLimit(middleware.RateLimitConfig{
    RequestsPerSecond: 10,
    Burst:             20,
}))
```

**Features**:
- Token bucket algorithm
- Per-IP or per-user limiting
- RFC 6585 headers (RateLimit-*)
- 429 Too Many Requests

### 7. CircuitBreaker

**Package**: `middleware.CircuitBreaker(config)`
**Dependencies**: Zero
**Coverage**: 95.5%

**Usage**:
```go
router.Use(middleware.CircuitBreaker(middleware.CircuitBreakerConfig{
    Threshold:   5,           // Open after 5 failures
    Timeout:     30 * time.Second,
    MaxRequests: 3,           // Half-open with 3 requests
}))
```

**Features**:
- Failure threshold
- Auto-recovery (half-open state)
- Timeout-based reset
- 503 Service Unavailable when open

### 8. Secure

**Package**: `middleware.Secure(config)`
**Dependencies**: Zero
**Coverage**: 100%

**Usage**:
```go
router.Use(middleware.Secure(middleware.SecureConfig{
    ContentSecurityPolicy: "default-src 'self'",
    HSTSMaxAge:           31536000,  // 1 year
}))
```

**Features**:
- OWASP 2025 compliant
- CSP, HSTS, X-Frame-Options, etc.
- XSS protection
- Clickjacking prevention

---

## Examples

**11 complete examples** demonstrating all features:

### Basic Examples

**01-hello-world** (`examples/01-hello-world/`):
- Simple routing
- Basic handlers
- Query parameters
- ~100 LOC

**02-rest-api-crud** (`examples/02-rest-api-crud/`):
- Full CRUD operations
- Type-safe handlers
- RFC 9457 errors
- ~340 LOC

**04-content-negotiation** (`examples/04-content-negotiation/`):
- Multi-format responses
- Accept header handling
- AI agent support (Markdown)
- ~200 LOC

**05-middleware** (`examples/05-middleware/`):
- All 8 middleware
- Route groups
- JWT authentication
- ~300 LOC

### Validation Examples

**validation/** (`examples/validation/`):

6 progressive examples showing validation:

1. **01-basic**: Minimal validation demo (~60 LOC)
2. **02-rest-api-crud**: Full CRUD with validation (~340 LOC)
3. **03-custom-validator**: Custom validation rules (~120 LOC)
4. **04-nested-structs**: Nested struct validation (~100 LOC)
5. **05-custom-messages**: Custom error messages (~100 LOC)
6. **06-production**: Production-ready setup (~400 LOC)

**Total**: 17 Go source files, ~1400 LOC code, ~2200 lines documentation

---

## Comparison with Other Routers

### vs Gin, Echo, Fiber, Chi

| Feature | fursy | Gin | Echo | Fiber | Chi |
|---------|-------|-----|------|-------|-----|
| **Type Safety** | ✅ Compile-time (`Box[Req, Res]`) | ❌ Runtime | ❌ Runtime | ❌ Runtime | ❌ Runtime |
| **Auto Validation** | ✅ Set once, validate all | ❌ Manual | ❌ Manual | ❌ Manual | ❌ Manual |
| **RFC 9457 Errors** | ✅ Built-in | ❌ Custom | ❌ Custom | ❌ Custom | ❌ Custom |
| **Content Negotiation** | ✅ RFC 9110 | 🔧 Partial | 🔧 Partial | 🔧 Partial | ❌ No |
| **OpenAPI Generation** | ✅ Built-in | 🔧 Plugin | 🔧 Plugin | 🔧 Plugin | 🔧 Plugin |
| **Zero Deps (core)** | ✅ Yes | ❌ No | ❌ No | ❌ No | ✅ Yes |
| **Performance** | ⭐⭐⭐⭐⭐ 256 ns/op | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Go Version** | 1.25+ | 1.13+ | 1.17+ | 1.17+ | 1.16+ |

**fursy unique advantages**:
1. **Type-safe generics** - Only router with compile-time type safety
2. **RFC 9457 built-in** - Standard error format everywhere
3. **RFC 9110 content negotiation** - Full quality value support
4. **Minimal dependencies** - Core = stdlib only
5. **Modern Go** - Built for Go 1.25+
6. **AI-ready** - Markdown responses for agents

### vs ozzo-routing (Reference)

**fursy builds on ozzo-routing** but adds:

- ✅ Generic type-safe handlers (`Box[Req, Res]`)
- ✅ RFC 9457 Problem Details (standard errors)
- ✅ OpenAPI 3.1 generation (automatic)
- ✅ Better performance (<500ns vs ~38μs)
- ✅ Modern Go 1.25+ features

**Preserved from ozzo**:
- ✅ Middleware pipeline architecture
- ✅ Zero dependencies (core)
- ✅ Clean, simple API

---

## Key Files Reference

### Core Routing

**`router.go`**:
- Main Router type
- Route registration (GET, POST, PUT, DELETE, etc.)
- ServeHTTP implementation
- Wrapper over `internal/radix`

**`internal/radix/tree.go`**:
- Radix tree implementation
- Zero-allocation routing
- Parameter extraction
- 87.9% coverage

**`internal/radix/node.go`**:
- Tree node structure
- Static/param/wildcard nodes
- Edge list management

### Context Types

**`context.go`, `context_base.go`**:
- Non-generic Context type
- Request/response helpers
- Parameter/query access
- Convenience methods (OK, Created, etc.)

**`context_generic.go`**:
- Generic Box[Req, Res] type
- Type-safe request/response
- Bind() for validation
- Generic convenience methods

### Handlers

**`handler.go`**:
- HandlerFunc (non-generic)
- Middleware signature

**`handler_generic.go`**:
- Generic Handler[Req, Res]
- Type-safe handler signature

### Route Groups

**`group.go`**:
- Group type
- Nested grouping
- Middleware inheritance
- Prefix management

### Error Handling

**`error.go`**:
- Error helpers
- Status code constants

**`problem.go`**:
- RFC 9457 Problem type
- Standard error responses
- Extension support

### Features

**`openapi.go`**:
- OpenAPI 3.1 generation
- Type introspection
- Schema generation

**`validation.go`**:
- Validation interface
- Pluggable validators
- Error formatting

**`negotiate.go`**:
- Content negotiation
- Accept header parsing
- Quality value handling

**`version.go`**:
- API versioning
- Version extraction

### Middleware

**`middleware/logger.go`**: Structured logging
**`middleware/recovery.go`**: Panic recovery
**`middleware/cors.go`**: CORS headers
**`middleware/basicauth.go`**: Basic Auth
**`middleware/jwt.go`**: JWT authentication
**`middleware/ratelimit.go`**: Rate limiting
**`middleware/circuitbreaker.go`**: Circuit breaker
**`middleware/secure.go`**: Security headers

### Plugins

**`plugins/opentelemetry/middleware.go`**: Tracing
**`plugins/opentelemetry/metrics.go`**: Metrics
**`plugins/validator/validator.go`**: Validation

---

## Performance Metrics

### Routing Performance

**Static routes**:
- 256 ns/op ✅
- 1 alloc/op ✅
- ~10.5M ops/s throughput

**Parametric routes**:
- 326 ns/op ✅ (1 param)
- 344 ns/op ✅ (2 params)
- 561 ns/op ✅ (4 params - deep nesting)
- 1 alloc/op for all ✅

**Wildcard routes**:
- 539 ns/op ✅
- 1 alloc/op ✅

### Context Operations

**Parameter extraction**:
- 3.7 ns/op ✅
- 0 allocs/op ✅

**Query parsing**:
- 21.8 ns/op ✅
- 0 allocs/op ✅

**Response rendering**:
- String: 145 ns/op, 2 allocs
- JSON: 2512 ns/op, 18 allocs (encoder overhead)

### Middleware Performance

**Logger middleware**:
- ~200 ns/op overhead
- 0 allocs/op (slog)

**Full middleware chain** (8 middleware):
- ~1.7M req/s throughput
- Minimal overhead per middleware

### Comparison with httprouter

**fursy** achieves similar performance to httprouter (industry benchmark) while adding:
- Type-safe handlers
- RFC 9457 errors
- OpenAPI generation
- Content negotiation
- Validation

**See**: `docs/PERFORMANCE.md` for complete benchmarks (19 benchmarks total).

---

## Common Gotchas

### 1. MUST use encoding/json/v2 ⚠️ CRITICAL

```go
// ❌ WRONG:
import "encoding/json"

// ✅ CORRECT:
import "encoding/json/v2"
```

**Why**: Go 1.25+ requires new JSON API.

### 2. MUST use log/slog for logging

```go
// ❌ WRONG:
import "log"
log.Printf("message")

// ✅ CORRECT:
import "log/slog"
slog.Info("message", "key", value)
```

### 3. ALWAYS read LINTER_RULES.md before coding

**Location**: `.claude/LINTER_RULES.md`

**Critical rules**:
- Use `s == ""` instead of `len(s) == 0`
- All comments must end with period
- Add package comment to every file
- No if-else after return
- Use `errors.Is()` for error comparison

**8-point checklist** - use it before every commit!

### 4. internal/ packages are NOT importable externally

```go
// ❌ CANNOT DO (from external package):
import "github.com/coregx/fursy/internal/radix"
// Error: use of internal package not allowed

// ✅ USE PUBLIC API:
import "github.com/coregx/fursy"
```

**Why**: `internal/` is Go special directory - prevents external imports.

### 5. ALWAYS run race detector before commit

```bash
# ❌ INSUFFICIENT:
go test ./...

# ✅ REQUIRED:
go test -race ./...
```

**Why**: Catches data races that normal tests miss.

### 6. Coverage must be >85%

```bash
# Check coverage
go test -coverprofile=coverage.txt ./...
go tool cover -func=coverage.txt | grep total
```

**Current**: 88.9% (exceeds target) ✅

### 7. Golangci-lint MUST pass with 0 issues

```bash
golangci-lint run
# Output: MUST be clean (no issues)
```

**Current**: 0 issues ✅

### 8. Never add dependencies to core without discussion

**Core** (`github.com/coregx/fursy/`):
- Routing, Context, Groups: ONLY stdlib
- Middleware exceptions: jwt (golang-jwt/jwt), ratelimit (x/time)
- Everything else: ONLY stdlib

**Plugins** (`plugins/`):
- Can have dependencies (isolated)

### 9. Commit messages - no AI attribution

```bash
# ❌ WRONG:
git commit -m "feat: add feature

Co-Authored-By: Claude <noreply@anthropic.com>"

# ✅ CORRECT:
git commit -m "feat: add feature"
```

**Why**: Commits must look professional. Claude is already in contributors.

### 10. Empty type for Box[Req, Res]

**When no request/response body needed**:

```go
type Empty struct{}

// DELETE with no body
router.DELETE[Empty, Empty]("/users/:id",
    func(box *fursy.Box[Empty, Empty]) error {
        deleteUser(box.Param("id"))
        return box.NoContentSuccess()  // 204
    })
```

### 11. Bind() is required for validation

```go
router.POST[CreateUserRequest, UserResponse]("/users",
    func(box *fursy.Box[CreateUserRequest, UserResponse]) error {
        // ❌ WRONG: Accessing ReqBody before Bind()
        // user := box.ReqBody  // nil!

        // ✅ CORRECT: Bind first
        if err := box.Bind(); err != nil {
            return err
        }
        user := box.ReqBody  // Now populated and validated!
    })
```

### 12. Middleware order matters

```go
router := fursy.New()

// ✅ CORRECT ORDER:
router.Use(middleware.Recovery())  // 1. Catch panics first
router.Use(middleware.Logger())    // 2. Log everything
router.Use(middleware.CORS())      // 3. CORS before auth
router.Use(middleware.JWT(secret)) // 4. Auth last

// ❌ WRONG ORDER:
router.Use(middleware.JWT(secret)) // Panic after JWT = no logging!
router.Use(middleware.Logger())
router.Use(middleware.Recovery())
```

---

## Contributing Guidelines

### Before Starting Work

1. **Read STATUS.md** (`.claude/STATUS.md`)
   - Current status
   - Active tasks
   - Known issues

2. **Read LINTER_RULES.md** (`.claude/LINTER_RULES.md`)
   - 8-point checklist
   - 31 common errors
   - Code standards

3. **Read CLAUDE.md** (`.claude/CLAUDE.md`)
   - Project configuration
   - Architecture decisions
   - Workflow

4. **Check Kanban** (`docs/dev/kanban/`)
   - See what's in progress
   - Pick task from backlog

### Development Process

1. **Create feature branch**:
   ```bash
   git checkout develop
   git checkout -b feature/TASK-XXX-description
   ```

2. **Write tests first** (TDD):
   ```bash
   # Write test
   go test -v -run TestNewFeature ./...

   # Implement feature
   # ...

   # Verify test passes
   go test -v -run TestNewFeature ./...
   ```

3. **Run full test suite**:
   ```bash
   go test -race ./...
   ```

4. **Check coverage** (must be >85%):
   ```bash
   go test -coverprofile=coverage.txt ./...
   go tool cover -func=coverage.txt | grep total
   ```

5. **Lint** (must be 0 issues):
   ```bash
   golangci-lint run
   ```

6. **Update STATUS.md** (for significant changes):
   ```bash
   # Update metrics, progress, etc.
   vim .claude/STATUS.md
   ```

7. **Commit**:
   ```bash
   git add .
   git commit -m "feat(scope): description"
   ```

8. **Push and merge**:
   ```bash
   git push origin feature/TASK-XXX-description
   # Create PR to develop (squash merge)
   ```

### Pre-Release Checklist

**Before any release**:

```bash
# Full validation
bash scripts/pre-release-check.sh

# Includes:
# - All tests
# - Race detector
# - Coverage check
# - Linter
# - Build verification
```

### Code Review Standards

**Every PR must**:
- Pass all tests
- Pass race detector
- Have >85% coverage
- Pass linter (0 issues)
- Follow naming conventions
- Include benchmarks (for performance-critical code)
- Update STATUS.md (if significant)

---

## Quick Reference

### Essential Commands

```bash
# Testing
go test ./...                              # All tests
go test -race ./...                        # With race detector ⭐
go test -coverprofile=coverage.txt ./...  # Coverage
go test -v -run TestName ./...            # Specific test
go test -bench=. -benchmem ./...          # Benchmarks

# Linting
golangci-lint run                          # Lint (MUST be 0 issues) ⭐

# Pre-release
bash scripts/pre-release-check.sh          # Full validation ⭐

# Build
go build ./...                             # All packages
```

### File Locations

- **STATUS.md**: `.claude/STATUS.md` ⭐⭐
- **LINTER_RULES.md**: `.claude/LINTER_RULES.md` ⭐
- **CLAUDE.md**: `.claude/CLAUDE.md`
- **Performance**: `docs/PERFORMANCE.md`
- **Examples**: `examples/`
- **Middleware**: `middleware/`
- **Plugins**: `plugins/`

### Import Statements

```go
// Core
import "github.com/coregx/fursy"

// Middleware (built-in)
import "github.com/coregx/fursy/middleware"

// Plugins (separate modules)
import "github.com/coregx/fursy/plugins/opentelemetry"
import "github.com/coregx/fursy/plugins/validator"
```

### Key URLs

- **Repository**: `github.com/coregx/fursy`
- **Organization**: `github.com/coregx`
- **Sister Project**: [Relica](https://github.com/coregx/relica)
- **Go Docs**: (will be) `pkg.go.dev/github.com/coregx/fursy`

---

## Summary for AI Agents

### What is fursy?

**fursy** is a production-ready HTTP router for Go 1.25+ that uniquely combines:

1. **Type-safe generic handlers** `Box[Req, Res]` - First in Go ecosystem
2. **RFC 9457 Problem Details** - Standard error format everywhere
3. **OpenAPI 3.1 generation** - Automatic from code
4. **Minimal dependencies** - Core = stdlib only
5. **256 ns/op routing** - Zero-allocation, 1 alloc/op
6. **8 production middleware** - Logger, Recovery, CORS, BasicAuth, JWT, RateLimit, CircuitBreaker, Secure
7. **RFC 9110 content negotiation** - Multi-format responses including Markdown for AI agents

### Current Status (2025-11-17)

- **v0.1.0** released (rebranded from FURY)
- **88.9% test coverage** (150+ tests, 19 benchmarks)
- **0 linter issues** (golangci-lint strict)
- **Phase 3 complete** (Foundation, API Excellence, Production Features)
- **Phase 4 ready** (Ecosystem: docs, community, plugins)

### How AI Agents Should Work

1. **ALWAYS read STATUS.md first** (`.claude/STATUS.md`)
2. **ALWAYS read LINTER_RULES.md before coding** (`.claude/LINTER_RULES.md`)
3. **MUST use `encoding/json/v2`** (not `encoding/json`)
4. **MUST use `log/slog`** (not `log`)
5. **MUST run `go test -race`** before commit
6. **MUST pass `golangci-lint run`** with 0 issues
7. **NO `Co-Authored-By: Claude`** in commits
8. **Maintain >85% coverage** (currently 88.9%)

### Key Differentiators

**Why fursy is unique**:
- Only Go router with compile-time type-safe handlers
- Only router with built-in RFC 9457 (standard error format)
- Only router with full RFC 9110 content negotiation including AI agent support
- Minimal dependencies while providing production features
- Modern Go 1.25+ design from the ground up

### Where to Look

- **Current work**: `.claude/STATUS.md` → "Current Work" section
- **Code standards**: `.claude/LINTER_RULES.md` → 8-point checklist
- **Examples**: `examples/` → 11 complete examples
- **Performance**: `docs/PERFORMANCE.md` → 19 benchmarks
- **API docs**: Code comments (godoc style)

---

**Version**: 1.0
**Last Updated**: 2025-11-17
**Project Version**: v0.1.0 (Production Ready)
**For Questions**: See `.claude/STATUS.md` and `.claude/CLAUDE.md`
