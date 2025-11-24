// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"database/sql"

	"github.com/coregx/fursy"
)

// Tx wraps a *sql.Tx transaction with context support.
//
// Transactions provide ACID guarantees for database operations.
// All operations within a transaction are atomic - they either
// all succeed (commit) or all fail (rollback).
type Tx struct {
	tx *sql.Tx
}

// BeginTx starts a new database transaction.
//
// The provided context is used until the transaction is committed or rolled back.
// If the context is canceled, the sql package will roll back the transaction.
//
// Example:
//
//	tx, err := db.BeginTx(ctx, nil)
//	if err != nil {
//	    return err
//	}
//	defer tx.Rollback() // Rollback if not committed
//
//	// ... perform operations ...
//
//	return tx.Commit()
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx}, nil
}

// Commit commits the transaction.
//
// Returns an error if the transaction has already been committed or rolled back.
func (t *Tx) Commit() error {
	return t.tx.Commit()
}

// Rollback aborts the transaction.
//
// Rollback is safe to call even if the transaction has already been committed.
// In that case, it returns sql.ErrTxDone.
func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

// Exec executes a query without returning rows within the transaction.
//
// Example:
//
//	_, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "Alice")
func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows within the transaction.
//
// Example:
//
//	rows, err := tx.Query(ctx, "SELECT id, name FROM users")
//	if err != nil {
//	    return err
//	}
//	defer rows.Close()
func (t *Tx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns at most one row within the transaction.
//
// Example:
//
//	var name string
//	err := tx.QueryRow(ctx, "SELECT name FROM users WHERE id = $1", userID).Scan(&name)
func (t *Tx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// WithTx executes a function within a database transaction.
//
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
//
// This is a convenience helper that handles transaction lifecycle automatically.
//
// Example:
//
//	err := database.WithTx(ctx, db, func(tx *database.Tx) error {
//	    _, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "Bob")
//	    if err != nil {
//	        return err // Automatic rollback
//	    }
//	    _, err = tx.Exec(ctx, "INSERT INTO audit (action) VALUES ($1)", "user_created")
//	    return err // Automatic commit on nil error
//	})
func WithTx(ctx context.Context, db *DB, fn func(*Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback() // Ignore rollback error, return original error.
		return err
	}

	return tx.Commit()
}

// TxMiddleware creates a middleware that wraps each request in a database transaction.
//
// The transaction is automatically committed if the handler succeeds (returns nil),
// or rolled back if the handler returns an error.
//
// This is useful for endpoints that require transactional guarantees.
//
// Example:
//
//	// Apply to specific routes that need transactions:
//	txGroup := router.Group("/api/v1")
//	txGroup.Use(database.Middleware(db))
//	txGroup.Use(database.TxMiddleware(db))
//
//	txGroup.POST("/users", func(c *fursy.Context) error {
//	    tx, _ := database.GetTx(c)
//	    // Use tx for all database operations
//	    // Auto-commit on success, auto-rollback on error
//	    return nil
//	})
func TxMiddleware(db *DB) fursy.HandlerFunc {
	return func(c *fursy.Context) error {
		tx, err := db.BeginTx(c.Request.Context(), nil)
		if err != nil {
			return err
		}

		// Store transaction in context.
		ctx := context.WithValue(c.Request.Context(), txKey, tx)
		c.Request = c.Request.WithContext(ctx)

		// Execute handler chain.
		err = c.Next()

		// Commit or rollback based on result.
		if err != nil {
			_ = tx.Rollback() // Ignore rollback error, return original error.
			return err
		}

		return tx.Commit()
	}
}

// GetTx retrieves the transaction from the context.
//
// Returns (nil, false) if TxMiddleware is not configured for this request.
//
// Example:
//
//	tx, ok := database.GetTx(c)
//	if !ok {
//	    return c.Problem(fursy.InternalServerError("Transaction not available"))
//	}
//	_, err := tx.Exec(c.Request.Context(), "INSERT INTO ...")
func GetTx(c *fursy.Context) (*Tx, bool) {
	tx, ok := c.Request.Context().Value(txKey).(*Tx)
	return tx, ok
}

// MustGetTx retrieves the transaction from the context or panics.
//
// This is a convenience helper for handlers where transaction is required.
// Panics with a descriptive message if TxMiddleware is not configured.
//
// Use this in handlers where transaction absence indicates a programming error
// (i.e., middleware misconfiguration), not a runtime error.
//
// Example:
//
//	txGroup := router.Group("/api")
//	txGroup.Use(database.TxMiddleware(db))
//
//	txGroup.POST("/transfer", func(c *fursy.Context) error {
//	    tx := database.MustGetTx(c) // Panic if TxMiddleware not configured
//	    _, err := tx.Exec(c.Request.Context(), "UPDATE accounts SET ...")
//	    return err
//	})
//
// For production APIs with proper error handling, use GetTxOrError() instead.
func MustGetTx(c *fursy.Context) *Tx {
	tx, ok := GetTx(c)
	if !ok {
		panic("database: transaction not available - ensure database.TxMiddleware(db) is used")
	}
	return tx
}

// GetTxOrError retrieves the transaction from the context or returns an RFC 9457 error.
//
// This is a convenience helper that combines GetTx() with error handling.
// Returns InternalServerError (500) if TxMiddleware is not configured.
//
// This is the recommended approach for production APIs where transaction
// unavailability should return a proper error response.
//
// Example:
//
//	txGroup.POST("/transfer", func(c *fursy.Context) error {
//	    tx, err := database.GetTxOrError(c)
//	    if err != nil {
//	        return c.Problem(err.(fursy.Problem))
//	    }
//	    _, err = tx.Exec(c.Request.Context(), "UPDATE accounts SET ...")
//	    return err
//	})
func GetTxOrError(c *fursy.Context) (*Tx, error) {
	tx, ok := GetTx(c)
	if !ok {
		return nil, fursy.InternalServerError("Transaction not available")
	}
	return tx, nil
}
