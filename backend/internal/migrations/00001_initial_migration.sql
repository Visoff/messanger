-- +goose Up
CREATE TABLE "users" (
  "id" uuid PRIMARY KEY NOT NULL DEFAULT (gen_random_uuid()),
  "username" text NOT NULL UNIQUE,
  "password_hash" text NOT NULL,
  "avatar_url" text,
  "metadata" json NOT NULL DEFAULT '{}',
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT 'now()',
  "deleted_at" timestamp,
  "last_seen_at" timestamp NOT NULL DEFAULT 'now()'
);

-- +goose Down
DROP TABLE "users";
