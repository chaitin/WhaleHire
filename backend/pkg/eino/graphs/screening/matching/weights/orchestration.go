package weights

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// WeightPlannerAgent 岗位权重规划 Agent
type WeightPlannerAgent struct {
	chain *compose.Chain[*domain.WeightInferenceInput, *domain.WeightInferenceResult]
}

type llmWeightOutput struct {
	Skill          float64  `json:"skill"`
	Responsibility float64  `json:"responsibility"`
	Experience     float64  `json:"experience"`
	Education      float64  `json:"education"`
	Industry       float64  `json:"industry"`
	Basic          float64  `json:"basic"`
	Rationale      []string `json:"rationale"`
}

// newInputLambda 将权重推理输入转换为模板变量
func newInputLambda(ctx context.Context, input *domain.WeightInferenceInput, opts ...any) (map[string]any, error) {
	if input == nil {
		return nil, fmt.Errorf("权重推理输入不能为空")
	}
	if input.JobProfile == nil || input.JobProfile.JobProfile == nil {
		return nil, fmt.Errorf("岗位信息不能为空")
	}

	payload := map[string]any{
		"job_profile": map[string]any{
			"base":             input.JobProfile.JobProfile,
			"responsibilities": input.JobProfile.Responsibilities,
			"skills":           input.JobProfile.Skills,
			"experience":       input.JobProfile.ExperienceRequirements,
			"education":        input.JobProfile.EducationRequirements,
			"industry":         input.JobProfile.IndustryRequirements,
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化岗位数据失败: %w", err)
	}

	return map[string]any{
		"job_profile_json": string(payloadJSON),
	}, nil
}

// newOutputLambda 解析模型输出为结构化权重
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*domain.WeightInferenceResult, error) {
	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("模型输出为空")
	}

	var output llmWeightOutput
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("解析模型输出失败: %w; raw=%s", err, msg.Content)
	}

	weights := domain.DimensionWeights{
		Skill:          output.Skill,
		Responsibility: output.Responsibility,
		Experience:     output.Experience,
		Education:      output.Education,
		Industry:       output.Industry,
		Basic:          output.Basic,
	}

	if err := validateWeights(weights); err != nil {
		return nil, err
	}

	result := &domain.WeightInferenceResult{
		Weights:   weights,
		Rationale: output.Rationale,
	}
	return result, nil
}

// validateWeights 校验权重取值
func validateWeights(weights domain.DimensionWeights) error {
	values := []struct {
		name  string
		value float64
	}{
		{"skill", weights.Skill},
		{"responsibility", weights.Responsibility},
		{"experience", weights.Experience},
		{"education", weights.Education},
		{"industry", weights.Industry},
		{"basic", weights.Basic},
	}

	var sum float64
	for _, item := range values {
		if item.value < 0 {
			return fmt.Errorf("维度 %s 权重不能为负: %.4f", item.name, item.value)
		}
		if item.value > 1 {
			return fmt.Errorf("维度 %s 权重超过1: %.4f", item.name, item.value)
		}
		sum += item.value
	}

	if sum <= 0 {
		return fmt.Errorf("权重总和必须大于0")
	}
	if diff := sum - 1; diff > 0.05 || diff < -0.05 {
		return fmt.Errorf("权重总和需接近1，当前为 %.4f", sum)
	}

	return nil
}

// NewWeightPlannerAgent 创建岗位权重规划 Agent
func NewWeightPlannerAgent(ctx context.Context, llm model.ToolCallingChatModel) (*WeightPlannerAgent, error) {
	chatTemplate, err := NewWeightPlannerChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建权重规划模板失败: %w", err)
	}

	chain := compose.NewChain[*domain.WeightInferenceInput, *domain.WeightInferenceResult]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model"), compose.WithNodeKey("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &WeightPlannerAgent{chain: chain}, nil
}

// GetChain 返回 Agent 处理链
func (a *WeightPlannerAgent) GetChain() *compose.Chain[*domain.WeightInferenceInput, *domain.WeightInferenceResult] {
	return a.chain
}

// Compile 编译链为 Runnable
func (a *WeightPlannerAgent) Compile(ctx context.Context) (compose.Runnable[*domain.WeightInferenceInput, *domain.WeightInferenceResult], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译权重规划链失败: %w", err)
	}
	return runnable, nil
}

// GetVersion 返回 Agent 版本
func (a *WeightPlannerAgent) GetVersion() string {
	return "1.1.0"
}
