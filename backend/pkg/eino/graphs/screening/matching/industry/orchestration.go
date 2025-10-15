package industry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// IndustryAgent 行业背景匹配Agent
type IndustryAgent struct {
	chain *compose.Chain[*domain.IndustryData, *domain.IndustryMatchDetail]
}

// 输入处理，将IndustryData转换为模板变量
func newInputLambda(ctx context.Context, input *domain.IndustryData, opts ...any) (map[string]any, error) {
	// 验证输入数据
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// 构建结构化的输入数据，便于LLM理解
	inputData := map[string]any{
		"job_industry_requirements": input.JobIndustryRequirements,
		"resume_experiences":        input.ResumeExperiences,
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
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*domain.IndustryMatchDetail, error) {

	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	var output domain.IndustryMatchDetail
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	return &output, nil
}

// NewIndustryAgent 创建行业背景匹配Agent
func NewIndustryAgent(ctx context.Context, llm model.ToolCallingChatModel) (*IndustryAgent, error) {
	// 创建聊天模板
	chatTemplate, err := NewIndustryChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[*domain.IndustryData, *domain.IndustryMatchDetail]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &IndustryAgent{
		chain: chain,
	}, nil
}

// GetChain 获取处理链
func (a *IndustryAgent) GetChain() *compose.Chain[*domain.IndustryData, *domain.IndustryMatchDetail] {
	return a.chain
}

// Compile 编译链为可执行的Runnable
func (a *IndustryAgent) Compile(ctx context.Context) (compose.Runnable[*domain.IndustryData, *domain.IndustryMatchDetail], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return runnable, nil
}

// Process 处理行业背景匹配
func (a *IndustryAgent) Process(ctx context.Context, input *domain.IndustryData) (*domain.IndustryMatchDetail, error) {
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
func (a *IndustryAgent) validateInput(input *domain.IndustryData) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	if input.JobIndustryRequirements == nil {
		return fmt.Errorf("job industry requirements cannot be nil")
	}
	if input.ResumeExperiences == nil {
		return fmt.Errorf("resume experiences cannot be nil")
	}
	return nil
}

// validateOutput 验证输出数据
func (a *IndustryAgent) validateOutput(output *domain.IndustryMatchDetail) error {
	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}
	if output.Score < 0 || output.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100, got %f", output.Score)
	}

	// 验证行业匹配信息
	for _, indMatch := range output.IndustryMatches {
		if indMatch.Score < 0 || indMatch.Score > 100 {
			return fmt.Errorf("industry match score must be between 0 and 100, got %f", indMatch.Score)
		}
		if indMatch.Relevance < 0 || indMatch.Relevance > 100 {
			return fmt.Errorf("industry relevance must be between 0 and 100, got %f", indMatch.Relevance)
		}
	}

	// 验证公司匹配信息
	for _, compMatch := range output.CompanyMatches {
		if compMatch.Score < 0 || compMatch.Score > 100 {
			return fmt.Errorf("company match score must be between 0 and 100, got %f", compMatch.Score)
		}
	}

	return nil
}

// GetAgentType 返回Agent类型
func (a *IndustryAgent) GetAgentType() string {
	return "IndustryAgent"
}

// GetVersion 返回Agent版本
func (a *IndustryAgent) GetVersion() string {
	return "1.0.0"
}
