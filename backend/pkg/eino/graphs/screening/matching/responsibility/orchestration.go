package responsibility

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ResponsibilityAgent 职责匹配Agent
type ResponsibilityAgent struct {
	chain *compose.Chain[*domain.ResponsibilityData, *domain.ResponsibilityMatchDetail]
}

// 输入处理，将ResponsibilityData转换为模板变量
func newInputLambda(ctx context.Context, input *domain.ResponsibilityData, opts ...any) (map[string]any, error) {
	// 验证输入数据
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// 构建结构化的输入数据，便于LLM理解
	inputData := map[string]any{
		"job_responsibilities": input.JobResponsibilities,
		"resume_experiences":   input.ResumeExperiences,
		"resume_projects":      input.ResumeProjects,
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
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*domain.ResponsibilityMatchDetail, error) {

	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	var output domain.ResponsibilityMatchDetail
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	return &output, nil
}

// NewResponsibilityAgent 创建职责匹配Agent
func NewResponsibilityAgent(ctx context.Context, llm model.ToolCallingChatModel) (*ResponsibilityAgent, error) {
	// 创建聊天模板
	chatTemplate, err := NewResponsibilityChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[*domain.ResponsibilityData, *domain.ResponsibilityMatchDetail]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model"), compose.WithNodeKey("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &ResponsibilityAgent{
		chain: chain,
	}, nil
}

// GetChain 获取处理链
func (a *ResponsibilityAgent) GetChain() *compose.Chain[*domain.ResponsibilityData, *domain.ResponsibilityMatchDetail] {
	return a.chain
}

// Compile 编译链为可执行的Runnable
func (a *ResponsibilityAgent) Compile(ctx context.Context) (compose.Runnable[*domain.ResponsibilityData, *domain.ResponsibilityMatchDetail], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return runnable, nil
}

// Process 处理职责匹配
func (a *ResponsibilityAgent) Process(ctx context.Context, input *domain.ResponsibilityData) (*domain.ResponsibilityMatchDetail, error) {
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
func (a *ResponsibilityAgent) validateInput(input *domain.ResponsibilityData) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	if input.JobResponsibilities == nil {
		return fmt.Errorf("job responsibilities cannot be nil")
	}
	if input.ResumeExperiences == nil {
		return fmt.Errorf("resume experiences cannot be nil")
	}
	return nil
}

// validateOutput 验证输出数据
func (a *ResponsibilityAgent) validateOutput(output *domain.ResponsibilityMatchDetail) error {
	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}

	if output.Score < 0 || output.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100, got %f", output.Score)
	}

	// 验证匹配的职责
	for _, matchedResp := range output.MatchedResponsibilities {
		if matchedResp.MatchScore < 0 || matchedResp.MatchScore > 100 {
			return fmt.Errorf("responsibility match score must be between 0 and 100, got %f", matchedResp.MatchScore)
		}
	}

	return nil
}

// GetAgentType 返回Agent类型
func (a *ResponsibilityAgent) GetAgentType() string {
	return "ResponsibilityAgent"
}

// GetVersion 返回Agent版本
func (a *ResponsibilityAgent) GetVersion() string {
	return "1.0.0"
}
