package websearch

import (
	"context"
	"fmt"

	"github.com/chaitin/WhaleHire/backend/pkg/eino/tools"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// WebSearchConfig 网页搜索配置
type WebSearchConfig struct {
	// 可以添加搜索相关配置，如超时时间、最大结果数等
	MaxResults int               `json:"max_results,omitempty"`
	Region     duckduckgo.Region `json:"region,omitempty"`
}

// WebSearchChain 网页搜索链结构
type WebSearchChain struct {
	tools []tool.BaseTool
	chain *compose.Chain[*UserMessage, *WebSearchResult]
}

// GetChain 获取处理链
func (w *WebSearchChain) GetChain() *compose.Chain[*UserMessage, *WebSearchResult] {
	return w.chain
}

func (w *WebSearchChain) Compile(ctx context.Context) (compose.Runnable[*UserMessage, *WebSearchResult], error) {
	agent, err := w.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return agent, nil
}

// createSearchTools 创建搜索工具
func createSearchTools(ctx context.Context, config *WebSearchConfig) ([]tool.BaseTool, error) {
	searchTool, err := tools.NewDDGSearchTool(ctx, &duckduckgo.Config{
		MaxResults: config.MaxResults,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create search tool: %w", err)
	}

	return []tool.BaseTool{searchTool}, nil
}

// buildToolInfos 构建工具信息
func buildToolInfos(ctx context.Context, tools []tool.BaseTool) ([]*schema.ToolInfo, error) {
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, tool := range tools {
		info, err := tool.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tool info: %w", err)
		}
		toolInfos = append(toolInfos, info)
	}
	return toolInfos, nil
}

// BuildWebSearchChain 构建网页搜索工作流链
// 工作流程：用户查询 -> 模型生成查询请求 -> 搜索工具 -> 返回结果
func NewWebSearchChain(ctx context.Context, searchChat model.ToolCallingChatModel) (*WebSearchChain, error) {
	return NewWebSearchChainWithConfig(ctx, searchChat, nil)
}

// BuildWebSearchChainWithConfig 使用配置构建网页搜索工作流链
func NewWebSearchChainWithConfig(ctx context.Context, searchChat model.ToolCallingChatModel, config *WebSearchConfig) (*WebSearchChain, error) {
	if config == nil {
		config = &WebSearchConfig{}
	}

	// 1. 创建搜索工具
	todoTools, err := createSearchTools(ctx, config)
	if err != nil {
		return nil, err
	}

	// 2. 创建 tools 节点
	todoToolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: todoTools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tools node: %w", err)
	}

	// 3. 创建聊天模板
	chatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 4. 构建工具信息并配置聊天模型
	toolInfos, err := buildToolInfos(ctx, todoTools)
	if err != nil {
		return nil, err
	}

	searchChat, err = searchChat.WithTools(toolInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to configure chat model with tools: %w", err)
	}

	// 5. 构建完整的处理链
	chain := compose.NewChain[*UserMessage, *WebSearchResult]()

	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newLambdaWithHistory), compose.WithNodeName("userinput")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chattemplate")).
		AppendChatModel(searchChat, compose.WithNodeName("chatmodel")).
		AppendToolsNode(todoToolsNode).
		AppendLambda(compose.InvokableLambdaWithOption(newLambdaCovertWebSearchResult), compose.WithNodeName("covertresult"))

	return &WebSearchChain{
		tools: todoTools,
		chain: chain,
	}, nil
}
