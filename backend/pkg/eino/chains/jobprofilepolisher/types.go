package jobprofilepolisher

// PolishJobPromptInput 一键润色链输入
type PolishJobPromptInput struct {
	Idea string `json:"idea"` // 用户原始想法或一句话描述
}

// PolishJobPromptResult 一键润色链输出
type PolishJobPromptResult struct {
	PolishedPrompt     string   `json:"polished_prompt"`     // 完整且可直接用于生成链的 Prompt
	SuggestedTitle     string   `json:"suggested_title"`     // AI 推荐的岗位标题
	ResponsibilityTips []string `json:"responsibility_tips"` // 岗位职责提示
	RequirementTips    []string `json:"requirement_tips"`    // 任职要求提示
	BonusTips          []string `json:"bonus_tips"`          // 加分项提示
}
