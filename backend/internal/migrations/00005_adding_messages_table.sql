-- +goose Up
CREATE TABLE "messages" (
  "id" uuid PRIMARY KEY NOT NULL DEFAULT (gen_random_uuid()),
  "chat_id" uuid NOT NULL,
  "topic_id" uuid,
  "sender_id" uuid NOT NULL,
  "reply_message_id" uuid,
  "content" text,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  "deleted_at" timestamp
);


-- +goose Down
DROP TABLE "messages";
