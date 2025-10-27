package jobprofilegenerator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// JobProfileGeneratorChain 岗位画像生成链
type JobProfileGeneratorChain struct {
	chain *compose.Chain[*JobProfileGenerateInput, *JobProfileGenerateResult]
}

// GetChain 返回链结构体
func (c *JobProfileGeneratorChain) GetChain() *compose.Chain[*JobProfileGenerateInput, *JobProfileGenerateResult] {
	return c.chain
}

// Compile 编译链为可执行 Runnable
func (c *JobProfileGeneratorChain) Compile(ctx context.Context) (compose.Runnable[*JobProfileGenerateInput, *JobProfileGenerateResult], error) {
	r, err := c.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return r, nil
}

func newInputLambda(ctx context.Context, input *JobProfileGenerateInput, _ ...any) (map[string]any, error) {
	if input == nil {
		return nil, fmt.Errorf("input is required")
	}

	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is empty")
	}

	return map[string]any{
		"prompt":       prompt,
		"current_date": time.Now().Format("2006-01-02"),
	}, nil
}

func newOutputLambda(ctx context.Context, msg *schema.Message, _ ...any) (*JobProfileGenerateResult, error) {
	if msg == nil || strings.TrimSpace(msg.Content) == "" {
		return nil, fmt.Errorf("empty model output")
	}

	result, err := decodeJobProfileGenerateResult([]byte(msg.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	if result.Profile == nil {
		return nil, fmt.Errorf("profile is empty")
	}

	return result, nil
}

// NewJobProfileGeneratorChain 构建岗位画像生成链
func NewJobProfileGeneratorChain(ctx context.Context, chatModel model.ToolCallingChatModel) (*JobProfileGeneratorChain, error) {
	chatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	chain := compose.NewChain[*JobProfileGenerateInput, *JobProfileGenerateResult]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &JobProfileGeneratorChain{chain: chain}, nil
}
