-- +goose Up
CREATE TYPE "topic_type" AS ENUM (
  'text_topic',
  'voice_topic'
);

CREATE TABLE "topics" (
  "id" uuid PRIMARY KEY NOT NULL DEFAULT (gen_random_uuid()),
  "chat_id" uuid NOT NULL,
  "title" text NOT NULL,
  "avatar_url" text,
  "type" topic_type NOT NULL,
  "creted_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  "deleted_at" timestamp
);


-- +goose Down
DROP TABLE "topics";
DROP TYPE "topic_type";
