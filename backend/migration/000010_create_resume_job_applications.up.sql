-- Migration: 000010_create_resume_job_applications
-- Created: 2024-01-01
-- Description: Create resume_job_applications table for many-to-many relationship between resumes and job positions
-- Create resume_job_applications table
CREATE TABLE "resume_job_applications" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "resume_id" uuid NOT NULL,
    "job_position_id" uuid NOT NULL,
    "status" varchar(50) NOT NULL DEFAULT 'applied',
    "source" varchar(100),
    "notes" text,
    "applied_at" timestamptz NOT NULL DEFAULT now(),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL
);
-- Add foreign key constraints
ALTER TABLE "resume_job_applications"
ADD CONSTRAINT "fk_resume_job_applications_resume_id" FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id") ON DELETE CASCADE;
ALTER TABLE "resume_job_applications"
ADD CONSTRAINT "fk_resume_job_applications_job_position_id" FOREIGN KEY ("job_position_id") REFERENCES "job_position" ("id") ON DELETE CASCADE;
-- Add unique constraint to prevent duplicate applications
ALTER TABLE "resume_job_applications"
ADD CONSTRAINT "uk_resume_job_applications_resume_job" UNIQUE ("resume_id", "job_position_id");
-- Add indexes for performance
CREATE INDEX "idx_resume_job_applications_resume_id" ON "resume_job_applications" ("resume_id");
CREATE INDEX "idx_resume_job_applications_job_position_id" ON "resume_job_applications" ("job_position_id");
CREATE INDEX "idx_resume_job_applications_status" ON "resume_job_applications" ("status");
CREATE INDEX "idx_resume_job_applications_applied_at" ON "resume_job_applications" ("applied_at");