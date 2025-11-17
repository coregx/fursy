# Security Policy

## Supported Versions

FURSY HTTP Router is currently in active development (0.x versions). We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.2.x   | :white_check_mark: |
| 0.1.x   | :white_check_mark: |
| < 0.1.0 | :x:                |

Future stable releases (v1.0+) will follow semantic versioning with LTS support.

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability in FURSY HTTP Router, please report it responsibly.

### How to Report

**DO NOT** open a public GitHub issue for security vulnerabilities.

Instead, please report security issues by:

1. **Private Security Advisory** (preferred):
   https://github.com/coregx/fursy/security/advisories/new

2. **Email** to maintainers:
   Create a private GitHub issue or contact via discussions

### What to Include

Please include the following information in your report:

- **Description** of the vulnerability
- **Steps to reproduce** the issue (include proof-of-concept if applicable)
- **Affected versions** (which versions are impacted)
- **Potential impact** (DoS, information disclosure, RCE, etc.)
- **Suggested fix** (if you have one)
- **Your contact information** (for follow-up questions)

### Response Timeline

- **Initial Response**: Within 48-72 hours
- **Triage & Assessment**: Within 1 week
- **Fix & Disclosure**: Coordinated with reporter

We aim to:
1. Acknowledge receipt within 72 hours
2. Provide an initial assessment within 1 week
3. Work with you on a coordinated disclosure timeline
4. Credit you in the security advisory (unless you prefer to remain anonymous)

## Security Considerations for HTTP Routing

HTTP routers process untrusted input from web requests, which introduces security risks.

### 1. Path Traversal Attacks

**Risk**: Malicious paths can exploit routing vulnerabilities.

**Attack Vectors**:
- Path traversal attempts (`../../etc/passwd`)
- Encoded slashes (`%2F`, `%2f`)
- Double encoding (`%252F`)
- Null bytes in paths (`/file%00.txt`)
- Unicode normalization exploits

**Mitigation in Library**:
- ✅ Path normalization before routing
- ✅ Validation of route parameters
- ✅ Sanitization of special characters
- ✅ No automatic file serving (user responsibility)
- ✅ Strict pattern matching in radix tree

**User Recommendations**:
```go
// ❌ BAD - Don't use route parameters directly for filesystem operations
router.GET("/files/:path", func(c *fursy.Context) error {
    filePath := c.Param("path")  // Could be "../../etc/passwd"
    return c.File(filePath)
})

// ✅ GOOD - Validate and sanitize paths
router.GET("/files/:path", func(c *fursy.Context) error {
    filename := filepath.Base(c.Param("path"))
    if !isValidFilename(filename) {
        return c.Error(400, fursy.BadRequest("Invalid filename"))
    }
    safePath := filepath.Join(safeDir, filename)
    return c.File(safePath)
})
```

### 2. Denial of Service (DoS) Attacks

**Risk**: Malicious requests can exhaust server resources.

**Attack Vectors**:
- Slowloris attacks (slow HTTP requests)
- Large request bodies (memory exhaustion)
- Regex DoS in route patterns (if regex added)
- Excessive route registrations
- Deep route nesting

**Mitigation**:
- ✅ Efficient radix tree routing (O(log n) lookups)
- ✅ Zero-allocation routing (1 alloc/op)
- ✅ Context pooling (prevents memory leaks)
- ✅ Graceful shutdown (prevents resource leaks)
- ✅ Circuit breaker middleware (prevents cascade failures)
- ✅ Rate limiting middleware (prevents abuse)

**User Recommendations**:
```go
// ✅ Use built-in middleware for protection
router.Use(fursy.RateLimit(100, time.Minute))  // 100 req/min
router.Use(fursy.CircuitBreaker(0.5, 100))     // 50% error rate threshold

// ✅ Set reasonable timeouts
router.Use(fursy.Timeout(30 * time.Second))

// ✅ Limit request body size
router.Use(fursy.BodyLimit(10 * 1024 * 1024))  // 10MB max
```

### 3. Injection Attacks

**Risk**: Unsanitized input can lead to injection vulnerabilities.

**Attack Vectors**:
- SQL injection (if database middleware used)
- Command injection
- Header injection
- Template injection
- Log injection

**Mitigation**:
- ✅ Parameter extraction is safe (no SQL/command execution)
- ✅ JSON parsing uses `encoding/json/v2` (safe unmarshaling)
- ✅ Header handling through stdlib (validated)
- 🔄 **User Responsibility**: Sanitize data before database/external use

**User Best Practices**:
```go
// ❌ BAD - Don't execute unsanitized input
router.POST("/search", func(c *fursy.Context) error {
    query := c.Query("q")
    // UNSAFE: SQL injection vulnerability
    db.Exec("SELECT * FROM users WHERE name = '" + query + "'")
})

// ✅ GOOD - Use parameterized queries
router.POST("/search", func(c *fursy.Context) error {
    query := c.Query("q")
    db.Exec("SELECT * FROM users WHERE name = ?", query)
})
```

### 4. Authentication & Authorization

**Risk**: Missing or weak authentication/authorization.

**Attack Vectors**:
- Missing authentication checks
- Insecure token storage
- Session fixation
- JWT vulnerabilities (weak secrets, no expiration)

**Mitigation in Library**:
- ✅ JWT middleware with secure defaults
- ✅ Token validation and expiration checks
- ✅ HTTPS enforcement option
- ✅ Secure cookie handling
- 🔄 **User Responsibility**: Implement proper authz logic

**User Best Practices**:
```go
// ✅ Use JWT middleware for authentication
jwtMiddleware := fursy.JWT(fursy.JWTConfig{
    Secret:     os.Getenv("JWT_SECRET"),
    Expiration: 24 * time.Hour,
})

// Protected routes
protected := router.Group("/api", jwtMiddleware)
protected.GET("/users", getUsersHandler)
protected.POST("/users", createUserHandler)

// ✅ Implement authorization in handlers
func getUsersHandler(c *fursy.Context) error {
    user := c.Get("user").(User)
    if !user.IsAdmin() {
        return c.Error(403, fursy.Forbidden("Admin required"))
    }
    // ...
}
```

### 5. Cross-Site Scripting (XSS)

**Risk**: Unsanitized output can lead to XSS.

**Attack Vectors**:
- Reflected XSS (user input in responses)
- Stored XSS (database-stored malicious content)
- DOM-based XSS (client-side rendering)

**Mitigation**:
- ✅ JSON responses automatically escaped (`encoding/json/v2`)
- ✅ Content-Type headers set correctly
- ✅ Security headers middleware (CSP, X-XSS-Protection)
- 🔄 **User Responsibility**: Sanitize HTML/JS output

**User Best Practices**:
```go
// ✅ Use security headers middleware
router.Use(fursy.SecurityHeaders(fursy.SecurityConfig{
    ContentSecurityPolicy: "default-src 'self'",
    XFrameOptions:         "DENY",
    XContentTypeOptions:   "nosniff",
    XSSProtection:         "1; mode=block",
}))

// ✅ Return JSON (auto-escaped)
router.GET("/user/:id", func(c *fursy.Context) error {
    user := getUserByID(c.Param("id"))
    return c.JSON(200, user)  // Automatically escaped
})

// ⚠️ If returning HTML, sanitize first
router.GET("/profile", func(c *fursy.Context) error {
    userInput := c.Query("name")
    sanitized := html.EscapeString(userInput)
    return c.String(200, "<h1>"+sanitized+"</h1>")
})
```

### 6. Cross-Site Request Forgery (CSRF)

**Risk**: Forged requests from malicious sites.

**Mitigation**:
- ✅ CSRF token middleware available
- ✅ SameSite cookie support
- ✅ Origin header validation
- 🔄 **User Responsibility**: Enable CSRF protection

**User Best Practices**:
```go
// ✅ Enable CSRF protection for state-changing operations
csrfMiddleware := fursy.CSRF(fursy.CSRFConfig{
    TokenLength: 32,
    CookieName:  "_csrf",
    HeaderName:  "X-CSRF-Token",
})

router.Use(csrfMiddleware)

// Safe methods (GET, HEAD, OPTIONS) are exempt
// POST, PUT, DELETE require CSRF token
```

## Security Best Practices for Users

### Input Validation

Always validate and sanitize user input:

```go
// Validate request body
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

router.POST("/users", func(c *fursy.Context) error {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        return c.Error(400, fursy.BadRequest("Invalid request"))
    }

    // Additional validation
    if err := validator.Validate(req); err != nil {
        return c.Error(400, fursy.BadRequest(err.Error()))
    }

    // Process validated request
    // ...
})
```

### Rate Limiting

Protect against abuse with rate limiting:

```go
// Global rate limit
router.Use(fursy.RateLimit(1000, time.Hour))

// Per-route rate limits
router.POST("/login",
    fursy.RateLimit(5, time.Minute),  // 5 attempts per minute
    loginHandler,
)
```

### Error Handling

Never leak sensitive information in errors:

```go
// ❌ BAD - Leaks internal details
router.GET("/users/:id", func(c *fursy.Context) error {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", c.Param("id"))
    if err != nil {
        return c.Error(500, fursy.InternalServerError(err.Error()))  // Leaks SQL!
    }
})

// ✅ GOOD - Generic error messages
router.GET("/users/:id", func(c *fursy.Context) error {
    user, err := db.Query("SELECT * FROM users WHERE id = ?", c.Param("id"))
    if err != nil {
        log.Printf("Database error: %v", err)  // Log internally
        return c.Error(500, fursy.InternalServerError("Failed to fetch user"))
    }
})
```

### HTTPS Enforcement

Always use HTTPS in production:

```go
// ✅ Redirect HTTP to HTTPS
router.Use(fursy.HTTPSRedirect())

// ✅ Set secure headers
router.Use(fursy.SecurityHeaders(fursy.SecurityConfig{
    HSTSMaxAge:            31536000,  // 1 year
    HSTSIncludeSubdomains: true,
    HSTSPreload:           true,
}))
```

## Known Security Considerations

### 1. Route Parameter Injection

**Status**: Active mitigation via validation.

**Risk Level**: Medium

**Description**: Route parameters are user-controlled and could contain malicious values.

**Mitigation**:
- All parameters validated before use
- No automatic file serving
- User responsibility to sanitize before external use

### 2. Memory Exhaustion

**Status**: Mitigated via context pooling and limits.

**Risk Level**: Low to Medium

**Description**: Malicious requests could attempt to exhaust server memory.

**Mitigation**:
- Context pooling (prevents allocation spikes)
- Maximum capacity limits (32/64 params/handlers)
- Circuit breaker (prevents cascade failures)
- 🔄 **TODO (v0.4.0)**: Additional memory monitoring

### 3. Regex DoS (If Regex Routes Added)

**Status**: Not applicable (no regex routes yet).

**Risk Level**: N/A (future consideration)

**Description**: Complex regex patterns can cause catastrophic backtracking.

**Mitigation** (if added):
- Regex compilation at route registration (not per-request)
- Timeout limits on regex matching
- Validation of regex patterns

### 4. Dependency Security

FURSY HTTP Router has minimal dependencies:

**Core Routing (zero dependencies)**:
- ✅ stdlib only (routing engine, context, groups)

**Middleware (minimal dependencies)**:
- `github.com/golang-jwt/jwt/v5` (JWT authentication)
- `golang.org/x/time` (rate limiting)
- Other middleware: stdlib only

**Plugins (optional)**:
- `go.opentelemetry.io/otel` (OpenTelemetry plugin)
- `github.com/go-playground/validator/v10` (validation plugin)

**Monitoring**:
- ✅ Dependabot enabled
- ✅ Regular dependency audits
- ✅ No C dependencies (pure Go)

## Security Testing

### Current Testing

- ✅ Unit tests with malicious input
- ✅ Integration tests with attack vectors
- ✅ Linting with 34+ security-focused linters
- ✅ Race detector (`go test -race`)
- ✅ Static analysis (`go vet`)

### Planned for v1.0

- 🔄 Fuzzing with go-fuzz
- 🔄 OWASP ZAP scanning
- 🔄 Static Application Security Testing (SAST)
- 🔄 Dynamic Application Security Testing (DAST)
- 🔄 Penetration testing

## Security Disclosure History

No security vulnerabilities have been reported or fixed yet (project is in 0.x development).

When vulnerabilities are addressed, they will be listed here with:
- **CVE ID** (if assigned)
- **Affected versions**
- **Fixed in version**
- **Severity** (Critical/High/Medium/Low)
- **Credit** to reporter

## Security Contact

- **GitHub Security Advisory**: https://github.com/coregx/fursy/security/advisories/new
- **Public Issues** (for non-sensitive bugs): https://github.com/coregx/fursy/issues
- **Discussions**: https://github.com/coregx/fursy/discussions

## Bug Bounty Program

FURSY HTTP Router does not currently have a bug bounty program. We rely on responsible disclosure from the security community.

If you report a valid security vulnerability:
- ✅ Public credit in security advisory (if desired)
- ✅ Acknowledgment in CHANGELOG
- ✅ Our gratitude and recognition in README
- ✅ Priority review and quick fix

---

**Thank you for helping keep FURSY HTTP Router secure!** 🔒

*Security is a journey, not a destination. We continuously improve our security posture with each release.*
