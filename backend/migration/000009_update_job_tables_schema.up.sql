-- 000009_update_job_tables_schema.up.sql
-- 更新职位相关表的结构

-- 1. 修改 department 表：移除 description 字段的长度限制
ALTER TABLE "department" ALTER COLUMN "description" TYPE text;

-- 2. 修改 job_position 表：添加 work_type 字段
ALTER TABLE "job_position" ADD COLUMN "work_type" character varying(50) NULL;

-- 3. 修改 job_responsibility 表：
-- 3.1 添加 weight 字段（可选，范围 0-100）
ALTER TABLE "job_responsibility" ADD COLUMN "weight" integer NULL;
ALTER TABLE "job_responsibility" ADD CONSTRAINT "job_responsibility_weight_check" CHECK ("weight" >= 0 AND "weight" <= 100);

-- 3.2 移除 sort_order 字段（如果存在）
ALTER TABLE "job_responsibility" DROP COLUMN IF EXISTS "sort_order";

-- 4. 修改 job_skill 表：
-- 4.1 修改 type 字段为支持中文值
ALTER TABLE "job_skill" DROP CONSTRAINT IF EXISTS "job_skill_type_check";
ALTER TABLE "job_skill" ALTER COLUMN "type" TYPE character varying(50);

-- 4.2 修改 weight 字段为可选
ALTER TABLE "job_skill" ALTER COLUMN "weight" DROP NOT NULL;

-- 5. 修改 job_education_requirement 表：
-- 5.1 重命名 min_degree 字段为 education_type
ALTER TABLE "job_education_requirement" RENAME COLUMN "min_degree" TO "education_type";

-- 5.2 修改 education_type 字段为可选，支持中文值
ALTER TABLE "job_education_requirement" ALTER COLUMN "education_type" DROP NOT NULL;
ALTER TABLE "job_education_requirement" ALTER COLUMN "education_type" TYPE character varying(50);

-- 5.3 修改 weight 字段为可选
ALTER TABLE "job_education_requirement" ALTER COLUMN "weight" DROP NOT NULL;

-- 6. 修改 job_experience_requirement 表：
-- 6.1 添加 experience_type 字段（可选）
ALTER TABLE "job_experience_requirement" ADD COLUMN "experience_type" character varying(50) NULL;

-- 6.2 修改 weight 字段为可选
ALTER TABLE "job_experience_requirement" ALTER COLUMN "weight" DROP NOT NULL;

-- 7. 修改 job_industry_requirement 表：
-- 7.1 修改 weight 字段为可选
ALTER TABLE "job_industry_requirement" ALTER COLUMN "weight" DROP NOT NULL;

-- 更新索引和约束
-- 为新增的字段创建索引
CREATE INDEX IF NOT EXISTS "idx_job_position_work_type" ON "job_position" ("work_type");
CREATE INDEX IF NOT EXISTS "idx_job_education_requirement_education_type" ON "job_education_requirement" ("education_type");
CREATE INDEX IF NOT EXISTS "idx_job_experience_requirement_experience_type" ON "job_experience_requirement" ("experience_type");