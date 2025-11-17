# REST API CRUD Example

Complete REST API demonstrating CRUD operations with the fursy HTTP router.

## Features

- **Type-safe handlers** using `fursy.Box[Req, Res]`
- **Full CRUD operations** for User entity
- **RFC 9457 Problem Details** for error responses
- **In-memory database** (thread-safe with `sync.RWMutex`)
- **Middleware** (Logger, Recovery)
- **Clean architecture** (handlers, models, database separation)

## Running the Example

```bash
# From this directory
go run .

# Or build first
go build -o rest-api-crud
./rest-api-crud
```

The server will start on http://localhost:8080

## API Endpoints

### Create User
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "username": "johndoe",
    "full_name": "John Doe",
    "age": 30
  }'
```

**Response (201 Created):**
```json
{
  "id": 1,
  "email": "john@example.com",
  "username": "johndoe",
  "full_name": "John Doe",
  "age": 30
}
```

### List All Users
```bash
curl http://localhost:8080/users
```

**Response (200 OK):**
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

### Get User by ID
```bash
curl http://localhost:8080/users/1
```

**Response (200 OK):**
```json
{
  "id": 1,
  "email": "john@example.com",
  "username": "johndoe",
  "full_name": "John Doe",
  "age": 30
}
```

### Update User
```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "age": 31
  }'
```

**Response (200 OK):**
```json
{
  "id": 1,
  "email": "john.doe@example.com",
  "username": "johndoe",
  "full_name": "John Doe",
  "age": 31
}
```

### Delete User
```bash
curl -X DELETE http://localhost:8080/users/1
```

**Response (204 No Content):** _(empty body)_

## Error Handling

The API uses RFC 9457 Problem Details for all errors:

### User Not Found (404)
```bash
curl http://localhost:8080/users/999
```

**Response:**
```json
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "User not found",
  "instance": "/users/999"
}
```

### Validation Error (400)
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe"
  }'
```

**Response:**
```json
{
  "type": "about:blank",
  "title": "Bad Request",
  "status": 400,
  "detail": "Email is required",
  "instance": "/users"
}
```

### Duplicate Username (409)
```bash
# Try to create user with existing username
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "another@example.com",
    "username": "johndoe",
    "full_name": "Another User",
    "age": 25
  }'
```

**Response:**
```json
{
  "type": "about:blank",
  "title": "Conflict",
  "status": 409,
  "detail": "Username already exists",
  "instance": "/users"
}
```

## Code Structure

```
02-rest-api-crud/
├── main.go       - Server setup and route registration
├── models.go     - Request/Response structures
├── handlers.go   - CRUD handlers
├── database.go   - In-memory database
├── go.mod        - Module file
└── README.md     - This file
```

## Key Concepts

### Type-Safe Handlers

```go
func (h *Handlers) CreateUser(c *fursy.Box[CreateUserRequest, UserResponse]) error {
    // c.ReqBody is *CreateUserRequest - type-safe!
    req := c.ReqBody

    // Create and return response
    return c.Created("/users/123", UserResponse{...})
}
```

### Automatic Request Binding

The `c.Bind()` method automatically:
- Parses JSON/XML/Form based on `Content-Type`
- Binds to the typed `ReqBody` field
- Returns RFC 9457 error on failure

### Thread-Safe Database

All database operations use `sync.RWMutex`:
- Read operations: `RLock()` / `RUnlock()`
- Write operations: `Lock()` / `Unlock()`

### Partial Updates

The `UpdateUserRequest` uses `omitempty` tags and checks for empty strings:

```go
type UpdateUserRequest struct {
    Email    string `json:"email,omitempty"`
    Username string `json:"username,omitempty"`
    FullName string `json:"full_name,omitempty"`
    Age      int    `json:"age,omitempty"`
}

// Only update provided fields
if req.Email != "" {
    user.Email = req.Email
}
```

## Testing the Full Workflow

```bash
# 1. Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","username":"alice","full_name":"Alice Smith","age":28}'

# 2. Create another user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"bob@example.com","username":"bob","full_name":"Bob Jones","age":35}'

# 3. List all users
curl http://localhost:8080/users

# 4. Get specific user
curl http://localhost:8080/users/1

# 5. Update user
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"age":29}'

# 6. Delete user
curl -X DELETE http://localhost:8080/users/2

# 7. Verify deletion
curl http://localhost:8080/users
```

## Next Steps

- Add request validation using `github.com/coregx/fursy/plugins/validator`
- Add pagination to list endpoint
- Add filtering and sorting
- Add authentication middleware
- Persist to real database (PostgreSQL, MongoDB, etc.)
- Add unit tests for handlers
- Add integration tests

## Related Examples

- **01-hello-world** - Basic fursy setup
- **validation/02-rest-api-crud** - This example with validation plugin
