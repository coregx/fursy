# 06-production - Production-Ready Setup

A complete production-ready example with JWT authentication, validation, structured logging, and graceful shutdown.

## What This Demonstrates

- JWT authentication with middleware
- Role-based access control (RBAC)
- Structured logging with `log/slog`
- Configuration from environment variables
- Graceful shutdown with signal handling
- Custom validation messages
- Complete API with protected routes
- Production error handling patterns

## How to Run

```bash
# With default settings
go run .

# With custom configuration
PORT=3000 JWT_SECRET=my-secret LOG_LEVEL=debug go run .
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `JWT_SECRET` | `your-secret-key...` | JWT signing secret |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `ENVIRONMENT` | `development` | Environment (development/staging/production) |
| `READ_TIMEOUT` | `15` | Read timeout in seconds |
| `WRITE_TIMEOUT` | `15` | Write timeout in seconds |

## API Endpoints

### Public Endpoints

#### POST /api/login

Login and receive JWT token.

**Request**:
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

**Response** (200):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "admin@example.com",
    "username": "admin",
    "role": "admin"
  }
}
```

### Protected Endpoints (Require Authentication)

All protected endpoints require `Authorization: Bearer <token>` header.

#### GET /api/profile

Get current user's profile.

**Request**:
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer $TOKEN"
```

**Response** (200):
```json
{
  "id": 1,
  "email": "admin@example.com",
  "username": "admin",
  "role": "admin"
}
```

#### PUT /api/profile

Update current user's profile.

**Request**:
```bash
curl -X PUT http://localhost:8080/api/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newusername",
    "email": "newemail@example.com"
  }'
```

**Validation Error** (422):
```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "Username must be at least 3 characters",
  "errors": {
    "username": "Username must be at least 3 characters"
  }
}
```

### Admin-Only Endpoints (Require `role: admin`)

#### POST /api/users

Create a new user (admin only).

**Request**:
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "johndoe",
    "password": "secret123",
    "role": "user"
  }'
```

**Response** (201):
```json
{
  "id": 2,
  "email": "user@example.com",
  "username": "johndoe",
  "role": "user"
}
```

**Validation Error** (422):
```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "Role must be one of: user admin",
  "errors": {
    "role": "Role must be one of: user admin"
  }
}
```

#### GET /api/users

List all users (admin only).

**Request**:
```bash
curl http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN"
```

**Response** (200):
```json
{
  "users": [
    {
      "id": 1,
      "email": "admin@example.com",
      "username": "admin",
      "role": "admin"
    },
    {
      "id": 2,
      "email": "user@example.com",
      "username": "johndoe",
      "role": "user"
    }
  ],
  "total": 2,
  "page": 1,
  "limit": 100
}
```

## File Structure

```
06-production/
├── main.go         # Application entry, routing, graceful shutdown
├── config.go       # Configuration from environment
├── models.go       # Request/Response types
├── handlers.go     # HTTP handlers
├── middleware.go   # JWT auth and RBAC
└── README.md       # This file
```

**Total Lines of Code**: ~400 LOC

## Key Production Patterns

### 1. JWT Authentication

```go
// Generate token on login
token, err := GenerateToken(cfg.JWTSecret, user)

// Validate token in middleware
protected.Use(AuthMiddleware(cfg.JWTSecret))

// Access user info in handlers
userID := c.Get("user_id").(int)
```

### 2. Role-Based Access Control

```go
// Require admin role
admin := protected.Group("")
admin.Use(RequireRole("admin"))
admin.POST[CreateUserRequest, UserResponse]("/users", HandleCreateUser)
```

### 3. Structured Logging

```go
slog.Info("User logged in", "user_id", user.ID, "username", user.Username)
slog.Error("Failed to create user", "error", err, "email", req.Email)
slog.Warn("Invalid credentials", "email", req.Email)
```

### 4. Graceful Shutdown

```go
// Setup signal handling
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

// Shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

### 5. Configuration Management

```go
// Load from environment
cfg := LoadConfig()

// With defaults
Port: getEnv("PORT", "8080"),
JWTSecret: getEnv("JWT_SECRET", "default-secret"),
```

### 6. Error Handling

```go
// Validation errors (automatic)
if err := c.Bind(); err != nil {
    return err  // 422 with validation details
}

// Business logic errors
if user == nil {
    return c.Problem(fursy.Unauthorized("Invalid credentials"))
}

// Server errors
if err != nil {
    slog.Error("Operation failed", "error", err)
    return c.Problem(fursy.InternalServerError("Internal error"))
}
```

## Complete Workflow

### 1. Start Server

```bash
go run .
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}' \
  | jq -r '.token' > token.txt

TOKEN=$(cat token.txt)
```

### 3. Get Profile

```bash
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Create User (Admin Only)

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email":"user@example.com",
    "username":"johndoe",
    "password":"secret123",
    "role":"user"
  }'
```

### 5. List Users (Admin Only)

```bash
curl http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN"
```

### 6. Graceful Shutdown

Press `Ctrl+C` and observe graceful shutdown:

```
{"time":"...","level":"INFO","msg":"Shutting down server..."}
{"time":"...","level":"INFO","msg":"Server exited"}
```

## Security Considerations

### Production Checklist

- [ ] Change `JWT_SECRET` to strong random value
- [ ] Use HTTPS in production
- [ ] Implement password hashing (bcrypt)
- [ ] Add rate limiting (see `fursy.RateLimit()`)
- [ ] Add CORS middleware (see `fursy.CORS()`)
- [ ] Add request ID tracking
- [ ] Implement proper session management
- [ ] Add audit logging
- [ ] Set up monitoring and alerting
- [ ] Use environment-specific configs

### JWT Best Practices

```go
// 1. Use strong secret (32+ bytes)
JWT_SECRET=$(openssl rand -base64 32)

// 2. Set reasonable expiration
"exp": time.Now().Add(24 * time.Hour).Unix()

// 3. Validate all claims
if !token.Valid {
    return error
}

// 4. Use HTTPS only
// 5. Store tokens securely (HttpOnly cookies)
```

### Password Security (Not Implemented - Add This!)

```go
import "golang.org/x/crypto/bcrypt"

// Hash password before storing
hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 12)

// Verify password on login
err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
```

## Deployment

### Docker

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### Kubernetes

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-config
data:
  PORT: "8080"
  LOG_LEVEL: "info"
  ENVIRONMENT: "production"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: api
        image: your-registry/api:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: api-config
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: api-secrets
              key: jwt-secret
```

## Monitoring

### Health Check Endpoint (Add This!)

```go
router.GET("/health", func(c *fursy.Context) error {
    return c.JSON(200, map[string]string{
        "status": "healthy",
        "version": "1.0.0",
    })
})
```

### Metrics (Add OpenTelemetry)

```go
import "github.com/coregx/fursy/plugins/opentelemetry"

router.Use(opentelemetry.Metrics())
```

## Testing

```bash
# Run all examples
go test ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...
```

## Next Steps

1. Add database (PostgreSQL, MySQL, MongoDB)
2. Implement password hashing (bcrypt)
3. Add OpenTelemetry tracing
4. Implement refresh tokens
5. Add email verification
6. Implement 2FA
7. Add rate limiting per user
8. Implement API key auth
9. Add caching (Redis)
10. Set up CI/CD pipeline

## License

MIT License - see [LICENSE](../../../LICENSE) for details.
