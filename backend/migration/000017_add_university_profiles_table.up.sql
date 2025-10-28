-- Migration: 000017_add_university_profiles_table
-- Created: 2024-12-05
-- 描述：初始化高校画像表并启用 pgvector 向量检索能力
-- 确保已启用 pgvector 扩展（幂等操作）
CREATE EXTENSION IF NOT EXISTS vector;
-- 创建高校画像主表
CREATE TABLE "university_profiles" (
    "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "deleted_at" timestamptz NULL,
    "name_cn" varchar(256) NOT NULL,
    "name_en" varchar(256),
    "alias" varchar(256),
    "country" varchar(128),
    "is_double_first_class" boolean NOT NULL DEFAULT false,
    "is_project_985" boolean NOT NULL DEFAULT false,
    "is_project_211" boolean NOT NULL DEFAULT false,
    "is_qs_top100" boolean NOT NULL DEFAULT false,
    "rank_qs" integer,
    "overall_score" double precision,
    "metadata" jsonb,
    "vector_content" text,
    "vector" vector(1024),
    "created_at" timestamptz NOT NULL DEFAULT now(),
    "updated_at" timestamptz NOT NULL DEFAULT now()
);
-- 设置表与关键字段注释，便于后续维护
COMMENT ON TABLE "university_profiles" IS '高校画像主表，用于存储高校基础信息、标签及向量嵌入';
COMMENT ON COLUMN "university_profiles"."name_cn" IS '高校中文名称';
COMMENT ON COLUMN "university_profiles"."name_en" IS '高校英文名称';
COMMENT ON COLUMN "university_profiles"."alias" IS '高校别名';
COMMENT ON COLUMN "university_profiles"."country" IS '所属国家或地区';
COMMENT ON COLUMN "university_profiles"."is_double_first_class" IS '是否为双一流院校';
COMMENT ON COLUMN "university_profiles"."is_project_985" IS '是否为985工程院校';
COMMENT ON COLUMN "university_profiles"."is_project_211" IS '是否为211工程院校';
COMMENT ON COLUMN "university_profiles"."is_qs_top100" IS '是否进入QS全球大学排名前100';
COMMENT ON COLUMN "university_profiles"."rank_qs" IS 'QS最新排名';
COMMENT ON COLUMN "university_profiles"."overall_score" IS '综合评分';
COMMENT ON COLUMN "university_profiles"."metadata" IS '附加元数据(JSON)，记录数据来源、同步信息等';
COMMENT ON COLUMN "university_profiles"."vector_content" IS '用于向量化的文档内容，预处理后的文本数据';
COMMENT ON COLUMN "university_profiles"."vector" IS '高校语义向量，维度与 embedding 配置保持一致';
-- 建立常用索引，加速查询
CREATE UNIQUE INDEX "idx_university_profiles_name_cn" ON "university_profiles" ("name_cn")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_university_profiles_metadata" ON "university_profiles" USING GIN ("metadata")
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_university_profiles_vector_ivfflat" ON "university_profiles" USING ivfflat ("vector" vector_cosine_ops) WITH (lists = 100)
WHERE "vector" IS NOT NULL
    AND "deleted_at" IS NULL;
CREATE INDEX "idx_university_profiles_flags" ON "university_profiles" (
    "is_double_first_class",
    "is_project_985",
    "is_project_211",
    "is_qs_top100"
)
WHERE "deleted_at" IS NULL;
CREATE INDEX "idx_university_profiles_country" ON "university_profiles" ("country")
WHERE "deleted_at" IS NULL;
-- 创建向量内容全文搜索索引（可选）
CREATE INDEX "idx_university_profiles_vector_content" ON "university_profiles" USING gin(to_tsvector('english', "vector_content"))
WHERE "vector_content" IS NOT NULL AND "deleted_at" IS NULL;