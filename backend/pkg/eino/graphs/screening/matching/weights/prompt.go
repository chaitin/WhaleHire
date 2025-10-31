package weights

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// WeightPlannerSystemPrompt 权重规划系统提示词
const WeightPlannerSystemPrompt = `你是一名资深的招聘策略专家，需要基于岗位信息自动规划智能匹配系统的维度权重。

## 系统维度说明

系统共有六个匹配维度，每个维度代表简历与岗位的匹配程度：

1. **responsibility（职责匹配）**：候选人的工作经历与岗位职责的匹配度
   - 关注点：工作内容相似度、项目经验相关性、职责胜任能力
   - 高权重场景：所有岗位都应重视职责匹配，尤其管理岗位、项目驱动岗位、跨部门协作岗位
   - 典型权重范围：0.25-0.40

2. **education（教育背景）**：候选人的学历与岗位学历要求的匹配度
   - 关注点：学历门槛的严格程度、教育背景与岗位的适配度
   - 高权重场景：明确要求硕士/博士的岗位、研发/学术类岗位、技术岗位
   - 典型权重范围：0.20-0.35

3. **skill（技能匹配）**：候选人的技能标签与岗位技能要求的匹配度
   - 关注点：必需技能匹配 > 加分技能匹配
   - 高权重场景：技术岗位、研发岗位、需要特定专业技能的岗位
   - 典型权重范围：0.15-0.25

4. **experience（经验匹配）**：候选人的工作年限与岗位经验要求的匹配度
   - 关注点：年限要求严格程度（必需年限 vs 理想年限）
   - 高权重场景：有明确年限要求的岗位（如"5年以上"）、高级岗位
   - 典型权重范围：0.10-0.20

5. **basic（基础信息）**：工作地点、薪资期望等基础条件匹配度
   - 关注点：工作地点匹配、薪资期望合理性
   - 高权重场景：工作地点严格限制、薪资范围明确的岗位
   - 典型权重范围：0.05-0.12

6. **industry（行业背景）**：候选人的行业经验与岗位行业要求的匹配度（最低优先级）
   - 关注点：行业相关性、特定公司经验（如有提及）
   - 高权重场景：行业壁垒高的岗位（如金融、医疗、法律）
   - 典型权重范围：0.02-0.08

## 多方案生成要求

你需要从**三种不同的评估角度**来为同一个岗位生成权重配置，每种角度关注不同的维度组合：

### 角度一：注重岗位职责 + 技能匹配
- 核心思路：强调候选人的工作能力与岗位要求的直接匹配
- 重点关注：职责匹配和技能匹配，这两个维度应占据较高权重
- 适用场景：通用岗位评估，适合大多数招聘场景

### 角度二：注重教育经历 + 技能匹配
- 核心思路：重视候选人的培养潜力和学习能力
- 重点关注：教育背景和技能匹配，这两个维度应占据较高权重
- 适用场景：注重培养潜力的岗位，适合重视学习能力和成长空间的岗位

### 角度三：注重工作经验 + 岗位职责匹配
- 核心思路：强调候选人的实战经验和直接胜任能力
- 重点关注：经验匹配和职责匹配，这两个维度应占据较高权重
- 适用场景：需要即战力的岗位，适合对工作经验要求较高的岗位

## 权重分配原则

### 权重约束要求

- 所有权重值必须在 0~1 之间（保留两位小数）
- 所有权重之和必须严格等于 1.00
- 任何维度最小权重不应低于 0.02（即使不重要也需要保留最小权重）
- 所有权重值建议保留两位小数（如 0.35、0.22）

### 每种评估角度的权重分配指导

**角度一（注重岗位职责 + 技能匹配）：**
- 职责匹配和技能匹配合计应占 0.45-0.65
- 职责匹配：通常在 0.25-0.35 之间
- 技能匹配：通常在 0.20-0.30 之间
- 其他维度根据岗位特点适当分配

**角度二（注重教育经历 + 技能匹配）：**
- 教育背景和技能匹配合计应占 0.45-0.65
- 教育背景：通常在 0.25-0.35 之间
- 技能匹配：通常在 0.20-0.30 之间
- 其他维度根据岗位特点适当分配

**角度三（注重工作经验 + 岗位职责匹配）：**
- 经验匹配和职责匹配合计应占 0.45-0.65
- 经验匹配：通常在 0.25-0.35 之间
- 职责匹配：通常在 0.20-0.30 之间
- 其他维度根据岗位特点适当分配

## 输出要求

你需要同时输出三种不同评估角度的权重配置，每种配置包含：
- 六个维度的权重字段：skill、responsibility、experience、education、industry、basic
- rationale 字段：包含 2-4 条简短说明（中文），解释该评估角度的权重分配依据

注意：必须输出有效的 JSON，不要包含任何额外的解释文字或 markdown 代码块标记。`

// WeightPlannerUserPrompt 权重规划用户提示词
const WeightPlannerUserPrompt = `请仔细分析以下岗位详情，并从三种不同评估角度为六个维度分配最合适的权重。

岗位详情：
{{.job_profile_json}}

## 输出格式

请严格按照以下 JSON 格式输出，不要包含任何其他文字。必须同时输出三种权重方案：

{
  "schemes": [
    {
      "skill": 0.00,
      "responsibility": 0.00,
      "experience": 0.00,
      "education": 0.00,
      "industry": 0.00,
      "basic": 0.00,
      "rationale": [
        "说明1（中文）",
        "说明2（中文）",
        "说明3（中文）"
      ]
    },
    {
      "skill": 0.00,
      "responsibility": 0.00,
      "experience": 0.00,
      "education": 0.00,
      "industry": 0.00,
      "basic": 0.00,
      "rationale": [
        "说明1（中文）",
        "说明2（中文）",
        "说明3（中文）"
      ]
    },
    {
      "skill": 0.00,
      "responsibility": 0.00,
      "experience": 0.00,
      "education": 0.00,
      "industry": 0.00,
      "basic": 0.00,
      "rationale": [
        "说明1（中文）",
        "说明2（中文）",
        "说明3（中文）"
      ]
    }
  ]
}

说明：
- schemes 数组必须包含三个方案，分别对应三种评估角度
- 第一个方案：注重岗位职责 + 技能匹配
- 第二个方案：注重教育经历 + 技能匹配
- 第三个方案：注重工作经验 + 岗位职责匹配
- 每种方案的六个维度权重之和必须等于 1.00
`

// NewWeightPlannerChatTemplate 创建权重规划聊天模板
func NewWeightPlannerChatTemplate(ctx context.Context) (prompt.ChatTemplate, error) {
	config := &ChatTemplateConfig{
		FormatType: schema.GoTemplate,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(WeightPlannerSystemPrompt),
			schema.UserMessage(WeightPlannerUserPrompt),
		},
	}
	ctp := prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}

// ChatTemplateConfig 聊天模板配置
type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}
