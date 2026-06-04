-- name: GetWorkspaceByID :one
SELECT * FROM workspace
WHERE id = $1;

-- name: GetWorkspacesByUser :many
SELECT w.* FROM workspace w
JOIN workspace_user wu ON wu.workspace_id = w.id
WHERE wu.user_id = $1
  AND wu.pending = false
UNION
SELECT * FROM workspace
WHERE owner_id = $1;

-- name: CreateWorkspace :one
INSERT INTO workspace (id, title, description, owner_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateWorkspace :one
UPDATE workspace
SET title       = COALESCE($2, title),
    description = COALESCE($3, description),
    updated_at  = current_timestamp
WHERE id = $1
RETURNING *;

-- name: DeleteWorkspace :exec
DELETE FROM workspace
WHERE id = $1;
