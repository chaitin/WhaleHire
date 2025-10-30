package domain

import (
	"time"

	"github.com/chaitin/WhaleHire/backend/consts"
)

const (
	TaskMetaDataNode    = "TaskMetaDataNode"
	DispatcherNode      = "DispatcherNode"
	BasicInfoAgent      = "BasicInfoAgent"
	SkillAgent          = "SkillAgent"
	ResponsibilityAgent = "ResponsibilityAgent"
	ExperienceAgent     = "ExperienceAgent"
	EducationAgent      = "EducationAgent"
	IndustryAgent       = "IndustryAgent"
	AggregatorAgent     = "AggregatorAgent"
)

// MatchInput 匹配输入结构
type MatchInput struct {
	JobProfile       *JobProfileDetail `json:"job_profile"`
	Resume           *ResumeDetail     `json:"resume"`
	DimensionWeights *DimensionWeights `json:"weights"`
	MatchTaskID      string            `json:"match_task_id"`
}

// JobResumeMatch 工作简历匹配结果
// @Description 候选人简历与岗位要求的综合匹配分析结果，包含各维度匹配详情和最终综合评分
type JobResumeMatch struct {
	TaskMetaData        *TaskMetaData              `json:"task_meta_data"`
	OverallScore        float64                    `json:"overall_score"`        // 综合匹配度 (0-100)
	SkillMatch          *SkillMatchDetail          `json:"skill_match"`          // 技能匹配详情
	ResponsibilityMatch *ResponsibilityMatchDetail `json:"responsibility_match"` // 职责匹配详情
	ExperienceMatch     *ExperienceMatchDetail     `json:"experience_match"`     // 经验匹配详情
	EducationMatch      *EducationMatchDetail      `json:"education_match"`      // 教育匹配详情
	IndustryMatch       *IndustryMatchDetail       `json:"industry_match"`       // 行业匹配详情
	BasicMatch          *BasicMatchDetail          `json:"basic_match"`          // 基本信息匹配
	MatchedAt           time.Time                  `json:"matched_at"`           // 匹配时间
	Recommendations     []string                   `json:"recommendations"`      // 匹配建议
}

// BasicMatchDetail 基本信息匹配详情
// @Description 候选人基本信息与岗位要求的匹配分析结果，包含地点、薪资、部门等基础信息匹配
type BasicMatchDetail struct {
	// Score 基本信息匹配总分 (0-100)
	Score float64 `json:"score" example:"85.5"`
	// SubScores 各子项匹配分数，如地点、薪资等
	SubScores map[string]float64 `json:"sub_scores"`
	// Evidence 匹配证据列表
	Evidence []string `json:"evidence" example:"[\"工作地点匹配\",\"薪资期望符合\"]"`
	// Notes 匹配备注说明
	Notes string `json:"notes" example:"候选人基本信息与岗位要求高度匹配"`
}

// SkillMatchDetail 技能匹配详情
// @Description 候选人技能与岗位技能要求的匹配分析结果，包含匹配技能、缺失技能和LLM分析
type SkillMatchDetail struct {
	// Score 技能匹配总分 (0-100)
	Score float64 `json:"score" example:"78.5"`
	// MatchedSkills 匹配的技能列表
	MatchedSkills []*MatchedSkill `json:"matched_skills"`
	// MissingSkills 缺失的技能列表
	MissingSkills []*JobSkill `json:"missing_skills"`
	// ExtraSkills 额外的技能列表
	ExtraSkills []string `json:"extra_skills" example:"[\"Docker\",\"Kubernetes\"]"`
	// LLMAnalysis 大模型整体分析结果
	LLMAnalysis *SkillLLMAnalysis `json:"llm_analysis"`
}

// MatchedSkill 匹配的技能
type MatchedSkill struct {
	// JobSkillID 岗位技能ID
	JobSkillID string `json:"job_skill_id" example:"skill_001"`
	// ResumeSkillID 简历技能ID，可选
	ResumeSkillID string `json:"resume_skill_id,omitempty" example:"resume_skill_001"`
	// MatchType 匹配类型: exact(完全匹配), semantic(语义匹配), related(相关匹配), none(无匹配)
	MatchType string `json:"match_type" example:"exact" enums:"exact,semantic,related,none"`
	// LLMScore 大模型评分 (0-100)
	LLMScore float64 `json:"llm_score" example:"85.0"`
	// ProficiencyGap 熟练度差距
	ProficiencyGap float64 `json:"proficiency_gap" example:"0.2"`
	// Score 该技能匹配得分
	Score float64 `json:"score" example:"82.5"`
	// LLMAnalysis 大模型分析详情
	LLMAnalysis *SkillItemAnalysis `json:"llm_analysis"`
}

// SkillLLMAnalysis 技能大模型分析
type SkillLLMAnalysis struct {
	// OverallMatch 整体匹配度 (0-100)
	OverallMatch float64 `json:"overall_match" example:"78.5"`
	// TechnicalFit 技术契合度
	TechnicalFit float64 `json:"technical_fit" example:"85.0"`
	// LearningCurve 学习曲线评估: low(低), medium(中), high(高)
	LearningCurve string `json:"learning_curve" example:"medium" enums:"low,medium,high"`
	// StrengthAreas 优势技能领域
	StrengthAreas []string `json:"strength_areas" example:"[\"后端开发\",\"数据库设计\"]"`
	// GapAreas 技能缺口领域
	GapAreas []string `json:"gap_areas" example:"[\"前端框架\",\"云原生技术\"]"`
	// Recommendations 技能提升建议
	Recommendations []string `json:"recommendations" example:"[\"建议学习React框架\",\"加强Docker容器化技能\"]"`
	// AnalysisDetail 详细分析说明
	AnalysisDetail string `json:"analysis_detail" example:"候选人在后端开发方面经验丰富，但前端技能有待提升"`
}

// SkillItemAnalysis 技能项分析
type SkillItemAnalysis struct {
	// MatchLevel 匹配等级: perfect(完美), good(良好), partial(部分), none(无匹配)
	MatchLevel string `json:"match_level" example:"good" enums:"perfect,good,partial,none"`
	// MatchPercentage 匹配百分比 (0-100)
	MatchPercentage float64 `json:"match_percentage" example:"85.0"`
	// ProficiencyGap 熟练度差距: none(无差距), minor(轻微), moderate(中等), major(较大)
	ProficiencyGap string `json:"proficiency_gap" example:"minor" enums:"none,minor,moderate,major"`
	// Transferability 技能可迁移性: high(高), medium(中), low(低)
	Transferability string `json:"transferability" example:"high" enums:"high,medium,low"`
	// LearningEffort 学习难度: minimal(最小), moderate(中等), significant(较大)
	LearningEffort string `json:"learning_effort" example:"moderate" enums:"minimal,moderate,significant"`
	// MatchReason 匹配原因说明
	MatchReason string `json:"match_reason" example:"具备相关技术栈经验，可快速上手"`
}

// ResponsibilityMatchDetail 职责匹配详情
// @Description 候选人工作经历与岗位职责要求的匹配分析结果，包含匹配职责和相关经验
type ResponsibilityMatchDetail struct {
	// Score 职责匹配总分 (0-100)
	Score float64 `json:"score" example:"82.0"`
	// MatchedResponsibilities 匹配的职责列表
	MatchedResponsibilities []*MatchedResponsibility `json:"matched_responsibilities"`
	// UnmatchedResponsibilities 未匹配的职责列表
	UnmatchedResponsibilities []*JobResponsibility `json:"unmatched_responsibilities"`
	// RelevantExperiences 相关工作经历ID列表
	RelevantExperiences []string `json:"relevant_experiences" example:"[\"exp_001\",\"exp_002\"]"`
}

// MatchedResponsibility 匹配的职责
type MatchedResponsibility struct {
	// JobResponsibilityID 岗位职责ID
	JobResponsibilityID string `json:"job_responsibility_id" example:"resp_001"`
	// ResumeExperienceID 简历工作经历ID，可选
	ResumeExperienceID string `json:"resume_experience_id,omitempty" example:"exp_001"`
	// LLMAnalysis 大模型分析结果
	LLMAnalysis *LLMMatchAnalysis `json:"llm_analysis"`
	// MatchScore 该职责匹配得分 (0-100)
	MatchScore float64 `json:"match_score" example:"85.5"`
	// MatchReason 匹配原因说明
	MatchReason string `json:"match_reason" example:"具备相关项目管理经验"`
}

// LLMMatchAnalysis 大模型匹配分析
type LLMMatchAnalysis struct {
	// MatchLevel 匹配等级: excellent(优秀), good(良好), fair(一般), poor(较差)
	MatchLevel string `json:"match_level" example:"good" enums:"excellent,good,fair,poor"`
	// MatchPercentage 匹配百分比 (0-100)
	MatchPercentage float64 `json:"match_percentage" example:"85.0"`
	// StrengthPoints 匹配优势点
	StrengthPoints []string `json:"strength_points" example:"[\"项目管理经验丰富\",\"团队协作能力强\"]"`
	// WeakPoints 不足之处
	WeakPoints []string `json:"weak_points" example:"[\"缺乏敏捷开发经验\",\"跨部门沟通待提升\"]"`
	// RecommendedActions 建议改进措施
	RecommendedActions []string `json:"recommended_actions" example:"[\"参加敏捷开发培训\",\"加强跨部门协作\"]"`
	// AnalysisDetail 详细分析说明
	AnalysisDetail string `json:"analysis_detail" example:"候选人在项目管理方面有丰富经验，但需要加强敏捷开发方法论"`
}

// ExperienceMatchDetail 经验匹配详情
// @Description 候选人工作经验与岗位经验要求的匹配分析结果，包含年限、职位和行业匹配
type ExperienceMatchDetail struct {
	// Score 经验匹配总分 (0-100)
	Score float64 `json:"score" example:"78.5"`
	// YearsMatch 工作年限匹配信息
	YearsMatch *YearsMatchInfo `json:"years_match"`
	// PositionMatches 职位匹配列表
	PositionMatches []*PositionMatchInfo `json:"position_matches"`
	// IndustryMatches 行业匹配列表
	IndustryMatches []*IndustryMatchInfo `json:"industry_matches"`
	// CareerProgression 职业发展轨迹
	CareerProgression *CareerProgressionInfo `json:"career_progression,omitempty"`
	// OverallAnalysis 整体工作经验匹配度分析总结
	OverallAnalysis string `json:"overall_analysis,omitempty"`
}

// YearsMatchInfo 年限匹配信息
type YearsMatchInfo struct {
	// RequiredYears 要求工作年限
	RequiredYears float64 `json:"required_years" example:"5.0"`
	// ActualYears 实际工作年限
	ActualYears float64 `json:"actual_years" example:"6.5"`
	// Score 年限匹配得分 (0-100)
	Score float64 `json:"score" example:"95.0"`
	// Gap 年限差距（正数表示超出要求，负数表示不足）
	Gap float64 `json:"gap" example:"1.5"`
	// Analysis 年限匹配分析说明
	Analysis string `json:"analysis,omitempty" example:"候选人总年限满足岗位要求"`
}

// PositionMatchInfo 职位匹配信息
type PositionMatchInfo struct {
	// ResumeExperienceID 简历工作经历ID
	ResumeExperienceID string `json:"resume_experience_id" example:"exp_001"`
	// ExperienceType 简历经历类型（work/organization/volunteer/internship）
	ExperienceType string `json:"experience_type,omitempty" example:"work"`
	// Position 职位名称
	Position string `json:"position" example:"高级软件工程师"`
	// Company 公司名称
	Company string `json:"company,omitempty" example:"阿里巴巴"`
	// Relevance 相关度 (0-1)
	Relevance float64 `json:"relevance" example:"0.85"`
	// Score 职位匹配得分 (0-100)
	Score float64 `json:"score" example:"85.0"`
	// Analysis 职位匹配分析说明
	Analysis string `json:"analysis,omitempty" example:"核心职责与目标岗位高度一致"`
}

// IndustryMatchInfo 行业匹配信息
type IndustryMatchInfo struct {
	// ResumeExperienceID 简历工作经历ID
	ResumeExperienceID string `json:"resume_experience_id" example:"exp_001"`
	// Company 公司名称
	Company string `json:"company" example:"腾讯科技"`
	// Industry 行业名称
	Industry string `json:"industry" example:"互联网"`
	// Relevance 相关度 (0-1)
	Relevance float64 `json:"relevance" example:"0.90"`
	// Score 行业匹配得分 (0-100)
	Score float64 `json:"score" example:"90.0"`
	// Analysis 行业匹配分析说明
	Analysis string `json:"analysis,omitempty" example:"行业业务模式与目标岗位高度一致"`
}

// CareerProgressionInfo 职业发展轨迹信息
type CareerProgressionInfo struct {
	// Score 职业发展轨迹得分 (0-100)
	Score float64 `json:"score" example:"85.0"`
	// Trend 职业发展趋势
	Trend string `json:"trend,omitempty" example:"上升"`
	// Analysis 职业发展轨迹分析说明
	Analysis string `json:"analysis,omitempty" example:"从初级到高级岗位稳步晋升"`
}

// EducationMatchDetail 教育匹配详情
// @Description 候选人教育背景与岗位教育要求的匹配分析结果，包含学历、专业和学校匹配
type EducationMatchDetail struct {
	// Score 教育匹配总分 (0-100)
	Score float64 `json:"score" example:"88.0"`
	// DegreeMatch 学历匹配信息
	DegreeMatch *DegreeMatchInfo `json:"degree_match"`
	// MajorMatches 专业匹配列表
	MajorMatches []*MajorMatchInfo `json:"major_matches"`
	// SchoolMatches 学校匹配列表
	SchoolMatches []*SchoolMatchInfo `json:"school_matches"`
	// OverallAnalysis 整体教育背景匹配度分析总结
	OverallAnalysis string `json:"overall_analysis,omitempty"`
}

// DegreeMatchInfo 学历匹配信息
type DegreeMatchInfo struct {
	// RequiredDegree 要求学历
	RequiredDegree string `json:"required_degree" example:"本科"`
	// ActualDegree 实际学历
	ActualDegree string `json:"actual_degree" example:"硕士"`
	// Score 学历匹配得分 (0-100)
	Score float64 `json:"score" example:"100.0"`
	// Meets 是否满足学历要求
	Meets bool `json:"meets" example:"true"`
}

// MajorMatchInfo 专业匹配信息
type MajorMatchInfo struct {
	// ResumeEducationID 简历教育经历ID
	ResumeEducationID string `json:"resume_education_id" example:"edu_001"`
	// Major 专业名称
	Major string `json:"major" example:"计算机科学与技术"`
	// Relevance 相关度 (0-100)
	Relevance float64 `json:"relevance" example:"95.0"`
	// Score 专业匹配得分 (0-100)
	Score float64 `json:"score" example:"95.0"`
}

// SchoolMatchInfo 学校匹配信息
type SchoolMatchInfo struct {
	// ResumeEducationID 简历教育经历ID
	ResumeEducationID string `json:"resume_education_id" example:"edu_001"`
	// School 学校名称
	School string `json:"school" example:"清华大学"`
	// Degree 学位等级
	Degree string `json:"degree,omitempty" example:"硕士"`
	// Major 专业名称
	Major string `json:"major,omitempty" example:"计算机科学与技术"`
	// GraduationYear 毕业年份
	GraduationYear int `json:"graduation_year,omitempty" example:"2022"`
	// Reputation 学校声誉评分 (0-100)
	Reputation float64 `json:"reputation" example:"98.0"`
	// Score 学校匹配得分 (0-100)
	Score float64 `json:"score" example:"98.0"`
	// UniversityTypes 学校类型标签，如 985/211/QS Top 等
	UniversityTypes []consts.UniversityType `json:"university_types,omitempty" example:"[\"985\",\"double_first_class\"]"`
	// GPA 候选人在该学校的绩点
	GPA *float64 `json:"gpa,omitempty" example:"3.8"`
	// Analysis 学校匹配分析说明
	Analysis string `json:"analysis,omitempty" example:"院校声誉与岗位期望高度吻合"`
}

// IndustryMatchDetail 行业匹配详情
// @Description 候选人行业背景与岗位行业要求的匹配分析结果，包含行业和公司匹配
type IndustryMatchDetail struct {
	// Score 行业匹配总分 (0-100)
	Score float64 `json:"score" example:"85.0"`
	// IndustryMatches 行业匹配列表
	IndustryMatches []*IndustryMatchInfo `json:"industry_matches"`
	// CompanyMatches 公司匹配列表
	CompanyMatches []*CompanyMatchInfo `json:"company_matches"`
	// IndustryDepth 行业深度评估
	IndustryDepth *IndustryDepthInfo `json:"industry_depth,omitempty"`
	// OverallAnalysis 整体行业背景匹配度分析总结
	OverallAnalysis string `json:"overall_analysis,omitempty"`
}

// IndustryDepthInfo 行业深度信息
// @Description 候选人在相关行业的工作年限与深度分析
type IndustryDepthInfo struct {
	// TotalYears 在相关行业总工作年限
	TotalYears float64 `json:"total_years" example:"5.0"`
	// Score 行业深度分数 (0-100)
	Score float64 `json:"score" example:"90.0"`
	// Analysis 行业深度分析说明
	Analysis string `json:"analysis" example:"在互联网行业有超过5年经验，具备深厚的行业积累"`
}

// CompanyMatchInfo 公司匹配信息
type CompanyMatchInfo struct {
	// ResumeExperienceID 简历工作经历ID
	ResumeExperienceID string `json:"resume_experience_id" example:"exp_001"`
	// Company 简历中的公司名称
	Company string `json:"company" example:"阿里巴巴"`
	// TargetCompany 目标公司名称
	TargetCompany string `json:"target_company" example:"腾讯"`
	// CompanySize 公司规模
	CompanySize string `json:"company_size,omitempty" example:"大型"`
	// Reputation 公司声誉分数 (0-100)
	Reputation float64 `json:"reputation,omitempty" example:"85.0"`
	// Score 公司匹配得分 (0-100)
	Score float64 `json:"score" example:"88.0"`
	// IsExact 是否完全匹配
	IsExact bool `json:"is_exact" example:"false"`
	// Analysis 公司匹配分析说明
	Analysis string `json:"analysis,omitempty" example:"公司规模与目标公司匹配，行业地位相当"`
}

// DimensionWeights 维度权重
// @Description 各个匹配维度的权重配置，用于计算综合匹配度
type DimensionWeights struct {
	// Skill 技能匹配权重，默认35%
	Skill float64 `json:"skill"`
	// Responsibility 职责匹配权重，默认20%
	Responsibility float64 `json:"responsibility" `
	// Experience 工作经验权重，默认20%
	Experience float64 `json:"experience"`
	// Education 教育背景权重，默认15%
	Education float64 `json:"education" `
	// Industry 行业背景权重，默认7%
	Industry float64 `json:"industry"`
	// Basic 基本信息权重，默认3%
	Basic float64 `json:"basic" `
}

// DefaultDimensionWeights 默认维度权重
var DefaultDimensionWeights = DimensionWeights{
	Skill:          0.35,
	Responsibility: 0.20,
	Experience:     0.20,
	Education:      0.15,
	Industry:       0.07,
	Basic:          0.03,
}

// Agent数据结构

// BasicInfoData 基本信息Agent数据
type BasicInfoData struct {
	JobProfile *JobProfileDetail `json:"job_profile"`
	Resume     *ResumeDetail     `json:"resume"`
}

// SkillData 技能Agent数据
type SkillData struct {
	JobSkills      []*JobSkill      `json:"job_skills"`
	ResumeSkills   []*ResumeSkill   `json:"resume_skills"`
	ResumeProjects []*ResumeProject `json:"resume_projects"` // 新增：用于提取技术栈
}

// ResponsibilityData 职责Agent数据
type ResponsibilityData struct {
	JobResponsibilities []*JobResponsibility `json:"job_responsibilities"`
	ResumeExperiences   []*ResumeExperience  `json:"resume_experiences"`
	ResumeProjects      []*ResumeProject     `json:"resume_projects"` // 新增：项目职责匹配
}

// ExperienceData 经验Agent数据
type ExperienceData struct {
	JobExperienceRequirements []*JobExperienceRequirement `json:"job_experience_requirements"`
	ResumeExperiences         []*ResumeExperience         `json:"resume_experiences"`
	ResumeYearsExperience     float64                     `json:"resume_years_experience"` // 新增：简历总工作年限
}

// EducationData 教育Agent数据
type EducationData struct {
	JobEducationRequirements []*JobEducationRequirement `json:"job_education_requirements"`
	ResumeEducations         []*ResumeEducation         `json:"resume_educations"`
}

// IndustryData 行业Agent数据
type IndustryData struct {
	JobIndustryRequirements []*JobIndustryRequirement `json:"job_industry_requirements"`
	ResumeExperiences       []*ResumeExperience       `json:"resume_experiences"`
}

// TaskMetaData 任务元数据
// @Description 筛选任务的元数据信息，包含任务ID、简历ID、匹配任务ID和维度权重
type TaskMetaData struct {
	// JobID 岗位ID
	JobID string `json:"job_id" `
	// ResumeID 简历ID
	ResumeID string `json:"resume_id"`
	// MatchTaskID 匹配任务ID
	MatchTaskID string `json:"match_task_id" `
	// DimensionWeights 维度权重配置
	DimensionWeights *DimensionWeights `json:"dimension_weights" `
}
