// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package database_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/database"
)

// Test 10: MustGetDB returns DB when configured.
func TestMustGetDB_Success(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	router := fursy.New()
	router.Use(database.Middleware(db))

	router.GET("/test", func(c *fursy.Context) error {
		retrievedDB := database.MustGetDB(c) // Should not panic.
		if retrievedDB != db {
			t.Error("MustGetDB returned wrong database")
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

// Test 11: MustGetDB panics when not configured.
func TestMustGetDB_Panic(t *testing.T) {
	router := fursy.New()
	// NO database middleware!

	router.GET("/test", func(c *fursy.Context) error {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGetDB should panic when DB not configured")
			}
		}()
		database.MustGetDB(c) // Should panic.
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}

// Test 12: GetDBOrError returns DB when configured.
func TestGetDBOrError_Success(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	router := fursy.New()
	router.Use(database.Middleware(db))

	router.GET("/test", func(c *fursy.Context) error {
		retrievedDB, err := database.GetDBOrError(c)
		if err != nil {
			return err
		}
		if retrievedDB != db {
			t.Error("GetDBOrError returned wrong database")
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

// Test 13: GetDBOrError returns error when not configured.
func TestGetDBOrError_Error(t *testing.T) {
	router := fursy.New()
	// NO database middleware!

	router.GET("/test", func(c *fursy.Context) error {
		db, err := database.GetDBOrError(c)
		if err == nil {
			t.Error("GetDBOrError should return error when DB not configured")
		}
		if db != nil {
			t.Error("GetDBOrError should return nil DB when not configured")
		}
		// Convert Problem error to response.
		var prob fursy.Problem
		if errors.As(err, &prob) {
			return c.Problem(prob)
		}
		return err
	})

	req := httptest.NewRequest("GET", "/test", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected 500 Internal Server Error, got %d", w.Code)
	}
}

// Test 14: MustGetTx returns Tx when configured.
func TestMustGetTx_Success(t *testing.T) {
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
		tx := database.MustGetTx(c) // Should not panic.
		_, err := tx.Exec(c.Request.Context(), "INSERT INTO test (name) VALUES (?)", "Helen")
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
		t.Errorf("expected 1 row, got %d", count)
	}
}

// Test 15: MustGetTx panics when not configured.
func TestMustGetTx_Panic(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	router := fursy.New()
	router.Use(database.Middleware(db))
	// NO TxMiddleware!

	router.POST("/insert", func(c *fursy.Context) error {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGetTx should panic when TxMiddleware not configured")
			}
		}()
		database.MustGetTx(c) // Should panic.
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/insert", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}

// Test 16: GetTxOrError returns Tx when configured.
func TestGetTxOrError_Success(t *testing.T) {
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
		tx, err := database.GetTxOrError(c)
		if err != nil {
			return err
		}
		_, err = tx.Exec(c.Request.Context(), "INSERT INTO test (name) VALUES (?)", "Ivy")
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
		t.Errorf("expected 1 row, got %d", count)
	}
}

// Test 17: GetTxOrError returns error when not configured.
func TestGetTxOrError_Error(t *testing.T) {
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	router := fursy.New()
	router.Use(database.Middleware(db))
	// NO TxMiddleware!

	router.POST("/insert", func(c *fursy.Context) error {
		tx, err := database.GetTxOrError(c)
		if err == nil {
			t.Error("GetTxOrError should return error when TxMiddleware not configured")
		}
		if tx != nil {
			t.Error("GetTxOrError should return nil Tx when not configured")
		}
		// Convert Problem error to response.
		var prob fursy.Problem
		if errors.As(err, &prob) {
			return c.Problem(prob)
		}
		return err
	})

	req := httptest.NewRequest("POST", "/insert", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected 500 Internal Server Error, got %d", w.Code)
	}
}
