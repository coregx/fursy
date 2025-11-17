# fursy Examples

Welcome to the **fursy** examples directory! This collection of working examples demonstrates every feature of fursy, from basic routing to production-ready patterns.

## Table of Contents

- [Learning Path](#learning-path)
- [Quick Start](#quick-start)
- [Basic Examples](#basic-examples)
- [Advanced Examples](#advanced-examples)
- [Validation Examples](#validation-examples)
- [Additional Resources](#additional-resources)
- [Contributing](#contributing)

---

## Learning Path

### Beginner (Start Here)

If you're new to fursy, follow this path:

1. **[01-hello-world/](#01-hello-world)** - Minimal setup (5 min)
2. **[02-rest-api-crud/](#02-rest-api-crud)** - Complete CRUD API (15 min)
3. **[validation/01-basic/](#validation01-basic)** - Basic validation (10 min)

### Intermediate

After mastering the basics:

4. **[validation/02-rest-api-crud/](#validation02-rest-api-crud)** - CRUD with validation (20 min)
5. **[validation/03-custom-validator/](#validation03-custom-validator)** - Custom validation rules (15 min)
6. **[04-content-negotiation/](#04-content-negotiation)** - Multi-format responses (20 min)
7. **[05-middleware/](#05-middleware)** - All middleware patterns (30 min)

### Advanced

Production-ready features:

8. **[validation/04-nested-structs/](#validation04-nested-structs)** - Nested validation (15 min)
9. **[validation/05-custom-messages/](#validation05-custom-messages)** - Custom error messages (10 min)
10. **[validation/06-production/](#validation06-production)** - Production setup (30 min)
11. **[06-opentelemetry/](#06-opentelemetry)** - Distributed tracing (45 min)

**Total Learning Time**: ~3.5 hours to master all features

---

## Quick Start

### Prerequisites

- **Go 1.25+** (required for fursy)
- Basic HTTP/REST knowledge
- `curl` or similar HTTP client

### Running Any Example

```bash
# Navigate to example directory
cd examples/01-hello-world

# Run the example
go run main.go

# Or build first
go build -o example
./example
```

Most examples start on `http://localhost:8080` by default.

---

## Basic Examples

These examples demonstrate core fursy features with minimal code.

### 01. Hello World

**Difficulty**: 🟢 Beginner
**Concepts**: Basic routing, handlers, JSON responses
**Time**: 5 minutes
**[View Example →](./01-hello-world/)**

The simplest possible fursy application - a single endpoint returning JSON. Perfect starting point!

**What You'll Learn**:
- Creating a fursy router
- Defining routes with `router.GET()`
- Returning JSON with `c.OK()`
- Running the server

**Quick Start**:
```bash
cd examples/01-hello-world
go run main.go
curl http://localhost:8080/
```

**Lines of Code**: ~30 LOC

---

### 02. REST API CRUD

**Difficulty**: 🟢 Beginner
**Concepts**: CRUD operations, type-safe handlers, in-memory database, RESTful patterns
**Time**: 15 minutes
**[View Example →](./02-rest-api-crud/)**

Complete REST API demonstrating Create, Read, Update, Delete operations with type-safe handlers.

**What You'll Learn**:
- Type-safe handlers with `Box[Req, Res]`
- RESTful endpoint design
- HTTP status codes (200, 201, 204, 404, 409)
- RFC 9457 Problem Details for errors
- Thread-safe in-memory database
- Partial updates with PUT

**Quick Start**:
```bash
cd examples/02-rest-api-crud
go run .
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","full_name":"Test User","age":25}'
```

**Lines of Code**: ~340 LOC (main, models, handlers, database)

---

## Advanced Examples

Production-ready features and advanced patterns.

### 04. Content Negotiation

**Difficulty**: 🟡 Intermediate
**Concepts**: HTTP content negotiation, multiple response formats, RFC 9110, q-values
**Time**: 20 minutes
**[View Example →](./04-content-negotiation/)**

Comprehensive demonstration of HTTP content negotiation - serve multiple formats (JSON, XML, HTML, Markdown) from a single endpoint.

**What You'll Learn**:
- `Accepts()` method for format checking
- `AcceptsAny()` for multi-format support with q-value priority
- `Negotiate()` for automatic format selection
- `Markdown()` for AI-friendly responses
- RFC 9110 quality value (q-value) handling
- Why AI agents (Claude, ChatGPT) prefer Markdown

**Quick Start**:
```bash
cd examples/04-content-negotiation
go run .

# Request Markdown (AI-friendly)
curl -H "Accept: text/markdown" http://localhost:8080/api/users

# Request HTML (browser-friendly)
curl -H "Accept: text/html" http://localhost:8080/api/users

# Request JSON (API-friendly)
curl -H "Accept: application/json" http://localhost:8080/api/users
```

**Lines of Code**: ~400 LOC

---

### 05. Middleware

**Difficulty**: 🟡 Intermediate
**Concepts**: All 8 built-in middleware, custom middleware, middleware ordering, group middleware
**Time**: 30 minutes
**[View Example →](./05-middleware/)**

Complete guide to fursy middleware - demonstrates all built-in middleware and custom patterns.

**What You'll Learn**:
- **Built-in Middleware** (8):
  - Logger - Structured request logging with `log/slog`
  - Recovery - Panic recovery with stack traces
  - CORS - Cross-Origin Resource Sharing
  - BasicAuth - HTTP Basic Authentication
  - JWT - JSON Web Token authentication (HS256, RS256, ES256)
  - RateLimit - Token bucket rate limiting
  - CircuitBreaker - Fault tolerance pattern
  - Secure - OWASP 2025 security headers
- **Custom Middleware** (8 examples):
  - RequestID, Timing, Skipper, APIKey, CacheControl, CompressionHint, Version, MethodOverride
- Middleware ordering best practices
- Group middleware for route groups
- Configuration patterns

**Quick Start**:
```bash
cd examples/05-middleware
go run .

# Generate JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/token | jq -r .token)

# Access protected endpoint
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/protected/users

# Test Basic Auth
curl -u admin:secret http://localhost:8080/basic/dashboard
```

**Lines of Code**: ~600 LOC

---

### 06. OpenTelemetry

**Difficulty**: 🔴 Advanced
**Concepts**: Distributed tracing, Jaeger integration, custom spans, W3C Trace Context, HTTP semantic conventions
**Time**: 45 minutes
**[View Example →](./06-opentelemetry/)**

Comprehensive OpenTelemetry instrumentation with Jaeger for distributed tracing and observability.

**What You'll Learn**:
- OpenTelemetry tracing middleware
- Jaeger integration via OTLP/HTTP
- Custom spans for database queries and external calls
- Error recording in traces
- HTTP semantic conventions
- Context propagation across services
- Performance monitoring for slow requests
- Using Jaeger UI for trace visualization

**Quick Start**:
```bash
cd examples/06-opentelemetry

# Start Jaeger (requires Docker)
docker-compose up -d

# Run application
go run main.go

# Generate traffic
curl http://localhost:8080/users/123

# View traces in Jaeger UI
open http://localhost:16686
```

**Lines of Code**: ~500 LOC + docker-compose.yml

**Dependencies**: Docker, Jaeger (via docker-compose)

---

## Validation Examples

The `validation/` directory contains 6 comprehensive examples demonstrating automatic request validation using the fursy validator plugin.

**[View All Validation Examples →](./validation/)**

All validation examples use:
- RFC 9457 Problem Details for validation errors (422 status)
- Type-safe handlers with `Box[Req, Res]`
- Automatic validation via `c.Bind()`
- go-playground/validator v10 under the hood

---

### validation/01-basic

**Difficulty**: 🟢 Beginner
**Concepts**: Basic validation tags, automatic error responses
**Time**: 10 minutes
**[View Example →](./validation/01-basic/)**

Minimal validation example - demonstrates core validation tags and setup.

**What You'll Learn**:
- Setting up validator with `router.SetValidator()`
- Common validation tags: `required`, `email`, `min`, `max`, `gte`, `lte`
- Automatic validation with `c.Bind()`
- RFC 9457 compliant validation errors

**Quick Start**:
```bash
cd examples/validation/01-basic
go run main.go

# Valid request
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","username":"john","age":25,"password":"secret123"}'

# Invalid request (see validation errors)
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"invalid","username":"ab","age":15,"password":"123"}'
```

**Lines of Code**: ~60 LOC

---

### validation/02-rest-api-crud

**Difficulty**: 🟡 Intermediate
**Concepts**: CRUD with validation, create vs update validation, omitempty for partial updates
**Time**: 20 minutes
**[View Example →](./validation/02-rest-api-crud/)**

Complete REST API with validation on create and update operations.

**What You'll Learn**:
- Different validation rules for create vs update
- Using `omitempty` for optional fields in updates
- Validating partial updates (PUT requests)
- Thread-safe database with validation
- RESTful status codes with validation (422 Unprocessable Entity)

**Quick Start**:
```bash
cd examples/validation/02-rest-api-crud
go run .

# Create user with validation
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","username":"johndoe","full_name":"John Doe","age":30,"password":"secret123"}'

# Partial update (only age)
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"age":31}'
```

**Lines of Code**: ~340 LOC

---

### validation/03-custom-validator

**Difficulty**: 🟡 Intermediate
**Concepts**: Custom validation functions, password strength, phone validation, business rules
**Time**: 15 minutes
**[View Example →](./validation/03-custom-validator/)**

Creating and registering custom validation functions for business-specific rules.

**What You'll Learn**:
- Registering custom validators with `v.RegisterCustomValidator()`
- Password strength validation (uppercase, lowercase, digit, special char)
- Phone number format validation
- Business logic validation (allowed domains)
- Using `validator.FieldLevel` API

**Custom Validators Demonstrated**:
- `strong_password` - Password complexity rules
- `phone` - Phone number format (international)
- `company_domain` - Email domain whitelist

**Quick Start**:
```bash
cd examples/validation/03-custom-validator
go run .

# Valid with strong password
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"john@example.com","username":"johndoe","password":"Secret123!","phone":"+1234567890"}'
```

**Lines of Code**: ~200 LOC

---

### validation/04-nested-structs

**Difficulty**: 🟡 Intermediate
**Concepts**: Nested struct validation, slice validation with dive, nested error format
**Time**: 15 minutes
**[View Example →](./validation/04-nested-structs/)**

Validation of deeply nested structures and slices.

**What You'll Learn**:
- Nested struct validation (Address within User)
- Slice validation with `dive` tag
- Validating each element in an array
- Nested error response format (`address.city`, `tags[0]`)
- Deep nesting patterns

**Quick Start**:
```bash
cd examples/validation/04-nested-structs
go run main.go

# Valid nested request
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "name":"John Doe",
    "email":"john@example.com",
    "address":{"street":"123 Main St","city":"San Francisco","state":"CA","zip_code":"94102","country":"US"},
    "tags":["developer","golang"]
  }'
```

**Lines of Code**: ~150 LOC

---

### validation/05-custom-messages

**Difficulty**: 🟢 Beginner
**Concepts**: Custom error messages, placeholders, user-friendly messages, localization
**Time**: 10 minutes
**[View Example →](./validation/05-custom-messages/)**

Customizing validation error messages with placeholders for better UX.

**What You'll Learn**:
- Setting custom messages via `validator.Options`
- Using placeholders: `{field}`, `{param}`, `{value}`
- User-friendly error messages
- Localization-ready patterns (English, Spanish examples)

**Quick Start**:
```bash
cd examples/validation/05-custom-messages
go run main.go

# See custom error messages
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"invalid","username":"ab","age":15,"password":"123"}'
```

**Lines of Code**: ~120 LOC

---

### validation/06-production

**Difficulty**: 🔴 Advanced
**Concepts**: Production setup, JWT auth, RBAC, graceful shutdown, configuration, structured logging
**Time**: 30 minutes
**[View Example →](./validation/06-production/)**

Production-ready example with JWT authentication, validation, structured logging, and graceful shutdown.

**What You'll Learn**:
- JWT authentication with middleware
- Role-based access control (admin, user)
- Structured logging with `log/slog`
- Configuration from environment variables
- Graceful shutdown with signal handling
- Custom validation messages
- Complete API with protected routes
- Production error handling patterns

**Quick Start**:
```bash
cd examples/validation/06-production
go run .

# Login and get JWT token
TOKEN=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}' \
  | jq -r '.token')

# Get profile
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer $TOKEN"

# Create user (admin only)
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","username":"johndoe","password":"secret123","role":"user"}'
```

**Lines of Code**: ~400 LOC

**Environment Variables**: `PORT`, `JWT_SECRET`, `LOG_LEVEL`, `ENVIRONMENT`

---

## Additional Resources

### Documentation

- **[Main README](../README.md)** - Project overview and quick start
- **[Middleware Documentation](../middleware/)** - Built-in middleware reference
- **[Validator Plugin](../plugins/validator/README.md)** - Complete validation guide
- **[OpenTelemetry Plugin](../plugins/opentelemetry/README.md)** - Observability setup
- **[pkg.go.dev](https://pkg.go.dev/github.com/coregx/fursy)** - API reference

### Middleware Reference

All built-in middleware demonstrated in [05-middleware/](./05-middleware/):

| Middleware | Purpose | Example |
|------------|---------|---------|
| `middleware.Logger()` | Structured request logging | `router.Use(middleware.Logger())` |
| `middleware.Recovery()` | Panic recovery with stack traces | `router.Use(middleware.Recovery())` |
| `middleware.CORS()` | Cross-Origin Resource Sharing | `router.Use(middleware.CORS())` |
| `middleware.BasicAuth()` | HTTP Basic Authentication | `protected.Use(middleware.BasicAuth(accounts))` |
| `middleware.JWT()` | JSON Web Token authentication | `protected.Use(middleware.JWT(secret))` |
| `middleware.RateLimit()` | Token bucket rate limiting | `router.Use(middleware.RateLimit(10, 20))` |
| `middleware.CircuitBreaker()` | Fault tolerance pattern | `external.Use(middleware.CircuitBreaker(...))` |
| `middleware.Secure()` | OWASP 2025 security headers | `router.Use(middleware.Secure())` |

### Validation Tags

Common validation tags (complete list in [validator plugin README](../plugins/validator/README.md)):

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field must be present | `validate:"required"` |
| `email` | Valid email format | `validate:"email"` |
| `min=3` | Minimum length/value | `validate:"min=3"` |
| `max=50` | Maximum length/value | `validate:"max=50"` |
| `gte=18` | Greater than or equal | `validate:"gte=18"` |
| `lte=120` | Less than or equal | `validate:"lte=120"` |
| `alphanum` | Alphanumeric only | `validate:"alphanum"` |
| `url` | Valid URL | `validate:"url"` |
| `dive` | Validate slice elements | `validate:"dive,min=2"` |
| `omitempty` | Skip if empty | `validate:"omitempty,email"` |

---

## Contributing

### Adding New Examples

Want to contribute an example? Great! Follow these guidelines:

1. **Choose a topic** - Find a feature not yet covered
2. **Keep it focused** - One concept per example
3. **Follow the structure**:
   ```
   examples/NN-topic-name/
   ├── main.go         # Main application
   ├── README.md       # Comprehensive guide
   └── go.mod          # Module file with replace directive
   ```
4. **Write a great README**:
   - What This Demonstrates
   - How to Run
   - Example Requests (with curl commands)
   - Key Concepts
   - Quick Start
   - Next Steps
5. **Use `replace` directive** in go.mod:
   ```go
   replace github.com/coregx/fursy => ../../
   ```
6. **Test thoroughly** - Ensure all curl commands work
7. **Update this index** - Add your example to the appropriate section

### Example Template

See [01-hello-world/](./01-hello-world/) for the simplest template.

### Pull Request Guidelines

- Clear description of what the example demonstrates
- Working code with no errors
- Comprehensive README with curl examples
- Add entry to this index (examples/README.md)
- Follow existing code style and patterns

---

## Summary

This examples directory demonstrates:

- ✅ **11 complete examples** covering all fursy features
- ✅ **Progressive learning path** from beginner to advanced
- ✅ **Production-ready patterns** (JWT, RBAC, logging, shutdown)
- ✅ **Comprehensive validation** (6 validation-specific examples)
- ✅ **Advanced features** (content negotiation, middleware, tracing)
- ✅ **Working code** with test commands for every example
- ✅ **~2,500 lines of example code** demonstrating best practices

**Total Learning Time**: ~3.5 hours to master all features

---

**Start your journey**: [01-hello-world/](./01-hello-world/)

**Need help?** Check the [main README](../README.md) or open an [issue on GitHub](https://github.com/coregx/fursy/issues).

**Happy coding with fursy!** 🚀
