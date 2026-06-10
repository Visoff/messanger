-- name: CreateChatInvitation :one
INSERT INTO invitation (
    user_id,
    chat_id
) VALUES (
    $1, $2
) RETURNING id;

-- name: GetInvitationByUserAndChat :one
SELECT * FROM invitation
WHERE user_id = $1 AND chat_id = $2 AND deleted_at IS NULL;

-- name: GetInvitationById :one
SELECT * FROM invitation
WHERE id = $1 AND deleted_at IS NULL;

-- name: UseInvitation :exec
UPDATE invitation SET deleted_at = NOW()
WHERE id = $1;
