# 01-basic - Minimal Validation Example

A minimal example demonstrating basic validation with fursy validator plugin.

## What This Demonstrates

- Setting up validator plugin with `router.SetValidator()`
- Using validation struct tags (`required`, `email`, `min`, `max`, `gte`, `lte`)
- Type-safe handlers with `Box[Req, Res]`
- Automatic validation via `c.Bind()`
- RFC 9457 compliant error responses

## How to Run

```bash
go run main.go
```

Server starts on `http://localhost:8080`.

## Example Requests

### Valid Request

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "john_doe",
    "age": 25,
    "password": "secret123"
  }'
```

**Expected Response** (201 Created):

```json
{
  "id": 1,
  "username": "john_doe",
  "email": "user@example.com"
}
```

### Invalid Request - All Fields Fail

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "username": "ab",
    "age": 15,
    "password": "123"
  }'
```

**Expected Response** (422 Unprocessable Entity):

```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "4 field(s) failed validation",
  "errors": {
    "email": "Email must be a valid email address",
    "username": "Username must be at least 3 characters long",
    "age": "Age must be greater than or equal to 18",
    "password": "Password must be at least 8 characters long"
  }
}
```

### Invalid Request - Missing Required Fields

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Expected Response** (422 Unprocessable Entity):

```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "4 field(s) failed validation",
  "errors": {
    "email": "Email is required",
    "username": "Username is required",
    "age": "Age must be greater than or equal to 18",
    "password": "Password is required"
  }
}
```

### Invalid Request - Age Out of Range

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "john",
    "age": 150,
    "password": "secret123"
  }'
```

**Expected Response** (422 Unprocessable Entity):

```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "Age must be less than or equal to 120",
  "errors": {
    "age": "Age must be less than or equal to 120"
  }
}
```

## Key Concepts

### 1. Validation Struct Tags

Define validation rules using struct tags:

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=50"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
    Password string `json:"password" validate:"required,min=8"`
}
```

**Common Tags**:
- `required` - Field must be present
- `email` - Must be valid email format
- `min=3` - Minimum length (for strings) or value (for numbers)
- `max=50` - Maximum length or value
- `gte=18` - Greater than or equal
- `lte=120` - Less than or equal

### 2. Type-Safe Handlers

Use `Box[Req, Res]` for type-safe request/response handling:

```go
router.POST[CreateUserRequest, UserResponse]("/users", createUser)

func createUser(c *fursy.Box[CreateUserRequest, UserResponse]) error {
    // c.ReqBody is type CreateUserRequest
    // Return type UserResponse
}
```

### 3. Automatic Validation

Validation happens when you call `c.Bind()`:

```go
if err := c.Bind(); err != nil {
    return err // Returns RFC 9457 validation errors
}

// ReqBody is validated - safe to use!
req := c.ReqBody
```

### 4. RFC 9457 Error Format

Validation errors automatically return RFC 9457 Problem Details:

```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "N field(s) failed validation",
  "errors": {
    "field_name": "Error message"
  }
}
```

## Code Structure

```
01-basic/
├── main.go        # Main application (~60 lines)
└── README.md      # This file
```

**Lines of Code**: ~60 LOC

## Next Steps

After mastering this example:

1. Try [02-rest-api-crud/](../02-rest-api-crud/) - Complete CRUD operations
2. Explore [03-custom-validator/](../03-custom-validator/) - Custom validation rules
3. Check [06-production/](../06-production/) - Production setup

## Validation Tags Reference

See complete list in [validator plugin README](../../../plugins/validator/README.md#common-validation-tags).

## Troubleshooting

### Validation Not Working?

Make sure you:
1. Set validator: `router.SetValidator(validator.New())`
2. Call `c.Bind()` in handler
3. Use correct tag name: `validate:` (not `valid:`)

### Want Custom Error Messages?

See [05-custom-messages/](../05-custom-messages/) example.

## License

MIT License - see [LICENSE](../../../LICENSE) for details.
