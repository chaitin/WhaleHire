-- Migration: add_job_profile_tables

-- Drop indexes first
DROP INDEX IF EXISTS "idx_job_position_status";
DROP INDEX IF EXISTS "idx_job_position_created_by";

-- Drop constraints
ALTER TABLE "job_position" DROP CONSTRAINT IF EXISTS "job_position_status_check";
ALTER TABLE "job_position" DROP CONSTRAINT IF EXISTS "job_position_created_by_fk";

-- Drop columns
ALTER TABLE "job_position" DROP COLUMN IF EXISTS "status";
ALTER TABLE "job_position" DROP COLUMN IF EXISTS "created_by";

-- Drop triggers
DROP TRIGGER IF EXISTS update_job_skill_updated_at ON job_skill;
DROP TRIGGER IF EXISTS update_job_responsibility_updated_at ON job_responsibility;
DROP TRIGGER IF EXISTS update_job_experience_requirement_updated_at ON job_experience_requirement;
DROP TRIGGER IF EXISTS update_job_education_requirement_updated_at ON job_education_requirement;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS "department_deleted_at_idx";
DROP INDEX IF EXISTS "idx_job_skillmeta_deleted_at";
DROP INDEX IF EXISTS "idx_job_position_deleted_at";
DROP INDEX IF EXISTS "idx_job_skill_deleted_at";
DROP INDEX IF EXISTS "idx_job_responsibility_deleted_at";
DROP INDEX IF EXISTS "idx_job_experience_requirement_deleted_at";
DROP INDEX IF EXISTS "idx_job_education_requirement_deleted_at";

-- Drop tables in reverse order of creation
DROP TABLE IF EXISTS "job_industry_requirement";
DROP TABLE IF EXISTS "job_experience_requirement";
DROP TABLE IF EXISTS "job_education_requirement";
DROP TABLE IF EXISTS "job_skill";
DROP TABLE IF EXISTS "job_skillmeta";
DROP TABLE IF EXISTS "job_responsibility";
DROP TABLE IF EXISTS "job_position";
DROP TABLE IF EXISTS "department";
