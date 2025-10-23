-- Migration: 000014_add_name_field_to_notification_settings (DOWN)
-- Created: 2024-01-01
-- Description: Rollback name field addition from notification_settings table
-- Drop the new indexes created in UP migration
DROP INDEX IF EXISTS idx_notification_settings_name_channel;
DROP INDEX IF EXISTS idx_notification_settings_name;
DROP INDEX IF EXISTS idx_notification_settings_name_channel_enabled;
-- Remove the name column
ALTER TABLE "notification_settings" DROP COLUMN IF EXISTS "name";
-- Recreate the original indexes that were dropped in UP migration
CREATE INDEX IF NOT EXISTS idx_notification_settings_channel ON notification_settings(channel);
CREATE INDEX IF NOT EXISTS idx_notification_settings_channel_enabled ON notification_settings(channel, enabled);