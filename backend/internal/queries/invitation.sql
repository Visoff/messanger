-- name: CreateChatInvitation :one
INSERT INTO invitation (
    user_id,
    chat_id
) VALUES (
    $1, $2
) RETURNING id;
