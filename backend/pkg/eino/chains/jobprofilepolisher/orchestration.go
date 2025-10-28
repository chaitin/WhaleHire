package jobprofilepolisher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// JobProfilePolisherChain 岗位 Prompt 润色链
type JobProfilePolisherChain struct {
	chain *compose.Chain[*PolishJobPromptInput, *PolishJobPromptResult]
}

// GetChain 返回编排链，用于进一步组合
func (c *JobProfilePolisherChain) GetChain() *compose.Chain[*PolishJobPromptInput, *PolishJobPromptResult] {
	return c.chain
}

// Compile 编译链得到可执行 Runnable
func (c *JobProfilePolisherChain) Compile(ctx context.Context) (compose.Runnable[*PolishJobPromptInput, *PolishJobPromptResult], error) {
	r, err := c.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return r, nil
}

func newInputLambda(ctx context.Context, input *PolishJobPromptInput, _ ...any) (map[string]any, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	idea := strings.TrimSpace(input.Idea)
	if idea == "" {
		return nil, fmt.Errorf("idea is empty")
	}

	return map[string]any{
		"idea":         idea,
		"current_date": time.Now().Format("2006-01-02"),
	}, nil
}

func newOutputLambda(ctx context.Context, msg *schema.Message, _ ...any) (*PolishJobPromptResult, error) {
	if msg == nil || strings.TrimSpace(msg.Content) == "" {
		return nil, fmt.Errorf("empty model output")
	}

	result, err := decodePolishJobPromptResult([]byte(msg.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	// polished prompt 必须有值
	if strings.TrimSpace(result.PolishedPrompt) == "" {
		return nil, fmt.Errorf("polished prompt is empty")
	}

	if strings.TrimSpace(result.SuggestedTitle) == "" {
		return nil, fmt.Errorf("suggested title is empty")
	}

	return result, nil
}

// NewJobProfilePolisherChain 构建岗位 Prompt 润色链
func NewJobProfilePolisherChain(ctx context.Context, chatModel model.ToolCallingChatModel) (*JobProfilePolisherChain, error) {
	chatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	chain := compose.NewChain[*PolishJobPromptInput, *PolishJobPromptResult]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &JobProfilePolisherChain{chain: chain}, nil
}
