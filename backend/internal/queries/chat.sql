-- name: ListChats :many
SELECT chats.* from chats
left join chat_members on chat_members.chat_id = chats.id
where chat_members.user_id = $1;

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
