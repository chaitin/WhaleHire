//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/chains/websearch"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
)

func main() {

	ctx := context.Background()
	factory := models.NewModelFactory()

	// 获取配置
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}
	// 注册OpenAI模型
	openaiConfig := &models.OpenAIConfig{
		APIKey:  cfg.GeneralAgent.LLM.APIKey,
		BaseURL: cfg.GeneralAgent.LLM.BaseURL,
		Model:   cfg.GeneralAgent.LLM.ModelName,
	}
	openaiManager := models.NewOpenAIModelManager(openaiConfig)
	factory.Register(models.ModelTypeOpenAI, "deepseek-chat-1", openaiManager)

	chatModel, err := factory.GetModel(ctx, models.ModelTypeOpenAI, "deepseek-chat-1")

	chain, err := websearch.NewWebSearchChain(ctx, chatModel)
	if err != nil {
		log.Fatalf("failed to create web search chain: %v", err)
	}

	agent, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("failed to compile chain: %v", err)
	}

	resp, err := agent.Invoke(ctx, &websearch.UserMessage{
		Query: "最新 AI 咨询",
	})

	if err != nil {
		log.Fatalf("failed to invoke agent: %v", err)
	}

	// 输出结果
	fmt.Printf("%+v", resp)

}
