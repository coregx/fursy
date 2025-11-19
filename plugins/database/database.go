// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package database provides database integration middleware for fursy HTTP router.
//
// This package provides:
//   - DB wrapper for *sql.DB with context support
//   - Middleware to share database connection across handlers
//   - Transaction helpers with auto-commit/rollback
//   - Context integration via c.DB() method
//
// Example:
//
//	import (
//	    "database/sql"
//	    "github.com/coregx/fursy"
//	    "github.com/coregx/fursy/plugins/database"
//	    _ "github.com/lib/pq" // PostgreSQL driver
//	)
//
//	sqlDB, _ := sql.Open("postgres", dsn)
//	db := database.NewDB(sqlDB)
//
//	router := fursy.New()
//	router.Use(database.Middleware(db))
//
//	router.GET("/users/:id", func(c *fursy.Context) error {
//	    db := c.DB()
//	    var user User
//	    err := db.QueryRow(c.Request.Context(),
//	        "SELECT * FROM users WHERE id = $1", c.Param("id")).
//	        Scan(&user.ID, &user.Name)
//	    if err != nil {
//	        return c.Problem(fursy.NotFound("User not found"))
//	    }
//	    return c.JSON(200, user)
//	})
package database

import (
	"context"
	"database/sql"

	"github.com/coregx/fursy"
)

// contextKey is a private type for storing database in context.
type contextKey int

const (
	dbKey contextKey = iota
	txKey
)

// DB wraps a *sql.DB connection with context support.
//
// It provides a thin wrapper around database/sql that integrates
// with fursy's context and middleware system.
type DB struct {
	db *sql.DB
}

// NewDB creates a new DB wrapper around a *sql.DB connection.
//
// Example:
//
//	sqlDB, err := sql.Open("postgres", dsn)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	db := database.NewDB(sqlDB)
func NewDB(db *sql.DB) *DB {
	return &DB{db: db}
}

// Middleware creates a middleware that stores the database in the request context.
//
// This allows handlers to access the database via c.DB() method.
//
// Example:
//
//	db := database.NewDB(sqlDB)
//	router.Use(database.Middleware(db))
//
//	router.GET("/users", func(c *fursy.Context) error {
//	    db := c.DB()
//	    // Use db for queries...
//	    return nil
//	})
func Middleware(db *DB) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		ctx := context.WithValue(c.Request.Context(), dbKey, db)
		c.Request = c.Request.WithContext(ctx)
		return c.Next()
	}
}

// GetDB retrieves the database from the context.
//
// Returns (nil, false) if database middleware is not configured.
//
// Example:
//
//	db, ok := database.GetDB(c)
//	if !ok {
//	    return c.Problem(fursy.InternalServerError("Database not configured"))
//	}
func GetDB(c *fursy.Context) (*DB, bool) {
	db, ok := c.Request.Context().Value(dbKey).(*DB)
	return db, ok
}

// MustGetDB retrieves the database from the context or panics.
//
// This is a convenience helper for handlers where database is required.
// Panics with a descriptive message if database middleware is not configured.
//
// Use this in handlers where database absence indicates a programming error
// (i.e., middleware misconfiguration), not a runtime error.
//
// Example:
//
//	router.GET("/users", func(c *fursy.Context) error {
//	    db := database.MustGetDB(c) // Panic if DB not configured
//	    rows, err := db.Query(c.Request.Context(), "SELECT * FROM users")
//	    // ...
//	    return c.JSON(200, users)
//	})
//
// For production APIs with proper error handling, use GetDBOrError() instead.
func MustGetDB(c *fursy.Context) *DB {
	db, ok := GetDB(c)
	if !ok {
		panic("database: middleware not configured - ensure database.Middleware(db) is used")
	}
	return db
}

// GetDBOrError retrieves the database from the context or returns an RFC 9457 error.
//
// This is a convenience helper that combines GetDB() with error handling.
// Returns InternalServerError (500) if database middleware is not configured.
//
// This is the recommended approach for production APIs where database
// misconfiguration should return a proper error response.
//
// Example:
//
//	router.GET("/users", func(c *fursy.Context) error {
//	    db, err := database.GetDBOrError(c)
//	    if err != nil {
//	        return c.Problem(err.(fursy.Problem))
//	    }
//	    rows, err := db.Query(c.Request.Context(), "SELECT * FROM users")
//	    // ...
//	    return c.JSON(200, users)
//	})
func GetDBOrError(c *fursy.Context) (*DB, error) {
	db, ok := GetDB(c)
	if !ok {
		return nil, fursy.InternalServerError("Database not configured")
	}
	return db, nil
}

// DB returns the underlying *sql.DB connection.
//
// This is useful when you need to access the raw database/sql API.
//
// Example:
//
//	stats := db.DB().Stats()
func (d *DB) DB() *sql.DB {
	return d.db
}

// Ping verifies a connection to the database is still alive.
//
// Example:
//
//	if err := db.Ping(ctx); err != nil {
//	    log.Fatal("Database connection lost:", err)
//	}
func (d *DB) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Close closes the database connection.
//
// It is rare to Close a DB, as the DB handle is meant to be
// long-lived and shared between many goroutines.
func (d *DB) Close() error {
	return d.db.Close()
}

// Exec executes a query without returning rows.
//
// Example:
//
//	result, err := db.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
//	if err != nil {
//	    return err
//	}
//	rowsAffected, _ := result.RowsAffected()
func (d *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows.
//
// Example:
//
//	rows, err := db.Query(ctx, "SELECT id, name FROM users")
//	if err != nil {
//	    return err
//	}
//	defer rows.Close()
//
//	for rows.Next() {
//	    var id int
//	    var name string
//	    rows.Scan(&id, &name)
//	}
func (d *DB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row.
//
// Example:
//
//	var user User
//	err := db.QueryRow(ctx, "SELECT id, name FROM users WHERE id = $1", userID).
//	    Scan(&user.ID, &user.Name)
//	if err == sql.ErrNoRows {
//	    return ErrNotFound
//	}
func (d *DB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}
