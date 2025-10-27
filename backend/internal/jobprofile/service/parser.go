package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofileparser"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
	"github.com/cloudwego/eino/compose"
)

type JobProfileParserService struct {
	modelFactory  *models.ModelFactory
	config        *config.Config
	logger        *slog.Logger
	skillMetaRepo domain.JobSkillMetaRepo

	// 预编译的 Chain
	parserChain compose.Runnable[*jobprofileparser.JobProfileParseInput, *jobprofileparser.JobProfileParseResult]
}

// NewJobProfileParserService 创建职位解析服务
func NewJobProfileParserService(config *config.Config, logger *slog.Logger, skillMetaRepo domain.JobSkillMetaRepo) *JobProfileParserService {
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
	service := &JobProfileParserService{
		modelFactory:  factory,
		config:        config,
		logger:        logger,
		skillMetaRepo: skillMetaRepo,
	}

	// 预编译 Chain
	ctx := context.Background()
	if err := service.initializeChain(ctx); err != nil {
		logger.Error("Failed to initialize parser chain", "error", err)
		// 返回未初始化的服务，让运行时处理错误
	}

	return service
}

// initializeChain 初始化并编译解析链
func (s *JobProfileParserService) initializeChain(ctx context.Context) error {
	// 获取模型
	chatModel, err := s.modelFactory.GetModel(ctx, models.ModelTypeOpenAI, s.config.GeneralAgent.LLM.ModelName)
	if err != nil {
		return fmt.Errorf("failed to get chat model: %w", err)
	}

	// 创建并编译解析链
	parserChain, err := jobprofileparser.NewJobProfileParserChain(ctx, chatModel)
	if err != nil {
		return fmt.Errorf("failed to create parser chain: %w", err)
	}
	s.parserChain, err = parserChain.Compile(ctx)
	if err != nil {
		return fmt.Errorf("failed to compile parser chain: %w", err)
	}

	return nil
}

// ParseJobProfile 解析职位描述
func (s *JobProfileParserService) ParseJobProfile(ctx context.Context, req *domain.ParseJobProfileReq) (*domain.ParseJobProfileResp, error) {
	// 检查 Chain 是否已初始化
	if s.parserChain == nil {
		s.logger.Error("Parser chain not initialized")
		return nil, fmt.Errorf("AI service temporarily unavailable, please try again later")
	}

	// 准备输入
	input := &jobprofileparser.JobProfileParseInput{
		Description: req.Description,
	}

	// 执行解析
	result, err := s.parserChain.Invoke(ctx, input)
	if err != nil {
		s.logger.Error("Failed to parse job profile", "error", err)
		return nil, err
	}

	// 转换结果
	resp, err := s.transformParseResult(ctx, result)
	if err != nil {
		s.logger.Error("failed to transform parse result", "error", err)
		return nil, err
	}

	return resp, nil
}

// transformParseResult 转换解析结果
func (s *JobProfileParserService) transformParseResult(ctx context.Context, result *jobprofileparser.JobProfileParseResult) (*domain.ParseJobProfileResp, error) {
	resp := &domain.ParseJobProfileResp{
		Name:      result.Name,
		WorkType:  result.WorkType,
		Location:  result.Location,
		SalaryMin: result.SalaryMin,
		SalaryMax: result.SalaryMax,
	}

	// 转换职责要求
	for _, responsibility := range result.Responsibilities {
		resp.Responsibilities = append(resp.Responsibilities, &domain.JobResponsibilityInput{
			Responsibility: responsibility.Responsibility,
		})
	}

	// 转换技能要求
	for _, skill := range result.Skills {
		skillID := ""
		// 从技能字典中查找技能ID
		if skillMeta, err := s.skillMetaRepo.GetByName(ctx, skill.Skill); err == nil {
			skillID = skillMeta.ID.String()
		}
		// 如果查找失败（技能不存在），skillID保持为空字符串

		resp.Skills = append(resp.Skills, &domain.JobSkillInput{
			SkillID:   &skillID,
			SkillName: skill.Skill,
			Type:      skill.Type,
		})
	}

	// 转换教育要求
	for _, education := range result.EducationRequirements {
		resp.EducationRequirements = append(resp.EducationRequirements, &domain.JobEducationRequirementInput{
			EducationType: education.EducationType,
		})
	}

	// 转换经验要求
	for _, experience := range result.ExperienceRequirements {
		resp.ExperienceRequirements = append(resp.ExperienceRequirements, &domain.JobExperienceRequirementInput{
			ExperienceType: experience.ExperienceType,
			MinYears:       experience.MinYears,
			IdealYears:     experience.IdealYears,
		})
	}

	// 转换行业要求
	for _, industry := range result.IndustryRequirements {
		resp.IndustryRequirements = append(resp.IndustryRequirements, &domain.JobIndustryRequirementInput{
			Industry:    industry.Industry,
			CompanyName: industry.CompanyName,
		})
	}

	return resp, nil
}
