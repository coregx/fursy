# REST API with Database Example

Complete REST API example demonstrating database integration with fursy router.

## Features

- **Database Middleware**: Share database connection across handlers
- **CRUD Operations**: Create, Read, Update, Delete
- **Transaction Management**: WithTx helper and manual transactions
- **Error Handling**: RFC 9457 Problem Details
- **Batch Operations**: Atomic multi-row inserts with transactions

## Prerequisites

```bash
go get github.com/coregx/fursy
go get github.com/coregx/fursy/plugins/database
go get github.com/mattn/go-sqlite3
```

## Run

```bash
go run main.go
```

Server starts on http://localhost:8080

Database file: `./users.db` (SQLite, persisted on disk)

## API Endpoints

### Create User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice"}'

# Response: 201 Created
{
  "id": 1,
  "name": "Alice"
}
```

### Get User

```bash
curl http://localhost:8080/users/1

# Response: 200 OK
{
  "id": 1,
  "name": "Alice"
}
```

### List Users

```bash
curl http://localhost:8080/users

# Response: 200 OK
[
  {
    "id": 1,
    "name": "Alice"
  },
  {
    "id": 2,
    "name": "Bob"
  }
]
```

### Update User

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated"}'

# Response: 200 OK
{
  "status": "updated"
}
```

### Delete User

```bash
curl -X DELETE http://localhost:8080/users/1

# Response: 204 No Content
```

### Batch Create (with Transaction)

```bash
curl -X POST http://localhost:8080/users/batch \
  -H "Content-Type: application/json" \
  -d '[{"name":"Bob"},{"name":"Charlie"},{"name":"Dave"}]'

# Response: 201 Created
{
  "inserted": 3
}
```

If any user in the batch is invalid, the entire transaction rolls back (all-or-nothing).

### Rename User (Manual Transaction)

```bash
curl -X POST http://localhost:8080/users/1/rename \
  -H "Content-Type: application/json" \
  -d '{"new_name":"Alice Smith"}'

# Response: 200 OK
{
  "old_name": "Alice",
  "new_name": "Alice Smith"
}
```

## Error Handling

All errors return RFC 9457 Problem Details:

```bash
# User not found
curl http://localhost:8080/users/999

# Response: 404 Not Found
{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "User not found"
}
```

```bash
# Invalid JSON
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d 'invalid'

# Response: 400 Bad Request
{
  "type": "about:blank",
  "title": "Bad Request",
  "status": 400,
  "detail": "Invalid JSON"
}
```

## Key Concepts

### Database Middleware

```go
db := database.NewDB(sqlDB)
router.Use(database.Middleware(db))
```

Makes database available in all handlers via `database.GetDB(c)`.

### CRUD Operations

Simple database operations without transactions:

```go
router.GET("/users/:id", func(c *fursy.Context) error {
    db, _ := database.GetDB(c)

    var user User
    err := db.QueryRow(ctx,
        "SELECT id, name FROM users WHERE id = ?",
        c.Param("id")).Scan(&user.ID, &user.Name)

    if err == sql.ErrNoRows {
        return c.Problem(fursy.NotFound("User not found"))
    }
    return c.OK(user)
})
```

### Transactions with WithTx

Automatic commit/rollback:

```go
err := database.WithTx(ctx, db, func(tx *database.Tx) error {
    // All operations succeed = commit
    // Any error = rollback
    _, err := tx.Exec(ctx, "INSERT INTO users ...")
    return err
})
```

### Manual Transactions

Full control over transaction lifecycle:

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback() // Rollback if not committed

// ... operations ...

return tx.Commit()
```

## Database Schema

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
);
```

## Clean Up

```bash
# Remove database file
rm users.db
```

## Next Steps

- Try PostgreSQL or MySQL drivers (same API, different connection string)
- Add validation with `plugins/validator`
- Add authentication with `middleware/jwt.go`
- Add request logging with `middleware/logger.go`
- Add transactions middleware for automatic transaction management

## Related Examples

- `01-hello-world` - Minimal fursy setup
- `02-rest-api-crud` - REST API without database (in-memory)
- `05-middleware` - All middleware examples

## Learn More

- [fursy Documentation](https://github.com/coregx/fursy)
- [plugins/database API](../../plugins/database/README.md)
- [RFC 9457 Problem Details](https://datatracker.ietf.org/doc/html/rfc9457)
