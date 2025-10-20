# 消息通知与异步队列设计方案

本文档描述 WhaleHire 平台在“简历解析完成”“岗位匹配完成”等关键节点向钉钉群发送通知的后端实现方案，重点覆盖整体架构、模块划分、消息队列设计、配置要求与实施步骤，为后续开发与联调提供依据。

## 1. 背景与目标

- 业务需要在多个耗时流程结束后及时告知招聘团队，避免人工轮询。
- 现有流程多为同步调用，缺少统一通知组件，易导致耦合与重发困难。
- 引入可靠的异步消息队列，确保通知可追踪、可重试，并便于扩展到企业微信等渠道。

目标：构建一套轻量但可扩展的通知子系统，支持事件抽象、模板化渲染与钉钉群机器人投递，保证消息可观测与幂等。

## 2. 总体架构概览

```
Resume/Matching Usecase → NotificationUsecase.PublishEvent()
        ↓
   NotificationEventRepo (落库)
        ↓
   MQ Producer (Redis Stream)
        ↓
┌─────────────────────────────────────┐
│ Notification Worker (消费 + 幂等)  │
│   ↘ Sender Adapter (DingTalk)       │
└─────────────────────────────────────┘
        ↓
  钉钉群机器人（签名 + 重试）
```

关键能力：

- 领域事件抽象：统一通知触发条件与负载。
- 异步队列：缓冲高峰流量，提供失败重试。
- Worker：读取消息，模板渲染并调用外部发送器。
- 监控与告警：记录投递状态、错误码与重试次数。

## 3. 模块与目录规划

- `backend/domain/notification.go`：集中定义 `NotificationEvent`、`NotificationChannel`、状态枚举与聚合根，并提供 Ent 模型到领域实体的转换方法。
- `backend/db/notificationevent.go`：Ent Schema，存储事件数据、重试次数、最后状态。
- `backend/internal/notification/usecase/`：提供 `PublishEvent`、`MarkDelivered` 等应用服务。
- `backend/internal/notification/repo/`：实现事件表与状态表的读写。
- `backend/internal/notification/templates/`：钉钉消息模板（Markdown/纯文本，支持变量）。
- `backend/internal/notification/adapter/dingtalk/`：封装 webhook、签名、发送与限流。
- `backend/internal/notification/worker/`：消费通知事件，调用模板与 Sender，处理重试。
- `backend/internal/notification/config.go`：加载钉钉 webhook、签名 key、默认重试策略等配置；若后续需要跨模块复用，再抽取至 `internal/config`。
- `backend/internal/queue/`：定义队列接口、Redis Stream 实现、统一 producer/consumer。
- `backend/internal/resume/usecase/` & `backend/internal/matching/usecase/`：在业务流程结束时注入 `NotificationUsecase` 发布事件。
- `backend/pkg/dingtalk/`：抽象钉钉通用工具（签名、timestamp、HTTP 请求）。

## 4. 消息流转流程

1. 业务 usecase 触发：简历解析或匹配成功后调用 `NotificationUsecase.PublishEvent(ctx, event)`。
2. Usecase 校验事件幂等键（如 `resume_id + job_id + event_type`），落库并写入 Redis Stream。
3. Worker 订阅 Redis Stream（消费者组），按批读取消息，加载模板渲染钉钉内容。
4. 发送成功：更新事件状态为 `Delivered`，记录响应；失败则记录原因并安排重试。
5. 重试超限或异常：标记为 `Failed`，并触发内部告警（Prometheus 指标或日志采集）。

## 5. 消息队列设计

- **实现**：基于现有 Redis，使用 Redis Stream；若未来引入专业 MQ（如 Kafka、RabbitMQ），保持接口兼容。
- **Producer**：`backend/internal/queue/redis/producer.go` 封装 `XAdd`，支持批量写入与 TraceID。
- **Consumer**：`backend/internal/queue/redis/consumer.go` 以消费者组消费，配置 `MaxLen` 和 `Block`。
- **消息格式**：字段包含 `event_id`、`event_type`、`payload`（JSON）、`retry_count`、`trace_id`。
- **幂等处理**：Worker 在消费前查询事件状态，若已 `Delivered` 则 `XAck` 并跳过。
- **重试策略**：指数退避（如 1min / 5min / 30min），超过阈值写入死信流 `notification:deadletter`。
- **监控指标**：消息积压、重试次数、失败率，暴露到 `pkg/metrics`。

## 6. 钉钉消息发送适配

- 支持文本与 Markdown 模板，模板中使用 `${var}` 占位符，由 Usecase 提供变量。
- 发送前增加签名：timestamp + secret 做 HmacSha256，满足钉钉安全要求。
- 网络请求使用统一 `pkg/request` 工具；默认超时与重试（如 3 次，每次 2s 回退）。
- 提供 `SendOptions`：@特定手机号、是否加急、是否同步返回。
- 发送失败时，归类错误（客户端/网络/钉钉限制），用于重试与监控。

## 7. 数据模型与结构

- `notification_events`（Ent Schema）
  - `id`（UUID）、`event_type`、`channel`、`status`（Pending/Delivering/Delivered/Failed）
  - `payload`（JSON）、`template_id`、`target`（webhook/群 ID）
  - `retry_count`、`max_retry`、`last_error`、`trace_id`
  - `created_at`、`scheduled_at`、`delivered_at`、`updated_at`
- 可选扩展：`notification_channels` 存储渠道配置，便于多渠道共存。
- Domain 层暴露 `NotificationEvent`、`NotificationPayload` 结构体，避免直接暴露 Ent 模型。

## 8. 配置与部署要求

- 新增 `.env.example` 配置：
  - `DINGTALK_NOTIFICATION_WEBHOOK`
  - `DINGTALK_NOTIFICATION_SECRET`
  - `NOTIFICATION_MAX_RETRY`
  - `NOTIFICATION_QUEUE_STREAM=notification:events`
- `docker-compose.dev.yml`：复用现有 Redis，无需新增容器；如需独立实例，可添加命名空间或数据库索引。
- `cmd/main.go`：在依赖注入中增加 NotificationUsecase、Queue Producer、Worker 启动。
- Worker 启动支持后台 goroutine，并提供 `Shutdown` 钩子确保优雅退出。

## 9. 监控、日志与告警

- 指标：`notification_events_total`、`notification_delivery_duration_seconds`、`notification_retry_total`。
- 日志：统一使用 `pkg/log`，记录 TraceID、事件 ID、模板 ID。
- 告警：当 dead-letter 队列有积压或失败率超过阈值，触发内部告警（如钉钉机器人自监控群）。
- 审计：在 `backend/domain/audit.go` 可增加通知相关审计动作，借助现有审计流水。

## 10. 实施里程碑

1. **基础设施**：完成 Ent Schema、Repo、Queue 抽象，编写迁移文件并验证。
2. **业务集成**：在简历解析与岗位匹配 usecase 中发布事件，编写模板示例。
3. **Worker 与发送器**：实现钉钉 Sender、Worker 消费逻辑、幂等与重试。
4. **监控上线**：补充指标、日志、告警配置，完成 e2e 测试（模拟事件 → 钉钉收到消息）。
5. **迭代与扩展**：支持多渠道通知、后台管理界面、模板动态配置。

通过上述设计，可以在保持服务内聚与模块化的前提下，引入可靠的通知链路，为后续扩展 AI 审核、面试提醒等场景打下基础。
