-- 000009_update_job_tables_schema.down.sql
-- 回滚职位相关表的结构修改

-- 删除新增的索引
DROP INDEX IF EXISTS "idx_job_experience_requirement_experience_type";
DROP INDEX IF EXISTS "idx_job_education_requirement_education_type";
DROP INDEX IF EXISTS "idx_job_position_work_type";

-- 7. 回滚 job_industry_requirement 表：
-- 7.1 恢复 industry 字段为必需
ALTER TABLE "job_industry_requirement" ALTER COLUMN "industry" SET NOT NULL;
-- 7.2 添加回 weight 字段
ALTER TABLE "job_industry_requirement" ADD COLUMN "weight" integer NOT NULL DEFAULT 0;
-- 7.3 添加 weight 字段约束
ALTER TABLE "job_industry_requirement" ADD CONSTRAINT "job_industry_requirement_weight_check" CHECK ("weight" >= 0 AND "weight" <= 100);

-- 6. 回滚 job_experience_requirement 表：
-- 6.1 删除 experience_type 字段
ALTER TABLE "job_experience_requirement" DROP COLUMN IF EXISTS "experience_type";
-- 6.2 添加回 weight 字段
ALTER TABLE "job_experience_requirement" ADD COLUMN "weight" integer NOT NULL DEFAULT 0;
-- 6.3 添加 weight 字段约束
ALTER TABLE "job_experience_requirement" ADD CONSTRAINT "job_experience_requirement_weight_check" CHECK ("weight" >= 0 AND "weight" <= 100);

-- 5. 回滚 job_education_requirement 表：
-- 5.1 恢复 education_type 字段为必需，并改回原来的类型
ALTER TABLE "job_education_requirement" ALTER COLUMN "education_type" SET NOT NULL;
ALTER TABLE "job_education_requirement" ALTER COLUMN "education_type" TYPE character varying(50);
-- 5.2 重命名 education_type 字段回 min_degree
ALTER TABLE "job_education_requirement" RENAME COLUMN "education_type" TO "min_degree";
-- 5.3 添加回 weight 字段
ALTER TABLE "job_education_requirement" ADD COLUMN "weight" integer NOT NULL DEFAULT 0;
-- 5.4 添加 weight 字段约束
ALTER TABLE "job_education_requirement" ADD CONSTRAINT "job_education_requirement_weight_check" CHECK ("weight" >= 0 AND "weight" <= 100);

-- 4. 回滚 job_skill 表：
-- 4.1 恢复 type 字段的原始约束和类型
ALTER TABLE "job_skill" ALTER COLUMN "type" TYPE character varying(16);
ALTER TABLE "job_skill" ADD CONSTRAINT "job_skill_type_check" CHECK ("type" IN ('required', 'bonus'));
-- 4.2 添加回 weight 字段
ALTER TABLE "job_skill" ADD COLUMN "weight" integer NOT NULL DEFAULT 0;
-- 4.3 添加 weight 字段约束
ALTER TABLE "job_skill" ADD CONSTRAINT "job_skill_weight_check" CHECK ("weight" >= 0 AND "weight" <= 100);

-- 3. 回滚 job_responsibility 表：
-- 3.1 添加回 sort_order 字段
ALTER TABLE "job_responsibility" ADD COLUMN "sort_order" integer NOT NULL DEFAULT 0;

-- 2. 回滚 job_position 表：删除 work_type 字段
ALTER TABLE "job_position" DROP COLUMN IF EXISTS "work_type";

-- 1. 回滚 department 表：恢复 description 字段的长度限制
ALTER TABLE "department" ALTER COLUMN "description" TYPE character varying(255);