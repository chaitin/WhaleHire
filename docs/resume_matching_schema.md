# 智能岗位画像与简历匹配数据库设计（PostgreSQL + pgvector 可选，UUID 主键）

本设计面向产品原型中的“智能匹配”功能（创建匹配任务、查看匹配结果列表与详情、分页下载等），并与当前智能匹配 Agent 代码路径 `/backend/pkg/eino/graphs/screening` 的输出结构对齐，可扩展、可追溯、可复现。

目标：

- 对岗位画像与简历匹配进行任务化管理（Task）；
- 存储每次匹配的候选集与细粒度的维度结果（技能/职责/经验/教育/行业/基础信息）；
- 支持模型/权重/规则的版本化与复现；
- 提供向量表以支持语义检索与预筛（可选，pgvector）。

与现有表的兼容性：

- 已存在的岗位/职责/技能/学历等表（`job_position`、`job_responsibility`、`job_skill`、`job_education_requirement`、`job_experience_requirement`、`job_industry_requirement`）以及简历相关表（`resumes`、`resume_education`、`resume_experience`、`resume_skill`、`resume_project`、`resume_document_parses`）保持不变；
- 新增匹配相关的任务、结果、配置与向量表，通过外键连接到既有实体。

---

## 1. 匹配任务表 `screening_tasks`

用于管理一次岗位与候选集的匹配任务。

```sql
CREATE TABLE screening_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_position_id UUID NOT NULL REFERENCES job_position(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending/running/completed/failed
    dimension_weights JSONB,                       -- 维度权重：与 types.DimensionWeights 对齐
    llm_config JSONB,                              -- LLM/Embedding 配置：provider/model/temperature/top_p 等
    notes TEXT,
    resume_total INT DEFAULT 0,                    -- 任务内候选总数
    resume_processed INT DEFAULT 0,                -- 已处理数
    resume_succeeded INT DEFAULT 0,                -- 成功匹配数
    resume_failed INT DEFAULT 0,                   -- 失败数
    agent_version VARCHAR(50),                     -- 例如 智能匹配Agent的版本号
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP                          -- 软删除时间（与 Ent SoftDeleteMixin 保持一致）
);

CREATE INDEX idx_screening_tasks_job ON screening_tasks(job_position_id);
CREATE INDEX idx_screening_tasks_status ON screening_tasks(status);
```

说明：

- `selection_strategy` 用于记录筛选简历的快照规则，确保任务可复现；
- `dimension_weights` 与 `/pkg/eino/graphs/screening/types/types.go` 中 `DimensionWeights` 字段一致；
- 任务级别的处理量指标便于列表页概览计数与进度条显示。

---

## 2. 任务候选关系表 `screening_task_resumes`

用于记录某个匹配任务包含的简历集合以及处理状态。

```sql
CREATE TABLE screening_task_resumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES screening_tasks(id) ON DELETE CASCADE,
    resume_id UUID NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'queued', -- queued/processing/done/failed/skipped
    error_message TEXT,
    ranking INT,                                  -- 排名（任务内）
    score NUMERIC(6,2),                           -- 综合分快照（任务维度）
    processed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP,
    UNIQUE(task_id, resume_id)
);

CREATE INDEX idx_str_task ON screening_task_resumes(task_id);
CREATE INDEX idx_str_status ON screening_task_resumes(status);
CREATE INDEX idx_str_score ON screening_task_resumes(score DESC);
```

说明：

- 任务执行时的候选快照，便于“分页、排序、下载结果”；
- `score` 与 `ranking` 便于任务维度的前 N 名展示；
- 若需要复现某次任务的候选集，直接依赖此表即可。

---

## 3. 匹配结果表 `screening_results`

对齐 Aggregator 输出的结构，保存维度细节与整体评分，支持详情页展示与审计追踪。

```sql
CREATE TABLE screening_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES screening_tasks(id) ON DELETE CASCADE,
    job_position_id UUID NOT NULL REFERENCES job_position(id) ON DELETE CASCADE,
    resume_id UUID NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    overall_score NUMERIC(6,2) NOT NULL,         -- 聚合总分 0-100
    match_level VARCHAR(20),                      -- excellent/good/fair/poor（可由分数映射）
    dimension_scores JSONB,                       -- {skill:xx, responsibility:xx, experience:xx, education:xx, industry:xx, basic:xx}
    skill_detail JSONB,                           -- types.SkillMatchDetail 序列化
    responsibility_detail JSONB,                  -- types.ResponsibilityMatchDetail
    experience_detail JSONB,                      -- types.ExperienceMatchDetail
    education_detail JSONB,                       -- types.EducationMatchDetail
    industry_detail JSONB,                        -- types.IndustryMatchDetail
    basic_detail JSONB,                           -- types.BasicMatchDetail
    recommendations JSONB,                        -- string[]
    trace_id VARCHAR(100),                        -- 可与链路/日志系统关联
    runtime_metadata JSONB,                       -- 运行时元数据：tokens/cost/latency 等
    sub_agent_versions JSONB,                     -- 每个 Agent 版本快照
    matched_at TIMESTAMP DEFAULT now(),
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP,
    UNIQUE(task_id, resume_id)
);

CREATE INDEX idx_sr_task ON screening_results(task_id);
CREATE INDEX idx_sr_job_resume ON screening_results(job_position_id, resume_id);
CREATE INDEX idx_sr_overall_score ON screening_results(overall_score DESC);
```

说明：

- 维度详情 JSONB 与 Agent 输出结构完全一致，便于直接落库和可视化；
- `runtime_metadata` 保存调用成本与延迟，便于后续优化；
- `trace_id` 用于串联链路追踪系统或内部日志。

---

## 4. 运行指标汇总表 `screening_run_metrics`

用于任务完成后的整体统计，支持列表页展示概览与趋势分析。

```sql
CREATE TABLE screening_run_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES screening_tasks(id) ON DELETE CASCADE,
    avg_score NUMERIC(6,2),
    histogram JSONB,                 -- 例如分数段分布 {"0-59":n1, "60-74":n2, ...}
    tokens_input BIGINT,             -- 任务总输入 tokens
    tokens_output BIGINT,            -- 任务总输出 tokens
    total_cost NUMERIC(12,4),        -- 模型调用成本统计（如美元）
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP
);
```

---

## 8. 典型查询与使用方式

- 获取任务列表及统计：

```sql
SELECT id, job_position_id, status, resume_total, resume_processed, resume_succeeded, resume_failed,
       started_at, finished_at
FROM screening_tasks
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;
```

- 按综合分获取任务内 Top-N 候选：

```sql
SELECT r.resume_id, r.overall_score, r.match_level
FROM screening_results r
WHERE r.task_id = $1
ORDER BY r.overall_score DESC
LIMIT 10;
```

- 详情页加载维度结果：

```sql
SELECT overall_score, dimension_scores, skill_detail, responsibility_detail,
       experience_detail, education_detail, industry_detail, basic_detail,
       recommendations
FROM screening_results WHERE task_id = $1 AND resume_id = $2;
```

---

## 9. 与 Agent 输出结构的映射

- `types.JobResumeMatch` → `screening_results`：
- `OverallScore` → `overall_score`
- `SkillMatch/ResponsibilityMatch/...` → `*_detail`
- `Recommendations` → `recommendations`
- `TaskMetaData` → `screening_tasks` 中的快照 + `screening_results.trace_id/sub_agent_versions`
- `types.DimensionWeights` → `screening_tasks.dimension_weights`

---

## 10. 演进与扩展建议

- 增加维度：如“证书/开源贡献/语言能力”等，仅需在 `dimension_scores` 增加字段并新建相应 `*_detail` JSONB；
- 增加多岗位/多任务批处理：可在 `screening_tasks` 引入 `batch_id` 进行批次管理；
- 增加审核流：新增 `screening_review_records(task_id, resume_id, reviewer_id, decision, notes)`；
- 增加缓存表：将热门任务 Top-N 结果缓存到 Redis 以提升列表页性能；
- 增加数据治理：对 `screening_results` 定期归档，保留 `id/task_id/resume_id/overall_score` 索引与少量聚合维度，详情 JSONB 迁移到冷存储。

---

## 11. 与 Ent 的落地（建议命名）

- `ent/schema/screening_task.go` → `screening_tasks`
- `ent/schema/screening_task_resume.go` → `screening_task_resumes`
- `ent/schema/screening_result.go` → `screening_results`
- `ent/schema/screening_run_metric.go` → `screening_run_metrics`

字段类型与约束应与本文 DDL 一致，Edge 参考现有 `job_position`、`resumes`、`resume_job_applications` 的写法。

---

## 12. 安全与合规

- 对 `screening_results.*_detail` 中含有的文本要做脱敏处理（邮箱/电话等）；
- 建议启用行级访问控制：仅任务创建人、部门管理员可读取匹配结果；
- 记录模型调用元数据以便成本核算与合规审计。

---

## 13. 依赖与部署注意事项

- DDL 仅为参考，最终以 Ent Schema 生成的迁移为准；
- 大字段 JSONB 建议开启 TOAST 压缩与合适的 autovacuum 配置；
- 对高频查询（任务列表、Top-N）建立覆盖索引与 `partial index`（如 `status='completed'`）。

---

本文档与 `docs/job_profile_schema.md` 一起构成岗位画像与智能匹配的数据层设计，满足当前产品原型与后续演进需求。
