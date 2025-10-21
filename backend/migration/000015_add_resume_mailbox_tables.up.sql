-- Migration: 000015_add_resume_mailbox_tables
-- Created: 2024-01-01
-- Description: Add resume mailbox collection tables for email-based resume collection feature
-- Create resume_mailbox_settings table
CREATE TABLE "resume_mailbox_settings" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "name" varchar(256) NOT NULL,
    "email_address" varchar(256) NOT NULL,
    "protocol" varchar(10) NOT NULL CHECK ("protocol" IN ('imap', 'pop3')),
    "host" varchar(256) NOT NULL,
    "port" integer NOT NULL,
    "use_ssl" boolean NOT NULL DEFAULT true,
    "folder" varchar(256),
    "auth_type" varchar(20) NOT NULL DEFAULT 'password' CHECK ("auth_type" IN ('password', 'oauth')),
    "encrypted_credential" jsonb NOT NULL,
    "uploader_id" uuid NOT NULL,
    "job_profile_id" uuid,
    "sync_interval_minutes" integer,
    "status" varchar(20) NOT NULL DEFAULT 'enabled' CHECK ("status" IN ('enabled', 'disabled')),
    "last_synced_at" timestamp with time zone,
    "last_error" text,
    "retry_count" integer NOT NULL DEFAULT 0,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL
);
-- Create resume_mailbox_cursors table
CREATE TABLE "resume_mailbox_cursors" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "mailbox_id" uuid NOT NULL,
    "protocol_cursor" text NOT NULL,
    "last_message_id" varchar(512),
    "deleted_at" timestamptz NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now()
);
-- Create resume_mailbox_statistics table
CREATE TABLE "resume_mailbox_statistics" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "mailbox_id" uuid NOT NULL,
    "date" timestamp with time zone NOT NULL,
    "synced_emails" integer NOT NULL DEFAULT 0,
    "parsed_resumes" integer NOT NULL DEFAULT 0,
    "failed_resumes" integer NOT NULL DEFAULT 0,
    "skipped_attachments" integer NOT NULL DEFAULT 0,
    "last_sync_duration_ms" integer NOT NULL DEFAULT 0,
    "deleted_at" timestamptz NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now()
);
-- Add foreign key constraints
ALTER TABLE "resume_mailbox_settings"
ADD CONSTRAINT "fk_resume_mailbox_settings_uploader" FOREIGN KEY ("uploader_id") REFERENCES "users"("id") ON DELETE CASCADE;
ALTER TABLE "resume_mailbox_settings"
ADD CONSTRAINT "fk_resume_mailbox_settings_job_profile" FOREIGN KEY ("job_profile_id") REFERENCES "job_position"("id") ON DELETE
SET NULL;
ALTER TABLE "resume_mailbox_cursors"
ADD CONSTRAINT "fk_resume_mailbox_cursors_mailbox" FOREIGN KEY ("mailbox_id") REFERENCES "resume_mailbox_settings"("id") ON DELETE CASCADE;
ALTER TABLE "resume_mailbox_statistics"
ADD CONSTRAINT "fk_resume_mailbox_statistics_mailbox" FOREIGN KEY ("mailbox_id") REFERENCES "resume_mailbox_settings"("id") ON DELETE CASCADE;
-- Create indexes for resume_mailbox_settings
CREATE UNIQUE INDEX "idx_resume_mailbox_settings_protocol_email" ON "resume_mailbox_settings"("protocol", "email_address")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_resume_mailbox_settings_status" ON "resume_mailbox_settings"("status")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_resume_mailbox_settings_uploader_id" ON "resume_mailbox_settings"("uploader_id")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_resume_mailbox_settings_job_profile_id" ON "resume_mailbox_settings"("job_profile_id")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_resume_mailbox_settings_last_synced_at" ON "resume_mailbox_settings"("last_synced_at")
WHERE "deleted_at" IS NULL;
-- Create indexes for resume_mailbox_cursors
CREATE INDEX "idx_resume_mailbox_cursors_mailbox_id" ON "resume_mailbox_cursors"("mailbox_id")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_resume_mailbox_cursors_updated_at" ON "resume_mailbox_cursors"("updated_at")
WHERE "deleted_at" IS NULL;
-- Create indexes for resume_mailbox_statistics
CREATE UNIQUE INDEX "idx_resume_mailbox_statistics_mailbox_date" ON "resume_mailbox_statistics"("mailbox_id", "date")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_resume_mailbox_statistics_mailbox_id" ON "resume_mailbox_statistics"("mailbox_id")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_resume_mailbox_statistics_date" ON "resume_mailbox_statistics"("date")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_resume_mailbox_statistics_updated_at" ON "resume_mailbox_statistics"("updated_at")
WHERE "deleted_at" IS NULL;
-- Add comments to tables
COMMENT ON TABLE "resume_mailbox_settings" IS '邮箱简历采集配置表';
COMMENT ON TABLE "resume_mailbox_cursors" IS '邮箱同步游标表';
COMMENT ON TABLE "resume_mailbox_statistics" IS '邮箱同步统计表';
-- Add comments to key columns
COMMENT ON COLUMN "resume_mailbox_settings"."name" IS '任务名称，页面展示用';
COMMENT ON COLUMN "resume_mailbox_settings"."email_address" IS '邮箱账号';
COMMENT ON COLUMN "resume_mailbox_settings"."protocol" IS '邮箱协议';
COMMENT ON COLUMN "resume_mailbox_settings"."encrypted_credential" IS '加密后的凭证信息(JSON格式)';
COMMENT ON COLUMN "resume_mailbox_settings"."uploader_id" IS '上传人用户ID';
COMMENT ON COLUMN "resume_mailbox_settings"."job_profile_id" IS '岗位画像ID';
COMMENT ON COLUMN "resume_mailbox_settings"."sync_interval_minutes" IS '自定义同步频率(分钟)，为空则使用平台默认';
COMMENT ON COLUMN "resume_mailbox_cursors"."protocol_cursor" IS '协议游标信息(IMAP UIDVALIDITY+UID 或 POP3 UIDL列表摘要)';
COMMENT ON COLUMN "resume_mailbox_cursors"."last_message_id" IS '最后处理的邮件ID，用于排查';
COMMENT ON COLUMN "resume_mailbox_statistics"."date" IS '统计日期(按日分区，零点对齐)';
COMMENT ON COLUMN "resume_mailbox_statistics"."synced_emails" IS '当日成功同步邮件数';
COMMENT ON COLUMN "resume_mailbox_statistics"."parsed_resumes" IS '成功解析入库的简历附件数量';
COMMENT ON COLUMN "resume_mailbox_statistics"."failed_resumes" IS '解析失败数量';
COMMENT ON COLUMN "resume_mailbox_statistics"."skipped_attachments" IS '被去重或过滤的附件数量';
COMMENT ON COLUMN "resume_mailbox_statistics"."last_sync_duration_ms" IS '最近一次同步耗时(毫秒)';