# Hello World Example

The simplest possible **fursy** application - a single endpoint that returns JSON.

## Features

- ✅ Minimal setup (< 30 lines of code)
- ✅ Single GET endpoint
- ✅ JSON response
- ✅ Structured logging with `log/slog`

## Code Structure

```go
func main() {
    router := fursy.New()

    router.GET("/", func(c *fursy.Context) error {
        return c.OK(map[string]string{
            "message": "Hello, World!",
            "status":  "success",
        })
    })

    http.ListenAndServe(":8080", router)
}
```

**Key Points:**
- `fursy.New()` - Create router
- `router.GET()` - Define route
- `c.OK()` - Return 200 OK with JSON body
- `http.ListenAndServe()` - Standard library HTTP server

## Running

### Option 1: Using `go run`

```bash
cd examples/01-hello-world
go run main.go
```

### Option 2: Build and run

```bash
cd examples/01-hello-world
go build -o hello-world
./hello-world  # Linux/macOS
# or
hello-world.exe  # Windows
```

**Output:**
```
INFO Server starting port=8080
INFO Visit: http://localhost:8080
```

## Testing

### cURL

```bash
# GET request
curl http://localhost:8080/

# With pretty-print
curl -s http://localhost:8080/ | jq

# With headers
curl -i http://localhost:8080/
```

**Expected Response:**
```json
{
  "message": "Hello, World!",
  "status": "success"
}
```

### Browser

Open [http://localhost:8080/](http://localhost:8080/) in your browser.

### HTTPie

```bash
http GET :8080
```

## Next Steps

- See [02-rest-api-crud](../02-rest-api-crud/) for a complete CRUD API
- See [validation examples](../validation/) for request validation
- Check the [main README](../../README.md) for advanced features

## Dependencies

**Zero external dependencies** - only stdlib!

```
github.com/coregx/fursy
```

---

**Learn more:** [fursy Documentation](../../README.md)
