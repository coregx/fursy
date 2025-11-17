package main

import (
	"errors"
	"sync"
)

var (
	// ErrUserNotFound is returned when user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists is returned when user already exists.
	ErrUserExists = errors.New("user already exists")
)

// Database represents an in-memory user database.
type Database struct {
	users  map[int]*User
	nextID int
	mu     sync.RWMutex
}

// NewDatabase creates a new in-memory database.
func NewDatabase() *Database {
	return &Database{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

// Create creates a new user.
func (db *Database) Create(req *CreateUserRequest) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if username already exists.
	for _, user := range db.users {
		if user.Username == req.Username {
			return nil, ErrUserExists
		}
	}

	// Create user.
	user := &User{
		ID:       db.nextID,
		Email:    req.Email,
		Username: req.Username,
		FullName: req.FullName,
		Age:      req.Age,
	}

	db.users[user.ID] = user
	db.nextID++

	return user, nil
}

// GetByID retrieves a user by ID.
func (db *Database) GetByID(id int) (*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetAll retrieves all users.
func (db *Database) GetAll() []*User {
	db.mu.RLock()
	defer db.mu.RUnlock()

	users := make([]*User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}

	return users
}

// Update updates a user.
func (db *Database) Update(id int, req *UpdateUserRequest) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, exists := db.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Update only provided fields.
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Age > 0 {
		user.Age = req.Age
	}

	return user, nil
}

// Delete deletes a user.
func (db *Database) Delete(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, exists := db.users[id]
	if !exists {
		return ErrUserNotFound
	}

	delete(db.users, id)
	return nil
}
