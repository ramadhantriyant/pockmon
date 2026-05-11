-- name: CreateCategory :one
INSERT INTO categories (id, user_id, name, type, color, icon, parent_category_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories
WHERE id = $1;

-- name: GetCategoryByIDAndUser :one
SELECT * FROM categories
WHERE id = $1 AND user_id = $2;

-- name: ListCategoriesByUser :many
SELECT * FROM categories
WHERE user_id = $1 OR is_system = true
ORDER BY type, name;

-- name: ListCategoriesByType :many
SELECT * FROM categories
WHERE (user_id = $1 OR is_system = true) AND type = $2
ORDER BY name;

-- name: ListSystemCategories :many
SELECT * FROM categories
WHERE is_system = true
ORDER BY type, name;

-- name: ListSubcategories :many
SELECT * FROM categories
WHERE parent_category_id = $1
ORDER BY name;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2,
    color = $3,
    icon = $4,
    parent_category_id = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $6
RETURNING *;

-- name: SetSystemCategory :exec
UPDATE categories
SET is_system = true
WHERE id = $1 AND user_id = $2;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = $1 AND user_id = $2;
