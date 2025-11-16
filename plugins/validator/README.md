# Validator Plugin for fursy

[![Go Reference](https://pkg.go.dev/badge/github.com/coregx/fursy/plugins/validator.svg)](https://pkg.go.dev/github.com/coregx/fursy/plugins/validator)
[![Go Report Card](https://goreportcard.com/badge/github.com/coregx/fursy/plugins/validator)](https://goreportcard.com/report/github.com/coregx/fursy/plugins/validator)
[![Coverage](https://img.shields.io/badge/coverage-94.3%25-brightgreen)](https://github.com/coregx/fursy/plugins/validator)

Production-ready validator plugin for [fursy](https://github.com/coregx/fursy) HTTP router, integrating [go-playground/validator/v10](https://github.com/go-playground/validator) with automatic conversion to RFC 9457 compliant error responses.

## Features

- ✅ **Seamless Integration** - Implements `fursy.Validator` interface
- ✅ **RFC 9457 Compliance** - Automatic conversion to Problem Details format
- ✅ **100+ Validation Tags** - Email, URL, UUID, credit card, and more
- ✅ **Custom Validators** - Register your own validation functions
- ✅ **Custom Error Messages** - Fully customizable error messages with placeholders
- ✅ **Nested Struct Validation** - Validates deeply nested structures
- ✅ **Zero Config** - Works out of the box with sensible defaults
- ✅ **Type Safe** - Full Go generics support with fursy's `Box[Req, Res]`
- ✅ **94.3% Test Coverage** - Production-ready and battle-tested

## Installation

```bash
go get github.com/coregx/fursy/plugins/validator
```

**Requirements**: Go 1.25+

## Quick Start

```go
package main

import (
    "github.com/coregx/fursy"
    "github.com/coregx/fursy/plugins/validator"
)

type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3,max=50"`
    Age      int    `json:"age" validate:"gte=18,lte=120"`
    Password string `json:"password" validate:"required,min=8"`
}

type UserResponse struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

func main() {
    router := fursy.New()

    // Set validator plugin (one line!)
    router.SetValidator(validator.New())

    // Use type-safe handlers - validation is automatic!
    router.POST[CreateUserRequest, UserResponse]("/users", func(c *fursy.Box[CreateUserRequest, UserResponse]) error {
        // c.Bind() automatically validates using struct tags
        if err := c.Bind(); err != nil {
            // Returns RFC 9457 Problem Details with validation errors
            return err
        }

        // ReqBody is validated and type-safe!
        user := createUser(c.ReqBody)
        return c.Created("/users/"+user.ID, user)
    })

    router.Run(":8080")
}
```

## Validation Error Response

When validation fails, the plugin returns RFC 9457 Problem Details:

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

## Common Validation Tags

### String Validations

| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field must be present | `validate:"required"` |
| `email` | Must be valid email | `validate:"email"` |
| `url` | Must be valid URL | `validate:"url"` |
| `uri` | Must be valid URI | `validate:"uri"` |
| `uuid` | Must be valid UUID | `validate:"uuid"` |
| `uuid4` | Must be valid UUID v4 | `validate:"uuid4"` |
| `alpha` | Only alphabetic characters | `validate:"alpha"` |
| `alphanum` | Only alphanumeric | `validate:"alphanum"` |
| `numeric` | Must be numeric | `validate:"numeric"` |
| `hexadecimal` | Must be hexadecimal | `validate:"hexadecimal"` |
| `hexcolor` | Must be hex color | `validate:"hexcolor"` |
| `rgb`, `rgba` | Must be RGB/RGBA color | `validate:"rgb"` |
| `hsl`, `hsla` | Must be HSL/HSLA color | `validate:"hsl"` |
| `isbn` | Must be valid ISBN | `validate:"isbn"` |
| `isbn10` | Must be valid ISBN-10 | `validate:"isbn10"` |
| `isbn13` | Must be valid ISBN-13 | `validate:"isbn13"` |
| `json` | Must be valid JSON | `validate:"json"` |
| `latitude` | Must be valid latitude | `validate:"latitude"` |
| `longitude` | Must be valid longitude | `validate:"longitude"` |
| `ssn` | Must be valid SSN | `validate:"ssn"` |
| `ipv4` | Must be valid IPv4 | `validate:"ipv4"` |
| `ipv6` | Must be valid IPv6 | `validate:"ipv6"` |
| `ip` | Must be valid IP | `validate:"ip"` |
| `mac` | Must be valid MAC address | `validate:"mac"` |

### String Length/Comparison

| Tag | Description | Example |
|-----|-------------|---------|
| `min` | Minimum length | `validate:"min=3"` |
| `max` | Maximum length | `validate:"max=50"` |
| `len` | Exact length | `validate:"len=10"` |
| `contains` | Must contain substring | `validate:"contains=@"` |
| `containsany` | Must contain any of | `validate:"containsany=abc"` |
| `excludes` | Must not contain | `validate:"excludes=admin"` |
| `excludesall` | Must not contain any | `validate:"excludesall=!@#"` |
| `startswith` | Must start with | `validate:"startswith=https://"` |
| `endswith` | Must end with | `validate:"endswith=.com"` |

### Number Validations

| Tag | Description | Example |
|-----|-------------|---------|
| `gt` | Greater than | `validate:"gt=0"` |
| `gte` | Greater than or equal | `validate:"gte=18"` |
| `lt` | Less than | `validate:"lt=100"` |
| `lte` | Less than or equal | `validate:"lte=120"` |
| `eq` | Equal to | `validate:"eq=42"` |
| `ne` | Not equal to | `validate:"ne=0"` |
| `oneof` | One of values | `validate:"oneof=red green blue"` |

### Date/Time

| Tag | Description | Example |
|-----|-------------|---------|
| `datetime` | Valid date/time format | `validate:"datetime=2006-01-02"` |

### Slices/Arrays

All string length tags work on slices for item count:
- `min=2` - At least 2 items
- `max=5` - At most 5 items
- `len=3` - Exactly 3 items

## Advanced Usage

### Custom Error Messages

```go
v := validator.New(&validator.Options{
    CustomMessages: map[string]string{
        "required": "{field} is required",
        "email":    "Please provide a valid email for {field}",
        "min":      "{field} must have at least {param} characters",
    },
})

router.SetValidator(v)
```

**Placeholders**:
- `{field}` - Field name
- `{value}` - Actual value that failed
- `{param}` - Validation parameter (e.g., "3" in `min=3`)

### Custom Validators

```go
v := validator.New()

// Register custom validator
v.RegisterCustomValidator("custom_domain", func(fl validator.FieldLevel) bool {
    domain := fl.Field().String()
    return strings.HasSuffix(domain, "@example.com")
})

router.SetValidator(v)

// Use in struct tags
type SignupRequest struct {
    Email string `validate:"required,custom_domain"`
}
```

### Validation Aliases

Create shorthand tags for complex rules:

```go
v := validator.New()

// Register alias
v.RegisterAlias("password", "required,min=8,max=72,containsany=!@#$%")

router.SetValidator(v)

// Use alias in struct tags
type User struct {
    Password string `validate:"password"`
}
```

### Nested Struct Validation

```go
type Address struct {
    Street  string `validate:"required"`
    City    string `validate:"required"`
    ZipCode string `validate:"required,numeric,len=5"`
}

type User struct {
    Name    string   `validate:"required"`
    Email   string   `validate:"required,email"`
    Address *Address `validate:"required"` // Validates nested struct
}
```

### Using JSON Tag Names in Errors

```go
v := validator.New()

v.RegisterTagNameFunc(func(fld reflect.StructField) string {
    name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
    if name == "-" {
        return ""
    }
    return name
})

router.SetValidator(v)

type User struct {
    Email string `json:"email_address" validate:"required,email"`
}
// Error will use "email_address" instead of "Email"
```

### Validating Single Variables

```go
v := validator.New()

// Validate a single value
email := "user@example.com"
err := v.Var(email, "required,email")
if err != nil {
    // Handle error
}
```

## Integration Patterns

### With Middleware

```go
func ValidateMiddleware() fursy.HandlerFunc {
    return func(c *fursy.Context) error {
        // Validation happens automatically in c.Bind()
        // This middleware can add extra checks
        return c.Next()
    }
}

router.Use(ValidateMiddleware())
```

### Manual Validation

```go
router.POST[CreateUserRequest, UserResponse]("/users", func(c *fursy.Box[CreateUserRequest, UserResponse]) error {
    // Option 1: Automatic validation via Bind
    if err := c.Bind(); err != nil {
        return err
    }

    // Option 2: Manual validation
    req := new(CreateUserRequest)
    if err := c.BindJSON(req); err != nil {
        return err
    }

    // Manually validate
    if err := c.Router().Validator().Validate(req); err != nil {
        return err
    }

    // ... process request
})
```

### Production Setup

```go
func setupValidator() *validator.Validator {
    v := validator.New(&validator.Options{
        CustomMessages: map[string]string{
            "required": "{field} is required",
            "email":    "{field} must be a valid email address",
            "min":      "{field} must be at least {param} characters",
            "max":      "{field} must not exceed {param} characters",
        },
    })

    // Register custom validators
    v.RegisterCustomValidator("strong_password", func(fl validator.FieldLevel) bool {
        pass := fl.Field().String()
        // Check for uppercase, lowercase, digit, special char
        return len(pass) >= 8 &&
            regexp.MustCompile(`[A-Z]`).MatchString(pass) &&
            regexp.MustCompile(`[a-z]`).MatchString(pass) &&
            regexp.MustCompile(`[0-9]`).MatchString(pass) &&
            regexp.MustCompile(`[!@#$%^&*]`).MatchString(pass)
    })

    // Register aliases
    v.RegisterAlias("username", "required,min=3,max=50,alphanum")
    v.RegisterAlias("password", "required,strong_password")

    return v
}

func main() {
    router := fursy.New()
    router.SetValidator(setupValidator())
    // ... setup routes
}
```

## Complete Example

See [examples/validation/](../../examples/validation/) for a complete working example.

## API Reference

### Validator

```go
type Validator struct {
    // contains filtered or unexported fields
}

// New creates a new Validator with optional configuration.
func New(opts ...*Options) *Validator

// Validate validates the given data and returns validation errors.
func (v *Validator) Validate(data any) error

// RegisterCustomValidator registers a custom validation function.
func (v *Validator) RegisterCustomValidator(tag string, fn validator.Func) error

// RegisterTagNameFunc registers a function to get custom field names for errors.
func (v *Validator) RegisterTagNameFunc(fn validator.TagNameFunc)

// RegisterAlias registers an alias for validation tags.
func (v *Validator) RegisterAlias(alias, tags string)

// Struct validates a struct and returns validation errors.
func (v *Validator) Struct(data any) error

// Var validates a single variable value against validation tags.
func (v *Validator) Var(field any, tag string) error
```

### Options

```go
type Options struct {
    // TagName is the struct tag name for validation rules.
    // Default: "validate"
    TagName string

    // CustomMessages provides custom error messages for specific tags.
    // Supports placeholders: {field}, {value}, {param}
    CustomMessages map[string]string
}
```

## Performance

- **94.3% test coverage** - Extensively tested
- **Zero allocations** - Efficient error conversion
- **Lazy initialization** - Validator created only when needed
- **Concurrent safe** - Thread-safe validation

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](../../LICENSE) for details.

## Links

- [fursy Documentation](https://github.com/coregx/fursy)
- [go-playground/validator](https://github.com/go-playground/validator)
- [RFC 9457 - Problem Details](https://www.rfc-editor.org/rfc/rfc9457.html)
- [Validation Tag Reference](https://github.com/go-playground/validator#baked-in-validations)

## Support

- GitHub Issues: [github.com/coregx/fursy/issues](https://github.com/coregx/fursy/issues)
- Documentation: [pkg.go.dev/github.com/coregx/fursy/plugins/validator](https://pkg.go.dev/github.com/coregx/fursy/plugins/validator)
