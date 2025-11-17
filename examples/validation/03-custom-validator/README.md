# 03-custom-validator - Custom Validation Rules

Demonstrates creating and registering custom validation functions for business-specific rules.

## What This Demonstrates

- Registering custom validators with `v.RegisterCustomValidator()`
- Password strength validation
- Phone number format validation
- Business logic validation (company domain)
- Using custom validators in struct tags

## How to Run

```bash
go run .
```

## Custom Validators

### 1. strong_password

Requirements:
- Minimum 8 characters
- At least one uppercase letter
- At least one lowercase letter
- At least one digit
- At least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

### 2. phone

Accepts various phone formats:
- `+1234567890`
- `(123) 456-7890`
- `123-456-7890`

Validates 10-15 digits for international compatibility.

### 3. company_domain

Checks if email ends with allowed company domains:
- `@example.com`
- `@company.org`
- `@enterprise.net`

## Example Requests

### Valid Request

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "username": "johndoe",
    "password": "Secret123!",
    "phone": "+1234567890"
  }'
```

**Response** (201 Created):
```json
{
  "id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "phone": "+1234567890"
}
```

### Invalid - Weak Password

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "username": "johndoe",
    "password": "weak",
    "phone": "+1234567890"
  }'
```

**Response** (422):
```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "Password failed strong_password validation",
  "errors": {
    "password": "Password failed strong_password validation"
  }
}
```

### Invalid - Wrong Domain

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@gmail.com",
    "username": "johndoe",
    "password": "Secret123!",
    "phone": "+1234567890"
  }'
```

**Response** (422):
```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "Email failed company_domain validation",
  "errors": {
    "email": "Email failed company_domain validation"
  }
}
```

## How to Create Custom Validators

```go
// 1. Define validator function
func validateStrongPassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    // ... validation logic
    return isValid
}

// 2. Register in main()
v := validator.New()
v.RegisterCustomValidator("strong_password", validateStrongPassword)
router.SetValidator(v)

// 3. Use in struct tags
type SignupRequest struct {
    Password string `validate:"required,strong_password"`
}
```

## Key Concepts

### validator.FieldLevel API

```go
func validateCustom(fl validator.FieldLevel) bool {
    fl.Field()          // Get field value (reflect.Value)
    fl.FieldName()      // Get field name
    fl.Param()          // Get validation parameter
    fl.Parent()         // Get parent struct
    return true         // Return validation result
}
```

### Common Patterns

**String validation**:
```go
value := fl.Field().String()
```

**Int validation**:
```go
value := fl.Field().Int()
```

**Using regex**:
```go
pattern := regexp.MustCompile(`^[A-Z]`)
return pattern.MatchString(value)
```

## Next Steps

1. Try [04-nested-structs/](../04-nested-structs/) - Nested validation
2. Explore [05-custom-messages/](../05-custom-messages/) - Custom error messages

## License

MIT License - see [LICENSE](../../../LICENSE) for details.
