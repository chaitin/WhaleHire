-- Migration: 000010_create_resume_job_applications
-- Created: 2024-01-01
-- Description: Rollback - Drop resume_job_applications table
-- Drop indexes first
DROP INDEX IF EXISTS "idx_resume_job_applications_applied_at";
DROP INDEX IF EXISTS "idx_resume_job_applications_status";
DROP INDEX IF EXISTS "idx_resume_job_applications_job_position_id";
DROP INDEX IF EXISTS "idx_resume_job_applications_resume_id";
-- Drop constraints
ALTER TABLE "resume_job_applications" DROP CONSTRAINT IF EXISTS "uk_resume_job_applications_resume_job";
ALTER TABLE "resume_job_applications" DROP CONSTRAINT IF EXISTS "fk_resume_job_applications_job_position_id";
ALTER TABLE "resume_job_applications" DROP CONSTRAINT IF EXISTS "fk_resume_job_applications_resume_id";
-- Drop the table
DROP TABLE IF EXISTS "resume_job_applications";