-- name: ListChats :many
SELECT * from chats;

-- name: CreateChat :one
INSERT INTO chats (
    title,
    type
) VALUES (
    $1, $2
) RETURNING *;
