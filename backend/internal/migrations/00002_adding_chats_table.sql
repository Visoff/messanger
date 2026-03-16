-- +goose Up
CREATE TYPE "chat_type" AS ENUM (
  'private',
  'group',
  'channel'
);

CREATE TABLE "chats" (
  "id" uuid PRIMARY KEY NOT NULL DEFAULT (gen_random_uuid()),
  "title" text NOT NULL,
  "type" chat_type NOT NULL,
  "avatar_url" text,
  "metadata" json NOT NULL DEFAULT '{}',
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  "deleted_at" timestamp
);

-- +goose Down
DROP TABLE "chats";
DROP TYPE "chat_type";
