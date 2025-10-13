-- 000009_update_job_tables_schema.down.sql
-- 回滚职位相关表的结构修改

-- 删除新增的索引
DROP INDEX IF EXISTS "idx_job_experience_requirement_experience_type";
DROP INDEX IF EXISTS "idx_job_education_requirement_education_type";
DROP INDEX IF EXISTS "idx_job_position_work_type";

-- 7. 回滚 job_industry_requirement 表：
-- 7.1 恢复 weight 字段为必需
ALTER TABLE "job_industry_requirement" ALTER COLUMN "weight" SET NOT NULL;

-- 6. 回滚 job_experience_requirement 表：
-- 6.1 删除 experience_type 字段
ALTER TABLE "job_experience_requirement" DROP COLUMN IF EXISTS "experience_type";

-- 6.2 恢复 weight 字段为必需
ALTER TABLE "job_experience_requirement" ALTER COLUMN "weight" SET NOT NULL;

-- 5. 回滚 job_education_requirement 表：
-- 5.1 恢复 weight 字段为必需
ALTER TABLE "job_education_requirement" ALTER COLUMN "weight" SET NOT NULL;

-- 5.2 恢复 education_type 字段为必需，并改回原来的类型
ALTER TABLE "job_education_requirement" ALTER COLUMN "education_type" SET NOT NULL;
ALTER TABLE "job_education_requirement" ALTER COLUMN "education_type" TYPE character varying(50);

-- 5.3 重命名 education_type 字段回 min_degree
ALTER TABLE "job_education_requirement" RENAME COLUMN "education_type" TO "min_degree";

-- 4. 回滚 job_skill 表：
-- 4.1 恢复 weight 字段为必需
ALTER TABLE "job_skill" ALTER COLUMN "weight" SET NOT NULL;

-- 4.2 恢复 type 字段的原始约束和类型
ALTER TABLE "job_skill" ALTER COLUMN "type" TYPE character varying(16);
ALTER TABLE "job_skill" ADD CONSTRAINT "job_skill_type_check" CHECK ("type" IN ('required', 'bonus'));

-- 3. 回滚 job_responsibility 表：
-- 3.1 添加回 sort_order 字段
ALTER TABLE "job_responsibility" ADD COLUMN "sort_order" integer NOT NULL DEFAULT 0;

-- 3.2 删除 weight 字段和约束
ALTER TABLE "job_responsibility" DROP CONSTRAINT IF EXISTS "job_responsibility_weight_check";
ALTER TABLE "job_responsibility" DROP COLUMN IF EXISTS "weight";

-- 2. 回滚 job_position 表：删除 work_type 字段
ALTER TABLE "job_position" DROP COLUMN IF EXISTS "work_type";

-- 1. 回滚 department 表：恢复 description 字段的长度限制
ALTER TABLE "department" ALTER COLUMN "description" TYPE character varying(255);