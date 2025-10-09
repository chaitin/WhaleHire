CREATE TABLE "department" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" character varying(100) NOT NULL,
    "description" character varying(255) NULL,
    "parent_id" uuid NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "job_position" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" character varying(100) NOT NULL,
    "department_id" uuid NOT NULL,
    "location" character varying(200) NULL,
    "salary_min" double precision NULL,
    "salary_max" double precision NULL,
    "description" text NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "job_position_department_fk" FOREIGN KEY ("department_id") REFERENCES "department" ("id") ON DELETE NO ACTION
);

CREATE TABLE "job_responsibility" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "job_id" uuid NOT NULL,
    "responsibility" text NOT NULL,
    "sort_order" integer NOT NULL DEFAULT 0,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "job_responsibility_job_fk" FOREIGN KEY ("job_id") REFERENCES "job_position" ("id") ON DELETE CASCADE
);

CREATE TABLE "job_skillmeta" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "name" character varying(100) NOT NULL,
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL,
    PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "job_skillmeta_name_key" ON "job_skillmeta" ("name");

CREATE TABLE "job_skill" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "job_id" uuid NOT NULL,
    "skill_id" uuid NOT NULL,
    "type" character varying(16) NOT NULL,
    "weight" integer NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "job_skill_job_fk" FOREIGN KEY ("job_id") REFERENCES "job_position" ("id") ON DELETE CASCADE,
    CONSTRAINT "job_skill_skill_fk" FOREIGN KEY ("skill_id") REFERENCES "job_skillmeta" ("id") ON DELETE CASCADE,
    CONSTRAINT "job_skill_type_check" CHECK ("type" IN ('required', 'bonus')),
    CONSTRAINT "job_skill_weight_check" CHECK ("weight" >= 0 AND "weight" <= 100)
);
CREATE UNIQUE INDEX "job_skill_job_id_skill_id_type_key" ON "job_skill" ("job_id", "skill_id", "type");

CREATE TABLE "job_education_requirement" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "job_id" uuid NOT NULL,
    "min_degree" character varying(50) NOT NULL,
    "weight" integer NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "job_education_requirement_job_fk" FOREIGN KEY ("job_id") REFERENCES "job_position" ("id") ON DELETE CASCADE,
    CONSTRAINT "job_education_requirement_weight_check" CHECK ("weight" >= 0 AND "weight" <= 100)
);

CREATE TABLE "job_experience_requirement" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "job_id" uuid NOT NULL,
    "min_years" integer NOT NULL DEFAULT 0,
    "ideal_years" integer NOT NULL DEFAULT 0,
    "weight" integer NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "job_experience_requirement_job_fk" FOREIGN KEY ("job_id") REFERENCES "job_position" ("id") ON DELETE CASCADE,
    CONSTRAINT "job_experience_requirement_min_years_check" CHECK ("min_years" >= 0),
    CONSTRAINT "job_experience_requirement_ideal_years_check" CHECK ("ideal_years" >= 0),
    CONSTRAINT "job_experience_requirement_weight_check" CHECK ("weight" >= 0 AND "weight" <= 100)
);

CREATE TABLE "job_industry_requirement" (
    "id" uuid NOT NULL DEFAULT gen_random_uuid(),
    "job_id" uuid NOT NULL,
    "industry" character varying(100) NOT NULL,
    "company_name" character varying(200) NULL,
    "weight" integer NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "job_industry_requirement_job_fk" FOREIGN KEY ("job_id") REFERENCES "job_position" ("id") ON DELETE CASCADE,
    CONSTRAINT "job_industry_requirement_weight_check" CHECK ("weight" >= 0 AND "weight" <= 100)
);

-- Create indexes for deleted_at columns to improve query performance
CREATE INDEX "department_deleted_at_idx" ON "department" ("deleted_at");
CREATE INDEX "idx_job_skillmeta_deleted_at" ON "job_skillmeta" ("deleted_at");
CREATE INDEX "idx_job_position_deleted_at" ON "job_position" ("deleted_at");
CREATE INDEX "idx_job_skill_deleted_at" ON "job_skill" ("deleted_at");
CREATE INDEX "idx_job_responsibility_deleted_at" ON "job_responsibility" ("deleted_at");
CREATE INDEX "idx_job_experience_requirement_deleted_at" ON "job_experience_requirement" ("deleted_at");
CREATE INDEX "idx_job_education_requirement_deleted_at" ON "job_education_requirement" ("deleted_at");

-- Create triggers to automatically update updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_job_skill_updated_at BEFORE UPDATE ON job_skill FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_job_responsibility_updated_at BEFORE UPDATE ON job_responsibility FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_job_experience_requirement_updated_at BEFORE UPDATE ON job_experience_requirement FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_job_education_requirement_updated_at BEFORE UPDATE ON job_education_requirement FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
