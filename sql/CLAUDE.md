# SQL Layer

## Workflow

1. Edit schema files in `sql/schema/` (migrations) or query files in `sql/queries/`.
2. Run `sqlc generate` from the project root to regenerate `internal/database/`.
3. Never edit anything inside `internal/database/` — it is fully generated.

## Migrations (`sql/schema/`)

Managed by **goose**. Files are prefixed with a sequence number and must never be renamed or reordered. To add a change, create a new file:

```
NNN_description.sql
```

Each file must contain goose directives:
```sql
-- +goose Up
ALTER TABLE ...;

-- +goose Down
ALTER TABLE ...;
```

Migrations run automatically when the server starts via `goose.Up()`.

## Queries (`sql/queries/`)

One file per domain entity. Each query must have a sqlc annotation:

```sql
-- name: QueryName :one    -- returns a single row
-- name: QueryName :many   -- returns a slice
-- name: QueryName :exec   -- no rows returned
```

After editing, regenerate:
```bash
sqlc generate
```

## sqlc Config (`sqlc.yaml`)

```yaml
engine: postgresql
sql_package: pgx/v5
emit_json_tags: true
emit_interface: true      # generates Querier interface in querier.go
emit_empty_slices: true
emit_pointers_for_null_types: true
```

`emit_pointers_for_null_types: true` means nullable columns become pointer types (`*string`, `*bool`) in the generated Go structs.

## pgtype Conventions

All primary keys are `pgtype.UUID`, monetary values are `pgtype.Numeric`, and dates are `pgtype.Date`. Use the helpers in `internal/util/convert.go` — never construct pgtype values by hand in handlers.

A zero-value `pgtype.UUID{}` (Valid: false) is stored as NULL, which is how optional FK columns (e.g., `parent_category_id`) are cleared.
