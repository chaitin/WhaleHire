package resumeparser

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ResumeParserChain 简历解析链结构
type ResumeParserChain struct {
	chain *compose.Chain[*ResumeParseInput, *ResumeParseResult]
}

// GetChain 获取处理链
func (c *ResumeParserChain) GetChain() *compose.Chain[*ResumeParseInput, *ResumeParseResult] {
	return c.chain
}

// Compile 编译链为可执行的Runnable
func (c *ResumeParserChain) Compile(ctx context.Context) (compose.Runnable[*ResumeParseInput, *ResumeParseResult], error) {
	r, err := c.chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile chain: %w", err)
	}
	return r, nil
}

// 输入处理，将 Resume 文本与历史对话传入模板
func newInputLambda(ctx context.Context, input *ResumeParseInput, opts ...any) (map[string]any, error) {
	return map[string]any{
		"resume": input.Resume,
		// "date":   time.Now().Format("2006-01-02"),
	}, nil
}

// 输出处理，将模型输出解析为结构化结果
func newOutputLambda(ctx context.Context, msg *schema.Message, opts ...any) (*ResumeParseResult, error) {

	if msg == nil || msg.Content == "" {
		return nil, fmt.Errorf("empty model output")
	}
	res, err := decodeResumeParseResult([]byte(msg.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w; raw=%s", err, msg.Content)
	}
	return res, nil
}

// NewResumeParserChain 构建简历解析工作流链
// 工作流程：输入 -> 模板 -> 聊天模型(GPT5/兼容) -> 结果解析
func NewResumeParserChain(ctx context.Context, chatModel model.ToolCallingChatModel) (*ResumeParserChain, error) {
	chatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat template: %w", err)
	}

	chain := compose.NewChain[*ResumeParseInput, *ResumeParseResult]()
	chain.
		AppendLambda(compose.InvokableLambdaWithOption(newInputLambda), compose.WithNodeName("input_processing")).
		AppendChatTemplate(chatTemplate, compose.WithNodeName("chat_template")).
		AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
		AppendLambda(compose.InvokableLambdaWithOption(newOutputLambda), compose.WithNodeName("output_processing"))

	return &ResumeParserChain{chain: chain}, nil
}
