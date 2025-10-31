package weights

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// WeightPlannerAgent 岗位权重规划 Agent
type WeightPlannerAgent struct {
	chain *compose.Chain[*domain.WeightInferenceInput, *domain.WeightInferenceResult]
}

// llmWeightSchemeOutput LLM 输出的单个权重方案
type llmWeightSchemeOutput struct {
	Skill          float64  `json:"skill"`
	Responsibility float64  `json:"responsibility"`
	Experience     float64  `json:"experience"`
	Education      float64  `json:"education"`
	Industry       float64  `json:"industry"`
	Basic          float64  `json:"basic"`
	Rationale      []string `json:"rationale"`
}

// llmWeightOutput LLM 输出的完整权重结构
type llmWeightOutput struct {
	Schemes []llmWeightSchemeOutput `json:"schemes"`
}

// newInputLambda 将权重推理输入转换为模板变量
func newInputLambda(ctx context.Context, input *domain.WeightInferenceInput, opts ...any) (map[string]any, error) {
	if input == nil {
		return nil, fmt.Errorf("权重推理输入不能为空")
	}
	if input.JobProfile == nil || input.JobProfile.JobProfile == nil {
		return nil, fmt.Errorf("岗位信息不能为空")
	}

	// 将 JobProfile 转换为易理解的文档格式
	jobProfileDoc := formatJobProfileAsDocument(input.JobProfile)

	return map[string]any{
		"job_profile_json": jobProfileDoc,
	}, nil
}

// formatJobProfileAsDocument 将岗位画像转换为易理解的文档格式
func formatJobProfileAsDocument(jobProfile *domain.JobProfileDetail) string {
	var doc strings.Builder

	// 基本信息
	base := jobProfile.JobProfile
	doc.WriteString("# 岗位信息\n\n")

	doc.WriteString("## 基本信息\n")
	doc.WriteString(fmt.Sprintf("- **岗位名称**: %s\n", base.Name))
	if base.Department != "" {
		doc.WriteString(fmt.Sprintf("- **所属部门**: %s\n", base.Department))
	}
	if base.WorkType != nil {
		doc.WriteString(fmt.Sprintf("- **工作性质**: %s\n", translateWorkType(*base.WorkType)))
	}
	if base.Location != nil && *base.Location != "" {
		doc.WriteString(fmt.Sprintf("- **工作地点**: %s\n", *base.Location))
	}
	if base.SalaryMin != nil && base.SalaryMax != nil {
		doc.WriteString(fmt.Sprintf("- **薪资范围**: %.0f - %.0f 元/月\n", *base.SalaryMin, *base.SalaryMax))
	} else if base.SalaryMin != nil {
		doc.WriteString(fmt.Sprintf("- **最低薪资**: %.0f 元/月\n", *base.SalaryMin))
	} else if base.SalaryMax != nil {
		doc.WriteString(fmt.Sprintf("- **最高薪资**: %.0f 元/月\n", *base.SalaryMax))
	}
	if base.Description != nil && *base.Description != "" {
		doc.WriteString(fmt.Sprintf("- **岗位描述**: %s\n", *base.Description))
	}
	doc.WriteString("\n")

	// 岗位职责
	if len(jobProfile.Responsibilities) > 0 {
		doc.WriteString("## 岗位职责\n")
		for i, resp := range jobProfile.Responsibilities {
			doc.WriteString(fmt.Sprintf("%d. %s\n", i+1, resp.Responsibility))
		}
		doc.WriteString("\n")
	}

	// 技能要求
	if len(jobProfile.Skills) > 0 {
		doc.WriteString("## 技能要求\n")

		// 必需技能
		requiredSkills := []string{}
		bonusSkills := []string{}
		for _, skill := range jobProfile.Skills {
			if skill.Type == string(consts.JobSkillTypeRequired) {
				requiredSkills = append(requiredSkills, skill.Skill)
			} else if skill.Type == string(consts.JobSkillTypeBonus) {
				bonusSkills = append(bonusSkills, skill.Skill)
			}
		}

		if len(requiredSkills) > 0 {
			doc.WriteString("### 必需技能\n")
			for i, skill := range requiredSkills {
				doc.WriteString(fmt.Sprintf("%d. %s\n", i+1, skill))
			}
		}

		if len(bonusSkills) > 0 {
			doc.WriteString("### 加分技能\n")
			for i, skill := range bonusSkills {
				doc.WriteString(fmt.Sprintf("%d. %s\n", i+1, skill))
			}
		}
		doc.WriteString("\n")
	}

	// 经验要求
	if len(jobProfile.ExperienceRequirements) > 0 {
		doc.WriteString("## 工作经验要求\n")
		for i, exp := range jobProfile.ExperienceRequirements {
			expText := translateExperienceType(exp.ExperienceType)
			if exp.MinYears > 0 || exp.IdealYears > 0 {
				if exp.MinYears > 0 && exp.IdealYears > 0 {
					doc.WriteString(fmt.Sprintf("%d. %s（最少年限：%d年，理想年限：%d年）\n", i+1, expText, exp.MinYears, exp.IdealYears))
				} else if exp.MinYears > 0 {
					doc.WriteString(fmt.Sprintf("%d. %s（最少年限：%d年）\n", i+1, expText, exp.MinYears))
				} else if exp.IdealYears > 0 {
					doc.WriteString(fmt.Sprintf("%d. %s（理想年限：%d年）\n", i+1, expText, exp.IdealYears))
				} else {
					doc.WriteString(fmt.Sprintf("%d. %s\n", i+1, expText))
				}
			} else {
				doc.WriteString(fmt.Sprintf("%d. %s\n", i+1, expText))
			}
		}
		doc.WriteString("\n")
	}

	// 学历要求
	if len(jobProfile.EducationRequirements) > 0 {
		doc.WriteString("## 学历要求\n")
		for i, edu := range jobProfile.EducationRequirements {
			doc.WriteString(fmt.Sprintf("%d. %s\n", i+1, translateEducationType(edu.EducationType)))
		}
		doc.WriteString("\n")
	}

	// 行业要求
	if len(jobProfile.IndustryRequirements) > 0 {
		doc.WriteString("## 行业背景要求\n")
		for i, industry := range jobProfile.IndustryRequirements {
			if industry.CompanyName != nil && *industry.CompanyName != "" {
				doc.WriteString(fmt.Sprintf("%d. %s（优先考虑有 %s 工作经验的候选人）\n", i+1, industry.Industry, *industry.CompanyName))
			} else {
				doc.WriteString(fmt.Sprintf("%d. %s\n", i+1, industry.Industry))
			}
		}
		doc.WriteString("\n")
	}

	return doc.String()
}

// translateWorkType 翻译工作性质
func translateWorkType(workType string) string {
	switch consts.JobWorkType(workType) {
	case consts.JobWorkTypeFullTime:
		return "全职"
	case consts.JobWorkTypePartTime:
		return "兼职"
	case consts.JobWorkTypeInternship:
		return "实习"
	case consts.JobWorkTypeOutsourcing:
		return "外包"
	default:
		return workType
	}
}

// translateExperienceType 翻译经验类型
func translateExperienceType(expType string) string {
	switch consts.JobExperienceType(expType) {
	case consts.JobExperienceTypeUnlimited:
		return "不限"
	case consts.JobExperienceTypeFreshGraduate:
		return "应届毕业生"
	case consts.JobExperienceTypeUnderOneYear:
		return "1年以下"
	case consts.JobExperienceTypeOneToThree:
		return "1-3年"
	case consts.JobExperienceTypeThreeToFive:
		return "3-5年"
	case consts.JobExperienceTypeFiveToTen:
		return "5-10年"
	case consts.JobExperienceTypeOverTen:
		return "10年以上"
	default:
		return expType
	}
}

// translateEducationType 翻译学历类型
func translateEducationType(eduType string) string {
	switch consts.JobEducationType(eduType) {
	case consts.JobEducationTypeUnlimited:
		return "不限"
	case consts.JobEducationTypeJunior:
		return "大专"
	case consts.JobEducationTypeBachelor:
		return "本科"
	case consts.JobEducationTypeMaster:
		return "硕士"
	case consts.JobEducationTypeDoctor:
		return "博士"
	default:
		return eduType
	}
}

// newOutputLambda 解析模型输出为结构化权重
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*domain.WeightInferenceResult, error) {
	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("模型输出为空")
	}

	var output llmWeightOutput
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("解析模型输出失败: %w; raw=%s", err, msg.Content)
	}

	// 验证 schemes 数组长度
	if len(output.Schemes) != 3 {
		return nil, fmt.Errorf("权重方案数量不正确，期望3个方案，实际得到%d个; raw=%s", len(output.Schemes), msg.Content)
	}

	// 定义三种方案的类型标签
	schemeTypes := []string{
		domain.WeightSchemeTypeDefault,        // 第一个方案：注重岗位职责 + 技能匹配
		domain.WeightSchemeTypeFreshGraduate,  // 第二个方案：注重教育经历 + 技能匹配
		domain.WeightSchemeTypeExperienced,    // 第三个方案：注重工作经验 + 岗位职责匹配
	}

	var weightSchemes []domain.WeightScheme
	for i, schemeOutput := range output.Schemes {
		weights := domain.DimensionWeights{
			Skill:          schemeOutput.Skill,
			Responsibility: schemeOutput.Responsibility,
			Experience:     schemeOutput.Experience,
			Education:      schemeOutput.Education,
			Industry:       schemeOutput.Industry,
			Basic:          schemeOutput.Basic,
		}

		// 验证权重
		if err := validateWeights(weights); err != nil {
			return nil, fmt.Errorf("方案 %d (%s) 权重验证失败: %w", i+1, schemeTypes[i], err)
		}

		weightScheme := domain.WeightScheme{
			Type:      schemeTypes[i],
			Weights:   weights,
			Rationale: schemeOutput.Rationale,
		}
		weightSchemes = append(weightSchemes, weightScheme)
	}

	result := &domain.WeightInferenceResult{
		WeightSchemes: weightSchemes,
	}
	return result, nil
}

// validateWeights 校验权重取值
func validateWeights(weights domain.DimensionWeights) error {
	values := []struct {
		name  string
		value float64
	}{
		{"skill", weights.Skill},
		{"responsibility", weights.Responsibility},
		{"experience", weights.Experience},
		{"education", weights.Education},
		{"industry", weights.Industry},
		{"basic", weights.Basic},
	}

	var sum float64
	for _, item := range values {
		if item.value < 0 {
			return fmt.Errorf("维度 %s 权重不能为负: %.4f", item.name, item.value)
		}
		if item.value > 1 {
			return fmt.Errorf("维度 %s 权重超过1: %.4f", item.name, item.value)
		}
		sum += item.value
	}

	if sum <= 0 {
		return fmt.Errorf("权重总和必须大于0")
	}
	if diff := sum - 1; diff > 0.05 || diff < -0.05 {
		return fmt.Errorf("权重总和需接近1，当前为 %.4f", sum)
	}

	return nil
}

// NewWeightPlannerAgent 创建岗位权重规划 Agent
func NewWeightPlannerAgent(ctx context.Context, llm model.ToolCallingChatModel) (*WeightPlannerAgent, error) {
	chatTemplate, err := NewWeightPlannerChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建权重规划模板失败: %w", err)
	}

	chain := compose.NewChain[*domain.WeightInferenceInput, *domain.WeightInferenceResult]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model"), compose.WithNodeKey("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &WeightPlannerAgent{chain: chain}, nil
}

// GetChain 返回 Agent 处理链
func (a *WeightPlannerAgent) GetChain() *compose.Chain[*domain.WeightInferenceInput, *domain.WeightInferenceResult] {
	return a.chain
}

// Compile 编译链为 Runnable
func (a *WeightPlannerAgent) Compile(ctx context.Context) (compose.Runnable[*domain.WeightInferenceInput, *domain.WeightInferenceResult], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译权重规划链失败: %w", err)
	}
	return runnable, nil
}

// GetVersion 返回 Agent 版本
func (a *WeightPlannerAgent) GetVersion() string {
	return "1.1.0"
}
