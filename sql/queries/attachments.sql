-- name: CreateAttachment :one
INSERT INTO attachments (id, transaction_id, file_name, file_path, file_type, file_size)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAttachmentByID :one
SELECT * FROM attachments
WHERE id = $1;

-- name: ListAttachmentsByTransaction :many
SELECT * FROM attachments
WHERE transaction_id = $1
ORDER BY uploaded_at DESC;

-- name: DeleteAttachment :exec
DELETE FROM attachments
WHERE id = $1;

-- name: DeleteAttachmentsByTransaction :exec
DELETE FROM attachments
WHERE transaction_id = $1;
