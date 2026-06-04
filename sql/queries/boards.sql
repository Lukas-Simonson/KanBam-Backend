-- name: GetBoardByID :one
SELECT * FROM board
WHERE id = $1;

-- name: GetBoardsByWorkspace :many
SELECT * FROM board
WHERE workspace_id = $1
ORDER BY created_at ASC;

-- name: CreateBoard :one
INSERT INTO board (id, title, description, prefix, workspace_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateBoard :one
UPDATE board
SET title       = COALESCE($2, title),
    description = COALESCE($3, description),
    is_archived = COALESCE($4, is_archived),
    updated_at  = current_timestamp
WHERE id = $1
RETURNING *;

-- name: IncrementBoardSequence :one
UPDATE board
SET sequence = sequence + 1
WHERE id = $1
RETURNING sequence;

-- name: DeleteBoard :exec
DELETE FROM board
WHERE id = $1;
