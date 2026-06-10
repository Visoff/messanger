-- name: ListTopicMessages :many
SELECT * from messages
where topic_id = $1;

-- name: ListChatMessages :many
SELECT * from messages
where chat_id = $1;

-- name: CreateTopicMessage :one
INSERT INTO messages (
    topic_id,
    sender_id,
    content,
    reply_message_id,
    chat_id
) VALUES (
    $1, $2, $3, $4, (select chat_id from topics where topics.id = $1)
) RETURNING *;

-- name: CreateChatMessage :one
INSERT INTO messages (
    chat_id,
    sender_id,
    content,
    reply_message_id
) VALUES (
    $1, $2, $3, $4
) RETURNING *;
