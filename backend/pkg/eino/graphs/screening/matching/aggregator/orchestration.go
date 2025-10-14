package aggregator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/types"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// AggregatorAgent 匹配结果聚合Agent
type AggregatorAgent struct {
	chain *compose.Chain[map[string]any, *types.JobResumeMatch]
}

// AggregatorInput 聚合器输入结构
type AggregatorInput struct {
	BasicMatch          *types.BasicMatchDetail          `json:"basic_match"`
	SkillMatch          *types.SkillMatchDetail          `json:"skill_match"`
	ResponsibilityMatch *types.ResponsibilityMatchDetail `json:"responsibility_match"`
	ExperienceMatch     *types.ExperienceMatchDetail     `json:"experience_match"`
	EducationMatch      *types.EducationMatchDetail      `json:"education_match"`
	IndustryMatch       *types.IndustryMatchDetail       `json:"industry_match"`
	Weights             *types.DimensionWeights          `json:"weights"`
	TaskMetaData        *types.TaskMetaData              `json:"-"` // 任务元数据不参与序列化
}

// 输入处理，将map[string]any转换为模板变量
func newInputLambda(ctx context.Context, input map[string]any, opts ...any) (map[string]any, error) {
	// 打印输入内容，方便调试
	fmt.Printf("aggregator Input=%v\n", input)

	// 构建聚合输入结构
	aggregatorInput := &AggregatorInput{}

	// 从map中提取各个Agent的输出结果
	if basicMatch, ok := input[types.BasicInfoAgent]; ok {
		if basicDetail, ok := basicMatch.(*types.BasicMatchDetail); ok {
			aggregatorInput.BasicMatch = basicDetail
		}
	}

	if skillMatch, ok := input[types.SkillAgent]; ok {
		if skillDetail, ok := skillMatch.(*types.SkillMatchDetail); ok {
			aggregatorInput.SkillMatch = skillDetail
		}
	}

	if responsibilityMatch, ok := input[types.ResponsibilityAgent]; ok {
		if respDetail, ok := responsibilityMatch.(*types.ResponsibilityMatchDetail); ok {
			aggregatorInput.ResponsibilityMatch = respDetail
		}
	}

	if experienceMatch, ok := input[types.ExperienceAgent]; ok {
		if expDetail, ok := experienceMatch.(*types.ExperienceMatchDetail); ok {
			aggregatorInput.ExperienceMatch = expDetail
		}
	}

	if educationMatch, ok := input[types.EducationAgent]; ok {
		if eduDetail, ok := educationMatch.(*types.EducationMatchDetail); ok {
			aggregatorInput.EducationMatch = eduDetail
		}
	}

	if industryMatch, ok := input[types.IndustryAgent]; ok {
		if indDetail, ok := industryMatch.(*types.IndustryMatchDetail); ok {
			aggregatorInput.IndustryMatch = indDetail
		}
	}

	// 提取任务元数据
	if taskMetaData, ok := input[types.TaskMetaDataNode]; ok {
		if metaData, ok := taskMetaData.(*types.TaskMetaData); ok {
			aggregatorInput.TaskMetaData = metaData
			// 从任务元数据中提取权重配置
			if metaData.DimensionWeights != nil {
				aggregatorInput.Weights = metaData.DimensionWeights
			} else {
				// 如果没有配置权重，使用默认权重
				aggregatorInput.Weights = &types.DefaultDimensionWeights
			}
		}
	}

	// 将聚合输入转换为JSON字符串
	inputJSON, err := json.Marshal(aggregatorInput)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal aggregator input: %w", err)
	}

	return map[string]any{
		"input":   string(inputJSON),
		"weights": aggregatorInput.Weights,
	}, nil
}

// 输出处理，将模型输出解析为结构化结果
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*types.JobResumeMatch, error) {
	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	var output types.JobResumeMatch
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	return &output, nil
}

// NewAggregatorAgent 创建匹配结果聚合Agent
func NewAggregatorAgent(ctx context.Context, llm model.ToolCallingChatModel) (*AggregatorAgent, error) {
	// 创建聊天模板
	chatTemplate, err := NewAggregatorChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[map[string]any, *types.JobResumeMatch]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &AggregatorAgent{
		chain: chain,
	}, nil
}

// GetChain 获取处理链
func (a *AggregatorAgent) GetChain() *compose.Chain[map[string]any, *types.JobResumeMatch] {
	return a.chain
}

// Compile 编译链为可执行的Runnable
func (a *AggregatorAgent) Compile(ctx context.Context) (compose.Runnable[map[string]any, *types.JobResumeMatch], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return runnable, nil
}

// Process 处理匹配结果聚合
func (a *AggregatorAgent) Process(ctx context.Context, input map[string]any) (*types.JobResumeMatch, error) {
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
func (a *AggregatorAgent) validateInput(input map[string]any) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	// 验证任务元数据
	taskMetaData, ok := input[types.TaskMetaDataNode]
	if !ok {
		return fmt.Errorf("task metadata is required")
	}
	if metaData, ok := taskMetaData.(*types.TaskMetaData); !ok {
		return fmt.Errorf("task metadata must be of type *TaskMetaData")
	} else {
		if metaData.JobID == "" {
			return fmt.Errorf("job_id in task metadata cannot be empty")
		}
		if metaData.ResumeID == "" {
			return fmt.Errorf("resume_id in task metadata cannot be empty")
		}
		if metaData.MatchTaskID == "" {
			return fmt.Errorf("match_task_id in task metadata cannot be empty")
		}
	}

	// 验证至少有一个Agent的输出
	agentOutputs := []string{
		types.BasicInfoAgent,
		types.SkillAgent,
		types.ResponsibilityAgent,
		types.ExperienceAgent,
		types.EducationAgent,
		types.IndustryAgent,
	}

	hasOutput := false
	for _, agentName := range agentOutputs {
		if output, ok := input[agentName]; ok {
			hasOutput = true
			// 验证输出类型
			switch agentName {
			case types.BasicInfoAgent:
				if _, ok := output.(*types.BasicMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *BasicMatchDetail", agentName)
				}
			case types.SkillAgent:
				if _, ok := output.(*types.SkillMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *SkillMatchDetail", agentName)
				}
			case types.ResponsibilityAgent:
				if _, ok := output.(*types.ResponsibilityMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *ResponsibilityMatchDetail", agentName)
				}
			case types.ExperienceAgent:
				if _, ok := output.(*types.ExperienceMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *ExperienceMatchDetail", agentName)
				}
			case types.EducationAgent:
				if _, ok := output.(*types.EducationMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *EducationMatchDetail", agentName)
				}
			case types.IndustryAgent:
				if _, ok := output.(*types.IndustryMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *IndustryMatchDetail", agentName)
				}
			}
		}
	}

	if !hasOutput {
		return fmt.Errorf("at least one agent output is required")
	}

	return nil
}

// validateOutput 验证输出数据
func (a *AggregatorAgent) validateOutput(output *types.JobResumeMatch) error {
	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}
	if output.OverallScore < 0 || output.OverallScore > 100 {
		return fmt.Errorf("overall score must be between 0 and 100, got %f", output.OverallScore)
	}

	// 验证各维度得分
	if output.BasicMatch != nil && (output.BasicMatch.Score < 0 || output.BasicMatch.Score > 100) {
		return fmt.Errorf("basic match score must be between 0 and 100, got %f", output.BasicMatch.Score)
	}
	if output.SkillMatch != nil && (output.SkillMatch.Score < 0 || output.SkillMatch.Score > 100) {
		return fmt.Errorf("skill match score must be between 0 and 100, got %f", output.SkillMatch.Score)
	}
	if output.ResponsibilityMatch != nil && (output.ResponsibilityMatch.Score < 0 || output.ResponsibilityMatch.Score > 100) {
		return fmt.Errorf("responsibility match score must be between 0 and 100, got %f", output.ResponsibilityMatch.Score)
	}
	if output.ExperienceMatch != nil && (output.ExperienceMatch.Score < 0 || output.ExperienceMatch.Score > 100) {
		return fmt.Errorf("experience match score must be between 0 and 100, got %f", output.ExperienceMatch.Score)
	}
	if output.EducationMatch != nil && (output.EducationMatch.Score < 0 || output.EducationMatch.Score > 100) {
		return fmt.Errorf("education match score must be between 0 and 100, got %f", output.EducationMatch.Score)
	}
	if output.IndustryMatch != nil && (output.IndustryMatch.Score < 0 || output.IndustryMatch.Score > 100) {
		return fmt.Errorf("industry match score must be between 0 and 100, got %f", output.IndustryMatch.Score)
	}

	return nil
}

// GetAgentType 返回Agent类型
func (a *AggregatorAgent) GetAgentType() string {
	return "AggregatorAgent"
}

// GetVersion 返回Agent版本
func (a *AggregatorAgent) GetVersion() string {
	return "1.0.0"
}
