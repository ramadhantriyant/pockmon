-- name: CreateTransfer :one
INSERT INTO transfers (
    id, user_id, from_account_id, to_account_id,
    from_transaction_id, to_transaction_id, amount, to_amount, exchange_rate,
    transfer_date, description
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetTransferByID :one
SELECT
    tr.*,
    fa.name AS from_account_name,
    ta.name AS to_account_name
FROM transfers tr
JOIN accounts fa ON tr.from_account_id = fa.id
JOIN accounts ta ON tr.to_account_id = ta.id
WHERE tr.id = $1 AND tr.user_id = $2;

-- name: ListTransfersByUser :many
SELECT
    tr.*,
    fa.name AS from_account_name,
    ta.name AS to_account_name
FROM transfers tr
JOIN accounts fa ON tr.from_account_id = fa.id
JOIN accounts ta ON tr.to_account_id = ta.id
WHERE tr.user_id = $1
ORDER BY tr.transfer_date DESC, tr.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTransfersByAccount :many
SELECT
    tr.*,
    fa.name AS from_account_name,
    ta.name AS to_account_name
FROM transfers tr
JOIN accounts fa ON tr.from_account_id = fa.id
JOIN accounts ta ON tr.to_account_id = ta.id
WHERE tr.user_id = $1
  AND (tr.from_account_id = $2 OR tr.to_account_id = $2)
ORDER BY tr.transfer_date DESC
LIMIT $3 OFFSET $4;

-- name: ListTransfersByDateRange :many
SELECT
    tr.*,
    fa.name AS from_account_name,
    ta.name AS to_account_name
FROM transfers tr
JOIN accounts fa ON tr.from_account_id = fa.id
JOIN accounts ta ON tr.to_account_id = ta.id
WHERE tr.user_id = $1
  AND tr.transfer_date BETWEEN $2 AND $3
ORDER BY tr.transfer_date DESC;

-- name: DeleteTransfer :exec
DELETE FROM transfers
WHERE id = $1 AND user_id = $2;
