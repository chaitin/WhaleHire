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
	Match           *domain.JobResumeMatch
	TokenUsages     map[string]*model.TokenUsage
	Version         string
	SubAgentVersion map[string]string
	Duration        time.Duration
	DimensionMap    map[string]float64
	Collector       *screening.AgentCallbackCollector
}

type matchingService struct {
	logger            *slog.Logger
	cfg               *config.Config
	factory           *models.ModelFactory
	version           string
	sub_agent_version map[string]string
	nodeRunRepo       domain.ScreeningNodeRunRepo
	screeningRepo     domain.ScreeningRepo

	// 图编译缓存相关
	compiledGraph    compose.Runnable[*domain.MatchInput, *domain.JobResumeMatch]
	currentGraph     *screening.ScreeningChatGraph
	currentModelType models.ModelType
	currentModelName string
	compileMutex     sync.RWMutex
}

// NewMatchingService 创建匹配服务
func NewMatchingService(cfg *config.Config, logger *slog.Logger, nodeRunRepo domain.ScreeningNodeRunRepo, screeningRepo domain.ScreeningRepo) (MatchingService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		logger = slog.Default()
	}
	if nodeRunRepo == nil {
		return nil, fmt.Errorf("nodeRunRepo is required")
	}
	if screeningRepo == nil {
		return nil, fmt.Errorf("screeningRepo is required")
	}

	factory := models.NewModelFactory()

	return &matchingService{
		logger:        logger,
		cfg:           cfg,
		factory:       factory,
		version:       "1.0.0",
		nodeRunRepo:   nodeRunRepo,
		screeningRepo: screeningRepo,
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

	// 获取编译后的图（支持并发复用）
	compiledGraph, err := s.ensureCompiledGraph(ctx, modelType, modelName)
	if err != nil {
		return nil, fmt.Errorf("ensure compiled graph failed: %w", err)
	}

	weights := buildDimensionWeights(req.DimensionWeights)
	matchInput := &domain.MatchInput{
		JobProfile:       req.JobProfile,
		Resume:           req.Resume,
		DimensionWeights: weights,
		MatchTaskID:      fmt.Sprintf("%s:%s", req.TaskID.String(), req.ResumeID.String()),
	}

	// 创建回调收集器并传入包装器，复用同一份回调状态
	collector := screening.NewAgentCallbackCollector()
	wrapper := NewCallbackCollectorWrapper(collector, s.nodeRunRepo, s.screeningRepo, s.currentGraph, req.TaskID, req.ResumeID, req.TaskID.String(), s.logger)

	// 包装器内部已包含收集器和数据库回调
	allOptions := wrapper.ComposeOptions()

	start := time.Now()
	match, err := compiledGraph.Invoke(ctx, matchInput, allOptions...)
	duration := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("invoke screening graph failed: %w", err)
	}

	result := &MatchResult{
		Match:           match,
		TokenUsages:     collector.TokenUsages(),
		Version:         s.version,
		SubAgentVersion: s.sub_agent_version,
		Duration:        duration,
		DimensionMap:    weightsToMap(weights),
		Collector:       collector,
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

// ensureCompiledGraph 确保图已编译并可复用
func (s *matchingService) ensureCompiledGraph(ctx context.Context, modelType models.ModelType, modelName string) (compose.Runnable[*domain.MatchInput, *domain.JobResumeMatch], error) {
	// 快速路径：如果已有编译好的图且模型配置未变化，直接返回
	s.compileMutex.RLock()
	if s.compiledGraph != nil && s.currentModelType == modelType && s.currentModelName == modelName {
		graph := s.compiledGraph
		s.compileMutex.RUnlock()
		return graph, nil
	}
	s.compileMutex.RUnlock()

	// 慢速路径：需要重新编译图
	s.compileMutex.Lock()
	defer s.compileMutex.Unlock()

	// 双重检查：可能在等待锁的过程中其他 goroutine 已经完成了编译
	if s.compiledGraph != nil && s.currentModelType == modelType && s.currentModelName == modelName {
		return s.compiledGraph, nil
	}

	// 获取模型
	chatModel, err := s.factory.GetModel(ctx, modelType, modelName)
	if err != nil {
		return nil, fmt.Errorf("获取对话模型失败: %w", err)
	}

	// 构建图
	graph, err := screening.NewScreeningChatGraph(ctx, chatModel, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("构建智能筛选图失败: %w", err)
	}

	// 编译图
	compiledGraph, err := graph.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译智能筛选图失败: %w", err)
	}

	// 缓存编译结果和模型配置
	s.compiledGraph = compiledGraph
	s.currentGraph = graph
	s.currentModelType = modelType
	s.currentModelName = modelName
	s.version = graph.GetVersion()
	s.sub_agent_version = graph.GetSubAgentVersions()

	return compiledGraph, nil
}

// InvalidateCompiledGraph 使编译后的图失效，强制下次调用时重新编译
// 当模型配置或图结构发生变化时调用此方法
func (s *matchingService) InvalidateCompiledGraph() {
	s.compileMutex.Lock()
	defer s.compileMutex.Unlock()

	s.compiledGraph = nil
	s.currentModelType = ""
	s.currentModelName = ""
}

// GetCompiledGraphInfo 获取当前编译图的信息（用于调试和监控）
func (s *matchingService) GetCompiledGraphInfo() (modelType models.ModelType, modelName string, isCompiled bool) {
	s.compileMutex.RLock()
	defer s.compileMutex.RUnlock()

	return s.currentModelType, s.currentModelName, s.compiledGraph != nil
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
