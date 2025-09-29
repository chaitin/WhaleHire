CREATE TABLE "resumes" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "name" character varying NULL,
    "gender" character varying NULL,
    "birthday" timestamptz NULL,
    "email" character varying NULL,
    "phone" character varying NULL,
    "current_city" character varying NULL,
    "highest_education" character varying NULL,
    "years_experience" double precision NULL,
    "resume_file_url" character varying NULL,
    "status" character varying NOT NULL DEFAULT 'pending',
    "error_message" character varying NULL,
    "parsed_at" timestamptz NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "user_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "resumes_users_resumes" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE NO ACTION
);
CREATE TABLE "resume_educations" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "school" character varying NULL,
    "degree" character varying NULL,
    "major" character varying NULL,
    "start_date" timestamptz NULL,
    "end_date" timestamptz NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "resume_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "resume_educations_resumes_educations" FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id") ON DELETE NO ACTION
);
CREATE TABLE "resume_experiences" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "company" character varying NULL,
    "position" character varying NULL,
    "title" character varying NULL,
    "start_date" timestamptz NULL,
    "end_date" timestamptz NULL,
    "description" text NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "resume_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "resume_experiences_resumes_experiences" FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id") ON DELETE NO ACTION
);
CREATE TABLE "resume_logs" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "action" character varying NOT NULL,
    "message" text NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "resume_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "resume_logs_resumes_logs" FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id") ON DELETE NO ACTION
);
CREATE TABLE "resume_skills" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "skill_name" character varying NULL,
    "level" character varying NULL,
    "description" character varying NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "resume_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "resume_skills_resumes_skills" FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id") ON DELETE NO ACTION
);
CREATE TABLE "resume_document_parses" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "file_id" character varying NULL,
    "content" text NULL,
    "file_type" character varying NULL,
    "filename" character varying NULL,
    "title" character varying NULL,
    "upload_at" timestamptz NULL,
    "status" character varying NOT NULL DEFAULT 'pending',
    "error_message" character varying NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "resume_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "resume_document_parses_resumes_document_parse" FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id") ON DELETE NO ACTION
);