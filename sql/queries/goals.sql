-- name: CreateGoal :one
INSERT INTO goals (id, user_id, account_id, name, description, target_amount, currency_code, target_date, goal_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetGoalByID :one
SELECT g.*, a.name AS account_name
FROM goals g
LEFT JOIN accounts a ON g.account_id = a.id
WHERE g.id = $1 AND g.user_id = $2;

-- name: ListGoalsByUser :many
SELECT g.*, a.name AS account_name
FROM goals g
LEFT JOIN accounts a ON g.account_id = a.id
WHERE g.user_id = $1
ORDER BY g.target_date ASC NULLS LAST, g.created_at DESC;

-- name: ListActiveGoalsByUser :many
SELECT g.*, a.name AS account_name
FROM goals g
LEFT JOIN accounts a ON g.account_id = a.id
WHERE g.user_id = $1 AND g.is_completed = false
ORDER BY g.target_date ASC NULLS LAST;

-- name: ListGoalsByType :many
SELECT g.*, a.name AS account_name
FROM goals g
LEFT JOIN accounts a ON g.account_id = a.id
WHERE g.user_id = $1 AND g.goal_type = $2
ORDER BY g.target_date ASC NULLS LAST;

-- name: UpdateGoal :one
UPDATE goals
SET name = $2,
    description = $3,
    target_amount = $4,
    target_date = $5,
    goal_type = $6,
    account_id = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $8
RETURNING *;

-- name: UpdateGoalCurrentAmount :one
UPDATE goals
SET current_amount = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $3
RETURNING *;

-- name: CompleteGoal :one
UPDATE goals
SET is_completed = true,
    completed_date = CURRENT_DATE,
    current_amount = target_amount,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteGoal :exec
DELETE FROM goals
WHERE id = $1 AND user_id = $2;

-- name: ListGoalsReachedTarget :many
SELECT id, user_id, name, target_amount, current_amount
FROM goals
WHERE is_completed = false
  AND current_amount >= target_amount;

-- name: GetGoalProgress :one
SELECT
    id,
    name,
    target_amount,
    current_amount,
    CASE WHEN target_amount > 0
        THEN ROUND((current_amount / target_amount) * 100, 2)
        ELSE 0
    END AS progress_percentage,
    target_amount - current_amount AS remaining_amount,
    target_date,
    is_completed
FROM goals
WHERE id = $1 AND user_id = $2;
