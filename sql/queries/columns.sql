-- name: GetColumnByID :one
SELECT * FROM kb_column
WHERE id = $1;

-- name: GetColumnsByBoard :many
SELECT * FROM kb_column
WHERE board_id = $1
ORDER BY position ASC;

-- name: CreateColumn :one
INSERT INTO kb_column (id, title, wip_limit, color, is_terminal, position, board_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateColumn :one
UPDATE kb_column
SET title       = COALESCE($2, title),
    wip_limit   = COALESCE($3, wip_limit),
    color       = COALESCE($4, color),
    is_terminal = COALESCE($5, is_terminal),
    position    = COALESCE($6, position),
    updated_at  = current_timestamp
WHERE id = $1
RETURNING *;

-- name: DeleteColumn :exec
DELETE FROM kb_column
WHERE id = $1;
