-- name: ListChatTopics :many
SELECT * from topics
where chat_id = $1;

-- name: ListTopicMembers :many
SELECT users.* from users
join chat_members on chat_members.user_id = users.id
join topics on topics.chat_id = chat_members.chat_id
where topics.id = $1;

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
