# Middleware Example - Comprehensive Guide

This example demonstrates **all** built-in middleware and custom middleware patterns in fursy.

## Table of Contents

1. [What are Middleware?](#what-are-middleware)
2. [Built-in Middleware](#built-in-middleware)
3. [Custom Middleware](#custom-middleware)
4. [Middleware Ordering](#middleware-ordering)
5. [Group Middleware](#group-middleware)
6. [Configuration](#configuration)
7. [Best Practices](#best-practices)
8. [Testing](#testing)
9. [Next Steps](#next-steps)

---

## What are Middleware?

Middleware are functions that execute **before** and **after** your route handlers. They form a **chain** where each middleware can:

- **Modify** the request or response
- **Validate** authentication, rate limits, etc.
- **Log** request information
- **Recover** from panics
- **Abort** the chain (return early)

### Middleware Pattern

```go
func MyMiddleware() fursy.HandlerFunc {
    return func(c *fursy.Context) error {
        // BEFORE handler (preprocessing)

        err := c.Next() // Call next middleware or handler

        // AFTER handler (postprocessing)

        return err
    }
}
```

### Why Use Middleware?

- **Separation of concerns** - Keep handlers focused on business logic
- **Reusability** - Write once, use everywhere
- **Composability** - Stack middleware for complex behavior
- **Maintainability** - Change cross-cutting concerns in one place

---

## Built-in Middleware

Fursy includes **8 production-ready** middleware covering common use cases:

### 1. Logger - Structured Request Logging

Logs HTTP requests with structured fields using `log/slog`.

**Features:**
- Method, path, status, latency, IP, bytes
- Configurable log levels (INFO/WARN/ERROR based on status)
- Skip paths (e.g., `/health`)
- Custom logger support

**Usage:**

```go
router.Use(middleware.Logger())
```

**With config:**

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Logger: logger,
    SkipPaths: []string{"/health", "/metrics"},
}))
```

**Output (JSON):**

```json
{
  "time": "2025-01-13T10:00:00Z",
  "level": "INFO",
  "msg": "HTTP request",
  "method": "GET",
  "path": "/api/users",
  "status": 200,
  "latency_ms": 12.5,
  "ip": "192.168.1.1",
  "bytes": 1234
}
```

---

### 2. Recovery - Panic Recovery

Recovers from panics and returns HTTP 500 with stack traces.

**Features:**
- Recovers panics without crashing
- Stack trace logging
- Configurable stack trace size
- Separate logger for errors

**Usage:**

```go
router.Use(middleware.Recovery())
```

**With config:**

```go
router.Use(middleware.RecoveryWithConfig(middleware.RecoveryConfig{
    Logger: slog.New(slog.NewJSONHandler(os.Stderr, nil)),
    DisableStackTrace: false, // Enable for debugging
    StackTraceSize: 8192,     // 8KB stack trace
}))
```

---

### 3. CORS - Cross-Origin Resource Sharing

Handles CORS preflight requests and sets CORS headers.

**Features:**
- Whitelist origins
- Configure methods, headers, credentials
- Preflight OPTIONS handling
- Max age caching

**Usage:**

```go
router.Use(middleware.CORS())
```

**With config:**

```go
router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: "https://example.com,https://foo.com",
    AllowMethods: "GET,POST,PUT,DELETE",
    AllowHeaders: "Content-Type,Authorization",
    ExposeHeaders: "X-Request-ID",
    AllowCredentials: true,
    MaxAge: 12 * time.Hour,
}))
```

---

### 4. BasicAuth - HTTP Basic Authentication

Validates username:password from Authorization header.

**Features:**
- RFC 7617 compliant
- Custom validator function
- Realm configuration
- Skipper support

**Usage:**

```go
accounts := map[string]string{
    "admin": "secret",
    "user":  "password123",
}
router.Use(middleware.BasicAuth(middleware.BasicAuthAccounts(accounts)))
```

**Access user in handler:**

```go
router.GET("/dashboard", func(c *fursy.Context) error {
    username := c.GetString(middleware.UserContextKey)
    return c.OK(map[string]string{"user": username})
})
```

**Test:**

```bash
curl -u admin:secret http://localhost:8080/basic/dashboard
```

---

### 5. JWT - JSON Web Token Authentication

Validates JWT tokens with modern security practices (2025).

**Features:**
- HS256, RS256, ES256 support
- Forbids "none" algorithm (security)
- Validates exp, nbf, iat claims
- Issuer and audience validation
- Custom claims support
- Multiple token sources (header, query, cookie)

**Usage (HS256):**

```go
secret := []byte("my-secret-key")
router.Use(middleware.JWT(secret))
```

**Usage (RS256):**

```go
publicKey, _ := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
router.Use(middleware.JWTWithConfig(middleware.JWTConfig{
    SigningKey: publicKey,
    SigningMethod: "RS256",
}))
```

**Access claims in handler:**

```go
router.GET("/protected", func(c *fursy.Context) error {
    claims := c.Get(middleware.JWTContextKey).(jwt.MapClaims)
    userID := claims["sub"].(string)
    return c.String(200, "Hello, "+userID)
})
```

**Generate token:**

```bash
curl -X POST http://localhost:8080/auth/token
```

**Use token:**

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/protected/users
```

---

### 6. RateLimit - Token Bucket Rate Limiting

Limits requests per IP/user/key using token bucket algorithm.

**Features:**
- Token bucket algorithm (golang.org/x/time/rate)
- Per-key limiting (IP, user, API key)
- Configurable rate and burst
- X-RateLimit-* headers (RFC draft)
- In-memory or custom store (Redis)
- Automatic cleanup

**Usage:**

```go
router.Use(middleware.RateLimitWithConfig(middleware.RateLimitConfig{
    Rate:  10,  // 10 requests per second
    Burst: 20,  // Allow bursts up to 20
    KeyFunc: func(c *fursy.Context) string {
        return c.Request.RemoteAddr // Rate limit by IP
    },
}))
```

**Headers:**

```
X-RateLimit-Limit: 10
X-RateLimit-Remaining: 7
X-RateLimit-Reset: 1673600000
```

**Response when limited:**

```
HTTP/1.1 429 Too Many Requests
Retry-After: 5
```

---

### 7. CircuitBreaker - Fault Tolerance

Prevents cascading failures using circuit breaker pattern.

**Features:**
- Three states: Closed, Open, Half-Open
- Configurable failure threshold
- Automatic recovery timeout
- State change callbacks
- Zero dependencies (95.5% coverage)

**Usage:**

```go
router.Use(middleware.CircuitBreakerWithConfig(middleware.CircuitBreakerConfig{
    ConsecutiveFailures: 5,                // Open after 5 failures
    Timeout:             60 * time.Second, // Stay open for 60s
    MaxRequests:         2,                // Allow 2 requests in half-open
    OnStateChange: func(from, to middleware.State) {
        log.Printf("Circuit: %s -> %s", from, to)
    },
}))
```

**States:**

- **Closed**: Normal operation, all requests pass through
- **Open**: Fail fast, reject all requests (503 Service Unavailable)
- **Half-Open**: Test recovery with limited requests

---

### 8. Secure - OWASP 2025 Security Headers

Sets security headers following OWASP best practices.

**Features:**
- X-Frame-Options (clickjacking)
- X-Content-Type-Options (MIME sniffing)
- Content-Security-Policy (XSS)
- Strict-Transport-Security (HTTPS)
- Referrer-Policy
- Cross-Origin policies

**Usage:**

```go
router.Use(middleware.SecureWithConfig(middleware.SecureConfig{
    XFrameOptions:      middleware.XFrameOptionsSameOrigin,
    ContentTypeNosniff: middleware.ContentTypeNosniffValue,
    ReferrerPolicy:     middleware.ReferrerPolicyStrictOrigin,
    HSTSMaxAge:         31536000, // 1 year
    ContentSecurityPolicy: "default-src 'self'",
}))
```

**Headers added:**

```
X-Frame-Options: SAMEORIGIN
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'
Strict-Transport-Security: max-age=31536000
```

---

## Custom Middleware

The example includes **8 custom middleware** demonstrating common patterns:

### 1. RequestID - Request Tracing

Adds unique request ID for distributed tracing.

```go
router.Use(RequestIDMiddleware())

// Access in handler
requestID := c.GetString("request_id")
```

**Headers:**

```
X-Request-ID: a1b2c3d4e5f6g7h8
```

---

### 2. Timing - Response Time

Measures and reports request processing time.

```go
router.Use(TimingMiddleware())

// Access in handler
duration := c.Get("response_time").(time.Duration)
```

**Headers:**

```
X-Response-Time: 12.5ms
```

---

### 3. Skipper - Conditional Middleware

Skip middleware for specific paths.

```go
router.Use(SkipperMiddleware(func(c *fursy.Context) bool {
    return c.Request.URL.Path == "/health"
}))
```

---

### 4. APIKey - API Key Authentication

Validates X-API-Key header.

```go
validKeys := map[string]bool{"key123": true}
router.Use(APIKeyMiddleware(validKeys))
```

---

### 5. CacheControl - HTTP Caching

Sets Cache-Control headers.

```go
// No cache
router.Use(CacheControlMiddleware("no-cache, no-store, must-revalidate"))

// Cache for 1 hour
router.Use(CacheControlMiddleware("public, max-age=3600"))
```

---

### 6. CompressionHint - Vary Header

Hints to proxies about compression.

```go
router.Use(CompressionHintMiddleware())
```

---

### 7. Version - API Versioning

Adds API version to headers.

```go
router.Use(VersionMiddleware("v1.2.3"))
```

---

### 8. MethodOverride - HTTP Method Override

Allows method override for HTML forms.

```go
router.Use(MethodOverrideMiddleware())

// POST /users?_method=DELETE -> DELETE /users
```

---

## Middleware Ordering

**CRITICAL:** Middleware order matters! Apply in this sequence:

### Global Middleware (All Routes)

```go
router.Use(middleware.Logger())      // 1. Log everything (first)
router.Use(middleware.Recovery())    // 2. Recover panics (second)
router.Use(RequestIDMiddleware())    // 3. Request tracing
router.Use(TimingMiddleware())       // 4. Measure time
router.Use(middleware.Secure())      // 5. Security headers
```

### Group Middleware (API/Protected Routes)

```go
// API group
api := router.Group("/api")
api.Use(middleware.CORS())           // 6. CORS for API
api.Use(middleware.RateLimit())      // 7. Rate limiting

// Protected group
protected := api.Group("/protected")
protected.Use(middleware.JWT())      // 8. Authentication
```

### Why This Order?

1. **Logger first** - Captures all requests (including failures)
2. **Recovery second** - Catches panics from all middleware/handlers
3. **Security early** - Apply security headers ASAP
4. **CORS before auth** - Handle preflight OPTIONS before JWT
5. **RateLimit before auth** - Prevent auth DoS
6. **Auth last** - Only for protected routes

---

## Group Middleware

Apply middleware to **route groups** instead of globally:

### Example 1: Public API with CORS

```go
api := router.Group("/api")
api.Use(middleware.CORS())

api.GET("/public", handler)  // Has CORS
```

### Example 2: Protected API with JWT

```go
protected := router.Group("/protected")
protected.Use(middleware.JWT(secret))

protected.GET("/users", handler)  // Requires JWT
```

### Example 3: Admin with BasicAuth

```go
admin := router.Group("/admin")
admin.Use(middleware.BasicAuth(accounts))

admin.GET("/dashboard", handler)  // Requires BasicAuth
```

### Example 4: Nested Groups

```go
api := router.Group("/api")
api.Use(middleware.CORS())
api.Use(middleware.RateLimit())

v1 := api.Group("/v1")
v1.Use(middleware.JWT(secret))

// Final path: /api/v1/users
// Middleware: CORS -> RateLimit -> JWT
v1.GET("/users", handler)
```

---

## Configuration

All built-in middleware support configuration:

### Pattern 1: Default Config

```go
router.Use(middleware.Logger())       // Uses defaults
router.Use(middleware.Recovery())
router.Use(middleware.CORS())
```

### Pattern 2: Custom Config

```go
router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Logger: customLogger,
    SkipPaths: []string{"/health"},
}))
```

### Pattern 3: Shared Config

```go
// Create once, reuse
corsConfig := middleware.CORSConfig{
    AllowOrigins: "https://example.com",
    AllowMethods: "GET,POST",
}

api1.Use(middleware.CORSWithConfig(corsConfig))
api2.Use(middleware.CORSWithConfig(corsConfig))
```

---

## Best Practices

### 1. Middleware Order Matters

Always apply in the recommended order (see above).

### 2. Use Group Middleware

Apply middleware only where needed:

```go
// ❌ BAD: JWT on all routes
router.Use(middleware.JWT(secret))

// ✅ GOOD: JWT only on protected routes
protected := router.Group("/protected")
protected.Use(middleware.JWT(secret))
```

### 3. Skip Health Checks

Don't log/rate-limit health checks:

```go
router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    SkipPaths: []string{"/health", "/metrics"},
}))
```

### 4. Configure CORS Properly

Don't use `AllowOrigins: "*"` with credentials:

```go
// ❌ BAD: Credentials with wildcard
middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: "*",
    AllowCredentials: true, // ERROR!
})

// ✅ GOOD: Specific origins with credentials
middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: "https://example.com",
    AllowCredentials: true,
})
```

### 5. Use Structured Logging

Always use `log/slog` for structured logs:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
    Logger: logger,
}))
```

### 6. Rate Limit by Key

Choose appropriate rate limit key:

```go
// Per-IP
KeyFunc: func(c *fursy.Context) string {
    return c.Request.RemoteAddr
}

// Per-User
KeyFunc: func(c *fursy.Context) string {
    return c.GetString("user_id")
}

// Per-API-Key
KeyFunc: func(c *fursy.Context) string {
    return c.Request.Header.Get("X-API-Key")
}

// Global
KeyFunc: func(c *fursy.Context) string {
    return "global"
}
```

### 7. Handle Errors Properly

Use RFC 9457 Problem Details:

```go
ErrorHandler: func(c *fursy.Context, err error) error {
    return c.Problem(fursy.Problem{
        Type:   "https://example.com/errors/unauthorized",
        Title:  "Unauthorized",
        Status: 401,
        Detail: err.Error(),
    })
}
```

### 8. Use Circuit Breaker for External Services

Wrap external API calls with circuit breaker:

```go
external := router.Group("/external")
external.Use(middleware.CircuitBreaker(...))
external.GET("/api", callExternalAPI)
```

---

## Testing

Test the example with `curl`:

### 1. Home (No Auth)

```bash
curl http://localhost:8080/
```

### 2. Health Check (No Logging)

```bash
curl http://localhost:8080/health
```

### 3. Panic Recovery

```bash
curl http://localhost:8080/panic
# Returns 500 with recovery message
```

### 4. Public API (CORS + RateLimit)

```bash
curl -i http://localhost:8080/api/public
# Check X-RateLimit-* headers
```

### 5. Generate JWT Token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/token | jq -r .token)
echo $TOKEN
```

### 6. Protected API (JWT Required)

```bash
# Without token (fails)
curl http://localhost:8080/api/protected/users

# With token (succeeds)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/protected/users
```

### 7. Basic Auth

```bash
curl -u admin:secret http://localhost:8080/basic/dashboard
```

### 8. Circuit Breaker

```bash
# First 3 requests fail (trigger circuit breaker)
curl http://localhost:8080/circuit/flaky
curl http://localhost:8080/circuit/flaky
curl http://localhost:8080/circuit/flaky

# Circuit opens - returns 503
curl http://localhost:8080/circuit/flaky

# Wait 10 seconds for timeout...
sleep 10

# Circuit half-opens - allows test requests
curl http://localhost:8080/circuit/flaky
```

### 9. Test CORS Preflight

```bash
curl -X OPTIONS \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  http://localhost:8080/api/public
```

### 10. Test Rate Limiting

```bash
# Rapid requests to trigger rate limit
for i in {1..25}; do
  curl -w "\n%{http_code}\n" http://localhost:8080/api/public
done
# Should see 429 after 20 requests
```

---

## Next Steps

Explore more examples:

1. **[Hello World](../01-hello-world/)** - Basic routing
2. **[REST API](../02-rest-api/)** - CRUD operations
3. **[Generic Handlers](../03-generic-handlers/)** - Type-safe handlers
4. **[Error Handling](../04-error-handling/)** - RFC 9457 Problem Details
5. **[Middleware](../05-middleware/)** ← You are here
6. **[OpenAPI](../06-openapi/)** - API documentation
7. **[Performance](../07-performance/)** - Benchmarks and optimization

---

## Summary

This example demonstrated:

- ✅ **8 built-in middleware**: Logger, Recovery, CORS, BasicAuth, JWT, RateLimit, CircuitBreaker, Secure
- ✅ **8 custom middleware**: RequestID, Timing, Skipper, APIKey, CacheControl, CompressionHint, Version, MethodOverride
- ✅ **Middleware ordering** - Critical for correct behavior
- ✅ **Group middleware** - Scoped to route groups
- ✅ **Configuration** - Customizing middleware behavior
- ✅ **Best practices** - Production-ready patterns
- ✅ **Testing** - curl commands for all endpoints

Fursy provides **everything you need** for production HTTP middleware in Go 1.25+!

---

**Next:** Check out [Generic Handlers](../03-generic-handlers/) for type-safe request/response handling!
