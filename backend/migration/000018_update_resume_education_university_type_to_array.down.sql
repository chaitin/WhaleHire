-- 回滚 resume_educations 表的 university_types 字段改为 university_type 单个类型
-- 1. 添加旧的 university_type 字段
ALTER TABLE resume_educations
ADD COLUMN university_type VARCHAR;
-- 2. 将 university_types 数组的第一个元素迁移回 university_type
UPDATE resume_educations
SET university_type = CASE
        WHEN university_types IS NOT NULL
        AND jsonb_array_length(university_types) > 0 THEN university_types->>0
        ELSE NULL
    END;
-- 3. 删除新的 university_types 字段
ALTER TABLE resume_educations DROP COLUMN university_types;
-- 删除 resumes 表的新字段
ALTER TABLE resumes DROP COLUMN IF EXISTS age,
    DROP COLUMN IF EXISTS personal_summary,
    DROP COLUMN IF EXISTS expected_salary,
    DROP COLUMN IF EXISTS expected_city,
    DROP COLUMN IF EXISTS available_date,
    DROP COLUMN IF EXISTS honors_certificates,
    DROP COLUMN IF EXISTS other_info;
-- 删除 resume_educations 表的新字段
ALTER TABLE resume_educations DROP COLUMN IF EXISTS gpa;
-- 删除 resume_experiences 表的新字段
ALTER TABLE resume_experiences DROP COLUMN IF EXISTS experience_type;