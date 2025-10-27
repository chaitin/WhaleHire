package jobprofileparser

// JobProfileParseInput 岗位画像解析输入
type JobProfileParseInput struct {
	Description string `json:"description"` // 岗位描述文本
}

// JobProfileParseResult 岗位画像解析结果，精简版本，只包含AI解析需要的核心数据
type JobProfileParseResult struct {
	// 基本信息
	Name      string   `json:"name"`
	WorkType  *string  `json:"work_type,omitempty"`
	Location  *string  `json:"location,omitempty"`
	SalaryMin *float64 `json:"salary_min,omitempty"`
	SalaryMax *float64 `json:"salary_max,omitempty"`

	// 详细信息 - 只包含业务核心字段
	Responsibilities       []*JobResponsibilitySimple        `json:"responsibilities"`
	Skills                 []*JobSkillSimple                 `json:"skills"`
	EducationRequirements  []*JobEducationRequirementSimple  `json:"education_requirements"`
	ExperienceRequirements []*JobExperienceRequirementSimple `json:"experience_requirements"`
	IndustryRequirements   []*JobIndustryRequirementSimple   `json:"industry_requirements"`
}

// 简化的子结构体，移除ID等系统字段
type JobResponsibilitySimple struct {
	Responsibility string `json:"responsibility"`
}

type JobSkillSimple struct {
	Skill string `json:"skill"`
	Type  string `json:"type"` // required/bonus
}

type JobEducationRequirementSimple struct {
	EducationType string `json:"education_type"`
}

type JobExperienceRequirementSimple struct {
	ExperienceType string `json:"experience_type"`
	MinYears       int    `json:"min_years"`
	IdealYears     int    `json:"ideal_years"`
}

type JobIndustryRequirementSimple struct {
	Industry    string  `json:"industry"`
	CompanyName *string `json:"company_name,omitempty"`
}
