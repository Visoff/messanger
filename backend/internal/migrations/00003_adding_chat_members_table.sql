-- +goose Up
CREATE TYPE "chat_role" AS ENUM (
  'owner',
  'admin',
  'member'
);

CREATE TABLE "chat_members" (
  "id" uuid PRIMARY KEY NOT NULL DEFAULT (gen_random_uuid()),
  "user_id" uuid NOT NULL,
  "chat_id" uuid NOT NULL,
  "role" chat_role NOT NULL DEFAULT 'member',
  "joined_at" timestamp NOT NULL DEFAULT (now()),
  "left_at" timestamp
);

-- +goose Down
DROP TABLE "chat_members";
DROP TYPE "chat_role";
