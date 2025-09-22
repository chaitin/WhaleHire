package retrieverchat

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/config"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// RetrieverChatGraph 文档召回聊天图结构
type RetrieverChatGraph struct {
	graph *compose.Graph[*UserMessage, []*schema.Message]
}

// GetGraph 获取图结构
func (r *RetrieverChatGraph) GetGraph() *compose.Graph[*UserMessage, []*schema.Message] {
	return r.graph
}

// Compile 编译链为可执行的Runnable
func (r *RetrieverChatGraph) Compile(ctx context.Context) (compose.Runnable[*UserMessage, []*schema.Message], error) {
	agent, err := r.graph.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return agent, nil
}
func newtest(ctx context.Context, input []*schema.Document, opts ...any) (output []*schema.Message, err error) {
	for _, doc := range input {
		fmt.Println(doc.ID, doc.Content)
		fmt.Println("-----------")
	}
	return output, nil
}

func newtest1(ctx context.Context, input string, opts ...any) (output string, err error) {
	fmt.Println(input)
	return input, nil
}

// NewRetrieverChatGraph 使用配置创建文档召回聊天图
func NewRetrieverChatGraph(ctx context.Context, chatModel model.ToolCallingChatModel, cfg *config.Config) (*RetrieverChatGraph, error) {
	const (
		ChatTemplate   = "ChatTemplate"
		RedisRetriever = "RedisRetriever"
		InputToQuery   = "InputToQuery"
		InputToHistory = "InputToHistory"
	)
	// 1. 创建文档召回器
	retrieverInstance, err := NewRetriever(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create retriever: %w", err)
	}

	// 2. 创建聊天模板
	chatTemplate, err := NewChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	g := compose.NewGraph[*UserMessage, []*schema.Message]()

	_ = g.AddLambdaNode(InputToHistory, compose.InvokableLambdaWithOption(newLambdaWithHistory), compose.WithNodeName("UserMessageToVariables"))
	_ = g.AddLambdaNode(InputToQuery, compose.InvokableLambdaWithOption(newLambdaWithQuery), compose.WithNodeName("UserMessageToQuery"))
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplate, compose.WithNodeName("ChatTemplate"))
	_ = g.AddRetrieverNode(RedisRetriever, retrieverInstance)
	_ = g.AddLambdaNode("test", compose.InvokableLambdaWithOption(newtest), compose.WithNodeName("test"))
	_ = g.AddLambdaNode("test1", compose.InvokableLambdaWithOption(newtest1), compose.WithNodeName("test"))

	_ = g.AddEdge(compose.START, InputToHistory)
	_ = g.AddEdge(compose.START, InputToQuery)
	_ = g.AddEdge(InputToQuery, "test1")
	_ = g.AddEdge(InputToQuery, RedisRetriever)
	_ = g.AddEdge(RedisRetriever, "test")
	_ = g.AddEdge("test", compose.END)

	// _ = g.AddEdge(RedisRetriever, ChatTemplate)
	// _ = g.AddEdge(InputToHistory, ChatTemplate)
	// _ = g.AddEdge(ChatTemplate, compose.END)

	return &RetrieverChatGraph{
		graph: g,
	}, nil
}
