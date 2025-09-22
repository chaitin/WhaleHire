package retrieverchat

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

var systemPrompt = `
你是一个专业的文档问答助手。你的任务是：
1. 严格依据提供的文档内容回答用户问题
2. 不要引入文档中未出现的信息，也不要编造
3. 如果文档中没有相关信息，请直接回答“没有在文档中找到答案”
4. 回答需简洁、准确，必要时可引用文档中的关键片段

提供的文档内容如下：
==== doc start ====
{documents}
==== doc end ====

当前时间: {date}
`

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

func NewChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	config := &ChatTemplateConfig{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(systemPrompt),
			schema.MessagesPlaceholder("history", false),
			schema.UserMessage("{input}"),
		},
	}
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}
