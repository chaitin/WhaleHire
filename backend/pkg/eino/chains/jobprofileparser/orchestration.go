package jobprofileparser

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// JobProfileParserChain 岗位画像解析链结构
type JobProfileParserChain struct {
	chain *compose.Chain[*JobProfileParseInput, *JobProfileParseResult]
}

// GetChain 获取处理链
func (c *JobProfileParserChain) GetChain() *compose.Chain[*JobProfileParseInput, *JobProfileParseResult] {
	return c.chain
}

// Compile 编译链为可执行的Runnable
func (c *JobProfileParserChain) Compile(ctx context.Context) (compose.Runnable[*JobProfileParseInput, *JobProfileParseResult], error) {
	r, err := c.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return r, nil
}

// 输入处理，将岗位描述文本传入模板
func newInputLambda(ctx context.Context, input *JobProfileParseInput, opts ...any) (map[string]any, error) {
	return map[string]any{
		"description":  input.Description,
		"current_date": time.Now().Format("2006-01-02"),
	}, nil
}

// 输出处理，将模型输出解析为结构化结果
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*JobProfileParseResult, error) {
	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	res, err := decodeJobProfileParseResult([]byte(msg.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}
	return res, nil
}

// NewJobProfileParserChain 构建岗位画像解析工作流链
// 工作流程：输入 -> 模板 -> 聊天模型(GPT5/兼容) -> 结果解析
func NewJobProfileParserChain(ctx context.Context, chatModel model.ToolCallingChatModel) (*JobProfileParserChain, error) {
	chatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	chain := compose.NewChain[*JobProfileParseInput, *JobProfileParseResult]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &JobProfileParserChain{chain: chain}, nil
}
