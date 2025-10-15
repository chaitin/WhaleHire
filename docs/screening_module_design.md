# 智能筛选模块设计方案（Screening）

本文档整理并固化前期沟通的设计方案，覆盖智能筛选业务模块的目标、架构设计、数据模型、接口设计、异步处理机制、监控与指标、错误处理与一致性以及实施计划。用于指导后续实现与联调，并为团队成员提供统一的参考。

## 1. 背景与目标

在大规模简历与岗位匹配场景下，智能筛选需要具备以下能力：
- 高并发接入与异步处理能力，避免阻塞前端请求和占用过多在线资源；
- 稳定可靠的数据落库与结果查询能力；
- 可观测与成本可控（Tokens、时长、费用）能力；
- 清晰的业务闭环：任务创建 → 异步执行 → 结果产出 → 指标统计 → 过程可视化。

目标：提供一个模块化的 Screening 业务组件，包含任务管理、结果查询、运行指标与监控，支持后续扩展与优化。

## 2. 模块与职责划分

- Domain（业务领域层）：定义业务实体、过滤器、请求与响应模型、接口契约（Usecase/Repo）。
- Usecase（应用服务层）：编排业务流程，组合仓储与外部服务（例如 MQ、LLM 工作流），对外提供面向业务的服务方法。
- Repo（数据访问层）：面向数据库模型进行 CRUD，完成实体转换与分页查询等。
- Handler（接口适配层）：HTTP 适配，参数校验、鉴权集成、统一响应与错误处理。
- Consts：定义枚举与常量。
- DB（Ent 模型）：持久化结构，生成器维护字段与关联。
- MQ/Worker：异步任务调度与执行，保证高吞吐与可扩展。
- Web 包：统一错误响应、分页、国际化与中间件。

## 3. 关键常量与枚举

定义于 `backend/consts/screening.go`：
- ScreeningTaskStatus：Pending｜Running｜Completed｜Failed
- ScreeningTaskResumeStatus：Pending｜Running｜Completed｜Failed
- MatchLevel：Excellent｜Good｜Fair｜Poor

均提供 `Values()` 与 `IsValid()` 方法用于校验与序列化。

## 4. 数据模型设计

### 4.1 数据库模型（Ent）

- ScreeningTask（`backend/db/screeningtask.go`）
  - 主要字段：
    - ID, DeletedAt
    - JobPositionID（注意：不是 JobProfileID）
    - CreatedBy
    - Status（Pending/Running/Completed/Failed）
    - DimensionWeights（map[string]interface{}，可选）
    - LlmConfig（map[string]interface{}，可选）
    - Notes
    - ResumeTotal/ResumeProcessed/ResumeSucceeded/ResumeFailed
    - AgentVersion
    - StartedAt/FinishedAt
    - CreatedAt/UpdatedAt

- ScreeningResult（`backend/db/screeningresult.go`）
  - 主要字段：
    - ID, DeletedAt
    - TaskID, JobPositionID, ResumeID
    - OverallScore, MatchLevel
    - DimensionScores（map[string]interface{}）
    - BasicDetail/EducationDetail/ExperienceDetail/IndustryDetail/ResponsibilityDetail/SkillDetail（建议以 JSON 字符串存储）
    - Recommendations, TraceID, RuntimeMetadata, SubAgentVersions
    - MatchedAt, CreatedAt, UpdatedAt

- ScreeningRunMetric（`backend/db/screeningrunmetric.go`）
  - 主要字段：
    - ID, DeletedAt
    - TaskID
    - AvgScore, Histogram（map[string]interface{}）
    - TokensInput/TokensOutput，TotalCost
    - CreatedAt/UpdatedAt

### 4.2 Domain 实体（建议方案）

为避免导入循环与类型耦合，Domain 层应以“领域可用”的轻量结构表达业务：

- ScreeningTaskEntity（用于 Repo 层与 Usecase 的内部交换）
  - 字段建议：
    - ID, JobPositionID, Status, Notes
    - ResumeTotal/ResumeSucceeded/ResumeFailed
    - CreatedBy, CreatedAt, UpdatedAt
    - StartedAt, FinishedAt（注：DB 为 FinishedAt，Domain 对外可称 CompletedAt）

- ScreeningTaskResumeEntity：TaskID, ResumeID, Status, ErrorMessage, ProcessedAt, CreatedAt, UpdatedAt

- ScreeningResultEntity：
  - TaskID, ResumeID, OverallScore
  - BasicInfoDetail/EducationDetail/ExperienceDetail/IndustryDetail/ResponsibilityDetail/SkillDetail（JSON 字符串）
  - CreatedAt, UpdatedAt

- ScreeningRunMetricEntity：
  - TaskID, TotalTokens, PromptTokens, CompletionTokens, TotalCost, ProcessingDuration
  - CreatedAt, UpdatedAt

### 4.3 对外展示模型（API 响应）

对外响应可解码上述 JSON 字段为结构体，避免上层重复解析：
- ScreeningResult：将各 *Detail 字段解析为结构体（例如 BasicMatchDetail/EducationMatchDetail 等）。
- ScreeningRunMetric：直接输出各指标。

为避免 Domain 引入 `pkg/eino/graphs/screening/types` 导致导入循环，建议：
1) Domain 层的响应结构使用通用 map 或自定义轻量结构；
2) 在 Handler 层（或 Usecase 层）完成到 `types` 的转换；
3) 或将 `types` 下的匹配详情结构抽取到独立的无业务依赖的包。

## 5. 接口与用例设计

### 5.1 Usecase 接口（建议签名）

```go
type ScreeningUsecase interface {
    CreateScreeningTask(ctx context.Context, req *CreateScreeningTaskReq) (*CreateScreeningTaskResp, error)
    GetScreeningTask(ctx context.Context, req *GetScreeningTaskReq) (*GetScreeningTaskResp, error)
    ListScreeningTasks(ctx context.Context, req *ListScreeningTasksReq) (*ListScreeningTasksResp, error)

    // 异步流程相关（可选）
    StartScreeningTask(ctx context.Context, req *StartScreeningTaskReq) (*StartScreeningTaskResp, error)

    GetScreeningResult(ctx context.Context, req *GetScreeningResultReq) (*GetScreeningResultResp, error)
    ListScreeningResults(ctx context.Context, req *ListScreeningResultsReq) (*ListScreeningResultsResp, error)

    GetScreeningMetrics(ctx context.Context, req *GetScreeningMetricsReq) (*GetScreeningMetricsResp, error)
}
```

### 5.2 Repo 接口（建议签名）

```go
type ScreeningRepo interface {
    CreateScreeningTask(ctx context.Context, task *ScreeningTaskEntity) (*ScreeningTaskEntity, error)
    GetScreeningTask(ctx context.Context, id uuid.UUID) (*ScreeningTaskEntity, error)
    ListScreeningTasks(ctx context.Context, filter *ScreeningTaskFilter) ([]*ScreeningTaskEntity, int, error)
    UpdateScreeningTask(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error

    CreateScreeningTaskResume(ctx context.Context, taskResume *ScreeningTaskResumeEntity) (*ScreeningTaskResumeEntity, error)
    GetScreeningTaskResume(ctx context.Context, taskID, resumeID uuid.UUID) (*ScreeningTaskResumeEntity, error)
    ListScreeningTaskResumes(ctx context.Context, filter *ScreeningTaskResumeFilter) ([]*ScreeningTaskResumeEntity, int, error)
    UpdateScreeningTaskResume(ctx context.Context, taskID, resumeID uuid.UUID, updates map[string]interface{}) error

    CreateScreeningResult(ctx context.Context, result *ScreeningResultEntity) (*ScreeningResultEntity, error)
    GetScreeningResult(ctx context.Context, taskID, resumeID uuid.UUID) (*ScreeningResultEntity, error)
    ListScreeningResults(ctx context.Context, filter *ScreeningResultFilter) ([]*ScreeningResultEntity, int, error)

    CreateScreeningRunMetric(ctx context.Context, metric *ScreeningRunMetricEntity) (*ScreeningRunMetricEntity, error)
    GetScreeningRunMetric(ctx context.Context, taskID uuid.UUID) (*ScreeningRunMetricEntity, error)
}
```

### 5.3 请求/响应模型（建议）

```go
type CreateScreeningTaskReq struct {
    JobPositionID uuid.UUID   `json:"job_position_id"` // 注意与 DB 一致
    ResumeIDs     []uuid.UUID `json:"resume_ids"`
    Notes         string      `json:"notes,omitempty"`
}

type CreateScreeningTaskResp struct { TaskID uuid.UUID `json:"task_id"` }

type GetScreeningTaskReq struct { TaskID uuid.UUID `json:"task_id"` }

type GetScreeningTaskResp struct {
    Task    *ScreeningTask         `json:"task"`
    Resumes []*ScreeningTaskResume `json:"resumes"`
    Metrics *ScreeningRunMetric    `json:"metrics,omitempty"`
}

type ListScreeningTasksReq struct {
    JobPositionID *uuid.UUID                  `json:"job_position_id,omitempty"`
    Status        *consts.ScreeningTaskStatus `json:"status,omitempty"`
    CreatedBy     *uuid.UUID                  `json:"created_by,omitempty"`
    StartTime     *time.Time                  `json:"start_time,omitempty"`
    EndTime       *time.Time                  `json:"end_time,omitempty"`
    Page          int                         `json:"page"`
    PageSize      int                         `json:"page_size"`
}

type ListScreeningTasksResp struct {
    Tasks      []*ScreeningTask `json:"tasks"`
    Total      int              `json:"total"`
    Page       int              `json:"page"`
    PageSize   int              `json:"page_size"`
    TotalPages int              `json:"total_pages"`
}

type StartScreeningTaskReq struct { TaskID uuid.UUID `json:"task_id"` }
type StartScreeningTaskResp struct { TaskID uuid.UUID `json:"task_id"` }

type GetScreeningResultReq struct { TaskID, ResumeID uuid.UUID }
type GetScreeningResultResp struct { Result *ScreeningResult }

type ListScreeningResultsReq struct {
    TaskID   *uuid.UUID
    ResumeID *uuid.UUID
    MinScore *float64
    MaxScore *float64
    Page     int
    PageSize int
}
type ListScreeningResultsResp struct {
    Results    []*ScreeningResult
    Total      int
    Page       int
    PageSize   int
    TotalPages int
}

type GetScreeningMetricsReq struct { TaskID uuid.UUID }
type GetScreeningMetricsResp struct { Metrics *ScreeningRunMetric }
```

## 6. 异步处理架构设计

### 6.1 流程说明

1) 前端/API 创建筛选任务（CreateScreeningTask），落库任务与需要处理的 ResumeIDs。
2) 通过消息队列（MQ）发布“任务开始”或“任务-简历处理”消息（可一简历一消息）。
3) Worker 消费消息，执行智能匹配流程：
   - 加载岗位信息与简历内容；
   - 调用 LLM/规则引擎，生成各维度匹配详情与评分；
   - 计算并存储 ScreeningResult（JSON 详情字符串 + 数值指标）；
   - 累计运行指标（Tokens、时长、成本），在 ScreeningRunMetric 中更新；
   - 更新任务进度（ResumeProcessed/Succeeded/Failed），任务结束时更新 Status=Completed 和 FinishedAt。
4) API 层提供查询：任务详情、结果列表、指标数据。

### 6.2 可靠性与可扩展性

- MQ：支持至少一次投递；消费者幂等处理（基于 TaskID+ResumeID 唯一键）；失败重试与 DLQ（死信队列）。
- Worker：水平扩展、多进程并发；支持速率限制与超时；支持断点续处理。
- 成本控制：统一记录 TokensInput/TokensOutput、TotalCost；支持批处理策略。

### 6.3 消息模型（示例）

```json
{
  "type": "screening.resume.process",
  "task_id": "<uuid>",
  "resume_id": "<uuid>",
  "job_position_id": "<uuid>",
  "context": {
    "agent_version": "vX.Y.Z",
    "llm_config": { "model": "gpt-4o-mini", "temperature": 0.2 }
  }
}
```

## 7. API 设计（REST）

> 统一前缀：`/api/v1/screening`

- POST `/tasks` 创建任务
  - Request：CreateScreeningTaskReq
  - Response：CreateScreeningTaskResp

- GET `/tasks/:task_id` 任务详情
  - Response：GetScreeningTaskResp

- GET `/tasks` 任务列表
  - Query：ListScreeningTasksReq（page/page_size、status、job_position_id、created_by、start_time/end_time）
  - Response：ListScreeningTasksResp

- POST `/tasks/:task_id/start` 启动任务（异步）
  - Request：StartScreeningTaskReq
  - Response：StartScreeningTaskResp

- GET `/results` 结果列表
  - Query：ListScreeningResultsReq（task_id/resume_id/min_score/max_score/page/page_size）
  - Response：ListScreeningResultsResp

- GET `/metrics/:task_id` 指标
  - Response：GetScreeningMetricsResp

### 7.1 统一响应与错误处理

基于 `backend/pkg/web`：
- 成功：`ctx.Success(data)`
- 失败：`ctx.Failed(httpStatus, errID, err)` 或返回 `BusinessErr/Err` 统一处理。
- 支持国际化（i18n）与错误码映射。

## 8. 监控与指标

- 运行指标：
  - TokensInput/TokensOutput、TotalCost、ProcessingDuration、AvgScore、Histogram。
- 过程监控：
  - 任务状态流转（Pending → Running → Completed/Failed）；
  - 成功/失败数量与原因（ErrorMessage）；
- 可观测性：
  - 日志、TraceID、错误追踪；
  - Prometheus 指标（待扩展）。

## 9. 错误处理与一致性

- 幂等性：以 (TaskID, ResumeID) 唯一索引保证重复消费可安全覆盖更新。
- 失败重试：Worker 层控制重试与退避策略，超过阈值入 DLQ；API 层返回可定位的错误信息。
- 事务与一致性：写入结果与运行指标尽量在同一事务或具备补偿机制，保证总量与进度一致。

## 10. 实施与修复计划

1) 类型与接口统一：
   - Domain 与 DB 对齐 JobPositionID 字段命名（替换 JobProfileID）。
   - Domain Usecase/Repo 请求与响应签名统一为本文档建议版本。
   - Domain 层不直接依赖 `pkg/eino/graphs/screening/types`，在 Handler/Usecase 层进行转换或抽取公共类型包。

2) 异步机制落地：
   - 选型并接入 MQ（如 NATS/Kafka/RabbitMQ），定义消息主题与格式；
   - 实现 Worker 消费逻辑、并发与幂等；
   - 增加运行指标的采集与聚合。

3) 监控与指标：
   - 打通日志与 Trace；
   - 暴露 Prometheus 指标（可选）；
   - 按任务维度的统计与看板。

4) API 文档与测试：
   - Swagger/OpenAPI 注释完善；
   - 单元测试与集成测试覆盖核心流程；
   - 基于 web 包的统一错误与分页测试。

## 11. 附录：枚举与校验

- ScreeningTaskStatus/ScreeningTaskResumeStatus/MatchLevel 的 `Values()` 与 `IsValid()` 用于表单校验与查询过滤。
- List 请求的 `page/page_size` 建议限制：page ≥ 1，1 ≤ page_size ≤ 100。
- 时间过滤采用 `StartTime/EndTime`（UTC）并在 Repo 层进行范围查询。

---

本文档为实施过程中的基线方案，后续如有调整（例如细化维度详情结构或引入新的指标项），请在 docs 目录中版本化更新并在 PR 中注明变更点。