-- Migration: update_resume_user_id_to_uploader_id
-- Created: 2025-01-27 15:45:00
-- Description: Rollback - Rename uploader_id column back to user_id in resumes table
-- Drop the new constraint
ALTER TABLE "resumes" DROP CONSTRAINT "resumes_users_uploader";
-- Add back the original constraint
ALTER TABLE "resumes"
ADD CONSTRAINT "resumes_users_resumes" FOREIGN KEY ("uploader_id") REFERENCES "users" ("id") ON DELETE NO ACTION;
-- Rename the column back from uploader_id to user_id
ALTER TABLE "resumes"
    RENAME COLUMN "uploader_id" TO "user_id";