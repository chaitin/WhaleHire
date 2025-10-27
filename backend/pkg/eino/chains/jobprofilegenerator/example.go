//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofilegenerator"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
)

// ExampleUsage 展示如何使用岗位画像生成链
func main() {
	ctx := context.Background()

	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	factory := models.NewModelFactory()
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

	chain, err := jobprofilegenerator.NewJobProfileGeneratorChain(ctx, chatModel)
	if err != nil {
		log.Fatalf("创建生成链失败: %v", err)
	}

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译生成链失败: %v", err)
	}

	input := &jobprofilegenerator.JobProfileGenerateInput{
		Prompt: `我们需要招聘一位AI产品经理，负责公司人工智能相关产品的全生命周期管理。该岗位将主导AI产品的需求分析、功能设计、产品路线图规划以及跨部门协作推进产品落地。候选人需要具备3年以上AI或互联网产品经验，熟悉机器学习/深度学习技术原理，有成功的AI产品商业化经验者优先。工作地点为北京，需要与算法团队、工程团队和业务部门紧密合作，推动AI技术转化为具有市场竞争力的产品。`,
	}

	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		log.Fatalf("执行生成失败: %v", err)
	}

	fmt.Println("=== 岗位画像生成结果 ===")
	fmt.Printf("岗位名称：%s\n", result.Profile.Name)
	if result.Profile.WorkType != nil {
		fmt.Printf("工作性质：%s\n", *result.Profile.WorkType)
	}
	if result.Profile.Location != nil {
		fmt.Printf("工作地点：%s\n", *result.Profile.Location)
	}

	fmt.Println("\n岗位职责：")
	for idx, resp := range result.Profile.Responsibilities {
		fmt.Printf("%d. %s\n", idx+1, resp.Responsibility)
	}

	fmt.Println("\nMarkdown 描述：")
	fmt.Println(result.DescriptionMarkdown)

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("\n=== 完整JSON结果 ===\n%s\n", string(jsonData))
}
