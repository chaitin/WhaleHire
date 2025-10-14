//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/graphs/retrieverchat"
	"github.com/chaitin/WhaleHire/backend/pkg/eino/models"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/compose"
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

	// 创建 callback handler 示例
	// 这个 handler 将在组件执行的不同阶段被调用，用于日志记录、监控等目的
	handler := callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			log.Printf("[CALLBACK] 组件开始执行 - 名称: %s, 类型: %s", info.Name, info.Type)
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			log.Printf("[CALLBACK] 组件执行完成 - 名称: %s, 类型: %s", info.Name, info.Type)
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			log.Printf("[CALLBACK] 组件执行出错 - 名称: %s, 类型: %s, 错误: %v", info.Name, info.Type, err)
			return ctx
		}).
		Build()

	// 执行
	input := &retrieverchat.UserMessage{
		Input: "请问 Eino 是什么框架",
	}
	output, err := agent.Invoke(ctx, input, // 只对特定节点生效的 call option
		compose.WithCallbacks(handler).DesignateNode("node_1"))
	if err != nil {
		panic(err)
	}

	// 输出结果
	for _, msg := range output {
		fmt.Printf("%+v", msg)
	}
}
