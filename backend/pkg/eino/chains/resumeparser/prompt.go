package resumeparser

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

var systemPrompt = `
你是一个专业的简历解析助手。请从提供的简历内容中提取信息，并严格按照以下JSON格式输出。

## 输出格式要求
1. 必须输出有效的JSON格式
2. 所有字段名必须与schema完全一致
3. 日期格式使用 RFC3339 标准 (例如: "2023-01-15T00:00:00Z")
4. 如果某个字段信息不存在，使用null或空字符串
5. 数组字段即使为空也要输出空数组[]

## JSON Schema:
{{
  "basic_info": {{
    "name": "string (必填)",
    "phone": "string (必填)",
    "email": "string (必填)", 
    "gender": "string (可选: 男/女/未知)",
    "birthday": "string|null (RFC3339格式日期)",
    "current_city": "string (当前居住地)",
    "highest_education": "string (最高学历: 专科/本科/硕士/博士等)",
    "years_experience": "number (工作年限，支持小数，如3.5年)"
  }},
  "educations": [
    {{
      "school": "string (学校名称)",
      "major": "string (专业)",
      "degree": "string (学位: 本科/硕士/博士/专科等)",
      "start_date": "string|null (RFC3339格式)",
      "end_date": "string|null (RFC3339格式)",
      "gpa": "string (可选: GPA成绩)"
    }}
  ],
  "experiences": [
    {{
      "company": "string (公司名称)",
      "position": "string (职位名称)",
      "start_date": "string|null (RFC3339格式)",
      "end_date": "string|null (RFC3339格式，在职可为null)",
      "description": "string (工作描述和职责)",
      "achievements": "string (可选: 主要成就和业绩)"
    }}
  ],
  "skills": [
    {{
      "name": "string (技能名称)",
      "level": "string (熟练程度: 初级/中级/高级/精通)",
      "description": "string (可选: 技能描述)"
    }}
  ]
}}

## 输出示例:
{{
  "basic_info": {{
    "name": "张三",
    "phone": "13800138000",
    "email": "zhangsan@example.com",
    "gender": "男",
    "birthday": "1990-05-15T00:00:00Z",
    "current_city": "北京市",
    "highest_education": "本科",
    "years_experience": 3.3
  }},
  "educations": [
    {{
      "school": "清华大学",
      "major": "计算机科学与技术",
      "degree": "本科",
      "start_date": "2008-09-01T00:00:00Z",
      "end_date": "2012-06-30T00:00:00Z",
      "gpa": "3.8"
    }}
  ],
  "experiences": [
    {{
      "company": "阿里巴巴",
      "position": "高级软件工程师",
      "start_date": "2020-03-01T00:00:00Z",
      "end_date": null,
      "description": "负责电商平台后端开发，参与微服务架构设计",
      "achievements": "优化系统性能提升30%，获得年度优秀员工"
    }}
  ],
  "skills": [
    {{
      "name": "Java",
      "level": "精通",
      "description": "5年开发经验"
    }},
    {{
      "name": "Spring Boot",
      "level": "高级",
      "description": "熟练使用微服务开发"
    }}
  ]
}}

请仔细分析简历内容，提取所有相关信息，并严格按照上述格式输出JSON。
`

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	config := &ChatTemplateConfig{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(systemPrompt),
			schema.MessagesPlaceholder("history", false),
			schema.UserMessage("{resume}"),
		},
	}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
