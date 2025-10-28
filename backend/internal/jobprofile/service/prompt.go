package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofilegenerator"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofileparser"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofilepolisher"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
	"github.com/cloudwego/eino/compose"
)

// JobProfilePromptService AI 岗位画像 Prompt 服务
type JobProfilePromptService struct {
	modelFactory  *models.ModelFactory
	config        *config.Config
	logger        *slog.Logger
	skillMetaRepo domain.JobSkillMetaRepo

	// 预编译的 Chain
	polisherChain  compose.Runnable[*jobprofilepolisher.PolishJobPromptInput, *jobprofilepolisher.PolishJobPromptResult]
	generatorChain compose.Runnable[*jobprofilegenerator.JobProfileGenerateInput, *jobprofilegenerator.JobProfileGenerateResult]
}

// NewJobProfilePromptService 创建 AI 岗位画像 Prompt 服务
func NewJobProfilePromptService(config *config.Config, logger *slog.Logger, skillMetaRepo domain.JobSkillMetaRepo) *JobProfilePromptService {
	factory := models.NewModelFactory()

	// 注册OpenAI模型
	openaiConfig := &models.OpenAIConfig{
		APIKey:         config.GeneralAgent.LLM.APIKey,
		BaseURL:        config.GeneralAgent.LLM.BaseURL,
		Model:          config.GeneralAgent.LLM.ModelName,
		ResponseFormat: "json_object",
	}
	openaiManager := models.NewOpenAIModelManager(openaiConfig)
	factory.Register(models.ModelTypeOpenAI, config.GeneralAgent.LLM.ModelName, openaiManager)

	// 创建服务实例
	service := &JobProfilePromptService{
		modelFactory:  factory,
		config:        config,
		logger:        logger,
		skillMetaRepo: skillMetaRepo,
	}

	// 预编译 Chain
	ctx := context.Background()
	if err := service.initializeChains(ctx); err != nil {
		logger.Error("Failed to initialize chains", "error", err)
		// 返回未初始化的服务，让运行时处理错误
	}

	return service
}

// initializeChains 初始化并编译所有 Chain
func (s *JobProfilePromptService) initializeChains(ctx context.Context) error {
	// 获取模型
	chatModel, err := s.modelFactory.GetModel(ctx, models.ModelTypeOpenAI, s.config.GeneralAgent.LLM.ModelName)
	if err != nil {
		return fmt.Errorf("failed to get chat model: %w", err)
	}

	// 创建并编译润色链
	polisherChain, err := jobprofilepolisher.NewJobProfilePolisherChain(ctx, chatModel)
	if err != nil {
		return fmt.Errorf("failed to create polisher chain: %w", err)
	}
	s.polisherChain, err = polisherChain.Compile(ctx)
	if err != nil {
		return fmt.Errorf("failed to compile polisher chain: %w", err)
	}

	// 创建并编译生成链
	generatorChain, err := jobprofilegenerator.NewJobProfileGeneratorChain(ctx, chatModel)
	if err != nil {
		return fmt.Errorf("failed to create generator chain: %w", err)
	}
	s.generatorChain, err = generatorChain.Compile(ctx)
	if err != nil {
		return fmt.Errorf("failed to compile generator chain: %w", err)
	}

	return nil
}

// PolishPrompt 润色岗位需求描述
func (s *JobProfilePromptService) PolishPrompt(ctx context.Context, req *domain.PolishJobPromptReq) (*domain.PolishJobPromptResp, error) {
	// 参数校验
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	idea := strings.TrimSpace(req.Idea)
	if idea == "" {
		return nil, fmt.Errorf("idea is required")
	}

	// 长度限制检查（5-500个字符）
	if len([]rune(idea)) < 5 || len([]rune(idea)) > 500 {
		return nil, fmt.Errorf("idea length must be between 5 and 500 characters")
	}

	// 检查 Chain 是否已初始化
	if s.polisherChain == nil {
		s.logger.Error("Polisher chain not initialized")
		return nil, fmt.Errorf("AI service temporarily unavailable, please try again later")
	}

	// 准备输入
	input := &jobprofilepolisher.PolishJobPromptInput{
		Idea: idea,
	}

	// 执行润色
	result, err := s.polisherChain.Invoke(ctx, input)
	if err != nil {
		s.logger.Error("Failed to polish job prompt", "error", err, "idea_hash", hashString(idea))
		return nil, fmt.Errorf("AI service temporarily unavailable, please try again later")
	}

	// 转换结果
	resp := &domain.PolishJobPromptResp{
		PolishedPrompt:     result.PolishedPrompt,
		SuggestedTitle:     result.SuggestedTitle,
		ResponsibilityTips: result.ResponsibilityTips,
		RequirementTips:    result.RequirementTips,
		BonusTips:          result.BonusTips,
	}

	// 记录日志
	s.logger.Info("ai.job_profile.polish",
		"suggested_title", result.SuggestedTitle,
		"responsibility_tips_count", len(result.ResponsibilityTips),
		"requirement_tips_count", len(result.RequirementTips),
		"bonus_tips_count", len(result.BonusTips),
	)

	return resp, nil
}

// GenerateByPrompt 基于 Prompt 生成岗位画像
func (s *JobProfilePromptService) GenerateByPrompt(ctx context.Context, req *domain.GenerateJobProfileReq) (*domain.GenerateJobProfileResp, error) {
	// 参数校验
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// 检查 Chain 是否已初始化
	if s.generatorChain == nil {
		s.logger.Error("Generator chain not initialized")
		return nil, fmt.Errorf("AI service temporarily unavailable, please try again later")
	}

	// 准备输入
	input := &jobprofilegenerator.JobProfileGenerateInput{
		Prompt: prompt,
	}

	// 执行生成
	result, err := s.generatorChain.Invoke(ctx, input)
	if err != nil {
		s.logger.Error("Failed to generate job profile", "error", err, "prompt_hash", hashString(prompt))
		return nil, fmt.Errorf("AI service temporarily unavailable, please try again later")
	}

	// 转换结果
	resp := &domain.GenerateJobProfileResp{
		Profile:             s.transformParseResult(ctx, result.Profile),
		DescriptionMarkdown: result.DescriptionMarkdown,
	}

	// 记录日志
	s.logger.Info("ai.job_profile.generate",
		"profile_name", result.Profile.Name,
		"responsibilities_count", len(result.Profile.Responsibilities),
		"skills_count", len(result.Profile.Skills),
		"description_length", len(result.DescriptionMarkdown),
	)

	return resp, nil
}

// transformGenerateResult 将 jobprofilegenerator 的结果转换为 domain 层的响应
func (s *JobProfilePromptService) transformParseResult(ctx context.Context, profile *jobprofileparser.JobProfileParseResult) *domain.ParseJobProfileResp {
	if profile == nil {
		return &domain.ParseJobProfileResp{}
	}

	// 构建 ParseJobProfileResp 结构
	parseResp := &domain.ParseJobProfileResp{
		Name:      profile.Name,
		WorkType:  profile.WorkType,
		Location:  profile.Location,
		SalaryMin: profile.SalaryMin,
		SalaryMax: profile.SalaryMax,
	}

	// 转换职责
	for _, r := range profile.Responsibilities {
		if r != nil {
			parseResp.Responsibilities = append(parseResp.Responsibilities, &domain.JobResponsibilityInput{
				Responsibility: r.Responsibility,
			})
		}
	}

	// 转换技能
	for _, skill := range profile.Skills {
		if skill != nil {
			skillID := ""
			// 从技能字典中查找技能ID
			if skillMeta, err := s.skillMetaRepo.GetByName(ctx, skill.Skill); err == nil {
				skillID = skillMeta.ID.String()
			}
			// 如果查找失败（技能不存在），skillID保持为空字符串

			parseResp.Skills = append(parseResp.Skills, &domain.JobSkillInput{
				SkillID:   &skillID,
				SkillName: skill.Skill,
				Type:      skill.Type,
			})
		}
	}

	// 转换学历要求
	for _, e := range profile.EducationRequirements {
		if e != nil {
			parseResp.EducationRequirements = append(parseResp.EducationRequirements, &domain.JobEducationRequirementInput{
				EducationType: e.EducationType,
			})
		}
	}

	// 转换经验要求
	for _, e := range profile.ExperienceRequirements {
		if e != nil {
			parseResp.ExperienceRequirements = append(parseResp.ExperienceRequirements, &domain.JobExperienceRequirementInput{
				ExperienceType: e.ExperienceType,
				MinYears:       e.MinYears,
				IdealYears:     e.IdealYears,
			})
		}
	}

	// 转换行业要求
	for _, i := range profile.IndustryRequirements {
		if i != nil {
			parseResp.IndustryRequirements = append(parseResp.IndustryRequirements, &domain.JobIndustryRequirementInput{
				Industry:    i.Industry,
				CompanyName: i.CompanyName,
			})
		}
	}

	return parseResp
}

// hashString 对字符串进行简单哈希，用于日志记录
func hashString(s string) string {
	if len(s) <= 10 {
		return "***"
	}
	return fmt.Sprintf("%x", len(s))
}
