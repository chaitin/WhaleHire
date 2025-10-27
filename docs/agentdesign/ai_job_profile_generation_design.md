# AI 岗位画像生成功能设计

## 1. 背景与目标

- 现有系统支持手动填写岗位画像，字段较多、编写成本高，难以快速批量产出标准化内容。
- 企业希望输入一句“想要招聘的数据分析师”即可得到完善的岗位职责、任职要求、加分项，同时生成可直接落库的岗位记录。
- 目标是在保持人工编辑模式的同时，引入 AI 辅助链路：提示语润色 → 岗位画像生成 → 用户确认或继续编辑 → 保存。

## 2. 范围与不在范围

- **范围内**：后端 AI 流程设计、接口定义、领域模型扩展、前端交互改造、Description 字段写入规范、测试与监控。

## 3. 用户旅程

1. 用户进入“AI 岗位画像生成”页签，输入一句话需求。
2. 点击“一键润色”，后台返回结构化 Prompt 建议与关键提纲，前端展示可继续编辑。
3. 用户确认润色后的 Prompt，点击“生成内容”。
4. 后台根据 Prompt 生成岗位画像结构体 + 三段式文本描述，前端渲染预览。
5. 用户如需调整，可回退修改 Prompt 或直接编辑字段；确认后点击“保存内容”。
6. 前端调用原有 `POST /api/v1/job-profiles` 接口创建岗位，Description 字段写入 AI 生成文本。

## 4. 总体方案

- 新增 **Prompt 润色链** 与 **岗位画像生成链**，均基于现有 Eino + OpenAI 适配层实现。
- 领域用例扩展：`JobProfileUsecase` 增加 `PolishPrompt` 与 `GenerateByPrompt` 两个能力，复用现有 `ParseJobProfile` 服务完成结构化解析。
- 接口层新增 `POST /api/v1/job-profiles/ai/polish` 与 `POST /api/v1/job-profiles/ai/generate`，统一走用户鉴权；生成接口返回描述文本、结构化字段与原始 Prompt。
- 前端保留现有列表与历史模块，新增 AI 交互输入区、润色结果编辑区、生成内容预览区。
- Description 字段统一存储为 Markdown 风格三段，便于前端展示和后续导出。

## 5. 后端设计

### 5.1 模块关系

- `backend/pkg/eino/chains/jobprofilepolisher`：Prompt 润色链，产出结构化 Prompt。
- `backend/pkg/eino/chains/jobprofilegenerator`：在润色结果基础上给出岗位三段描述及解析输入。
- `backend/internal/jobprofile/service`：
  - 新增 `JobProfilePromptService`（封装润色链）；
  - 扩展现有 `JobProfileParserService`，提供将生成链输出转为领域对象的辅助方法。
- `backend/internal/jobprofile/usecase/job_profile.go`：
  - 新增方法 `PolishPrompt(ctx, req)`、`GenerateByPrompt(ctx, req)`；
  - 利用 `BuildDescriptionText(responsibilities, requiredSkills, bonusSkills)` 生成 Description。
- `backend/internal/jobprofile/handler/v1`：
  - 新增 `ai.go`，挂载 `/api/v1/job-profiles/ai` 路由。

### 5.2 领域模型扩展

```go
type PolishJobPromptReq struct {
    Idea string `json:"idea" validate:"required"` // 用户初始输入
}

type PolishJobPromptResp struct {
    PolishedPrompt     string   `json:"polished_prompt"`
    SuggestedTitle     string   `json:"suggested_title"`
    ResponsibilityTips []string `json:"responsibility_tips"`
    RequirementTips    []string `json:"requirement_tips"`
    BonusTips          []string `json:"bonus_tips"`
}

type GenerateJobProfileReq struct {
    Prompt         string  `json:"prompt" validate:"required"`
    DepartmentID   *string `json:"department_id,omitempty"` // 生成时可提前绑定部门
}

type GenerateJobProfileResp struct {
    Profile             *ParseJobProfileResp            `json:"profile"`
    DescriptionMarkdown string                          `json:"description_markdown"`
}
```

### 5.3 服务流程

1. `PolishPrompt`
   - 校验 idea。
   - 调用 polisher 链生成润色结果，转为 `PolishJobPromptResp`。
   - 记录日志 `ai.job_profile.polish`，并在日志中记录提示条目数量。
2. `GenerateByPrompt`
   - 校验 prompt。
   - 调用生成链产出结构化的 `ParseJobProfileResp` 以及职责/任职要求/加分项的文本化建议。
   - 在链内根据建议和结构化数据拼装 Markdown 描述并返回。
   - Markdown 文本遵循三段式模板：
     ```
     岗位职责：
     - ...
     任职要求：
     - ...
     加分项：
     - ...
     ```
   - 拆分技能列表，标记 required/bonus，返回给前端。

### 5.4 接口设计

| 接口                               | 方法 | 请求体                  | 响应体                   | 说明                             |
| ---------------------------------- | ---- | ----------------------- | ------------------------ | -------------------------------- |
| `/api/v1/job-profiles/ai/polish`   | POST | `PolishJobPromptReq`    | `PolishJobPromptResp`    | Prompt 润色，需用户登录          |
| `/api/v1/job-profiles/ai/generate` | POST | `GenerateJobProfileReq` | `GenerateJobProfileResp` | 基于 Prompt 生成岗位画像，不落库 |

> 最终保存仍调用现有 `POST /api/v1/job-profiles`，前端传入 `GenerateJobProfileResp.Profile` 中的结构字段，并将 `DescriptionMarkdown` 填入 `CreateJobProfileReq.Description`。

### 5.5 AI 提示词设计要点

- **润色链**：明确要求输出 JSON，包含 `polished_prompt`、`suggested_title` 及三类提示数组（职责/任职/加分），每类控制在 1~2 条聚焦的改进建议，数组允许为空但字段不可缺失。
- **生成链**：输入为润色后的 Prompt，输出单行 JSON，顶层包含 `profile`（遵循解析链 schema）与 `sections{responsibilities[],requirements[],bonuses[]}`，链路内部基于这些信息拼装 Markdown 描述并返回给调用方。
- 统一设置 `ResponseFormat=json_object`，沿用现有 OpenAI 工厂与配置项。

### 5.6 错误处理与幂等

- 链执行失败时返回 502 并提示“AI 服务暂时不可用，请稍后重试”。
- 对用户输入做长度限制（例如 5~500 个汉字），防止 prompt 注入。
- 记录请求/响应摘要但脱敏，Sensitive 字段如 idea 不落库，只写入结构化日志。

## 7. 配置与部署

- 复用 `config.GeneralAgent.LLM`；如需区分模型，可新增 `GeneralAgent.JobProfile.*` 节点。
- 默认超时：润色 8s、生成 15s。通过配置文件暴露。
- 新增 `AI_JOB_PROFILE_POLISH_TIMEOUT`、`AI_JOB_PROFILE_GENERATE_TIMEOUT` 环境变量示例到 `backend/.env.example` 与运维文档。

## 8. 安全与权限

- 所有 AI 接口继续走用户鉴权中间件；仅具备岗位管理权限的角色可调用。
- 防止 prompt 注入：在系统层面禁止出现对外部功能的系统命令，例如对 prompt 做黑名单词匹配。
- 日志脱敏：仅记录哈希后的用户 ID、耗时、token 数，避免输出完整岗位内容。

## 10. 测试计划

- **单元测试**：用链路 mock 验证 usecase 参数校验、Description 拼装、技能拆分逻辑。
- **集成测试**：在沙箱环境使用真实模型或 fixture 响应，校验接口契约（JSON Schema）。
- **前端测试**：至少补 UI 交互自测，验证状态切换、异常提示、保存流程。
- **回归测试**：确保手动创建入口不受影响，特别是 `CreateJobProfileReq` 校验逻辑。

## 11. 风险与缓解

- 模型输出格式不稳定 → 加强提示、增加 JSON 解析兜底（重试/手动修复）。
- 描述过长导致存储字段超限 → 在生成链 Prompt 中限制每段 10 条以内，并在保存前截断。
- 生成内容与业务需求不符 → 允许用户在预览区手动编辑；提供“重新生成”按钮。
- LLM 成本不可控 → 记录 token，设置每日调用上限（配置化）。

---

本设计方案确保在现有手动流程上平滑叠加 AI 体验，保障结构化数据一致性，并为后续扩展留出模型与提示语的可配置空间。
