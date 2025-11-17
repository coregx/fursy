// Package main demonstrates the simplest possible fursy application.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/coregx/fursy"
)

func main() {
	// Create a new fursy router.
	router := fursy.New()

	// Define a simple GET endpoint.
	router.GET("/", func(c *fursy.Context) error {
		return c.OK(map[string]string{
			"message": "Hello, World!",
			"status":  "success",
		})
	})

	// Start the server.
	slog.Info("Server starting", "port", 8080)
	slog.Info("Visit: http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
