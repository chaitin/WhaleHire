package intent

import "github.com/cloudwego/eino/schema"

// IntentInput 意图分类输入
type IntentInput struct {
	Query   string            `json:"query"`
	History []*schema.Message `json:"history,omitempty"`
}

type IntentResult struct {
	Intent string `json:"intent"`
}

// 定义意图常量
const (
	IntentSearchOnline   = "SearchOnline"
	IntentAnswerDirectly = "AnswerDirectly"
	// 未来扩展：
	IntentImageGeneration = "ImageGeneration"
	IntentCalculator      = "Calculator"
	IntentDatabaseQuery   = "DatabaseQuery"
)
