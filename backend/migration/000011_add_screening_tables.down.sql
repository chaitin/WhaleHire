-- Migration: 000011_add_screening_tables
-- Created: 2024-01-01
-- Description: Rollback - Drop screening tables for AI-powered resume screening system

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS "idx_screening_run_metrics_created_at";
DROP INDEX IF EXISTS "idx_screening_run_metrics_avg_score";
DROP INDEX IF EXISTS "idx_screening_run_metrics_task_id";

DROP INDEX IF EXISTS "idx_screening_results_matched_at";
DROP INDEX IF EXISTS "idx_screening_results_match_level";
DROP INDEX IF EXISTS "idx_screening_results_overall_score";
DROP INDEX IF EXISTS "idx_screening_results_job_position_resume";
DROP INDEX IF EXISTS "idx_screening_results_task_id";

DROP INDEX IF EXISTS "idx_screening_task_resumes_task_ranking";
DROP INDEX IF EXISTS "idx_screening_task_resumes_score";
DROP INDEX IF EXISTS "idx_screening_task_resumes_status";
DROP INDEX IF EXISTS "idx_screening_task_resumes_task_id";

DROP INDEX IF EXISTS "idx_screening_tasks_created_at";
DROP INDEX IF EXISTS "idx_screening_tasks_created_by";
DROP INDEX IF EXISTS "idx_screening_tasks_status";
DROP INDEX IF EXISTS "idx_screening_tasks_job_position_id";

-- Drop check constraints
ALTER TABLE "screening_results" DROP CONSTRAINT IF EXISTS "chk_screening_results_match_level";

-- Drop unique constraints
ALTER TABLE "screening_results" DROP CONSTRAINT IF EXISTS "uk_screening_results_task_resume";
ALTER TABLE "screening_task_resumes" DROP CONSTRAINT IF EXISTS "uk_screening_task_resumes_task_resume";

-- Drop foreign key constraints for screening_run_metrics
ALTER TABLE "screening_run_metrics" DROP CONSTRAINT IF EXISTS "fk_screening_run_metrics_task_id";

-- Drop foreign key constraints for screening_results
ALTER TABLE "screening_results" DROP CONSTRAINT IF EXISTS "fk_screening_results_resume_id";
ALTER TABLE "screening_results" DROP CONSTRAINT IF EXISTS "fk_screening_results_job_position_id";
ALTER TABLE "screening_results" DROP CONSTRAINT IF EXISTS "fk_screening_results_task_id";

-- Drop foreign key constraints for screening_task_resumes
ALTER TABLE "screening_task_resumes" DROP CONSTRAINT IF EXISTS "fk_screening_task_resumes_resume_id";
ALTER TABLE "screening_task_resumes" DROP CONSTRAINT IF EXISTS "fk_screening_task_resumes_task_id";

-- Drop foreign key constraints for screening_tasks
ALTER TABLE "screening_tasks" DROP CONSTRAINT IF EXISTS "fk_screening_tasks_created_by";
ALTER TABLE "screening_tasks" DROP CONSTRAINT IF EXISTS "fk_screening_tasks_job_position_id";

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS "screening_run_metrics";
DROP TABLE IF EXISTS "screening_results";
DROP TABLE IF EXISTS "screening_task_resumes";
DROP TABLE IF EXISTS "screening_tasks";