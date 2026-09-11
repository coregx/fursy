package main

import (
	"errors"
	"strconv"

	"github.com/coregx/fursy"
)

// Handlers contains the HTTP handlers for the bookstore API.
type Handlers struct {
	store *BookStore
}

// NewHandlers creates a Handlers bound to store.
func NewHandlers(store *BookStore) *Handlers {
	return &Handlers{store: store}
}

// ListBooks handles GET /books.
func (h *Handlers) ListBooks(c *fursy.Box[fursy.Empty, BookList]) error {
	return c.OK(h.store.List())
}

// CreateBook handles POST /books and returns 201 with a Location header.
func (h *Handlers) CreateBook(c *fursy.Box[CreateBookRequest, Book]) error {
	req := c.ReqBody
	if req.Title == "" {
		return c.Problem(fursy.BadRequest("title is required"))
	}
	if req.Author == "" {
		return c.Problem(fursy.BadRequest("author is required"))
	}
	if req.Year < 1450 || req.Year > 2100 {
		return c.Problem(fursy.BadRequest("year must be between 1450 and 2100"))
	}

	book := h.store.Create(*req)
	return c.Created("/books/"+strconv.Itoa(book.ID), book)
}

// GetBook handles GET /books/:id.
func (h *Handlers) GetBook(c *fursy.Box[fursy.Empty, Book]) error {
	id, ok := bookID(c.Context)
	if !ok {
		return c.Problem(fursy.BadRequest("invalid book id"))
	}

	book, err := h.store.Get(id)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			return c.Problem(fursy.NotFound("Book not found"))
		}
		return c.Problem(fursy.InternalServerError("Failed to get book"))
	}
	return c.OK(book)
}

// UpdateBook handles PATCH /books/:id (partial update).
func (h *Handlers) UpdateBook(c *fursy.Box[UpdateBookRequest, Book]) error {
	id, ok := bookID(c.Context)
	if !ok {
		return c.Problem(fursy.BadRequest("invalid book id"))
	}

	book, err := h.store.Update(id, *c.ReqBody)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			return c.Problem(fursy.NotFound("Book not found"))
		}
		return c.Problem(fursy.InternalServerError("Failed to update book"))
	}
	return c.OK(book)
}

// DeleteBook handles DELETE /books/:id and returns 204.
func (h *Handlers) DeleteBook(c *fursy.Box[fursy.Empty, fursy.Empty]) error {
	id, ok := bookID(c.Context)
	if !ok {
		return c.Problem(fursy.BadRequest("invalid book id"))
	}

	if err := h.store.Delete(id); err != nil {
		if errors.Is(err, ErrBookNotFound) {
			return c.Problem(fursy.NotFound("Book not found"))
		}
		return c.Problem(fursy.InternalServerError("Failed to delete book"))
	}
	return c.NoContentSuccess()
}

// SearchBooks handles POST /books/search.
//
// The request body is marked optional in the generated OpenAPI spec
// (OptionalRequestBody). At runtime fursy still requires a JSON body, so send
// {} to search with no filters.
func (h *Handlers) SearchBooks(c *fursy.Box[SearchBooksRequest, BookList]) error {
	return c.OK(h.store.Search(*c.ReqBody))
}

// Stats handles GET /admin/stats.
func (h *Handlers) Stats(c *fursy.Box[fursy.Empty, Stats]) error {
	return c.OK(h.store.Stats())
}

// bookID parses the :id path parameter.
func bookID(c *fursy.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
