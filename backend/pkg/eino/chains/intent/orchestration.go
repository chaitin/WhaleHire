package intent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// IntentClassificationChain 意图分类链结构
type IntentClassificationChain struct {
	chain *compose.Chain[*IntentInput, *IntentResult]
}

// GetChain 获取处理链
func (ic *IntentClassificationChain) GetChain() *compose.Chain[*IntentInput, *IntentResult] {
	return ic.chain
}

// Compile 编译链为可执行的Runnable
func (ic *IntentClassificationChain) Compile(ctx context.Context) (compose.Runnable[*IntentInput, *IntentResult], error) {
	agent, err := ic.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile intent classification chain: %w", err)
	}
	return agent, nil
}

// newIntentInputLambda 将输入转换为模板所需的格式
func newIntentInputLambda(ctx context.Context, input *IntentInput, opts ...any) (output map[string]any, err error) {
	return map[string]any{
		"query":   input.Query,
		"history": input.History,
	}, nil
}

// newIntentOutputLambda 将模型输出转换为意图结果
func newIntentOutputLambda(ctx context.Context, input *schema.Message, opts ...any) (output *IntentResult, err error) {
	if input.Content == "" {
		return nil, fmt.Errorf("no output from model")
	}

	// 获取最后一条消息的内容
	lastMessage := input
	if lastMessage.Content == "" {
		return nil, fmt.Errorf("empty content in model output")
	}

	content := lastMessage.Content

	// 尝试解析JSON格式的意图结果
	var result IntentResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// 如果JSON解析失败，尝试简单的文本匹配
		if containsIntent(content, IntentSearchOnline) {
			result.Intent = IntentSearchOnline
		} else {
			result.Intent = IntentAnswerDirectly
		}
	}

	// 验证意图是否有效
	if !isValidIntent(result.Intent) {
		result.Intent = IntentAnswerDirectly // 默认意图
	}

	return &result, nil
}

// containsIntent 检查内容是否包含特定意图
func containsIntent(content, intent string) bool {
	return len(content) > 0 && len(intent) > 0 &&
		(content == intent ||
			fmt.Sprintf("\"%s\"", intent) == content ||
			fmt.Sprintf("'%s'", intent) == content)
}

// isValidIntent 验证意图是否有效
func isValidIntent(intent string) bool {
	switch intent {
	case IntentSearchOnline, IntentAnswerDirectly, IntentImageGeneration, IntentCalculator, IntentDatabaseQuery:
		return true
	default:
		return false
	}
}

// NewIntentClassificationChain 创建意图分类链
func NewIntentClassificationChain(ctx context.Context, chatModel model.ToolCallingChatModel) (*IntentClassificationChain, error) {
	// 1. 创建聊天模板
	chatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	// 2. 构建完整的处理链
	// 输入处理 -> 聊天模板 -> 聊天模型 -> 输出处理
	chain := compose.NewChain[*IntentInput, *IntentResult]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newIntentInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newIntentOutputLambda), compose.WithNodeName("output_processing"))

	return &IntentClassificationChain{
		chain: chain,
	}, nil
}
