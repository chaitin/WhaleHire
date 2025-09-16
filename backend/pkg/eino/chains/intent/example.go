//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/schema"
	"github.com/ptonlix/whalehire/backend/pkg/eino/chains/intent"
	"github.com/ptonlix/whalehire/backend/pkg/eino/models"
)

func main() {

	ctx := context.Background()
	factory := models.NewModelFactory()
	// 注册OpenAI模型
	openaiConfig := &models.OpenAIConfig{
		APIKey:  "sk-b1d10d968f074cf083f842adf86b4f2c",
		BaseURL: "https://api.deepseek.com/v1",
		Model:   "deepseek-chat",
	}
	openaiManager := models.NewOpenAIModelManager(openaiConfig)
	factory.Register(models.ModelTypeOpenAI, "deepseek-chat-1", openaiManager)

	chatModel, err := factory.GetModel(ctx, models.ModelTypeOpenAI, "deepseek-chat-1")

	chain, err := intent.NewIntentClassificationChain(ctx, chatModel)
	if err != nil {
		log.Fatalf("failed to create web search chain: %v", err)
	}

	agent, err := chain.Compile(ctx)
	if err != nil {
		log.Fatalf("failed to compile chain: %v", err)
	}

	resp, err := agent.Invoke(ctx, &intent.IntentInput{
		Query: "特斯拉是什么样的公司",
		History: []*schema.Message{
			{
				Role:    schema.User,
				Content: "你好",
			},
			{
				Role:    schema.Assistant,
				Content: "我是DeepSeek",
			},
		},
	})

	if err != nil {
		log.Fatalf("failed to invoke agent: %v", err)
	}

	// 输出结果
	fmt.Println(resp.Intent)

}
