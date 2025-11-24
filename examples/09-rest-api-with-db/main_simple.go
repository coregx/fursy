// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package main demonstrates a simplified REST API with database integration using dbcontext helpers.
//
// This example shows the recommended pattern using Get DB OrError() for production code.
//
// Run:
//
//	go run main_simple.go
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
package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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
	// Open SQLite database.
	sqlDB, err := sql.Open("sqlite", "./users_simple.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	// Create users table.
	_, err = sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Wrap database with fursy integration.
	db := database.NewDB(sqlDB)

	// Create router.
	router := fursy.New()
	router.Use(database.Middleware(db))

	// CREATE user - Using GetDBOrError (recommended pattern).
	router.POST("/users", func(c *fursy.Context) error {
		db, err := database.GetDBOrError(c)
		if err != nil {
			var prob fursy.Problem
			if errors.As(err, &prob) {
				return c.Problem(prob)
			}
			return err
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
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		id, _ := result.LastInsertId()
		user.ID = int(id)
		return c.Created(user)
	})

	// READ user by ID - Using MustGetDB (prototyping pattern).
	router.GET("/users/:id", func(c *fursy.Context) error {
		db := database.MustGetDB(c) // Simpler, but panics if misconfigured.

		var user User
		err := db.QueryRow(c.Request.Context(),
			"SELECT id, name FROM users WHERE id = ?", c.Param("id")).
			Scan(&user.ID, &user.Name)

		if err == sql.ErrNoRows {
			return c.Problem(fursy.NotFound("User not found"))
		}
		if err != nil {
			return c.Problem(fursy.InternalServerError(err.Error()))
		}

		return c.OK(user)
	})

	// Start server.
	log.Println("Simplified REST API listening on :8080")
	log.Println("Database: ./users_simple.db (SQLite)")
	log.Println("")
	log.Println("Try:")
	log.Println("  curl -X POST http://localhost:8080/users -H 'Content-Type: application/json' -d '{\"name\":\"Alice\"}'")
	log.Println("  curl http://localhost:8080/users/1")
	log.Fatal(router.Run(":8080"))
}
