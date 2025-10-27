package jobprofilepolisher

import (
	"encoding/json"
	"strings"
)

type rawPolishJobPromptResult struct {
	PolishedPrompt     string   `json:"polished_prompt"`
	SuggestedTitle     string   `json:"suggested_title"`
	ResponsibilityTips []string `json:"responsibility_tips"`
	RequirementTips    []string `json:"requirement_tips"`
	BonusTips          []string `json:"bonus_tips"`
}

func decodePolishJobPromptResult(data []byte) (*PolishJobPromptResult, error) {
	var raw rawPolishJobPromptResult
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	return &PolishJobPromptResult{
		PolishedPrompt:     strings.TrimSpace(raw.PolishedPrompt),
		SuggestedTitle:     strings.TrimSpace(raw.SuggestedTitle),
		ResponsibilityTips: normalizeStringSlice(raw.ResponsibilityTips),
		RequirementTips:    normalizeStringSlice(raw.RequirementTips),
		BonusTips:          normalizeStringSlice(raw.BonusTips),
	}, nil
}

func normalizeStringSlice(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return []string{}
	}
	if len(result) > 2 {
		return result[:2]
	}
	return result
}
