# fursy plugins/database

Database integration plugin for fursy HTTP router. Provides seamless integration with `database/sql` for any SQL driver (PostgreSQL, MySQL, SQLite, etc.).

## Features

- **Database Middleware**: Share database connection across handlers
- **Transaction Helpers**: Easy transaction management with auto-commit/rollback
- **Context Integration**: `c.DB()` for convenient database access
- **Generic SQL Support**: Works with any `database/sql` driver
- **Zero External Dependencies**: Only stdlib `database/sql`

## Installation

```bash
go get github.com/coregx/fursy/plugins/database
go get github.com/lib/pq  # PostgreSQL driver (example)
```

## Quick Start

```go
package main

import (
    "database/sql"
    "github.com/coregx/fursy"
    "github.com/coregx/fursy/plugins/database"
    _ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
    // Open database connection.
    sqlDB, err := sql.Open("postgres", "user=postgres dbname=mydb sslmode=disable")
    if err != nil {
        panic(err)
    }
    defer sqlDB.Close()

    // Wrap with fursy database plugin.
    db := database.NewDB(sqlDB)

    // Create router with database middleware.
    router := fursy.New()
    router.Use(database.Middleware(db))

    // Use database in handlers.
    router.GET("/users/:id", func(c *fursy.Context) error {
        retrievedDB, ok := database.GetDB(c)
        if !ok {
            return c.Problem(fursy.InternalServerError("Database not configured"))
        }

        var user User
        err := retrievedDB.QueryRow(c.Request.Context(),
            "SELECT * FROM users WHERE id = $1", c.Param("id")).
            Scan(&user.ID, &user.Name)

        if err == sql.ErrNoRows {
            return c.Problem(fursy.NotFound("User not found"))
        }
        if err != nil {
            return c.Problem(fursy.InternalServerError(err.Error()))
        }

        return c.JSON(200, user)
    })

    router.Run(":8080")
}
```

## API Reference

### DB Type

```go
type DB struct {
    // Wraps *sql.DB with context support
}
```

#### Methods

- `NewDB(db *sql.DB) *DB` - Create new DB wrapper
- `DB() *sql.DB` - Get underlying `*sql.DB`
- `Ping(ctx context.Context) error` - Verify connection
- `Close() error` - Close connection
- `Exec(ctx, query, args...)` - Execute query without rows
- `Query(ctx, query, args...)` - Execute query returning rows
- `QueryRow(ctx, query, args...)` - Execute query returning single row
- `BeginTx(ctx, opts) (*Tx, error)` - Start transaction

### Middleware

```go
func Middleware(db *DB) fursy.HandlerFunc
```

Stores database in request context, making it available to all handlers.

**Usage:**

```go
db := database.NewDB(sqlDB)
router.Use(database.Middleware(db))
```

### GetDB Helper

```go
func GetDB(c *fursy.Context) (*DB, bool)
```

Retrieves database from context.

**Returns:**
- `*DB, true` if database is configured
- `nil, false` if middleware not configured

**Usage:**

```go
db, ok := database.GetDB(c)
if !ok {
    return c.Problem(fursy.InternalServerError("Database not configured"))
}
```

## Transactions

### Tx Type

```go
type Tx struct {
    // Wraps *sql.Tx with context support
}
```

#### Methods

- `Commit() error` - Commit transaction
- `Rollback() error` - Rollback transaction
- `Exec(ctx, query, args...)` - Execute query without rows
- `Query(ctx, query, args...)` - Execute query returning rows
- `QueryRow(ctx, query, args...)` - Execute query returning single row

### Manual Transactions

```go
router.POST("/transfer", func(c *fursy.Context) error {
    db, _ := database.GetDB(c)

    tx, err := db.BeginTx(c.Request.Context(), nil)
    if err != nil {
        return err
    }
    defer tx.Rollback() // Rollback if not committed

    // ... do work ...

    return tx.Commit()
})
```

### WithTx Helper

```go
func WithTx(ctx context.Context, db *DB, fn func(*Tx) error) error
```

Executes a function within a transaction. Automatically commits on success, rolls back on error.

**Usage:**

```go
err := database.WithTx(c.Request.Context(), db, func(tx *database.Tx) error {
    _, err := tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "Alice")
    if err != nil {
        return err // Automatic rollback
    }
    _, err = tx.Exec(ctx, "INSERT INTO audit (action) VALUES ($1)", "user_created")
    return err // Automatic commit on nil error
})
```

### Transaction Middleware

```go
func TxMiddleware(db *DB) fursy.HandlerFunc
```

Wraps each request in a database transaction. Auto-commits on success, auto-rolls back on error.

**Usage:**

```go
// Apply to specific routes that need transactions.
txGroup := router.Group("/api/v1")
txGroup.Use(database.Middleware(db))
txGroup.Use(database.TxMiddleware(db))

txGroup.POST("/users", func(c *fursy.Context) error {
    tx, _ := database.GetTx(c)
    // Use tx for all database operations
    // Auto-commit on success, auto-rollback on error
    return nil
})
```

**GetTx Helper:**

```go
func GetTx(c *fursy.Context) (*Tx, bool)
```

Retrieves transaction from context (requires TxMiddleware).

## Examples

### CRUD Operations

See [examples/09-rest-api-with-db](../../examples/09-rest-api-with-db/) for a complete REST API example with:
- Create, Read, Update, Delete operations
- Transaction management
- Error handling with RFC 9457
- Batch operations

### Batch Insert with Transaction

```go
router.POST("/users/batch", func(c *fursy.Context) error {
    db, _ := database.GetDB(c)

    var users []User
    json.NewDecoder(c.Request.Body).Decode(&users)

    var count int
    err := database.WithTx(c.Request.Context(), db, func(tx *database.Tx) error {
        for _, user := range users {
            _, err := tx.Exec(c.Request.Context(),
                "INSERT INTO users (name) VALUES ($1)", user.Name)
            if err != nil {
                return err // Rollback entire batch
            }
            count++
        }
        return nil // Commit all inserts
    })

    if err != nil {
        return c.Problem(fursy.InternalServerError(err.Error()))
    }

    return c.Created(map[string]int{"inserted": count})
})
```

### Error Handling

```go
router.GET("/users/:id", func(c *fursy.Context) error {
    db, _ := database.GetDB(c)

    var user User
    err := db.QueryRow(c.Request.Context(),
        "SELECT id, name FROM users WHERE id = $1", c.Param("id")).
        Scan(&user.ID, &user.Name)

    if err == sql.ErrNoRows {
        return c.Problem(fursy.NotFound("User not found"))
    }
    if err != nil {
        return c.Problem(fursy.InternalServerError(err.Error()))
    }

    return c.JSON(200, user)
})
```

## Supported Databases

Any `database/sql` compatible driver:

- **PostgreSQL**: `github.com/lib/pq` or `github.com/jackc/pgx/v5/stdlib`
- **MySQL**: `github.com/go-sql-driver/mysql`
- **SQLite**: `github.com/mattn/go-sqlite3`
- **SQL Server**: `github.com/denisenkom/go-mssqldb`
- **Oracle**: `github.com/godror/godror`

## dbcontext Pattern

The **dbcontext pattern** refers to best practices for managing database connections in request context. This plugin provides three approaches with different trade-offs:

### Approach 1: GetDBOrError (Recommended for Production)

**Use when**: You want clean error handling with RFC 9457 Problem Details.

```go
router.GET("/users/:id", func(c *fursy.Context) error {
    db, err := database.GetDBOrError(c)
    if err != nil {
        return err // Returns 500 Internal Server Error
    }

    var user User
    err = db.QueryRow(c.Request.Context(),
        "SELECT id, name FROM users WHERE id = $1", c.Param("id")).
        Scan(&user.ID, &user.Name)

    if err == sql.ErrNoRows {
        return c.Problem(fursy.NotFound("User not found"))
    }
    return c.JSON(200, user)
})
```

**Pros:**
- Clean, production-ready error handling
- Returns RFC 9457 compliant errors
- Single line to get DB with error handling

**Cons:**
- Requires error check on every handler

### Approach 2: MustGetDB (For Prototyping)

**Use when**: Rapid prototyping or when DB absence indicates programming error.

```go
router.GET("/users", func(c *fursy.Context) error {
    db := database.MustGetDB(c) // Panics if middleware not configured

    rows, err := db.Query(c.Request.Context(), "SELECT * FROM users")
    // ... handle query errors only
})
```

**Pros:**
- Minimal boilerplate
- Fast to write during development

**Cons:**
- Panics on misconfiguration (not production-friendly)
- Less explicit error handling

### Approach 3: GetDB (Manual Control)

**Use when**: You need custom error handling or conditional DB usage.

```go
router.GET("/users", func(c *fursy.Context) error {
    db, ok := database.GetDB(c)
    if !ok {
        // Custom error handling
        return c.Problem(fursy.Problem{
            Type:   "https://example.com/errors/db-not-configured",
            Title:  "Database Unavailable",
            Status: 503,
            Detail: "Service is temporarily unavailable",
        })
    }
    // Use db...
})
```

**Pros:**
- Full control over error handling
- Can return custom error responses

**Cons:**
- More verbose
- Requires manual error construction

### When to Use Each Approach

| Scenario | Recommended Approach |
|----------|---------------------|
| **Production REST API** | `GetDBOrError()` - Clean errors |
| **Internal Admin Panel** | `MustGetDB()` - Fast prototyping |
| **Microservice Health Check** | `GetDB()` - Custom 503 responses |
| **Conditional DB Usage** | `GetDB()` - Check availability |

### Transaction Patterns

Similar helpers exist for transactions:

**GetTxOrError (Recommended):**
```go
txGroup := router.Group("/api")
txGroup.Use(database.TxMiddleware(db))

txGroup.POST("/transfer", func(c *fursy.Context) error {
    tx, err := database.GetTxOrError(c)
    if err != nil {
        return err
    }
    // Use tx - auto-commit on success, auto-rollback on error
})
```

**MustGetTx (Prototyping):**
```go
txGroup.POST("/batch", func(c *fursy.Context) error {
    tx := database.MustGetTx(c) // Panics if TxMiddleware not configured
    // Use tx...
})
```

### Repository Pattern Integration

Combine dbcontext with repository pattern for clean separation:

```go
// Repository encapsulates database operations
type UserRepository struct {
    db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.QueryRow(ctx,
        "SELECT id, name FROM users WHERE id = $1", id).
        Scan(&user.ID, &user.Name)

    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    return &user, err
}

// Handler using repository pattern
router.GET("/users/:id", func(c *fursy.Context) error {
    db, err := database.GetDBOrError(c)
    if err != nil {
        return err
    }

    repo := NewUserRepository(db)
    user, err := repo.FindByID(c.Request.Context(), c.Param("id"))
    if err == ErrUserNotFound {
        return c.Problem(fursy.NotFound("User not found"))
    }
    if err != nil {
        return c.Problem(fursy.InternalServerError(err.Error()))
    }

    return c.JSON(200, user)
})
```

**Benefits:**
- Testable (mock repository interface)
- Clean separation of concerns
- Reusable across handlers
- Type-safe domain errors

## Best Practices

### Connection Pooling

Configure connection pool settings for production:

```go
sqlDB, _ := sql.Open("postgres", dsn)

// Configure pool
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)

db := database.NewDB(sqlDB)
```

### Use Transactions for Multi-Step Operations

Always use transactions for operations that modify multiple rows or tables:

```go
database.WithTx(ctx, db, func(tx *database.Tx) error {
    // Step 1: Insert user
    // Step 2: Insert audit log
    // Both succeed or both fail
    return nil
})
```

### Context Cancellation

Always pass request context to database operations:

```go
db.QueryRow(c.Request.Context(), query, args...) // Use c.Request.Context()
```

This ensures:
- Operations are canceled if client disconnects
- Timeout policies are enforced
- Graceful shutdown works correctly

### Prepared Statements

For frequently executed queries, use prepared statements:

```go
stmt, _ := db.DB().PrepareContext(ctx, "SELECT * FROM users WHERE id = $1")
defer stmt.Close()
// Use stmt.QueryRowContext(ctx, id)
```

## Testing

Run tests:

```bash
cd plugins/database
go test -v ./...
```

Run with coverage:

```bash
go test -coverprofile=coverage.txt ./...
go tool cover -html=coverage.txt
```

## License

MIT License - see [LICENSE](../../LICENSE) for details.

## Related

- [fursy Router](https://github.com/coregx/fursy)
- [RFC 9457 Problem Details](https://datatracker.ietf.org/doc/html/rfc9457)
- [database/sql Package](https://pkg.go.dev/database/sql)

## See Also

- [fursy Router](../../README.md) - Main router documentation
- [plugins/stream](../stream/README.md) - SSE + WebSocket integration
- [Examples](../../examples/)
  - [REST API with Database](../../examples/09-rest-api-with-db/)

## Database Drivers

Compatible with any `database/sql` driver:

- PostgreSQL: `github.com/lib/pq`
- MySQL: `github.com/go-sql-driver/mysql`
- SQLite: `modernc.org/sqlite` (pure Go, no CGO)
- SQL Server: `github.com/denisenkom/go-mssqldb`

See [database/sql drivers](https://github.com/golang/go/wiki/SQLDrivers) for full list.
