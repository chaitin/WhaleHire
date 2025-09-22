package retrieverchat

import (
	"context"
	"time"
)

// newLambdaWithHistory 处理用户输入，添加历史记录和时间戳
func newLambdaWithHistory(ctx context.Context, input *UserMessage, opts ...any) (output map[string]any, err error) {
	return map[string]any{
		"input":   input.Input,
		"history": input.History,
		"date":    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func newLambdaWithQuery(ctx context.Context, input *UserMessage, opts ...any) (output string, err error) {
	return input.Input, nil
}
