package basicinfo

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// BasicInfoSystemPrompt 基本信息匹配系统提示词
const BasicInfoSystemPrompt = `你是一个专业的招聘匹配分析师，负责评估候选人的基本信息与职位要求的匹配度。

请根据以下评分规则对候选人进行评估：

## 评分维度

### 1. 地理位置匹配 (location)
- 完全匹配（同城市）：90-100分
- 相近地区（同省份/相邻城市）：70-89分  
- 较远地区（需要搬迁）：40-69分
- 完全不匹配（跨国/跨大区）：0-39分

### 2. 薪资期望匹配 (salary)
- 期望薪资在预算范围内：90-100分
- 期望薪资略高于预算（10%以内）：70-89分
- 期望薪资明显高于预算（10-30%）：40-69分
- 期望薪资严重超出预算（30%以上）：0-39分

### 3. 部门/职能匹配 (department)
- 完全匹配目标部门：90-100分
- 相关部门经验：70-89分
- 有一定相关性：40-69分
- 完全不相关：0-39分

## 输出格式

请严格按照以下JSON格式输出结果（字段名使用下划线命名）：

{
  "score": 总分（0-100的浮点数）,
  "sub_scores": {
    "location": 地理位置分数（0-100的浮点数）,
    "salary": 薪资匹配分数（0-100的浮点数）,
    "department": 部门匹配分数（0-100的浮点数）
  },
  "evidence": [
    "包含关键信息的理由说明"
  ],
  "notes": "整体结论与补充说明"
}

注意：
1. 所有分数必须是0-100之间的数值
2. 总分 = location*0.5 + salary*0.3 + department*0.2
3. evidence用于列出支撑评分的关键信息，每条不超过60个汉字，如信息缺失需指出
4. notes字段用于给出综合结论、风险提示或补充说明`

// BasicInfoUserPrompt 用户输入提示词模板
const BasicInfoUserPrompt = `请分析以下候选人基本信息与职位要求的匹配度：

## 输入数据说明
以下JSON数据仅包含与基本信息匹配相关的关键字段：
- job_profile: 岗位名称、所属部门、工作地点、薪资区间等核心信息（已剔除职责、技能等冗余内容）
- resume: 候选人姓名、当前城市、工作年限、近期经历摘要等基础信息（仅保留与基础匹配相关的内容）
- notes: 可能出现的提示信息，标记出缺失或需特别注意的要素

## 待分析数据
{{.input}}

## 分析要求
请仔细对比职位要求与候选人信息，重点关注：
1. **地理位置匹配度**：对比职位工作地点与候选人期望工作地点/当前居住地
2. **薪资期望匹配度**：对比职位薪资范围与候选人期望薪资
3. **部门职能匹配度**：对比职位所属部门与候选人相关工作经验
4. **证据与说明**：在evidence字段中列出关键事实支撑评分，在notes字段中总结总体结论及信息缺口

请根据系统提示中的详细评分规则，对候选人进行全面评估并严格按照JSON格式输出结果。`

// NewBasicInfoChatTemplate 创建基本信息匹配的聊天模板
func NewBasicInfoChatTemplate(ctx context.Context) (prompt.ChatTemplate, error) {
	config := &ChatTemplateConfig{
		FormatType: schema.GoTemplate,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(BasicInfoSystemPrompt),
			schema.UserMessage(BasicInfoUserPrompt),
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
