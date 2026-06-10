-- +goose Up
DELETE FROM "chat_members"
WHERE "id" IN (
    SELECT "id"
    FROM (
        SELECT "id", ROW_NUMBER() OVER (PARTITION BY "user_id", "chat_id" ORDER BY "joined_at") AS rn
        FROM "chat_members"
    ) AS dups
    WHERE rn > 1
);

ALTER TABLE "chat_members" ADD CONSTRAINT "uq_chat_members_user_chat" UNIQUE ("user_id", "chat_id");

-- +goose Down
ALTER TABLE "chat_members" DROP CONSTRAINT "uq_chat_members_user_chat";
