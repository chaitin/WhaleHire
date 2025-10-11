-- Add university_type field to resume_educations table
ALTER TABLE "resume_educations"
ADD COLUMN "university_type" character varying NULL;
-- Create resume_projects table
CREATE TABLE "resume_projects" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "name" character varying NULL,
    "role" character varying NULL,
    "company" character varying NULL,
    "start_date" timestamptz NULL,
    "end_date" timestamptz NULL,
    "description" text NULL,
    "responsibilities" text NULL,
    "achievements" text NULL,
    "technologies" character varying NULL,
    "project_url" character varying NULL,
    "project_type" character varying NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "resume_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "resume_projects_resumes_projects" FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id") ON DELETE NO ACTION
);