package skill

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// SkillAgent 技能匹配Agent
type SkillAgent struct {
	chain *compose.Chain[*domain.SkillData, *domain.SkillMatchDetail]
}

// 输入处理，将SkillData转换为模板变量
func newInputLambda(ctx context.Context, input *domain.SkillData, opts ...any) (map[string]any, error) {
	// 验证输入数据
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// 构建结构化的输入数据，便于LLM理解
	inputData := map[string]any{
		"job_skills":      input.JobSkills,
		"resume_skills":   input.ResumeSkills,
		"resume_projects": input.ResumeProjects,
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
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*domain.SkillMatchDetail, error) {

	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	var output domain.SkillMatchDetail
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	return &output, nil
}

// NewSkillAgent 创建技能匹配Agent
func NewSkillAgent(ctx context.Context, llm model.ToolCallingChatModel) (*SkillAgent, error) {
	// 创建聊天模板
	chatTemplate, err := NewSkillChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[*domain.SkillData, *domain.SkillMatchDetail]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model"), compose.WithNodeKey("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &SkillAgent{
		chain: chain,
	}, nil
}

// GetChain 获取处理链
func (a *SkillAgent) GetChain() *compose.Chain[*domain.SkillData, *domain.SkillMatchDetail] {
	return a.chain
}

// Compile 编译链为可执行的Runnable
func (a *SkillAgent) Compile(ctx context.Context) (compose.Runnable[*domain.SkillData, *domain.SkillMatchDetail], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return runnable, nil
}

// Process 处理技能匹配
func (a *SkillAgent) Process(ctx context.Context, input *domain.SkillData) (*domain.SkillMatchDetail, error) {
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
func (a *SkillAgent) validateInput(input *domain.SkillData) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	if input.JobSkills == nil {
		return fmt.Errorf("job skills cannot be nil")
	}
	if input.ResumeSkills == nil {
		return fmt.Errorf("resume skills cannot be nil")
	}
	return nil
}

// validateOutput 验证输出数据
func (a *SkillAgent) validateOutput(output *domain.SkillMatchDetail) error {
	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}

	if output.Score < 0 || output.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100, got %f", output.Score)
	}

	// 验证匹配的技能
	for _, matchedSkill := range output.MatchedSkills {
		if matchedSkill.LLMScore < 0 || matchedSkill.LLMScore > 100 {
			return fmt.Errorf("skill match score must be between 0 and 100, got %f", matchedSkill.LLMScore)
		}
	}

	return nil
}

// GetAgentType 返回Agent类型
func (a *SkillAgent) GetAgentType() string {
	return "SkillAgent"
}

// GetVersion 返回Agent版本
func (a *SkillAgent) GetVersion() string {
	return "1.0.0"
}
