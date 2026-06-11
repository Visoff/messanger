-- name: ListUserCategories :many
SELECT * FROM user_categories WHERE user_id = $1 ORDER BY created_at;

-- name: CreateUserCategory :one
INSERT INTO user_categories (user_id, name, chat_ids) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateUserCategory :one
UPDATE user_categories SET name = $2, chat_ids = $3, updated_at = NOW() WHERE id = $1 AND user_id = $4 RETURNING *;

-- name: DeleteUserCategory :exec
DELETE FROM user_categories WHERE id = $1 AND user_id = $2;
