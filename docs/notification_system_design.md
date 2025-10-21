# 消息通知系统总体设计

本文档基于仓库当前实现，对 WhaleHire 的钉钉通知子系统进行概览性说明，帮助研发了解现状、扩展点与运行要求。重点涵盖系统架构、模块划分、事件流向及数据模型。

## 1. 背景与目标

- 简历解析、岗位匹配、筛选任务等流程耗时且异步，业务需要在完成后及时向协作群同步结果，避免人工轮询。
- 平台需要统一的事件模型、可追踪的投递链路以及模板化的通知内容，便于后续扩展更多渠道。
- 当前实现聚焦于钉钉机器人投递，但整体架构已抽象生产者、消费者、事件仓储与发送适配器，支持后续扩容。

通过该子系统，我们希望在保持代码内聚的同时，实现事件幂等、队列缓冲、可配置渠道与可回查的通知履历。

## 2. 系统架构

```
业务用例 (简历/筛选等)
        │ PublishEvent()
        ▼
domain.NotificationUsecase
        │ 写库 + 推队列
        ▼
NotificationEventRepo (Ent)
        │
Redis Stream notification:events
        │
NotificationWorker (消费者组 notification-worker)
        │
DingTalkAdapter ─→ pkg/dingtalk ─→ 钉钉机器人
```

- 业务 usecase 注入 `NotificationUsecase`，构造 payload 并发布事件。
- Usecase 落库 `notification_events`，随后使用 Redis Stream 作为异步缓冲。事件记录的 `target` 会保存首个启用配置的 webhook，便于回溯。
- `NotificationWorker` 消费 Stream、查询事件数据、渲染模板并调用钉钉适配器。适配器会在投递阶段读取所有启用的钉钉配置，实现“一次触发、多群广播”。
- 投递成功后更新事件状态，失败则累加重试信息，由运营或后台任务触发 `RetryEvent` 重新入队。

## 3. 核心模块

- **领域模型（domain）**  
  `backend/domain/notification.go` 定义事件实体、四类 payload（简历解析、批量解析、岗位匹配、筛选任务）以及幂等键生成策略；`backend/domain/notification_setting.go` 现包含通知配置的唯一名称 `name`、分页查询参数与多通道接口协议，支撑一个渠道维护多条配置。

- **应用服务（usecase）**  
  `backend/internal/notification/usecase/notification.go` 负责幂等校验、选取启用的通知配置、写入事件并推送到 `notification:events` 流；`backend/internal/notification/usecase/notification_setting.go` 新增按名称 + 通道的唯一性校验、列表分页、按通道批量获取配置等能力，为多群推送提供数据源。

- **仓储层（repo）**  
  `backend/internal/notification/repo/notification_event.go` 与 `.../notification_setting.go` 基于 Ent 客户端读写 `notification_events` 与 `notification_settings` 表，支撑状态更新、重试计数、启用配置查询等场景；`notification_setting` 仓储已支持名称去重、按通道批量返回启用配置以及带条件的分页检索。

- **队列抽象（queue）**  
  `backend/internal/queue/queue.go` 定义 Producer/Consumer 接口，`backend/internal/queue/redis/redis.go` 基于 Redis Stream 实现发布、订阅、死信处理等能力；`backend/internal/queue_provider.go` 负责将 Redis 客户端注入为 Producer/Consumer。

- **发送适配器与模板**  
  `backend/internal/notification/adapter/dingtalk.go` 使用 `backend/pkg/dingtalk` 封装签名与 Markdown 消息发送，模板位于 `backend/templates/notification/*.tmpl`。适配器的 `broadcastMessage` 会拉取所有启用的钉钉配置、使用 3 个并发槽发送 Markdown，并内置 3 次指数退避重试与错误聚合，确保一次事件可覆盖多个群。

- **异步 Worker**  
  `backend/internal/notification/worker/notification_worker.go` 消费 `notification:events` 流、分发到适配器并更新状态；`backend/internal/notification/worker/servicer.go` 将 Worker 封装为 `service.Servicer`，在 `backend/cmd/main.go` 中与 HTTP 服务一同纳入生命周期管理。

- **业务集成点**  
  当前已在 `backend/internal/resume/usecase/resume.go` 与 `backend/internal/screening/usecase/screening.go` 等用例中注入通知能力，后续场景可复用 `domain.NotificationPayload` 接口快速接入；如需针对特定群定向通知，可在业务层构造多个 payload 并分别调用 `PublishEvent`。

- **运营接口**  
  `backend/internal/notification/handler/v1/notification_setting.go` 暴露 REST API，允许运营配置钉钉 webhook、重试次数、超时等信息，配套 `errcode` 做错误映射，并提供启用列表、按通道查询（`/by-channel/:channel`）与分页检索能力。

## 4. 消息流转与幂等

1. 业务用例组装具体 payload（例如 `domain.ResumeParseCompletedPayload`），调用 `PublishEvent`。
2. Usecase 基于 payload 生成幂等键（如 `resume_parse_<resume_id>`），若已有同一 Trace ID 的事件则跳过重复投递。
3. Usecase 查询启用的通知设置，使用第一个配置填充事件的 `channel`/`target` 字段并落库；真正的投递目标会在适配器层面再次读取全部启用配置，以支持多群广播。
4. 事件写入 `notification_events`，并以结构化字段推送至 Redis Stream `notification:events`。
5. `NotificationWorker` 以消费者组 `notification-worker` 读取消息，回查事件记录。若状态已是 `delivered` 会直接确认消息，避免重复发送。
6. Worker 调用 `DingTalkAdapter` 渲染模板，适配器通过 `broadcastMessage` 限流并发、执行最多 3 次重试，将 Markdown 消息推送到所有启用的钉钉群；发送成功后调用仓储 `MarkAsDelivered`，失败则通过 `IncrementRetryCount` 记录错误与次数。
7. Worker 无论成功或失败都会 `ACK` 当前 Stream 消息，防止 Redis 重复投递；失败的事件保留在数据库，后续可通过 `RetryEvent`（或运营界面）再次推送。
8. `PublishEventWithDelay` 支持设置计划执行时间，延迟事件会在字段 `scheduled_at` 中标记，由消费者在读取时判断是否到期。

## 5. 数据模型

- **notification_events**  
  核心字段包含 `id`, `event_type`, `channel`, `status`, `payload`(JSON), `template_id`, `target`, `retry_count`, `max_retry`, `timeout`, `last_error`, `trace_id`, `scheduled_at`, `delivered_at`, `created_at`, `updated_at`。迁移脚本 `backend/migration/000013_add_notification_tables.up.sql` 已创建索引以支撑按状态/类型检索。

- **notification_settings**  
  字段包括 `id`, `name`, `channel`, `enabled`, `dingtalk_config`(JSON，含 `webhook_url`/`token`/`secret`), `max_retry`, `timeout`, `description`, `created_at`, `updated_at`。同一通道可维护多条启用记录，通过唯一名称区分；API 支持按通道拉取全部配置、分页检索及关键词搜索。

- **Redis Stream 消息结构**  
  `event_id`, `event_type`, `channel`, `payload`, `template_id`, `target`, `trace_id`, `created_at` 等字段会被序列化后写入 Stream；Worker 会解析 JSON 字符串恢复为结构体。

## 6. 配置与运维

- **服务启动**：`backend/cmd/main.go` 使用 `worker.NewServicer` 将通知 Worker 加入统一的 `service.Service`，随主进程启动/停止；确保 Redis 在启动前可用。
- **渠道管理**：通过 `/api/v1/notification-settings` 系列接口管理钉钉配置，支持启用/停用、设置最大重试次数与超时时间。新增 `/api/v1/notification-settings/by-channel/:channel` 与分页查询接口，便于按渠道维护多条配置；Usecase/Adapter 会读取所有 `enabled=true` 的配置并群发。
- **模板更新**：Markdown 模板位于 `backend/templates/notification`，更新时需保证语法正确并覆盖主要字段。适配器启动时缓存模板，如需热更新目前需重启服务。
- **重试策略**：失败事件保留在数据库，通过 `notificationUsecase.RetryEvent` 再次推送至 Stream。后续若需要后台自动重试，可在 Worker 中引入延迟策略或使用定时任务扫描 `status=failed` 事件。
- **依赖组件**：当前仅依赖现有 Redis 服务；无需额外消息中间件或环境变量。钉钉凭据由数据库配置维护。

## 7. 后续演进方向

- 扩展新的渠道适配器（企业微信、邮件等），实现 `NotificationChannel` 多态化。
- 在 Worker 中补充失败重试调度、死信队列与 Prometheus 指标，增强可观测性，并记录各群投递结果（成功/失败）。
- 引入模板管理后台，实现模板在线编辑与多语言版本。
- 针对高频事件可增加批量消息合并或更精细的限流策略，避免钉钉机器人触发频控。

以上内容覆盖了当前通知系统的主要组成与运行机制，为后续功能扩展和稳定性建设提供参考。
