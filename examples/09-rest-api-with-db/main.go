// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package main demonstrates a complete REST API with database integration.
//
// This example shows:
//   - Database middleware configuration
//   - CRUD operations (Create, Read, Update, Delete)
//   - Transaction management
//   - Error handling with RFC 9457 Problem Details
//   - Batch operations with transactions
//
// Run:
//
//	go run main.go
//
// Test:
//
//	# Create user
//	curl -X POST http://localhost:8080/users \
//	  -H "Content-Type: application/json" \
//	  -d '{"name":"Alice"}'
//
//	# Get user
//	curl http://localhost:8080/users/1
//
//	# List users
//	curl http://localhost:8080/users
//
//	# Delete user
//	curl -X DELETE http://localhost:8080/users/1
//
//	# Batch create users (with transaction)
//	curl -X POST http://localhost:8080/users/batch \
//	  -H "Content-Type: application/json" \
//	  -d '[{"name":"Bob"},{"name":"Charlie"}]'
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/database"
	_ "modernc.org/sqlite" // Pure Go SQLite driver (no CGO required)
)

// User represents a user in the system.
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name" validate:"required"`
}

func main() {
	// Open SQLite database (file-based for persistence).
	// Note: modernc.org/sqlite uses "sqlite" driver name (not "sqlite3").
	sqlDB, err := sql.Open("sqlite", "./users.db")
	if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
		log.Fatal(err)
	}
	defer sqlDB.Close()

	// Create users table if not exists.
	_, err = sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
		log.Fatal(err)
	}

	// Wrap database with fursy integration.
	db := database.NewDB(sqlDB)

	// Create router.
	router := fursy.New()

	// Configure database middleware (makes db available in all handlers).
	router.Use(database.Middleware(db))

	// ===================
	// CRUD Endpoints
	// ===================

	// CREATE user.
	router.POST("/users", func(c *fursy.Context) error {
		// Use GetDBOrError for production-ready error handling.
		db, err := database.GetDBOrError(c)
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(err.(fursy.Problem))
		}

		var user User
		if err := json.NewDecoder(c.Request.Body).Decode(&user); err != nil {
			return c.Problem(fursy.BadRequest("Invalid JSON"))
		}

		if user.Name == "" {
			return c.Problem(fursy.BadRequest("Name is required"))
		}

		result, err := db.Exec(c.Request.Context(),
			"INSERT INTO users (name) VALUES (?)", user.Name)
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		id, _ := result.LastInsertId()
		user.ID = int(id)
		return c.Created(user)
	})

	// READ user by ID.
	router.GET("/users/:id", func(c *fursy.Context) error {
		db, err := database.GetDBOrError(c)
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(err.(fursy.Problem))
		}

		var user User
		err = db.QueryRow(c.Request.Context(),
			"SELECT id, name FROM users WHERE id = ?", c.Param("id")).
			Scan(&user.ID, &user.Name)

		if err == sql.ErrNoRows {
			return c.Problem(fursy.NotFound("User not found"))
		}
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		return c.OK(user)
	})

	// LIST all users.
	router.GET("/users", func(c *fursy.Context) error {
		db, err := database.GetDBOrError(c)
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(err.(fursy.Problem))
		}

		rows, err := db.Query(c.Request.Context(), "SELECT id, name FROM users")
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
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

		return c.OK(users)
	})

	// UPDATE user.
	router.PUT("/users/:id", func(c *fursy.Context) error {
		db, err := database.GetDBOrError(c)
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(err.(fursy.Problem))
		}

		var user User
		if err := json.NewDecoder(c.Request.Body).Decode(&user); err != nil {
			return c.Problem(fursy.BadRequest("Invalid JSON"))
		}

		if user.Name == "" {
			return c.Problem(fursy.BadRequest("Name is required"))
		}

		result, err := db.Exec(c.Request.Context(),
			"UPDATE users SET name = ? WHERE id = ?", user.Name, c.Param("id"))
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.Problem(fursy.NotFound("User not found"))
		}

		return c.OK(map[string]string{"status": "updated"})
	})

	// DELETE user.
	router.DELETE("/users/:id", func(c *fursy.Context) error {
		db, err := database.GetDBOrError(c)
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(err.(fursy.Problem))
		}

		result, err := db.Exec(c.Request.Context(),
			"DELETE FROM users WHERE id = ?", c.Param("id"))
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return c.Problem(fursy.NotFound("User not found"))
		}

		return c.NoContentSuccess()
	})

	// ===================
	// Batch Operations with Transactions
	// ===================

	// Batch create users (uses transaction to ensure atomicity).
	router.POST("/users/batch", func(c *fursy.Context) error {
		db, err := database.GetDBOrError(c)
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(err.(fursy.Problem))
		}

		var users []User
		if err := json.NewDecoder(c.Request.Body).Decode(&users); err != nil {
			return c.Problem(fursy.BadRequest("Invalid JSON"))
		}

		if len(users) == 0 {
			return c.Problem(fursy.BadRequest("No users provided"))
		}

		var count int
		err = database.WithTx(c.Request.Context(), db, func(tx *database.Tx) error {
			for _, user := range users {
				if user.Name == "" {
					return fursy.ErrBadRequest // Rollback entire transaction.
				}
				_, err := tx.Exec(c.Request.Context(),
					"INSERT INTO users (name) VALUES (?)", user.Name)
				if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
					return err // Rollback on any error.
				}
				count++
			}
			return nil // Commit transaction.
		})

		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		return c.Created(map[string]int{"inserted": count})
	})

	// ===================
	// Example: Manual Transaction
	// ===================

	// Transfer-like operation (demonstrates manual transaction management).
	router.POST("/users/:id/rename", func(c *fursy.Context) error {
		db, err := database.GetDBOrError(c)
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(err.(fursy.Problem))
		}

		var req struct {
			NewName string `json:"new_name"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
			return c.Problem(fursy.BadRequest("Invalid JSON"))
		}

		// Start manual transaction.
		tx, err := db.BeginTx(c.Request.Context(), nil)
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(fursy.InternalServerError(err.Error()))
		}
		defer tx.Rollback() // Rollback if not committed.

		// Check user exists.
		var oldName string
		err = tx.QueryRow(c.Request.Context(),
			"SELECT name FROM users WHERE id = ?", c.Param("id")).Scan(&oldName)
		if err == sql.ErrNoRows {
			return c.Problem(fursy.NotFound("User not found"))
		}
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		// Update name.
		_, err = tx.Exec(c.Request.Context(),
			"UPDATE users SET name = ? WHERE id = ?", req.NewName, c.Param("id"))
		if err != nil {
			return c.Problem(err.(fursy.Problem))
		}
		// Continue processing...
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		// Commit transaction.
		if err := tx.Commit(); err != nil {
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		return c.OK(map[string]string{
			"old_name": oldName,
			"new_name": req.NewName,
		})
	})

	// Start server.
	log.Println("REST API with database listening on :8080")
	log.Println("Database: ./users.db (SQLite)")
	log.Println("")
	log.Println("Try:")
	log.Println("  curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' -d '{\"name\":\"Alice\"}'")
	log.Println("  curl http://localhost:8080/users")
	log.Println("  curl http://localhost:8080/users/1")
	log.Println("  curl -X DELETE http://localhost:8080/users/1")
	log.Fatal(router.Run(":8080"))
}
