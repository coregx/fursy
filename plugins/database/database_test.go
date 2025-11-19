// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package database_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/database"
	_ "modernc.org/sqlite" // Pure Go SQLite driver for testing
)

// setupDB creates an in-memory SQLite database for testing.
func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:") // modernc.org/sqlite uses "sqlite" driver name
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	return db
}

// Test 1: Middleware stores DB in context.
func TestMiddleware(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	router := fursy.New()
	router.Use(database.Middleware(db))

	router.GET("/test", func(c *fursy.Context) error {
		retrievedDB, ok := database.GetDB(c)
		if !ok {
			t.Error("database not found in context")
		}
		if retrievedDB != db {
			t.Error("wrong database retrieved from context")
		}
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// Test 2: GetDB returns false if not in context.
func TestGetDB_NotFound(t *testing.T) {
	router := fursy.New()

	router.GET("/test", func(c *fursy.Context) error {
		_, ok := database.GetDB(c)
		if ok {
			t.Error("expected database not found, but got ok=true")
		}
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}

// Test 3: DB wrapper methods work correctly.
func TestDB_Methods(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	ctx := context.Background()

	// Test Ping.
	if err := db.Ping(ctx); err != nil {
		t.Errorf("Ping() failed: %v", err)
	}

	// Test DB() returns underlying *sql.DB.
	if db.DB() != sqlDB {
		t.Error("DB() returned wrong *sql.DB")
	}

	// Test Exec - create table and insert in same transaction for sqlite.
	_, err := db.Exec(ctx, `
		CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT);
		INSERT INTO test (name) VALUES ('Alice');
	`)
	if err != nil {
		t.Errorf("Exec() failed: %v", err)
	}

	// Test Query.
	rows, err := db.Query(ctx, "SELECT id, name FROM test")
	if err != nil {
		t.Errorf("Query() failed: %v", err)
	}
	if rows != nil {
		defer rows.Close()

		if !rows.Next() {
			t.Error("expected at least one row")
		}
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			t.Errorf("Scan() failed: %v", err)
		}
		if name != "Alice" {
			t.Errorf("expected name 'Alice', got %q", name)
		}
	}

	// Note: QueryRow is tested in other tests (TestTx_Commit, integration_test).
	// Skipping here due to modernc.org/sqlite connection isolation quirks in memory mode.
}

// Test 4: Transaction commit.
func TestTx_Commit(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	ctx := context.Background()

	// Create table.
	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	// Start transaction.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Insert within transaction.
	_, err = tx.Exec(ctx, "INSERT INTO test (name) VALUES (?)", "Bob")
	if err != nil {
		t.Fatal(err)
	}

	// Commit.
	if err := tx.Commit(); err != nil {
		t.Errorf("Commit() failed: %v", err)
	}

	// Verify data persisted.
	var count int
	db.QueryRow(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 row after commit, got %d", count)
	}
}

// Test 5: Transaction rollback.
func TestTx_Rollback(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO test (name) VALUES (?)", "Charlie")
	if err != nil {
		t.Fatal(err)
	}

	// Rollback.
	if err := tx.Rollback(); err != nil {
		t.Errorf("Rollback() failed: %v", err)
	}

	// Verify data NOT persisted.
	var count int
	db.QueryRow(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 rows after rollback, got %d", count)
	}
}

// Test 6: WithTx helper commits on success.
func TestWithTx_Success(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	// Use WithTx.
	err = database.WithTx(ctx, db, func(tx *database.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO test (name) VALUES (?)", "Dave")
		return err
	})

	if err != nil {
		t.Fatalf("WithTx() failed: %v", err)
	}

	// Verify committed.
	var count int
	db.QueryRow(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 row (commit), got %d", count)
	}
}

// Test 7: WithTx helper rolls back on error.
func TestWithTx_Error(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	// Use WithTx with error.
	err = database.WithTx(ctx, db, func(tx *database.Tx) error {
		_, err := tx.Exec(ctx, "INSERT INTO test (name) VALUES (?)", "Eve")
		if err != nil {
			return err
		}
		return sql.ErrNoRows // Simulate error.
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Verify rolled back.
	var count int
	db.QueryRow(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 rows (rollback), got %d", count)
	}
}

// Test 8: TxMiddleware auto-commits.
func TestTxMiddleware_Commit(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	router := fursy.New()
	router.Use(database.Middleware(db))
	router.Use(database.TxMiddleware(db))

	router.POST("/insert", func(c *fursy.Context) error {
		tx, ok := database.GetTx(c)
		if !ok {
			return c.Problem(fursy.InternalServerError("Transaction not available"))
		}
		_, err := tx.Exec(c.Request.Context(), "INSERT INTO test (name) VALUES (?)", "Frank")
		if err != nil {
			return err
		}
		return c.JSON(200, map[string]string{"status": "inserted"})
	})

	req := httptest.NewRequest("POST", "/insert", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Verify committed.
	var count int
	db.QueryRow(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 row (commit), got %d", count)
	}
}

// Test 9: TxMiddleware auto-rolls back on error.
func TestTxMiddleware_Rollback(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	ctx := context.Background()

	_, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatal(err)
	}

	router := fursy.New()
	router.Use(database.Middleware(db))
	router.Use(database.TxMiddleware(db))

	router.POST("/insert", func(c *fursy.Context) error {
		tx, ok := database.GetTx(c)
		if !ok {
			return c.Problem(fursy.InternalServerError("Transaction not available"))
		}
		_, err := tx.Exec(c.Request.Context(), "INSERT INTO test (name) VALUES (?)", "Grace")
		if err != nil {
			return err
		}
		// Simulate error - use explicit error to trigger rollback.
		c.Problem(fursy.BadRequest("Intentional error")) // Write response
		return fursy.ErrBadRequest                       // Return error to trigger rollback
	})

	req := httptest.NewRequest("POST", "/insert", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Verify rolled back.
	var count int
	db.QueryRow(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 rows (rollback), got %d", count)
	}
}
