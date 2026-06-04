-- name: GetWorkspaceMembers :many
SELECT u.id, u.name, u.email, u.created_at, wu.role
FROM workspace_user wu
JOIN kb_user u ON u.id = wu.user_id
WHERE wu.workspace_id = $1
  AND wu.pending = false;

-- name: GetWorkspaceMember :one
SELECT u.id, u.name, u.email, u.created_at, wu.role
FROM workspace_user wu
JOIN kb_user u ON u.id = wu.user_id
WHERE wu.workspace_id = $1
  AND wu.user_id = $2;

-- name: GetWorkspaceMembership :one
SELECT * FROM workspace_user
WHERE workspace_id = $1
  AND user_id = $2;

-- name: AddWorkspaceMember :exec
INSERT INTO workspace_user (user_id, workspace_id, role, pending)
VALUES ($1, $2, $3, false);

-- name: UpdateWorkspaceMemberRole :exec
UPDATE workspace_user
SET role = $3
WHERE workspace_id = $1
  AND user_id = $2;

-- name: RemoveWorkspaceMember :exec
DELETE FROM workspace_user
WHERE workspace_id = $1
  AND user_id = $2;
