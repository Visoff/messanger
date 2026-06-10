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

-- name: UpdateChatMuted :one
UPDATE chats SET
    metadata = $2,
    updated_at = NOW()
WHERE id = $1 RETURNING *;
