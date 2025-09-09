ALTER TABLE "admin_login_histories"
ALTER COLUMN "id" DROP DEFAULT;
ALTER TABLE "admin_roles"
ALTER COLUMN "id" DROP DEFAULT;
ALTER TABLE "admins"
ALTER COLUMN "id" DROP DEFAULT;
ALTER TABLE "user_identities"
ALTER COLUMN "id" DROP DEFAULT;
ALTER TABLE "user_login_histories"
ALTER COLUMN "id" DROP DEFAULT;
ALTER TABLE "users"
ALTER COLUMN "id" DROP DEFAULT;
CREATE TABLE "conversations" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "title" character varying NOT NULL,
    "metadata" jsonb NULL,
    "status" character varying NOT NULL DEFAULT 'active',
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "user_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "conversations_users_conversations" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION
);
CREATE TABLE "messages" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "role" character varying NOT NULL,
    "agent_name" character varying NULL,
    "type" character varying NOT NULL DEFAULT 'text',
    "content" text NULL,
    "media_url" character varying NULL,
    "sequence" bigint NOT NULL DEFAULT 0,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "conversation_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "messages_conversations_messages" FOREIGN KEY ("conversation_id") REFERENCES "conversations" ("id") ON DELETE NO ACTION
);
CREATE TABLE "attachments" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "type" character varying NOT NULL,
    "url" character varying NOT NULL,
    "metadata" jsonb NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "message_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "attachments_messages_attachments" FOREIGN KEY ("message_id") REFERENCES "messages" ("id") ON DELETE NO ACTION
);