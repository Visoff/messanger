-- +goose Up
CREATE TABLE "push_subscriptions" (
  "id" uuid PRIMARY KEY NOT NULL DEFAULT (gen_random_uuid()),
  "user_id" uuid NOT NULL,
  "endpoint" text NOT NULL,
  "p256dh" text NOT NULL,
  "auth" text NOT NULL
);


-- +goose Down
DROP TABLE "push_subscriptions";
