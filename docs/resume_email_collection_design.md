# 简历邮箱采集总体设计

本文档描述 WhaleHire 平台新增的“多邮箱简历采集”能力，覆盖需求背景、系统架构、核心模块、数据模型、关键流程与运维要点，以指导后续的后端、前端与基础设施实现。

## 1. 背景与目标

- 招聘团队通常将候选人简历发送到多个招聘邮箱（公司官方、职位专用、第三方渠道等），目前需人工导出再导入平台，效率低且易遗漏。
- 现有系统仅支持手动上传或爬虫导入，缺乏对邮箱场景的自动化处理，导致协作人员无法及时获得解析结果。
- 目标是在平台内统一配置多个 IMAP/POP3 邮箱，定时同步新邮件，自动抓取简历附件并触发现有的解析、上传人归属和岗位画像绑定流程，实现“零侵扰”的简历入库体验。

## 2. 业务范围与角色

- **运营/招聘负责人**：在前端页面管理邮箱配置（新增、编辑、开启/停用、测试连接），指定“简历上传人”和“岗位画像”。
- **自动采集服务**：按照配置定时同步邮件，解析附件并调用现有简历解析/入库用例。
- **系统管理员**：维护默认同步频率、凭证加密密钥、监控指标，处理异常告警。

不在本期范围：邮件回复、手动标记已读/未读、解析失败后的自动通知（列入后续演进）。

## 3. 系统架构概览

```
          +------------------+
          | 邮件服务器 (IMAP/POP3)|
          +---------+--------+
                    |
          定时任务 / 事件触发
                    |
          +---------v---------+
          | Mailbox Syncer    |  <-- 读取配置、调用协议适配器
          +---------+---------+
                    |
             结构化邮件数据
                    |
          +---------v---------+
          | Attachment Router |  <-- 去重、识别简历附件、存储原文件
          +---------+---------+
                    |
             简历解析请求 (队列/用例)
                    |
          +---------v---------+
          | Resume Parser UC  |  <-- 复用现有解析、上传人/岗位关联
          +---------+---------+
                    |
          +---------v---------+
          | 数据库 / 日志 / 指标 |
          +------------------+
```

- 架构延续现有“用例 + 仓储 + 适配器”的清洁分层，新增邮箱配置仓储、协议适配器与调度 Worker。
- 邮箱同步输出的结构化邮件会进入统一的附件路由模块，与当前手动上传入口共享解析流程，避免重复实现。

## 4. 核心模块划分

- **MailboxSettingUsecase**：提供 CRUD、分页查询、状态切换、连接测试能力，校验协议参数、上传人、岗位画像的存在性。
- **MailboxSettingRepo (Ent)**：读写 `resume_mailbox_settings` 表，负责敏感字段加解密、分页与唯一性校验（协议 + 邮箱地址唯一）。
- **CredentialVault**：基于现有密钥管理方案（优先使用 KMS 或配置的 AES 密钥）对邮箱密码、OAuth token 等敏感信息进行加密存储。
- **Scheduler & Sync Worker**：按配置的 `sync_interval` 或平台默认频率调度任务，支持按邮箱独立暂停/继续；Worker 负责拉取邮件、增量追踪与错误重试。
- **Protocol Adapter 层**：实现 `IMAPAdapter` 与 `POP3Adapter`，遵循统一接口 `Fetch(startCursor) ([]MailItem, nextCursor, error)`，封装 TLS、收件箱选择、UID/UIDL 管理等细节。
- **Attachment Router**：对邮件正文、附件进行 MIME 解析，过滤非简历附件，执行去重（基于文件哈希 + 邮件 Message-ID），并将原文件保存到对象存储或 `tmp/` 后转交解析用例。
- **ResumeParseIntegration**：调用已有 `ResumeUploadUsecase`，传入上传人（邮箱配置指定）与岗位画像 ID，保持与手动上传一致的校验与日志记录。
- **Monitoring & Alerting**：暴露同步成功/失败次数、延迟、剩余待解析邮件数量；失败超过阈值时写入告警（后续可复用通知系统推送）。

## 5. 数据模型设计

- **resume_mailbox_settings**

  - `id` (UUID)
  - `name`：任务名称，页面展示用
  - `email_address`：邮箱账号，唯一约束（同协议 + 账号不可重复）
  - `protocol`：`imap` / `pop3`
  - `host` / `port` / `use_ssl` / `folder`（IMAP 专用）
  - `auth_type`：`password` / `oauth`
  - `encrypted_credential`：密文（JSON，含密码或 OAuth token）
  - `uploader_id`：上传人用户 ID（复用简历上传逻辑）
  - `job_profile_id`：岗位画像 ID
  - `sync_interval_minutes`：自定义同步频率（为空则使用平台默认）
  - `status`：`enabled` / `disabled`
  - `last_synced_at` / `last_error` / `retry_count`
  - `deleted_at`：软删除时间，遵循平台表结构规范
  - `created_at` / `updated_at`

- **resume_mailbox_cursors**

  - `id` (UUID)
  - `mailbox_id`：外键
  - `protocol_cursor`：IMAP UIDVALIDITY + UID 或 POP3 UIDL 列表摘要
  - `last_message_id`：冗余字段帮助排查
  - `deleted_at`：软删除标记，删除配置时与主表同步软删
  - `created_at` / `updated_at`

- **resume_mailbox_statistics**

  - `id` (UUID)
  - `mailbox_id`：外键
  - `date`：按日分区（零点对齐），支持查看趋势
  - `synced_emails`：当日成功同步邮件数
  - `parsed_resumes`：成功解析入库的简历附件数量
  - `failed_resumes`：解析失败数量
  - `skipped_attachments`：被去重或过滤的附件数量
  - `last_sync_duration_ms`：最近一次同步耗时（毫秒），用于分析性能
  - `deleted_at`：软删除标记，删除配置时与主表同步软删
  - `created_at` / `updated_at`

## 6. 关键流程说明

### 6.1 邮箱配置管理

1. 前端调用 `/api/v1/resume-mailboxes` 相关接口创建配置，填写协议、服务器参数、上传人、岗位画像。
2. 后端 Usecase 校验账号唯一性、检查上传人/岗位画像是否有效，使用 `CredentialVault` 加密敏感字段后写库。
3. 连接测试接口会即时拉起协议适配器执行登录、列出 1 条邮件验证权限，不会写入游标。
4. 状态切换接口控制 `enabled` 字段，Scheduler 根据该字段决定是否调度同步任务。

### 6.2 邮件同步与解析

1. Scheduler 按邮箱配置生成同步任务，并将任务加入 `resume-mailbox-sync` 队列（或使用现有 Worker 框架）。
2. Sync Worker 读取配置与游标，创建协议适配器：
   - **IMAP**：使用 UID 搜索 `SINCE lastSync`，并持久化最新 UIDVALIDITY+UID。
   - **POP3**：读取 UIDL 列表，忽略已处理 UIDL，持久化最新集合摘要。
3. Worker 将新邮件转为统一结构 `MailItem`（含 Message-ID、发送时间、附件列表、正文片段）。
4. Attachment Router 遍历附件，识别常见简历格式（PDF、DOC/DOCX、HTML、图片等），对每个有效附件：
   - 计算哈希用于去重（与已有解析记录对比）。
   - 将原文件暂存至对象存储/本地临时目录。
   - 生成解析任务并调用 `ResumeUploadUsecase.UploadFromMailbox`（新接口），传入上传人、岗位画像、来源标识（邮箱 ID）。
5. 解析成功后更新 `last_synced_at`、重置 `retry_count`，失败则记录 `last_error` 并根据重试策略延迟下一次尝试。
6. Worker 根据本次处理结果累加 `resume_mailbox_statistics`（当日分区），并记录最后一次同步耗时，为前端展示与运维分析提供数据。

### 6.3 去重与异常处理

- 邮件级去重依据 `Message-ID` + 附件文件哈希，避免重复解析同一封邮件或附件。
- 若解析失败（格式错误、解析服务异常），将附件保留，并在 `resume_mailbox_logs` 中记录原因；达到最大重试（默认 3 次）后保持失败状态等待人工处理。
- 同步层异常（网络、身份失败）将指数退避重试，同时触发告警。

## 7. 对外接口规划

- `GET /api/v1/resume-mailboxes`：分页列表，支持按状态、协议、邮箱地址搜索。
- `POST /api/v1/resume-mailboxes`：创建配置。
- `PUT /api/v1/resume-mailboxes/{id}`：更新配置（禁止修改协议与邮箱地址，可提示删除后重建）。
- `POST /api/v1/resume-mailboxes/{id}/test-connection`：即时连接测试。
- `POST /api/v1/resume-mailboxes/{id}/toggle`：启用/禁用。
- `POST /api/v1/resume-mailboxes/{id}/sync-now`：手动触发一次同步（用于排查）。
- `GET /api/v1/resume-mailboxes/{id}/stats?range=30d`：按日返回统计数据（同步邮件数、成功入库、失败、去重数等），支持范围聚合。
- 前端页面复用现有“简历收集配置”模块，新增邮箱 tab，展示同步统计（最后同步时间、同步数、失败次数等）。

所有接口返回信息需覆盖上传人姓名、岗位画像名称、同步统计，便于与页面图示对齐。

## 8. 安全与合规要求

- 邮箱凭证必须使用 AES-256-GCM（或对接外部 KMS）加密存储，密钥来源于环境变量并由运维管理。
- Worker 在内存解密后应尽快覆盖敏感变量，避免写入日志；日志中仅记录掩码后的邮箱地址。
- 支持配置 IMAP/POP3 SSL/TLS；如需兼容明文端口，需在配置层显式标记并提示风险。
- 限制同步频率下限（例如不小于 5 分钟）防止触发邮箱服务商限流。

## 9. 运维与监控

- 新增 Prometheus 指标：
  - `resume_mailbox_sync_total{mailbox_id,status}`：同步批次数。
  - `resume_mailbox_attachment_total{mailbox_id,outcome}`：附件处理结果。
  - `resume_mailbox_last_sync_timestamp{mailbox_id}`：最后成功同步时间。
- 结合 `resume_mailbox_statistics` 表产出日报/周报，向运营展示各邮箱的解析成功率与去重比例。
- 日志：按邮箱 ID 切分，记录开始/结束时间、附件数、失败原因。
- 告警：当 `last_synced_at` 超过阈值（例如 24 小时）或连续失败超过 N 次时，通过通知系统推送运维告警。

## 10. 上线与迁移计划

1. 数据库迁移：添加 `resume_mailbox_settings`、`resume_mailbox_cursors` 表，生成 Ent Schema 与迁移 SQL。
2. 后端实现 Usecase、Repo、Handler、Worker、协议适配器，并编写集成测试（可基于 dockerized mail server 模拟）。
3. 前端新增邮箱配置页面、表单与列表 UI，复用岗位/上传人选择组件。
4. 灰度上线：先配置内部测试邮箱验证流程，确认解析链路与去重正常，再开放给运营团队。
5. 编写运行手册，说明凭证管理、常见错误处理、日志定位方法。

## 11. 后续演进方向

- 支持 OAuth 认证（如 Microsoft 365、Gmail），减少明文密码存储。
- 引入邮件正文解析，将正文关键字段（候选人姓名、联系方式）与附件结果交叉校验，提高准确率。
- 增加解析失败自动通知（钉钉/邮箱）与手动标记“已处理”功能，打通异常闭环。
- 支持多语言模板与国际化字段，为海外团队引入多语言邮箱。
- 与已有爬虫任务统一调度面板，提供采集任务健康度大盘。

以上设计为后续实现提供方向性指导，可在开发阶段根据实际情况细化接口协议与数据结构。
