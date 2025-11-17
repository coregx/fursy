# 05-custom-messages - Custom Error Messages

Demonstrates customizing validation error messages with placeholders.

## What This Demonstrates

- Setting custom error messages via `validator.Options`
- Using placeholders: `{field}`, `{param}`, `{value}`
- User-friendly error messages
- Localization-ready patterns

## How to Run

```bash
go run main.go
```

## Custom Messages Configuration

```go
v := validator.New(&validator.Options{
    CustomMessages: map[string]string{
        "required": "{field} is required and cannot be empty",
        "email":    "Please provide a valid email address for {field}",
        "min":      "{field} must be at least {param} characters long",
        "max":      "{field} must not exceed {param} characters",
        "gte":      "{field} must be {param} or greater",
        "lte":      "{field} must be {param} or less",
    },
})
```

## Placeholders

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `{field}` | Field name | "Email" |
| `{param}` | Validation parameter | "3" in `min=3` |
| `{value}` | Actual field value | "ab" |

## Example Requests

### Invalid Request

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid",
    "username": "ab",
    "age": 15,
    "password": "123"
  }'
```

**Default Error Messages** (without custom messages):
```json
{
  "errors": {
    "email": "Email must be a valid email address",
    "username": "Username must be at least 3 characters long",
    "age": "Age must be greater than or equal to 18",
    "password": "Password must be at least 8 characters long"
  }
}
```

**Custom Error Messages** (with custom messages):
```json
{
  "type": "about:blank",
  "title": "Validation Failed",
  "status": 422,
  "detail": "4 field(s) failed validation",
  "errors": {
    "email": "Please provide a valid email address for Email",
    "username": "Username must be at least 3 characters long",
    "age": "Age must be 18 or greater",
    "password": "Password must be at least 8 characters long"
  }
}
```

## Message Customization Patterns

### User-Friendly Messages

```go
CustomMessages: map[string]string{
    "required": "Oops! You forgot to fill in {field}",
    "email":    "That doesn't look like a valid email address",
    "min":      "Please enter at least {param} characters for {field}",
}
```

### Technical Messages

```go
CustomMessages: map[string]string{
    "required": "Field '{field}' is mandatory",
    "email":    "Invalid email format in field '{field}'",
    "min":      "Field '{field}' length < {param} (min required)",
}
```

### Localization (Spanish)

```go
CustomMessages: map[string]string{
    "required": "{field} es requerido",
    "email":    "Por favor proporcione un correo válido",
    "min":      "{field} debe tener al menos {param} caracteres",
}
```

## Complete Example with All Common Tags

```go
v := validator.New(&validator.Options{
    CustomMessages: map[string]string{
        // Required/Optional
        "required": "{field} is required",
        "omitempty": "{field} is optional",

        // String validations
        "email":     "Please enter a valid email",
        "url":       "Please enter a valid URL",
        "uuid":      "{field} must be a valid UUID",
        "alpha":     "{field} must contain only letters",
        "alphanum":  "{field} must be alphanumeric",
        "numeric":   "{field} must be numeric",

        // Length
        "min":       "{field} must be at least {param} characters",
        "max":       "{field} must not exceed {param} characters",
        "len":       "{field} must be exactly {param} characters",

        // Numbers
        "gt":        "{field} must be greater than {param}",
        "gte":       "{field} must be {param} or greater",
        "lt":        "{field} must be less than {param}",
        "lte":       "{field} must be {param} or less",
        "eq":        "{field} must equal {param}",

        // Strings
        "contains":   "{field} must contain '{param}'",
        "startswith": "{field} must start with '{param}'",
        "endswith":   "{field} must end with '{param}'",
    },
})
```

## Advanced: Multiple Languages

```go
// Define message sets
var (
    englishMessages = map[string]string{
        "required": "{field} is required",
        "email":    "Invalid email address",
    }

    spanishMessages = map[string]string{
        "required": "{field} es requerido",
        "email":    "Correo electrónico inválido",
    }
)

// Select based on request header
func getMessages(lang string) map[string]string {
    switch lang {
    case "es":
        return spanishMessages
    default:
        return englishMessages
    }
}

// In middleware
func LocalizationMiddleware() fursy.HandlerFunc {
    return func(c *fursy.Context) error {
        lang := c.Header("Accept-Language")
        messages := getMessages(lang)

        v := validator.New(&validator.Options{
            CustomMessages: messages,
        })

        // Re-set validator for this request
        // (In production, cache validators per language)

        return c.Next()
    }
}
```

## Key Concepts

### Placeholder Interpolation

The validator plugin automatically replaces placeholders:

```go
// Message template
"min": "{field} must be at least {param} characters"

// Validation rule
`validate:"min=3"`

// Field name
"Username"

// Result
"Username must be at least 3 characters"
```

### Default Messages

If no custom message is provided, default messages are used:

- `required` → "Field is required"
- `email` → "Field must be a valid email address"
- `min` → "Field must be at least X characters long"

## Next Steps

1. Check [06-production/](../06-production/) - Production setup with all features

## License

MIT License - see [LICENSE](../../../LICENSE) for details.
