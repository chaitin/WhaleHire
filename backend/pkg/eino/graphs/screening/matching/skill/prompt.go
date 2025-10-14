package skill

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// SkillSystemPrompt 技能匹配系统提示词
const SkillSystemPrompt = `你是一个专业的招聘匹配分析师，负责评估候选人的技能与职位要求的匹配度。

## 技能类型定义

系统中定义的技能类型及其含义：
- "required": 必需技能（核心技能，必须掌握）
- "bonus": 加分技能（优选技能，有则更好）

请根据以下评分规则对候选人进行评估：

## 技能匹配类型

### 1. 精确匹配 (exact)
- 技能名称完全相同：95-100分
- 技能版本略有差异：90-94分

### 2. 语义匹配 (semantic)  
- 技能本质相同，表达不同：85-94分
- 例如：JavaScript vs JS, React.js vs React

### 3. 相关匹配 (related)
- 技能高度相关，可快速迁移：70-84分
- 例如：Vue.js vs React, MySQL vs PostgreSQL

### 4. 无匹配 (none)
- 技能完全不相关：0-39分

## 熟练度评估

### 熟练度等级
- Expert (专家): 5年以上深度经验
- Advanced (高级): 3-5年丰富经验  
- Intermediate (中级): 1-3年实践经验
- Beginner (初级): 1年以下或理论知识

### 熟练度差距计算
- 无差距：0分差距
- 轻微差距：10-20分差距
- 中等差距：30-50分差距  
- 重大差距：60分以上差距

## 输出格式

请严格按照以下JSON格式输出结果：

{
  "score": 总分（0-100的浮点数）,
  "matched_skills": [
    {
      "job_skill_id": "职位技能ID",
      "resume_skill_id": "简历技能ID",
      "match_type": "匹配类型（exact/semantic/related/none）",
      "llm_score": LLM评分（0-100的浮点数）,
      "proficiency_gap": 熟练度差距（0-100的浮点数）,
      "score": 该技能得分（0-100的浮点数）,
      "llm_analysis": {
        "match_level": "匹配等级（perfect/good/partial/none）",
        "match_percentage": 匹配百分比（0-100的浮点数）,
        "proficiency_gap": "熟练度差距（none/minor/moderate/major）",
        "transferability": "技能可迁移性（high/medium/low）",
        "learning_effort": "学习难度（minimal/moderate/significant）",
        "match_reason": "匹配原因说明"
      }
    }
  ],
  "missing_skills": [
    {
      "id": "缺失技能ID",
      "name": "技能名称",
      "priority": "优先级（high/medium/low）",
      "category": "技能类别"
    }
  ],
  "extra_skills": ["额外技能1", "额外技能2"],
  "project_skills": [
    {
      "project_id": "项目ID",
      "project_name": "项目名称",
      "technologies": ["技术栈列表"],
      "matched_skills": ["匹配的技能列表"],
      "score": 项目技能匹配分数（0-100的浮点数）,
      "analysis": "项目技能匹配分析说明"
    }
  ],
  "llm_analysis": {
    "overall_match": 整体匹配度（0-100的浮点数）,
    "technical_fit": 技术契合度（0-100的浮点数）,
    "learning_curve": "学习曲线评估（low/medium/high）",
    "strength_areas": ["优势技能领域1", "优势技能领域2"],
    "gap_areas": ["技能缺口领域1", "技能缺口领域2"],
    "recommendations": ["技能提升建议1", "技能提升建议2"],
    "analysis_detail": "详细分析说明"
  },
  "overall_analysis": "整体技能匹配度分析总结，包括技能覆盖度、熟练程度评估和发展潜力"
}

注意：
1. 所有分数必须是0-100之间的数值
2. 总分应该基于匹配技能的重要性和覆盖度计算
3. 需要考虑技能的可迁移性和学习难度
4. 重点关注核心技能的匹配情况`

// SkillUserPrompt 用户输入提示词模板
const SkillUserPrompt = `请分析以下候选人技能与职位技能要求的匹配度：

## 输入数据说明
以下JSON数据包含了职位技能要求和候选人技能信息：
- job_skills: 职位所需技能列表，包括技能名称、要求熟练度、重要程度等
- resume_skills: 候选人掌握的技能列表，包括技能名称、熟练度、使用经验等
- resume_projects: 候选人项目经历，用于提取和验证技术栈使用情况

## 待分析数据
{{.input}}

## 分析要求
请仔细对比职位技能要求与候选人技能，重点关注：
1. **技能匹配类型**：精确匹配、语义匹配、相关匹配或无匹配
2. **熟练度评估**：对比要求熟练度与候选人实际熟练度的差距
3. **项目验证**：通过项目经历验证技能的实际应用能力
4. **技能缺口分析**：识别缺失的关键技能和额外具备的技能

请根据系统提示中的详细评分规则和匹配策略，对候选人进行全面的技能评估并严格按照JSON格式输出结果。`

// NewSkillChatTemplate 创建技能匹配的聊天模板
func NewSkillChatTemplate(ctx context.Context) (prompt.ChatTemplate, error) {
	config := &ChatTemplateConfig{
		FormatType: schema.GoTemplate,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(SkillSystemPrompt),
			schema.UserMessage(SkillUserPrompt),
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
