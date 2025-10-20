-- Migration: 000013_add_notification_tables
-- Created: 2024-01-01
-- Description: Create notification system tables for event-driven notifications
-- Create notification_settings table
CREATE TABLE "notification_settings" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "channel" varchar(50) NOT NULL,
    "enabled" boolean NOT NULL DEFAULT true,
    "dingtalk_config" jsonb,
    "max_retry" integer NOT NULL DEFAULT 3,
    "timeout" integer NOT NULL DEFAULT 300,
    "description" varchar(500),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL
);
-- Create notification_events table
CREATE TABLE "notification_events" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "event_type" varchar(100) NOT NULL,
    "channel" varchar(50) NOT NULL DEFAULT 'dingtalk',
    "status" varchar(50) NOT NULL DEFAULT 'pending',
    "payload" jsonb NOT NULL,
    "template_id" varchar(100) NOT NULL,
    "target" varchar(500) NOT NULL,
    "retry_count" integer NOT NULL DEFAULT 0,
    "max_retry" integer NOT NULL DEFAULT 3,
    "timeout" integer NOT NULL DEFAULT 300,
    "last_error" text,
    "trace_id" varchar(100),
    "scheduled_at" timestamptz,
    "delivered_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL
);
-- Create indexes for notification_settings
CREATE INDEX idx_notification_settings_channel ON notification_settings(channel);
CREATE INDEX idx_notification_settings_enabled ON notification_settings(enabled);
CREATE INDEX idx_notification_settings_channel_enabled ON notification_settings(channel, enabled);
-- Create indexes for notification_events
CREATE INDEX idx_notification_events_event_type ON notification_events(event_type);
CREATE INDEX idx_notification_events_channel ON notification_events(channel);
CREATE INDEX idx_notification_events_status ON notification_events(status);
CREATE INDEX idx_notification_events_trace_id ON notification_events(trace_id);
CREATE INDEX idx_notification_events_scheduled_at ON notification_events(scheduled_at);
CREATE INDEX idx_notification_events_created_at ON notification_events(created_at);
-- Create composite indexes for notification_events
CREATE INDEX idx_notification_events_status_scheduled_at ON notification_events(status, scheduled_at);
CREATE INDEX idx_notification_events_event_type_status ON notification_events(event_type, status);
-- Add constraints for notification_settings
ALTER TABLE notification_settings
ADD CONSTRAINT chk_notification_settings_channel CHECK (
        channel IN ('dingtalk', 'email', 'webhook', 'sms')
    );
ALTER TABLE notification_settings
ADD CONSTRAINT chk_notification_settings_max_retry CHECK (
        max_retry >= 0
        AND max_retry <= 10
    );
ALTER TABLE notification_settings
ADD CONSTRAINT chk_notification_settings_timeout CHECK (
        timeout >= 1
        AND timeout <= 300
    );
-- Add constraints for notification_events
ALTER TABLE notification_events
ADD CONSTRAINT chk_notification_events_channel CHECK (
        channel IN ('dingtalk', 'email', 'webhook', 'sms')
    );
ALTER TABLE notification_events
ADD CONSTRAINT chk_notification_events_status CHECK (
        status IN ('pending', 'delivering', 'delivered', 'failed')
    );
ALTER TABLE notification_events
ADD CONSTRAINT chk_notification_events_retry_count CHECK (retry_count >= 0);
ALTER TABLE notification_events
ADD CONSTRAINT chk_notification_events_max_retry CHECK (
        max_retry >= 0
        AND max_retry <= 10
    );
ALTER TABLE notification_events
ADD CONSTRAINT chk_notification_events_timeout CHECK (
        timeout >= 1
        AND timeout <= 300
    );
-- Add comments for better documentation
COMMENT ON TABLE notification_settings IS '通知设置表，存储各种通知渠道的配置信息';
COMMENT ON COLUMN notification_settings.channel IS '通知渠道类型';
COMMENT ON COLUMN notification_settings.enabled IS '是否启用该通知渠道';
COMMENT ON COLUMN notification_settings.dingtalk_config IS '钉钉通知配置（JSON格式）';
COMMENT ON COLUMN notification_settings.max_retry IS '最大重试次数';
COMMENT ON COLUMN notification_settings.timeout IS '超时时间（秒）';
COMMENT ON COLUMN notification_settings.description IS '通知设置描述';
COMMENT ON TABLE notification_events IS '通知事件表，存储待发送和已发送的通知事件';
COMMENT ON COLUMN notification_events.event_type IS '事件类型';
COMMENT ON COLUMN notification_events.channel IS '通知渠道';
COMMENT ON COLUMN notification_events.status IS '通知状态';
COMMENT ON COLUMN notification_events.payload IS '通知负载数据（JSON格式）';
COMMENT ON COLUMN notification_events.template_id IS '通知模板ID';
COMMENT ON COLUMN notification_events.target IS '目标地址，如webhook URL';
COMMENT ON COLUMN notification_events.retry_count IS '当前重试次数';
COMMENT ON COLUMN notification_events.max_retry IS '最大重试次数';
COMMENT ON COLUMN notification_events.timeout IS '超时时间（秒）';
COMMENT ON COLUMN notification_events.last_error IS '最后一次错误信息';
COMMENT ON COLUMN notification_events.trace_id IS '链路追踪ID';
COMMENT ON COLUMN notification_events.scheduled_at IS '计划投递时间';
COMMENT ON COLUMN notification_events.delivered_at IS '实际投递时间';