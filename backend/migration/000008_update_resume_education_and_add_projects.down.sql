-- Drop resume_projects table
DROP TABLE IF EXISTS "resume_projects";

-- Remove university_type field from resume_educations table
ALTER TABLE "resume_educations" DROP COLUMN IF EXISTS "university_type";