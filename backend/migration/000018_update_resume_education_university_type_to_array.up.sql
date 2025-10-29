-- 将 resume_educations 表的 university_type 字段改为 university_types，支持数组类型
-- 1. 添加新的 university_types 字段（JSON 类型）
ALTER TABLE resume_educations
ADD COLUMN university_types JSONB;
-- 2. 将现有的 university_type 数据迁移到 university_types 数组中
UPDATE resume_educations
SET university_types = CASE
        WHEN university_type IS NOT NULL
        AND university_type != '' THEN jsonb_build_array(university_type)
        ELSE '[]'::jsonb
    END;
-- 3. 删除旧的 university_type 字段
ALTER TABLE resume_educations DROP COLUMN university_type;
-- 为 resumes 表添加新字段
ALTER TABLE resumes
ADD COLUMN age INTEGER,
    ADD COLUMN personal_summary TEXT,
    ADD COLUMN expected_salary VARCHAR(255),
    ADD COLUMN expected_city VARCHAR(255),
    ADD COLUMN available_date TIMESTAMP WITH TIME ZONE,
    ADD COLUMN honors_certificates TEXT,
    ADD COLUMN other_info TEXT;
-- 为 resume_educations 表添加新字段
ALTER TABLE resume_educations
ADD COLUMN gpa DOUBLE PRECISION;
-- 为 resume_experiences 表添加新字段
ALTER TABLE resume_experiences
ADD COLUMN experience_type VARCHAR(255) DEFAULT 'work';
-- 添加注释说明
COMMENT ON COLUMN resumes.age IS '年龄';
COMMENT ON COLUMN resumes.personal_summary IS '个人简介（自我评价）';
COMMENT ON COLUMN resumes.expected_salary IS '期望薪资';
COMMENT ON COLUMN resumes.expected_city IS '期望城市';
COMMENT ON COLUMN resumes.available_date IS '到岗时间';
COMMENT ON COLUMN resumes.honors_certificates IS '荣誉与资格证书';
COMMENT ON COLUMN resumes.other_info IS '其它信息';
COMMENT ON COLUMN resume_educations.gpa IS '在校学分绩点';
COMMENT ON COLUMN resume_experiences.experience_type IS '经历类型: work, organization, volunteer, internship';