//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"

	"github.com/ptonlix/whalehire/backend/config"
	"github.com/ptonlix/whalehire/backend/pkg/eino/graphs/retrieverchat"
	"github.com/ptonlix/whalehire/backend/pkg/eino/models"
)

func main() {
	ctx := context.Background()
	// 获取配置
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	factory := models.NewModelFactory()
	// 注册OpenAI模型
	openaiConfig := &models.OpenAIConfig{
		APIKey:  cfg.GeneralAgent.LLM.APIKey,
		BaseURL: cfg.GeneralAgent.LLM.BaseURL,
		Model:   cfg.GeneralAgent.LLM.ModelName,
	}
	openaiManager := models.NewOpenAIModelManager(openaiConfig)
	factory.Register(models.ModelTypeOpenAI, "deepseek-chat-1", openaiManager)

	chatModel, err := factory.GetModel(ctx, models.ModelTypeOpenAI, "deepseek-chat-1")

	graph, err := retrieverchat.NewRetrieverChatGraph(ctx, chatModel, cfg)

	agent, err := graph.Compile(ctx)
	if err != nil {
		panic(err)
	}

	// 执行
	input := &retrieverchat.UserMessage{
		Input: "请问 Eino 是什么框架",
	}
	output, err := agent.Invoke(ctx, input)
	if err != nil {
		panic(err)
	}

	// 输出结果
	for _, msg := range output {
		fmt.Printf("%+v", msg)
	}
}
