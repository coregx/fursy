# 04-nested-structs - Nested Struct Validation

Demonstrates validation of deeply nested structures and slices.

## What This Demonstrates

- Nested struct validation (`Address` within `User`)
- Slice validation with `dive` tag
- Validating each element in an array
- Nested error response format

## How to Run

```bash
go run main.go
```

## Validation Rules

### User
- `name`: required, 2-100 characters
- `email`: required, valid email
- `address`: required nested struct
- `tags`: 1-5 tags, each 2-20 characters (`dive` validates each tag)

### Address (nested)
- `street`: required, min 5 characters
- `city`: required, min 2 characters
- `state`: required, exactly 2 characters (e.g., "CA")
- `zip_code`: required, exactly 5 numeric digits
- `country`: required, exactly 2 characters (e.g., "US")

## Example Requests

### Valid Request

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "address": {
      "street": "123 Main Street",
      "city": "San Francisco",
      "state": "CA",
      "zip_code": "94102",
      "country": "US"
    },
    "tags": ["developer", "golang", "backend"]
  }'
```

**Response** (201):
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "address": {
    "street": "123 Main Street",
    "city": "San Francisco",
    "state": "CA",
    "zip_code": "94102",
    "country": "US"
  },
  "tags": ["developer", "golang", "backend"]
}
```

### Invalid - Missing Nested Fields

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "address": {
      "street": "123",
      "city": "SF",
      "state": "California",
      "zip_code": "941",
      "country": "USA"
    },
    "tags": ["x", "toolongtagname123456789"]
  }'
```

**Response** (422):
```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "6 field(s) failed validation",
  "errors": {
    "address.street": "Street must be at least 5 characters long",
    "address.state": "State must be 2 characters long",
    "address.zip_code": "ZipCode must be 5 characters long",
    "address.country": "Country must be 2 characters long",
    "tags[0]": "Tags[0] must be at least 2 characters long",
    "tags[1]": "Tags[1] must not exceed 20 characters"
  }
}
```

## Key Concepts

### Nested Struct Validation

```go
type Address struct {
    City string `validate:"required,min=2"`
}

type User struct {
    Address *Address `validate:"required"`  // Validates nested struct
}
```

The validator automatically validates all fields in nested structs!

### Slice Validation with `dive`

```go
type User struct {
    Tags []string `validate:"required,min=1,max=5,dive,min=2,max=20"`
}
```

- `required,min=1,max=5` - validates the slice itself (1-5 elements)
- `dive` - tells validator to validate each element
- `min=2,max=20` - each tag must be 2-20 characters

### Nested Error Format

Errors use dot notation for nested fields:
- `address.city` - nested struct field
- `tags[0]` - slice element at index 0
- `user.address.street` - deeply nested field

## Common Patterns

### Optional Nested Struct

```go
type User struct {
    Address *Address `validate:"omitempty"`  // Optional
}
```

### Slice of Nested Structs

```go
type User struct {
    Addresses []*Address `validate:"dive"`  // Validate each address
}
```

### Deep Nesting

```go
type Company struct {
    Users []User `validate:"dive"`
}

type User struct {
    Address *Address `validate:"required"`
}

type Address struct {
    City string `validate:"required"`
}
```

All levels validated automatically!

## Next Steps

1. Try [05-custom-messages/](../05-custom-messages/) - Custom error messages
2. Check [06-production/](../06-production/) - Production setup

## License

MIT License - see [LICENSE](../../../LICENSE) for details.
