# Pockmon

Personal finance REST API in Go. Firebase handles authentication identity; this service owns the financial data (accounts, transactions, categories, budgets, goals).

## Tech Stack

- **Go 1.26** with **Gin** web framework
- **PostgreSQL** via `pgx/v5` (pgxpool for connection pooling)
- **sqlc** for type-safe query generation — all code in `internal/database/` is generated, never edit it directly
- **goose** for database migrations
- **Firebase Admin SDK** for token verification and user management

## Commands

```bash
# Run the server
go run ./cmd/api/

# Build
go build ./...

# After editing sql/queries/*.sql, regenerate the database layer
sqlc generate

# Migrations run automatically on startup via goose.Up()
# To run manually against a DB:
goose -dir sql/schema postgres "$DB_URL" up
```

## Environment

Requires a `.env` file (or environment variables) and the Firebase service account key:

```
DB_URL=postgres://user:password@localhost:5432/pockmon
```

`firebase-pockmon.json` — Firebase service account key, must be present at the project root.

## Directory Structure

```
cmd/api/
  main.go        # entry point: load env, connect DB, run migrations, start server
  server.go      # Gin router setup, all route definitions
  db.go          # DB connection pool and goose migration runner

internal/
  config/        # Config struct (DB pool, Querier, FirebaseApp, AuthClient)
  database/      # sqlc-generated — NEVER edit manually
  handler/       # One file per domain: auth, user, category, account, transaction
  middleware/    # firebase.go (Auth), error.go (AppError, ErrorHandler), logger.go
  util/          # convert.go (pgtype helpers), seed.go (default categories)

sql/
  schema/        # goose migration files (001_users.sql … 011_account_adjustments.sql)
  queries/       # sqlc query files — edit these, then run sqlc generate
```

## Error Handling Pattern

All handlers use this pattern — never `c.AbortWithStatusJSON`:

```go
c.Error(&gin.Error{
    Err:  middleware.NewAppError(http.StatusXxx, "client error", "client message").WithInternal(err.Error()),
    Type: gin.ErrorTypePublic,
})
c.Abort()
return
```

The `ErrorHandler` middleware reads the last error and writes a consistent JSON envelope:
```json
{"status": 404, "error": "not found", "message": "account not found", "path": "...", "timestamp": "..."}
```

`WithInternal(detail)` attaches a server-only message that the Logger middleware logs but `ErrorHandler` never sends to the client.

## Auth Flow

1. Client authenticates with Firebase directly and obtains a JWT.
2. Client sends `Authorization: Bearer <jwt>` on every request.
3. `middleware.Auth` calls `authClient.VerifyIDTokenAndCheckRevoked` and stores the token as `c.Set("firebaseToken", token)`.
4. Handlers retrieve it via `c.MustGet("firebaseToken").(*auth.Token)` and look up the local user with `GetUserByFirebaseUID(token.UID)`.

The local `users` table exists as a FK anchor for financial data and stores `currency_code` and `is_admin`. Display name, email, and photo URL come from the token claims.

## Key Patterns

**pgtype helpers** — always use `util/convert.go` to convert Go types to pgtype:
```go
util.GenerateUUID()          // new UUIDv7 → pgtype.UUID
util.GetUUID(string)         // parse string → pgtype.UUID (returns error on invalid)
util.GetNumeric(float64)     // → pgtype.Numeric
util.GetDate(string)         // → pgtype.Date  (format: "2006-01-02")
```

**Not-found detection** — import `"github.com/jackc/pgx/v5"` and check:
```go
if errors.Is(err, pgx.ErrNoRows) { /* 404 */ }
```

**DB transactions** — use `h.config.DB.Begin(ctx)` then `database.New(tx)` for atomic operations (e.g., CreateTransaction + UpdateAccountBalance):
```go
tx, err := h.config.DB.Begin(ctx)
defer tx.Rollback(ctx)
q := database.New(tx)
// … operations …
tx.Commit(ctx)
```

**Admin check** — user.IsAdmin must be checked explicitly in handlers that require it (ListUsers, ToggleAdmin, DeleteUser). There is no admin middleware.

**First-user promotion** — `Register` calls `CountUser` after inserting; if count == 1, promotes the new user to admin via `SetUserAdmin`.

**Account balance** — `CreateTransaction` updates `accounts.current_balance` atomically within a DB transaction. Income adds to balance, expense subtracts.

## Middleware Chain

`gin.Recovery` → `middleware.Logger` → `middleware.ErrorHandler` → handler

`ErrorHandler` must come after `Logger` because it writes the response body; both run on the way back up after `c.Next()`.

## Validation Conventions

- Category `type` must be `expense` or `income` — use `binding:"required,oneof=expense income"` in request structs.
- Transaction `type` must be `expense` or `income`.
- UUID path params that fail parsing → **400** (not 404 or 500).
- DB not-found (pgx.ErrNoRows) → **404**.
- 204 responses use `c.Status(http.StatusNoContent)`, never `c.JSON(204, nil)`.

## Routes

```
POST   /auth/register
GET    /auth/me
PUT    /auth/me/currency/:code

GET    /user                   (admin only)
PUT    /user/:id               (admin only — toggle is_admin)
DELETE /user/:id               (admin only — deletes DB row + Firebase account)

GET    /api/category
GET    /api/category/:id
POST   /api/category
PUT    /api/category/:id
DELETE /api/category/:id

GET    /api/account
GET    /api/account/:id
POST   /api/account
PUT    /api/account/:id

GET    /api/transaction                    (?limit=20&offset=0, max limit 100)
GET    /api/transaction/category/:id
GET    /api/transaction/account/:id
GET    /api/transaction/type/:type         (income | expense)
GET    /api/transaction/tags               (?tags=food&tags=transport)
POST   /api/transaction
PUT    /api/transaction/:id
DELETE /api/transaction/:id
```

All routes require `Authorization: Bearer <firebase-jwt>`.
