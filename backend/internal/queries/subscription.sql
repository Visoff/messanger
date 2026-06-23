-- name: CreateWebPushSubscription :exec
INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
VALUES ($1, $2, $3, $4);

-- name: UpdateWebPushSubscription :exec
UPDATE push_subscriptions
SET user_id = $1, p256dh = $2, auth = $3
WHERE id = $4;

-- name: GetSubscriptionByEndpoint :one
SELECT * FROM push_subscriptions
WHERE endpoint = $1 LIMIT 1;

-- name: GetAllSubscriptions :many
SELECT * FROM push_subscriptions;

-- name: GetUserSubscriptions :many
SELECT * FROM push_subscriptions
WHERE user_id = $1;
