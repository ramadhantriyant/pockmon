# Pockmon

Personal finance REST API written in Go. Firebase handles authentication identity; this service owns all financial data — accounts, transactions, categories, budgets, goals, and recurring transactions.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| Web framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | PostgreSQL via [pgx/v5](https://github.com/jackc/pgx) (connection pooling with pgxpool) |
| Query generation | [sqlc](https://sqlc.dev/) — all code in `internal/database/` is generated, never edit it directly |
| Migrations | [goose](https://github.com/pressly/goose) — runs automatically on startup |
| Authentication | Firebase Admin SDK (token verification) |
| File storage | Google Cloud Storage (signed URL upload pattern) |
| Scheduling | [robfig/cron v3](https://github.com/robfig/cron) — daily jobs for recurring transactions and notifications |

## Prerequisites

- Go 1.26+
- PostgreSQL 14+
- A Firebase project with **Authentication** and **Storage** enabled
- A Firebase service account key (for the Admin SDK)
- `sqlc` — only needed when editing SQL queries (`go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`)

## Environment Variables

Copy `.env.example` to `.env` and fill in the values:

```env
# PostgreSQL connection string
DB_URL=postgres://user:password@localhost:5432/pockmon

# Firebase service account key as a JSON string
# Paste the entire contents of your firebase-*.json file as a single-line string
GOOGLE_CREDENTIALS_JSON={"type":"service_account",...}

# Google Cloud Storage bucket name (from Firebase Storage)
# Format: <project-id>.firebasestorage.app
STORAGE_BUCKET=your-project.firebasestorage.app
```

> **How credentials work:** On startup the server writes `GOOGLE_CREDENTIALS_JSON` to a temp file and sets `GOOGLE_APPLICATION_CREDENTIALS` to point at it. Both the Firebase Admin SDK and GCS client pick it up via Application Default Credentials (ADC). The file persists for the lifetime of the process because ADC re-reads it on every token refresh.

## Running

```bash
# Development
go run ./cmd/api/

# Build
go build -o pockmon ./cmd/api/

# Run binary
./pockmon
```

The server starts on `:8080`. Database migrations run automatically on every startup via `goose.Up()`.

## Project Structure

```
cmd/api/
  main.go        # Entry point: load env → setup credentials → connect DB → run migrations → start server
  server.go      # Gin router and all route definitions
  db.go          # DB connection pool, goose migration runner, and Google credential setup

internal/
  config/        # Config struct (DB pool, Querier, FirebaseApp, AuthClient, StorageClient)
  database/      # sqlc-generated code — never edit manually
  handler/       # One file per domain (auth, user, category, account, transaction, budget, goal, …)
  middleware/    # Auth (Firebase JWT), ErrorHandler, Logger
  scheduler/     # Daily cron: processes recurring transactions, sends budget/goal/bill notifications
  util/          # UUID/Numeric/Date helpers, default category seeder

sql/
  schema/        # Goose migration files (001_users … 011_account_adjustments)
  queries/       # sqlc input files — edit these, then run sqlc generate
```

## Authentication

All routes require a Firebase ID token:

```
Authorization: Bearer <firebase-jwt>
```

1. The client authenticates directly with Firebase and obtains a JWT.
2. `middleware.Auth` calls `VerifyIDTokenAndCheckRevoked` on every request.
3. Handlers look up the local user via `GetUserByFirebaseUID(token.UID)`.

The local `users` table is a FK anchor for all financial data. It stores `currency_code` and `is_admin`. Profile fields (display name, email, photo) come from the Firebase token claims.

**First user** is automatically promoted to admin after registration.

## API Reference

### Auth

| Method | Path | Description |
|---|---|---|
| `POST` | `/auth/register` | Create local user record. Body: `{ "currency_code": "USD" }` |
| `GET` | `/auth/me` | Get current user profile |
| `PUT` | `/auth/me/currency/:code` | Update preferred currency |

### Users _(admin only)_

| Method | Path | Description |
|---|---|---|
| `GET` | `/user` | List all users |
| `PUT` | `/user/:id` | Toggle admin status |
| `DELETE` | `/user/:id` | Delete user (DB row + Firebase account) |

### Categories

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/category` | List categories (own + system defaults) |
| `GET` | `/api/category/:id` | Get single category |
| `POST` | `/api/category` | Create. Body: `{ name, type: "income"\|"expense", icon?, color? }` |
| `PUT` | `/api/category/:id` | Update |
| `DELETE` | `/api/category/:id` | Delete (non-system categories only) |

### Accounts

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/account` | List accounts |
| `GET` | `/api/account/:id` | Get account |
| `POST` | `/api/account` | Create. Body: `{ name, account_type, currency_code, initial_balance? }` |
| `PUT` | `/api/account/:id` | Update |
| `DELETE` | `/api/account/:id` | Deactivate |
| `GET` | `/api/account/:id/adjustment` | List balance adjustments |
| `POST` | `/api/account/:id/adjustment` | Create manual balance adjustment |

### Transactions

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/transaction` | List (`?limit=20&offset=0`, max 100) |
| `GET` | `/api/transaction/:id` | Get single transaction |
| `GET` | `/api/transaction/type/:type` | Filter by `income` or `expense` |
| `GET` | `/api/transaction/category/:id` | Filter by category |
| `GET` | `/api/transaction/account/:id` | Filter by account |
| `GET` | `/api/transaction/tags` | Filter by tags (`?tags=food&tags=transport`) |
| `POST` | `/api/transaction` | Create (updates account balance atomically) |
| `PUT` | `/api/transaction/:id` | Update |
| `DELETE` | `/api/transaction/:id` | Delete (reverses account balance) |
| `GET` | `/api/transaction/:id/attachment` | List attachments |
| `GET` | `/api/transaction/:id/attachment/upload-url` | Get signed PUT URL (`?file_name=&file_type=`) |
| `POST` | `/api/transaction/:id/attachment` | Confirm attachment after upload |

### Attachments

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/attachment/:id` | Get attachment with signed download URL |
| `DELETE` | `/api/attachment/:id` | Delete (GCS object + DB record) |

### Recurring Transactions

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/recurring` | List all |
| `GET` | `/api/recurring/active` | List active only |
| `GET` | `/api/recurring/:id` | Get single |
| `POST` | `/api/recurring` | Create. Body: `{ name, type, amount, account_id, frequency, start_date, auto_create? }` |
| `PUT` | `/api/recurring/:id` | Update (type and start_date are immutable) |
| `DELETE` | `/api/recurring/:id` | Deactivate |

Supported frequencies: `daily`, `weekly`, `biweekly`, `monthly`, `quarterly`, `yearly`.

When `auto_create` is `true`, the daily cron creates the transaction automatically. When `false`, a reminder notification is sent instead.

### Budgets

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/budget` | List |
| `GET` | `/api/budget/active` | List active |
| `GET` | `/api/budget/spending` | List with spending (`?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD`) |
| `GET` | `/api/budget/alerts` | List budgets exceeding alert threshold |
| `GET` | `/api/budget/:id` | Get |
| `GET` | `/api/budget/:id/spending` | Get with spending for date range |
| `POST` | `/api/budget` | Create. Body: `{ name, category_id, amount, period, start_date, end_date?, alert_threshold? }` |
| `PUT` | `/api/budget/:id` | Update |
| `DELETE` | `/api/budget/:id` | Deactivate |

`alert_threshold` defaults to `80` (%). Supported periods: `daily`, `weekly`, `monthly`, `quarterly`, `yearly`.

### Goals

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/goal` | List all |
| `GET` | `/api/goal/active` | List active (non-completed) |
| `GET` | `/api/goal/type/:type` | Filter by type |
| `GET` | `/api/goal/:id` | Get |
| `GET` | `/api/goal/:id/progress` | Get with progress details |
| `POST` | `/api/goal` | Create. Body: `{ name, goal_type, target_amount, target_date?, description? }` |
| `PUT` | `/api/goal/:id` | Update |
| `PATCH` | `/api/goal/:id/contribute` | Add contribution. Body: `{ amount }` |
| `PATCH` | `/api/goal/:id/complete` | Mark as completed |
| `DELETE` | `/api/goal/:id` | Delete |

Supported types: `savings`, `debt_payoff`, `investment`, `purchase`, `other`.

### Transfers

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/transfer` | List |
| `GET` | `/api/transfer/:id` | Get |
| `GET` | `/api/transfer/account/:id` | List for account |
| `POST` | `/api/transfer` | Create transfer between accounts |
| `DELETE` | `/api/transfer/:id` | Delete |

### Notifications

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/notification` | List (paginated) |
| `GET` | `/api/notification/unread` | List unread |
| `PATCH` | `/api/notification/:id/read` | Mark as read |
| `PATCH` | `/api/notification/read-all` | Mark all as read |
| `DELETE` | `/api/notification/:id` | Delete |
| `DELETE` | `/api/notification/read` | Delete all read notifications |

## Background Scheduler

A cron job runs daily at midnight (`0 0 * * *`) and performs four tasks:

1. **Recurring transactions** — finds all active recurring transactions where `next_due_date <= today`, creates the transaction and updates the account balance atomically, then advances `next_due_date` to the next period. Self-healing: missed days are processed on the next run.
2. **Budget alerts** — notifies users when spending exceeds their configured `alert_threshold` in the current calendar month.
3. **Bill reminders** — notifies users about recurring transactions with `auto_create = false` that are due within the next 3 days.
4. **Goal milestones** — notifies users when `current_amount >= target_amount` but the goal is not yet marked complete.

## File Uploads

Attachments use a two-step signed URL pattern — files never pass through the API server:

1. `GET /api/transaction/:id/attachment/upload-url?file_name=receipt.pdf&file_type=application/pdf`  
   Returns a signed `PUT` URL (15-minute expiry) and the `object_path`.
2. Client uploads directly to the signed GCS URL.
3. `POST /api/transaction/:id/attachment` with `{ file_name, file_path, file_type?, file_size? }`  
   Records the attachment in the database.

Download URLs are signed GET URLs (1-hour expiry), returned alongside every attachment response.

## Database Migrations

Migrations in `sql/schema/` are managed by goose and run automatically on startup. To run manually:

```bash
goose -dir sql/schema postgres "$DB_URL" up

# Check status
goose -dir sql/schema postgres "$DB_URL" status

# Roll back one step
goose -dir sql/schema postgres "$DB_URL" down
```

## Modifying Queries

1. Edit the relevant file in `sql/queries/`
2. Run `sqlc generate`
3. Use the generated functions from `internal/database/`

Never edit `internal/database/` directly — it is overwritten on every `sqlc generate`.

## Error Response Format

All errors return a consistent JSON envelope:

```json
{
  "status": 404,
  "error": "not found",
  "message": "account not found",
  "path": "/api/account/abc",
  "timestamp": "2026-05-12T00:00:00Z"
}
```

Internal error details are logged server-side and never sent to the client.
