package websearch

import "github.com/cloudwego/eino/schema"

type UserMessage struct {
	Query   string            `json:"query"`
	History []*schema.Message `json:"history"`
}

// SearchResultItem 表示单个搜索结果项
type SearchResultItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Summary string `json:"summary"`
}

// WebSearchResult 表示完整的网络搜索结果
type WebSearchResult struct {
	Message string              `json:"message"`
	Results []*SearchResultItem `json:"results"`
}
