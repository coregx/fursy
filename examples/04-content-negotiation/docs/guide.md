# Content Negotiation in fursy

**A comprehensive guide to HTTP content negotiation following RFC 9110**

---

## What is Content Negotiation?

**Content negotiation** (RFC 9110, Section 12) is the mechanism where the server selects the best representation of a resource based on the client's preferences, expressed through HTTP headers.

### Why It Matters

Different clients prefer different formats:

- **Web browsers**: HTML for rendering
- **API clients**: JSON for easy parsing
- **RSS readers**: XML/Atom feeds
- **AI agents** (Claude, ChatGPT): **Markdown** for optimal comprehension
- **Legacy systems**: Plain text or XML

Instead of creating separate endpoints for each format, content negotiation allows **one endpoint** to serve multiple formats intelligently.

## The Accept Header

Clients specify preferences using the `Accept` header:

```http
Accept: text/markdown, text/html;q=0.9, application/json;q=0.8
```

### Quality Values (q-values)

The `q` parameter (0.0 - 1.0) indicates preference:

- **q=1.0** (default): Highest preference
- **q=0.9**: Slightly lower preference
- **q=0.8**: Even lower preference
- **q=0.0**: Not acceptable

**Example interpretation:**
```
Accept: text/markdown, text/html;q=0.9, application/json;q=0.8
```
- **Most preferred**: text/markdown (q=1.0, implicit)
- **Second**: text/html (q=0.9)
- **Third**: application/json (q=0.8)

## fursy Content Negotiation Methods

### 1. Accepts(mediaType) - Simple Check

Check if client accepts a specific media type:

```go
router.GET("/docs", func(c *fursy.Context) error {
    if c.Accepts(fursy.MIMETextMarkdown) {
        // Client accepts markdown - preferred by AI agents
        return c.Markdown("# Documentation\n\nContent here...")
    }

    // Fallback to JSON
    return c.JSON(200, map[string]string{
        "title": "Documentation",
        "content": "Content here...",
    })
})
```

**Use case**: When you have a **preferred format** and a **fallback**.

### 2. AcceptsAny(...mediaTypes) - Priority Selection

Returns the **best match** based on q-values:

```go
router.GET("/users", func(c *fursy.Context) error {
    users := []User{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
    }

    // Check client preferences in order
    switch c.AcceptsAny(fursy.MIMETextMarkdown, fursy.MIMETextHTML, fursy.MIMEApplicationJSON) {
    case fursy.MIMETextMarkdown:
        // AI-friendly format
        md := "# Users\n\n"
        for _, u := range users {
            md += fmt.Sprintf("- **%s** (ID: %d)\n", u.Name, u.ID)
        }
        return c.Markdown(md)

    case fursy.MIMETextHTML:
        // Browser-friendly format
        html := "<h1>Users</h1><ul>"
        for _, u := range users {
            html += fmt.Sprintf("<li><b>%s</b> (ID: %d)</li>", u.Name, u.ID)
        }
        html += "</ul>"
        c.SetHeader("Content-Type", fursy.MIMETextHTML+"; charset=utf-8")
        return c.String(200, html)

    default:
        // API-friendly format (JSON fallback)
        return c.JSON(200, users)
    }
})
```

**Use case**: When you support **multiple formats** and want to respect client preferences.

### 3. Negotiate(status, data) - Automatic

Automatically selects the best format and serializes data:

```go
router.GET("/api/data", func(c *fursy.Context) error {
    data := map[string]any{
        "version": "1.0",
        "status":  "healthy",
        "uptime":  3600,
    }

    // Automatically responds in JSON, XML, or plain text
    return c.Negotiate(200, data)
})
```

**Use case**: When you want **automatic negotiation** without manual format handling.

**Supported formats** (in order):
1. application/json
2. application/xml, text/xml
3. text/plain

## Q-Value Priority Handling

fursy automatically respects RFC 9110 q-value priorities.

### Example 1: AI Agent Preference

```http
GET /docs HTTP/1.1
Accept: text/markdown;q=1.0, application/json;q=0.5
```

**Result**: fursy returns **text/markdown** (highest q-value).

### Example 2: Browser Preference

```http
GET /docs HTTP/1.1
Accept: text/html, application/json;q=0.9, */*;q=0.8
```

**Result**: fursy returns **text/html** (q=1.0 implicit, highest priority).

### Example 3: API Client

```http
GET /api/users HTTP/1.1
Accept: application/json
```

**Result**: fursy returns **application/json**.

### Example 4: Multiple Preferences

```http
GET /users HTTP/1.1
Accept: application/json;q=0.8, text/markdown;q=0.9, text/html;q=0.7
```

**Priority order**:
1. **text/markdown** (q=0.9) ← Selected
2. application/json (q=0.8)
3. text/html (q=0.7)

**Result**: fursy returns **text/markdown**.

## Why Markdown for AI Agents?

AI language models (Claude, ChatGPT, etc.) **excel at processing markdown**:

- **Better comprehension**: Markdown structure is semantic
- **Code blocks**: Properly formatted code examples
- **Lists and tables**: Easier to parse than HTML
- **Headings**: Clear hierarchy for context understanding

**Example - API documentation for AI:**

```markdown
# Users API

## List Users

**Endpoint**: GET /users

**Response**:
- ID (integer): User identifier
- Name (string): Full name
- Email (string): Contact email

**Example**:
```json
{
  "users": [
    {"id": 1, "name": "Alice", "email": "alice@example.com"}
  ]
}
```

## Browser vs API vs AI Agent

| Client | Accept Header | Response Format | Use Case |
|--------|--------------|-----------------|----------|
| **Browser** | text/html | HTML | Human reading, rendering |
| **API Client** | application/json | JSON | Machine parsing, apps |
| **AI Agent** | text/markdown | Markdown | LLM comprehension, analysis |
| **RSS Reader** | application/xml | XML | Feed aggregation |

## Best Practices

### 1. Always Provide a Fallback

```go
// ✅ GOOD: Fallback to JSON
switch c.AcceptsAny(fursy.MIMETextMarkdown, fursy.MIMEApplicationJSON) {
case fursy.MIMETextMarkdown:
    return c.Markdown(content)
default:
    return c.JSON(200, data)
}

// ❌ BAD: No fallback
if c.Accepts(fursy.MIMETextMarkdown) {
    return c.Markdown(content)
}
// If client doesn't accept markdown, returns nothing!
```

### 2. Use Negotiate() for Simple Cases

```go
// ✅ GOOD: Simple automatic negotiation
return c.Negotiate(200, data)

// ❌ OVERKILL: Manual handling for standard formats
switch c.AcceptsAny(fursy.MIMEApplicationJSON, fursy.MIMEApplicationXML) {
case fursy.MIMEApplicationJSON:
    return c.JSON(200, data)
case fursy.MIMEApplicationXML:
    return c.XML(200, data)
}
// Just use Negotiate() instead!
```

### 3. Set Vary Header for Caching

fursy automatically sets `Vary: Accept` when using `Negotiate()`, but for manual negotiation:

```go
// ✅ GOOD: Set Vary header
c.SetHeader("Vary", "Accept")
if c.Accepts(fursy.MIMETextMarkdown) {
    return c.Markdown(content)
}
```

This tells caches (CDN, browser) to store different versions based on Accept header.

### 4. Document Supported Formats

In your API documentation, clearly list supported formats:

```markdown
## GET /docs

Returns API documentation.

**Supported formats**:
- `text/markdown` - AI-friendly markdown (recommended for LLMs)
- `text/html` - Browser-friendly HTML
- `application/json` - Structured data

**Example**:
```bash
curl -H "Accept: text/markdown" http://localhost:8080/docs
```

## Testing Content Negotiation

### Using cURL

```bash
# Request markdown (AI agents)
curl -H "Accept: text/markdown" http://localhost:8080/docs

# Request HTML (browsers)
curl -H "Accept: text/html" http://localhost:8080/docs

# Request JSON (API clients)
curl -H "Accept: application/json" http://localhost:8080/docs

# Multiple preferences with q-values
curl -H "Accept: text/markdown;q=0.9, application/json;q=0.5" http://localhost:8080/docs

# Wildcard (accept anything)
curl -H "Accept: */*" http://localhost:8080/docs
```

### Testing Q-Value Priority

```bash
# Markdown preferred
curl -H "Accept: application/json;q=0.5, text/markdown;q=1.0" http://localhost:8080/users

# JSON preferred
curl -H "Accept: application/json;q=1.0, text/markdown;q=0.5" http://localhost:8080/users

# HTML preferred
curl -H "Accept: text/html;q=0.9, application/json;q=0.7" http://localhost:8080/users
```

## RFC 9110 References

- **Section 12**: Content Negotiation
- **Section 12.5.1**: Accept header field
- **Section 12.5.1**: Quality values (q-values)
- **Section 15.5.7**: 406 Not Acceptable status code

## Summary

Content negotiation in fursy:

1. **Accepts(mediaType)** - Check if client accepts a specific type
2. **AcceptsAny(...mediaTypes)** - Get best match from multiple types
3. **Negotiate(status, data)** - Automatic format selection
4. **Q-values** - Respect client priority (higher q = higher priority)
5. **Markdown** - Preferred format for AI agents (Claude, ChatGPT)
6. **Fallbacks** - Always provide a default format

Use content negotiation to build **flexible APIs** that serve browsers, API clients, and AI agents from a **single endpoint**.

---

*This guide is optimized for markdown rendering and AI agent comprehension.*
