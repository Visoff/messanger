-- name: ListChats :many
SELECT chats.* from chats
left join chat_members on chat_members.chat_id = chats.id
where chat_members.user_id = $1;

-- name: ListChatMembers :many
SELECT users.* from users
join chat_members on chat_members.user_id = users.id
where chat_members.chat_id = $1;

-- name: GetChat :one
SELECT * FROM chats
WHERE id = $1;

-- name: CreateChat :one
INSERT INTO chats (
    title,
    type
) VALUES (
    $1, $2
) RETURNING *;

-- name: AddUserToChat :exec
INSERT INTO chat_members (
    user_id,
    chat_id,
    role
) VALUES (
    $1, $2, $3
);

-- name: JoinUserToChat :exec
INSERT INTO chat_members (
    user_id,
    chat_id
) VALUES (
    $1, $2
);

-- name: UpdateChat :one
UPDATE chats SET
    title = $2,
    metadata = $3,
    updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: RemoveUserFromChat :exec
DELETE FROM chat_members
WHERE user_id = $1 AND chat_id = $2;

-- name: CheckPrivateChatExists :one
SELECT chats.id FROM chats
JOIN chat_members cm1 ON cm1.chat_id = chats.id
JOIN chat_members cm2 ON cm2.chat_id = chats.id
WHERE chats.type = 'private'
  AND cm1.user_id = $1
  AND cm2.user_id = $2
LIMIT 1;

-- name: ListChatsWithLastMessage :many
SELECT chats.id, chats.title, chats.type, chats.avatar_url, chats.metadata, chats.created_at, chats.updated_at, chats.deleted_at,
       m.id as m_id, m.chat_id as m_chat_id, m.topic_id as m_topic_id, m.sender_id as m_sender_id,
       m.reply_message_id as m_reply_message_id, m.content as m_content,
       m.created_at as m_created_at, m.updated_at as m_updated_at, m.deleted_at as m_deleted_at
FROM chats
LEFT JOIN chat_members ON chat_members.chat_id = chats.id
LEFT JOIN LATERAL (
    SELECT * FROM messages
    WHERE messages.chat_id = chats.id AND messages.deleted_at IS NULL
    ORDER BY messages.created_at DESC
    LIMIT 1
) m ON true
WHERE chat_members.user_id = $1;

-- name: UpdateChatMuted :one
UPDATE chats SET
    metadata = $2,
    updated_at = NOW()
WHERE id = $1 RETURNING *;
