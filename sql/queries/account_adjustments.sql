-- name: CreateAccountAdjustment :one
INSERT INTO account_adjustments (account_id, user_id, amount, previous_balance, new_balance, reason, adjustment_date)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAccountAdjustmentByID :one
SELECT adj.*, a.name AS account_name
FROM account_adjustments adj
JOIN accounts a ON adj.account_id = a.id
WHERE adj.id = $1 AND adj.user_id = $2;

-- name: ListAccountAdjustmentsByAccount :many
SELECT adj.*, a.name AS account_name
FROM account_adjustments adj
JOIN accounts a ON adj.account_id = a.id
WHERE adj.account_id = $1 AND adj.user_id = $2
ORDER BY adj.adjustment_date DESC, adj.created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListAccountAdjustmentsByUser :many
SELECT adj.*, a.name AS account_name
FROM account_adjustments adj
JOIN accounts a ON adj.account_id = a.id
WHERE adj.user_id = $1
ORDER BY adj.adjustment_date DESC, adj.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAccountAdjustmentsByDateRange :many
SELECT adj.*, a.name AS account_name
FROM account_adjustments adj
JOIN accounts a ON adj.account_id = a.id
WHERE adj.user_id = $1
  AND adj.adjustment_date BETWEEN $2 AND $3
ORDER BY adj.adjustment_date DESC;

-- name: GetLatestAdjustmentByAccount :one
SELECT * FROM account_adjustments
WHERE account_id = $1
ORDER BY adjustment_date DESC, created_at DESC
LIMIT 1;
