-- name: CreateUser :one
INSERT INTO users (id, firebase_uid, currency_code)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByFirebaseUID :one
SELECT * FROM users
WHERE firebase_uid = $1;

-- name: UpdateUserCurrency :one
UPDATE users
SET currency_code = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
