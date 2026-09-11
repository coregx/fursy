# 03 — REST API with generated OpenAPI

A small bookstore API that showcases how fursy's **type-safe handlers** drive
**OpenAPI 3.1** generation. Unlike `02-rest-api-crud` (which focuses on CRUD
plumbing), this example focuses on the documentation surface you get for free.

## Features showcased

- **Schemas from types** — `Box[Req, Res]` request/response types become JSON Schemas.
- **Named components + `$ref`** — named types are registered once in
  `components.schemas` and referenced, not duplicated.
- **Nested composition** — `Book.publisher` → `$ref Publisher`;
  `BookList.books` → array of `$ref Book`.
- **Accurate status codes** — `SuccessStatus: 201` (POST) and `204` (DELETE, no body).
- **Required vs optional** — pointer fields in `UpdateBookRequest` make them optional.
- **Optional request body** — `OptionalRequestBody` on `POST /books/search`.
- **Auto path parameters** — `/books/:id` is declared as a required `in: path` param.
- **Auto `operationId`** — e.g. `getBooks`, `getBooksById`, `postBooks`.
- **Summaries, tags, servers, deprecation** — via `RouteOptions`, `WithInfo`, `WithServer`.
- **Route groups** — `/admin/stats` is registered through a `RouteGroup` and documented.

## Running the example

```bash
# From this directory
go run .
```

- Swagger UI: http://localhost:8080/
- OpenAPI document: http://localhost:8080/openapi.json

## Endpoints

| Method | Path | Success | Notes |
|--------|------|---------|-------|
| GET    | `/books` | 200 | `BookList` |
| POST   | `/books` | 201 | `CreateBookRequest` → `Book`, `Location` header |
| GET    | `/books/:id` | 200 | `Book` |
| PATCH  | `/books/:id` | 200 | `UpdateBookRequest` (partial) |
| DELETE | `/books/:id` | 204 | no body |
| POST   | `/books/search` | 200 | optional body; send `{}` to match all |
| GET    | `/admin/stats` | 200 | grouped under `/admin` |
| GET    | `/legacy/books` | 200 | marked `deprecated: true` |

## Try it

```bash
# List books (seeded with two)
curl http://localhost:8080/books

# Create a book
curl -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{"title":"The Art of Computer Programming","author":"Donald Knuth","publisher_id":2,"year":1968,"tags":["computing"]}'

# Partial update
curl -X PATCH http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{"year":1844}'

# Search (body is optional in the spec; at runtime send at least {})
curl -X POST http://localhost:8080/books/search \
  -H "Content-Type: application/json" \
  -d '{"tag":"computing"}'

# Delete
curl -i -X DELETE http://localhost:8080/books/2
```

## What to look for in `/openapi.json`

- **`components.schemas`** contains `Book`, `Publisher`, `BookList`,
  `CreateBookRequest`, `UpdateBookRequest`, `SearchBooksRequest`, `Stats`, and
  `Problem`.
- **`$ref` reuse** — `Book.publisher` references `Publisher`; `BookList.books`
  items reference `Book`; the same `Book` component is reused by every operation.
- **Status codes** — `POST /books` documents `201`, `DELETE /books/:id`
  documents `204` (no content), and errors document `400`/`500` as `Problem`.
- **Optionality** — `CreateBookRequest` lists all non-`omitempty` fields as
  `required`; `UpdateBookRequest` (pointer fields) has no required fields.
- **`parameters`** — `GET /books/{id}` declares `id` as `required: true`, `in: path`.
- **`operationId`** — deterministic and unique, e.g. `getBooksById`, `postBooksSearch`.
- **`deprecated: true`** on `GET /legacy/books`.

## Note on `OptionalRequestBody`

`OptionalRequestBody` affects the **document** (`required: false` on the request
body). fursy's request binder still expects a JSON body at runtime, so send `{}`
when you have no filters.

## Files

```
03-rest-api-with-openapi/
├── main.go      - server setup, routes, and OpenAPI metadata
├── models.go    - request/response types (schemas are inferred from these)
├── handlers.go  - type-safe handlers
├── store.go     - in-memory, thread-safe store
├── index.html   - Swagger UI
├── go.mod
└── README.md
```

## Related

- [`02-rest-api-crud`](../02-rest-api-crud/) — CRUD without OpenAPI annotation.
- [`01-hello-world`](../01-hello-world/) — minimal setup.
