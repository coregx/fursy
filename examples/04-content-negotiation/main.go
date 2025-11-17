// Package main demonstrates comprehensive content negotiation with the fursy HTTP router.
//
// This example showcases:
//   - Accepts() method for checking specific MIME types
//   - AcceptsAny() method for multi-format support with q-value priority
//   - Markdown() method for AI-friendly responses
//   - Negotiate() method for automatic format selection
//   - RFC 9110 q-value handling and priority
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
)

// User represents a user entity.
type User struct {
	ID    int    `json:"id" xml:"id"`
	Name  string `json:"name" xml:"name"`
	Email string `json:"email" xml:"email"`
}

// sampleUsers contains demo user data.
var sampleUsers = []User{
	{ID: 1, Name: "Alice Johnson", Email: "alice@example.com"},
	{ID: 2, Name: "Bob Smith", Email: "bob@example.com"},
	{ID: 3, Name: "Charlie Brown", Email: "charlie@example.com"},
}

func main() {
	// Create router with middleware.
	router := fursy.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// 1. Documentation endpoint - markdown/JSON support
	// Demonstrates: Accepts() for simple format checking
	router.GET("/docs", handleDocs)

	// 2. User list endpoint - markdown/HTML/JSON support
	// Demonstrates: AcceptsAny() for multiple formats with q-value priority
	router.GET("/api/users", handleUsers)

	// 3. Automatic negotiation - JSON/XML/plain text
	// Demonstrates: Negotiate() for automatic format selection
	router.GET("/api/data", handleData)

	// 4. Health check - simple JSON response
	router.GET("/health", handleHealth)

	// Start server.
	port := ":8080"
	slog.Info("Content Negotiation Example Server", "port", port)
	slog.Info("Try different Accept headers to see content negotiation in action")
	slog.Info("")
	slog.Info("Example requests:")
	slog.Info("  curl -H 'Accept: text/markdown' http://localhost:8080/docs")
	slog.Info("  curl -H 'Accept: text/html' http://localhost:8080/api/users")
	slog.Info("  curl -H 'Accept: application/json' http://localhost:8080/api/data")
	slog.Info("  curl http://localhost:8080/health")
	slog.Info("")

	if err := http.ListenAndServe(port, router); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

// handleDocs serves API documentation in markdown or JSON format.
//
// This endpoint demonstrates the Accepts() method for simple format checking.
// AI agents (Claude, ChatGPT) prefer markdown, while API clients prefer JSON.
//
// Supported formats:
//   - text/markdown (preferred by AI agents)
//   - application/json (fallback for API clients)
func handleDocs(c *fursy.Context) error {
	// Read markdown documentation from file.
	apiMd, err := os.ReadFile("docs/api.md")
	if err != nil {
		slog.Error("Failed to read api.md", "error", err)
		return c.Problem(fursy.InternalServerError("Failed to load documentation"))
	}

	// Check if client accepts markdown.
	if c.Accepts(fursy.MIMETextMarkdown) {
		// Serve markdown - optimal for AI agents.
		return c.Markdown(string(apiMd))
	}

	// Fallback to JSON for API clients.
	return c.JSON(200, map[string]string{
		"title":       "fursy HTTP Router API Reference",
		"version":     "v0.1.0",
		"format":      "markdown",
		"description": "API documentation is available in markdown format. Use Accept: text/markdown header.",
		"hint":        "AI agents like Claude prefer: curl -H 'Accept: text/markdown' http://localhost:8080/docs",
	})
}

// handleUsers serves a user list in markdown, HTML, or JSON format.
//
// This endpoint demonstrates the AcceptsAny() method for supporting multiple
// formats with automatic q-value priority handling per RFC 9110.
//
// Supported formats:
//   - text/markdown (AI-friendly with structured formatting)
//   - text/html (browser-friendly with HTML rendering)
//   - application/json (API-friendly, default fallback)
//
// Q-value examples:
//   - Accept: text/markdown;q=1.0, application/json;q=0.5
//     Result: markdown (higher q-value)
//   - Accept: text/html;q=0.9, application/json;q=0.7
//     Result: HTML (higher q-value)
func handleUsers(c *fursy.Context) error {
	// Determine the best format based on Accept header and q-values.
	switch c.AcceptsAny(fursy.MIMETextMarkdown, fursy.MIMETextHTML, fursy.MIMEApplicationJSON) {
	case fursy.MIMETextMarkdown:
		// AI-friendly markdown format with structured data.
		md := "# Users\n\n"
		md += "Total users: " + fmt.Sprintf("%d", len(sampleUsers)) + "\n\n"
		md += "## User List\n\n"

		for _, u := range sampleUsers {
			md += fmt.Sprintf("### %s\n\n", u.Name)
			md += fmt.Sprintf("- **ID**: %d\n", u.ID)
			md += fmt.Sprintf("- **Email**: %s\n", u.Email)
			md += "\n"
		}

		md += "---\n\n"
		md += "*This response is optimized for AI agents like Claude and ChatGPT.*\n"

		return c.Markdown(md)

	case fursy.MIMETextHTML:
		// Browser-friendly HTML format.
		html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Users - fursy Example</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 40px auto;
            padding: 0 20px;
            line-height: 1.6;
            color: #333;
        }
        h1 {
            color: #2563eb;
            border-bottom: 3px solid #2563eb;
            padding-bottom: 10px;
        }
        .user-card {
            background: #f8fafc;
            border-left: 4px solid #2563eb;
            padding: 15px 20px;
            margin: 15px 0;
            border-radius: 4px;
        }
        .user-card h3 {
            margin: 0 0 10px 0;
            color: #1e40af;
        }
        .user-info {
            color: #64748b;
            font-size: 14px;
        }
        .total {
            background: #eff6ff;
            padding: 10px 15px;
            border-radius: 4px;
            margin: 20px 0;
        }
        footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #e2e8f0;
            color: #64748b;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <h1>Users</h1>
    <div class="total">
        <strong>Total users:</strong> ` + fmt.Sprintf("%d", len(sampleUsers)) + `
    </div>
`

		for _, u := range sampleUsers {
			html += fmt.Sprintf(`
    <div class="user-card">
        <h3>%s</h3>
        <div class="user-info">
            <strong>ID:</strong> %d<br>
            <strong>Email:</strong> <a href="mailto:%s">%s</a>
        </div>
    </div>
`, u.Name, u.ID, u.Email, u.Email)
		}

		html += `
    <footer>
        <p><em>This response is optimized for browser viewing. Try different Accept headers:</em></p>
        <ul>
            <li><code>Accept: text/markdown</code> - AI-friendly format</li>
            <li><code>Accept: application/json</code> - API-friendly format</li>
        </ul>
    </footer>
</body>
</html>`

		// Send HTML response using String method with HTML content type.
		c.SetHeader("Content-Type", fursy.MIMETextHTML+"; charset=utf-8")
		return c.String(200, html)

	default:
		// API-friendly JSON format (default fallback).
		return c.JSON(200, map[string]any{
			"users": sampleUsers,
			"total": len(sampleUsers),
			"meta": map[string]string{
				"format": "json",
				"hint":   "Try Accept: text/markdown for AI-friendly format, or Accept: text/html for browser viewing",
			},
		})
	}
}

// handleData demonstrates automatic content negotiation with Negotiate().
//
// The Negotiate() method automatically selects the best format from:
//   - application/json (default)
//   - application/xml, text/xml
//   - text/plain
//
// This is useful when you want automatic format selection without manual
// format handling. fursy respects RFC 9110 q-values for priority.
func handleData(c *fursy.Context) error {
	data := map[string]any{
		"service": "fursy-content-negotiation",
		"version": "1.0.0",
		"status":  "healthy",
		"features": []string{
			"Accepts() - check specific MIME type",
			"AcceptsAny() - check multiple types with q-value priority",
			"Negotiate() - automatic format selection",
			"Markdown() - AI-friendly responses",
		},
		"stats": map[string]int{
			"uptime_seconds": 3600,
			"total_requests": 1234,
			"active_users":   42,
		},
	}

	// Automatically negotiate and respond in the best format.
	// Supports: JSON, XML, plain text.
	return c.Negotiate(200, data)
}

// handleHealth is a simple health check endpoint.
//
// Returns JSON status information. This demonstrates a simple endpoint
// without content negotiation (JSON only).
func handleHealth(c *fursy.Context) error {
	return c.OK(map[string]any{
		"status":  "healthy",
		"service": "content-negotiation-example",
		"version": "1.0.0",
	})
}
