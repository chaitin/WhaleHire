package education

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// EducationAgent 教育背景匹配Agent
type EducationAgent struct {
	chain *compose.Chain[*domain.EducationData, *domain.EducationMatchDetail]
}

// 输入处理，将EducationData转换为模板变量
func newInputLambda(ctx context.Context, input *domain.EducationData, opts ...any) (map[string]any, error) {
	// 验证输入数据
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	// 构建结构化的输入数据，便于LLM理解
	inputData := map[string]any{
		"job_education_requirements": input.JobEducationRequirements,
		"resume_educations":          input.ResumeEducations,
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
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*domain.EducationMatchDetail, error) {

	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	var output domain.EducationMatchDetail
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	return &output, nil
}

// NewEducationAgent 创建教育背景匹配Agent
func NewEducationAgent(ctx context.Context, llm model.ToolCallingChatModel) (*EducationAgent, error) {
	// 创建聊天模板
	chatTemplate, err := NewEducationChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[*domain.EducationData, *domain.EducationMatchDetail]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model"), compose.WithNodeKey("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &EducationAgent{
		chain: chain,
	}, nil
}

// GetChain 获取处理链
func (a *EducationAgent) GetChain() *compose.Chain[*domain.EducationData, *domain.EducationMatchDetail] {
	return a.chain
}

// Compile 编译链为可执行的Runnable
func (a *EducationAgent) Compile(ctx context.Context) (compose.Runnable[*domain.EducationData, *domain.EducationMatchDetail], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return runnable, nil
}

// Process 处理教育背景匹配
func (a *EducationAgent) Process(ctx context.Context, input *domain.EducationData) (*domain.EducationMatchDetail, error) {
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
func (a *EducationAgent) validateInput(input *domain.EducationData) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	if input.JobEducationRequirements == nil {
		return fmt.Errorf("job education requirements cannot be nil")
	}
	if input.ResumeEducations == nil {
		return fmt.Errorf("resume educations cannot be nil")
	}
	return nil
}

// validateOutput 验证输出数据
func (a *EducationAgent) validateOutput(output *domain.EducationMatchDetail) error {
	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}
	if output.Score < 0 || output.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100, got %f", output.Score)
	}

	// 验证学历匹配
	if output.DegreeMatch != nil {
		if output.DegreeMatch.Score < 0 || output.DegreeMatch.Score > 100 {
			return fmt.Errorf("degree match score must be between 0 and 100, got %f", output.DegreeMatch.Score)
		}
	}

	// 验证专业匹配
	for _, majorMatch := range output.MajorMatches {
		if majorMatch.Score < 0 || majorMatch.Score > 100 {
			return fmt.Errorf("major match score must be between 0 and 100, got %f", majorMatch.Score)
		}
		if majorMatch.Relevance < 0 || majorMatch.Relevance > 100 {
			return fmt.Errorf("major relevance must be between 0 and 100, got %f", majorMatch.Relevance)
		}
	}

	// 验证学校匹配
	for _, schoolMatch := range output.SchoolMatches {
		if schoolMatch.Score < 0 || schoolMatch.Score > 100 {
			return fmt.Errorf("school match score must be between 0 and 100, got %f", schoolMatch.Score)
		}
		if schoolMatch.Reputation < 0 || schoolMatch.Reputation > 100 {
			return fmt.Errorf("school reputation must be between 0 and 100, got %f", schoolMatch.Reputation)
		}
		if schoolMatch.GraduationYear < 0 {
			return fmt.Errorf("school graduation_year must be non-negative, got %d", schoolMatch.GraduationYear)
		}
		if schoolMatch.GPA != nil && *schoolMatch.GPA < 0 {
			return fmt.Errorf("school gpa must be non-negative, got %f", *schoolMatch.GPA)
		}
	}

	return nil
}

// GetAgentType 返回Agent类型
func (a *EducationAgent) GetAgentType() string {
	return "EducationAgent"
}

// GetVersion 返回Agent版本
func (a *EducationAgent) GetVersion() string {
	return "1.1.0"
}
