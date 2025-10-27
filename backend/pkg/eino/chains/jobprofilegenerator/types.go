package jobprofilegenerator

import "github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofileparser"

// JobProfileGenerateInput 岗位画像生成链输入
type JobProfileGenerateInput struct {
	Prompt string `json:"prompt"` // 已润色的 Prompt
}

// JobProfileGenerateResult 岗位画像生成链输出
type JobProfileGenerateResult struct {
	Profile             *jobprofileparser.JobProfileParseResult `json:"profile"`              // 结构化岗位画像
	DescriptionMarkdown string                                  `json:"description_markdown"` // Markdown 描述
}
