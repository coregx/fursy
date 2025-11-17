# Validation Examples for fursy

This directory contains comprehensive validation examples demonstrating fursy's type-safe validation system with the [validator plugin](../../plugins/validator/).

## Prerequisites

- Go 1.25 or higher
- fursy HTTP router
- validator plugin

Install dependencies:

```bash
go get github.com/coregx/fursy
go get github.com/coregx/fursy/plugins/validator
```

## Examples Index

### [01-basic/](./01-basic/) - Minimal Validation Example
**Complexity**: Beginner
**Lines of Code**: ~50-60
**Time to Complete**: 5-10 minutes

A minimal example demonstrating basic validation with a single POST endpoint.

**What you'll learn**:
- Setting up validator plugin
- Using validation struct tags
- Automatic validation with `c.Bind()`
- RFC 9457 error responses

**Run**:
```bash
cd 01-basic
go run main.go
```

---

### [02-rest-api-crud/](./02-rest-api-crud/) - Complete CRUD with Validation
**Complexity**: Intermediate
**Lines of Code**: ~200-250
**Time to Complete**: 20-30 minutes

A complete REST API with CRUD operations and validation on create/update.

**What you'll learn**:
- Type-safe handlers with `Box[Req, Res]`
- Multiple request types with validation
- In-memory database operations
- Validation on both POST and PUT
- RESTful API patterns

**Run**:
```bash
cd 02-rest-api-crud
go run .
```

---

### [03-custom-validator/](./03-custom-validator/) - Custom Validation Rules
**Complexity**: Intermediate
**Lines of Code**: ~100-120
**Time to Complete**: 15-20 minutes

Demonstrates creating and registering custom validation functions.

**What you'll learn**:
- Registering custom validators
- Password strength validation
- Phone number format validation
- Business logic validation rules

**Run**:
```bash
cd 03-custom-validator
go run .
```

---

### [04-nested-structs/](./04-nested-structs/) - Nested Struct Validation
**Complexity**: Intermediate
**Lines of Code**: ~80-100
**Time to Complete**: 10-15 minutes

Shows how to validate deeply nested structures.

**What you'll learn**:
- Nested struct validation
- Using `dive` tag for slices
- Nested error response format
- Validating complex data structures

**Run**:
```bash
cd 04-nested-structs
go run main.go
```

---

### [05-custom-messages/](./05-custom-messages/) - Custom Error Messages
**Complexity**: Intermediate
**Lines of Code**: ~80-100
**Time to Complete**: 10-15 minutes

Demonstrates customizing validation error messages.

**What you'll learn**:
- Custom error messages with placeholders
- Message interpolation (`{field}`, `{param}`)
- User-friendly error messages
- Localization patterns

**Run**:
```bash
cd 05-custom-messages
go run main.go
```

---

### [06-production/](./06-production/) - Production-Ready Setup
**Complexity**: Advanced
**Lines of Code**: ~300-400
**Time to Complete**: 30-45 minutes

A production-ready example with JWT auth, validation, logging, and graceful shutdown.

**What you'll learn**:
- JWT authentication integration
- Structured logging with `log/slog`
- Configuration from environment variables
- Graceful shutdown patterns
- Production error handling
- Complete API setup

**Run**:
```bash
cd 06-production
go run .
```

---

## Recommended Learning Path

For best learning experience, follow this order:

1. **Start with 01-basic/** - Understand the fundamentals
2. **Move to 02-rest-api-crud/** - Learn complete CRUD patterns
3. **Try 03-custom-validator/** - Add custom validation logic
4. **Explore 04-nested-structs/** - Handle complex data
5. **Study 05-custom-messages/** - Improve user experience
6. **Finish with 06-production/** - See it all together

## Common Validation Tags

### String Validations
- `required` - Field must be present
- `email` - Must be valid email
- `url` - Must be valid URL
- `min=3` - Minimum length
- `max=50` - Maximum length
- `alpha` - Only letters
- `alphanum` - Letters and numbers

### Number Validations
- `gte=18` - Greater than or equal
- `lte=120` - Less than or equal
- `gt=0` - Greater than
- `lt=100` - Less than

### Format Validations
- `uuid` - Valid UUID
- `ipv4` - Valid IPv4 address
- `json` - Valid JSON string
- `datetime=2006-01-02` - Date format

See [validator plugin README](../../plugins/validator/README.md) for complete list of 100+ validation tags.

## Testing Examples

Each example can be tested with curl:

```bash
# Valid request
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","username":"john","age":25,"password":"secret123"}'

# Invalid request (triggers validation errors)
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"invalid","username":"ab","age":15,"password":"123"}'
```

Expected validation error response (RFC 9457 format):

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

## Key Concepts

### Type-Safe Handlers

fursy uses Go generics for type-safe request/response handling:

```go
// Define request and response types
type CreateUserRequest struct {
    Email string `json:"email" validate:"required,email"`
    Name  string `json:"name" validate:"required,min=3"`
}

type UserResponse struct {
    ID    int    `json:"id"`
    Email string `json:"email"`
}

// Type-safe handler with automatic validation
router.POST[CreateUserRequest, UserResponse]("/users",
    func(c *fursy.Box[CreateUserRequest, UserResponse]) error {
        if err := c.Bind(); err != nil {
            return err // Returns RFC 9457 validation errors
        }

        // ReqBody is validated and type-safe!
        user := createUser(c.ReqBody)
        return c.JSON(200, user)
    },
)
```

### Automatic Validation

Validation happens automatically when you call `c.Bind()`:

1. Request body is parsed (JSON/XML/Form)
2. Data is unmarshaled into `Req` type
3. Validation tags are checked
4. If validation fails, RFC 9457 error is returned
5. If successful, `c.ReqBody` contains validated data

### RFC 9457 Compliance

All validation errors return RFC 9457 Problem Details format:

- `type` - Problem type URI
- `title` - Human-readable summary
- `status` - HTTP status code (422 for validation)
- `detail` - Detailed explanation
- `errors` - Field-level validation errors

## Integration with fursy Features

### Middleware
Validation works seamlessly with middleware:

```go
router.Use(middleware.Logger())
router.Use(middleware.Recovery())

// Validation happens in handler via c.Bind()
router.POST[Req, Res]("/api/users", handler)
```

### OpenAPI Generation
Validation tags are used for OpenAPI schema generation:

```go
type User struct {
    Email string `json:"email" validate:"required,email" doc:"User email address"`
    Age   int    `json:"age" validate:"gte=18,lte=120" doc:"User age (18-120)"`
}
```

### Content Negotiation
Validation errors respect Accept header:

- `Accept: application/json` → JSON response
- `Accept: application/xml` → XML response
- `Accept: text/plain` → Plain text response

## Troubleshooting

### Validation not working?

1. **Check validator is set**:
   ```go
   router.SetValidator(validator.New())
   ```

2. **Ensure c.Bind() is called**:
   ```go
   if err := c.Bind(); err != nil {
       return err
   }
   ```

3. **Verify struct tags are correct**:
   ```go
   type User struct {
       Email string `validate:"required,email"` // Correct
       // Email string `valid:"required,email"` // Wrong tag name!
   }
   ```

### Custom error messages not showing?

Use `validator.Options`:

```go
v := validator.New(&validator.Options{
    CustomMessages: map[string]string{
        "required": "{field} is required",
        "email": "{field} must be valid email",
    },
})
router.SetValidator(v)
```

### Nested validation failing?

Use `dive` tag for slices:

```go
type User struct {
    Tags []string `validate:"required,dive,min=2"` // Validate each tag
}
```

## Additional Resources

- [fursy Documentation](https://github.com/coregx/fursy)
- [Validator Plugin](../../plugins/validator/)
- [go-playground/validator Tags](https://github.com/go-playground/validator#baked-in-validations)
- [RFC 9457 Specification](https://www.rfc-editor.org/rfc/rfc9457.html)

## Contributing

Found an issue or want to add an example? Please open an issue or PR at [github.com/coregx/fursy](https://github.com/coregx/fursy).

## License

MIT License - see [LICENSE](../../LICENSE) for details.
