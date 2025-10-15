-- Migration: 000011_add_screening_tables
-- Created: 2024-01-01
-- Description: Create screening tables for AI-powered resume screening system

-- Create screening_tasks table
CREATE TABLE "screening_tasks" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "job_position_id" uuid NOT NULL,
    "created_by" uuid NOT NULL,
    "status" varchar(50) NOT NULL DEFAULT 'pending',
    "dimension_weights" jsonb,
    "llm_config" jsonb,
    "notes" text,
    "resume_total" integer NOT NULL DEFAULT 0,
    "resume_processed" integer NOT NULL DEFAULT 0,
    "resume_succeeded" integer NOT NULL DEFAULT 0,
    "resume_failed" integer NOT NULL DEFAULT 0,
    "agent_version" varchar(50),
    "started_at" timestamptz,
    "finished_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL
);

-- Create screening_task_resumes table
CREATE TABLE "screening_task_resumes" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "task_id" uuid NOT NULL,
    "resume_id" uuid NOT NULL,
    "status" varchar(50) NOT NULL DEFAULT 'pending',
    "error_message" text,
    "ranking" integer,
    "score" double precision,
    "processed_at" timestamptz,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL
);

-- Create screening_results table
CREATE TABLE "screening_results" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "task_id" uuid NOT NULL,
    "job_position_id" uuid NOT NULL,
    "resume_id" uuid NOT NULL,
    "overall_score" double precision NOT NULL,
    "match_level" varchar(20),
    "dimension_scores" jsonb,
    "skill_detail" jsonb,
    "responsibility_detail" jsonb,
    "experience_detail" jsonb,
    "education_detail" jsonb,
    "industry_detail" jsonb,
    "basic_detail" jsonb,
    "recommendations" jsonb,
    "trace_id" varchar(100),
    "runtime_metadata" jsonb,
    "sub_agent_versions" jsonb,
    "matched_at" timestamptz NOT NULL DEFAULT now(),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL
);

-- Create screening_run_metrics table
CREATE TABLE "screening_run_metrics" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "task_id" uuid NOT NULL,
    "avg_score" double precision,
    "histogram" jsonb,
    "tokens_input" bigint,
    "tokens_output" bigint,
    "total_cost" double precision,
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now(),
    "deleted_at" timestamptz NULL
);

-- Add foreign key constraints for screening_tasks
ALTER TABLE "screening_tasks"
ADD CONSTRAINT "fk_screening_tasks_job_position_id" FOREIGN KEY ("job_position_id") REFERENCES "job_position" ("id") ON DELETE CASCADE;

ALTER TABLE "screening_tasks"
ADD CONSTRAINT "fk_screening_tasks_created_by" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON DELETE CASCADE;

-- Add foreign key constraints for screening_task_resumes
ALTER TABLE "screening_task_resumes"
ADD CONSTRAINT "fk_screening_task_resumes_task_id" FOREIGN KEY ("task_id") REFERENCES "screening_tasks" ("id") ON DELETE CASCADE;

ALTER TABLE "screening_task_resumes"
ADD CONSTRAINT "fk_screening_task_resumes_resume_id" FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id") ON DELETE CASCADE;

-- Add foreign key constraints for screening_results
ALTER TABLE "screening_results"
ADD CONSTRAINT "fk_screening_results_task_id" FOREIGN KEY ("task_id") REFERENCES "screening_tasks" ("id") ON DELETE CASCADE;

ALTER TABLE "screening_results"
ADD CONSTRAINT "fk_screening_results_job_position_id" FOREIGN KEY ("job_position_id") REFERENCES "job_position" ("id") ON DELETE CASCADE;

ALTER TABLE "screening_results"
ADD CONSTRAINT "fk_screening_results_resume_id" FOREIGN KEY ("resume_id") REFERENCES "resumes" ("id") ON DELETE CASCADE;

-- Add foreign key constraints for screening_run_metrics
ALTER TABLE "screening_run_metrics"
ADD CONSTRAINT "fk_screening_run_metrics_task_id" FOREIGN KEY ("task_id") REFERENCES "screening_tasks" ("id") ON DELETE CASCADE;

-- Add unique constraints
ALTER TABLE "screening_task_resumes"
ADD CONSTRAINT "uk_screening_task_resumes_task_resume" UNIQUE ("task_id", "resume_id");

ALTER TABLE "screening_results"
ADD CONSTRAINT "uk_screening_results_task_resume" UNIQUE ("task_id", "resume_id");

-- Add indexes for screening_tasks
CREATE INDEX "idx_screening_tasks_job_position_id" ON "screening_tasks" ("job_position_id");
CREATE INDEX "idx_screening_tasks_status" ON "screening_tasks" ("status");
CREATE INDEX "idx_screening_tasks_created_by" ON "screening_tasks" ("created_by");
CREATE INDEX "idx_screening_tasks_created_at" ON "screening_tasks" ("created_at");

-- Add indexes for screening_task_resumes
CREATE INDEX "idx_screening_task_resumes_task_id" ON "screening_task_resumes" ("task_id");
CREATE INDEX "idx_screening_task_resumes_status" ON "screening_task_resumes" ("status");
CREATE INDEX "idx_screening_task_resumes_score" ON "screening_task_resumes" ("score" DESC);
CREATE INDEX "idx_screening_task_resumes_task_ranking" ON "screening_task_resumes" ("task_id", "ranking");

-- Add indexes for screening_results
CREATE INDEX "idx_screening_results_task_id" ON "screening_results" ("task_id");
CREATE INDEX "idx_screening_results_job_position_resume" ON "screening_results" ("job_position_id", "resume_id");
CREATE INDEX "idx_screening_results_overall_score" ON "screening_results" ("overall_score" DESC);
CREATE INDEX "idx_screening_results_match_level" ON "screening_results" ("match_level");
CREATE INDEX "idx_screening_results_matched_at" ON "screening_results" ("matched_at");

-- Add indexes for screening_run_metrics
CREATE INDEX "idx_screening_run_metrics_task_id" ON "screening_run_metrics" ("task_id");
CREATE INDEX "idx_screening_run_metrics_avg_score" ON "screening_run_metrics" ("avg_score");
CREATE INDEX "idx_screening_run_metrics_created_at" ON "screening_run_metrics" ("created_at");

-- Add check constraints for match_level enum
ALTER TABLE "screening_results"
ADD CONSTRAINT "chk_screening_results_match_level" CHECK ("match_level" IN ('excellent', 'good', 'fair', 'poor'));

-- Add comments for better documentation
COMMENT ON TABLE "screening_tasks" IS '智能筛选任务表';
COMMENT ON TABLE "screening_task_resumes" IS '筛选任务-简历关联表';
COMMENT ON TABLE "screening_results" IS '筛选结果表';
COMMENT ON TABLE "screening_run_metrics" IS '筛选运行指标表';

COMMENT ON COLUMN "screening_tasks"."status" IS '任务状态: pending/running/completed/failed';
COMMENT ON COLUMN "screening_tasks"."dimension_weights" IS '维度权重配置';
COMMENT ON COLUMN "screening_tasks"."llm_config" IS 'LLM/Embedding配置';
COMMENT ON COLUMN "screening_tasks"."notes" IS '任务备注';
COMMENT ON COLUMN "screening_tasks"."resume_total" IS '任务内候选总数';
COMMENT ON COLUMN "screening_tasks"."resume_processed" IS '已处理数';
COMMENT ON COLUMN "screening_tasks"."resume_succeeded" IS '成功匹配数';
COMMENT ON COLUMN "screening_tasks"."resume_failed" IS '失败数';
COMMENT ON COLUMN "screening_tasks"."agent_version" IS '智能匹配Agent版本号';

COMMENT ON COLUMN "screening_task_resumes"."status" IS '处理状态: pending/running/completed/failed';
COMMENT ON COLUMN "screening_task_resumes"."error_message" IS '错误信息';
COMMENT ON COLUMN "screening_task_resumes"."ranking" IS '排名（任务内）';
COMMENT ON COLUMN "screening_task_resumes"."score" IS '综合分快照（任务维度）';
COMMENT ON COLUMN "screening_task_resumes"."processed_at" IS '处理完成时间';

COMMENT ON COLUMN "screening_results"."overall_score" IS '聚合总分 0-100';
COMMENT ON COLUMN "screening_results"."match_level" IS '匹配等级: excellent/good/fair/poor';
COMMENT ON COLUMN "screening_results"."dimension_scores" IS '各维度分数';
COMMENT ON COLUMN "screening_results"."skill_detail" IS '技能匹配详情';
COMMENT ON COLUMN "screening_results"."responsibility_detail" IS '职责匹配详情';
COMMENT ON COLUMN "screening_results"."experience_detail" IS '经验匹配详情';
COMMENT ON COLUMN "screening_results"."education_detail" IS '教育匹配详情';
COMMENT ON COLUMN "screening_results"."industry_detail" IS '行业匹配详情';
COMMENT ON COLUMN "screening_results"."basic_detail" IS '基本信息匹配详情';
COMMENT ON COLUMN "screening_results"."recommendations" IS '匹配建议';
COMMENT ON COLUMN "screening_results"."trace_id" IS '链路追踪ID';
COMMENT ON COLUMN "screening_results"."runtime_metadata" IS '运行时元数据';
COMMENT ON COLUMN "screening_results"."sub_agent_versions" IS '各Agent版本快照';

COMMENT ON COLUMN "screening_run_metrics"."avg_score" IS '平均分数';
COMMENT ON COLUMN "screening_run_metrics"."histogram" IS '分数段分布';
COMMENT ON COLUMN "screening_run_metrics"."tokens_input" IS '任务总输入tokens';
COMMENT ON COLUMN "screening_run_metrics"."tokens_output" IS '任务总输出tokens';
COMMENT ON COLUMN "screening_run_metrics"."total_cost" IS '模型调用成本统计';