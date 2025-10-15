package domain

import (
	"time"
)

const (
	TaskMetaDataNode    = "task_meta_data"
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
type BasicMatchDetail struct {
	Score     float64            `json:"score"`
	SubScores map[string]float64 `json:"sub_scores"`
	Evidence  []string           `json:"evidence"`
	Notes     string             `json:"notes"`
}

// SkillMatchDetail 技能匹配详情
type SkillMatchDetail struct {
	Score         float64           `json:"score"`
	MatchedSkills []*MatchedSkill   `json:"matched_skills"` // 匹配的技能
	MissingSkills []*JobSkill       `json:"missing_skills"` // 缺失的技能
	ExtraSkills   []string          `json:"extra_skills"`   // 额外的技能
	LLMAnalysis   *SkillLLMAnalysis `json:"llm_analysis"`   // 大模型整体分析
}

// MatchedSkill 匹配的技能
type MatchedSkill struct {
	JobSkillID     string             `json:"job_skill_id"`
	ResumeSkillID  string             `json:"resume_skill_id,omitempty"`
	MatchType      string             `json:"match_type"`      // exact, semantic, related, none
	LLMScore       float64            `json:"llm_score"`       // 大模型评分 (0-100)
	ProficiencyGap float64            `json:"proficiency_gap"` // 熟练度差距
	Score          float64            `json:"score"`           // 该技能得分
	LLMAnalysis    *SkillItemAnalysis `json:"llm_analysis"`    // 大模型分析详情
}

// SkillLLMAnalysis 技能大模型分析
type SkillLLMAnalysis struct {
	OverallMatch    float64  `json:"overall_match"`   // 整体匹配度 (0-100)
	TechnicalFit    float64  `json:"technical_fit"`   // 技术契合度
	LearningCurve   string   `json:"learning_curve"`  // 学习曲线评估 (low/medium/high)
	StrengthAreas   []string `json:"strength_areas"`  // 优势技能领域
	GapAreas        []string `json:"gap_areas"`       // 技能缺口领域
	Recommendations []string `json:"recommendations"` // 技能提升建议
	AnalysisDetail  string   `json:"analysis_detail"` // 详细分析说明
}

// SkillItemAnalysis 技能项分析
type SkillItemAnalysis struct {
	MatchLevel      string  `json:"match_level"`      // 匹配等级: perfect/good/partial/none
	MatchPercentage float64 `json:"match_percentage"` // 匹配百分比 (0-100)
	ProficiencyGap  string  `json:"proficiency_gap"`  // 熟练度差距: none/minor/moderate/major
	Transferability string  `json:"transferability"`  // 技能可迁移性: high/medium/low
	LearningEffort  string  `json:"learning_effort"`  // 学习难度: minimal/moderate/significant
	MatchReason     string  `json:"match_reason"`     // 匹配原因说明
}

// ResponsibilityMatchDetail 职责匹配详情
type ResponsibilityMatchDetail struct {
	Score                     float64                  `json:"score"`
	MatchedResponsibilities   []*MatchedResponsibility `json:"matched_responsibilities"`   // 匹配的职责
	UnmatchedResponsibilities []*JobResponsibility     `json:"unmatched_responsibilities"` // 未匹配的职责
	RelevantExperiences       []string                 `json:"relevant_experiences"`       // 相关工作经历ID
}

// MatchedResponsibility 匹配的职责
type MatchedResponsibility struct {
	JobResponsibilityID string            `json:"job_responsibility_id"`
	ResumeExperienceID  string            `json:"resume_experience_id,omitempty"`
	LLMAnalysis         *LLMMatchAnalysis `json:"llm_analysis"` // 大模型分析结果
	MatchScore          float64           `json:"match_score"`  // 该职责匹配得分
	MatchReason         string            `json:"match_reason"` // 匹配原因说明
}

// LLMMatchAnalysis 大模型匹配分析
type LLMMatchAnalysis struct {
	MatchLevel         string   `json:"match_level"`         // 匹配等级: excellent/good/fair/poor
	MatchPercentage    float64  `json:"match_percentage"`    // 匹配百分比 (0-100)
	StrengthPoints     []string `json:"strength_points"`     // 匹配优势点
	WeakPoints         []string `json:"weak_points"`         // 不足之处
	RecommendedActions []string `json:"recommended_actions"` // 建议改进措施
	AnalysisDetail     string   `json:"analysis_detail"`     // 详细分析说明
}

// ExperienceMatchDetail 经验匹配详情
type ExperienceMatchDetail struct {
	Score           float64              `json:"score"`
	YearsMatch      *YearsMatchInfo      `json:"years_match"`
	PositionMatches []*PositionMatchInfo `json:"position_matches"`
	IndustryMatches []*IndustryMatchInfo `json:"industry_matches"`
}

// YearsMatchInfo 年限匹配信息
type YearsMatchInfo struct {
	RequiredYears float64 `json:"required_years"`
	ActualYears   float64 `json:"actual_years"`
	Score         float64 `json:"score"`
	Gap           float64 `json:"gap"`
}

// PositionMatchInfo 职位匹配信息
type PositionMatchInfo struct {
	ResumeExperienceID string  `json:"resume_experience_id"`
	Position           string  `json:"position"`
	Relevance          float64 `json:"relevance"`
	Score              float64 `json:"score"`
}

// IndustryMatchInfo 行业匹配信息
type IndustryMatchInfo struct {
	ResumeExperienceID string  `json:"resume_experience_id"`
	Company            string  `json:"company"`
	Industry           string  `json:"industry"`
	Relevance          float64 `json:"relevance"`
	Score              float64 `json:"score"`
}

// EducationMatchDetail 教育匹配详情
type EducationMatchDetail struct {
	Score         float64            `json:"score"`
	DegreeMatch   *DegreeMatchInfo   `json:"degree_match"`
	MajorMatches  []*MajorMatchInfo  `json:"major_matches"`
	SchoolMatches []*SchoolMatchInfo `json:"school_matches"`
}

// DegreeMatchInfo 学历匹配信息
type DegreeMatchInfo struct {
	RequiredDegree string  `json:"required_degree"`
	ActualDegree   string  `json:"actual_degree"`
	Score          float64 `json:"score"`
	Meets          bool    `json:"meets"`
}

// MajorMatchInfo 专业匹配信息
type MajorMatchInfo struct {
	ResumeEducationID string  `json:"resume_education_id"`
	Major             string  `json:"major"`
	Relevance         float64 `json:"relevance"`
	Score             float64 `json:"score"`
}

// SchoolMatchInfo 学校匹配信息
type SchoolMatchInfo struct {
	ResumeEducationID string  `json:"resume_education_id"`
	School            string  `json:"school"`
	Reputation        float64 `json:"reputation"`
	Score             float64 `json:"score"`
}

// IndustryMatchDetail 行业匹配详情
type IndustryMatchDetail struct {
	Score           float64              `json:"score"`
	IndustryMatches []*IndustryMatchInfo `json:"industry_matches"`
	CompanyMatches  []*CompanyMatchInfo  `json:"company_matches"`
}

// CompanyMatchInfo 公司匹配信息
type CompanyMatchInfo struct {
	ResumeExperienceID string  `json:"resume_experience_id"`
	Company            string  `json:"company"`
	TargetCompany      string  `json:"target_company"`
	Score              float64 `json:"score"`
	IsExact            bool    `json:"is_exact"`
}

// MatchLevel 匹配等级
type MatchLevel string

const (
	MatchLevelExcellent MatchLevel = "excellent" // 90-100分
	MatchLevelGood      MatchLevel = "good"      // 75-89分
	MatchLevelFair      MatchLevel = "fair"      // 60-74分
	MatchLevelPoor      MatchLevel = "poor"      // 0-59分
)

// GetMatchLevel 根据分数获取匹配等级
func GetMatchLevel(score float64) MatchLevel {
	switch {
	case score >= 90:
		return MatchLevelExcellent
	case score >= 75:
		return MatchLevelGood
	case score >= 60:
		return MatchLevelFair
	default:
		return MatchLevelPoor
	}
}

// DimensionWeights 维度权重
type DimensionWeights struct {
	Skill          float64 `json:"skill"`          // 技能匹配权重 35%
	Responsibility float64 `json:"responsibility"` // 职责匹配权重 20%
	Experience     float64 `json:"experience"`     // 工作经验权重 20%
	Education      float64 `json:"education"`      // 教育背景权重 15%
	Industry       float64 `json:"industry"`       // 行业背景权重 7%
	Basic          float64 `json:"basic"`          // 基本信息权重 3%
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

// 任务数据
type TaskMetaData struct {
	JobID            string            `json:"job_id"`
	ResumeID         string            `json:"resume_id"`
	MatchTaskID      string            `json:"match_task_id"`
	DimensionWeights *DimensionWeights `json:"dimension_weights"`
}
