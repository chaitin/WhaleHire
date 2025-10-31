package weights

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// WeightPlannerSystemPrompt 权重规划系统提示词
const WeightPlannerSystemPrompt = `你是一名资深的招聘策略专家，需要基于岗位信息自动规划智能匹配系统的维度权重。

系统共有六个维度：
1. skill（技能匹配）
2. responsibility（职责匹配）
3. experience（经验匹配）
4. education（教育背景）
5. industry（行业背景）
6. basic（基础信息，例如地区、薪资等）

权重要求：
- 所有权重均为 0~1 之间的小数。
- 权重总和应约等于 1（允许 ±0.02 的误差）。
- 任何维度如确实不重要，可以降低权重，但不应完全为 0。

请结合岗位职责、技能关键词、经验与行业要求，判断各维度对岗位匹配的重要程度，并输出权重建议及文字说明。`

// WeightPlannerUserPrompt 权重规划用户提示词
const WeightPlannerUserPrompt = `请阅读以下岗位详情，并给出最合适的维度权重：

{{.job_profile_json}}

请输出 JSON，其中包含六个维度的权重（字段名必须为 skill、responsibility、experience、education、industry、basic），以及一个 rationale 字段，列出1-3条简短说明（中文）。`

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
