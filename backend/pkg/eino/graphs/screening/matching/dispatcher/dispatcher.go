package dispatcher

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/domain"
)

// DispatchData 分发数据类型
type DispatchData map[string]any

// NewDispatchData 创建新的分发数据
func NewDispatchData(input *domain.MatchInput) DispatchData {
	if input == nil || input.JobProfile == nil || input.Resume == nil {
		return make(DispatchData)
	}

	data := make(DispatchData)

	// 基本信息 Agent 数据
	data[domain.BasicInfoAgent] = &domain.BasicInfoData{
		JobProfile: input.JobProfile,
		Resume:     input.Resume,
	}

	// 技能 Agent 数据
	data[domain.SkillAgent] = &domain.SkillData{
		JobSkills:      input.JobProfile.Skills,
		ResumeSkills:   input.Resume.Skills,
		ResumeProjects: input.Resume.Projects, // 新增：用于提取技术栈
	}

	// 职责 Agent 数据
	data[domain.ResponsibilityAgent] = &domain.ResponsibilityData{
		JobResponsibilities: input.JobProfile.Responsibilities,
		ResumeExperiences:   input.Resume.Experiences,
		ResumeProjects:      input.Resume.Projects, // 新增：项目职责匹配
	}

	// 经验 Agent 数据
	data[domain.ExperienceAgent] = &domain.ExperienceData{
		JobExperienceRequirements: input.JobProfile.ExperienceRequirements,
		ResumeExperiences:         input.Resume.Experiences,
		ResumeYearsExperience:     input.Resume.YearsExperience, // 新增：简历总工作年限
	}

	// 教育 Agent 数据
	data[domain.EducationAgent] = &domain.EducationData{
		JobEducationRequirements: input.JobProfile.EducationRequirements,
		ResumeEducations:         input.Resume.Educations,
	}

	// 行业 Agent 数据
	data[domain.IndustryAgent] = &domain.IndustryData{
		JobIndustryRequirements: input.JobProfile.IndustryRequirements,
		ResumeExperiences:       input.Resume.Experiences,
	}

	// 任务元数据
	data[domain.TaskMetaDataNode] = &domain.TaskMetaData{
		JobID:            input.JobProfile.ID,
		ResumeID:         input.Resume.ID,
		MatchTaskID:      input.MatchTaskID,
		DimensionWeights: input.DimensionWeights,
	}

	return data
}

// Dispatcher 分发器结构体
type Dispatcher struct{}

// NewDispatcher 创建新的分发器实例
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// Process 处理匹配输入，将数据分发给各个Agent
func (d *Dispatcher) Process(ctx context.Context, input *domain.MatchInput) (map[string]any, error) {
	// 验证输入数据
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	if input.JobProfile == nil || input.Resume == nil {
		return nil, fmt.Errorf("job profile and resume are required")
	}

	// 使用新的 NewDispatchData 函数创建分发数据
	data := NewDispatchData(input)

	// 检查是否有数据
	if len(data) == 0 {
		return nil, fmt.Errorf("failed to create dispatch data: input validation failed")
	}

	return data, nil
}

// GetAgentType 返回Agent类型
func (a *Dispatcher) GetAgentType() string {
	return "Dispatcher"
}

// GetVersion 返回Agent版本
func (a *Dispatcher) GetVersion() string {
	return "1.1.0"
}
