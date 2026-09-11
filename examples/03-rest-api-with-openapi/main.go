// Package main demonstrates how fursy's type-safe handlers drive OpenAPI 3.1
// generation: request/response schemas, named components with $ref, status
// codes, path parameters, operationIds, tags, groups, and deprecation.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/middleware"
)

func main() {
	router := fursy.New()

	// Use middleware.
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// API metadata for the generated OpenAPI document.
	router.WithInfo(fursy.Info{
		Title:       "Bookstore API",
		Version:     "1.0.0",
		Description: "Example API showcasing fursy's type-safe handlers and generated OpenAPI 3.1.",
	})
	router.WithServer(fursy.Server{
		URL:         "http://localhost:8080",
		Description: "Local development",
	})

	// Create store and handlers.
	store := NewBookStore()
	handlers := NewHandlers(store)

	// Register routes.
	setupRoutes(router, handlers)

	// Serve the generated OpenAPI 3.1 document.
	router.ServeOpenAPI("/openapi.json")

	// Serve Swagger UI at the root.
	router.Handle("GET", "/", func(c *fursy.Context) error {
		html, err := os.ReadFile("index.html")
		if err != nil {
			return c.Problem(fursy.InternalServerError("Failed to load Swagger UI"))
		}
		c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.Response.WriteHeader(http.StatusOK)
		_, err = c.Response.Write(html)
		return err
	})

	slog.Info("Bookstore API listening on http://localhost:8080")
	slog.Info("Swagger UI", "url", "http://localhost:8080/")
	slog.Info("OpenAPI spec", "url", "http://localhost:8080/openapi.json")

	if err := http.ListenAndServe(":8080", router); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func setupRoutes(router *fursy.Router, h *Handlers) {
	books := []string{"books"}

	// Schemas for Book, BookList, CreateBookRequest, etc. are inferred from the
	// Box[Req, Res] type parameters; the named types become components.schemas.
	router.GET("/books", h.ListBooks, &fursy.RouteOptions{
		Summary: "List all books",
		Tags:    books,
	})
	router.POST("/books", h.CreateBook, &fursy.RouteOptions{
		Summary:       "Create a book",
		Tags:          books,
		SuccessStatus: http.StatusCreated, // 201 Created (with a response schema)
	})
	router.GET("/books/:id", h.GetBook, &fursy.RouteOptions{
		Summary: "Get a book by ID",
		Tags:    books, // :id is auto-declared as a required path parameter
	})
	router.PATCH("/books/:id", h.UpdateBook, &fursy.RouteOptions{
		Summary: "Partially update a book",
		Tags:    books,
	})
	router.DELETE("/books/:id", h.DeleteBook, &fursy.RouteOptions{
		Summary:       "Delete a book",
		Tags:          books,
		SuccessStatus: http.StatusNoContent, // 204 No Content (no response body)
	})
	router.POST("/books/search", h.SearchBooks, &fursy.RouteOptions{
		Summary:             "Search books",
		Description:         "All filter fields are optional.",
		Tags:                books,
		OptionalRequestBody: true, // required: false in the spec
	})

	// Grouped routes are documented too (prefixed path + optional metadata).
	admin := router.Group("/admin")
	admin.GET("/stats", h.Stats, &fursy.RouteOptions{
		Summary: "Get store statistics",
		Tags:    []string{"admin"},
	})

	// Deprecated operations are flagged in the spec.
	router.GET("/legacy/books", h.ListBooks, &fursy.RouteOptions{
		Summary:    "List all books (legacy)",
		Tags:       books,
		Deprecated: true,
	})
}
