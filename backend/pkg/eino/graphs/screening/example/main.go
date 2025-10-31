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
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/screening/matching/weights"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
	"github.com/cloudwego/eino-ext/callbacks/langsmith"
	"github.com/cloudwego/eino/callbacks"
)

func main() {

	// 获取配置
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	cfgls := &langsmith.Config{
		APIKey: cfg.Langsmith.APIKey,
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

	//1.调用调试服务初始化函数
	// err := devops.Init(ctx)
	// if err != nil {
	// 	log.Fatalf("[eino dev] init failed, err=%v", err)
	// 	return
	// }

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
	// 准备测试数据
	matchInput := createTestMatchInput()

	agent, err := weights.NewWeightPlannerAgent(ctx, chatModel)
	if err != nil {
		log.Fatalf("创建权重规划 Agent 失败: %v", err)
	}

	runnable, err := agent.Compile(ctx)
	if err != nil {
		log.Fatalf("编译权重规划链失败: %v", err)
	}

	resultWeight, err := runnable.Invoke(ctx, &domain.WeightInferenceInput{
		JobProfile: matchInput.JobProfile,
	})
	if err != nil {
		log.Fatalf("执行权重规划失败: %v", err)
	}

	fmt.Println("权重规划结果:", resultWeight)

	matchInput.DimensionWeights = &resultWeight.Weights

	// 创建screening图
	screeningGraph, err := screening.NewScreeningChatGraph(ctx, chatModel, cfg)
	if err != nil {
		log.Fatalf("failed to create screening graph: %v", err)
	}

	compiledGraph, err := screeningGraph.Compile(ctx)
	if err != nil {
		log.Fatalf("failed to compile screening graph: %v", err)
	}

	outputCollector := screening.NewAgentCallbackCollector()
	outputCollector.EnableDetailedCallbacks() // 开启详细回调便于单独示例调试

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
	fmt.Println("----------------------------------------")

	// 详细匹配结果
	printMatchDetails(result)

	// 输出完整JSON结果（可选）
	if jsonData, err := json.MarshalIndent(result, "", "  "); err == nil {
		fmt.Println("\n完整匹配结果JSON:")
		fmt.Println(string(jsonData))
	}

	// 打印各 Agent 的 Token 消耗情况
	fmt.Println("\n========== Token 消耗统计 ==========")
	var totalPromptTokens, totalCompletionTokens, totalTokens int

	if basicTokenUsage, ok := outputCollector.BasicInfoTokenUsage(); ok {
		fmt.Printf("[BasicInfoAgent] Prompt: %d, Completion: %d, Total: %d tokens\n",
			basicTokenUsage.PromptTokens, basicTokenUsage.CompletionTokens, basicTokenUsage.TotalTokens)
		totalPromptTokens += basicTokenUsage.PromptTokens
		totalCompletionTokens += basicTokenUsage.CompletionTokens
		totalTokens += basicTokenUsage.TotalTokens
	}

	if educationTokenUsage, ok := outputCollector.EducationTokenUsage(); ok {
		fmt.Printf("[EducationAgent] Prompt: %d, Completion: %d, Total: %d tokens\n",
			educationTokenUsage.PromptTokens, educationTokenUsage.CompletionTokens, educationTokenUsage.TotalTokens)
		totalPromptTokens += educationTokenUsage.PromptTokens
		totalCompletionTokens += educationTokenUsage.CompletionTokens
		totalTokens += educationTokenUsage.TotalTokens
	}

	if experienceTokenUsage, ok := outputCollector.ExperienceTokenUsage(); ok {
		fmt.Printf("[ExperienceAgent] Prompt: %d, Completion: %d, Total: %d tokens\n",
			experienceTokenUsage.PromptTokens, experienceTokenUsage.CompletionTokens, experienceTokenUsage.TotalTokens)
		totalPromptTokens += experienceTokenUsage.PromptTokens
		totalCompletionTokens += experienceTokenUsage.CompletionTokens
		totalTokens += experienceTokenUsage.TotalTokens
	}

	if industryTokenUsage, ok := outputCollector.IndustryTokenUsage(); ok {
		fmt.Printf("[IndustryAgent] Prompt: %d, Completion: %d, Total: %d tokens\n",
			industryTokenUsage.PromptTokens, industryTokenUsage.CompletionTokens, industryTokenUsage.TotalTokens)
		totalPromptTokens += industryTokenUsage.PromptTokens
		totalCompletionTokens += industryTokenUsage.CompletionTokens
		totalTokens += industryTokenUsage.TotalTokens
	}

	if responsibilityTokenUsage, ok := outputCollector.ResponsibilityTokenUsage(); ok {
		fmt.Printf("[ResponsibilityAgent] Prompt: %d, Completion: %d, Total: %d tokens\n",
			responsibilityTokenUsage.PromptTokens, responsibilityTokenUsage.CompletionTokens, responsibilityTokenUsage.TotalTokens)
		totalPromptTokens += responsibilityTokenUsage.PromptTokens
		totalCompletionTokens += responsibilityTokenUsage.CompletionTokens
		totalTokens += responsibilityTokenUsage.TotalTokens
	}

	if skillTokenUsage, ok := outputCollector.SkillTokenUsage(); ok {
		fmt.Printf("[SkillAgent] Prompt: %d, Completion: %d, Total: %d tokens\n",
			skillTokenUsage.PromptTokens, skillTokenUsage.CompletionTokens, skillTokenUsage.TotalTokens)
		totalPromptTokens += skillTokenUsage.PromptTokens
		totalCompletionTokens += skillTokenUsage.CompletionTokens
		totalTokens += skillTokenUsage.TotalTokens
	}

	if aggregatorTokenUsage, ok := outputCollector.AggregatorTokenUsage(); ok {
		fmt.Printf("[AggregatorAgent] Prompt: %d, Completion: %d, Total: %d tokens\n",
			aggregatorTokenUsage.PromptTokens, aggregatorTokenUsage.CompletionTokens, aggregatorTokenUsage.TotalTokens)
		totalPromptTokens += aggregatorTokenUsage.PromptTokens
		totalCompletionTokens += aggregatorTokenUsage.CompletionTokens
		totalTokens += aggregatorTokenUsage.TotalTokens
	}

	fmt.Println("----------------------------------------")
	fmt.Printf("总计消耗: Prompt: %d, Completion: %d, Total: %d tokens\n",
		totalPromptTokens, totalCompletionTokens, totalTokens)
	fmt.Println("========================================")

	// Blocking process exits
	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// <-sigs

	// Exit
	log.Printf("[eino dev] shutting down\n")
}

// createTestMatchInput 创建测试用的匹配输入数据
func createTestMatchInput() *domain.MatchInput {
	// 创建测试职位数据 - 蚂蚁集团支付系统高级后端工程师
	jobProfile := &domain.JobProfileDetail{
		JobProfile: &domain.JobProfile{
			ID:          "job-ant-pay-001",
			Name:        "支付系统高级后端工程师",
			Description: stringPtr("负责蚂蚁集团支付宝核心支付系统的架构设计与开发，支撑日均数十亿笔交易，要求具备高并发、分布式系统设计经验，熟悉金融支付业务场景"),
			Location:    stringPtr("杭州"),
			SalaryMin:   float64Ptr(35000),
			SalaryMax:   float64Ptr(60000),
		},
		Responsibilities: []*domain.JobResponsibility{
			{
				ID:             "resp-001",
				JobID:          "job-ant-pay-001",
				Responsibility: "负责支付核心链路系统的架构设计、开发和优化，确保系统高可用性和稳定性",
			},
			{
				ID:             "resp-002",
				JobID:          "job-ant-pay-001",
				Responsibility: "参与分布式事务处理、资金安全、风控系统等关键模块技术方案设计",
			},
			{
				ID:             "resp-003",
				JobID:          "job-ant-pay-001",
				Responsibility: "优化系统性能，解决高并发场景下的技术难题，提升系统吞吐量和响应速度",
			},
			{
				ID:             "resp-004",
				JobID:          "job-ant-pay-001",
				Responsibility: "参与技术团队建设，指导初中级工程师，推动技术规范和最佳实践落地",
			},
			{
				ID:             "resp-005",
				JobID:          "job-ant-pay-001",
				Responsibility: "与产品、运营团队协作，支撑业务快速发展和创新需求",
			},
		},
		Skills: []*domain.JobSkill{
			{
				ID:      "skill-001",
				JobID:   "job-ant-pay-001",
				SkillID: "java",
				Skill:   "Java",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-002",
				JobID:   "job-ant-pay-001",
				SkillID: "spring-boot",
				Skill:   "Spring Boot",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-003",
				JobID:   "job-ant-pay-001",
				SkillID: "mysql",
				Skill:   "MySQL",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-004",
				JobID:   "job-ant-pay-001",
				SkillID: "redis",
				Skill:   "Redis",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-005",
				JobID:   "job-ant-pay-001",
				SkillID: "kafka",
				Skill:   "Kafka",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-006",
				JobID:   "job-ant-pay-001",
				SkillID: "distributed-system",
				Skill:   "分布式系统",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-007",
				JobID:   "job-ant-pay-001",
				SkillID: "microservices",
				Skill:   "微服务架构",
				Type:    string(domain.SkillTypeRequired),
			},
			{
				ID:      "skill-008",
				JobID:   "job-ant-pay-001",
				SkillID: "kubernetes",
				Skill:   "Kubernetes",
				Type:    string(domain.SkillTypeBonus),
			},
			{
				ID:      "skill-009",
				JobID:   "job-ant-pay-001",
				SkillID: "go",
				Skill:   "Go语言",
				Type:    string(domain.SkillTypeBonus),
			},
		},
		ExperienceRequirements: []*domain.JobExperienceRequirement{
			{
				ID:             "exp-001",
				JobID:          "job-ant-pay-001",
				ExperienceType: "five_to_ten_years",
				MinYears:       5,
				IdealYears:     8,
			},
		},
		EducationRequirements: []*domain.JobEducationRequirement{
			{
				ID:            "edu-001",
				JobID:         "job-ant-pay-001",
				EducationType: "bachelor",
			},
		},
		IndustryRequirements: []*domain.JobIndustryRequirement{
			{
				ID:       "ind-001",
				JobID:    "job-ant-pay-001",
				Industry: "金融科技",
			},
			{
				ID:       "ind-002",
				JobID:    "job-ant-pay-001",
				Industry: "互联网",
			},
		},
	}

	// 创建测试简历数据 - 来自美团支付团队的候选人
	resume := &domain.ResumeDetail{
		Resume: &domain.Resume{
			ID:               "resume-li-ming-001",
			Name:             "李明",
			Gender:           "男",
			Email:            "liming.dev@gmail.com",
			Phone:            "13912345678",
			CurrentCity:      "北京",
			HighestEducation: "硕士",
			YearsExperience:  6.5,
		},
		Experiences: []*domain.ResumeExperience{
			{
				ID:          "exp-001",
				ResumeID:    "resume-li-ming-001",
				Company:     "美团",
				Position:    "高级后端工程师",
				Title:       "支付系统技术专家",
				StartDate:   timePtr(time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:     nil, // 当前工作
				Description: "负责美团支付核心系统架构设计与开发，主导了支付路由、资金清结算、风控系统等核心模块重构。使用Java + Spring Boot构建微服务架构，日处理交易量超过5000万笔。主要技术栈：Java、Spring Boot、MySQL、Redis、Kafka、Elasticsearch。负责团队技术方案评审，指导3名初级工程师。",
			},
			{
				ID:          "exp-002",
				ResumeID:    "resume-li-ming-001",
				Company:     "滴滴出行",
				Position:    "后端工程师",
				Title:       "高级后端工程师",
				StartDate:   timePtr(time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:     timePtr(time.Date(2021, 2, 28, 0, 0, 0, 0, time.UTC)),
				Description: "参与滴滴钱包支付系统开发，负责订单支付、退款、对账等核心业务模块。使用分布式事务保证数据一致性，通过缓存优化和数据库分库分表提升系统性能。参与了从单体架构向微服务架构的迁移项目，负责用户钱包服务的拆分和重构。",
			},
			{
				ID:          "exp-003",
				ResumeID:    "resume-li-ming-001",
				Company:     "京东",
				Position:    "软件工程师",
				Title:       "后端开发工程师",
				StartDate:   timePtr(time.Date(2017, 7, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:     timePtr(time.Date(2019, 6, 30, 0, 0, 0, 0, time.UTC)),
				Description: "参与京东商城订单系统开发，负责订单创建、库存扣减、价格计算等核心逻辑。使用Spring框架开发RESTful API，通过Redis缓存提升系统响应速度。参与了双11大促备战，通过性能优化和容量规划确保系统稳定运行。",
			},
		},
		Educations: []*domain.ResumeEducation{
			{
				ID:        "edu-001",
				ResumeID:  "resume-li-ming-001",
				School:    "华中科技大学",
				Degree:    "硕士",
				Major:     "软件工程",
				StartDate: timePtr(time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:   timePtr(time.Date(2017, 6, 30, 0, 0, 0, 0, time.UTC)),
			},
			{
				ID:        "edu-002",
				ResumeID:  "resume-li-ming-001",
				School:    "华中科技大学",
				Degree:    "本科",
				Major:     "计算机科学与技术",
				StartDate: timePtr(time.Date(2011, 9, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:   timePtr(time.Date(2015, 6, 30, 0, 0, 0, 0, time.UTC)),
			},
		},
		Skills: []*domain.ResumeSkill{
			{
				ID:          "skill-001",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "Java",
				Level:       "精通",
				Description: "6年Java开发经验，熟练掌握Java 8/11新特性，深入理解JVM原理和性能调优",
			},
			{
				ID:          "skill-002",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "Spring Boot",
				Level:       "精通",
				Description: "5年Spring Boot开发经验，熟悉Spring生态，包括Spring Cloud微服务架构",
			},
			{
				ID:          "skill-003",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "MySQL",
				Level:       "熟练",
				Description: "熟悉MySQL数据库设计、索引优化、分库分表，有大数据量处理经验",
			},
			{
				ID:          "skill-004",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "Redis",
				Level:       "熟练",
				Description: "熟练使用Redis进行缓存设计，了解Redis集群和哨兵模式",
			},
			{
				ID:          "skill-005",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "Kafka",
				Level:       "熟练",
				Description: "熟悉Kafka消息队列，用于异步处理和系统解耦，有高吞吐量场景使用经验",
			},
			{
				ID:          "skill-006",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "分布式系统",
				Level:       "熟练",
				Description: "具备分布式系统设计经验，熟悉CAP理论、分布式事务、服务治理等",
			},
			{
				ID:          "skill-007",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "微服务架构",
				Level:       "熟练",
				Description: "有微服务架构设计和实施经验，熟悉服务拆分、API网关、配置中心等",
			},
			{
				ID:          "skill-008",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "Docker",
				Level:       "熟练",
				Description: "熟练使用Docker进行应用容器化，有生产环境部署经验",
			},
			{
				ID:          "skill-009",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "Elasticsearch",
				Level:       "一般",
				Description: "了解Elasticsearch搜索引擎，用于日志分析和数据检索",
			},
			{
				ID:          "skill-010",
				ResumeID:    "resume-li-ming-001",
				SkillName:   "Python",
				Level:       "一般",
				Description: "有Python开发经验，主要用于数据分析和脚本编写",
			},
		},
		Projects: []*domain.ResumeProject{
			{
				ID:           "proj-001",
				ResumeID:     "resume-li-ming-001",
				Name:         "美团支付核心系统重构",
				Role:         "技术负责人",
				StartDate:    timePtr(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:      timePtr(time.Date(2023, 6, 30, 0, 0, 0, 0, time.UTC)),
				Description:  "主导美团支付核心系统从单体向微服务架构迁移，拆分为支付路由、资金清算、风控等10+个微服务。通过引入分布式事务框架Seata保证数据一致性，使用Redis + MySQL实现高可用架构。项目上线后系统可用性提升至99.99%，平均响应时间降低40%。",
				Technologies: "Java,Spring Boot,MySQL,Redis,Kafka,Seata,Docker",
			},
			{
				ID:           "proj-002",
				ResumeID:     "resume-li-ming-001",
				Name:         "高并发支付网关优化",
				Role:         "核心开发",
				StartDate:    timePtr(time.Date(2021, 8, 1, 0, 0, 0, 0, time.UTC)),
				EndDate:      timePtr(time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC)),
				Description:  "针对支付网关在高峰期出现的性能瓶颈进行优化，通过连接池优化、缓存策略调整、数据库读写分离等手段，将系统TPS从2万提升至8万。引入熔断降级机制，提升系统稳定性。",
				Technologies: "Java,Spring Boot,MySQL,Redis,Hystrix",
			},
		},
	}

	return &domain.MatchInput{
		JobProfile:       jobProfile,
		Resume:           resume,
		MatchTaskID:      "match-task-ant-pay-001",
		DimensionWeights: &domain.DefaultDimensionWeights,
	}
}

// printMatchDetails 打印匹配详情
func printMatchDetails(result *domain.JobResumeMatch) {
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
