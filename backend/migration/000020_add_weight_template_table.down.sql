-- Migration: 000020_add_weight_template_table (Rollback)
-- Created: 2025-01-10
-- Description: Drop weight_template table

-- Drop indexes
DROP INDEX IF EXISTS "idx_weight_template_created_at";
DROP INDEX IF EXISTS "idx_weight_template_created_by";

-- Drop foreign key constraint
ALTER TABLE "weight_template"
DROP CONSTRAINT IF EXISTS "weight_template_created_by_fk";

-- Drop table
DROP TABLE IF EXISTS "weight_template";

