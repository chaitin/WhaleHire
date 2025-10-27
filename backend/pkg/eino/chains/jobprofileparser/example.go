//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofileparser"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
)

// ExampleUsage 展示如何使用岗位画像解析链
func main() {
	ctx := context.Background()
	// 获取配置
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	factory := models.NewModelFactory()
	// 注册OpenAI兼容模型（例如 GPT5 名称通过配置传入）
	openaiConfig := &models.OpenAIConfig{
		APIKey:  cfg.GeneralAgent.LLM.APIKey,
		BaseURL: cfg.GeneralAgent.LLM.BaseURL,
		Model:   cfg.GeneralAgent.LLM.ModelName,
	}
	openaiManager := models.NewOpenAIModelManager(openaiConfig)
	factory.Register(models.ModelTypeOpenAI, cfg.GeneralAgent.LLM.ModelName, openaiManager)

	chatModel, err := factory.GetModel(ctx, models.ModelTypeOpenAI, cfg.GeneralAgent.LLM.ModelName)
	if err != nil {
		log.Fatalf("failed to get model: %v", err)
	}

	// 创建岗位画像解析链
	chain, err := jobprofileparser.NewJobProfileParserChain(ctx, chatModel)
	if err != nil {
		log.Fatalf("创建解析链失败: %v", err)
	}

	// 编译链
	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译链失败: %v", err)
	}

	// 准备输入数据
	input := &jobprofileparser.JobProfileParseInput{
		Description: `
岗位名称：资深后端工程师
工作性质：全职
工作地点：北京市朝阳区
薪资范围：25000-35000元/月

我们正在寻找一名资深后端工程师加入我们的技术团队。

职位职责：
1. 负责公司核心业务系统的后端架构设计和开发
2. 参与技术方案评审，制定技术标准和规范
3. 优化系统性能，保障服务的高可用性和稳定性
4. 指导初级工程师，参与代码审查
5. 参与产品需求分析，提供技术解决方案

任职要求：
- 本科及以上学历，计算机相关专业
- 3-5年后端开发经验，有大型互联网项目经验优先
- 精通Java/Go/Python中的至少一种语言
- 熟悉Spring Boot、微服务架构
- 熟悉MySQL、Redis等数据库技术
- 了解Docker、Kubernetes等容器技术加分
- 有金融、电商行业经验优先
- 良好的沟通能力和团队协作精神

我们提供：
- 有竞争力的薪资待遇
- 完善的培训体系
- 良好的职业发展空间
		`,
	}

	// 执行解析
	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		log.Fatalf("执行解析失败: %v", err)
	}

	// 输出结果
	fmt.Println("=== 岗位画像解析结果 ===")

	// 输出基本信息
	fmt.Println("\n基本信息:")
	fmt.Printf("岗位名称: %s\n", result.Name)
	if result.WorkType != nil {
		fmt.Printf("工作性质: %s\n", *result.WorkType)
	}
	if result.Location != nil {
		fmt.Printf("工作地点: %s\n", *result.Location)
	}
	if result.SalaryMin != nil && result.SalaryMax != nil {
		fmt.Printf("薪资范围: %.0f-%.0f元/月\n", *result.SalaryMin, *result.SalaryMax)
	}

	fmt.Println("\n职责列表:")
	for i, resp := range result.Responsibilities {
		fmt.Printf("%d. %s\n", i+1, resp.Responsibility)
	}

	fmt.Println("\n技能要求:")
	for i, skill := range result.Skills {
		fmt.Printf("%d. %s (%s)\n", i+1, skill.Skill, skill.Type)
	}

	fmt.Println("\n学历要求:")
	for i, edu := range result.EducationRequirements {
		fmt.Printf("%d. %s\n", i+1, edu.EducationType)
	}

	fmt.Println("\n经验要求:")
	for i, exp := range result.ExperienceRequirements {
		fmt.Printf("%d. %s (最少%d年，理想%d年)\n", i+1, exp.ExperienceType, exp.MinYears, exp.IdealYears)
	}

	fmt.Println("\n行业要求:")
	for i, ind := range result.IndustryRequirements {
		companyInfo := ""
		if ind.CompanyName != nil {
			companyInfo = fmt.Sprintf(" (公司: %s)", *ind.CompanyName)
		}
		fmt.Printf("%d. %s%s\n", i+1, ind.Industry, companyInfo)
	}

	// 输出完整JSON结果
	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("\n=== 完整JSON结果 ===\n%s\n", string(jsonData))
}
