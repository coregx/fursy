package fursy_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coregx/fursy"
	"github.com/coregx/fursy/plugins/database"
	_ "modernc.org/sqlite"
)

// TestIntegration_Database_Basic tests basic database integration.
func TestIntegration_Database_Basic(t *testing.T) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	db := database.NewDB(sqlDB)
	router := fursy.New()
	router.Use(database.Middleware(db))

	router.GET("/ping", func(c *fursy.Context) error {
		dbConn, ok := database.GetDB(c)
		if !ok {
			return c.Problem(fursy.Problem{Status: 500, Detail: "DB not found"})
		}
		if err := dbConn.Ping(context.Background()); err != nil {
			return c.Problem(fursy.Problem{Status: 500, Detail: err.Error()})
		}
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/ping", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
