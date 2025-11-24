// Package database provides database connection and simple query execution.
package database

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite" // SQLite driver
)

// DB wraps database/sql with helper methods.
type DB struct {
	*sql.DB
}

// Open opens a database connection.
func Open(dsn string) (*DB, error) {
	// Open SQLite database
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	// Test connection
	if err := db.PingContext(context.Background()); err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}
