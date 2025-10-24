package resumeparser

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
)

// rawResumeParseResult 用于承接模型返回的原始JSON，方便进行二次清洗
type rawResumeParseResult struct {
	BasicInfo   *rawParsedBasicInfo   `json:"basic_info"`
	Educations  []rawParsedEducation  `json:"educations"`
	Experiences []rawParsedExperience `json:"experiences"`
	Skills      []rawParsedSkill      `json:"skills"`
	Projects    []rawParsedProject    `json:"projects"`
}

type rawParsedBasicInfo struct {
	Name             string   `json:"name"`
	Phone            string   `json:"phone"`
	Email            string   `json:"email"`
	Gender           string   `json:"gender"`
	Birthday         *string  `json:"birthday"`
	CurrentCity      string   `json:"current_city"`
	HighestEducation string   `json:"highest_education"`
	YearsExperience  *float64 `json:"years_experience"`
}

type rawParsedEducation struct {
	School         string  `json:"school"`
	Major          string  `json:"major"`
	Degree         string  `json:"degree"`
	UniversityType string  `json:"university_type"`
	StartDate      *string `json:"start_date"`
	EndDate        *string `json:"end_date"`
	GPA            string  `json:"gpa"`
}

type rawParsedExperience struct {
	Company      string  `json:"company"`
	Position     string  `json:"position"`
	StartDate    *string `json:"start_date"`
	EndDate      *string `json:"end_date"`
	Description  string  `json:"description"`
	Achievements string  `json:"achievements"`
}

type rawParsedSkill struct {
	Name        string `json:"name"`
	Level       string `json:"level"`
	Description string `json:"description"`
}

type rawParsedProject struct {
	Name             string  `json:"name"`
	Role             string  `json:"role"`
	Company          string  `json:"company"`
	Description      string  `json:"description"`
	Responsibilities string  `json:"responsibilities"`
	Achievements     string  `json:"achievements"`
	Technologies     string  `json:"technologies"`
	ProjectURL       string  `json:"project_url"`
	ProjectType      string  `json:"project_type"`
	StartDate        *string `json:"start_date"`
	EndDate          *string `json:"end_date"`
}

func decodeResumeParseResult(data []byte) (*ResumeParseResult, error) {
	var raw rawResumeParseResult
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	result := &ResumeParseResult{}

	if raw.BasicInfo != nil {
		basic := &domain.ParsedBasicInfo{
			Name:             strings.TrimSpace(raw.BasicInfo.Name),
			Phone:            strings.TrimSpace(raw.BasicInfo.Phone),
			Email:            strings.TrimSpace(raw.BasicInfo.Email),
			Gender:           strings.TrimSpace(raw.BasicInfo.Gender),
			CurrentCity:      strings.TrimSpace(raw.BasicInfo.CurrentCity),
			HighestEducation: strings.TrimSpace(raw.BasicInfo.HighestEducation),
		}
		if raw.BasicInfo.Birthday != nil {
			basic.Birthday = parseFlexibleTime(raw.BasicInfo.Birthday)
		}
		if raw.BasicInfo.YearsExperience != nil {
			basic.YearsExperience = *raw.BasicInfo.YearsExperience
		}
		result.BasicInfo = basic
	}

	if len(raw.Educations) > 0 {
		result.Educations = make([]*domain.ParsedEducation, 0, len(raw.Educations))
		for _, item := range raw.Educations {
			edu := &domain.ParsedEducation{
				School:    strings.TrimSpace(item.School),
				Major:     strings.TrimSpace(item.Major),
				Degree:    strings.TrimSpace(item.Degree),
				GPA:       strings.TrimSpace(item.GPA),
				StartDate: parseFlexibleTime(item.StartDate),
				EndDate:   parseFlexibleTime(item.EndDate),
			}
			if val := strings.TrimSpace(item.UniversityType); val != "" {
				edu.UniversityType = normalizeUniversityType(val)
			}
			result.Educations = append(result.Educations, edu)
		}
	}

	if len(raw.Experiences) > 0 {
		result.Experiences = make([]*domain.ParsedExperience, 0, len(raw.Experiences))
		for _, item := range raw.Experiences {
			exp := &domain.ParsedExperience{
				Company:      strings.TrimSpace(item.Company),
				Position:     strings.TrimSpace(item.Position),
				Description:  strings.TrimSpace(item.Description),
				Achievements: strings.TrimSpace(item.Achievements),
				StartDate:    parseFlexibleTime(item.StartDate),
				EndDate:      parseFlexibleTime(item.EndDate),
			}
			result.Experiences = append(result.Experiences, exp)
		}
	}

	if len(raw.Skills) > 0 {
		result.Skills = make([]*domain.ParsedSkill, 0, len(raw.Skills))
		for _, item := range raw.Skills {
			skill := &domain.ParsedSkill{
				Name:        strings.TrimSpace(item.Name),
				Level:       strings.TrimSpace(item.Level),
				Description: strings.TrimSpace(item.Description),
			}
			result.Skills = append(result.Skills, skill)
		}
	}

	if len(raw.Projects) > 0 {
		result.Projects = make([]*domain.ParsedProject, 0, len(raw.Projects))
		for _, item := range raw.Projects {
			project := &domain.ParsedProject{
				Name:             strings.TrimSpace(item.Name),
				Role:             strings.TrimSpace(item.Role),
				Company:          strings.TrimSpace(item.Company),
				Description:      strings.TrimSpace(item.Description),
				Responsibilities: strings.TrimSpace(item.Responsibilities),
				Achievements:     strings.TrimSpace(item.Achievements),
				Technologies:     strings.TrimSpace(item.Technologies),
				ProjectURL:       strings.TrimSpace(item.ProjectURL),
				ProjectType:      strings.TrimSpace(item.ProjectType),
				StartDate:        parseFlexibleTime(item.StartDate),
				EndDate:          parseFlexibleTime(item.EndDate),
			}
			result.Projects = append(result.Projects, project)
		}
	}

	return result, nil
}

// parseFlexibleTime 尝试解析多种日期格式，无法解析时返回nil以避免整个流程失败
func parseFlexibleTime(raw *string) *time.Time {
	if raw == nil {
		return nil
	}
	s := strings.TrimSpace(*raw)
	if s == "" || strings.EqualFold(s, "null") {
		return nil
	}

	normalized := normalizeDateString(s)

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
		"2006-1-2",
		"2006/01/02",
		"2006/1/2",
		"2006.01.02",
		"2006.1.2",
		"2006-01",
		"2006-1",
		"2006/01",
		"2006/1",
		"2006.01",
		"2006.1",
		"2006",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, normalized); err == nil {
			tt := t.UTC()
			return &tt
		}
	}

	if !strings.Contains(normalized, "T") {
		// 尝试补齐时间部分再解析
		if t, err := time.Parse(time.RFC3339, normalized+"T00:00:00Z"); err == nil {
			tt := t.UTC()
			return &tt
		}
	}

	return nil
}

func normalizeDateString(s string) string {
	replacer := strings.NewReplacer(
		"年", "-",
		"月", "-",
		"日", "",
		"/", "-",
		".", "-",
	)
	normalized := replacer.Replace(strings.TrimSpace(s))
	return strings.Trim(normalized, "-")
}

func normalizeUniversityType(value string) consts.UniversityType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(consts.UniversityType211):
		return consts.UniversityType211
	case string(consts.UniversityType985):
		return consts.UniversityType985
	case string(consts.UniversityTypeOrdinary):
		return consts.UniversityTypeOrdinary
	default:
		return consts.UniversityType(value)
	}
}
