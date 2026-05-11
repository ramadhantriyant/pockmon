-- name: CreateAccount :one
INSERT INTO accounts (id, user_id, name, type, currency_code, initial_balance, current_balance, include_in_total, color, icon, notes)
VALUES ($1, $2, $3, $4, $5, $6, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetAccountByID :one
SELECT * FROM accounts
WHERE id = $1 AND user_id = $2;

-- name: ListAccountsByUser :many
SELECT * FROM accounts
WHERE user_id = $1
ORDER BY name;

-- name: ListActiveAccountsByUser :many
SELECT * FROM accounts
WHERE user_id = $1 AND is_active = true
ORDER BY name;

-- name: UpdateAccount :one
UPDATE accounts
SET name = $2,
    type = $3,
    currency_code = $4,
    include_in_total = $5,
    color = $6,
    icon = $7,
    notes = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $9
RETURNING *;

-- name: UpdateAccountBalance :one
UPDATE accounts
SET current_balance = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeactivateAccount :exec
UPDATE accounts
SET is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2;

-- name: GetTotalBalanceByUser :one
SELECT COALESCE(SUM(current_balance), 0) AS total_balance
FROM accounts
WHERE user_id = $1 AND is_active = true AND include_in_total = true;

-- name: GetAccountsByType :many
SELECT * FROM accounts
WHERE user_id = $1 AND type = $2 AND is_active = true
ORDER BY name;
