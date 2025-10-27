package jobprofilegenerator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofileparser"
)

type rawGenerateResult struct {
	Profile  rawJobProfile         `json:"profile"`
	Sections rawJobProfileSections `json:"sections"`
}

type rawJobProfile struct {
	Name                   string                        `json:"name"`
	WorkType               *string                       `json:"work_type"`
	Location               *string                       `json:"location"`
	SalaryMin              *float64                      `json:"salary_min"`
	SalaryMax              *float64                      `json:"salary_max"`
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

type rawJobProfileSections struct {
	Responsibilities []string `json:"responsibilities"`
	Requirements     []string `json:"requirements"`
	Bonuses          []string `json:"bonuses"`
}

func decodeJobProfileGenerateResult(data []byte) (*JobProfileGenerateResult, error) {
	var raw rawGenerateResult
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	profile := raw.Profile.toResult()

	markdown := buildMarkdownSections(
		normalizeLines(raw.Sections.Responsibilities),
		normalizeLines(raw.Sections.Requirements),
		normalizeLines(raw.Sections.Bonuses),
		profile,
	)

	return &JobProfileGenerateResult{
		Profile:             profile,
		DescriptionMarkdown: markdown,
	}, nil
}

func (raw *rawJobProfile) toResult() *jobprofileparser.JobProfileParseResult {
	result := &jobprofileparser.JobProfileParseResult{
		Name:                   strings.TrimSpace(raw.Name),
		WorkType:               normalizeWorkType(raw.WorkType),
		Location:               normalizePtrString(raw.Location),
		SalaryMin:              raw.SalaryMin,
		SalaryMax:              raw.SalaryMax,
		Responsibilities:       make([]*jobprofileparser.JobResponsibilitySimple, 0, len(raw.Responsibilities)),
		Skills:                 make([]*jobprofileparser.JobSkillSimple, 0, len(raw.Skills)),
		EducationRequirements:  make([]*jobprofileparser.JobEducationRequirementSimple, 0, len(raw.EducationRequirements)),
		ExperienceRequirements: make([]*jobprofileparser.JobExperienceRequirementSimple, 0, len(raw.ExperienceRequirements)),
		IndustryRequirements:   make([]*jobprofileparser.JobIndustryRequirementSimple, 0, len(raw.IndustryRequirements)),
	}

	for _, resp := range raw.Responsibilities {
		if trimmed := strings.TrimSpace(resp.Responsibility); trimmed != "" {
			result.Responsibilities = append(result.Responsibilities, &jobprofileparser.JobResponsibilitySimple{
				Responsibility: trimmed,
			})
		}
	}

	for _, skill := range raw.Skills {
		if trimmed := strings.TrimSpace(skill.Skill); trimmed != "" {
			result.Skills = append(result.Skills, &jobprofileparser.JobSkillSimple{
				Skill: trimmed,
				Type:  normalizeSkillType(skill.Type),
			})
		}
	}

	for _, edu := range raw.EducationRequirements {
		if trimmed := strings.TrimSpace(edu.EducationType); trimmed != "" {
			result.EducationRequirements = append(result.EducationRequirements, &jobprofileparser.JobEducationRequirementSimple{
				EducationType: normalizeEducationType(trimmed),
			})
		}
	}

	for _, exp := range raw.ExperienceRequirements {
		if trimmed := strings.TrimSpace(exp.ExperienceType); trimmed != "" {
			result.ExperienceRequirements = append(result.ExperienceRequirements, &jobprofileparser.JobExperienceRequirementSimple{
				ExperienceType: normalizeExperienceType(trimmed),
				MinYears:       exp.MinYears,
				IdealYears:     exp.IdealYears,
			})
		}
	}

	for _, ind := range raw.IndustryRequirements {
		if trimmed := strings.TrimSpace(ind.Industry); trimmed != "" {
			var company *string
			if ind.CompanyName != nil {
				if name := strings.TrimSpace(*ind.CompanyName); name != "" {
					company = &name
				}
			}
			result.IndustryRequirements = append(result.IndustryRequirements, &jobprofileparser.JobIndustryRequirementSimple{
				Industry:    trimmed,
				CompanyName: company,
			})
		}
	}

	return result
}

func normalizePtrString(value *string) *string {
	if value == nil {
		return nil
	}
	if trimmed := strings.TrimSpace(*value); trimmed != "" {
		return &trimmed
	}
	return nil
}

func normalizeLines(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func buildMarkdownSections(responsibilities, requirements, bonuses []string, profile *jobprofileparser.JobProfileParseResult) string {
	if len(responsibilities) == 0 {
		responsibilities = responsibilitiesFromProfile(profile)
	}
	if len(requirements) == 0 {
		requirements = requirementsFromProfile(profile)
	}
	if len(bonuses) == 0 {
		bonuses = bonusesFromProfile(profile)
	}

	builder := &strings.Builder{}
	writeMarkdownList(builder, "岗位职责", responsibilities)
	writeMarkdownList(builder, "任职要求", requirements)
	writeMarkdownList(builder, "加分项", bonuses)

	return strings.TrimSpace(builder.String())
}

func writeMarkdownList(builder *strings.Builder, title string, items []string) {
	builder.WriteString(title + "：\n")
	if len(items) == 0 {
		builder.WriteString("- 暂无特别要求\n\n")
		return
	}
	for _, item := range items {
		builder.WriteString("- " + item + "\n")
	}
	builder.WriteString("\n")
}

func responsibilitiesFromProfile(profile *jobprofileparser.JobProfileParseResult) []string {
	result := make([]string, 0, len(profile.Responsibilities))
	for _, resp := range profile.Responsibilities {
		if resp != nil && strings.TrimSpace(resp.Responsibility) != "" {
			result = append(result, strings.TrimSpace(resp.Responsibility))
		}
	}
	return result
}

func requirementsFromProfile(profile *jobprofileparser.JobProfileParseResult) []string {
	result := make([]string, 0, len(profile.Skills)+len(profile.EducationRequirements)+len(profile.ExperienceRequirements))
	seen := map[string]struct{}{}

	for _, skill := range profile.Skills {
		if skill == nil {
			continue
		}
		if strings.EqualFold(skill.Type, "required") {
			entry := fmt.Sprintf("熟练掌握%s", strings.TrimSpace(skill.Skill))
			if _, ok := seen[entry]; !ok && strings.TrimSpace(skill.Skill) != "" {
				result = append(result, entry)
				seen[entry] = struct{}{}
			}
		}
	}

	for _, edu := range profile.EducationRequirements {
		if edu == nil {
			continue
		}
		entry := educationDisplay(edu.EducationType)
		if entry != "" {
			if _, ok := seen[entry]; !ok {
				result = append(result, entry)
				seen[entry] = struct{}{}
			}
		}
	}

	for _, exp := range profile.ExperienceRequirements {
		if exp == nil {
			continue
		}
		entry := experienceDisplay(exp)
		if entry != "" {
			if _, ok := seen[entry]; !ok {
				result = append(result, entry)
				seen[entry] = struct{}{}
			}
		}
	}

	return result
}

func bonusesFromProfile(profile *jobprofileparser.JobProfileParseResult) []string {
	result := make([]string, 0, len(profile.Skills)+len(profile.IndustryRequirements))
	seen := map[string]struct{}{}

	for _, skill := range profile.Skills {
		if skill == nil {
			continue
		}
		if strings.EqualFold(skill.Type, "bonus") && strings.TrimSpace(skill.Skill) != "" {
			entry := fmt.Sprintf("具备%s经验更佳", strings.TrimSpace(skill.Skill))
			if _, ok := seen[entry]; !ok {
				result = append(result, entry)
				seen[entry] = struct{}{}
			}
		}
	}

	for _, ind := range profile.IndustryRequirements {
		if ind == nil {
			continue
		}
		entry := ""
		if ind.CompanyName != nil && strings.TrimSpace(*ind.CompanyName) != "" {
			entry = fmt.Sprintf("有%s工作背景优先", strings.TrimSpace(*ind.CompanyName))
		} else if strings.TrimSpace(ind.Industry) != "" {
			entry = fmt.Sprintf("有%s行业经验优先", strings.TrimSpace(ind.Industry))
		}
		if entry != "" {
			if _, ok := seen[entry]; !ok {
				result = append(result, entry)
				seen[entry] = struct{}{}
			}
		}
	}

	return result
}

func educationDisplay(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "unlimited":
		return "学历不限"
	case "junior_college":
		return "大专及以上学历"
	case "bachelor":
		return "本科及以上学历"
	case "master":
		return "硕士及以上学历"
	case "doctor":
		return "博士学历优先"
	default:
		return ""
	}
}

func experienceDisplay(exp *jobprofileparser.JobExperienceRequirementSimple) string {
	if exp == nil {
		return ""
	}
	experienceType := strings.ToLower(strings.TrimSpace(exp.ExperienceType))
	switch experienceType {
	case "unlimited":
		return "经验不限，欢迎积极学习"
	case "fresh_graduate":
		return "欢迎应届毕业生申请"
	case "under_one_year":
		return "具备1年以内相关经验"
	case "one_to_three_years":
		return rangeYears("1-3年相关经验", exp.MinYears, exp.IdealYears)
	case "three_to_five_years":
		return rangeYears("3-5年相关经验", exp.MinYears, exp.IdealYears)
	case "five_to_ten_years":
		return rangeYears("5-10年相关经验", exp.MinYears, exp.IdealYears)
	case "over_ten_years":
		return "具备10年以上丰富经验"
	default:
		if exp.MinYears > 0 {
			return fmt.Sprintf("具备%d年以上相关经验", exp.MinYears)
		}
	}
	return ""
}

func rangeYears(defaultText string, minYears, idealYears int) string {
	if minYears <= 0 && idealYears <= 0 {
		return defaultText
	}
	if minYears > 0 && idealYears > 0 {
		if minYears == idealYears {
			return fmt.Sprintf("具备%d年相关经验", minYears)
		}
		return fmt.Sprintf("具备%d-%d年相关经验", minYears, idealYears)
	}
	if minYears > 0 {
		return fmt.Sprintf("具备%d年以上相关经验", minYears)
	}
	return defaultText
}

func normalizeSkillType(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "required", "must", "必需", "必须", "必备":
		return "required"
	case "bonus", "加分", "优先", "preferred":
		return "bonus"
	default:
		return "required"
	}
}

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
		return "bachelor"
	}
}

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
		return "one_to_three_years"
	}
}

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
