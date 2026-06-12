-- name: ListChats :many
SELECT chats.* from chats
left join chat_members on chat_members.chat_id = chats.id
where chat_members.user_id = $1 AND chat_members.left_at IS NULL;

-- name: ListChatMembers :many
SELECT users.* from users
join chat_members on chat_members.user_id = users.id
where chat_members.chat_id = $1 AND chat_members.left_at IS NULL;

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
) ON CONFLICT (user_id, chat_id) DO UPDATE SET
    left_at = NULL,
    role = EXCLUDED.role;

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
UPDATE chat_members SET left_at = NOW()
WHERE user_id = $1 AND chat_id = $2;

-- name: CheckPrivateChatExists :one
SELECT chats.id FROM chats
JOIN chat_members cm1 ON cm1.chat_id = chats.id
JOIN chat_members cm2 ON cm2.chat_id = chats.id
WHERE chats.type = 'private'
  AND cm1.user_id = $1 AND cm1.left_at IS NULL
  AND cm2.user_id = $2 AND cm2.left_at IS NULL
LIMIT 1;

-- name: ListUserChats :many
SELECT c.* FROM chats c
JOIN chat_members cm ON cm.chat_id = c.id
WHERE cm.user_id = $1 AND cm.left_at IS NULL
ORDER BY c.updated_at DESC;

-- name: GetChatLastMessage :one
SELECT * FROM messages
WHERE chat_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateChatAvatar :one
UPDATE chats SET
    avatar_url = $2,
    updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: ListChatsLastMessages :many
SELECT DISTINCT ON (messages.chat_id) * FROM messages
WHERE messages.chat_id = ANY($1::uuid[])
  AND messages.deleted_at IS NULL
ORDER BY messages.chat_id, messages.created_at DESC;

-- name: UpdateChatUpdatedAt :exec
UPDATE chats SET updated_at = NOW() WHERE id = $1;

-- name: GetUserChatRole :one
SELECT role FROM chat_members WHERE user_id = $1 AND chat_id = $2 AND left_at IS NULL;

-- name: ListChatMembersWithRoles :many
SELECT users.*, chat_members.role FROM users
JOIN chat_members ON chat_members.user_id = users.id
WHERE chat_members.chat_id = $1 AND chat_members.left_at IS NULL;

-- name: UpdateChatMuted :one
UPDATE chats SET
    metadata = $2,
    updated_at = NOW()
WHERE id = $1 RETURNING *;
