-- name: ListChatTopics :many
SELECT * from topics
where chat_id = $1;

-- name: CreateChatTopic :one
INSERT INTO topics (
    chat_id,
    title,
    type
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetTopic :one
SELECT * from topics
where id = $1;
