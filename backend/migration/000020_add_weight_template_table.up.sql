-- Migration: 000020_add_weight_template_table
-- Created: 2025-01-10
-- Description: Create weight_template table for storing screening weight templates

-- Create weight_template table
CREATE TABLE "weight_template" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "name" varchar(100) NOT NULL,
    "description" varchar(500),
    "weights" jsonb NOT NULL,
    "created_by" uuid NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now()
);

-- Add foreign key constraint
ALTER TABLE "weight_template"
ADD CONSTRAINT "weight_template_created_by_fk" 
FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE NO ACTION;

-- Create indexes
CREATE INDEX "idx_weight_template_created_by" ON "weight_template" ("created_by");
CREATE INDEX "idx_weight_template_created_at" ON "weight_template" ("created_at");

-- Add comments
COMMENT ON TABLE "weight_template" IS '权重模板表，用于存储筛查权重配置模板';
COMMENT ON COLUMN "weight_template"."id" IS '模板ID';
COMMENT ON COLUMN "weight_template"."name" IS '模板名称';
COMMENT ON COLUMN "weight_template"."description" IS '模板描述';
COMMENT ON COLUMN "weight_template"."weights" IS '权重配置 JSONB';
COMMENT ON COLUMN "weight_template"."created_by" IS '创建者ID';
COMMENT ON COLUMN "weight_template"."created_at" IS '创建时间';
COMMENT ON COLUMN "weight_template"."updated_at" IS '更新时间';
COMMENT ON COLUMN "weight_template"."deleted_at" IS '删除时间（软删除）';

