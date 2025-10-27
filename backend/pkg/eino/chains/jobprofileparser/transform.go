package jobprofileparser

import (
	"encoding/json"
	"strings"
)

// rawJobProfileParseResult 用于承接模型返回的原始JSON，方便进行二次清洗
type rawJobProfileParseResult struct {
	// 基本信息
	Name      string   `json:"name"`
	WorkType  *string  `json:"work_type"`
	Location  *string  `json:"location"`
	SalaryMin *float64 `json:"salary_min"`
	SalaryMax *float64 `json:"salary_max"`

	// 详细信息 - 简化版本，移除ID字段
	Responsibilities       []rawJobResponsibility        `json:"responsibilities"`
	Skills                 []rawJobSkill                 `json:"skills"`
	EducationRequirements  []rawJobEducationRequirement  `json:"education_requirements"`
	ExperienceRequirements []rawJobExperienceRequirement `json:"experience_requirements"`
	IndustryRequirements   []rawJobIndustryRequirement   `json:"industry_requirements"`
}

type rawJobResponsibility struct {
	Responsibility string `json:"responsibility"`
}

type rawJobSkill struct {
	Skill string `json:"skill"`
	Type  string `json:"type"`
}

type rawJobEducationRequirement struct {
	EducationType string `json:"education_type"`
}

type rawJobExperienceRequirement struct {
	ExperienceType string `json:"experience_type"`
	MinYears       int    `json:"min_years"`
	IdealYears     int    `json:"ideal_years"`
}

type rawJobIndustryRequirement struct {
	Industry    string  `json:"industry"`
	CompanyName *string `json:"company_name"`
}

func decodeJobProfileParseResult(data []byte) (*JobProfileParseResult, error) {
	var raw rawJobProfileParseResult
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := &JobProfileParseResult{
		Name:                   strings.TrimSpace(raw.Name),
		WorkType:               normalizeWorkType(raw.WorkType),
		Location:               normalizeLocation(raw.Location),
		SalaryMin:              raw.SalaryMin,
		SalaryMax:              raw.SalaryMax,
		Responsibilities:       make([]*JobResponsibilitySimple, 0, len(raw.Responsibilities)),
		Skills:                 make([]*JobSkillSimple, 0, len(raw.Skills)),
		EducationRequirements:  make([]*JobEducationRequirementSimple, 0, len(raw.EducationRequirements)),
		ExperienceRequirements: make([]*JobExperienceRequirementSimple, 0, len(raw.ExperienceRequirements)),
		IndustryRequirements:   make([]*JobIndustryRequirementSimple, 0, len(raw.IndustryRequirements)),
	}

	// 转换职责
	for _, r := range raw.Responsibilities {
		if strings.TrimSpace(r.Responsibility) == "" {
			continue
		}
		result.Responsibilities = append(result.Responsibilities, &JobResponsibilitySimple{
			Responsibility: strings.TrimSpace(r.Responsibility),
		})
	}

	// 转换技能
	for _, s := range raw.Skills {
		if strings.TrimSpace(s.Skill) == "" {
			continue
		}
		skillType := normalizeSkillType(s.Type)
		result.Skills = append(result.Skills, &JobSkillSimple{
			Skill: strings.TrimSpace(s.Skill),
			Type:  skillType,
		})
	}

	// 转换学历要求
	for _, e := range raw.EducationRequirements {
		if strings.TrimSpace(e.EducationType) == "" {
			continue
		}
		educationType := normalizeEducationType(e.EducationType)
		result.EducationRequirements = append(result.EducationRequirements, &JobEducationRequirementSimple{
			EducationType: educationType,
		})
	}

	// 转换经验要求
	for _, exp := range raw.ExperienceRequirements {
		if strings.TrimSpace(exp.ExperienceType) == "" {
			continue
		}
		experienceType := normalizeExperienceType(exp.ExperienceType)
		result.ExperienceRequirements = append(result.ExperienceRequirements, &JobExperienceRequirementSimple{
			ExperienceType: experienceType,
			MinYears:       exp.MinYears,
			IdealYears:     exp.IdealYears,
		})
	}

	// 转换行业要求
	for _, ind := range raw.IndustryRequirements {
		if strings.TrimSpace(ind.Industry) == "" {
			continue
		}
		result.IndustryRequirements = append(result.IndustryRequirements, &JobIndustryRequirementSimple{
			Industry:    strings.TrimSpace(ind.Industry),
			CompanyName: ind.CompanyName,
		})
	}

	return result, nil
}

// normalizeSkillType 标准化技能类型
func normalizeSkillType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "required", "必需", "必须", "必备":
		return "required"
	case "bonus", "加分", "优先", "preferred":
		return "bonus"
	default:
		return "required" // 默认为必需技能
	}
}

// normalizeEducationType 标准化学历类型
func normalizeEducationType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "unlimited", "不限", "无要求":
		return "unlimited"
	case "junior_college", "大专", "专科":
		return "junior_college"
	case "bachelor", "本科", "学士":
		return "bachelor"
	case "master", "硕士", "研究生":
		return "master"
	case "doctor", "博士", "phd":
		return "doctor"
	default:
		return "bachelor" // 默认为本科
	}
}

// normalizeExperienceType 标准化经验类型
func normalizeExperienceType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "unlimited", "不限", "无要求":
		return "unlimited"
	case "fresh_graduate", "应届", "应届毕业生", "校招":
		return "fresh_graduate"
	case "under_one_year", "1年以下", "一年以下":
		return "under_one_year"
	case "one_to_three_years", "1-3年", "1到3年", "一到三年":
		return "one_to_three_years"
	case "three_to_five_years", "3-5年", "3到5年", "三到五年":
		return "three_to_five_years"
	case "five_to_ten_years", "5-10年", "5到10年", "五到十年":
		return "five_to_ten_years"
	case "over_ten_years", "10年以上", "十年以上":
		return "over_ten_years"
	default:
		return "one_to_three_years" // 默认为1-3年
	}
}

// normalizeWorkType 标准化工作性质
func normalizeWorkType(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*value))
	switch normalized {
	case "full_time", "全职", "正式":
		result := "full_time"
		return &result
	case "part_time", "兼职", "非全职":
		result := "part_time"
		return &result
	case "internship", "实习", "实习生":
		result := "internship"
		return &result
	case "outsourcing", "外包", "外派":
		result := "outsourcing"
		return &result
	default:
		return nil
	}
}

// normalizeLocation 标准化工作地点
func normalizeLocation(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
