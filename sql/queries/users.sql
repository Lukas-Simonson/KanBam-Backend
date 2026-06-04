-- name: GetUserByEmail :one
SELECT * FROM kb_user
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM kb_user
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO kb_user (id, name, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE kb_user
SET password_hash = $2,
    updated_at = current_timestamp
WHERE id = $1;
