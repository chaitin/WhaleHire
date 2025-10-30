# Agent 链路执行与 LLM 消耗观测方案

本文档提出一套可扩展、可复用的通用观测方案，用于记录任何基于 Eino `compose` 框架实现的 Agent/Chain 在调用大语言模型（LLM）时的 Token 消耗、输入输出摘要与执行错误。方案兼容 `backend/pkg/eino/chains` 下的岗位画像链路（`jobprofilegenerator`、`jobprofileparser`、`jobprofilepolisher`）以及未来新增的 Agent，帮助团队在成本、可观测性与合规审计之间取得平衡。

## 1. 背景与需求

- 多条 Agent 链路依赖大模型 `chat_model` 节点，但缺乏统一的 Token、耗时、费用等观测手段，导致成本不可控。
- 链路执行的输入参数、模型输出与错误散落在各层，调试或审计时难以追溯，线上告警缺少上下文。
- 现有 `screening` 模块的回调收集器高度定制，复用到其他 Agent 需要大量改造。
- 需要一套通用能力，既能无侵入地对接不同 Agent，又能适配多种模型类型、扩展成本计算策略，并满足各类数据安全要求。

## 2. 设计目标

- **统一抽象**：对模型调用与节点执行事件提供一致的数据结构与回调接口，避免重复造轮子。
- **低侵入**：Agent 链路仅需通过 `compose.Option` 注入回调，无需改写业务逻辑。
- **高可拓展性**：支持多 Agent、多模型节点、不同模型类型；可插拔 Collector，满足日志、监控、审计等多种需求。
- **数据安全**：提供脱敏 Hook，支持将原始数据与摘要数据分别落地，确保合规。
- **上下游打通**：采集数据可与任务 ID、Trace ID 等上下文信息关联，为监控告警、成本核算、问题排查提供统一入口。

## 3. 总体方案概览

```
Agent/Chain ──> Compose Node (chat_model / lambda / tool)
                 │
                 ▼
         回调适配层（UsageCallbacks + ExecutionCallbacks）
                 │
                 ▼
        Collector (内存聚合 / 外部上报 / 自定义实现)
                 │
        ┌────────┴────────┐
        ▼                 ▼
  运行指标持久化       日志、监控、审计
```

- 同一套回调可以在任意链路复用，只需指定 `agentName` 与需要监听的节点集合。
- Collector 可实现为内存聚合器、实时上报器或混合模式；默认实现聚合 Token 消耗与执行事件。
- Usecase 层或上游服务在执行链路前创建 Collector，执行结束后读取聚合结果并按需落库或上报。

## 4. 通用回调与 Collector 设计

### 4.1 模型消耗事件 (`UsageEvent`)

```go
type UsageEvent struct {
    AgentName        string
    NodePath         compose.NodePath
    ModelName        string
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    Latency          time.Duration
    StartedAt        time.Time
    FinishedAt       time.Time
    Extra            map[string]any
}

type UsageCollector interface {
    OnModelUsage(ctx context.Context, event *UsageEvent)
}
```

- `ModelUsageCallback` 实现 `callbacks.Callback`。
- 在 `OnEnd` 阶段读取 `model.TokenUsage` 与模型名称，构建 `UsageEvent`。
- `Extra` 可用于携带模型配置、场景信息等自定义数据。

### 4.2 执行事件 (`ExecutionEvent`)

```go
type ExecutionPhase string

const (
    ExecutionPhaseStart   ExecutionPhase = "start"
    ExecutionPhaseSuccess ExecutionPhase = "success"
    ExecutionPhaseError   ExecutionPhase = "error"
)

type ExecutionEvent struct {
    AgentName   string
    NodeName    string
    Phase       ExecutionPhase
    Input       any
    Output      any
    Err         error
    TraceID     string
    StartedAt   time.Time
    FinishedAt  time.Time
    Extra       map[string]any
}

type ExecutionCollector interface {
    OnExecutionEvent(ctx context.Context, event *ExecutionEvent)
}
```

- `ExecutionCallback` 可注入任意节点（例如 `input_processing`、`chat_model`、`output_processing`）。
- 默认在 `OnStart`、`OnEnd`、`OnError` 中触发，记录节点输入、输出与错误。
- 提供 `Redactor` Hook 统一脱敏：

```go
type Redactor func(node string, phase ExecutionPhase, payload any) any
```

### 4.3 默认 Collector

```go
type AggregatedUsage struct {
    AgentName        string
    NodeName         string
    ModelName        string
    Calls            int
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    TotalLatency     time.Duration
}

type AggregatedExecution struct {
    AgentName string
    NodeName  string
    Success   int
    Failed    int
    LastError string
}

type DefaultCollector struct {
    usageEvents     []*UsageEvent
    executionEvents []*ExecutionEvent
    usageIndex      map[string]*AggregatedUsage
    executionIndex  map[string]*AggregatedExecution
    mu              sync.RWMutex
}
```

- 同时实现 `UsageCollector` 与 `ExecutionCollector` 接口。
- 提供 `Snapshot()`、`FindUsageByAgent()`、`FindExecutionByAgent()` 等方法，方便 Usecase 获取结果。
- Collector 内部可配置 `Redactor`、`MetadataProvider` 等策略函数，统一处理敏感信息与上下文注入。

## 5. Compose 注入方式

为避免各链路重复配置，提供通用 Option：

```go
func WithModelUsage(agent string, collector UsageCollector, paths ...compose.NodePath) compose.Option
func WithExecutionLogging(agent string, collector ExecutionCollector, opts ...ExecutionOption) compose.Option
```

- `paths` 未传入时，默认监听 `compose.NewNodePath(agent, "chat_model")`。
- `ExecutionOption` 支持：
  - `WithNodes(nodes ...string)`：指定需要监听的节点名称。
  - `WithRedactor(redactor Redactor)`：注入脱敏逻辑。
  - `WithMetadata(k string, v any)`：额外附加上下文信息。
- 链路在构建阶段按需调用上述 Option，即可完成回调注入。

## 6. 集成模式

### 6.1 Usecase 层通用流程

1. 初始化 Collector（可选自定义实现以对接日志系统）。
2. 构建链路时通过 `ChainOption` 注入 `WithModelUsage` 与 `WithExecutionLogging`。
3. 执行链路，等待结果。
4. 通过 Collector 读取聚合数据：
   - 若需落库，写入运行指标表或执行日志表。
   - 若需上报监控，将指标转换为 Prometheus/StatsD 指标。
   - 若需调试，输出日志或写入审计系统。

### 6.2 接入示例：岗位画像链路

岗位画像链路可将 `agentName` 设置为：

- `job_profile_generator`
- `job_profile_parser`
- `job_profile_polisher`

在构建链路时注入：

```go
chain.AppendOption(
    usage.WithModelUsage(agentName, collector),
    usage.WithExecutionLogging(agentName, collector,
        usage.WithNodes("input_processing", "chat_model", "output_processing"),
        usage.WithRedactor(redactor),
    ),
)
```

其他业务链路（如简历解析、意图识别）可复用同样的模式，仅需更换 `agentName` 与监听节点。

## 7. 数据模型与输出格式

### 7.1 事件字段约定

| 字段                | 含义                            | 示例                                 |
| ------------------- | ------------------------------- | ------------------------------------ |
| `agent_name`        | 链路或 Agent 名称               | `job_profile_generator`              |
| `node_path`         | Eino 节点路径                   | `job_profile_generator.chat_model`   |
| `model_name`        | 实际调用的模型                  | `gpt-4o`                             |
| `prompt_tokens`     | 输入 Token 数                   | `1024`                               |
| `completion_tokens` | 输出 Token 数                   | `256`                                |
| `total_tokens`      | 总 Token 数                     | `1280`                               |
| `latency_ms`        | 模型调用耗时                    | `823`                                |
| `started_at`        | 调用起始时间（UTC）             | `2024-07-21T10:01:02Z`               |
| `finished_at`       | 调用结束时间（UTC）             | `2024-07-21T10:01:02.823Z`           |
| `phase`             | 执行阶段（start/success/error） | `success`                            |
| `input_snapshot`    | 输入摘要（脱敏后）              | `{prompt: "...(截断)", meta: {...}}` |
| `output_snapshot`   | 输出摘要（脱敏后）              | `{result: {...}}`                    |
| `error_message`     | 错误说明（仅失败时填充）        | `"parse json failed"`                |
| `metadata`          | 扩展字段（请求 ID、渠道等）     | `{trace_id: "...", channel: "web"}`  |

### 7.2 输出载体建议

- Usecase 返回值可附带事件或聚合结果，以便上层调用方感知成本。
- 日志/监控采集时默认使用聚合数据，原始事件仅用于审计或问题排查。
- 对输入/输出需执行脱敏与截断，敏感字段（姓名、邮箱、手机号等）采用掩码或散列处理。

## 8. 扩展与兼容性

- **多模型类型**：可通过额外回调支持 Embedding、图像模型等，Collector 只需增加对应字段。
- **跨 Agent 共享**：同一请求可复用 Collector，通过 `metadata` 区分业务上下文。
- **成本核算**：提供 `CostCalculator` 接口，根据模型单价或外部计费服务计算费用。
- **错误回溯**：利用 `ExecutionPhaseError` 与 `UsageEvent` 中的指标，快速定位失败节点。
- **可观测性集成**：结合 Trace ID/Span ID，对接 OpenTelemetry，实现全链路追踪。
- **实时告警**：将聚合指标转化为 Prometheus 指标，设置 Token 消耗、错误率、耗时等阈值。

## 9. 实施步骤

1. **基础设施落地**：创建 `backend/pkg/eino/callbacks/usage` 包，实现 `UsageEvent`、`ExecutionEvent`、默认 Collector。
2. **回调注入封装**：提供 `WithModelUsage`、`WithExecutionLogging` 等 Option，封装节点注册流程。
3. **链路改造**：在各 Agent 构造函数中加入可选参数，允许注入上述 Option。
4. **Usecase 集成**：在业务层创建 Collector，执行链路后读取指标并输出到日志/监控/数据库。
5. **数据落库（可选）**：依据下述表结构设计落库方案，实现运行指标与执行日志的持久化。
6. **监控告警（可选）**：结合 Prometheus/APM，将聚合数据转化为指标并配置阈值。
7. **渐进迁移**：逐步将现有 `screening` 回调收集器迁移到通用方案，统一观测口径。

## 附录：PostgreSQL 表结构设计（可选）

为支撑多业务场景的数据持久化，推荐使用通用的 Agent 指标与执行日志表。下述设计可直接用于岗位画像场景，也可通过 `agent_name` 区分其他 Agent。

### 1. `agent_run_metrics`

| 字段名              | 类型            | 说明                                      |
| ------------------- | --------------- | ----------------------------------------- |
| `id`                | `UUID`          | 主键，默认 `gen_random_uuid()`            |
| `agent_name`        | `VARCHAR(64)`   | Agent/Chain 标识                          |
| `scenario`          | `VARCHAR(64)`   | 业务场景（如 `job_profile`、`screening`） |
| `job_reference_id`  | `UUID`          | 关联任务/请求 ID（可选）                  |
| `model_name`        | `VARCHAR(128)`  | 模型名称                                  |
| `total_calls`       | `INTEGER`       | 调用次数                                  |
| `prompt_tokens`     | `INTEGER`       | 累计输入 Token                            |
| `completion_tokens` | `INTEGER`       | 累计输出 Token                            |
| `total_tokens`      | `INTEGER`       | 总 Token 数                               |
| `total_latency_ms`  | `INTEGER`       | 累计耗时（毫秒）                          |
| `avg_latency_ms`    | `INTEGER`       | 平均耗时（毫秒）                          |
| `success_calls`     | `INTEGER`       | 成功次数                                  |
| `failed_calls`      | `INTEGER`       | 失败次数                                  |
| `total_cost`        | `NUMERIC(18,6)` | 总成本（按需计算）                        |
| `metadata`          | `JSONB`         | 额外信息（请求来源、模型配置等）          |
| `deleted_at`        | `TIMESTAMPTZ`   | 删除时间                                  |
| `created_at`        | `TIMESTAMPTZ`   | 创建时间，默认 `NOW()`                    |
| `updated_at`        | `TIMESTAMPTZ`   | 更新时间，默认 `NOW()`                    |

索引建议：

- `idx_agent_run_metrics_agent_scenario` (`agent_name`, `scenario`)
- `idx_agent_run_metrics_job_reference` (`job_reference_id`)
- `idx_agent_run_metrics_created_at` (`created_at`)

示例建表 SQL：

```sql
CREATE TABLE agent_run_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_name VARCHAR(64) NOT NULL,
    scenario VARCHAR(64) NULL,
    job_reference_id UUID NULL,
    model_name VARCHAR(128) NOT NULL,
    total_calls INTEGER NOT NULL DEFAULT 0,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    total_latency_ms INTEGER NOT NULL DEFAULT 0,
    avg_latency_ms INTEGER NOT NULL DEFAULT 0,
    success_calls INTEGER NOT NULL DEFAULT 0,
    failed_calls INTEGER NOT NULL DEFAULT 0,
    total_cost NUMERIC(18,6) NOT NULL DEFAULT 0,
    metadata JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agent_run_metrics_agent_scenario
    ON agent_run_metrics (agent_name, scenario);

CREATE INDEX idx_agent_run_metrics_job_reference
    ON agent_run_metrics (job_reference_id);

CREATE INDEX idx_agent_run_metrics_created_at
    ON agent_run_metrics (created_at);
```

### 2. `agent_execution_logs`

| 字段名             | 类型           | 说明                              |
| ------------------ | -------------- | --------------------------------- |
| `id`               | `UUID`         | 主键，默认 `gen_random_uuid()`    |
| `agent_name`       | `VARCHAR(64)`  | Agent/Chain 标识                  |
| `scenario`         | `VARCHAR(64)`  | 业务场景（可选）                  |
| `job_reference_id` | `UUID`         | 关联任务/请求 ID（可选）          |
| `trace_id`         | `VARCHAR(128)` | Trace ID / 请求 ID                |
| `node_name`        | `VARCHAR(128)` | 节点名称（如 `input_processing`） |
| `phase`            | `VARCHAR(16)`  | `start` / `success` / `error`     |
| `input_snapshot`   | `JSONB`        | 输入摘要（脱敏后）                |
| `output_snapshot`  | `JSONB`        | 输出摘要（脱敏后）                |
| `error_message`    | `TEXT`         | 错误信息（失败时填充）            |
| `started_at`       | `TIMESTAMPTZ`  | 节点开始时间                      |
| `finished_at`      | `TIMESTAMPTZ`  | 节点结束时间                      |
| `metadata`         | `JSONB`        | 其他上下文信息（渠道、操作者等）  |
| `deleted_at`       | `TIMESTAMPTZ`  | 删除时间                          |
| `created_at`       | `TIMESTAMPTZ`  | 创建时间，默认 `NOW()`            |
| `updated_at`       | `TIMESTAMPTZ`  | 更新时间，默认 `NOW()`            |

索引建议：

- `idx_agent_execution_logs_agent_node` (`agent_name`, `node_name`)
- `idx_agent_execution_logs_trace` (`trace_id`)
- `idx_agent_execution_logs_job_reference` (`job_reference_id`)
- `idx_agent_execution_logs_created_at` (`created_at`)

示例建表 SQL：

```sql
CREATE TABLE agent_execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_name VARCHAR(64) NOT NULL,
    scenario VARCHAR(64) NULL,
    job_reference_id UUID NULL,
    trace_id VARCHAR(128) NULL,
    node_name VARCHAR(128) NOT NULL,
    phase VARCHAR(16) NOT NULL,
    input_snapshot JSONB NULL,
    output_snapshot JSONB NULL,
    error_message TEXT NULL,
    started_at TIMESTAMPTZ NULL,
    finished_at TIMESTAMPTZ NULL,
    metadata JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agent_execution_logs_agent_node
    ON agent_execution_logs (agent_name, node_name);

CREATE INDEX idx_agent_execution_logs_trace
    ON agent_execution_logs (trace_id);

CREATE INDEX idx_agent_execution_logs_job_reference
    ON agent_execution_logs (job_reference_id);

CREATE INDEX idx_agent_execution_logs_created_at
    ON agent_execution_logs (created_at);
```

### 3. 注意事项

- **脱敏策略**：Collector 中统一处理敏感字段，避免原文落库。必要时将原文写入独立的受控存储，并设置访问权限。
- **生命周期管理**：执行日志数据量大，建议按时间或任务状态定期归档/清理；运行指标可长期保留用于趋势分析。
- **Ent 集成**：如使用 Ent，可在 `backend/ent/schema` 新增对应 Schema，字段类型与索引与上述设计保持一致。
- **扩展字段**：若需记录模型超参、费用明细等，可通过 `metadata` 存储 JSON，减少迁移成本。

通过以上通用方案，可以在保持链路编排简洁的同时，为所有 Agent 链路建立统一的执行观测与成本分析能力，为后续的监控告警、故障排查以及成本优化提供坚实基础。
