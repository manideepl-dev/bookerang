# Bookerang

A small Go HTTP API for sharing books with nearby users.

## Run locally

Requirements:

- Go 1.24+
- PostgreSQL with PostGIS enabled

Set these environment variables (a local `.env` file is also supported):

```text
DATABASE_URL=postgres://user:password@localhost:5432/bookerang?sslmode=disable
JWT_SECRET=replace-with-a-long-random-secret
CORS_ORIGIN=http://localhost:3000
PORT=8080
```

Apply the SQL files in `migrations/` in order, then start the API:

```sh
go test ./...
go run .
```

The server listens on `http://localhost:8080` by default.

## Project layout

- `main.go`: configuration, database connection, dependency wiring, and HTTP server.
- `internal/httpapi`: HTTP routes, middleware, JSON DTOs, and transport-level validation.
- `internal/services`: application use cases, validation, and typed errors.
- `internal/repository`: PostgreSQL queries and persistence errors.
- `internal/domain`: data passed between application layers.
- `internal/jwt`: token creation and validation.
- `migrations`: database schema changes, applied in filename order.

Keep JSON tags in `httpapi` types. Keep SQL and database details in `repository`. Services should express business rules without knowing HTTP details.

## Learning path

1. Start with `main.go` and trace how dependencies are constructed.
2. Read one request end to end: route in `internal/httpapi/routes.go`, use case in `internal/services`, then SQL in `internal/repository`.
3. Practice Go errors by following `errors.Is` from services to handlers.
4. Add or improve one service test using the small repository fakes in `internal/services/*_test.go`.
5. Add an HTTP handler test with `httptest` in `internal/httpapi`.
6. Learn SQL transactions by studying `BookRepository.AddBook` and its rollback behavior.
7. Run `go test ./...`, `go test -race ./...`, and `go vet ./...` before deploying changes.

This project currently expects migrations to be applied as a deployment step. Before the schema grows, add an automated migration command or a migration tool to the deployment pipeline so application code and schema cannot drift.
