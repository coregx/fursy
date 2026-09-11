package main

// Publisher is a nested object used to demonstrate $ref composition in the
// generated OpenAPI document (Book.publisher -> #/components/schemas/Publisher).
type Publisher struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Book is the primary response model.
type Book struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Publisher Publisher `json:"publisher"`
	Year      int       `json:"year"`
	ISBN      string    `json:"isbn,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
}

// BookList is a named wrapper: it becomes its own component whose "books" field
// is an array of $ref Book.
type BookList struct {
	Books []Book `json:"books"`
	Total int    `json:"total"`
}

// CreateBookRequest is the POST /books request body. Fields without omitempty
// are emitted as required in the schema.
type CreateBookRequest struct {
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	PublisherID int      `json:"publisher_id"`
	Year        int      `json:"year"`
	ISBN        string   `json:"isbn,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateBookRequest is the PATCH /books/:id body. Pointer fields let the handler
// tell "omitted" from "set to a zero value", so the schema marks none required.
type UpdateBookRequest struct {
	Title       *string  `json:"title,omitempty"`
	Author      *string  `json:"author,omitempty"`
	PublisherID *int     `json:"publisher_id,omitempty"`
	Year        *int     `json:"year,omitempty"`
	ISBN        *string  `json:"isbn,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// SearchBooksRequest is the optional POST /books/search body.
type SearchBooksRequest struct {
	TitleContains string `json:"title_contains,omitempty"`
	Tag           string `json:"tag,omitempty"`
}

// Stats is returned by the grouped admin endpoint.
type Stats struct {
	TotalBooks int `json:"total_books"`
	TotalTags  int `json:"total_tags"`
}
