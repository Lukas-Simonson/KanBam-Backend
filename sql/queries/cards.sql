-- name: GetCardByID :one
SELECT * FROM card
WHERE id = $1;

-- name: GetCardsByBoard :many
SELECT * FROM card
WHERE board_id = $1
ORDER BY position ASC;

-- name: CreateCard :one
INSERT INTO card (id, title, description, position, reference, workspace_id, board_id, column_id, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateCard :one
UPDATE card
SET title       = COALESCE($2, title),
    description = COALESCE($3, description),
    is_archived = COALESCE($4, is_archived),
    position    = COALESCE($5, position),
    column_id   = COALESCE($6, column_id),
    user_id     = COALESCE($7, user_id),
    updated_at  = current_timestamp
WHERE id = $1
RETURNING *;

-- name: DeleteCard :exec
DELETE FROM card
WHERE id = $1;

-- name: GetCardTags :many
SELECT t.* FROM tag t
JOIN card_tag ct ON ct.tag_id = t.id
WHERE ct.card_id = $1;

-- name: AddCardTag :exec
INSERT INTO card_tag (card_id, tag_id)
VALUES ($1, $2);

-- name: RemoveCardTag :exec
DELETE FROM card_tag
WHERE card_id = $1
  AND tag_id = $2;

-- name: RemoveAllCardTags :exec
DELETE FROM card_tag
WHERE card_id = $1;
