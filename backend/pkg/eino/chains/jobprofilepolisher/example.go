//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/jobprofilepolisher"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
)

// ExampleUsage 展示如何使用岗位 Prompt 润色链
func main() {
	ctx := context.Background()

	// 初始化配置（需提前准备好 GeneralAgent.LLM 相关配置）
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

	chain, err := jobprofilepolisher.NewJobProfilePolisherChain(ctx, chatModel)
	if err != nil {
		log.Fatalf("创建润色链失败: %v", err)
	}

	runnable, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("编译链失败: %v", err)
	}

	input := &jobprofilepolisher.PolishJobPromptInput{
		Idea: "我想招聘一位AI产品经理。",
	}

	result, err := runnable.Invoke(ctx, input)
	if err != nil {
		log.Fatalf("执行润色失败: %v", err)
	}

	fmt.Println("=== Prompt 润色结果 ===")
	fmt.Printf("\n推荐岗位标题：%s\n", result.SuggestedTitle)
	fmt.Printf("\n润色后的 Prompt：\n%s\n", result.PolishedPrompt)

	printList := func(title string, items []string) {
		fmt.Printf("\n%s:\n", title)
		if len(items) == 0 {
			fmt.Println("（暂无建议）")
			return
		}
		for i, item := range items {
			fmt.Printf("%d. %s\n", i+1, item)
		}
	}

	printList("岗位职责提示", result.ResponsibilityTips)
	printList("任职要求提示", result.RequirementTips)
	printList("加分项提示", result.BonusTips)

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("\n=== 完整JSON结果 ===\n%s\n", string(jsonData))
}
