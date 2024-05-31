-- name: InsertAvatar :one
INSERT INTO avatars (bot_name, file_id, url) VALUES ($1, $2, $3) RETURNING *;

-- name: GetAvatar :one
SELECT * FROM avatars WHERE bot_name = $1 AND deleted IS NULL;

-- name: DeleteAvatar :one
UPDATE avatars SET deleted = now() WHERE bot_name = $1 AND deleted IS NULL RETURNING *;