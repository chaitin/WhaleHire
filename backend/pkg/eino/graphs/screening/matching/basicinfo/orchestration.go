package basicinfo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/types"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// BasicInfoAgent 基本信息匹配Agent
type BasicInfoAgent struct {
	chain *compose.Chain[*types.BasicInfoData, *types.BasicMatchDetail]
}

// 输入处理，将BasicInfoData转换为模板变量
func newInputLambda(ctx context.Context, input *types.BasicInfoData, opts ...any) (map[string]any, error) {
	// 验证输入数据
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// 构建结构化的输入数据，便于LLM理解
	inputData := map[string]any{
		"job_profile": input.JobProfile,
		"resume":      input.Resume,
	}

	// 将输入转换为JSON字符串
	inputJSON, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	return map[string]any{
		"input": string(inputJSON),
	}, nil
}

// 输出处理，将模型输出解析为结构化结果
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*types.BasicMatchDetail, error) {

	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	fmt.Printf("basicinfo msg=%v\n", msg.Content)

	var output types.BasicMatchDetail
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	return &output, nil
}

// NewBasicInfoAgent 创建基本信息匹配Agent
func NewBasicInfoAgent(ctx context.Context, llm model.ToolCallingChatModel) (*BasicInfoAgent, error) {
	// 创建聊天模板
	chatTemplate, err := NewBasicInfoChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[*types.BasicInfoData, *types.BasicMatchDetail]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &BasicInfoAgent{
		chain: chain,
	}, nil
}

// GetChain 获取处理链
func (a *BasicInfoAgent) GetChain() *compose.Chain[*types.BasicInfoData, *types.BasicMatchDetail] {
	return a.chain
}

// Compile 编译链为可执行的Runnable
func (a *BasicInfoAgent) Compile(ctx context.Context) (compose.Runnable[*types.BasicInfoData, *types.BasicMatchDetail], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return runnable, nil
}

// Process 处理基本信息匹配
func (a *BasicInfoAgent) Process(ctx context.Context, input *types.BasicInfoData) (*types.BasicMatchDetail, error) {
	// 验证输入
	if err := a.validateInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// 编译并执行处理链
	runnable, err := a.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}

	output, err := runnable.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to process: %w", err)
	}

	// 验证输出
	if err := a.validateOutput(output); err != nil {
		return nil, fmt.Errorf("invalid output: %w", err)
	}

	return output, nil
}

// validateInput 验证输入数据
func (a *BasicInfoAgent) validateInput(input *types.BasicInfoData) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	if input.JobProfile == nil {
		return fmt.Errorf("job profile cannot be nil")
	}
	if input.Resume == nil {
		return fmt.Errorf("resume cannot be nil")
	}
	return nil
}

// validateOutput 验证输出数据
func (a *BasicInfoAgent) validateOutput(output *types.BasicMatchDetail) error {
	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}
	if output.Score < 0 || output.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100, got %f", output.Score)
	}

	// 验证子分数
	for key, score := range output.SubScores {
		if score < 0 || score > 100 {
			return fmt.Errorf("sub score '%s' must be between 0 and 100, got %f", key, score)
		}
	}

	return nil
}

// GetAgentType 返回Agent类型
func (a *BasicInfoAgent) GetAgentType() string {
	return "BasicInfoAgent"
}

// GetVersion 返回Agent版本
func (a *BasicInfoAgent) GetVersion() string {
	return "1.0.0"
}
