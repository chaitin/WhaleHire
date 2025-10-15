package experience

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ExperienceAgent 工作经验匹配Agent
type ExperienceAgent struct {
	chain *compose.Chain[*domain.ExperienceData, *domain.ExperienceMatchDetail]
}

// 输入处理，将ExperienceData转换为模板变量
func newInputLambda(ctx context.Context, input *domain.ExperienceData, opts ...any) (map[string]any, error) {
	// 验证输入数据
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// 构建结构化的输入数据，便于LLM理解
	inputData := map[string]any{
		"job_experience_requirements": input.JobExperienceRequirements,
		"resume_experiences":          input.ResumeExperiences,
		"resume_years_experience":     input.ResumeYearsExperience,
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
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*domain.ExperienceMatchDetail, error) {

	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	fmt.Printf("experience msg=%v\n", msg.Content)

	var output domain.ExperienceMatchDetail
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	return &output, nil
}

// NewExperienceAgent 创建工作经验匹配Agent
func NewExperienceAgent(ctx context.Context, llm model.ToolCallingChatModel) (*ExperienceAgent, error) {
	// 创建聊天模板
	chatTemplate, err := NewExperienceChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[*domain.ExperienceData, *domain.ExperienceMatchDetail]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &ExperienceAgent{
		chain: chain,
	}, nil
}

// GetChain 获取处理链
func (a *ExperienceAgent) GetChain() *compose.Chain[*domain.ExperienceData, *domain.ExperienceMatchDetail] {
	return a.chain
}

// Compile 编译链为可执行的Runnable
func (a *ExperienceAgent) Compile(ctx context.Context) (compose.Runnable[*domain.ExperienceData, *domain.ExperienceMatchDetail], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return runnable, nil
}

// Process 处理工作经验匹配
func (a *ExperienceAgent) Process(ctx context.Context, input *domain.ExperienceData) (*domain.ExperienceMatchDetail, error) {
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
func (a *ExperienceAgent) validateInput(input *domain.ExperienceData) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	if input.JobExperienceRequirements == nil {
		return fmt.Errorf("job experience requirements cannot be nil")
	}
	if input.ResumeExperiences == nil {
		return fmt.Errorf("resume experiences cannot be nil")
	}
	return nil
}

// validateOutput 验证输出数据
func (a *ExperienceAgent) validateOutput(output *domain.ExperienceMatchDetail) error {
	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}
	if output.Score < 0 || output.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100, got %f", output.Score)
	}

	// 验证年限匹配信息
	if output.YearsMatch != nil {
		if output.YearsMatch.Score < 0 || output.YearsMatch.Score > 100 {
			return fmt.Errorf("years match score must be between 0 and 100, got %f", output.YearsMatch.Score)
		}
	}

	// 验证职位匹配信息
	for _, posMatch := range output.PositionMatches {
		if posMatch.Score < 0 || posMatch.Score > 100 {
			return fmt.Errorf("position match score must be between 0 and 100, got %f", posMatch.Score)
		}
		if posMatch.Relevance < 0 || posMatch.Relevance > 100 {
			return fmt.Errorf("position relevance must be between 0 and 100, got %f", posMatch.Relevance)
		}
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

	return nil
}

// GetAgentType 返回Agent类型
func (a *ExperienceAgent) GetAgentType() string {
	return "ExperienceAgent"
}

// GetVersion 返回Agent版本
func (a *ExperienceAgent) GetVersion() string {
	return "1.0.0"
}
