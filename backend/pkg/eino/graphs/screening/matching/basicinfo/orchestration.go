package basicinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// BasicInfoAgent 基本信息匹配Agent
type BasicInfoAgent struct {
	chain *compose.Chain[*domain.BasicInfoData, *domain.BasicMatchDetail]
}

// 输入处理，将BasicInfoData转换为模板变量
func newInputLambda(ctx context.Context, input *domain.BasicInfoData, opts ...any) (map[string]any, error) {
	// 验证输入数据
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}
	// 仅保留与基础匹配相关的字段，避免上下文过长
	inputData := map[string]any{
		"job_profile": buildJobBasicInfo(input.JobProfile),
		"resume":      buildResumeBasicInfo(input.Resume),
	}

	// 将输入转换为JSON字符串
	inputJSON, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	return map[string]any{
		"input": string(inputJSON),
	}, nil
}

// 输出处理，将模型输出解析为结构化结果
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*domain.BasicMatchDetail, error) {

	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	var output domain.BasicMatchDetail
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	return &output, nil
}

// NewBasicInfoAgent 创建基本信息匹配Agent
func NewBasicInfoAgent(ctx context.Context, llm model.ToolCallingChatModel) (*BasicInfoAgent, error) {
	// 创建聊天模板
	chatTemplate, err := NewBasicInfoChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[*domain.BasicInfoData, *domain.BasicMatchDetail]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model"), compose.WithNodeKey("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &BasicInfoAgent{
		chain: chain,
	}, nil
}

// GetChain 获取处理链
func (a *BasicInfoAgent) GetChain() *compose.Chain[*domain.BasicInfoData, *domain.BasicMatchDetail] {
	return a.chain
}

// Compile 编译链为可执行的Runnable
func (a *BasicInfoAgent) Compile(ctx context.Context) (compose.Runnable[*domain.BasicInfoData, *domain.BasicMatchDetail], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return runnable, nil
}

// Process 处理基本信息匹配
func (a *BasicInfoAgent) Process(ctx context.Context, input *domain.BasicInfoData) (*domain.BasicMatchDetail, error) {
	// 验证输入
	if err := a.validateInput(input); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// 编译并执行处理链
	runnable, err := a.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}

	output, err := runnable.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to process: %w", err)
	}

	// 验证输出
	if err := a.validateOutput(output); err != nil {
		return nil, fmt.Errorf("invalid output: %w", err)
	}

	return output, nil
}

// validateInput 验证输入数据
func (a *BasicInfoAgent) validateInput(input *domain.BasicInfoData) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	if input.JobProfile == nil {
		return fmt.Errorf("job profile cannot be nil")
	}
	if input.Resume == nil {
		return fmt.Errorf("resume cannot be nil")
	}
	return nil
}

// validateOutput 验证输出数据
func (a *BasicInfoAgent) validateOutput(output *domain.BasicMatchDetail) error {
	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}
	if output.Score < 0 || output.Score > 100 {
		return fmt.Errorf("score must be between 0 and 100, got %f", output.Score)
	}

	return nil
}

// GetAgentType 返回Agent类型
func (a *BasicInfoAgent) GetAgentType() string {
	return "BasicInfoAgent"
}

// GetVersion 返回Agent版本
func (a *BasicInfoAgent) GetVersion() string {
	return "1.1.0"
}

type jobBasicInfo struct {
	Name        string       `json:"name,omitempty"`
	Department  string       `json:"department,omitempty"`
	WorkType    *string      `json:"work_type,omitempty"`
	Location    *string      `json:"location,omitempty"`
	SalaryRange *salaryRange `json:"salary_range,omitempty"`
	Notes       []string     `json:"notes,omitempty"`
}

type resumeBasicInfo struct {
	Name              string              `json:"name,omitempty"`
	Age               *int                `json:"age,omitempty"`
	CurrentCity       string              `json:"current_city,omitempty"`
	ExpectedCity      string              `json:"expected_city,omitempty"`
	YearsExperience   *float64            `json:"years_experience,omitempty"`
	ExpectedSalary    *salaryRange        `json:"expected_salary,omitempty"`
	ExpectedSalaryRaw string              `json:"expected_salary_raw,omitempty"`
	EmploymentStatus  string              `json:"employment_status,omitempty"`
	PersonalSummary   string              `json:"personal_summary,omitempty"`
	Honors            []string            `json:"honors,omitempty"`
	OtherInfo         string              `json:"other_info,omitempty"`
	RecentExperiences []experienceSummary `json:"recent_experiences,omitempty"`
	Notes             []string            `json:"notes,omitempty"`
}

type salaryRange struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

type experienceSummary struct {
	Company        string `json:"company,omitempty"`
	Position       string `json:"position,omitempty"`
	ExperienceType string `json:"experience_type,omitempty"`
	Title          string `json:"title,omitempty"`
	Description    string `json:"description,omitempty"`
	Start          string `json:"start,omitempty"`
	End            string `json:"end,omitempty"`
}

func buildJobBasicInfo(detail *domain.JobProfileDetail) jobBasicInfo {
	info := jobBasicInfo{}
	if detail == nil || detail.JobProfile == nil {
		return info
	}

	base := detail.JobProfile
	info.Name = base.Name
	info.Department = base.Department
	info.WorkType = base.WorkType
	info.Location = base.Location

	if base.SalaryMin != nil || base.SalaryMax != nil {
		info.SalaryRange = &salaryRange{
			Min: base.SalaryMin,
			Max: base.SalaryMax,
		}
	}

	if base.Location == nil || strings.TrimSpace(*base.Location) == "" {
		info.Notes = append(info.Notes, "岗位未提供明确的工作地点信息")
	}
	if base.SalaryMin == nil && base.SalaryMax == nil {
		info.Notes = append(info.Notes, "岗位未提供薪资预算范围")
	}

	return info
}

func buildResumeBasicInfo(detail *domain.ResumeDetail) resumeBasicInfo {
	info := resumeBasicInfo{}
	if detail == nil || detail.Resume == nil {
		return info
	}

	base := detail.Resume
	info.Name = base.Name
	info.CurrentCity = base.CurrentCity
	info.ExpectedCity = base.ExpectedCity
	info.PersonalSummary = strings.TrimSpace(base.PersonalSummary)
	info.OtherInfo = strings.TrimSpace(base.OtherInfo)
	if base.Age != nil && *base.Age > 0 {
		info.Age = base.Age
	}

	if base.EmploymentStatus != nil {
		info.EmploymentStatus = string(*base.EmploymentStatus)
	}

	if base.YearsExperience > 0 {
		info.YearsExperience = float64Ptr(base.YearsExperience)
	}

	if expectedRange, raw := extractExpectedSalary(detail); expectedRange != nil {
		info.ExpectedSalary = expectedRange
		info.ExpectedSalaryRaw = raw
	} else if raw != "" {
		info.ExpectedSalaryRaw = raw
	}

	info.RecentExperiences = summarizeExperiences(detail.Experiences, 3)

	if strings.TrimSpace(base.CurrentCity) == "" {
		info.Notes = append(info.Notes, "简历未提供当前城市信息")
	}
	if strings.TrimSpace(base.ExpectedCity) == "" {
		info.Notes = append(info.Notes, "简历未提供明确的期望城市")
	}
	switch {
	case info.ExpectedSalary != nil:
		// no-op
	case strings.TrimSpace(base.ExpectedSalary) == "":
		info.Notes = append(info.Notes, "简历未提供明确的期望薪资信息")
	default:
		info.Notes = append(info.Notes, "期望薪资格式暂未解析，请人工确认")
	}
	if info.EmploymentStatus == "" {
		info.Notes = append(info.Notes, "简历未提供当前工作状态")
	}

	if strings.TrimSpace(base.HonorsCertificates) != "" {
		info.Honors = splitAndTrim(base.HonorsCertificates)
	}

	return info
}

func summarizeExperiences(exps []*domain.ResumeExperience, limit int) []experienceSummary {
	if len(exps) == 0 {
		return nil
	}

	sorted := make([]*domain.ResumeExperience, 0, len(exps))
	sorted = append(sorted, exps...)

	sort.Slice(sorted, func(i, j int) bool {
		typeWeight := func(exp *domain.ResumeExperience) int {
			if exp == nil {
				return 1
			}
			if exp.ExperienceType == consts.ExperienceTypeWork {
				return 0
			}
			return 1
		}
		if wI, wJ := typeWeight(sorted[i]), typeWeight(sorted[j]); wI != wJ {
			return wI < wJ
		}
		startI := comparableTime(sorted[i])
		startJ := comparableTime(sorted[j])
		if startI.Equal(startJ) {
			return endTime(sorted[i]).After(endTime(sorted[j]))
		}
		return startI.After(startJ)
	})

	if limit <= 0 || limit > len(sorted) {
		limit = len(sorted)
	}

	result := make([]experienceSummary, 0, limit)
	for _, exp := range sorted[:limit] {
		if exp == nil {
			continue
		}
		item := experienceSummary{
			Company:        strings.TrimSpace(exp.Company),
			Position:       strings.TrimSpace(exp.Position),
			ExperienceType: string(exp.ExperienceType),
			Title:          strings.TrimSpace(exp.Title),
		}
		if desc := strings.TrimSpace(exp.Description); desc != "" {
			item.Description = truncateText(desc, 120)
		}
		if exp.StartDate != nil && !exp.StartDate.IsZero() {
			item.Start = exp.StartDate.Format("2006-01")
		}
		if exp.EndDate != nil && !exp.EndDate.IsZero() {
			item.End = exp.EndDate.Format("2006-01")
		} else {
			item.End = "至今"
		}
		result = append(result, item)
	}
	return result
}

func extractExpectedSalary(detail *domain.ResumeDetail) (*salaryRange, string) {
	if detail == nil {
		return nil, ""
	}
	if detail.Resume == nil {
		return nil, ""
	}
	raw := strings.TrimSpace(detail.Resume.ExpectedSalary)
	if raw == "" {
		return nil, ""
	}
	if rng := parseExpectedSalaryRange(raw); rng != nil {
		return rng, raw
	}
	return nil, raw
}

func comparableTime(exp *domain.ResumeExperience) time.Time {
	if exp == nil {
		return time.Time{}
	}
	if exp.StartDate != nil && !exp.StartDate.IsZero() {
		return *exp.StartDate
	}
	if exp.EndDate != nil && !exp.EndDate.IsZero() {
		return *exp.EndDate
	}
	return time.Time{}
}

func endTime(exp *domain.ResumeExperience) time.Time {
	if exp == nil {
		return time.Time{}
	}
	if exp.EndDate != nil && !exp.EndDate.IsZero() {
		return *exp.EndDate
	}
	return comparableTime(exp)
}

func truncateText(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit]) + "..."
}

func float64Ptr(v float64) *float64 {
	return &v
}

var salaryNumberPattern = regexp.MustCompile(`\d+(?:\.\d+)?`)

func parseExpectedSalaryRange(value string) *salaryRange {
	if value = strings.TrimSpace(value); value == "" {
		return nil
	}
	normalized := strings.ToLower(value)
	multiplier := 1.0
	switch {
	case strings.Contains(normalized, "k"):
		multiplier = 1000
	case strings.Contains(normalized, "千"):
		multiplier = 1000
	case strings.Contains(normalized, "万"):
		multiplier = 10000
	}

	matches := salaryNumberPattern.FindAllString(normalized, -1)
	if len(matches) == 0 {
		return nil
	}

	parse := func(s string) (float64, bool) {
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return num * multiplier, true
	}

	var minVal, maxVal float64
	var ok bool
	if minVal, ok = parse(matches[0]); !ok {
		return nil
	}
	if len(matches) > 1 {
		if maxVal, ok = parse(matches[1]); !ok {
			return nil
		}
	} else {
		maxVal = minVal
		if strings.Contains(normalized, "上") || strings.Contains(normalized, "以上") {
			maxVal = minVal
		}
	}

	if maxVal < minVal {
		maxVal, minVal = minVal, maxVal
	}

	return &salaryRange{
		Min: float64Ptr(minVal),
		Max: float64Ptr(maxVal),
	}
}

func splitAndTrim(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool {
		switch r {
		case '\n', ',', '，', ';', '；':
			return true
		default:
			return false
		}
	})
	if len(parts) == 0 {
		return nil
	}
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
