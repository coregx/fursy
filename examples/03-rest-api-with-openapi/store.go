package main

import (
	"errors"
	"sort"
	"strings"
	"sync"
)

// ErrBookNotFound is returned when a book does not exist.
var ErrBookNotFound = errors.New("book not found")

// publishers is static reference data so the example needs no external database.
var publishers = map[int]Publisher{
	1: {ID: 1, Name: "Analytical Press"},
	2: {ID: 2, Name: "Mind & Machine"},
}

func publisherByID(id int) Publisher {
	if p, ok := publishers[id]; ok {
		return p
	}
	return Publisher{ID: id, Name: "Unknown"}
}

// BookStore is a small in-memory, thread-safe book store.
type BookStore struct {
	mu     sync.RWMutex
	books  map[int]Book
	nextID int
}

// NewBookStore creates a store seeded with two books.
func NewBookStore() *BookStore {
	s := &BookStore{books: make(map[int]Book), nextID: 1}
	s.seed("The Analytical Engine", "Ada Lovelace", 1, 1843, "978-0-00-000001-0", []string{"history", "computing"})
	s.seed("Computing Machinery and Intelligence", "Alan Turing", 2, 1950, "", []string{"computing", "ai"})
	return s
}

func (s *BookStore) seed(title, author string, publisherID, year int, isbn string, tags []string) {
	book := Book{
		ID:        s.nextID,
		Title:     title,
		Author:    author,
		Publisher: publisherByID(publisherID),
		Year:      year,
		ISBN:      isbn,
		Tags:      tags,
	}
	s.books[book.ID] = book
	s.nextID++
}

// List returns all books ordered by ID.
func (s *BookStore) List() BookList {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return newBookList(s.books)
}

// Search returns books matching the (all-optional) filter.
func (s *BookStore) Search(filter SearchBooksRequest) BookList {
	s.mu.RLock()
	defer s.mu.RUnlock()

	matches := make(map[int]Book)
	for id, book := range s.books {
		if filter.TitleContains != "" &&
			!strings.Contains(strings.ToLower(book.Title), strings.ToLower(filter.TitleContains)) {
			continue
		}
		if filter.Tag != "" && !hasTag(book.Tags, filter.Tag) {
			continue
		}
		matches[id] = book
	}
	return newBookList(matches)
}

// Stats returns aggregate counts.
func (s *BookStore) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tags := make(map[string]bool)
	for _, book := range s.books {
		for _, tag := range book.Tags {
			tags[tag] = true
		}
	}
	return Stats{TotalBooks: len(s.books), TotalTags: len(tags)}
}

// Get returns a single book by ID.
func (s *BookStore) Get(id int) (Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book, ok := s.books[id]
	if !ok {
		return Book{}, ErrBookNotFound
	}
	return book, nil
}

// Create inserts a new book and returns it.
func (s *BookStore) Create(req CreateBookRequest) Book {
	s.mu.Lock()
	defer s.mu.Unlock()

	book := Book{
		ID:        s.nextID,
		Title:     req.Title,
		Author:    req.Author,
		Publisher: publisherByID(req.PublisherID),
		Year:      req.Year,
		ISBN:      req.ISBN,
		Tags:      req.Tags,
	}
	s.books[book.ID] = book
	s.nextID++
	return book
}

// Update applies the provided (non-nil) fields and returns the updated book.
func (s *BookStore) Update(id int, req UpdateBookRequest) (Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	book, ok := s.books[id]
	if !ok {
		return Book{}, ErrBookNotFound
	}

	if req.Title != nil {
		book.Title = *req.Title
	}
	if req.Author != nil {
		book.Author = *req.Author
	}
	if req.PublisherID != nil {
		book.Publisher = publisherByID(*req.PublisherID)
	}
	if req.Year != nil {
		book.Year = *req.Year
	}
	if req.ISBN != nil {
		book.ISBN = *req.ISBN
	}
	if req.Tags != nil {
		book.Tags = req.Tags
	}

	s.books[id] = book
	return book, nil
}

// Delete removes a book by ID.
func (s *BookStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.books[id]; !ok {
		return ErrBookNotFound
	}
	delete(s.books, id)
	return nil
}

// newBookList builds a deterministic BookList from a book map.
func newBookList(books map[int]Book) BookList {
	list := make([]Book, 0, len(books))
	for _, book := range books {
		list = append(list, book)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return BookList{Books: list, Total: len(list)}
}

func hasTag(tags []string, want string) bool {
	for _, tag := range tags {
		if tag == want {
			return true
		}
	}
	return false
}
