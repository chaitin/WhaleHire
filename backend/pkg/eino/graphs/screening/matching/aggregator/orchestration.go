package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/chaitin/WhaleHire/backend/consts"
	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// AggregatorAgent 匹配结果聚合Agent
type AggregatorAgent struct {
	chain *compose.Chain[map[string]any, *domain.JobResumeMatch]
}

type aggregatorState struct {
	Input     *AggregatorInput
	Overall   float64
	MatchedAt time.Time
}

// AggregatorInput 聚合器输入结构
type AggregatorInput struct {
	BasicMatch          *domain.BasicMatchDetail          `json:"basic_match"`
	SkillMatch          *domain.SkillMatchDetail          `json:"skill_match"`
	ResponsibilityMatch *domain.ResponsibilityMatchDetail `json:"responsibility_match"`
	ExperienceMatch     *domain.ExperienceMatchDetail     `json:"experience_match"`
	EducationMatch      *domain.EducationMatchDetail      `json:"education_match"`
	IndustryMatch       *domain.IndustryMatchDetail       `json:"industry_match"`
	Weights             *domain.DimensionWeights          `json:"weights"`
	TaskMetaData        *domain.TaskMetaData              `json:"-"` // 任务元数据不参与序列化
}

type llmDimensionSummary struct {
	Key        string   `json:"key"`
	Name       string   `json:"name"`
	Score      float64  `json:"score"`
	Highlights []string `json:"highlights,omitempty"`
	Risks      []string `json:"risks,omitempty"`
}

type llmAggregatorInput struct {
	OverallScore       float64                 `json:"overall_score"`
	DimensionWeights   domain.DimensionWeights `json:"dimension_weights"`
	DimensionScores    map[string]float64      `json:"dimension_scores"`
	DimensionSummaries []llmDimensionSummary   `json:"dimension_summaries"`
	Strengths          []string                `json:"strengths,omitempty"`
	Risks              []string                `json:"risks,omitempty"`
	MissingInfo        []string                `json:"missing_information,omitempty"`
}

type llmAggregatorOutput struct {
	OverallScore    *float64 `json:"overall_score,omitempty"`
	Recommendations []string `json:"recommendations"`
}

// 输入处理，将map[string]any转换为模板变量
func newInputLambda(ctx context.Context, input map[string]any, opts ...any) (map[string]any, error) {
	// 构建聚合输入结构
	aggregatorInput := &AggregatorInput{}

	// 从map中提取各个Agent的输出结果
	if basicMatch, ok := input[domain.BasicInfoAgent]; ok {
		if basicDetail, ok := basicMatch.(*domain.BasicMatchDetail); ok {
			aggregatorInput.BasicMatch = basicDetail
		}
	}

	if skillMatch, ok := input[domain.SkillAgent]; ok {
		if skillDetail, ok := skillMatch.(*domain.SkillMatchDetail); ok {
			aggregatorInput.SkillMatch = skillDetail
		}
	}

	if responsibilityMatch, ok := input[domain.ResponsibilityAgent]; ok {
		if respDetail, ok := responsibilityMatch.(*domain.ResponsibilityMatchDetail); ok {
			aggregatorInput.ResponsibilityMatch = respDetail
		}
	}

	if experienceMatch, ok := input[domain.ExperienceAgent]; ok {
		if expDetail, ok := experienceMatch.(*domain.ExperienceMatchDetail); ok {
			aggregatorInput.ExperienceMatch = expDetail
		}
	}

	if educationMatch, ok := input[domain.EducationAgent]; ok {
		if eduDetail, ok := educationMatch.(*domain.EducationMatchDetail); ok {
			aggregatorInput.EducationMatch = eduDetail
		}
	}

	if industryMatch, ok := input[domain.IndustryAgent]; ok {
		if indDetail, ok := industryMatch.(*domain.IndustryMatchDetail); ok {
			aggregatorInput.IndustryMatch = indDetail
		}
	}

	// 提取任务元数据
	if taskMetaData, ok := input[domain.TaskMetaDataNode]; ok {
		if metaData, ok := taskMetaData.(*domain.TaskMetaData); ok {
			aggregatorInput.TaskMetaData = metaData
			// 优先使用 TaskMetaData 中的权重（来自请求）
			if metaData.DimensionWeights != nil {
				aggregatorInput.Weights = metaData.DimensionWeights
			}
		}
	}

	weights := aggregatorInput.Weights
	if weights == nil {
		defaultWeights := domain.DefaultDimensionWeights
		weights = &defaultWeights
		aggregatorInput.Weights = weights
	}

	dimensionScores := collectDimensionScores(aggregatorInput)
	overallScore := calculateOverallScore(weights, dimensionScores)
	matchedAt := time.Now()

	llmInput := buildLLMInput(aggregatorInput, dimensionScores, overallScore)
	inputJSON, err := json.Marshal(llmInput)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal llm input: %w", err)
	}

	if err := compose.ProcessState[*aggregatorState](ctx, func(_ context.Context, state *aggregatorState) error {
		state.Input = aggregatorInput
		state.Overall = overallScore
		state.MatchedAt = matchedAt
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to store aggregator state: %w", err)
	}

	return map[string]any{
		"llm_input": string(inputJSON),
	}, nil
}

// 输出处理，将模型输出解析为结构化结果
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*domain.JobResumeMatch, error) {
	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}

	var output llmAggregatorOutput
	if err := json.Unmarshal([]byte(msg.Content), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}

	if len(output.Recommendations) == 0 {
		return nil, fmt.Errorf("recommendations cannot be empty; raw=%s", msg.Content)
	}

	var result *domain.JobResumeMatch
	if err := compose.ProcessState[*aggregatorState](ctx, func(_ context.Context, state *aggregatorState) error {
		if state == nil || state.Input == nil {
			return fmt.Errorf("aggregator state not initialized")
		}

		overall := state.Overall
		if output.OverallScore != nil {
			overall = *output.OverallScore
		}

		result = buildFinalMatch(state.Input, overall, state.MatchedAt, output.Recommendations)
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

// NewAggregatorAgent 创建匹配结果聚合Agent
func NewAggregatorAgent(ctx context.Context, llm model.ToolCallingChatModel) (*AggregatorAgent, error) {
	// 创建聊天模板
	chatTemplate, err := NewAggregatorChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[map[string]any, *domain.JobResumeMatch](
		compose.WithGenLocalState(func(context.Context) *aggregatorState {
			return &aggregatorState{}
		}),
	)
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(llm, compose.WithNodeName("chat_model"), compose.WithNodeKey("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &AggregatorAgent{
		chain: chain,
	}, nil
}

// GetChain 获取处理链
func (a *AggregatorAgent) GetChain() *compose.Chain[map[string]any, *domain.JobResumeMatch] {
	return a.chain
}

// Compile 编译链为可执行的Runnable
func (a *AggregatorAgent) Compile(ctx context.Context) (compose.Runnable[map[string]any, *domain.JobResumeMatch], error) {
	runnable, err := a.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return runnable, nil
}

// Process 处理匹配结果聚合
func (a *AggregatorAgent) Process(ctx context.Context, input map[string]any) (*domain.JobResumeMatch, error) {
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
func (a *AggregatorAgent) validateInput(input map[string]any) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	// 验证任务元数据
	taskMetaData, ok := input[domain.TaskMetaDataNode]
	if !ok {
		return fmt.Errorf("task metadata is required")
	}
	if metaData, ok := taskMetaData.(*domain.TaskMetaData); !ok {
		return fmt.Errorf("task metadata must be of type *TaskMetaData")
	} else {
		if metaData.JobID == "" {
			return fmt.Errorf("job_id in task metadata cannot be empty")
		}
		if metaData.ResumeID == "" {
			return fmt.Errorf("resume_id in task metadata cannot be empty")
		}
		if metaData.MatchTaskID == "" {
			return fmt.Errorf("match_task_id in task metadata cannot be empty")
		}
	}

	// 验证至少有一个Agent的输出
	agentOutputs := []string{
		domain.BasicInfoAgent,
		domain.SkillAgent,
		domain.ResponsibilityAgent,
		domain.ExperienceAgent,
		domain.EducationAgent,
		domain.IndustryAgent,
	}

	hasOutput := false
	for _, agentName := range agentOutputs {
		if output, ok := input[agentName]; ok {
			hasOutput = true
			// 验证输出类型
			switch agentName {
			case domain.BasicInfoAgent:
				if _, ok := output.(*domain.BasicMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *BasicMatchDetail", agentName)
				}
			case domain.SkillAgent:
				if _, ok := output.(*domain.SkillMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *SkillMatchDetail", agentName)
				}
			case domain.ResponsibilityAgent:
				if _, ok := output.(*domain.ResponsibilityMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *ResponsibilityMatchDetail", agentName)
				}
			case domain.ExperienceAgent:
				if _, ok := output.(*domain.ExperienceMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *ExperienceMatchDetail", agentName)
				}
			case domain.EducationAgent:
				if _, ok := output.(*domain.EducationMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *EducationMatchDetail", agentName)
				}
			case domain.IndustryAgent:
				if _, ok := output.(*domain.IndustryMatchDetail); !ok {
					return fmt.Errorf("%s output must be of type *IndustryMatchDetail", agentName)
				}
			}
		}
	}

	if !hasOutput {
		return fmt.Errorf("at least one agent output is required")
	}

	return nil
}

// validateOutput 验证输出数据
func (a *AggregatorAgent) validateOutput(output *domain.JobResumeMatch) error {
	if output == nil {
		return fmt.Errorf("output cannot be nil")
	}
	if output.OverallScore < 0 || output.OverallScore > 100 {
		return fmt.Errorf("overall score must be between 0 and 100, got %f", output.OverallScore)
	}

	// 验证各维度得分
	if output.BasicMatch != nil && (output.BasicMatch.Score < 0 || output.BasicMatch.Score > 100) {
		return fmt.Errorf("basic match score must be between 0 and 100, got %f", output.BasicMatch.Score)
	}
	if output.SkillMatch != nil && (output.SkillMatch.Score < 0 || output.SkillMatch.Score > 100) {
		return fmt.Errorf("skill match score must be between 0 and 100, got %f", output.SkillMatch.Score)
	}
	if output.ResponsibilityMatch != nil && (output.ResponsibilityMatch.Score < 0 || output.ResponsibilityMatch.Score > 100) {
		return fmt.Errorf("responsibility match score must be between 0 and 100, got %f", output.ResponsibilityMatch.Score)
	}
	if output.ExperienceMatch != nil && (output.ExperienceMatch.Score < 0 || output.ExperienceMatch.Score > 100) {
		return fmt.Errorf("experience match score must be between 0 and 100, got %f", output.ExperienceMatch.Score)
	}
	if output.EducationMatch != nil && (output.EducationMatch.Score < 0 || output.EducationMatch.Score > 100) {
		return fmt.Errorf("education match score must be between 0 and 100, got %f", output.EducationMatch.Score)
	}
	if output.IndustryMatch != nil && (output.IndustryMatch.Score < 0 || output.IndustryMatch.Score > 100) {
		return fmt.Errorf("industry match score must be between 0 and 100, got %f", output.IndustryMatch.Score)
	}

	return nil
}

// GetAgentType 返回Agent类型
func (a *AggregatorAgent) GetAgentType() string {
	return "AggregatorAgent"
}

// GetVersion 返回Agent版本
func (a *AggregatorAgent) GetVersion() string {
	return "1.1.0"
}

func collectDimensionScores(input *AggregatorInput) map[string]float64 {
	scores := map[string]float64{}
	if input.SkillMatch != nil {
		scores["skill"] = input.SkillMatch.Score
	}
	if input.ResponsibilityMatch != nil {
		scores["responsibility"] = input.ResponsibilityMatch.Score
	}
	if input.ExperienceMatch != nil {
		scores["experience"] = input.ExperienceMatch.Score
	}
	if input.EducationMatch != nil {
		scores["education"] = input.EducationMatch.Score
	}
	if input.IndustryMatch != nil {
		scores["industry"] = input.IndustryMatch.Score
	}
	if input.BasicMatch != nil {
		scores["basic"] = input.BasicMatch.Score
	}
	return scores
}

func calculateOverallScore(weights *domain.DimensionWeights, scores map[string]float64) float64 {
	w := domain.DefaultDimensionWeights
	if weights != nil {
		w = *weights
	}
	value := scores["skill"]*w.Skill +
		scores["responsibility"]*w.Responsibility +
		scores["experience"]*w.Experience +
		scores["education"]*w.Education +
		scores["industry"]*w.Industry +
		scores["basic"]*w.Basic
	return math.Round(value*10) / 10
}

func buildLLMInput(input *AggregatorInput, scores map[string]float64, overall float64) *llmAggregatorInput {
	weights := domain.DefaultDimensionWeights
	if input.Weights != nil {
		weights = *input.Weights
	}

	strengths, risks := collectStrengthsAndRisks(input)
	missing := collectMissingInfo(input)

	return &llmAggregatorInput{
		OverallScore:       overall,
		DimensionWeights:   weights,
		DimensionScores:    scores,
		DimensionSummaries: buildDimensionSummaries(input),
		Strengths:          strengths,
		Risks:              risks,
		MissingInfo:        missing,
	}
}

func buildDimensionSummaries(input *AggregatorInput) []llmDimensionSummary {
	summaries := []llmDimensionSummary{
		{
			Key:   "skill",
			Name:  "技能匹配",
			Score: scoreOrZero(input.SkillMatch),
			Highlights: nonEmptySlice(collectFirstN(mergeSlices(
				extractSkillStrengths(input.SkillMatch),
			), 6)),
			Risks: nonEmptySlice(collectFirstN(mergeSlices(
				extractSkillRisks(input.SkillMatch),
			), 6)),
		},
		{
			Key:   "responsibility",
			Name:  "职责匹配",
			Score: scoreOrZero(input.ResponsibilityMatch),
			Highlights: nonEmptySlice(collectFirstN(mergeSlices(
				extractResponsibilityStrengths(input.ResponsibilityMatch),
			), 6)),
			Risks: nonEmptySlice(collectFirstN(mergeSlices(
				extractResponsibilityRisks(input.ResponsibilityMatch),
			), 6)),
		},
		{
			Key:   "experience",
			Name:  "经验匹配",
			Score: scoreOrZero(input.ExperienceMatch),
			Highlights: nonEmptySlice(collectFirstN(mergeSlices(
				extractExperienceHighlights(input.ExperienceMatch),
			), 6)),
			Risks: nonEmptySlice(collectFirstN(mergeSlices(
				extractExperienceRisks(input.ExperienceMatch),
			), 6)),
		},
		{
			Key:        "education",
			Name:       "教育匹配",
			Score:      scoreOrZero(input.EducationMatch),
			Highlights: nonEmptySlice(extractEducationHighlights(input.EducationMatch)),
			Risks:      nonEmptySlice(extractEducationRisks(input.EducationMatch)),
		},
		{
			Key:        "industry",
			Name:       "行业匹配",
			Score:      scoreOrZero(input.IndustryMatch),
			Highlights: nonEmptySlice(extractIndustryHighlights(input.IndustryMatch)),
			Risks:      nonEmptySlice(extractIndustryRisks(input.IndustryMatch)),
		},
		{
			Key:        "basic",
			Name:       "基本信息",
			Score:      scoreOrZero(input.BasicMatch),
			Highlights: nonEmptySlice(extractBasicHighlights(input.BasicMatch)),
			Risks:      nonEmptySlice(extractBasicRisks(input.BasicMatch)),
		},
	}

	return summaries
}

func buildFinalMatch(input *AggregatorInput, overall float64, matchedAt time.Time, recs []string) *domain.JobResumeMatch {
	if matchedAt.IsZero() {
		matchedAt = time.Now()
	}

	return &domain.JobResumeMatch{
		TaskMetaData:        input.TaskMetaData,
		OverallScore:        overall,
		SkillMatch:          input.SkillMatch,
		ResponsibilityMatch: input.ResponsibilityMatch,
		ExperienceMatch:     input.ExperienceMatch,
		EducationMatch:      input.EducationMatch,
		IndustryMatch:       input.IndustryMatch,
		BasicMatch:          input.BasicMatch,
		MatchedAt:           matchedAt,
		Recommendations:     recs,
	}
}

func collectStrengthsAndRisks(input *AggregatorInput) ([]string, []string) {
	var strengths []string
	var risks []string

	strengths = append(strengths, extractSkillStrengths(input.SkillMatch)...)
	risks = append(risks, extractSkillRisks(input.SkillMatch)...)

	strengths = append(strengths, extractResponsibilityStrengths(input.ResponsibilityMatch)...)
	risks = append(risks, extractResponsibilityRisks(input.ResponsibilityMatch)...)

	strengths = append(strengths, extractExperienceHighlights(input.ExperienceMatch)...)
	risks = append(risks, extractExperienceRisks(input.ExperienceMatch)...)

	strengths = append(strengths, extractEducationHighlights(input.EducationMatch)...)
	risks = append(risks, extractEducationRisks(input.EducationMatch)...)

	strengths = append(strengths, extractIndustryHighlights(input.IndustryMatch)...)
	risks = append(risks, extractIndustryRisks(input.IndustryMatch)...)

	strengths = append(strengths, extractBasicHighlights(input.BasicMatch)...)
	risks = append(risks, extractBasicRisks(input.BasicMatch)...)

	return deduplicate(strengths), deduplicate(risks)
}

func collectMissingInfo(input *AggregatorInput) []string {
	var missing []string
	if input.SkillMatch != nil {
		for _, skill := range input.SkillMatch.MissingSkills {
			if skill != nil && skill.Skill != "" {
				missing = append(missing, fmt.Sprintf("缺少关键技能：%s", skill.Skill))
			}
		}
	}

	if input.EducationMatch != nil {
		for _, school := range input.EducationMatch.SchoolMatches {
			if school == nil {
				continue
			}
			if len(school.UniversityTypes) == 0 && school.School != "" {
				missing = append(missing, fmt.Sprintf("院校[%s]缺少类型标签，请人工确认", school.School))
			}
			if school.GPA == nil && school.School != "" {
				missing = append(missing, fmt.Sprintf("院校[%s]未提供GPA信息", school.School))
			}
		}
	}

	if input.BasicMatch != nil {
		for _, evidence := range input.BasicMatch.Evidence {
			if containsMissingHint(evidence) {
				missing = append(missing, evidence)
			}
		}
		if containsMissingHint(input.BasicMatch.Notes) {
			missing = append(missing, input.BasicMatch.Notes)
		}
	}

	return deduplicate(missing)
}

func containsMissingHint(text string) bool {
	if text == "" {
		return false
	}
	keywords := []string{"缺", "未提供", "无法", "待补充", "缺失"}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

func scoreOrZero(item any) float64 {
	switch v := item.(type) {
	case *domain.SkillMatchDetail:
		if v == nil {
			return 0
		}
		return v.Score
	case *domain.ResponsibilityMatchDetail:
		if v == nil {
			return 0
		}
		return v.Score
	case *domain.ExperienceMatchDetail:
		if v == nil {
			return 0
		}
		return v.Score
	case *domain.EducationMatchDetail:
		if v == nil {
			return 0
		}
		return v.Score
	case *domain.IndustryMatchDetail:
		if v == nil {
			return 0
		}
		return v.Score
	case *domain.BasicMatchDetail:
		if v == nil {
			return 0
		}
		return v.Score
	default:
		return 0
	}
}

func extractSkillStrengths(skill *domain.SkillMatchDetail) []string {
	if skill == nil {
		return nil
	}
	var strengths []string
	if skill.LLMAnalysis != nil {
		strengths = append(strengths, skill.LLMAnalysis.StrengthAreas...)
	}
	for _, matched := range skill.MatchedSkills {
		if matched != nil && matched.LLMAnalysis != nil && matched.LLMAnalysis.MatchReason != "" {
			strengths = append(strengths, matched.LLMAnalysis.MatchReason)
		}
	}
	return strengths
}

func extractSkillRisks(skill *domain.SkillMatchDetail) []string {
	if skill == nil {
		return nil
	}
	var risks []string
	if skill.LLMAnalysis != nil {
		risks = append(risks, skill.LLMAnalysis.GapAreas...)
	}
	for _, matched := range skill.MatchedSkills {
		if matched != nil && matched.LLMAnalysis != nil && matched.LLMAnalysis.ProficiencyGap != "" && matched.LLMAnalysis.ProficiencyGap != "none" {
			risks = append(risks, fmt.Sprintf("%s熟练度差距：%s", matched.JobSkillID, matched.LLMAnalysis.ProficiencyGap))
		}
	}
	for _, miss := range skill.MissingSkills {
		if miss != nil && miss.Skill != "" {
			risks = append(risks, fmt.Sprintf("缺少技能：%s", miss.Skill))
		}
	}
	return risks
}

func extractResponsibilityStrengths(resp *domain.ResponsibilityMatchDetail) []string {
	if resp == nil {
		return nil
	}
	var strengths []string
	for _, matched := range resp.MatchedResponsibilities {
		if matched == nil {
			continue
		}
		if matched.LLMAnalysis != nil {
			strengths = append(strengths, matched.LLMAnalysis.StrengthPoints...)
		}
		if matched.MatchReason != "" {
			strengths = append(strengths, matched.MatchReason)
		}
	}
	return strengths
}

func extractResponsibilityRisks(resp *domain.ResponsibilityMatchDetail) []string {
	if resp == nil {
		return nil
	}
	var risks []string
	for _, matched := range resp.MatchedResponsibilities {
		if matched == nil {
			continue
		}
		if matched.LLMAnalysis != nil {
			risks = append(risks, matched.LLMAnalysis.WeakPoints...)
		}
	}
	for _, unmatched := range resp.UnmatchedResponsibilities {
		if unmatched != nil && unmatched.Responsibility != "" {
			risks = append(risks, fmt.Sprintf("未覆盖职责：%s", unmatched.Responsibility))
		}
	}
	return risks
}

func extractExperienceHighlights(exp *domain.ExperienceMatchDetail) []string {
	if exp == nil {
		return nil
	}
	var highlights []string
	if exp.YearsMatch != nil {
		highlights = append(highlights, fmt.Sprintf("实际年限 %.1f 年，要求 %.1f 年", exp.YearsMatch.ActualYears, exp.YearsMatch.RequiredYears))
	}
	for _, pos := range exp.PositionMatches {
		if pos != nil && pos.Position != "" {
			label := fmt.Sprintf("相关职位：%s（相关度%.0f分）", pos.Position, pos.Score)
			if pos.ExperienceType != "" && pos.ExperienceType != string(consts.ExperienceTypeWork) {
				label += fmt.Sprintf("，来源：%s", translateExperienceType(pos.ExperienceType))
			}
			highlights = append(highlights, label)
		}
	}
	return highlights
}

func extractExperienceRisks(exp *domain.ExperienceMatchDetail) []string {
	if exp == nil {
		return nil
	}
	var risks []string
	if exp.YearsMatch != nil && exp.YearsMatch.RequiredYears > 0 && exp.YearsMatch.ActualYears < exp.YearsMatch.RequiredYears {
		risks = append(risks, fmt.Sprintf("工作年限不足：缺少 %.1f 年", exp.YearsMatch.Gap))
	}
	return risks
}

func extractEducationHighlights(edu *domain.EducationMatchDetail) []string {
	if edu == nil {
		return nil
	}
	var highlights []string
	if edu.DegreeMatch != nil && edu.DegreeMatch.Meets {
		if edu.DegreeMatch.ActualDegree != "" {
			highlights = append(highlights, fmt.Sprintf("学历符合要求：%s", edu.DegreeMatch.ActualDegree))
		} else {
			highlights = append(highlights, "学历符合要求")
		}
	}
	for _, school := range edu.SchoolMatches {
		if school != nil && school.School != "" {
			var detailParts []string
			if len(school.UniversityTypes) > 0 {
				detailParts = append(detailParts, fmt.Sprintf("类型：%s", strings.Join(translateUniversityTypes(school.UniversityTypes), "/")))
			}
			if school.GPA != nil && *school.GPA > 0 {
				detailParts = append(detailParts, fmt.Sprintf("GPA %.2f", *school.GPA))
			}
			detail := ""
			if len(detailParts) > 0 {
				detail = "，" + strings.Join(detailParts, "，")
			}
			highlights = append(highlights, fmt.Sprintf("院校：%s（声誉%.0f分%s）", school.School, school.Score, detail))
		}
	}
	return highlights
}

func extractEducationRisks(edu *domain.EducationMatchDetail) []string {
	if edu == nil {
		return nil
	}
	var risks []string
	if edu.DegreeMatch != nil && !edu.DegreeMatch.Meets {
		risks = append(risks, fmt.Sprintf("学历不满足要求：需要%s，实际%s", edu.DegreeMatch.RequiredDegree, edu.DegreeMatch.ActualDegree))
	}
	for _, major := range edu.MajorMatches {
		if major != nil && major.Relevance < 70 {
			risks = append(risks, fmt.Sprintf("专业相关性低：%s（%.0f分）", major.Major, major.Relevance))
		}
	}
	for _, school := range edu.SchoolMatches {
		if school == nil || school.School == "" {
			continue
		}
		if school.Score < 70 {
			risks = append(risks, fmt.Sprintf("院校匹配偏低：%s（%.0f分）", school.School, school.Score))
		}
		if len(school.UniversityTypes) == 0 {
			risks = append(risks, fmt.Sprintf("院校[%s]缺少类型信息", school.School))
		}
		if school.GPA == nil {
			risks = append(risks, fmt.Sprintf("院校[%s]未提供GPA", school.School))
		}
	}
	return risks
}

func extractIndustryHighlights(ind *domain.IndustryMatchDetail) []string {
	if ind == nil {
		return nil
	}
	var highlights []string
	for _, item := range ind.IndustryMatches {
		if item != nil && item.Industry != "" {
			highlights = append(highlights, fmt.Sprintf("相关行业：%s（%.0f分）", item.Industry, item.Score))
		}
	}
	return highlights
}

func extractIndustryRisks(ind *domain.IndustryMatchDetail) []string {
	if ind == nil {
		return nil
	}
	var risks []string
	for _, item := range ind.IndustryMatches {
		if item != nil && item.Score < 60 && item.Industry != "" {
			risks = append(risks, fmt.Sprintf("行业匹配偏低：%s（%.0f分）", item.Industry, item.Score))
		}
	}
	return risks
}

func extractBasicHighlights(basic *domain.BasicMatchDetail) []string {
	if basic == nil {
		return nil
	}
	return append([]string{}, basic.Evidence...)
}

func extractBasicRisks(basic *domain.BasicMatchDetail) []string {
	if basic == nil {
		return nil
	}
	var risks []string
	for _, evidence := range basic.Evidence {
		if containsMissingHint(evidence) {
			risks = append(risks, evidence)
		}
	}
	if containsMissingHint(basic.Notes) {
		risks = append(risks, basic.Notes)
	}
	return risks
}

func mergeSlices(slices ...[]string) []string {
	var result []string
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

func deduplicate(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	var result []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func collectFirstN(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return items[:n]
}

func nonEmptySlice(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	return items
}

func translateUniversityTypes(types []consts.UniversityType) []string {
	if len(types) == 0 {
		return nil
	}
	result := make([]string, 0, len(types))
	for _, t := range types {
		if label := translateUniversityType(t); label != "" {
			result = append(result, label)
		}
	}
	return result
}

func translateUniversityType(t consts.UniversityType) string {
	switch t {
	case consts.UniversityType985:
		return "985"
	case consts.UniversityType211:
		return "211"
	case consts.UniversityTypeDoubleFirstClass:
		return "双一流"
	case consts.UniversityTypeQSTop100:
		return "QS前100"
	case consts.UniversityTypeOrdinary:
		return "普通高校"
	default:
		return string(t)
	}
}

func translateExperienceType(t string) string {
	switch t {
	case string(consts.ExperienceTypeWork):
		return "全职"
	case string(consts.ExperienceTypeInternship):
		return "实习"
	case string(consts.ExperienceTypeOrganization):
		return "组织/社团"
	case string(consts.ExperienceTypeVolunteer):
		return "志愿"
	default:
		return t
	}
}
