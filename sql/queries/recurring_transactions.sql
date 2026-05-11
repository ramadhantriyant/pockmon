-- name: CreateRecurringTransaction :one
INSERT INTO recurring_transactions (
    id, user_id, account_id, category_id, type,
    amount, currency_code, description, frequency,
    start_date, end_date, next_due_date, auto_create
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetRecurringTransactionByID :one
SELECT r.*, c.name AS category_name, a.name AS account_name
FROM recurring_transactions r
LEFT JOIN categories c ON r.category_id = c.id
LEFT JOIN accounts a ON r.account_id = a.id
WHERE r.id = $1 AND r.user_id = $2;

-- name: ListRecurringTransactionsByUser :many
SELECT r.*, c.name AS category_name, a.name AS account_name
FROM recurring_transactions r
LEFT JOIN categories c ON r.category_id = c.id
LEFT JOIN accounts a ON r.account_id = a.id
WHERE r.user_id = $1
ORDER BY r.next_due_date ASC;

-- name: ListActiveRecurringTransactionsByUser :many
SELECT r.*, c.name AS category_name, a.name AS account_name
FROM recurring_transactions r
LEFT JOIN categories c ON r.category_id = c.id
LEFT JOIN accounts a ON r.account_id = a.id
WHERE r.user_id = $1
  AND r.is_active = true
  AND (r.end_date IS NULL OR r.end_date >= CURRENT_DATE)
ORDER BY r.next_due_date ASC;

-- name: ListDueRecurringTransactions :many
SELECT r.*, c.name AS category_name, a.name AS account_name
FROM recurring_transactions r
LEFT JOIN categories c ON r.category_id = c.id
LEFT JOIN accounts a ON r.account_id = a.id
WHERE r.is_active = true
  AND r.next_due_date <= $1
  AND (r.end_date IS NULL OR r.end_date >= CURRENT_DATE)
ORDER BY r.next_due_date ASC;

-- name: ListAutoCreateDueRecurringTransactions :many
SELECT r.*, c.name AS category_name, a.name AS account_name
FROM recurring_transactions r
LEFT JOIN categories c ON r.category_id = c.id
LEFT JOIN accounts a ON r.account_id = a.id
WHERE r.is_active = true
  AND r.auto_create = true
  AND r.next_due_date <= $1
  AND (r.end_date IS NULL OR r.end_date >= CURRENT_DATE)
ORDER BY r.next_due_date ASC;

-- name: UpdateRecurringTransaction :one
UPDATE recurring_transactions
SET account_id = $2,
    category_id = $3,
    amount = $4,
    description = $5,
    frequency = $6,
    end_date = $7,
    auto_create = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $9
RETURNING *;

-- name: UpdateNextDueDate :one
UPDATE recurring_transactions
SET next_due_date = $2,
    last_processed_date = CURRENT_DATE,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeactivateRecurringTransaction :exec
UPDATE recurring_transactions
SET is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2;

-- name: DeleteRecurringTransaction :exec
DELETE FROM recurring_transactions
WHERE id = $1 AND user_id = $2;
