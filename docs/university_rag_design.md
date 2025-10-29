# 简历大学标签 RAG 方案设计

## 背景

- 现有简历解析链（`backend/pkg/eino/chains/resumeparser`）直接依赖 LLM 在提示词中判断 `university_type`，无法保证对于高校别名、缩写、英文译名的统一识别，导致 `双一流` / `985` / `211` 等标签漏判或误判。
- 团队已经整理了《dataset/merged_university_dataset_2025.csv》，覆盖国内双一流、985、211 以及海外 QS Top 100 的高校信息，却尚未进入解析链路。
- 项目规划在 Eino 框架下引入向量检索（pgvector），结合 RAG 提供高校权威标签，用于简历解析 AgentChain。

## 目标

- 建立高校知识库，将 CSV 数据清洗后落地 PostgreSQL + pgvector，并补充必要标签。
- 在解析链中，通过高校中文名称检索知识库，为每条教育经历补充 `双一流`、`985`、`211`、`QS Top 100` 等标签。
- 保持与现有 `ParsedResumeData` 结构兼容，同时向下游暴露更精细的标签信息，提升解析准确率与可追溯性。

## 现状分析

- **链路**：`ParserService.ParseResume` → Eino Chain → LLM 输出 JSON → `transform.go` 清洗 → `ResumeUsecase` 写库。
- **数据结构**：`domain.ParsedEducation` 目前仅有 `UniversityType consts.UniversityType` 字段，无法表达多标签组合。
- **向量基础设施**：仓库已配置 pgvector 镜像，`config.Config` 中预留 `Embedding`、`Retriever`、`Redis.Vector` 配置，但简历链未使用向量检索。
- **数据集**：CSV 包含英文名/中文名/地域/排名信息，需派生是否为 `双一流` / `985` / `211` / `QS Top 100`。

## 方案概览

- **数据侧**：新增 `university_profiles` Ent Schema，向量列使用 `vector(embedding_dimension)`，并建立布尔标签字段。
- **服务侧**：实现 `UniversityKnowledgeService`，包含 CSV/JSON 导入、向量嵌入与检索、标签推断、命中日志。
- **Eino 链路**：
  - 在 `backend/pkg/eino/chains` 下新增 `universityindexing` 工作流，复用 Eino `compose` 管线完成“文档生成 → 嵌入 → 自定义 pgvector Indexer”流程，索引器实现参考官方指南并放在 `backend/pkg/eino/components/indexer/pgvector`。
  - 新增 `universityretrieval` 链路，由“输入名称 → 嵌入 → 自定义 pgvector Retriever → 召回整理”构成，Retriever 组件按官方指南在 `backend/pkg/eino/components/retriever/pgvector` 实现，并支持 TopK/阈值配置。
- **Agent 改造**：在 `ParserService.ParseResume` 中加入高校标签补全步骤；Eino Chain 维持原解析逻辑，仅调整 prompt 以避免模型主观判定。
- **对外输出**：扩展 `domain.ParsedEducation` 新增 `UniversityTags []consts.UniversityBadge`；同时保留 `UniversityType` 作为向下兼容字段（基于主标签回填）。
- **运维**：提供 CLI/Make Task 完成知识库重建，支持定期更新，并跟踪命中率。

以下章节按模块展开。

## 数据建模与存储

### 表结构设计

- `university_profiles`
  - `id` (UUID, PK)
  - `name_cn` (TEXT, 唯一，支持简体中文全称)
  - `name_en` (TEXT, 可空)
  - `country` (TEXT)
  - `is_double_first_class` (BOOLEAN)
  - `is_project_985` (BOOLEAN)
  - `is_project_211` (BOOLEAN)
  - `is_qs_top100` (BOOLEAN)
  - `rank_2025` (INT, 可空)
  - `overall_score` (NUMERIC, 可空)
  - `metadata` (JSONB，记录数据源、维护时间、原始行)
  - `vector` (VECTOR，维度来自 `config.Embedding.Dimension`)
  - 索引：`UNIQUE(name_cn)`、`GIN(metadata)`、`IVFFLAT(vector vector_cosine_ops)` 等。

### CSV 清洗要点

- 将布尔字段统一转换为布尔值；对海外院校 `rank_2025 <= 100` 标记 `is_qs_top100`。
- 生成标准化名称：去除空格、括号内专业信息、统一全角/半角符号。
- 追加标准化字段（如去除“大学”“学院”的简称、去空格版本）写入 `metadata`，用于后续规则匹配。
- 输出 JSONB 元数据记录原行内容、导入时间，方便追溯。

### pgvector 初始化

- 在数据库迁移中启用扩展：`CREATE EXTENSION IF NOT EXISTS vector;`
- 通过 Ent Schema 增加 `field.Other("vector", &pgvector.Vector{}).Optional()`（需要自定义类型，建议复用现有 pgvector 驱动或引入 [pgvector-go](https://github.com/pgvector/pgvector-go)）。
- 为向量字段创建索引：`CREATE INDEX ... USING ivfflat (vector vector_cosine_ops) WITH (lists = 100);`

## 向量嵌入与知识库构建（基于 Eino）

### 嵌入模型选择

- 默认使用配置 `embedding.model_name`（当前为 `bge-m3`），维度与 `embedding.dimension` 保持一致。
- 若外部服务返回维度与配置不符，需要在导入阶段校验并更新配置/维度。
- 对中文名称采用同一嵌入模型，检索时可以将英文名称与中文合并输入（如 `"麻省理工学院 MIT Massachusetts Institute of Technology"`）提升 recall。

### 导入流程

1. `cmd/university_indexer/main.go` CLI 读取 CSV（允许传参指定数据源），输出 `[]*schema.Document`，每个 `Document` 的 `Content` 为高校标准描述，`MetaData` 存放 JSON 化属性。
2. 执行数据标准化（全称、简称、去空格版本等）并写入 `metadata`，同时将高校 ID 作为 `Document.ID`，方便后续幂等更新。
3. 调用新建的 `UniversityKnowledgeIndexingChain`（位于 `backend/pkg/eino/chains/universityindexing`）：链路包含
   - 自定义 `DocumentTransformer`（可选）补充内容；
   - `Embedding` 节点生成向量；
   - `PgvectorIndexer` 节点将结果批量写入 `university_profiles`。
4. `PgvectorIndexer` 在写入时负责 `UPSERT`（使用 `ON CONFLICT (name_cn)`），并更新向量列、布尔标签、`metadata`、`updated_at` 等字段。
5. CLI 过程中记录导入统计，必要时写入独立日志表（可选）便于回滚。

### Make/脚本与自动化

- 新增 `make university-index`：调用 CLI，支持 `CSV_PATH` 环境变量，便于本地快速导入/调试。
- 提供后台 API（如 `POST /api/v1/resume/universities/reindex`）触发重建流程：服务端调用 `UniversityKnowledgeIndexingChain`，并将任务放入异步队列/后台协程执行，返回任务 ID 供前端轮询进度。
- CI 或运维脚本在部署后可直接调用该 API 完成重构建，避免登陆服务器手动执行 CLI；CLI 方式作为兜底手段保留。

## 检索与标签判定流程

### 查询入口

- 新建 `backend/internal/resume/service/university_matcher.go`（或独立包 `backend/internal/university`）：
  - `MatchByName(ctx context.Context, schoolName string) (*domain.UniversityMatch, error)`
  - `BatchMatch(ctx context.Context, names []string) (map[string]*domain.UniversityMatch, error)`
  - 内部流程：
    1. 使用标准化规则（去空格、去“大学/学院”等后缀、比较全称/简称）对输入进行归一化，并在 `university_profiles` 做一次精确/模糊匹配；命中直接返回标签。
    2. 若未命中，则构造 `UniversityRetrievalChain` 的输入（包含原始名称与标准化名称），链路结构示例：
       - `Lambda`：准备 `{ "query": schoolNameNormalized }`
       - `Embedding`：生成查询向量（与索引时同模型）
       - `PgvectorRetriever`：调用自定义 Retriever，执行相似度检索并返回 `[]schema.Document`
       - `Lambda`：转换为 `domain.UniversityMatch`，过滤低于 `distance_threshold` 的结果
    3. 对候选结果应用中文/英文标准化、拼音首字母匹配等 heuristics，再决定最终标签与命中来源（exact/vector）。
    4. 返回命中高校、`[]UniversityBadge`、匹配置信度、命中来源、命中向量距离等信息。
  - 将每次匹配写入监控日志（`logger.Info` + 可选表 `university_match_logs`），便于评估召回率。

### 标签规则

- `UniversityBadge` 新增到 `consts` 包，例如：

  ```go
  type UniversityBadge string

  const (
      UniversityBadgeDoubleFirstClass UniversityBadge = "double_first_class"
      UniversityBadge985              UniversityBadge = "project_985"
      UniversityBadge211              UniversityBadge = "project_211"
      UniversityBadgeQSTop100         UniversityBadge = "qs_top100"
  )
  ```

- 主标签回填策略：
  - 若 `is_project_985` → `UniversityType = consts.UniversityType985`
  - 否则若 `is_project_211` → `UniversityType = consts.UniversityType211`
  - 否则 `UniversityType = consts.UniversityTypeOrdinary`
  - 同时在新字段 `UniversityTags` 写入所有 `badge`（海外 QS Top 100 仍将 `UniversityType` 置为 `ordinary`，但 tags 包含 `qs_top100`）。

### 批量处理

- 简历解析结果可能包含多条教育经历，需批量嵌入减少网络调用。
- 在 `MatchByName` 内部缓存短期结果（LRU/`sync.Map`）以避免同一次解析重复查询。

## AgentChain 改造

### Prompt 调整

- 更新 `backend/pkg/eino/chains/resumeparser/prompt.go`：
  - 将 `university_type` 字段描述为“由系统后处理填写”，提示模型不要猜测，可输出空字符串/`null`。
  - 确保模型仍准确抽取 `school` 名称，并保持中文原文。

### 解析链扩展

- 在 `transform.go` 中新增后处理函数 `enrichUniversityTags(ctx context.Context, result *ResumeParseResult) error`：
  - 遍历 `result.Educations`，调用 `UniversityKnowledgeService` 批量匹配。
  - 更新 `UniversityType`、新增 `UniversityTags` 字段。
  - 将匹配置信度写入 `ParsedEducation` 的新字段（例如 `UniversityMatchScore float64`）或记录到日志。
- 为保持 compose 结构清晰，可在 `ParserService.ParseResume` 中调用该函数，避免修改 Eino Chain 节点。

### Domain & Ent 变更

- `domain/ParsedEducation` 新增：
  ```go
  UniversityTags      []consts.UniversityBadge `json:"university_tags,omitempty"`
  UniversityMatchFrom string                   `json:"university_match_from,omitempty"` // exact / vector / manual
  UniversityMatchScore float64                 `json:"university_match_score,omitempty"`
  ```
- `db.ResumeEducation` Ent 实体新增：
  - `university_tags` (JSON 或 TEXT[])
  - `university_match_score` (FLOAT, 可空)
  - `university_match_from` (TEXT)
- `internal/resume/repo` 与 `usecase` 对应写库/读库逻辑更新。
- Swagger/前端 DTO（如有）同步调整。

## ParserService 改造

- 在 `ParseResume` 成功后：
  1. 调用 `knowledgeSvc.BatchMatch` 获取标签。
  2. 更新 `result.Educations`。
  3. 记录命中日志，便于排查未命中案例。
- 在失败时（未找到匹配或阈值过低）保留空标签，并在日志中标记。
- 若后续要支持手工校正，可新增接口写回 `university_profiles` 并同步刷新向量字段。

## pgvector 接入细节

- 自定义 Indexer：在 `backend/pkg/eino/components/indexer/pgvector` 下实现 `type Indexer struct{ ... }`，满足 `indexer.Indexer` 接口，参考官方 [Indexer 指南](https://www.cloudwego.io/zh/docs/eino/core_modules/components/indexer_guide/)。
  - `NewIndexer(ctx, cfg)` 中初始化数据库连接（可使用 `database/sql` + `pgx` 或直接复用项目的 `store` 包）。
  - `Index(ctx, docs []*schema.Document, opts ...indexer.Option)` 负责将向量、内容以及 `MetaData` 持久化到 `university_profiles`：
    ```go
    func (i *Indexer) Index(ctx context.Context, docs []*schema.Document, opts ...indexer.Option) error {
        tx, err := i.db.BeginTx(ctx, &sql.TxOptions{})
        ...
        for _, doc := range docs {
            payload := extractMetadata(doc.MetaData) // 布尔标签、国家、rank 等
            _, err = tx.ExecContext(ctx, `
                INSERT INTO university_profiles (id, name_cn, name_en, country,
                    is_double_first_class, is_project_985, is_project_211, is_qs_top100,
                    rank_2025, overall_score, metadata, vector, updated_at)
                VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())
                ON CONFLICT (name_cn)
                DO UPDATE SET
                    name_en = EXCLUDED.name_en,
                    country = EXCLUDED.country,
                    is_double_first_class = EXCLUDED.is_double_first_class,
                    is_project_985 = EXCLUDED.is_project_985,
                    is_project_211 = EXCLUDED.is_project_211,
                    is_qs_top100 = EXCLUDED.is_qs_top100,
                    rank_2025 = EXCLUDED.rank_2025,
                    overall_score = EXCLUDED.overall_score,
                    metadata = EXCLUDED.metadata,
                    vector = EXCLUDED.vector,
                    updated_at = NOW();
            `, id, doc.Content, payload.Vector, ...)
        }
        return tx.Commit()
    }
    ```
  - `Delete`, `Close` 等方法按需实现，支持更新/重建。
- 自定义 Retriever：在 `backend/pkg/eino/components/retriever/pgvector` 下实现 `type Retriever struct{ ... }`，满足 `retriever.Retriever` 接口，参考官方 [Retriever 指南](https://www.cloudwego.io/zh/docs/eino/core_modules/components/retriever_guide/)。
  - `NewRetriever(ctx, cfg)` 初始化数据库连接、阈值、TopK 等配置。
  - `Retrieve(ctx context.Context, query *schema.Query, opts ...retriever.Option) ([]*schema.Document, error)` 中执行向量相似度查询，可使用下方 SQL；将结果映射为 `schema.Document` 并附带 `Score`、`Metadata`。
  - 为兼容批量查询，可实现 `BatchRetrieve` 或在外层对多次调用做缓存。
- `UniversityKnowledgeIndexingChain` 在 `orchestration.go` 中注册该 Indexer，类似现有 `knowledgeindexing` 链，只是改为 `pgvector.NewIndexer`。
- `UniversityRetrievalChain` 在 `orchestration.go` 中注册 `pgvector.NewRetriever`，并通过 `compose` 将查询 Embedding 与检索节点串联，必要时在链尾增加 `Lambda` 节点整理召回结果。
- 为 PostgreSQL 启用 `CREATE EXTENSION IF NOT EXISTS vector;`，并在迁移中创建 `IVFFLAT` 或 `HNSW` 索引：`CREATE INDEX ... USING ivfflat (vector vector_cosine_ops) WITH (lists = 100);`
- Retriever 内部可复用 `database/sql` 查询，示例 SQL 如下：
  ```sql
  SELECT p.*, 1 - (p.vector <=> $1) AS score
  FROM university_profiles p
  WHERE p.vector <#> $1 < $distance
  ORDER BY p.vector <=> $1
  LIMIT $topK;
  ```

## 配置与开关

- `config.Config` 中新增：
  - `University` 模块配置：`TopK`、`DistanceThreshold`、`EnableVector`.
  - `Embedding` 模块需确保与高校知识库一致，必要时补充 `BatchSize`, `Timeout`.
- 在 `.env.example` 提示新增 PGVECTOR 连接参数。
- `docker-compose` 确认数据库镜像已启用 `pgvector` 扩展（当前镜像满足）。

## 测试与验证计划

- **单元测试**：
  - `UniversityKnowledgeService` 的名称标准化、标签推断。
  - `ParsedEducation` enrich 函数，验证空标签与命中情形。
- **集成测试**：
  - 使用内存/临时 PostgreSQL（或 Docker Testcontainer）插入部分高校数据，执行检索 API。
  - `ParserService.ParseResume` 在没有知识库时继续工作（降级路径）。
- **验收测试**：
  - 选取包含“清华大学”“北大”“MIT”等简历片段，验证标签输出。
  - 对常见别名（如“北航”“浙大”）进行回归。

## 运维与监控

- 导入 CLI 输出统计：总行数、去重后高校数、标准化名称覆盖数量、向量调用耗时。
- 匹配服务记录命中率、阈值分布，结合 Prometheus（如已有）或 Slog 采样。
- 支持 `make university-reindex` 定期同步数据；部署流水线中加入导入步骤。
- 数据更新前先备份原表（`CREATE TABLE ... AS SELECT ...`），必要时可回滚。

## 风险与缓解

- **向量维度不匹配**：导入时强校验，必要时降级为精确匹配。
- **命名歧义**：同名高校（国内外）可能混淆，需通过国家字段与标准化规则区分。
- **性能**：pgvector 检索需建立合适索引/并行，若 QPS 增加可引入缓存或旁路 Redis。
- **LLM 输出差异**：若模型输出学校英文名，需在标准化中处理大小写与空格。
- **数据维护**：CSV 更新需同步运行导入脚本，否则标签陈旧。

## 里程碑建议

1. **阶段 1**：完成 Ent Schema、数据库迁移、CSV → DB 导入 CLI，验证 pgvector 检索（~3 天）。
2. **阶段 2**：实现 `UniversityKnowledgeService`、扩展 `ParsedEducation`、完成解析链后处理（~4 天）。
3. **阶段 3**：完善配置、测试、日志与文档，联调前端/用例，准备上线策略（~3 天）。

## 后续展望

- 扩展至国际多语种别名、QS Top 500、更细粒度排名。
- 和候选人教育匹配、岗位画像推荐（已有 RAG 模块）共享知识库与检索层。
- 通过后台管理页提供知识库补充与标签纠错能力。
