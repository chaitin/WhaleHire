-- Migration: 000014_add_name_field_to_notification_settings
-- Created: 2024-01-01
-- Description: Add name field to notification_settings table to support multiple configurations per channel type
-- Add name field to notification_settings table
ALTER TABLE "notification_settings"
ADD COLUMN "name" varchar(256) NOT NULL DEFAULT '';
-- Update existing records to have a default name based on channel
UPDATE "notification_settings"
SET "name" = 'default_' || "channel"
WHERE "name" = '';
-- Remove the old channel index
DROP INDEX IF EXISTS idx_notification_settings_channel;
-- Create new unique index on (name, channel) combination
CREATE UNIQUE INDEX idx_notification_settings_name_channel ON notification_settings(name, channel);
-- Create index on name field for faster lookups
CREATE INDEX idx_notification_settings_name ON notification_settings(name);
-- Update the composite index to include name
DROP INDEX IF EXISTS idx_notification_settings_channel_enabled;
CREATE INDEX idx_notification_settings_name_channel_enabled ON notification_settings(name, channel, enabled);