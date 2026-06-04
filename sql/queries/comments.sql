-- name: GetCommentByID :one
SELECT * FROM comment
WHERE id = $1;

-- name: GetCommentsByCard :many
SELECT * FROM comment
WHERE card_id = $1
ORDER BY created_at ASC;

-- name: CreateComment :one
INSERT INTO comment (id, body, card_id, user_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateComment :one
UPDATE comment
SET body       = $2,
    updated_at = current_timestamp
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comment
WHERE id = $1;
