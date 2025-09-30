//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/resumeparser"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
)

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

	chain, err := resumeparser.NewResumeParserChain(ctx, chatModel)
	if err != nil {
		log.Fatalf("failed to create resume parser chain: %v", err)
	}

	agent, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("failed to compile chain: %v", err)
	}

	// 示例简历文本
	sample := `张三，邮箱 zhangsan@example.com，电话 13800001234。毕业于清华大学计算机科学与技术专业，本科。曾在字节跳动担任后端工程师（2021-07 至 2024-03），负责服务治理与性能优化。熟悉 Go、Python、Docker、Kubernetes。`
	out, err := agent.Invoke(ctx, &resumeparser.ResumeParseInput{Resume: sample})
	if err != nil {
		log.Fatalf("failed to invoke agent: %v", err)
	}

	fmt.Printf("%+v\n", out)
}
