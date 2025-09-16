package websearch

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudwego/eino/schema"
)

// newLambda2 component initialization function of node 'InputToHistory' in graph 'EinoAgent'
func newLambdaWithHistory(ctx context.Context, input *UserMessage, opts ...any) (output map[string]any, err error) {
	return map[string]any{
		"query":   input.Query,
		"history": input.History,
		"date":    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func newLambdaCovertWebSearchResult(ctx context.Context, input []*schema.Message, opts ...any) (output *WebSearchResult, err error) {
	var result WebSearchResult
	toolresult := input[0].Content

	err = json.Unmarshal([]byte(toolresult), &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
