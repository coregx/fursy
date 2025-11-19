# Production Boilerplate - DDD Architecture

> **Complete, production-ready REST API** with real-time features demonstrating modern Go development practices using **fursy + stream + Relica**.

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 🎯 What is This?

This is a **reference implementation** of a production-ready REST API showcasing:

- **Domain-Driven Design (DDD)** with Rich Models
- **Clean Architecture** (by meaning, not directories)
- **Type-safe HTTP routing** with fursy
- **Real-time communications** (SSE + WebSocket)
- **JWT Authentication** & Authorization
- **Database persistence** with Relica
- **RFC 9457 Problem Details** for errors
- **Docker support** for deployment

### Key Features

✅ User management (registration, login, profile management)
✅ JWT-based authentication
✅ Role-based authorization (user/admin)
✅ Real-time notifications via Server-Sent Events (SSE)
✅ Real-time chat via WebSocket
✅ Database migrations
✅ Docker Compose setup
✅ Graceful shutdown
✅ Structured logging (log/slog)
✅ RFC 9457 error responses

---

## 🏗️ Architecture

### Design Philosophy

**"DDD + Rich Model by meaning, not by directories"**

- **Rich Domain Models**: Business logic lives in entities, not service layers
- **Value Objects**: Enforce invariants (Email, Password, Role, Status)
- **Repository Pattern**: Clean data access abstraction
- **Service Layer**: Orchestrates use cases, delegates to domain models
- **API Layer**: Thin HTTP handlers, converts DTOs ↔ Domain models

### Project Structure

```
10-production-boilerplate/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
│
├── internal/
│   ├── config/                     # Configuration
│   │   └── config.go
│   │
│   ├── user/                       # User bounded context
│   │   ├── entity.go              # User entity (RICH MODEL)
│   │   ├── repository.go          # Data access interface + impl
│   │   ├── service.go             # Business logic / use cases
│   │   ├── api.go                 # HTTP handlers
│   │   └── validation.go          # Request validation
│   │
│   ├── notification/               # Notification bounded context
│   │   ├── entity.go
│   │   ├── service.go
│   │   └── api.go                 # SSE handler
│   │
│   ├── chat/                       # Chat bounded context
│   │   ├── entity.go
│   │   ├── service.go
│   │   └── api.go                 # WebSocket handler
│   │
│   └── shared/                     # Shared kernel
│       ├── database/              # Database connection
│       ├── auth/                  # JWT utilities + middleware
│       └── response/              # HTTP response helpers
│
├── migrations/                     # Database migrations
├── docker-compose.yml              # Docker setup
├── Makefile                        # Build automation
└── .env.example                    # Environment template
```

### Layering (By Meaning)

```
┌─────────────────────────────────────┐
│         API Layer                   │  HTTP handlers, routing
│   (user/api.go, chat/api.go)       │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Service Layer                 │  Use cases, orchestration
│  (user/service.go, chat/service.go) │
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Domain Layer                  │  Business logic, entities
│     (user/entity.go)                │  Value Objects, Rich Models
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│   Infrastructure Layer              │  Database, external services
│   (user/repository.go, database/)   │
└─────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.25+**
- **SQLite** (for local development)
- **Make** (optional, for convenience)
- **Docker** (optional, for containerized deployment)

### 1. Clone & Setup

```bash
# Clone repository
cd examples/10-production-boilerplate

# Copy environment template
cp .env.example .env

# Edit .env with your settings (optional, defaults work for local dev)
# nano .env

# Download dependencies
go mod download
```

### 2. Run Database Migrations

```bash
# Create database and apply migrations
make migrate-up
```

This creates `data/app.db` SQLite database with users and messages tables.

### 3. Run the Application

```bash
# Option 1: Using Go
go run ./cmd/api

# Option 2: Using Make
make run

# Option 3: Build and run binary
make build
./server
```

The API will be available at `http://localhost:8080`.

### 4. Test the API

```bash
# Check health
curl http://localhost:8080/health

# Register a new user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecurePass123",
    "name": "Admin User"
  }'

# Login and get JWT token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecurePass123"
  }'

# Get user profile (requires JWT token)
curl http://localhost:8080/api/users/me \
  -H "Authorization: Bearer <your-jwt-token>"
```

---

## 📡 API Endpoints

### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login and get JWT token |

### User Management (Protected)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/users/me` | Get current user profile | JWT |
| PUT | `/api/users/me` | Update profile | JWT |
| POST | `/api/users/me/password` | Change password | JWT |
| GET | `/api/users` | List all users | Admin only |
| POST | `/api/users/:id/ban` | Ban user | Admin only |
| POST | `/api/users/:id/promote` | Promote to admin | Admin only |

### Real-time Features (Protected)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/notifications/stream` | SSE notification stream | JWT |
| POST | `/api/notifications/broadcast` | Broadcast notification | Admin only |
| GET | `/api/chat/ws` | WebSocket chat connection | JWT |

### Example Requests

#### Register User
```bash
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}

Response: 201 Created
{
  "id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "role": "user",
  "status": "active",
  "created_at": "2025-01-18T..."
}
```

#### Login
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

Response: 200 OK
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Update Profile
```bash
PUT /api/users/me
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "John Smith"
}

Response: 200 OK
{
  "id": "uuid",
  "email": "user@example.com",
  "name": "John Smith",
  ...
}
```

#### SSE Notifications
```bash
GET /api/notifications/stream
Authorization: Bearer <token>
Accept: text/event-stream

Response: (Server-Sent Events stream)
data: {"id":"...","user_id":"system","message":"Hello!","type":"info",...}

data: {"id":"...","user_id":"system","message":"Another notification","type":"success",...}
```

#### WebSocket Chat
```javascript
// JavaScript client example
const ws = new WebSocket('ws://localhost:8080/api/chat/ws');
ws.onopen = () => {
    ws.send(JSON.stringify({ message: 'Hello, chat!' }));
};
ws.onmessage = (event) => {
    console.log('Received:', event.data);
};
```

---

## 🔐 Authentication & Authorization

### JWT Authentication

All protected endpoints require a JWT token in the `Authorization` header:

```
Authorization: Bearer <your-jwt-token>
```

Get token by:
1. Register a new user: `POST /api/auth/register`
2. Login: `POST /api/auth/login` → returns `{ "token": "..." }`
3. Use token in subsequent requests

### Role-Based Authorization

- **User role**: Can access own profile, notifications, chat
- **Admin role**: Additionally can list all users, ban users, promote users, broadcast notifications

To promote a user to admin, use the database directly (for initial admin setup):
```sql
UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
```

---

## 🐳 Docker Deployment

### Using Docker Compose

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Manual Docker Build

```bash
# Build image
docker build -t production-boilerplate .

# Run container
docker run -p 8080:8080 \
  -e JWT_SECRET=your-secret \
  -v $(pwd)/data:/data \
  production-boilerplate
```

---

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...
```

### Test Coverage Goals

- **Domain models**: >90% (business logic)
- **Services**: >85%
- **Repositories**: >80%
- **Handlers**: >75%

---

## 🏛️ Architecture Patterns

### 1. Rich Domain Models

Business logic lives **inside** entities, not in service layers:

```go
// ✅ GOOD: Business logic in entity
func (u *User) ChangePassword(oldPassword, newPassword string) error {
    if !u.Password.Check(oldPassword) {
        return errors.New("invalid old password")
    }
    // ... validation and update
}

// ❌ BAD: Anemic model (dumb data bag)
type User struct {
    ID       string
    Password string
}
func (s *UserService) ChangePassword(user *User, old, new string) error {
    // Business logic in service (wrong!)
}
```

### 2. Value Objects

Enforce invariants and validation:

```go
type Email struct { value string }

func NewEmail(email string) (Email, error) {
    if !emailRegex.MatchString(email) {
        return Email{}, errors.New("invalid email")
    }
    return Email{value: email}, nil
}
```

### 3. Repository Pattern

Clean data access abstraction:

```go
type Repository interface {
    Get(ctx context.Context, id string) (*User, error)
    Create(ctx context.Context, user *User) error
    // ...
}
```

### 4. Service Layer (Orchestration Only)

Services orchestrate, but delegate business logic to entities:

```go
func (s *serviceImpl) BanUser(ctx context.Context, userID string) error {
    user, err := s.repo.Get(ctx, userID)
    if err != nil {
        return err
    }

    // Delegate to entity (business logic lives there!)
    if err := user.Ban(); err != nil {
        return err
    }

    return s.repo.Update(ctx, user)
}
```

---

## 📝 Configuration

Environment variables (`.env` file):

```bash
# Server
PORT=8080
ENV=development

# Database
DATABASE_DSN=./data/app.db

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRATION=24h

# CORS
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
```

---

## 📊 Database Migrations

Migrations are plain SQL files in `migrations/` directory:

```
migrations/
├── 000001_create_users.up.sql      # Create users table
├── 000001_create_users.down.sql    # Drop users table
├── 000002_create_messages.up.sql   # Create messages table
└── 000002_create_messages.down.sql # Drop messages table
```

Apply migrations:
```bash
make migrate-up    # Apply all migrations
make migrate-down  # Rollback all migrations
```

---

## 🛠️ Development Workflow

### 1. Add a New Feature

```bash
# 1. Create feature in appropriate bounded context
mkdir internal/newfeature

# 2. Implement Rich Domain Model
# internal/newfeature/entity.go

# 3. Implement Repository
# internal/newfeature/repository.go

# 4. Implement Service (orchestration only!)
# internal/newfeature/service.go

# 5. Implement API handlers
# internal/newfeature/api.go

# 6. Wire dependencies in main.go
```

### 2. Make Changes

```bash
# Format code
go fmt ./...

# Run linters
golangci-lint run

# Run tests
go test -race ./...

# Build
go build ./cmd/api
```

---

## 🎓 Learning Resources

### Concepts Demonstrated

- **Domain-Driven Design (DDD)**: Rich models, bounded contexts, value objects
- **Clean Architecture**: Layered design with dependency inversion
- **SOLID Principles**: Single Responsibility, Dependency Injection, etc.
- **Repository Pattern**: Data access abstraction
- **JWT Authentication**: Stateless auth with claims
- **Server-Sent Events (SSE)**: Real-time server→client push
- **WebSocket**: Bidirectional real-time communication
- **RFC 9457 Problem Details**: Standard error responses

### Recommended Reading

- [Domain-Driven Design](https://www.amazon.com/Domain-Driven-Design-Tackling-Complexity-Software/dp/0321125215)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [fursy Documentation](https://github.com/coregx/fursy)
- [stream Documentation](https://github.com/coregx/stream)
- [Relica Documentation](https://github.com/coregx/relica)

---

## 📄 License

MIT License - see LICENSE file for details.

---

## 🤝 Contributing

Contributions welcome! This is a reference implementation for learning and demonstration purposes.

---

## 📬 Support

For questions or issues:
- Open an issue in the [fursy repository](https://github.com/coregx/fursy/issues)
- Check the [examples directory](../) for more examples

---

**Built with**:
- [fursy](https://github.com/coregx/fursy) - Fast Universal Routing SYstem
- [stream](https://github.com/coregx/stream) - SSE + WebSocket library
- [Relica](https://github.com/coregx/relica) - Fluent SQL query builder
- Go 1.25+ modern features

**Status**: Production-ready reference implementation ✅
