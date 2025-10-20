-- Migration Rollback: 000013_add_notification_tables
-- Description: Drop notification system tables
-- Drop notification_events table
DROP TABLE IF EXISTS "notification_events";
-- Drop notification_settings table
DROP TABLE IF EXISTS "notification_settings";