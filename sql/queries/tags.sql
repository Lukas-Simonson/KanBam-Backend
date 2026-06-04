-- name: GetTagByID :one
SELECT * FROM tag
WHERE id = $1;

-- name: GetTagsByWorkspace :many
SELECT * FROM tag
WHERE workspace_id = $1
ORDER BY name ASC;

-- name: CreateTag :one
INSERT INTO tag (id, name, color, workspace_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateTag :one
UPDATE tag
SET name       = COALESCE($2, name),
    color      = COALESCE($3, color),
    updated_at = current_timestamp
WHERE id = $1
RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tag
WHERE id = $1;
