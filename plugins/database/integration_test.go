// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package database_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/database"
	_ "modernc.org/sqlite" // Pure Go SQLite driver for testing
)

// User is a test model for integration testing.
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Integration Test: Full CRUD example.
//
//nolint:gocognit // Integration test with multiple subtests is inherently complex
func TestIntegration_CRUD(t *testing.T) {
	// Setup database.
	sqlDB := setupDB(t)
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	ctx := context.Background()

	// Create table.
	_, err := db.Exec(ctx, `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Setup router.
	router := fursy.New()
	router.Use(database.Middleware(db))

	// CREATE endpoint.
	router.POST("/users", func(c *fursy.Context) error {
		retrievedDB, ok := database.GetDB(c)
		if !ok {
			return c.Problem(fursy.InternalServerError("DB not configured"))
		}

		var user User
		if err := json.NewDecoder(c.Request.Body).Decode(&user); err != nil {
			return c.Problem(fursy.BadRequest(err.Error()))
		}

		result, err := retrievedDB.Exec(c.Request.Context(),
			"INSERT INTO users (name) VALUES (?)", user.Name)
		if err != nil {
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		id, _ := result.LastInsertId()
		user.ID = int(id)
		return c.JSON(201, user)
	})

	// READ endpoint.
	router.GET("/users/:id", func(c *fursy.Context) error {
		retrievedDB, ok := database.GetDB(c)
		if !ok {
			return c.Problem(fursy.InternalServerError("DB not configured"))
		}

		var user User
		err := retrievedDB.QueryRow(c.Request.Context(),
			"SELECT id, name FROM users WHERE id = ?", c.Param("id")).
			Scan(&user.ID, &user.Name)

		if err == sql.ErrNoRows {
			return c.Problem(fursy.NotFound("User not found"))
		}
		if err != nil {
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		return c.JSON(200, user)
	})

	// LIST endpoint.
	router.GET("/users", func(c *fursy.Context) error {
		retrievedDB, ok := database.GetDB(c)
		if !ok {
			return c.Problem(fursy.InternalServerError("DB not configured"))
		}

		rows, err := retrievedDB.Query(c.Request.Context(), "SELECT id, name FROM users")
		if err != nil {
			return c.Problem(fursy.InternalServerError(err.Error()))
		}
		defer rows.Close()

		users := []User{}
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Name); err != nil {
				return c.Problem(fursy.InternalServerError(err.Error()))
			}
			users = append(users, user)
		}

		return c.JSON(200, users)
	})

	// DELETE endpoint.
	router.DELETE("/users/:id", func(c *fursy.Context) error {
		retrievedDB, ok := database.GetDB(c)
		if !ok {
			return c.Problem(fursy.InternalServerError("DB not configured"))
		}

		_, err := retrievedDB.Exec(c.Request.Context(),
			"DELETE FROM users WHERE id = ?", c.Param("id"))
		if err != nil {
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		return c.JSON(200, map[string]string{"status": "deleted"})
	})

	// Test CREATE.
	t.Run("CREATE user", func(t *testing.T) {
		createReq := `{"name":"Alice"}`
		req := httptest.NewRequest("POST", "/users", strings.NewReader(createReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 201 {
			t.Fatalf("CREATE failed: expected 201, got %d", w.Code)
		}

		var created User
		if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if created.Name != "Alice" {
			t.Errorf("wrong name: got %q, want %q", created.Name, "Alice")
		}
		if created.ID == 0 {
			t.Error("expected non-zero ID")
		}
	})

	// Test READ.
	t.Run("READ user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/1", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("READ failed: expected 200, got %d", w.Code)
		}

		var read User
		if err := json.Unmarshal(w.Body.Bytes(), &read); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if read.Name != "Alice" {
			t.Errorf("wrong name: got %q, want %q", read.Name, "Alice")
		}
		if read.ID != 1 {
			t.Errorf("wrong ID: got %d, want %d", read.ID, 1)
		}
	})

	// Test LIST (after creating second user).
	t.Run("LIST users", func(t *testing.T) {
		// Create second user.
		createReq := `{"name":"Bob"}`
		req := httptest.NewRequest("POST", "/users", strings.NewReader(createReq))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// List all users.
		req = httptest.NewRequest("GET", "/users", http.NoBody)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("LIST failed: expected 200, got %d", w.Code)
		}

		var users []User
		if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if len(users) != 2 {
			t.Errorf("expected 2 users, got %d", len(users))
		}
	})

	// Test DELETE.
	t.Run("DELETE user", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/users/1", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("DELETE failed: expected 200, got %d", w.Code)
		}

		// Verify deleted.
		req = httptest.NewRequest("GET", "/users/1", http.NoBody)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 404 {
			t.Errorf("expected 404 after delete, got %d", w.Code)
		}
	})
}
