-- name: CreateTransaction :one
INSERT INTO transactions (
    id, user_id, account_id, category_id, type,
    amount, currency_code, transaction_date, description, notes,
    payee, location, tags, is_recurring, recurring_transaction_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: GetTransactionByID :one
SELECT t.*, c.name AS category_name, a.name AS account_name
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
LEFT JOIN accounts a ON t.account_id = a.id
WHERE t.id = $1 AND t.user_id = $2;

-- name: ListTransactionsByUser :many
SELECT t.*, c.name AS category_name, a.name AS account_name
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
LEFT JOIN accounts a ON t.account_id = a.id
WHERE t.user_id = $1
ORDER BY t.transaction_date DESC, t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTransactionsByAccount :many
SELECT t.*, c.name AS category_name
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.user_id = $1 AND t.account_id = $2
ORDER BY t.transaction_date DESC, t.created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListTransactionsByCategory :many
SELECT t.*, a.name AS account_name
FROM transactions t
LEFT JOIN accounts a ON t.account_id = a.id
WHERE t.user_id = $1 AND t.category_id = $2
ORDER BY t.transaction_date DESC
LIMIT $3 OFFSET $4;

-- name: ListTransactionsByDateRange :many
SELECT t.*, c.name AS category_name, a.name AS account_name
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
LEFT JOIN accounts a ON t.account_id = a.id
WHERE t.user_id = $1
  AND t.transaction_date BETWEEN $2 AND $3
ORDER BY t.transaction_date DESC, t.created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListTransactionsByType :many
SELECT t.*, c.name AS category_name, a.name AS account_name
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
LEFT JOIN accounts a ON t.account_id = a.id
WHERE t.user_id = $1 AND t.type = $2
ORDER BY t.transaction_date DESC
LIMIT $3 OFFSET $4;

-- name: ListTransactionsByTag :many
SELECT t.*, c.name AS category_name, a.name AS account_name
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
LEFT JOIN accounts a ON t.account_id = a.id
WHERE t.user_id = $1 AND $2 = ANY(t.tags)
ORDER BY t.transaction_date DESC
LIMIT $3 OFFSET $4;

-- name: UpdateTransaction :one
UPDATE transactions
SET category_id = $2,
    amount = $3,
    transaction_date = $4,
    description = $5,
    notes = $6,
    payee = $7,
    location = $8,
    tags = $9,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $10
RETURNING *;

-- name: DeleteTransaction :exec
DELETE FROM transactions
WHERE id = $1 AND user_id = $2;

-- name: GetMonthlySpendingByCategory :many
SELECT
    c.id AS category_id,
    c.name AS category_name,
    c.color,
    c.icon,
    SUM(t.amount) AS total_amount,
    COUNT(t.id) AS transaction_count
FROM transactions t
JOIN categories c ON t.category_id = c.id
WHERE t.user_id = $1
  AND t.type = 'expense'
  AND DATE_TRUNC('month', t.transaction_date) = DATE_TRUNC('month', $2::DATE)
GROUP BY c.id, c.name, c.color, c.icon
ORDER BY total_amount DESC;

-- name: GetIncomeVsExpenseByMonth :many
SELECT
    DATE_TRUNC('month', transaction_date) AS month,
    SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) AS total_income,
    SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) AS total_expense
FROM transactions
WHERE user_id = $1
  AND transaction_date BETWEEN $2 AND $3
GROUP BY DATE_TRUNC('month', transaction_date)
ORDER BY month;

-- name: CountTransactionsByUser :one
SELECT COUNT(*) FROM transactions
WHERE user_id = $1;

-- name: GetTransactionSummaryByDateRange :one
SELECT
    COUNT(*) AS transaction_count,
    SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) AS total_income,
    SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) AS total_expense
FROM transactions
WHERE user_id = $1
  AND transaction_date BETWEEN $2 AND $3;
