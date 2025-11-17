# Content Negotiation Example

**Comprehensive demonstration of HTTP content negotiation with the fursy router.**

This example showcases RFC 9110 compliant content negotiation, supporting multiple formats (JSON, XML, HTML, Markdown, plain text) from a single endpoint, with automatic q-value priority handling.

## What is Content Negotiation?

**Content negotiation** (RFC 9110, Section 12) is the mechanism where the server selects the best representation of a resource based on the client's preferences, expressed through the HTTP `Accept` header.

### Why It Matters

Different clients prefer different formats:

- **Web browsers** → HTML (for rendering)
- **API clients** → JSON (for parsing)
- **AI agents** (Claude, ChatGPT) → **Markdown** (for optimal comprehension)
- **RSS readers** → XML (for feeds)
- **Legacy systems** → Plain text or XML

Instead of creating separate endpoints for each format, content negotiation allows **one endpoint** to intelligently serve multiple formats.

## Features Demonstrated

This example demonstrates all fursy content negotiation methods:

| Method | Purpose | Use Case |
|--------|---------|----------|
| **`Accepts(mediaType)`** | Check if client accepts a specific type | Simple format checking with fallback |
| **`AcceptsAny(...mediaTypes)`** | Get best match from multiple types | Multi-format support with q-value priority |
| **`Negotiate(status, data)`** | Automatic format selection | Auto-negotiation for standard formats |
| **`Markdown(content)`** | Return markdown response | AI-friendly documentation |

### Q-Value Priority Handling

fursy automatically respects RFC 9110 **quality values** (q-values):

```http
Accept: text/markdown;q=1.0, application/json;q=0.5
```

- `q=1.0` (default): Highest preference
- `q=0.9`: Slightly lower preference
- `q=0.5`: Lower preference
- `q=0.0`: Not acceptable

The server selects the format with the **highest q-value** that it supports.

## Running the Example

```bash
# From this directory
go run .

# Or build first
go build -o content-negotiation
./content-negotiation
```

The server will start on http://localhost:8080

## Endpoints

### 1. GET /docs - Documentation with Markdown/JSON

Demonstrates `Accepts()` method for simple format checking.

**Supported formats:**
- `text/markdown` - AI-friendly markdown (preferred by LLMs)
- `application/json` - API-friendly JSON (fallback)

**Test with cURL:**

```bash
# Request markdown (AI agents like Claude prefer this)
curl -H "Accept: text/markdown" http://localhost:8080/docs

# Request JSON (API clients)
curl -H "Accept: application/json" http://localhost:8080/docs

# No Accept header (defaults to first check - markdown if available)
curl http://localhost:8080/docs
```

**Response (Markdown):**
```markdown
# fursy HTTP Router API Reference

**Version**: v0.1.0
**Package**: github.com/coregx/fursy

## Overview

fursy (Fast Universal Routing SYstem) is a production-ready HTTP router...
```

**Response (JSON):**
```json
{
  "title": "fursy HTTP Router API Reference",
  "version": "v0.1.0",
  "format": "markdown",
  "description": "API documentation is available in markdown format...",
  "hint": "AI agents like Claude prefer: curl -H 'Accept: text/markdown' ..."
}
```

### 2. GET /api/users - User List with Markdown/HTML/JSON

Demonstrates `AcceptsAny()` method for multi-format support with q-value priority.

**Supported formats:**
- `text/markdown` - AI-friendly structured format
- `text/html` - Browser-friendly HTML rendering
- `application/json` - API-friendly JSON (default fallback)

**Test with cURL:**

```bash
# Request markdown (AI agents)
curl -H "Accept: text/markdown" http://localhost:8080/api/users

# Request HTML (browsers)
curl -H "Accept: text/html" http://localhost:8080/api/users

# Request JSON (API clients)
curl -H "Accept: application/json" http://localhost:8080/api/users

# No Accept header (defaults to JSON fallback)
curl http://localhost:8080/api/users
```

**Response (Markdown):**
```markdown
# Users

Total users: 3

## User List

### Alice Johnson

- **ID**: 1
- **Email**: alice@example.com

### Bob Smith

- **ID**: 2
- **Email**: bob@example.com

### Charlie Brown

- **ID**: 3
- **Email**: charlie@example.com

---

*This response is optimized for AI agents like Claude and ChatGPT.*
```

**Response (HTML):**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Users - fursy Example</title>
    <style>
        /* Styled user cards with modern CSS */
    </style>
</head>
<body>
    <h1>Users</h1>
    <div class="total">Total users: 3</div>
    <div class="user-card">
        <h3>Alice Johnson</h3>
        <div class="user-info">
            <strong>ID:</strong> 1<br>
            <strong>Email:</strong> alice@example.com
        </div>
    </div>
    <!-- More user cards... -->
</body>
</html>
```

**Response (JSON):**
```json
{
  "users": [
    {
      "id": 1,
      "name": "Alice Johnson",
      "email": "alice@example.com"
    },
    {
      "id": 2,
      "name": "Bob Smith",
      "email": "bob@example.com"
    },
    {
      "id": 3,
      "name": "Charlie Brown",
      "email": "charlie@example.com"
    }
  ],
  "total": 3,
  "meta": {
    "format": "json",
    "hint": "Try Accept: text/markdown for AI-friendly format..."
  }
}
```

### 3. GET /api/data - Automatic Negotiation (JSON/XML/Plain Text)

Demonstrates `Negotiate()` method for automatic format selection.

**Supported formats** (automatic):
- `application/json` (default)
- `application/xml`, `text/xml`
- `text/plain`

**Test with cURL:**

```bash
# Request JSON (default)
curl http://localhost:8080/api/data

# Request XML
curl -H "Accept: application/xml" http://localhost:8080/api/data

# Request plain text
curl -H "Accept: text/plain" http://localhost:8080/api/data

# Wildcard (server chooses best - JSON)
curl -H "Accept: */*" http://localhost:8080/api/data
```

**Response (JSON):**
```json
{
  "service": "fursy-content-negotiation",
  "version": "1.0.0",
  "status": "healthy",
  "features": [
    "Accepts() - check specific MIME type",
    "AcceptsAny() - check multiple types with q-value priority",
    "Negotiate() - automatic format selection",
    "Markdown() - AI-friendly responses"
  ],
  "stats": {
    "uptime_seconds": 3600,
    "total_requests": 1234,
    "active_users": 42
  }
}
```

**Response (XML):**
```xml
<map>
  <service>fursy-content-negotiation</service>
  <version>1.0.0</version>
  <status>healthy</status>
  <features>Accepts() - check specific MIME type</features>
  <features>AcceptsAny() - check multiple types with q-value priority</features>
  <features>Negotiate() - automatic format selection</features>
  <features>Markdown() - AI-friendly responses</features>
  <stats>
    <uptime_seconds>3600</uptime_seconds>
    <total_requests>1234</total_requests>
    <active_users>42</active_users>
  </stats>
</map>
```

**Response (Plain Text):**
```
service: fursy-content-negotiation
version: 1.0.0
status: healthy
features: [Accepts() - check specific MIME type, AcceptsAny() - ...]
stats: map[active_users:42 total_requests:1234 uptime_seconds:3600]
```

### 4. GET /health - Simple Health Check

Simple JSON-only endpoint without content negotiation.

**Test with cURL:**

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "content-negotiation-example",
  "version": "1.0.0"
}
```

## Q-Value Priority Examples

fursy automatically handles RFC 9110 quality values (q-values) to select the best format.

### Example 1: AI Agent with Markdown Preference

```bash
curl -H "Accept: text/markdown;q=1.0, application/json;q=0.5" \
  http://localhost:8080/api/users
```

**Priority:**
1. text/markdown (q=1.0) ← **Selected**
2. application/json (q=0.5)

**Result:** Returns **markdown** format.

### Example 2: Browser with HTML Preference

```bash
curl -H "Accept: text/html;q=0.9, application/json;q=0.7, */*;q=0.1" \
  http://localhost:8080/api/users
```

**Priority:**
1. text/html (q=0.9) ← **Selected**
2. application/json (q=0.7)
3. */* (q=0.1)

**Result:** Returns **HTML** format.

### Example 3: API Client with JSON Preference

```bash
curl -H "Accept: application/json;q=1.0, text/markdown;q=0.8" \
  http://localhost:8080/api/users
```

**Priority:**
1. application/json (q=1.0) ← **Selected**
2. text/markdown (q=0.8)

**Result:** Returns **JSON** format.

### Example 4: Multiple Equal Preferences (First Match Wins)

```bash
curl -H "Accept: text/markdown, text/html, application/json" \
  http://localhost:8080/api/users
```

**Priority:** All have q=1.0 (implicit), so fursy selects the **first match** from the server's offered list.

**Server offers:** `[text/markdown, text/html, application/json]`

**Result:** Returns **markdown** (first offered that matches).

### Example 5: Q-Value Ties (Specificity Wins)

```bash
curl -H "Accept: text/*;q=0.9, text/markdown;q=0.9, application/json;q=0.7" \
  http://localhost:8080/api/users
```

**Priority:**
1. text/markdown (q=0.9, more specific) ← **Selected**
2. text/* (q=0.9, less specific wildcard)
3. application/json (q=0.7)

**Result:** Returns **markdown** (more specific type wins on tie).

### Example 6: No Acceptable Format

```bash
curl -H "Accept: image/png" http://localhost:8080/api/users
```

**Result:** Returns **default fallback** (JSON) since image/png is not supported.

## Why Markdown for AI Agents?

AI language models (Claude, ChatGPT, GPT-4, etc.) **excel at processing markdown**:

### Better Comprehension
- **Semantic structure**: Headings, lists, tables are clearly defined
- **Code blocks**: Syntax highlighting and proper formatting
- **Emphasis**: Bold, italic, and other markdown features preserve meaning

### Example Comparison

**JSON (Good for machines, harder for AI):**
```json
{
  "user": {
    "name": "Alice",
    "email": "alice@example.com",
    "bio": "Software engineer with 5 years experience in Go."
  }
}
```

**Markdown (Optimal for AI):**
```markdown
## Alice

**Email:** alice@example.com

**Bio:** Software engineer with 5 years experience in Go.
```

The markdown version provides:
- Clear hierarchy (heading)
- Emphasis (bold labels)
- Natural language flow

AI models can better understand context, relationships, and semantics from markdown.

## Code Examples

### Using Accepts() - Simple Format Check

```go
router.GET("/docs", func(c *fursy.Context) error {
    if c.Accepts(fursy.MIMETextMarkdown) {
        // Client accepts markdown - preferred by AI agents
        return c.Markdown(markdownContent)
    }

    // Fallback to JSON
    return c.JSON(200, data)
})
```

**Use case:** When you have a **preferred format** and a **fallback**.

### Using AcceptsAny() - Multi-Format with Priority

```go
router.GET("/users", func(c *fursy.Context) error {
    users := getUserList()

    switch c.AcceptsAny(fursy.MIMETextMarkdown, fursy.MIMETextHTML, fursy.MIMEApplicationJSON) {
    case fursy.MIMETextMarkdown:
        return c.Markdown(renderMarkdown(users))
    case fursy.MIMETextHTML:
        c.SetHeader("Content-Type", fursy.MIMETextHTML+"; charset=utf-8")
        return c.String(200, renderHTML(users))
    default:
        return c.JSON(200, users)
    }
})
```

**Use case:** When you support **multiple formats** and respect client preferences.

### Using Negotiate() - Automatic Selection

```go
router.GET("/api/data", func(c *fursy.Context) error {
    data := getData()

    // Automatically responds in JSON, XML, or plain text
    return c.Negotiate(200, data)
})
```

**Use case:** When you want **automatic negotiation** without manual handling.

## Browser Testing

You can also test in a browser by visiting:

- http://localhost:8080/api/users (browsers send `Accept: text/html` by default)
- http://localhost:8080/docs (will show JSON in browser)

**Browser Accept header (typical):**
```
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
```

Browsers prefer HTML, so `GET /api/users` will render the styled HTML user list.

## Testing with Different Tools

### cURL

```bash
# Explicit Accept header
curl -H "Accept: text/markdown" http://localhost:8080/docs

# Verbose output (see headers)
curl -v -H "Accept: application/json" http://localhost:8080/api/users
```

### HTTPie

```bash
# HTTPie automatically sends Accept: application/json
http GET http://localhost:8080/api/data

# Override Accept header
http GET http://localhost:8080/api/data Accept:application/xml
```

### Postman

1. Create GET request to `http://localhost:8080/api/users`
2. Go to **Headers** tab
3. Add `Accept: text/markdown`
4. Send request and view markdown response

### Browser DevTools

1. Open http://localhost:8080/api/users in browser
2. Open DevTools (F12) → Network tab
3. Refresh page
4. Click on the request
5. View **Request Headers** → See `Accept: text/html, ...`
6. View **Response** → See HTML rendered

## File Structure

```
04-content-negotiation/
├── main.go        - Server with 4 endpoints demonstrating negotiation
├── go.mod         - Module file with replace directive
├── README.md      - This file (comprehensive documentation)
└── docs/          - Sample markdown documentation files
    ├── api.md     - API reference in markdown
    └── guide.md   - Content negotiation guide in markdown
```

## Key Concepts

### 1. Accept Header

The client specifies preferences:

```http
Accept: text/markdown;q=1.0, application/json;q=0.5
```

- `text/markdown` with quality value 1.0 (highest)
- `application/json` with quality value 0.5 (lower)

### 2. Quality Values (q-values)

Range: 0.0 - 1.0 (default: 1.0)

- Higher q-value = higher preference
- Server selects format with highest q-value it supports
- On tie, more specific type wins (e.g., `text/markdown` > `text/*`)

### 3. Vary Header

fursy automatically sets `Vary: Accept` when using `Negotiate()`:

```http
Vary: Accept
```

This tells caches (CDN, browser) to store different versions based on Accept header.

### 4. Fallback Strategy

Always provide a fallback:

```go
// ✅ GOOD
switch c.AcceptsAny(fursy.MIMETextMarkdown, fursy.MIMEApplicationJSON) {
case fursy.MIMETextMarkdown:
    return c.Markdown(content)
default:
    return c.JSON(200, data)  // Fallback
}

// ❌ BAD: No fallback!
if c.Accepts(fursy.MIMETextMarkdown) {
    return c.Markdown(content)
}
// Returns nothing if client doesn't accept markdown!
```

## RFC 9110 References

This example follows **RFC 9110** (HTTP Semantics):

- **Section 12**: Content Negotiation
- **Section 12.5.1**: Accept header field
- **Section 12.4.2**: Quality values (q-values)
- **Section 15.5.7**: 406 Not Acceptable status code

## Benefits of Content Negotiation

### 1. Single Endpoint, Multiple Formats

Instead of:
- `/api/users.json`
- `/api/users.xml`
- `/api/users.html`

Use one endpoint:
- `/api/users` (responds in JSON, XML, HTML, Markdown based on Accept header)

### 2. Client Flexibility

Different clients can request the format they prefer:

- **Mobile app** → JSON (compact, easy to parse)
- **Web browser** → HTML (rendered view)
- **AI agent** → Markdown (optimal comprehension)
- **RSS reader** → XML (feed format)

### 3. Future-Proof

Add new formats without changing endpoint URLs:

```go
// Easy to add new format later
case fursy.MIMEApplicationYAML:
    return c.YAML(200, data)
```

### 4. Standards Compliant

Follows HTTP specifications (RFC 9110), ensuring compatibility with:
- Browsers
- API clients
- Proxies and CDNs
- HTTP caches

## Best Practices

### 1. Always Provide Fallback

```go
default:
    return c.JSON(200, data)  // Safe fallback
```

### 2. Use Negotiate() for Standard Formats

```go
// ✅ Simple and automatic
return c.Negotiate(200, data)

// ❌ Overkill for standard formats
switch c.AcceptsAny(fursy.MIMEApplicationJSON, fursy.MIMEApplicationXML) {
    // Just use Negotiate()!
}
```

### 3. Document Supported Formats

In your API docs, list supported formats:

```markdown
## GET /api/users

**Supported formats:**
- `text/markdown` - AI-friendly
- `text/html` - Browser-friendly
- `application/json` - API-friendly (default)
```

### 4. Test with Multiple Accept Headers

Always test your endpoints with different Accept headers to ensure proper negotiation.

## Next Steps

- Add more response formats (YAML, CSV, etc.)
- Implement content negotiation for request bodies (Content-Type negotiation)
- Add compression negotiation (Accept-Encoding)
- Add language negotiation (Accept-Language)
- Implement 406 Not Acceptable for strict negotiation
- Add caching with proper Vary headers

## Related Examples

- **01-hello-world** - Basic fursy setup
- **02-rest-api-crud** - CRUD operations with type-safe handlers
- **validation/** - Request validation examples

## Further Reading

- **docs/api.md** - Complete API reference in markdown
- **docs/guide.md** - Detailed content negotiation guide
- [RFC 9110 - HTTP Semantics](https://datatracker.ietf.org/doc/html/rfc9110)
- [MDN - Content Negotiation](https://developer.mozilla.org/en-US/docs/Web/HTTP/Content_negotiation)

---

**Note:** This example is designed to be educational and comprehensive, demonstrating all fursy content negotiation capabilities with practical, real-world examples.
