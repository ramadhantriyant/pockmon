-- name: CreateBudget :one
INSERT INTO budgets (id, user_id, category_id, name, amount, period, start_date, end_date, alert_threshold)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetBudgetByID :one
SELECT b.*, c.name AS category_name, c.color, c.icon
FROM budgets b
JOIN categories c ON b.category_id = c.id
WHERE b.id = $1 AND b.user_id = $2;

-- name: ListBudgetsByUser :many
SELECT b.*, c.name AS category_name, c.color, c.icon
FROM budgets b
JOIN categories c ON b.category_id = c.id
WHERE b.user_id = $1
ORDER BY b.name;

-- name: ListActiveBudgetsByUser :many
SELECT b.*, c.name AS category_name, c.color, c.icon
FROM budgets b
JOIN categories c ON b.category_id = c.id
WHERE b.user_id = $1
  AND b.is_active = true
  AND (b.end_date IS NULL OR b.end_date >= CURRENT_DATE)
ORDER BY b.name;

-- name: GetBudgetWithSpending :one
SELECT
    b.*,
    c.name AS category_name,
    c.color,
    c.icon,
    COALESCE(SUM(t.amount), 0) AS spent_amount,
    b.amount - COALESCE(SUM(t.amount), 0) AS remaining_amount,
    CASE WHEN b.amount > 0
        THEN ROUND((COALESCE(SUM(t.amount), 0) / b.amount) * 100, 2)
        ELSE 0
    END AS spent_percentage
FROM budgets b
JOIN categories c ON b.category_id = c.id
LEFT JOIN transactions t ON t.category_id = b.category_id
    AND t.user_id = b.user_id
    AND t.type = 'expense'
    AND t.transaction_date BETWEEN $2 AND $3
WHERE b.id = $1 AND b.user_id = $4
GROUP BY b.id, c.name, c.color, c.icon;

-- name: ListBudgetsWithSpending :many
SELECT
    b.*,
    c.name AS category_name,
    c.color,
    c.icon,
    COALESCE(SUM(t.amount), 0) AS spent_amount,
    b.amount - COALESCE(SUM(t.amount), 0) AS remaining_amount,
    CASE WHEN b.amount > 0
        THEN ROUND((COALESCE(SUM(t.amount), 0) / b.amount) * 100, 2)
        ELSE 0
    END AS spent_percentage
FROM budgets b
JOIN categories c ON b.category_id = c.id
LEFT JOIN transactions t ON t.category_id = b.category_id
    AND t.user_id = b.user_id
    AND t.type = 'expense'
    AND t.transaction_date BETWEEN $2 AND $3
WHERE b.user_id = $1 AND b.is_active = true
GROUP BY b.id, c.name, c.color, c.icon
ORDER BY spent_percentage DESC;

-- name: ListBudgetsExceedingThreshold :many
SELECT
    b.*,
    c.name AS category_name,
    COALESCE(SUM(t.amount), 0) AS spent_amount,
    CASE WHEN b.amount > 0
        THEN ROUND((COALESCE(SUM(t.amount), 0) / b.amount) * 100, 2)
        ELSE 0
    END AS spent_percentage
FROM budgets b
JOIN categories c ON b.category_id = c.id
LEFT JOIN transactions t ON t.category_id = b.category_id
    AND t.user_id = b.user_id
    AND t.type = 'expense'
    AND t.transaction_date BETWEEN $2 AND $3
WHERE b.user_id = $1 AND b.is_active = true
GROUP BY b.id, c.name
HAVING CASE WHEN b.amount > 0
    THEN (COALESCE(SUM(t.amount), 0) / b.amount) * 100
    ELSE 0
END >= b.alert_threshold;

-- name: UpdateBudget :one
UPDATE budgets
SET name = $2,
    amount = $3,
    period = $4,
    start_date = $5,
    end_date = $6,
    alert_threshold = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $8
RETURNING *;

-- name: DeactivateBudget :exec
UPDATE budgets
SET is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2;

-- name: DeleteBudget :exec
DELETE FROM budgets
WHERE id = $1 AND user_id = $2;
