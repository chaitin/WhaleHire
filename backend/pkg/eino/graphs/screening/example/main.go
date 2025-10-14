//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/domain"
	screening "github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/types"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
	"github.com/cloudwego/eino-ext/callbacks/langsmith"
	"github.com/cloudwego/eino/callbacks"
)

func main() {

	cfgls := &langsmith.Config{
		APIKey: "xxx",
	}
	// ft := langsmith.NewFlowTrace(cfg)
	cbh, err := langsmith.NewLangsmithHandler(cfgls)
	if err != nil {
		log.Fatal(err)
	}

	// 设置全局上报handler
	callbacks.AppendGlobalHandlers(cbh)

	ctx := context.Background()
	ctx = langsmith.SetTrace(ctx,
		langsmith.WithSessionName("WhaleHire Screening"), // 设置langsmith上报项目名称
	)

	// 1.调用调试服务初始化函数
	// err := devops.Init(ctx)
	// if err != nil {
	// 	log.Fatalf("[eino dev] init failed, err=%v", err)
	// 	return
	// }

	// 获取配置
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	factory := models.NewModelFactory()
	// 注册OpenAI兼容模型（例如 GPT5 名称通过配置传入）
	openaiConfig := &models.OpenAIConfig{
		APIKey:         cfg.GeneralAgent.LLM.APIKey,
		BaseURL:        cfg.GeneralAgent.LLM.BaseURL,
		Model:          cfg.GeneralAgent.LLM.ModelName,
		ResponseFormat: "json_object",
	}
	openaiManager := models.NewOpenAIModelManager(openaiConfig)
	factory.Register(models.ModelTypeOpenAI, cfg.GeneralAgent.LLM.ModelName, openaiManager)

	chatModel, err := factory.GetModel(ctx, models.ModelTypeOpenAI, cfg.GeneralAgent.LLM.ModelName)
	if err != nil {
		log.Fatalf("failed to get model: %v", err)
	}

	// 创建screening图
	screeningGraph, err := screening.NewScreeningChatGraph(ctx, chatModel, cfg)
	if err != nil {
		log.Fatalf("failed to create screening graph: %v", err)
	}

	compiledGraph, err := screeningGraph.Compile(ctx)
	if err != nil {
		log.Fatalf("failed to compile screening graph: %v", err)
	}

	outputCollector := screening.NewAgentOutputCollector()

	// 准备测试数据
	matchInput := createTestMatchInput()

	fmt.Println("开始执行简历智能匹配...")
	fmt.Printf("职位: %s\n", matchInput.JobProfile.Name)
	fmt.Printf("候选人: %s\n", matchInput.Resume.Name)
	fmt.Println("----------------------------------------")

	// 执行匹配
	result, err := compiledGraph.Invoke(ctx, matchInput, outputCollector.ComposeOptions()...)
	if err != nil {
		log.Fatalf("failed to invoke screening graph: %v", err)
	}

	// 输出结果
	fmt.Printf("匹配完成！综合得分: %.2f\n", result.OverallScore)
	fmt.Printf("匹配等级: %s\n", types.GetMatchLevel(result.OverallScore))
	fmt.Println("----------------------------------------")

	// 详细匹配结果
	printMatchDetails(result)

	// 输出完整JSON结果（可选）
	if jsonData, err := json.MarshalIndent(result, "", "  "); err == nil {
		fmt.Println("\n完整匹配结果JSON:")
		fmt.Println(string(jsonData))
	}

	// 打印各 Agent 输出，便于回调调试
	if basic, ok := outputCollector.BasicInfo(); ok {
		fmt.Printf("\n[BasicInfoAgent] 匹配得分: %.2f\n", basic.Score)
	}
	if education, ok := outputCollector.Education(); ok {
		fmt.Printf("[EducationAgent] 匹配得分: %.2f\n", education.Score)
	}
	if experience, ok := outputCollector.Experience(); ok {
		fmt.Printf("[ExperienceAgent] 匹配得分: %.2f\n", experience.Score)
	}
	if industry, ok := outputCollector.Industry(); ok {
		fmt.Printf("[IndustryAgent] 匹配得分: %.2f\n", industry.Score)
	}
	if responsibility, ok := outputCollector.Responsibility(); ok {
		fmt.Printf("[ResponsibilityAgent] 匹配得分: %.2f\n", responsibility.Score)
	}
	if skill, ok := outputCollector.Skill(); ok {
		fmt.Printf("[SkillAgent] 匹配得分: %.2f\n", skill.Score)
	}
	if aggregated, ok := outputCollector.AggregatedMatch(); ok {
		fmt.Printf("[AggregatorAgent] 综合得分: %.2f\n", aggregated.OverallScore)
	}

	// Blocking process exits
	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// <-sigs

	// Exit
	log.Printf("[eino dev] shutting down\n")
}

// createTestMatchInput 创建测试用的匹配输入数据
func createTestMatchInput() *types.MatchInput {
	// 创建测试职位数据
	jobProfile := &domain.JobProfileDetail{
		JobProfile: &domain.JobProfile{
			ID:          "job-001",
			Name:        "高级Go后端工程师",
			Description: stringPtr("负责后端服务开发，微服务架构设计与实现"),
			Location:    stringPtr("北京"),
			SalaryMin:   float64Ptr(25000),
			SalaryMax:   float64Ptr(40000),
		},
		Responsibilities: []*domain.JobResponsibility{
			{
				ID:             "resp-001",
				JobID:          "job-001",
				Responsibility: "负责后端API设计与开发",
			},
			{
				ID:             "resp-002",
				JobID:          "job-001",
				Responsibility: "参与系统架构设计和技术选型",
			},
			{
				ID:             "resp-003",
				JobID:          "job-001",
				Responsibility: "优化系统性能，解决技术难题",
			},
		},
		Skills: []*domain.JobSkill{
			{
				ID:      "skill-001",
				JobID:   "job-001",
				SkillID: "go-lang",
				Skill:   "Go语言",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-002",
				JobID:   "job-001",
				SkillID: "mysql",
				Skill:   "MySQL",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-003",
				JobID:   "job-001",
				SkillID: "redis",
				Skill:   "Redis",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-004",
				JobID:   "job-001",
				SkillID: "docker",
				Skill:   "Docker",
				Type:    string(domain.SkillTypeBonus),
			},
		},
		ExperienceRequirements: []*domain.JobExperienceRequirement{
			{
				ID:             "exp-001",
				JobID:          "job-001",
				ExperienceType: "three_to_five_years",
				MinYears:       3,
				IdealYears:     5,
			},
		},
		EducationRequirements: []*domain.JobEducationRequirement{
			{
				ID:            "edu-001",
				JobID:         "job-001",
				EducationType: "bachelor",
			},
		},
		IndustryRequirements: []*domain.JobIndustryRequirement{
			{
				ID:       "ind-001",
				JobID:    "job-001",
				Industry: "互联网",
			},
		},
	}

	// 创建测试简历数据
	resume := &domain.ResumeDetail{
		Resume: &domain.Resume{
			ID:               "resume-001",
			Name:             "张三",
			Gender:           "男",
			Email:            "zhangsan@example.com",
			Phone:            "13800001234",
			CurrentCity:      "北京",
			HighestEducation: "本科",
			YearsExperience:  4.5,
		},
		Experiences: []*domain.ResumeExperience{
			{
				ID:          "exp-001",
				ResumeID:    "resume-001",
				Company:     "字节跳动",
				Position:    "后端工程师",
				Title:       "高级后端工程师",
				StartDate:   timePtr(time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:     timePtr(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)),
				Description: "负责推荐系统后端服务开发，使用Go语言开发微服务，处理日均千万级请求",
			},
			{
				ID:          "exp-002",
				ResumeID:    "resume-001",
				Company:     "腾讯",
				Position:    "软件工程师",
				Title:       "后端开发工程师",
				StartDate:   timePtr(time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:     timePtr(time.Date(2021, 6, 30, 0, 0, 0, 0, time.UTC)),
				Description: "参与社交产品后端开发，负责用户系统和消息系统的设计与实现",
			},
		},
		Educations: []*domain.ResumeEducation{
			{
				ID:        "edu-001",
				ResumeID:  "resume-001",
				School:    "清华大学",
				Degree:    "本科",
				Major:     "计算机科学与技术",
				StartDate: timePtr(time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:   timePtr(time.Date(2019, 6, 30, 0, 0, 0, 0, time.UTC)),
			},
		},
		Skills: []*domain.ResumeSkill{
			{
				ID:          "skill-001",
				ResumeID:    "resume-001",
				SkillName:   "Go",
				Level:       "熟练",
				Description: "3年Go开发经验，熟悉gin、echo等框架",
			},
			{
				ID:          "skill-002",
				ResumeID:    "resume-001",
				SkillName:   "MySQL",
				Level:       "熟练",
				Description: "熟悉MySQL数据库设计和优化",
			},
			{
				ID:          "skill-003",
				ResumeID:    "resume-001",
				SkillName:   "Redis",
				Level:       "熟练",
				Description: "熟悉Redis缓存设计和使用",
			},
			{
				ID:          "skill-004",
				ResumeID:    "resume-001",
				SkillName:   "Python",
				Level:       "一般",
				Description: "有Python开发经验，主要用于数据处理",
			},
		},
	}

	return &types.MatchInput{
		JobProfile:       jobProfile,
		Resume:           resume,
		MatchTaskID:      "match-task-001",
		DimensionWeights: &types.DefaultDimensionWeights,
	}
}

// printMatchDetails 打印匹配详情
func printMatchDetails(result *types.JobResumeMatch) {
	fmt.Println("详细匹配结果:")

	if result.SkillMatch != nil {
		fmt.Printf("技能匹配: %.2f分\n", result.SkillMatch.Score)
		fmt.Printf("  匹配技能数: %d\n", len(result.SkillMatch.MatchedSkills))
		fmt.Printf("  缺失技能数: %d\n", len(result.SkillMatch.MissingSkills))
		fmt.Printf("  额外技能数: %d\n", len(result.SkillMatch.ExtraSkills))
	}

	if result.ExperienceMatch != nil {
		fmt.Printf("经验匹配: %.2f分\n", result.ExperienceMatch.Score)
		if result.ExperienceMatch.YearsMatch != nil {
			fmt.Printf("  工作年限: %.1f年 (要求: %.1f年)\n",
				result.ExperienceMatch.YearsMatch.ActualYears,
				result.ExperienceMatch.YearsMatch.RequiredYears)
		}
	}

	if result.EducationMatch != nil {
		fmt.Printf("教育匹配: %.2f分\n", result.EducationMatch.Score)
		if result.EducationMatch.DegreeMatch != nil {
			fmt.Printf("  学历: %s (要求: %s)\n",
				result.EducationMatch.DegreeMatch.ActualDegree,
				result.EducationMatch.DegreeMatch.RequiredDegree)
		}
	}

	if result.ResponsibilityMatch != nil {
		fmt.Printf("职责匹配: %.2f分\n", result.ResponsibilityMatch.Score)
		fmt.Printf("  匹配职责数: %d\n", len(result.ResponsibilityMatch.MatchedResponsibilities))
	}

	if result.IndustryMatch != nil {
		fmt.Printf("行业匹配: %.2f分\n", result.IndustryMatch.Score)
	}

	if result.BasicMatch != nil {
		fmt.Printf("基本信息匹配: %.2f分\n", result.BasicMatch.Score)
	}

	if len(result.Recommendations) > 0 {
		fmt.Println("\n匹配建议:")
		for i, rec := range result.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec)
		}
	}
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func timePtr(t time.Time) *time.Time {
	return &t
}
