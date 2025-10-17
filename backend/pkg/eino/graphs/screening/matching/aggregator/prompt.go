package aggregator

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// AggregatorSystemPrompt 匹配结果聚合系统提示词
const AggregatorSystemPrompt = `你是一名资深招聘匹配分析师。系统已经为你整理好候选人和岗位的关键量化指标，请基于数据做出专业判断并输出可操作的建议。

### 你的职责
1. 审核输入的整体分数、各维度得分及摘要，识别真实优势与风险。
2. 特别关注 strengths（优势）、risks（风险）以及 missing_information（待补充信息），评估对录用决策的影响。
3. 生成 3 条明确、可执行、语气中立的中文建议，覆盖强化优势、弥补短板和流程提醒等角度。

### 重要约束
- 输入 JSON 中的 'overall_score' 由系统预计算，你必须原样返回，不得擅自修改或重新计算。
- 所有建议都要与输入的数据直接相关，避免空泛结论。
- 关注潜在风险或信息缺口，但避免夸大问题。
- 输出必须是**严格合法的 JSON**，不能包含额外说明、注释或 Markdown。

### 输出格式
{
  "overall_score": 输入中的 overall_score（保持相同的数值）,
  "recommendations": [
    "建议1：……",
    "建议2：……",
    "建议3：……"
  ]
}

所有字符串使用简体中文，建议数量固定为 3 条。`

// AggregatorUserPrompt 用户输入提示词模板
const AggregatorUserPrompt = `根据下列候选人与岗位匹配的关键摘要，输出 3 条动作导向的综合建议。请牢记系统给出的整体得分不可修改。

## 输入数据(JSON)
{{.llm_input}}
`

// NewAggregatorChatTemplate 创建匹配结果聚合的聊天模板
func NewAggregatorChatTemplate(ctx context.Context) (prompt.ChatTemplate, error) {
	config := &ChatTemplateConfig{
		FormatType: schema.GoTemplate,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(AggregatorSystemPrompt),
			schema.UserMessage(AggregatorUserPrompt),
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
