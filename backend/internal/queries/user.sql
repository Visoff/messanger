-- name: CreateUser :one
INSERT INTO users (
    username,
    password_hash
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUser :one
UPDATE users SET
    username = $2,
    avatar_url = $3,
    metadata = $4,
    updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: UpdateUserAvatar :one
UPDATE users SET
    avatar_url = $2,
    updated_at = NOW()
WHERE id = $1 RETURNING *;
