# 02-rest-api-crud - Complete REST API with Validation

A complete REST API demonstrating CRUD operations with automatic validation on create and update operations.

## What This Demonstrates

- Complete CRUD operations (Create, Read, Update, Delete)
- Validation on POST and PUT requests
- Type-safe handlers with `Box[Req, Res]`
- In-memory database with thread-safe operations
- RESTful API patterns and status codes
- Error handling with RFC 9457 Problem Details
- Partial updates with optional validation

## How to Run

```bash
go run .
```

Server starts on `http://localhost:8080`.

## API Endpoints

| Method | Endpoint | Description | Validation |
|--------|----------|-------------|------------|
| POST   | `/users` | Create user | ✅ Required |
| GET    | `/users` | List all users | ❌ None |
| GET    | `/users/:id` | Get user by ID | ❌ None |
| PUT    | `/users/:id` | Update user | ✅ Optional fields |
| DELETE | `/users/:id` | Delete user | ❌ None |

## Example Requests

### 1. Create User (POST /users)

**Valid Request**:

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "username": "johndoe",
    "full_name": "John Doe",
    "age": 30,
    "password": "secret123"
  }'
```

**Response** (201 Created):

```json
{
  "id": 1,
  "email": "john@example.com",
  "username": "johndoe",
  "full_name": "John Doe",
  "age": 30
}
```

**Headers**:
```
Location: /users/1
```

**Invalid Request** (validation errors):

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "username": "ab",
    "full_name": "J",
    "age": 15,
    "password": "123"
  }'
```

**Response** (422 Unprocessable Entity):

```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "5 field(s) failed validation",
  "errors": {
    "email": "Email must be a valid email address",
    "username": "Username must be at least 3 characters long",
    "full_name": "FullName must be at least 2 characters long",
    "age": "Age must be greater than or equal to 18",
    "password": "Password must be at least 8 characters long"
  }
}
```

### 2. List All Users (GET /users)

```bash
curl http://localhost:8080/users
```

**Response** (200 OK):

```json
{
  "users": [
    {
      "id": 1,
      "email": "john@example.com",
      "username": "johndoe",
      "full_name": "John Doe",
      "age": 30
    }
  ],
  "total": 1
}
```

### 3. Get User by ID (GET /users/:id)

```bash
curl http://localhost:8080/users/1
```

**Response** (200 OK):

```json
{
  "id": 1,
  "email": "john@example.com",
  "username": "johndoe",
  "full_name": "John Doe",
  "age": 30
}
```

**Not Found**:

```bash
curl http://localhost:8080/users/999
```

**Response** (404 Not Found):

```json
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "User not found"
}
```

### 4. Update User (PUT /users/:id)

**Partial Update** (only email and age):

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newemail@example.com",
    "age": 31
  }'
```

**Response** (200 OK):

```json
{
  "id": 1,
  "email": "newemail@example.com",
  "username": "johndoe",
  "full_name": "John Doe",
  "age": 31
}
```

**Invalid Update**:

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "age": 17
  }'
```

**Response** (422 Unprocessable Entity):

```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "2 field(s) failed validation",
  "errors": {
    "email": "Email must be a valid email address",
    "age": "Age must be greater than or equal to 18"
  }
}
```

### 5. Delete User (DELETE /users/:id)

```bash
curl -X DELETE http://localhost:8080/users/1
```

**Response** (204 No Content):

No body returned, just HTTP 204 status.

## File Structure

```
02-rest-api-crud/
├── main.go        # Application entry point and route setup (~40 lines)
├── models.go      # Request/Response types with validation (~50 lines)
├── database.go    # In-memory database implementation (~100 lines)
├── handlers.go    # HTTP handlers for CRUD operations (~150 lines)
└── README.md      # This file
```

**Total Lines of Code**: ~340 LOC

## Key Concepts

### 1. Validation on Create vs Update

**Create (POST)** - All fields required:

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    FullName string `json:"full_name" validate:"required,min=2,max=100"`
    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
    Password string `json:"password" validate:"required,min=8,max=72"`
}
```

**Update (PUT)** - All fields optional:

```go
type UpdateUserRequest struct {
    Email    string `json:"email,omitempty" validate:"omitempty,email"`
    Username string `json:"username,omitempty" validate:"omitempty,min=3,max=50,alphanum"`
    FullName string `json:"full_name,omitempty" validate:"omitempty,min=2,max=100"`
    Age      int    `json:"age,omitempty" validate:"omitempty,gte=18,lte=120"`
}
```

Use `omitempty` in JSON tag and validation tag for optional fields!

### 2. RESTful Status Codes

| Operation | Success Status | Error Statuses |
|-----------|----------------|----------------|
| POST /users | 201 Created | 422 Validation, 409 Conflict |
| GET /users | 200 OK | - |
| GET /users/:id | 200 OK | 404 Not Found, 400 Bad Request |
| PUT /users/:id | 200 OK | 422 Validation, 404 Not Found |
| DELETE /users/:id | 204 No Content | 404 Not Found |

### 3. Thread-Safe Database

The in-memory database uses `sync.RWMutex` for thread safety:

```go
type Database struct {
    users  map[int]*User
    nextID int
    mu     sync.RWMutex  // Protects concurrent access
}

func (db *Database) Create(req *CreateUserRequest) (*User, error) {
    db.mu.Lock()         // Write lock
    defer db.mu.Unlock()
    // ... create user
}

func (db *Database) GetByID(id int) (*User, error) {
    db.mu.RLock()        // Read lock (allows concurrent reads)
    defer db.mu.RUnlock()
    // ... retrieve user
}
```

### 4. Type-Safe Routing

Each endpoint uses type-safe handlers:

```go
// POST /users - expects CreateUserRequest, returns UserResponse
router.POST[CreateUserRequest, UserResponse]("/users", h.CreateUser)

// GET /users - no request body (Empty), returns UserListResponse
router.GET[fursy.Empty, UserListResponse]("/users", h.ListUsers)

// GET /users/:id - no request body, returns single UserResponse
router.GET[fursy.Empty, UserResponse]("/users/:id", h.GetUser)
```

### 5. Error Handling Patterns

**Validation Errors** (automatic):

```go
if err := c.Bind(); err != nil {
    return err  // Returns 422 with validation errors
}
```

**Business Logic Errors**:

```go
user, err := h.db.Create(c.ReqBody)
if err != nil {
    if errors.Is(err, ErrUserExists) {
        return c.Problem(fursy.Conflict("Username already exists"))
    }
    return c.Problem(fursy.InternalServerError("Failed to create user"))
}
```

## Testing the API

### Test Script

```bash
#!/bin/bash

# Create user
echo "Creating user..."
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","full_name":"Test User","age":25,"password":"secret123"}'

echo -e "\n\nListing users..."
curl http://localhost:8080/users

echo -e "\n\nGetting user 1..."
curl http://localhost:8080/users/1

echo -e "\n\nUpdating user 1..."
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"age":26}'

echo -e "\n\nDeleting user 1..."
curl -X DELETE http://localhost:8080/users/1

echo -e "\n\nListing users after delete..."
curl http://localhost:8080/users
```

## Validation Rules

### CreateUserRequest

| Field | Rules | Error Examples |
|-------|-------|----------------|
| email | required, email | "invalid" → Email must be valid |
| username | required, min=3, max=50, alphanum | "ab" → min 3 chars |
| full_name | required, min=2, max=100 | "A" → min 2 chars |
| age | required, gte=18, lte=120 | 17 → must be ≥18 |
| password | required, min=8, max=72 | "123" → min 8 chars |

### UpdateUserRequest

All fields optional but validated if provided:

| Field | Rules | Error Examples |
|-------|-------|----------------|
| email | omitempty, email | "invalid" → Email must be valid |
| username | omitempty, min=3, max=50, alphanum | "ab" → min 3 chars |
| full_name | omitempty, min=2, max=100 | "A" → min 2 chars |
| age | omitempty, gte=18, lte=120 | 17 → must be ≥18 |

## Next Steps

After mastering this example:

1. Try [03-custom-validator/](../03-custom-validator/) - Custom validation rules
2. Explore [04-nested-structs/](../04-nested-structs/) - Nested validation
3. Check [06-production/](../06-production/) - Production setup with auth

## Common Patterns

### Handling URL Parameters

```go
idStr := c.Param("id")
id, err := strconv.Atoi(idStr)
if err != nil {
    return c.Problem(fursy.BadRequest("Invalid user ID"))
}
```

### Using fursy.Empty for No Request Body

```go
// GET endpoints that don't need request body
router.GET[fursy.Empty, UserResponse]("/users/:id", h.GetUser)
```

### Location Header on Create

```go
return c.Created("/users/"+strconv.Itoa(user.ID), resp)
```

## License

MIT License - see [LICENSE](../../../LICENSE) for details.
