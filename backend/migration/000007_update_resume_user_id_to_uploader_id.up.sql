-- Migration: update_resume_user_id_to_uploader_id
-- Created: 2025-01-27 15:45:00
-- Description: Rename user_id column to uploader_id in resumes table to better reflect its purpose
-- Rename the column from user_id to uploader_id
ALTER TABLE "resumes"
    RENAME COLUMN "user_id" TO "uploader_id";
-- Update the foreign key constraint name to reflect the new column name
-- First drop the existing constraint
ALTER TABLE "resumes" DROP CONSTRAINT "resumes_users_resumes";
-- Add the new constraint with updated name
ALTER TABLE "resumes"
ADD CONSTRAINT "resumes_users_uploader" FOREIGN KEY ("uploader_id") REFERENCES "users" ("id") ON DELETE NO ACTION;