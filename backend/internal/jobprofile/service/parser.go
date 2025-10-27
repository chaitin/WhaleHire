package service

import (
	"context"
	"log/slog"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofileparser"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
)

type JobProfileParserService struct {
	modelFactory  *models.ModelFactory
	config        *config.Config
	logger        *slog.Logger
	skillMetaRepo domain.JobSkillMetaRepo
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

	return &JobProfileParserService{
		modelFactory:  factory,
		config:        config,
		logger:        logger,
		skillMetaRepo: skillMetaRepo,
	}
}

// ParseJobProfile 解析职位描述
func (s *JobProfileParserService) ParseJobProfile(ctx context.Context, req *domain.ParseJobProfileReq) (*domain.ParseJobProfileResp, error) {
	// 获取模型
	chatModel, err := s.modelFactory.GetModel(ctx, models.ModelTypeOpenAI, s.config.GeneralAgent.LLM.ModelName)
	if err != nil {
		s.logger.Error("Failed to get chat model", "error", err)
		return nil, err
	}

	// 创建解析链
	parserChain, err := jobprofileparser.NewJobProfileParserChain(ctx, chatModel)
	if err != nil {
		s.logger.Error("Failed to create job profile parser chain", "error", err)
		return nil, err
	}

	// 编译解析链
	compiledChain, err := parserChain.Compile(ctx)
	if err != nil {
		s.logger.Error("Failed to compile job profile parser chain", "error", err)
		return nil, err
	}

	// 准备输入
	input := &jobprofileparser.JobProfileParseInput{
		Description: req.Description,
	}

	// 执行解析
	result, err := compiledChain.Invoke(ctx, input)
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
