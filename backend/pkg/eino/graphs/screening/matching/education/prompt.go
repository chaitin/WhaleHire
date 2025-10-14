package education

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// EducationSystemPrompt 教育背景匹配系统提示词
const EducationSystemPrompt = `你是一个专业的招聘匹配分析师，负责评估候选人的教育背景与职位要求的匹配度。

## 学历类型定义

系统中定义的学历类型及其含义：
- "unlimited": 不限学历要求
- "junior_college": 大专学历
- "bachelor": 本科学历
- "master": 硕士学历
- "doctor": 博士学历

请根据以下评分规则对候选人进行评估：

## 学历匹配评分

### 学历等级对应关系
- 博士 (PhD): 最高等级
- 硕士 (Master): 高等级
- 学士 (Bachelor): 中等级
- 专科 (Associate): 基础等级
- 高中及以下: 最低等级

### 学历匹配规则
- 完全匹配或超出要求：90-100分
- 低一个等级：70-89分
- 低两个等级：40-69分
- 低三个等级及以上：0-39分

## 专业匹配评分

### 专业相关性等级
- 完全匹配：95-100分
- 高度相关：85-94分
- 中度相关：70-84分
- 低度相关：50-69分
- 不相关：0-49分

### 专业匹配权重
- 核心专业要求：权重 70%
- 相关专业背景：权重 30%

## 学校声誉评分

### 学校等级划分
- 顶尖院校 (985/211/双一流): 90-100分
- 重点院校: 80-89分
- 普通本科院校: 70-79分
- 专科院校: 60-69分
- 其他院校: 50-59分

### 海外院校评估
- QS排名前50: 95-100分
- QS排名51-100: 90-94分
- QS排名101-200: 85-89分
- QS排名201-500: 80-84分
- 其他认可院校: 75-79分

## 输出格式

请严格按照以下JSON格式输出结果：

{
  "score": 总分（0-100的浮点数）,
  "degree_match": {
    "required_degree": "要求学历",
    "actual_degree": "实际学历",
    "score": 学历匹配分数（0-100的浮点数）,
    "meets": 是否满足要求（布尔值）
  },
  "major_matches": [
    {
      "resume_education_id": "简历教育经历ID",
      "major": "专业名称",
      "relevance": 专业相关性（0-100的浮点数）,
      "score": 该专业匹配分数（0-100的浮点数）
    }
  ],
  "school_matches": [
    {
      "resume_education_id": "简历教育经历ID",
      "school": "学校名称",
      "degree": "学位等级",
      "major": "专业名称",
      "graduation_year": 毕业年份,
      "reputation": 学校声誉分数（0-100的浮点数）,
      "score": 该学校匹配分数（0-100的浮点数）,
      "analysis": "学校匹配分析说明"
    }
  ],
  "overall_analysis": "整体教育背景匹配度分析总结"
}

## 综合评分计算

总分 = 学历匹配分数 × 0.4 + 专业匹配分数 × 0.4 + 学校声誉分数 × 0.2

注意：
1. 所有分数必须是0-100之间的数值
2. 学历是基础门槛，不满足基本要求会显著影响总分
3. 专业相关性是核心评估指标
4. 学校声誉作为加分项，但不是决定性因素
5. 需要考虑教育背景的时效性和持续学习能力`

// EducationUserPrompt 用户输入提示词模板
const EducationUserPrompt = `请分析以下候选人教育背景与职位教育要求的匹配度：

## 输入数据说明
以下JSON数据包含了职位教育要求和候选人教育背景信息：
- job_education_requirements: 职位对教育背景的要求，包括学历层次、专业要求、院校要求等
- resume_educations: 候选人教育经历列表，包括学校、专业、学历、毕业时间等

## 待分析数据
{{.input}}

## 分析要求
请仔细对比职位教育要求与候选人教育背景，重点关注：
1. **学历层次匹配**：候选人学历与职位要求学历的对比分析
2. **专业相关性**：所学专业与职位需求专业的匹配程度和相关度
3. **院校声誉**：毕业院校的知名度、排名和行业认可度评估
4. **教育质量**：综合评估教育背景对职位胜任能力的支撑程度

请根据系统提示中的详细评分规则和权重分配，对候选人进行全面的教育背景评估并严格按照JSON格式输出结果。`

// NewEducationChatTemplate 创建教育背景匹配的聊天模板
func NewEducationChatTemplate(ctx context.Context) (prompt.ChatTemplate, error) {
	config := &ChatTemplateConfig{
		FormatType: schema.GoTemplate,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(EducationSystemPrompt),
			schema.UserMessage(EducationUserPrompt),
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
