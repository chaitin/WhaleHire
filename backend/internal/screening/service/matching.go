package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	screening "github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// MatchingService 智能匹配服务接口
type MatchingService interface {
	Match(ctx context.Context, req *MatchRequest) (*MatchResult, error)
	Version() string
}

// MatchRequest 匹配请求
type MatchRequest struct {
	TaskID           uuid.UUID
	ResumeID         uuid.UUID
	JobProfile       *domain.JobProfileDetail
	Resume           *domain.ResumeDetail
	DimensionWeights map[string]float64
	LLMConfig        map[string]any
}

// MatchResult 匹配结果
type MatchResult struct {
	Match        *domain.JobResumeMatch
	TokenUsages  map[string]*model.TokenUsage
	Version      string
	Duration     time.Duration
	DimensionMap map[string]float64
}

type matchingService struct {
	logger   *slog.Logger
	cfg      *config.Config
	factory  *models.ModelFactory
	runnable compose.Runnable[*domain.MatchInput, *domain.JobResumeMatch]
	version  string
	initOnce sync.Once
	initErr  error
}

// NewMatchingService 创建匹配服务
func NewMatchingService(cfg *config.Config, logger *slog.Logger) (MatchingService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	factory := models.NewModelFactory()

	return &matchingService{
		logger:  logger,
		cfg:     cfg,
		factory: factory,
		version: "1.0.0",
	}, nil
}

// Match 执行智能匹配
func (s *matchingService) Match(ctx context.Context, req *MatchRequest) (*MatchResult, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if req.JobProfile == nil || req.JobProfile.JobProfile == nil {
		return nil, fmt.Errorf("job profile detail is required")
	}
	if req.Resume == nil || req.Resume.Resume == nil {
		return nil, fmt.Errorf("resume detail is required")
	}

	// 动态配置模型
	modelType, modelName, err := s.setupModel(req.LLMConfig)
	if err != nil {
		return nil, fmt.Errorf("setup model failed: %w", err)
	}

	if err := s.ensureRunnable(ctx, modelType, modelName); err != nil {
		return nil, err
	}

	weights := buildDimensionWeights(req.DimensionWeights)
	matchInput := &domain.MatchInput{
		JobProfile:       req.JobProfile,
		Resume:           req.Resume,
		DimensionWeights: weights,
		MatchTaskID:      fmt.Sprintf("%s:%s", req.TaskID.String(), req.ResumeID.String()),
	}

	collector := screening.NewAgentOutputCollector()
	start := time.Now()
	match, err := screening.InvokeWithCollector(ctx, s.runnable, matchInput, collector)
	duration := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("invoke screening graph failed: %w", err)
	}

	result := &MatchResult{
		Match:        match,
		TokenUsages:  collector.TokenUsages(),
		Version:      s.version,
		Duration:     duration,
		DimensionMap: weightsToMap(weights),
	}
	return result, nil
}

// Version 返回当前Agent版本
func (s *matchingService) Version() string {
	return s.version
}

// setupModel 根据LLM配置设置模型
func (s *matchingService) setupModel(llmConfig map[string]any) (models.ModelType, string, error) {
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
		return "", "", fmt.Errorf("parse LLM config failed: %w", err)
	}

	// 根据配置创建并注册模型
	switch modelType {
	case models.ModelTypeOpenAI:
		openaiConfig, err := s.buildOpenAIConfig(llmConfig, modelName)
		if err != nil {
			return "", "", fmt.Errorf("build OpenAI config failed: %w", err)
		}
		s.factory.Register(models.ModelTypeOpenAI, modelName, models.NewOpenAIModelManager(openaiConfig))
	default:
		return "", "", fmt.Errorf("unsupported model type: %s", modelType)
	}

	return modelType, modelName, nil
}

// parseLLMConfig 解析LLM配置
func (s *matchingService) parseLLMConfig(llmConfig map[string]any) (models.ModelType, string, error) {
	modelTypeStr, ok := llmConfig["type"].(string)
	if !ok {
		return "", "", fmt.Errorf("model type is required")
	}

	modelName, ok := llmConfig["model"].(string)
	if !ok {
		return "", "", fmt.Errorf("model name is required")
	}

	var modelType models.ModelType
	switch modelTypeStr {
	case "openai":
		modelType = models.ModelTypeOpenAI
	default:
		return "", "", fmt.Errorf("unsupported model type: %s", modelTypeStr)
	}

	return modelType, modelName, nil
}

// buildOpenAIConfig 构建OpenAI配置
func (s *matchingService) buildOpenAIConfig(llmConfig map[string]any, modelName string) (*models.OpenAIConfig, error) {
	apiKey, ok := llmConfig["api_key"].(string)
	if !ok {
		return nil, fmt.Errorf("api_key is required for OpenAI model")
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

func (s *matchingService) ensureRunnable(ctx context.Context, modelType models.ModelType, modelName string) error {
	s.initOnce.Do(func() {
		chatModel, err := s.factory.GetModel(ctx, modelType, modelName)
		if err != nil {
			s.initErr = fmt.Errorf("获取对话模型失败: %w", err)
			return
		}
		graph, err := screening.NewScreeningChatGraph(ctx, chatModel, s.cfg)
		if err != nil {
			s.initErr = fmt.Errorf("构建智能筛选图失败: %w", err)
			return
		}
		runnable, err := graph.Compile(ctx)
		if err != nil {
			s.initErr = fmt.Errorf("编译智能筛选图失败: %w", err)
			return
		}
		s.runnable = runnable
		s.version = graph.GetVersion()
	})
	return s.initErr
}

func buildDimensionWeights(overrides map[string]float64) *domain.DimensionWeights {
	weights := domain.DefaultDimensionWeights
	if overrides == nil {
		return &weights
	}
	if v, ok := overrides["skill"]; ok {
		weights.Skill = v
	}
	if v, ok := overrides["responsibility"]; ok {
		weights.Responsibility = v
	}
	if v, ok := overrides["experience"]; ok {
		weights.Experience = v
	}
	if v, ok := overrides["education"]; ok {
		weights.Education = v
	}
	if v, ok := overrides["industry"]; ok {
		weights.Industry = v
	}
	if v, ok := overrides["basic"]; ok {
		weights.Basic = v
	}
	return &weights
}

func weightsToMap(weights *domain.DimensionWeights) map[string]float64 {
	if weights == nil {
		return nil
	}
	return map[string]float64{
		"skill":          weights.Skill,
		"responsibility": weights.Responsibility,
		"experience":     weights.Experience,
		"education":      weights.Education,
		"industry":       weights.Industry,
		"basic":          weights.Basic,
	}
}
