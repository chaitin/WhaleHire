-- Migration: 000015_add_resume_mailbox_tables (DOWN)
-- Created: 2024-01-01
-- Description: Rollback resume mailbox collection tables

-- Drop indexes for resume_mailbox_statistics
DROP INDEX IF EXISTS "idx_resume_mailbox_statistics_updated_at";
DROP INDEX IF EXISTS "idx_resume_mailbox_statistics_date";
DROP INDEX IF EXISTS "idx_resume_mailbox_statistics_mailbox_id";
DROP INDEX IF EXISTS "idx_resume_mailbox_statistics_mailbox_date";

-- Drop indexes for resume_mailbox_cursors
DROP INDEX IF EXISTS "idx_resume_mailbox_cursors_updated_at";
DROP INDEX IF EXISTS "idx_resume_mailbox_cursors_mailbox_id";

-- Drop indexes for resume_mailbox_settings
DROP INDEX IF EXISTS "idx_resume_mailbox_settings_last_synced_at";
DROP INDEX IF EXISTS "idx_resume_mailbox_settings_job_profile_id";
DROP INDEX IF EXISTS "idx_resume_mailbox_settings_uploader_id";
DROP INDEX IF EXISTS "idx_resume_mailbox_settings_status";
DROP INDEX IF EXISTS "idx_resume_mailbox_settings_protocol_email";

-- Drop foreign key constraints
ALTER TABLE "resume_mailbox_statistics" DROP CONSTRAINT IF EXISTS "fk_resume_mailbox_statistics_mailbox";
ALTER TABLE "resume_mailbox_cursors" DROP CONSTRAINT IF EXISTS "fk_resume_mailbox_cursors_mailbox";
ALTER TABLE "resume_mailbox_settings" DROP CONSTRAINT IF EXISTS "fk_resume_mailbox_settings_job_profile";
ALTER TABLE "resume_mailbox_settings" DROP CONSTRAINT IF EXISTS "fk_resume_mailbox_settings_uploader";

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS "resume_mailbox_statistics";
DROP TABLE IF EXISTS "resume_mailbox_cursors";
DROP TABLE IF EXISTS "resume_mailbox_settings";