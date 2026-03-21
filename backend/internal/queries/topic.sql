-- name: ListChatTopics :many
SELECT * from topics
where chat_id = $1;
