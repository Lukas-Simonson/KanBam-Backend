-- name: GetActivityByID :one
SELECT * FROM activity
WHERE id = $1;

-- name: CreateActivity :one
INSERT INTO activity (id, workspace_id, user_id, board_id, column_id, card_id, comment_id, tag_id, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetWorkspaceActivity :many
SELECT * FROM activity
WHERE workspace_id = $1
  AND ($2::uuid    IS NULL OR board_id = $2)
  AND ($3::uuid    IS NULL OR card_id  = $3)
  AND ($4::uuid    IS NULL OR user_id  = $4)
  AND ($5::timestamptz IS NULL OR created_at > $5)
  AND ($6::timestamptz IS NULL OR created_at < $6)
ORDER BY created_at DESC;

-- name: GetBoardActivity :many
SELECT * FROM activity
WHERE workspace_id = $1
  AND board_id     = $2
  AND ($3::uuid    IS NULL OR card_id = $3)
  AND ($4::uuid    IS NULL OR user_id = $4)
  AND ($5::timestamptz IS NULL OR created_at > $5)
  AND ($6::timestamptz IS NULL OR created_at < $6)
ORDER BY created_at DESC;

-- name: GetCardActivity :many
SELECT * FROM activity
WHERE card_id    = $1
  AND ($2::uuid  IS NULL OR user_id = $2)
  AND ($3::timestamptz IS NULL OR created_at > $3)
  AND ($4::timestamptz IS NULL OR created_at < $4)
ORDER BY created_at DESC;
