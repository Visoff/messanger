-- name: CreateWebPushSubscription :exec
INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
VALUES ($1, $2, $3, $4);

-- name: GetAllSubscriptions :many
SELECT * FROM push_subscriptions;
