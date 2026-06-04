# KanBam Backend — Go Package Selection Plan

## Context

KanBam is a Kanban board REST API (Workspace → Board → Column → Card hierarchy) with JWT auth, Basic Auth for login, role-based access control, and a Postgres database. The schema is already defined in `KanBam-Schema.sql`. The goal is to choose and initialize the Go packages before writing any application code.

---

## Recommended Packages

### HTTP Router — `github.com/go-chi/chi/v5`
Chi handles the deeply-nested route groups in this API cleanly (e.g. `/workspaces/{workspaceID}/boards/{boardID}/columns`). Its `middleware` sub-package provides request ID, logger, recoverer, and it makes writing a JWT middleware straightforward because it stays on top of standard `net/http`.

### PostgreSQL Driver — `github.com/jackc/pgx/v5` + `pgxpool`
pgx is the highest-performance, most feature-complete Postgres driver for Go. `pgxpool` provides connection pooling and is used directly by sqlc-generated code.

### SQL Code Generation — `github.com/sqlc-dev/sqlc` (CLI tool, not a runtime dep)
sqlc reads `KanBam-Schema.sql` and hand-written `.sql` query files, then generates type-safe Go query functions and model structs. This eliminates row-scanning boilerplate and catches SQL/Go type mismatches at compile time. Installed via `go install` as a dev tool.

### JWT — `github.com/golang-jwt/jwt/v5`
The canonical Go JWT library. Used to sign tokens on register/login and validate `Authorization: Bearer` headers in Chi middleware.

### Password Hashing — `golang.org/x/crypto` (bcrypt)
`bcrypt.GenerateFromPassword` / `bcrypt.CompareHashAndPassword` for storing and verifying passwords. Part of the Go extended standard library.

### Input Validation — `github.com/go-playground/validator/v10`
Struct-tag validation (e.g. `validate:"required,min=8"`, `validate:"hexcolor"`, `validate:"email"`). Covers all the constraints in the API spec (password complexity, hex color regex, alphanumeric username, UUID format).

### UUID — `github.com/google/uuid`
`uuid.New()` generates UUID v4. Used for every primary key.

### Environment Config — `github.com/caarlos0/env/v11`
Parses environment variables directly into a typed Go struct (e.g. `DATABASE_URL`, `JWT_SECRET`, `PORT`). Zero dependencies, replaces viper for simple 12-factor config.

### `.env` Loading — `github.com/joho/godotenv`
Loads a `.env` file in development so env vars are available before `caarlos0/env` reads them. Only called in `main.go` behind a dev guard.

---

## Package Summary Table

| Concern | Package | Runtime dep? |
|---------|---------|-------------|
| HTTP routing | `github.com/go-chi/chi/v5` | yes |
| Postgres driver | `github.com/jackc/pgx/v5` | yes |
| SQL codegen | `github.com/sqlc-dev/sqlc` | no (tool) |
| JWT | `github.com/golang-jwt/jwt/v5` | yes |
| Password hashing | `golang.org/x/crypto` | yes |
| Validation | `github.com/go-playground/validator/v10` | yes |
| UUID | `github.com/google/uuid` | yes |
| Env config | `github.com/caarlos0/env/v11` | yes |
| .env loader | `github.com/joho/godotenv` | yes (dev) |

---

## Implementation Steps

1. **Add runtime dependencies** to `go.mod`:
   ```
   go get github.com/go-chi/chi/v5
   go get github.com/jackc/pgx/v5
   go get github.com/golang-jwt/jwt/v5
   go get golang.org/x/crypto
   go get github.com/go-playground/validator/v10
   go get github.com/google/uuid
   go get github.com/caarlos0/env/v11
   go get github.com/joho/godotenv
   ```

2. **Install sqlc CLI** (dev machine only):
   ```
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   ```

3. **Create `sqlc.yaml`** at repo root pointing at `KanBam-Schema.sql` and a `sql/queries/` directory, targeting the pgx/v5 driver and outputting to `internal/db/`.

4. **Fix the schema** — `KanBam-Schema.sql` has two syntax errors (missing commas after column definitions in the `Column` and `CardTag` tables) that will prevent sqlc from parsing it.

5. **Run `sqlc generate`** to produce `internal/db/models.go` and `internal/db/query.sql.go` from the schema + query files.

---

## Verification

- `go build ./...` compiles with no errors after `go get`
- `sqlc generate` produces files in `internal/db/` without errors
- A minimal `main.go` that spins up a Chi router on `:8080` and connects to Postgres via `pgxpool.New` returns 200 on `GET /health`
