package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/weights"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// WeightPreviewService 权重预览服务接口
type WeightPreviewService interface {
	PreviewWeights(ctx context.Context, jobProfile *domain.JobProfileDetail, llmConfig map[string]any) (*domain.WeightInferenceResult, map[string]int64, string, error)
	Version() string
}

type weightPreviewService struct {
	cfg              *config.Config
	factory          *models.ModelFactory
	logger           *slog.Logger
	version          string
	compiledRunnable compose.Runnable[*domain.WeightInferenceInput, *domain.WeightInferenceResult]
	currentAgent     *weights.WeightPlannerAgent
	currentModelType models.ModelType
	currentModelName string
}

// NewWeightPreviewService 创建权重预览服务
func NewWeightPreviewService(cfg *config.Config, logger *slog.Logger) (WeightPreviewService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &weightPreviewService{
		cfg:     cfg,
		factory: models.NewModelFactory(),
		logger:  logger,
		version: "1.1.0",
	}, nil
}

// PreviewWeights 预览权重，根据岗位信息自动推理维度权重
func (s *weightPreviewService) PreviewWeights(ctx context.Context, jobProfile *domain.JobProfileDetail, llmConfig map[string]any) (*domain.WeightInferenceResult, map[string]int64, string, error) {
	if jobProfile == nil || jobProfile.JobProfile == nil {
		return nil, nil, "", fmt.Errorf("岗位画像不能为空")
	}

	// 设置模型
	modelType, modelName, err := s.setupModel(llmConfig)
	if err != nil {
		return nil, nil, "", fmt.Errorf("设置模型失败: %w", err)
	}

	// 确保链已编译
	runnable, _, err := s.ensureCompiled(ctx, modelType, modelName)
	if err != nil {
		return nil, nil, "", fmt.Errorf("确保链已编译失败: %w", err)
	}

	// 构造输入
	input := &domain.WeightInferenceInput{
		JobProfile: jobProfile,
	}

	// 创建回调收集器来收集 token 使用情况
	var tokenUsage map[string]int64
	var collectedUsage *model.TokenUsage

	options := []compose.Option{
		compose.WithCallbacks(
			callbacks.NewHandlerBuilder().
				OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
					// 尝试从输出中提取 token 使用情况
					if callbackOutput := model.ConvCallbackOutput(output); callbackOutput != nil && callbackOutput.TokenUsage != nil {
						collectedUsage = callbackOutput.TokenUsage
					}
					return ctx
				}).Build(),
		),
	}

	// 调用推理
	result, err := runnable.Invoke(ctx, input, options...)
	if err != nil {
		s.logger.Warn("权重推理失败，回退到默认权重", "error", err, "job_id", jobProfile.JobProfile.ID)
		// 推理失败时回退到三种默认权重方案
		return &domain.WeightInferenceResult{
			WeightSchemes: []domain.WeightScheme{
				{
					Type:      domain.WeightSchemeTypeDefault,
					Weights:   domain.DefaultDimensionWeights,
					Rationale: []string{"推理失败，使用默认权重配置"},
				},
				{
					Type:      domain.WeightSchemeTypeFreshGraduate,
					Weights:   domain.DefaultDimensionWeights,
					Rationale: []string{"推理失败，使用默认权重配置"},
				},
				{
					Type:      domain.WeightSchemeTypeExperienced,
					Weights:   domain.DefaultDimensionWeights,
					Rationale: []string{"推理失败，使用默认权重配置"},
				},
			},
		}, nil, s.version, nil
	}

	// 归一化每个方案的权重
	for i := range result.WeightSchemes {
		sanitizedWeights := sanitizeWeights(result.WeightSchemes[i].Weights)
		result.WeightSchemes[i].Weights = sanitizedWeights
	}

	// 转换 token 使用情况
	if collectedUsage != nil {
		tokenUsage = map[string]int64{
			"input_tokens":  int64(collectedUsage.PromptTokens),
			"output_tokens": int64(collectedUsage.CompletionTokens),
			"total_tokens":  int64(collectedUsage.TotalTokens),
		}
	}

	return result, tokenUsage, s.version, nil
}

// Version 返回当前Agent版本
func (s *weightPreviewService) Version() string {
	return s.version
}

// ensureCompiled 确保链已编译并可复用
func (s *weightPreviewService) ensureCompiled(ctx context.Context, modelType models.ModelType, modelName string) (compose.Runnable[*domain.WeightInferenceInput, *domain.WeightInferenceResult], *weights.WeightPlannerAgent, error) {
	// 如果已编译且模型配置未变化，直接返回
	if s.compiledRunnable != nil && s.currentAgent != nil && s.currentModelType == modelType && s.currentModelName == modelName {
		return s.compiledRunnable, s.currentAgent, nil
	}

	// 获取模型
	chatModel, err := s.factory.GetModel(ctx, modelType, modelName)
	if err != nil {
		return nil, nil, fmt.Errorf("获取对话模型失败: %w", err)
	}

	// 创建 WeightPlannerAgent
	agent, err := weights.NewWeightPlannerAgent(ctx, chatModel)
	if err != nil {
		return nil, nil, fmt.Errorf("创建权重规划Agent失败: %w", err)
	}

	// 编译链
	runnable, err := agent.Compile(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("编译权重规划链失败: %w", err)
	}

	// 缓存编译结果和模型配置
	s.compiledRunnable = runnable
	s.currentAgent = agent
	s.currentModelType = modelType
	s.currentModelName = modelName
	s.version = agent.GetVersion()

	return runnable, agent, nil
}

// setupModel 根据LLM配置设置模型
func (s *weightPreviewService) setupModel(llmConfig map[string]any) (models.ModelType, string, error) {
	// 如果没有提供LLM配置，使用系统默认配置
	if len(llmConfig) == 0 {
		modelName := s.cfg.GeneralAgent.LLM.ModelName
		if modelName == "" {
			modelName = "gpt-4o-mini"
		}

		openaiConfig := &models.OpenAIConfig{
			APIKey:         s.cfg.GeneralAgent.LLM.APIKey,
			BaseURL:        s.cfg.GeneralAgent.LLM.BaseURL,
			Model:          modelName,
			ResponseFormat: "json_object",
		}

		s.factory.Register(models.ModelTypeOpenAI, modelName, models.NewOpenAIModelManager(openaiConfig))
		return models.ModelTypeOpenAI, modelName, nil
	}

	// 解析用户提供的LLM配置
	modelType, modelName, err := s.parseLLMConfig(llmConfig)
	if err != nil {
		return "", "", fmt.Errorf("解析LLM配置失败: %w", err)
	}

	// 根据配置创建并注册模型
	switch modelType {
	case models.ModelTypeOpenAI:
		openaiConfig, err := s.buildOpenAIConfig(llmConfig, modelName)
		if err != nil {
			return "", "", fmt.Errorf("构建OpenAI配置失败: %w", err)
		}
		s.factory.Register(models.ModelTypeOpenAI, modelName, models.NewOpenAIModelManager(openaiConfig))
	default:
		return "", "", fmt.Errorf("不支持的模型类型: %s", modelType)
	}

	return modelType, modelName, nil
}

// parseLLMConfig 解析LLM配置
func (s *weightPreviewService) parseLLMConfig(llmConfig map[string]any) (models.ModelType, string, error) {
	modelTypeStr, ok := llmConfig["type"].(string)
	if !ok {
		return "", "", fmt.Errorf("模型类型不能为空")
	}

	modelName, ok := llmConfig["model"].(string)
	if !ok {
		return "", "", fmt.Errorf("模型名称不能为空")
	}

	var modelType models.ModelType
	switch modelTypeStr {
	case "openai":
		modelType = models.ModelTypeOpenAI
	default:
		return "", "", fmt.Errorf("不支持的模型类型: %s", modelTypeStr)
	}

	return modelType, modelName, nil
}

// buildOpenAIConfig 构建OpenAI配置
func (s *weightPreviewService) buildOpenAIConfig(llmConfig map[string]any, modelName string) (*models.OpenAIConfig, error) {
	apiKey, ok := llmConfig["api_key"].(string)
	if !ok {
		return nil, fmt.Errorf("OpenAI模型需要api_key")
	}

	baseURL, _ := llmConfig["base_url"].(string)
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &models.OpenAIConfig{
		APIKey:         apiKey,
		BaseURL:        baseURL,
		Model:          modelName,
		ResponseFormat: "json_object",
	}, nil
}

// sanitizeWeights 权重归一化与校验
// 处理负值、总和异常等边界情况，确保权重合理性
func sanitizeWeights(weights domain.DimensionWeights) domain.DimensionWeights {
	sanitized := weights

	// 1. 将负值置零
	if sanitized.Skill < 0 {
		sanitized.Skill = 0
	}
	if sanitized.Responsibility < 0 {
		sanitized.Responsibility = 0
	}
	if sanitized.Experience < 0 {
		sanitized.Experience = 0
	}
	if sanitized.Education < 0 {
		sanitized.Education = 0
	}
	if sanitized.Industry < 0 {
		sanitized.Industry = 0
	}
	if sanitized.Basic < 0 {
		sanitized.Basic = 0
	}

	// 2. 应用最小权重下限（0.03），避免某个维度被完全忽略
	const minWeight = 0.03
	if sanitized.Skill < minWeight && sanitized.Skill > 0 {
		sanitized.Skill = minWeight
	}
	if sanitized.Responsibility < minWeight && sanitized.Responsibility > 0 {
		sanitized.Responsibility = minWeight
	}
	if sanitized.Experience < minWeight && sanitized.Experience > 0 {
		sanitized.Experience = minWeight
	}
	if sanitized.Education < minWeight && sanitized.Education > 0 {
		sanitized.Education = minWeight
	}
	if sanitized.Industry < minWeight && sanitized.Industry > 0 {
		sanitized.Industry = minWeight
	}
	if sanitized.Basic < minWeight && sanitized.Basic > 0 {
		sanitized.Basic = minWeight
	}

	// 3. 计算总和
	sum := sanitized.Skill + sanitized.Responsibility + sanitized.Experience +
		sanitized.Education + sanitized.Industry + sanitized.Basic

	// 4. 若总和接近 0 则回退默认权重
	if sum < 0.01 {
		return domain.DefaultDimensionWeights
	}

	// 5. 对总和 ≠1 的情况按比例归一化
	if math.Abs(sum-1.0) > 0.01 {
		sanitized.Skill /= sum
		sanitized.Responsibility /= sum
		sanitized.Experience /= sum
		sanitized.Education /= sum
		sanitized.Industry /= sum
		sanitized.Basic /= sum
	}

	return sanitized
}
